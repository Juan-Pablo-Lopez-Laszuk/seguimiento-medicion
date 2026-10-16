package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// edicionHistorias agrupa los handlers de editar y eliminar una historia desde su página (HU-06).
type edicionHistorias struct {
	casos *service.EdicionHistorias
}

// formEdicionHistoria son los datos de la página historia_editar.html: la historia (para el título y los enlaces), lo
// que cargó el usuario como texto, las opciones del desplegable y el error de cada campo.
type formEdicionHistoria struct {
	Proyecto    project.Proyecto
	Historia    backlog.Historia
	Titulo      string
	Descripcion string
	Prioridad   backlog.Prioridad
	Prioridades []backlog.Prioridad
	Errores     map[string]string
}

func nuevoFormEdicionHistoria(p project.Proyecto, h backlog.Historia) formEdicionHistoria {
	return formEdicionHistoria{
		Proyecto: p, Historia: h,
		Titulo: h.Titulo, Descripcion: h.Descripcion, Prioridad: h.Prioridad,
		Prioridades: backlog.Prioridades,
	}
}

func renderEdicionHistoria(w http.ResponseWriter, codigo int, form formEdicionHistoria) {
	render(w, codigo, "historia_editar.html", Pagina{Titulo: "Editar " + form.Historia.Identificador(), Datos: form})
}

// formulario muestra el formulario de editar con los datos actuales de la historia.
func (h edicionHistorias) formulario(w http.ResponseWriter, r *http.Request) {
	proyectoID, numero, ok := idsDeHistoria(w, r)
	if !ok {
		return
	}
	p, historia, err := h.casos.Buscar(r.Context(), proyectoID, numero)
	if respondeError(w, err, "mostrar el formulario de editar") {
		return
	}
	renderEdicionHistoria(w, http.StatusOK, nuevoFormEdicionHistoria(p, historia))
}

// editar recibe el formulario. Si los datos valen vuelve a la página de la historia (redirección 303); si no, muestra el
// formulario con los errores debajo de cada campo y lo que el usuario escribió (422).
func (h edicionHistorias) editar(w http.ResponseWriter, r *http.Request) {
	proyectoID, numero, ok := idsDeHistoria(w, r)
	if !ok {
		return
	}
	cambios := backlog.Cambios{
		Titulo:      r.PostFormValue("titulo"),
		Descripcion: r.PostFormValue("descripcion"),
		Prioridad:   backlog.Prioridad(r.PostFormValue("prioridad")),
	}
	_, err := h.casos.Editar(r.Context(), proyectoID, numero, cambios)

	var errs domain.ErroresValidacion
	switch {
	case err == nil:
		http.Redirect(w, r, fmt.Sprintf("/proyectos/%d/backlog/%d?editada=1", proyectoID, numero), http.StatusSeeOther)
	case errors.As(err, &errs):
		p, historia, errBuscar := h.casos.Buscar(r.Context(), proyectoID, numero)
		if respondeError(w, errBuscar, "volver a mostrar el formulario de editar") {
			return
		}
		form := nuevoFormEdicionHistoria(p, historia)
		form.Titulo, form.Descripcion, form.Prioridad = cambios.Titulo, cambios.Descripcion, cambios.Prioridad
		form.Errores = make(map[string]string, len(errs))
		for campo, e := range errs {
			form.Errores[campo] = e.Error()
		}
		renderEdicionHistoria(w, http.StatusUnprocessableEntity, form)
	default:
		respondeError(w, err, "editar la historia")
	}
}
