package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// historias agrupa los handlers del Product Backlog (HU-05).
type historias struct {
	casos     *service.Historias
	criterios *service.Criterios // para mostrar el progreso de cada historia en la lista (HU-07)
}

// listaHistorias son los datos de la página backlog.html.
type listaHistorias struct {
	Proyecto  project.Proyecto
	Historias []backlog.Historia
	Creada    int // número de la historia recién creada, para el aviso "Historia HU-N creada"; 0 si no hay
	// Progresos tiene el progreso de los criterios de cada historia, por su ID (HU-07).
	Progresos map[int64]backlog.Progreso
}

// formHistoria son los datos de la página historia_nueva.html: el proyecto (para el título y los enlaces), lo que
// cargó el usuario como texto, las opciones de los desplegables y el error de cada campo.
type formHistoria struct {
	Proyecto     project.Proyecto
	Titulo       string
	Descripcion  string
	Prioridad    backlog.Prioridad
	StoryPoints  string
	Prioridades  []backlog.Prioridad
	Escala       []int
	Errores      map[string]string
	ErrorGeneral string
}

// formVacio devuelve el formulario sin datos cargados, con las opciones de los desplegables.
func formVacio(p project.Proyecto) formHistoria {
	return formHistoria{Proyecto: p, Prioridades: backlog.Prioridades, Escala: backlog.EscalaStoryPoints}
}

// listar muestra el proyecto y las historias de su backlog.
func (h historias) listar(w http.ResponseWriter, r *http.Request) {
	p, lista, ok := h.cargar(w, r)
	if !ok {
		return
	}
	progresos, err := h.criterios.Progresos(r.Context(), lista)
	if err != nil {
		log.Printf("progreso de las historias del proyecto %d: %v", p.ID, err)
		http.Error(w, "No se pudo cargar el backlog. Probá de nuevo en unos minutos.", http.StatusInternalServerError)
		return
	}
	creada, _ := strconv.Atoi(r.URL.Query().Get("creada"))
	render(w, http.StatusOK, "backlog.html", Pagina{
		Titulo: "Backlog",
		Datos:  listaHistorias{Proyecto: p, Historias: lista, Creada: max(creada, 0), Progresos: progresos},
	})
}

// nueva muestra el formulario vacío.
func (h historias) nueva(w http.ResponseWriter, r *http.Request) {
	p, _, ok := h.cargar(w, r)
	if !ok {
		return
	}
	render(w, http.StatusOK, "historia_nueva.html", Pagina{Titulo: "Nueva historia", Datos: formVacio(p)})
}

// crear recibe el formulario. Si los datos son válidos vuelve a la lista (redirección 303); si no, muestra el
// formulario con los errores y los datos cargados.
func (h historias) crear(w http.ResponseWriter, r *http.Request) {
	id, ok := idDeProyecto(w, r)
	if !ok {
		return
	}
	form := formVacio(project.Proyecto{})
	form.Titulo = r.PostFormValue("titulo")
	form.Descripcion = r.PostFormValue("descripcion")
	form.Prioridad = backlog.Prioridad(r.PostFormValue("prioridad"))
	form.StoryPoints = r.PostFormValue("story_points")
	creada, err := h.casos.Crear(r.Context(), id, backlog.Datos{
		Titulo:      form.Titulo,
		Descripcion: form.Descripcion,
		Prioridad:   form.Prioridad,
		StoryPoints: form.StoryPoints,
	})

	var errs domain.ErroresValidacion
	switch {
	case err == nil:
		http.Redirect(w, r, fmt.Sprintf("/proyectos/%d/backlog?creada=%d", id, creada.Numero), http.StatusSeeOther)
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
		render(w, http.StatusUnprocessableEntity, "historia_nueva.html", Pagina{Titulo: "Nueva historia", Datos: form})
	default:
		log.Printf("crear historia del proyecto %d: %v", id, err)
		form.Proyecto.ID = id
		form.ErrorGeneral = "No se pudo guardar la historia. Probá de nuevo en unos minutos."
		render(w, http.StatusInternalServerError, "historia_nueva.html", Pagina{Titulo: "Nueva historia", Datos: form})
	}
}

// cargar busca el proyecto de la URL y sus historias. Si no puede, ya respondió (404 o 500) y ok es false.
func (h historias) cargar(w http.ResponseWriter, r *http.Request) (p project.Proyecto, lista []backlog.Historia, ok bool) {
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
		log.Printf("listar historias del proyecto %d: %v", id, err)
		http.Error(w, "No se pudo cargar el backlog. Probá de nuevo en unos minutos.", http.StatusInternalServerError)
		return p, nil, false
	}
	return p, lista, true
}
