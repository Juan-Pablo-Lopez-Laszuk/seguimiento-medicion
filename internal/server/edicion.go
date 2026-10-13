package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// edicion agrupa los handlers de modificar un proyecto (HU-02).
type edicion struct {
	casos *service.EdicionProyectos
}

// formEdicion son los datos de la página proyecto_editar.html: el proyecto, lo que cargó el usuario como texto y el
// error de cada campo. Finalizado oculta el formulario y muestra el aviso (CA-02.3).
type formEdicion struct {
	ID           int64
	Nombre       string
	Descripcion  string
	FechaInicio  string
	FechaFin     string
	Errores      map[string]string
	ErrorGeneral string
	Finalizado   bool
}

const tituloEdicion = "Modificar proyecto"

// formulario muestra el formulario con los datos actuales del proyecto, o el aviso si está Finalizado.
func (h edicion) formulario(w http.ResponseWriter, r *http.Request) {
	id, ok := idDeProyecto(w, r)
	if !ok {
		return
	}
	p, err := h.casos.Preparar(r.Context(), id)
	switch {
	case err == nil:
		render(w, http.StatusOK, "proyecto_editar.html", Pagina{Titulo: tituloEdicion, Datos: formEdicion{
			ID:          p.ID,
			Nombre:      p.Nombre,
			Descripcion: p.Descripcion,
			FechaInicio: fechaParaInput(p.FechaInicio),
			FechaFin:    fechaParaInput(p.FechaFin),
		}})
	case errors.Is(err, project.ErrNoEncontrado):
		proyectoNoEncontrado(w)
	case errors.Is(err, project.ErrProyectoFinalizado):
		render(w, http.StatusConflict, "proyecto_editar.html", Pagina{Titulo: tituloEdicion, Datos: formEdicion{ID: id, Finalizado: true}})
	default:
		log.Printf("preparar la modificación del proyecto %d: %v", id, err)
		http.Error(w, "No se pudo cargar el proyecto. Probá de nuevo en unos minutos.", http.StatusInternalServerError)
	}
}

// modificar recibe el formulario. Si los datos son válidos vuelve a la vista del proyecto con el mensaje
// (redirección 303); si no, muestra el formulario con los errores y los datos cargados.
func (h edicion) modificar(w http.ResponseWriter, r *http.Request) {
	id, ok := idDeProyecto(w, r)
	if !ok {
		return
	}
	form := formEdicion{
		ID:          id,
		Nombre:      r.PostFormValue("nombre"),
		Descripcion: r.PostFormValue("descripcion"),
		FechaInicio: r.PostFormValue("fecha_inicio"),
		FechaFin:    r.PostFormValue("fecha_fin"),
	}
	_, err := h.casos.Modificar(r.Context(), id, project.Datos{
		Nombre:      form.Nombre,
		Descripcion: form.Descripcion,
		FechaInicio: fechaDeFormulario(form.FechaInicio),
		FechaFin:    fechaDeFormulario(form.FechaFin),
	})

	var errs domain.ErroresValidacion
	switch {
	case err == nil:
		http.Redirect(w, r, fmt.Sprintf("/proyectos/%d?modificado=1", id), http.StatusSeeOther)
	case errors.Is(err, project.ErrNoEncontrado):
		proyectoNoEncontrado(w)
	case errors.Is(err, project.ErrProyectoFinalizado):
		render(w, http.StatusConflict, "proyecto_editar.html", Pagina{Titulo: tituloEdicion, Datos: formEdicion{ID: id, Finalizado: true}})
	case errors.As(err, &errs):
		form.Errores = make(map[string]string, len(errs))
		for campo, e := range errs {
			form.Errores[campo] = e.Error()
		}
		render(w, http.StatusUnprocessableEntity, "proyecto_editar.html", Pagina{Titulo: tituloEdicion, Datos: form})
	default:
		log.Printf("modificar el proyecto %d: %v", id, err)
		form.ErrorGeneral = "No se pudo guardar el proyecto. Probá de nuevo en unos minutos."
		render(w, http.StatusInternalServerError, "proyecto_editar.html", Pagina{Titulo: tituloEdicion, Datos: form})
	}
}
