package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// criterios agrupa los handlers de los criterios de aceptación de una historia (HU-07).
type criterios struct {
	casos *service.Criterios
}

// paginaHistoria son los datos de la página historia.html: la historia con sus criterios y su progreso, y lo que
// cambia según el pedido (el mensaje de éxito, el aviso y los errores del formulario de nuevo criterio).
type paginaHistoria struct {
	service.VistaHistoria
	Porcentaje  int    // progreso en porcentaje, para la barra
	Mensaje     string // mensaje de éxito después de agregar o marcar
	Aviso       string // por ejemplo, que la historia está Hecha y sus criterios no se pueden cambiar
	Descripcion string // lo que el usuario escribió en el formulario de nuevo criterio
	Errores     map[string]string
}

func nuevaPagina(v service.VistaHistoria) paginaHistoria {
	p := paginaHistoria{VistaHistoria: v}
	if v.Progreso.Total > 0 {
		p.Porcentaje = v.Progreso.Cumplidos * 100 / v.Progreso.Total
	}
	return p
}

// mensajeDeExito traduce los parámetros de la redirección a su texto. Solo conoce los valores que el servidor mismo
// pone: cualquier otro se ignora, así la página nunca muestra lo que venga en la URL.
func mensajeDeExito(consulta url.Values) string {
	switch {
	case consulta.Get("agregado") == "1":
		return "Criterio agregado."
	case consulta.Get("marcado") == "cumplido":
		return "Criterio marcado como cumplido."
	case consulta.Get("marcado") == "no-cumplido":
		return "Criterio marcado como no cumplido."
	case consulta.Get("estado") == "en-progreso":
		return "La historia pasó a En progreso."
	}
	return ""
}

// ver muestra la historia con sus criterios, su progreso y el formulario para agregar uno.
func (h criterios) ver(w http.ResponseWriter, r *http.Request) {
	proyectoID, numero, ok := idsDeHistoria(w, r)
	if !ok {
		return
	}
	vista, err := h.casos.Ver(r.Context(), proyectoID, numero)
	if respondeError(w, err, "ver la historia") {
		return
	}
	pagina := nuevaPagina(vista)
	pagina.Mensaje = mensajeDeExito(r.URL.Query())
	renderHistoria(w, http.StatusOK, pagina)
}

// agregar recibe el formulario. Si la descripción es válida vuelve a la página (redirección 303); si no, muestra la
// página con el error debajo del campo y lo que se escribió.
func (h criterios) agregar(w http.ResponseWriter, r *http.Request) {
	proyectoID, numero, ok := idsDeHistoria(w, r)
	if !ok {
		return
	}
	descripcion := r.PostFormValue("descripcion")
	_, err := h.casos.Agregar(r.Context(), proyectoID, numero, descripcion)

	var errs domain.ErroresValidacion
	switch {
	case err == nil:
		http.Redirect(w, r, fmt.Sprintf("/proyectos/%d/backlog/%d?agregado=1", proyectoID, numero), http.StatusSeeOther)
	case errors.As(err, &errs):
		pagina, ok := h.paginaDespuesDeUnError(w, r, proyectoID, numero)
		if !ok {
			return
		}
		pagina.Descripcion = descripcion
		pagina.Errores = make(map[string]string, len(errs))
		for campo, e := range errs {
			pagina.Errores[campo] = e.Error()
		}
		renderHistoria(w, http.StatusUnprocessableEntity, pagina)
	case errors.Is(err, backlog.ErrHistoriaHecha):
		h.avisoDeHistoriaHecha(w, r, proyectoID, numero, err)
	default:
		respondeError(w, err, "agregar un criterio")
	}
}

// marcar deja el criterio cumplido (cumplido=1) o no cumplido (cumplido=0) y vuelve a la página.
func (h criterios) marcar(w http.ResponseWriter, r *http.Request) {
	proyectoID, numero, ok := idsDeHistoria(w, r)
	if !ok {
		return
	}
	criterioID, err := strconv.ParseInt(chi.URLParam(r, "criterioID"), 10, 64)
	if err != nil {
		criterioNoEncontrado(w)
		return
	}
	valor := r.PostFormValue("cumplido")
	if valor != "1" && valor != "0" {
		http.Error(w, "Valor inválido.", http.StatusBadRequest)
		return
	}
	cumplido := valor == "1"
	_, err = h.casos.Marcar(r.Context(), proyectoID, numero, criterioID, cumplido)

	switch {
	case err == nil:
		resultado := "no-cumplido"
		if cumplido {
			resultado = "cumplido"
		}
		http.Redirect(w, r, fmt.Sprintf("/proyectos/%d/backlog/%d?marcado=%s", proyectoID, numero, resultado), http.StatusSeeOther)
	case errors.Is(err, backlog.ErrHistoriaHecha):
		h.avisoDeHistoriaHecha(w, r, proyectoID, numero, err)
	default:
		respondeError(w, err, "marcar un criterio")
	}
}

// avisoDeHistoriaHecha muestra la página con el aviso de que la historia está Hecha (409).
func (h criterios) avisoDeHistoriaHecha(w http.ResponseWriter, r *http.Request, proyectoID int64, numero int, err error) {
	pagina, ok := h.paginaDespuesDeUnError(w, r, proyectoID, numero)
	if !ok {
		return
	}
	pagina.Aviso = err.Error()
	renderHistoria(w, http.StatusConflict, pagina)
}

// paginaDespuesDeUnError vuelve a buscar la historia para mostrarla junto al error. Si no puede, ya respondió.
func (h criterios) paginaDespuesDeUnError(w http.ResponseWriter, r *http.Request, proyectoID int64, numero int) (paginaHistoria, bool) {
	vista, err := h.casos.Ver(r.Context(), proyectoID, numero)
	if respondeError(w, err, "volver a mostrar la historia") {
		return paginaHistoria{}, false
	}
	return nuevaPagina(vista), true
}

// respondeError responde 404 (proyecto, historia o criterio que no existen) o 500 si err no es nil, y dice si ya
// respondió. Con err nil no hace nada y devuelve false.
func respondeError(w http.ResponseWriter, err error, accion string) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, project.ErrNoEncontrado):
		proyectoNoEncontrado(w)
	case errors.Is(err, backlog.ErrHistoriaNoEncontrada):
		historiaNoEncontrada(w)
	case errors.Is(err, backlog.ErrCriterioNoEncontrado):
		criterioNoEncontrado(w)
	default:
		log.Printf("%s: %v", accion, err)
		http.Error(w, "No se pudo completar la operación. Probá de nuevo en unos minutos.", http.StatusInternalServerError)
	}
	return true
}

func renderHistoria(w http.ResponseWriter, codigo int, pagina paginaHistoria) {
	render(w, codigo, "historia.html", Pagina{Titulo: pagina.Historia.Identificador(), Datos: pagina})
}

// idsDeHistoria lee el {id} del proyecto y el {numero} de la historia de la URL. Si no son números responde 404, igual
// que si no existieran.
func idsDeHistoria(w http.ResponseWriter, r *http.Request) (proyectoID int64, numero int, ok bool) {
	proyectoID, ok = idDeProyecto(w, r)
	if !ok {
		return 0, 0, false
	}
	numero, err := strconv.Atoi(chi.URLParam(r, "numero"))
	if err != nil {
		historiaNoEncontrada(w)
		return 0, 0, false
	}
	return proyectoID, numero, true
}

func historiaNoEncontrada(w http.ResponseWriter) {
	http.Error(w, "No encontramos esa historia.", http.StatusNotFound)
}

func criterioNoEncontrado(w http.ResponseWriter) {
	http.Error(w, "No encontramos ese criterio.", http.StatusNotFound)
}
