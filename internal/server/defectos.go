package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

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
	Nuevo       defect.Datos
	Errores     map[string]string
	Severidades []defect.Severidad
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
	pagina := paginaDefectos{VistaDefectos: vista, Severidades: defect.Severidades}
	if r.URL.Query().Get("registrado") == "1" {
		pagina.Mensaje = "Defecto registrado."
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

func renderDefectos(w http.ResponseWriter, codigo int, pagina paginaDefectos) {
	render(w, codigo, "defectos.html", Pagina{Titulo: "Defectos", Datos: pagina})
}
