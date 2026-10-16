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

// pantallaEliminar son los datos de la página historia_eliminar.html: la historia, lo que se elimina con ella (ya dicho
// en singular o plural) y, si no se puede eliminar, el motivo.
type pantallaEliminar struct {
	Proyecto    project.Proyecto
	Historia    backlog.Historia
	Criterios   string // por ejemplo "2 criterios de aceptación"
	Tareas      string // por ejemplo "1 tarea"
	Impedimento string // vacío si se puede eliminar
}

func nuevaPantallaEliminar(v service.VistaEliminacion) pantallaEliminar {
	p := pantallaEliminar{
		Proyecto:  v.Proyecto,
		Historia:  v.Historia,
		Criterios: cantidad(v.Criterios, "criterio de aceptación", "criterios de aceptación"),
		Tareas:    cantidad(v.Tareas, "tarea", "tareas"),
	}
	if v.Impedimento != nil {
		p.Impedimento = v.Impedimento.Error()
	}
	return p
}

// cantidad dice "1 tarea", "2 tareas" o "0 tareas".
func cantidad(n int, singular, plural string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", singular)
	}
	return fmt.Sprintf("%d %s", n, plural)
}

func renderEliminar(w http.ResponseWriter, codigo int, pantalla pantallaEliminar) {
	render(w, codigo, "historia_eliminar.html", Pagina{Titulo: "Eliminar " + pantalla.Historia.Identificador(), Datos: pantalla})
}

// confirmar muestra la pantalla que pide confirmar la eliminación (CA-06.3): qué se elimina con la historia y, si no se
// puede eliminar, por qué. Mostrarla no elimina nada.
func (h edicionHistorias) confirmar(w http.ResponseWriter, r *http.Request) {
	proyectoID, numero, ok := idsDeHistoria(w, r)
	if !ok {
		return
	}
	vista, err := h.casos.PedirEliminacion(r.Context(), proyectoID, numero)
	if respondeError(w, err, "mostrar la confirmación de eliminar") {
		return
	}
	renderEliminar(w, http.StatusOK, nuevaPantallaEliminar(vista))
}

// eliminar recibe la confirmación. Si se puede, elimina la historia con sus criterios y sus tareas y vuelve al backlog
// (redirección 303); si no, muestra la pantalla con el motivo (409) y no borra nada.
func (h edicionHistorias) eliminar(w http.ResponseWriter, r *http.Request) {
	proyectoID, numero, ok := idsDeHistoria(w, r)
	if !ok {
		return
	}
	eliminada, err := h.casos.Eliminar(r.Context(), proyectoID, numero)

	switch {
	case err == nil:
		http.Redirect(w, r, fmt.Sprintf("/proyectos/%d/backlog?eliminada=%d", proyectoID, eliminada.Numero), http.StatusSeeOther)
	case errors.Is(err, backlog.ErrEliminarHistoriaHecha),
		errors.Is(err, backlog.ErrEliminarEnSprintCerrado),
		errors.Is(err, backlog.ErrHistoriaConRegistros):
		vista, errVista := h.casos.PedirEliminacion(r.Context(), proyectoID, numero)
		if respondeError(w, errVista, "volver a mostrar la confirmación de eliminar") {
			return
		}
		pantalla := nuevaPantallaEliminar(vista)
		pantalla.Impedimento = err.Error()
		renderEliminar(w, http.StatusConflict, pantalla)
	default:
		respondeError(w, err, "eliminar la historia")
	}
}
