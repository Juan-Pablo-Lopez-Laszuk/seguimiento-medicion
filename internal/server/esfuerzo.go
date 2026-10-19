package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// esfuerzo agrupa los handlers del registro de esfuerzo de una historia (HU-24).
type esfuerzo struct {
	casos *service.Esfuerzo
}

// paginaEsfuerzo son los datos de esfuerzo.html: la historia con sus registros y sus horas reales, y lo que cambia
// según el pedido (el mensaje de éxito y lo escrito en el formulario con sus errores).
type paginaEsfuerzo struct {
	service.VistaEsfuerzo
	Mensaje string
	Nuevo   effort.DatosRegistro
	Errores map[string]string
}

// ver muestra los registros de la historia, sus horas reales y el formulario.
func (h esfuerzo) ver(w http.ResponseWriter, r *http.Request) {
	proyectoID, numero, ok := idsDeHistoria(w, r)
	if !ok {
		return
	}
	vista, err := h.casos.Ver(r.Context(), proyectoID, numero)
	if respondeErrorDeTareas(w, err, "ver el esfuerzo") {
		return
	}
	pagina := paginaEsfuerzo{VistaEsfuerzo: vista, Nuevo: effort.DatosRegistro{Fecha: vista.Hoy.Format("2006-01-02")}}
	if r.URL.Query().Get("registrado") == "1" {
		pagina.Mensaje = "Horas registradas."
	}
	renderEsfuerzo(w, http.StatusOK, pagina)
}

// registrar recibe el formulario. Si los datos son válidos vuelve a la página (303); si no, la muestra con cada error
// debajo de su campo y lo que se escribió.
func (h esfuerzo) registrar(w http.ResponseWriter, r *http.Request) {
	proyectoID, numero, ok := idsDeHistoria(w, r)
	if !ok {
		return
	}
	// Un integrante o una tarea que no son un número son como no haber elegido ninguno.
	integrante, _ := strconv.ParseInt(r.PostFormValue("integrante"), 10, 64)
	tarea, _ := strconv.ParseInt(r.PostFormValue("tarea"), 10, 64)
	d := effort.DatosRegistro{
		IntegranteID: integrante, TareaID: tarea, Fecha: r.PostFormValue("fecha"),
		Actividad: r.PostFormValue("actividad"), Horas: r.PostFormValue("horas"),
	}
	_, err := h.casos.Registrar(r.Context(), proyectoID, numero, d)

	var errs domain.ErroresValidacion
	switch {
	case err == nil:
		http.Redirect(w, r, fmt.Sprintf("/proyectos/%d/backlog/%d/esfuerzo?registrado=1", proyectoID, numero), http.StatusSeeOther)
	case errors.As(err, &errs):
		vista, err := h.casos.Ver(r.Context(), proyectoID, numero)
		if respondeErrorDeTareas(w, err, "volver a mostrar el esfuerzo") {
			return
		}
		pagina := paginaEsfuerzo{VistaEsfuerzo: vista, Nuevo: d, Errores: make(map[string]string, len(errs))}
		for campo, e := range errs {
			pagina.Errores[campo] = e.Error()
		}
		renderEsfuerzo(w, http.StatusUnprocessableEntity, pagina)
	default:
		respondeErrorDeTareas(w, err, "registrar esfuerzo")
	}
}

func renderEsfuerzo(w http.ResponseWriter, codigo int, pagina paginaEsfuerzo) {
	render(w, codigo, "esfuerzo.html", Pagina{Titulo: "Esfuerzo de " + pagina.Historia.Identificador(), Datos: pagina})
}
