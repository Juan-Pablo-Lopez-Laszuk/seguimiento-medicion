package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// sprints agrupa los handlers de la épica E3 (HU-11).
type sprints struct {
	casos *service.Sprints
}

// listaSprints son los datos de la página sprints.html.
type listaSprints struct {
	Proyecto project.Proyecto
	Sprints  []sprint.Sprint
	Creado   string // número del sprint recién creado, para el aviso "Sprint N creado"
}

// formSprint son los datos de la página sprint_nuevo.html: el proyecto (para el título y el rango de
// fechas), lo que cargó el usuario como texto y el error de cada campo.
type formSprint struct {
	Proyecto     project.Proyecto
	Objetivo     string
	FechaInicio  string
	FechaFin     string
	Errores      map[string]string
	ErrorGeneral string
}

// listar muestra el proyecto y sus sprints.
func (h sprints) listar(w http.ResponseWriter, r *http.Request) {
	p, lista, ok := h.cargar(w, r)
	if !ok {
		return
	}
	render(w, http.StatusOK, "sprints.html", Pagina{
		Titulo: "Sprints",
		Datos:  listaSprints{Proyecto: p, Sprints: lista, Creado: r.URL.Query().Get("creado")},
	})
}

// nuevo muestra el formulario vacío.
func (h sprints) nuevo(w http.ResponseWriter, r *http.Request) {
	p, _, ok := h.cargar(w, r)
	if !ok {
		return
	}
	render(w, http.StatusOK, "sprint_nuevo.html", Pagina{Titulo: "Nuevo sprint", Datos: formSprint{Proyecto: p}})
}

// crear recibe el formulario. Si los datos son válidos vuelve a la lista (redirección 303); si no,
// muestra el formulario con los errores y los datos cargados.
func (h sprints) crear(w http.ResponseWriter, r *http.Request) {
	id, ok := idDeProyecto(w, r)
	if !ok {
		return
	}
	form := formSprint{
		Objetivo:    r.PostFormValue("objetivo"),
		FechaInicio: r.PostFormValue("fecha_inicio"),
		FechaFin:    r.PostFormValue("fecha_fin"),
	}
	creado, err := h.casos.Crear(r.Context(), id, sprint.Datos{
		Objetivo:    form.Objetivo,
		FechaInicio: fechaDeFormulario(form.FechaInicio),
		FechaFin:    fechaDeFormulario(form.FechaFin),
	})

	var errs domain.ErroresValidacion
	switch {
	case err == nil:
		http.Redirect(w, r, fmt.Sprintf("/proyectos/%d/sprints?creado=%d", id, creado.Numero), http.StatusSeeOther)
		return
	case errors.Is(err, project.ErrNoEncontrado):
		proyectoNoEncontrado(w)
		return
	case errors.As(err, &errs):
		form.Errores = make(map[string]string, len(errs))
		for campo, e := range errs {
			form.Errores[campo] = e.Error()
		}
		// El proyecto se vuelve a buscar solo para mostrar el título y el rango en el formulario.
		form.Proyecto, _, _ = h.casos.Listar(r.Context(), id)
		render(w, http.StatusUnprocessableEntity, "sprint_nuevo.html", Pagina{Titulo: "Nuevo sprint", Datos: form})
	default:
		log.Printf("crear sprint del proyecto %d: %v", id, err)
		form.ErrorGeneral = "No se pudo guardar el sprint. Probá de nuevo en unos minutos."
		render(w, http.StatusInternalServerError, "sprint_nuevo.html", Pagina{Titulo: "Nuevo sprint", Datos: form})
	}
}

// cargar busca el proyecto de la URL y sus sprints. Si no puede, ya respondió (404 o 500) y ok es false.
func (h sprints) cargar(w http.ResponseWriter, r *http.Request) (p project.Proyecto, lista []sprint.Sprint, ok bool) {
	id, ok := idDeProyecto(w, r)
	if !ok {
		return p, nil, false
	}
	p, lista, err := h.casos.Listar(r.Context(), id)
	switch {
	case errors.Is(err, project.ErrNoEncontrado):
		proyectoNoEncontrado(w)
		return p, nil, false
	case err != nil:
		log.Printf("listar sprints del proyecto %d: %v", id, err)
		http.Error(w, "No se pudieron cargar los sprints. Probá de nuevo en unos minutos.", http.StatusInternalServerError)
		return p, nil, false
	}
	return p, lista, true
}

// idDeProyecto lee el {id} de la URL. Si no es un número responde 404, igual que un proyecto que no existe.
func idDeProyecto(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		proyectoNoEncontrado(w)
		return 0, false
	}
	return id, true
}

func proyectoNoEncontrado(w http.ResponseWriter) {
	http.Error(w, "No encontramos ese proyecto.", http.StatusNotFound)
}

// fechaParaInput formatea una fecha como la espera un <input type="date"> (AAAA-MM-DD).
func fechaParaInput(t time.Time) string {
	return t.Format(time.DateOnly)
}
