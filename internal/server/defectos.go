package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/defect"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// defectos agrupa los handlers de los defectos de un proyecto (HU-27).
type defectos struct {
	casos *service.Defectos
}

// paginaDefectos son los datos de defectos.html.
type paginaDefectos struct {
	service.VistaDefectos
	Mensaje     string
	Aviso       string
	Nuevo       defect.Datos
	Errores     map[string]string
	Severidades []defect.Severidad
	Filtro      defect.Severidad // severidad del resumen (HU-34); "" = todas
}

// ResumenFiltrado devuelve el resumen de defectos con el filtro de la página (HU-34).
func (p paginaDefectos) ResumenFiltrado() service.ResumenDefectos {
	return p.Resumen(p.Filtro)
}

// ver muestra los defectos del proyecto y el formulario para registrar uno.
func (h defectos) ver(w http.ResponseWriter, r *http.Request) {
	proyectoID, ok := idDeProyecto(w, r)
	if !ok {
		return
	}
	vista, err := h.casos.Ver(r.Context(), proyectoID)
	if respondeErrorDeTareas(w, err, "ver los defectos") {
		return
	}
	pagina := paginaDefectos{VistaDefectos: vista, Severidades: defect.Severidades, Filtro: defect.Severidad(r.URL.Query().Get("severidad"))}
	if r.URL.Query().Get("registrado") == "1" {
		pagina.Mensaje = "Defecto registrado."
	}
	// Solo se muestra el mensaje de un estado conocido: el texto de la URL nunca llega a la página.
	if e, ok := estadosPorSlug[r.URL.Query().Get("estado")]; ok {
		pagina.Mensaje = "Defecto en estado " + string(e) + "."
	}
	renderDefectos(w, http.StatusOK, pagina)
}

// registrar recibe el formulario. Si los datos son válidos vuelve a la lista (303); si no, la muestra con cada error
// debajo de su campo y lo que se escribió.
func (h defectos) registrar(w http.ResponseWriter, r *http.Request) {
	proyectoID, ok := idDeProyecto(w, r)
	if !ok {
		return
	}
	// Un sprint o una historia que no son un número son como no haber elegido ninguno.
	sprintID, _ := strconv.ParseInt(r.PostFormValue("sprint"), 10, 64)
	historiaID, _ := strconv.ParseInt(r.PostFormValue("historia"), 10, 64)
	d := defect.Datos{
		Descripcion: r.PostFormValue("descripcion"), Severidad: defect.Severidad(r.PostFormValue("severidad")),
		SprintID: sprintID, HistoriaID: historiaID,
	}
	_, err := h.casos.Registrar(r.Context(), proyectoID, d)

	var errs domain.ErroresValidacion
	switch {
	case err == nil:
		http.Redirect(w, r, fmt.Sprintf("/proyectos/%d/defectos?registrado=1", proyectoID), http.StatusSeeOther)
	case errors.As(err, &errs):
		vista, err := h.casos.Ver(r.Context(), proyectoID)
		if respondeErrorDeTareas(w, err, "volver a mostrar los defectos") {
			return
		}
		pagina := paginaDefectos{VistaDefectos: vista, Nuevo: d, Severidades: defect.Severidades, Errores: make(map[string]string, len(errs))}
		for campo, e := range errs {
			pagina.Errores[campo] = e.Error()
		}
		renderDefectos(w, http.StatusUnprocessableEntity, pagina)
	default:
		respondeErrorDeTareas(w, err, "registrar un defecto")
	}
}

// estadosPorSlug son los estados a los que se puede pasar un defecto, por cómo aparecen en la URL (?estado=en-progreso).
var estadosPorSlug = map[string]defect.Estado{}

func init() {
	for _, e := range []defect.Estado{defect.EstadoEnProgreso, defect.EstadoResuelto, defect.EstadoCerrado, defect.EstadoReabierto} {
		estadosPorSlug[slugDeEstado(e)] = e
	}
}

func slugDeEstado(e defect.Estado) string {
	return strings.ReplaceAll(strings.ToLower(string(e)), " ", "-")
}

// cambiarEstado pasa el defecto al estado elegido (HU-28). Si el cambio vale vuelve a la lista (303); si la transición
// o el sprint de resolución no valen, muestra la lista con el motivo (409). Un estado que no existe es un pedido
// inválido (400) y un defecto que no es del proyecto, 404.
func (h defectos) cambiarEstado(w http.ResponseWriter, r *http.Request) {
	proyectoID, ok := idDeProyecto(w, r)
	if !ok {
		return
	}
	defectoID, err := strconv.ParseInt(chi.URLParam(r, "defecto"), 10, 64)
	if err != nil {
		defectoNoEncontrado(w)
		return
	}
	// Un sprint que no es un número es como no haber elegido ninguno: el dominio lo rechaza al resolver.
	sprintID, _ := strconv.ParseInt(r.PostFormValue("sprint"), 10, 64)
	d, err := h.casos.CambiarEstado(r.Context(), proyectoID, defectoID, defect.Estado(r.PostFormValue("estado")), sprintID)

	switch {
	case err == nil:
		http.Redirect(w, r, fmt.Sprintf("/proyectos/%d/defectos?estado=%s", proyectoID, slugDeEstado(d.Estado)), http.StatusSeeOther)
	case errors.Is(err, defect.ErrEstadoInvalido):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, defect.ErrDefectoNoEncontrado):
		defectoNoEncontrado(w)
	case errors.Is(err, defect.ErrTransicionInvalida), errors.Is(err, defect.ErrSprintResolucionInvalido),
		errors.Is(err, defect.ErrSprintResolucionAnterior):
		vista, errVista := h.casos.Ver(r.Context(), proyectoID)
		if respondeErrorDeTareas(w, errVista, "volver a mostrar los defectos") {
			return
		}
		renderDefectos(w, http.StatusConflict, paginaDefectos{VistaDefectos: vista, Severidades: defect.Severidades, Aviso: err.Error()})
	default:
		respondeErrorDeTareas(w, err, "cambiar el estado de un defecto")
	}
}

func defectoNoEncontrado(w http.ResponseWriter) {
	http.Error(w, "No encontramos ese defecto.", http.StatusNotFound)
}

func renderDefectos(w http.ResponseWriter, codigo int, pagina paginaDefectos) {
	render(w, codigo, "defectos.html", Pagina{Titulo: "Defectos", Datos: pagina})
}
