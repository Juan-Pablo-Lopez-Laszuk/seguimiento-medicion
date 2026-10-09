package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// integrantes agrupa los handlers de los integrantes de un proyecto (HU-03).
type integrantes struct {
	casos *service.Integrantes
}

// listaIntegrantes son los datos de la página integrantes.html.
type listaIntegrantes struct {
	Proyecto    project.Proyecto
	Integrantes []member.Integrante
	Registrado  bool   // aviso "Integrante registrado"
	Baja        bool   // aviso "Integrante dado de baja"
	Error       string // aviso de error de la baja (por ejemplo, ya estaba dado de baja)
}

// formIntegrante son los datos de la página integrante_nuevo.html.
type formIntegrante struct {
	Proyecto     project.Proyecto
	Nombre       string
	Email        string
	Rol          member.Rol
	Roles        []member.Rol
	Errores      map[string]string
	ErrorGeneral string
}

// listar muestra el proyecto y todos sus integrantes, activos y dados de baja.
func (h integrantes) listar(w http.ResponseWriter, r *http.Request) {
	datos, ok := h.cargar(w, r)
	if !ok {
		return
	}
	datos.Registrado = r.URL.Query().Get("registrado") == "1"
	datos.Baja = r.URL.Query().Get("baja") == "1"
	render(w, http.StatusOK, "integrantes.html", Pagina{Titulo: "Integrantes", Datos: datos})
}

// nuevo muestra el formulario vacío.
func (h integrantes) nuevo(w http.ResponseWriter, r *http.Request) {
	datos, ok := h.cargar(w, r)
	if !ok {
		return
	}
	render(w, http.StatusOK, "integrante_nuevo.html", Pagina{
		Titulo: "Nuevo integrante",
		Datos:  formIntegrante{Proyecto: datos.Proyecto, Roles: member.Roles},
	})
}

// registrar recibe el formulario. Si los datos son válidos vuelve a la lista (redirección 303); si no,
// muestra el formulario con los errores y los datos cargados.
func (h integrantes) registrar(w http.ResponseWriter, r *http.Request) {
	id, ok := idDeProyecto(w, r)
	if !ok {
		return
	}
	form := formIntegrante{
		Nombre: r.PostFormValue("nombre"),
		Email:  r.PostFormValue("email"),
		Rol:    member.Rol(r.PostFormValue("rol")),
		Roles:  member.Roles,
	}
	_, err := h.casos.Registrar(r.Context(), id, member.Datos{Nombre: form.Nombre, Email: form.Email, Rol: form.Rol})

	var errs domain.ErroresValidacion
	switch {
	case err == nil:
		http.Redirect(w, r, fmt.Sprintf("/proyectos/%d/integrantes?registrado=1", id), http.StatusSeeOther)
		return
	case errors.Is(err, project.ErrNoEncontrado):
		proyectoNoEncontrado(w)
		return
	case errors.As(err, &errs):
		form.Errores = make(map[string]string, len(errs))
		for campo, e := range errs {
			form.Errores[campo] = e.Error()
		}
		// El proyecto se vuelve a buscar solo para mostrar el título y los enlaces del formulario.
		form.Proyecto, _, _ = h.casos.Listar(r.Context(), id)
		render(w, http.StatusUnprocessableEntity, "integrante_nuevo.html", Pagina{Titulo: "Nuevo integrante", Datos: form})
	default:
		log.Printf("registrar integrante en el proyecto %d: %v", id, err)
		form.Proyecto.ID = id
		form.ErrorGeneral = "No se pudo guardar el integrante. Probá de nuevo en unos minutos."
		render(w, http.StatusInternalServerError, "integrante_nuevo.html", Pagina{Titulo: "Nuevo integrante", Datos: form})
	}
}

// darDeBaja deja inactivo al integrante y vuelve a la lista. Si ya estaba dado de baja, muestra la lista
// con el aviso (409).
func (h integrantes) darDeBaja(w http.ResponseWriter, r *http.Request) {
	proyectoID, ok := idDeProyecto(w, r)
	if !ok {
		return
	}
	id, err := strconv.ParseInt(chi.URLParam(r, "integranteID"), 10, 64)
	if err != nil {
		integranteNoEncontrado(w)
		return
	}
	_, err = h.casos.DarDeBaja(r.Context(), proyectoID, id)
	switch {
	case err == nil:
		http.Redirect(w, r, fmt.Sprintf("/proyectos/%d/integrantes?baja=1", proyectoID), http.StatusSeeOther)
	case errors.Is(err, member.ErrNoEncontrado):
		integranteNoEncontrado(w)
	case errors.Is(err, member.ErrYaDadoDeBaja):
		datos, ok := h.cargar(w, r)
		if !ok {
			return
		}
		datos.Error = err.Error()
		render(w, http.StatusConflict, "integrantes.html", Pagina{Titulo: "Integrantes", Datos: datos})
	default:
		log.Printf("dar de baja al integrante %d del proyecto %d: %v", id, proyectoID, err)
		http.Error(w, "No se pudo dar de baja al integrante. Probá de nuevo en unos minutos.", http.StatusInternalServerError)
	}
}

// cargar busca el proyecto de la URL y sus integrantes. Si no puede, ya respondió (404 o 500) y ok es false.
func (h integrantes) cargar(w http.ResponseWriter, r *http.Request) (listaIntegrantes, bool) {
	id, ok := idDeProyecto(w, r)
	if !ok {
		return listaIntegrantes{}, false
	}
	p, lista, err := h.casos.Listar(r.Context(), id)
	switch {
	case errors.Is(err, project.ErrNoEncontrado):
		proyectoNoEncontrado(w)
		return listaIntegrantes{}, false
	case err != nil:
		log.Printf("listar integrantes del proyecto %d: %v", id, err)
		http.Error(w, "No se pudieron cargar los integrantes. Probá de nuevo en unos minutos.", http.StatusInternalServerError)
		return listaIntegrantes{}, false
	}
	return listaIntegrantes{Proyecto: p, Integrantes: lista}, true
}

func integranteNoEncontrado(w http.ResponseWriter) {
	http.Error(w, "No encontramos ese integrante.", http.StatusNotFound)
}
