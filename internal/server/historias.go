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

// historias agrupa los handlers del Product Backlog (HU-05) y de su orden (HU-09).
type historias struct {
	casos     *service.Historias
	criterios *service.Criterios // para mostrar el progreso de cada historia en la lista (HU-07)
}

// listaHistorias son los datos de la página backlog.html.
type listaHistorias struct {
	Proyecto  project.Proyecto
	Historias []filaBacklog      // las que no están Hechas, en el orden del backlog (HU-09)
	Hechas    []backlog.Historia // las Hechas, la más reciente primero (HU-09 RN4)
	Creada    int                // número de la historia recién creada, para el aviso "Historia HU-N creada"; 0 si no hay
	Eliminada int                // número de la historia recién eliminada, para el aviso "Historia HU-N eliminada"; 0 si no hay (HU-06)
	Movida    string             // aviso de un movimiento hecho, "HU-4 subió un lugar." (HU-09); vacío si no hay
	Error     string             // por qué no se pudo mover una historia (HU-09); vacío si no hay
	// Filtro son los filtros que se están usando (HU-10); Mostradas, cuántas historias los cumplen, y Total, cuántas
	// tiene el proyecto. Estados y Prioridades son las opciones de los desplegables.
	Filtro      backlog.Filtro
	Mostradas   int
	Total       int
	Estados     []backlog.Estado
	Prioridades []backlog.Prioridad
	// Progresos tiene el progreso de los criterios de cada historia, por su ID (HU-07).
	Progresos map[int64]backlog.Progreso
}

// filaBacklog es una historia del backlog con los botones que le corresponden: Subir si no es la primera de su
// prioridad y Bajar si no es la última (HU-09 RN3).
type filaBacklog struct {
	backlog.Historia
	Subir, Bajar bool
}

// filasDelBacklog arma las filas de las historias pendientes, que ya vienen en el orden del backlog. Con reordenable en
// false (hay filtros puestos, HU-10 RN5) ninguna tiene Subir ni Bajar.
func filasDelBacklog(pendientes []backlog.Historia, reordenable bool) []filaBacklog {
	filas := make([]filaBacklog, len(pendientes))
	for i, h := range pendientes {
		filas[i] = filaBacklog{Historia: h,
			Subir: reordenable && i > 0 && pendientes[i-1].Prioridad == h.Prioridad,
			Bajar: reordenable && i < len(pendientes)-1 && pendientes[i+1].Prioridad == h.Prioridad,
		}
	}
	return filas
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

// listar muestra el proyecto y su backlog ordenado, con las Hechas aparte (HU-09), filtrado con los filtros de la URL
// (HU-10).
func (h historias) listar(w http.ResponseWriter, r *http.Request) {
	id, ok := idDeProyecto(w, r)
	if !ok {
		return
	}
	consulta := r.URL.Query()
	creada, _ := strconv.Atoi(consulta.Get("creada"))
	eliminada, _ := strconv.Atoi(consulta.Get("eliminada"))
	datos := listaHistorias{Creada: max(creada, 0), Eliminada: max(eliminada, 0),
		Filtro: backlog.NuevoFiltro(consulta.Get("estado"), consulta.Get("prioridad"), consulta.Get("texto"))}
	if n, _ := strconv.Atoi(consulta.Get("subio")); n > 0 {
		datos.Movida = fmt.Sprintf("HU-%d subió un lugar.", n)
	} else if n, _ := strconv.Atoi(consulta.Get("bajo")); n > 0 {
		datos.Movida = fmt.Sprintf("HU-%d bajó un lugar.", n)
	}
	h.mostrarBacklog(w, r, id, http.StatusOK, datos)
}

// mostrarBacklog completa los datos de la página con el backlog del proyecto, filtrado con datos.Filtro, y la muestra
// con ese código.
func (h historias) mostrarBacklog(w http.ResponseWriter, r *http.Request, id int64, codigo int, datos listaHistorias) {
	v, err := h.casos.Buscar(r.Context(), id, datos.Filtro)
	switch {
	case errors.Is(err, project.ErrNoEncontrado):
		proyectoNoEncontrado(w)
		return
	case err != nil:
		log.Printf("listar historias del proyecto %d: %v", id, err)
		http.Error(w, "No se pudo cargar el backlog. Probá de nuevo en unos minutos.", http.StatusInternalServerError)
		return
	}
	progresos, err := h.criterios.Progresos(r.Context(), append(append([]backlog.Historia{}, v.Pendientes...), v.Hechas...))
	if err != nil {
		log.Printf("progreso de las historias del proyecto %d: %v", id, err)
		http.Error(w, "No se pudo cargar el backlog. Probá de nuevo en unos minutos.", http.StatusInternalServerError)
		return
	}
	datos.Proyecto, datos.Historias, datos.Hechas, datos.Progresos = v.Proyecto, filasDelBacklog(v.Pendientes, !datos.Filtro.Activo()), v.Hechas, progresos
	datos.Mostradas, datos.Total = len(v.Pendientes)+len(v.Hechas), v.Total
	datos.Estados, datos.Prioridades = backlog.Estados, backlog.Prioridades
	render(w, codigo, "backlog.html", Pagina{Titulo: "Backlog", Datos: datos})
}

// mover sube o baja la historia un lugar dentro de su prioridad (HU-09) y vuelve al backlog (redirección 303) con el
// aviso. Si la regla no lo deja, muestra el backlog sin cambios con el motivo (409).
func (h historias) mover(m backlog.Movimiento) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		proyectoID, numero, ok := idsDeHistoria(w, r)
		if !ok {
			return
		}
		_, err := h.casos.Mover(r.Context(), proyectoID, numero, m)

		switch {
		case err == nil:
			aviso := "subio"
			if m == backlog.Bajar {
				aviso = "bajo"
			}
			http.Redirect(w, r, fmt.Sprintf("/proyectos/%d/backlog?%s=%d", proyectoID, aviso, numero), http.StatusSeeOther)
		case errors.Is(err, backlog.ErrYaEsLaPrimera), errors.Is(err, backlog.ErrYaEsLaUltima), errors.Is(err, backlog.ErrHechaNoSeReordena):
			h.mostrarBacklog(w, r, proyectoID, http.StatusConflict, listaHistorias{Error: err.Error()})
		case errors.Is(err, project.ErrNoEncontrado):
			proyectoNoEncontrado(w)
		case errors.Is(err, backlog.ErrHistoriaNoEncontrada):
			http.Error(w, err.Error(), http.StatusNotFound) // "no existe la historia HU-99"
		default:
			log.Printf("mover la historia HU-%d del proyecto %d: %v", numero, proyectoID, err)
			http.Error(w, "No se pudo guardar el orden. Probá de nuevo en unos minutos.", http.StatusInternalServerError)
		}
	}
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
