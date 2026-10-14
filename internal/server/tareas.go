package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// tareas agrupa los handlers de las tareas de una historia (HU-25).
type tareas struct {
	casos *service.Tareas
}

// paginaTareas son los datos de la página tareas.html: la historia con sus tareas y la suma de sus horas, y lo que
// cambia según el pedido (el mensaje de éxito, el aviso, el error y lo escrito en el formulario de tarea nueva).
type paginaTareas struct {
	service.VistaTareas
	Mensaje string            // mensaje de éxito después de agregar, cambiar horas o marcar
	Aviso   string            // por ejemplo, que la tarea está completada y no cambia su estimación
	Error   string            // un error al cambiar las horas de una tarea de la lista
	Nueva   effort.DatosTarea // lo que el usuario escribió en el formulario de tarea nueva
	Errores map[string]string
}

// mensajeDeTareas traduce los parámetros de la redirección a su texto. Solo conoce los valores que el servidor mismo
// pone: cualquier otro se ignora, así la página nunca muestra lo que venga en la URL.
func mensajeDeTareas(consulta url.Values) string {
	switch {
	case consulta.Get("agregada") == "1":
		return "Tarea agregada."
	case consulta.Get("horas") == "1":
		return "Horas estimadas actualizadas."
	case consulta.Get("marcada") == "completada":
		return "Tarea marcada como completada."
	case consulta.Get("marcada") == "pendiente":
		return "Tarea marcada como pendiente."
	}
	return ""
}

// ver muestra las tareas de la historia, la suma de sus horas y el formulario para agregar una.
func (h tareas) ver(w http.ResponseWriter, r *http.Request) {
	proyectoID, numero, ok := idsDeHistoria(w, r)
	if !ok {
		return
	}
	vista, err := h.casos.Ver(r.Context(), proyectoID, numero)
	if respondeErrorDeTareas(w, err, "ver las tareas") {
		return
	}
	renderTareas(w, http.StatusOK, paginaTareas{VistaTareas: vista, Mensaje: mensajeDeTareas(r.URL.Query())})
}

// agregar recibe el formulario. Si los datos son válidos vuelve a la página (redirección 303); si no, muestra la
// página con cada error debajo de su campo y lo que se escribió.
func (h tareas) agregar(w http.ResponseWriter, r *http.Request) {
	proyectoID, numero, ok := idsDeHistoria(w, r)
	if !ok {
		return
	}
	// Un responsable que no es un número es como no haber elegido ninguno.
	responsable, _ := strconv.ParseInt(r.PostFormValue("responsable"), 10, 64)
	d := effort.DatosTarea{Titulo: r.PostFormValue("titulo"), ResponsableID: responsable, Horas: r.PostFormValue("horas")}
	_, err := h.casos.Agregar(r.Context(), proyectoID, numero, d)

	var errs domain.ErroresValidacion
	switch {
	case err == nil:
		http.Redirect(w, r, fmt.Sprintf("/proyectos/%d/backlog/%d/tareas?agregada=1", proyectoID, numero), http.StatusSeeOther)
	case errors.As(err, &errs):
		pagina, ok := h.paginaDespuesDeUnError(w, r, proyectoID, numero)
		if !ok {
			return
		}
		pagina.Nueva = d
		pagina.Errores = make(map[string]string, len(errs))
		for campo, e := range errs {
			pagina.Errores[campo] = e.Error()
		}
		renderTareas(w, http.StatusUnprocessableEntity, pagina)
	default:
		respondeErrorDeTareas(w, err, "agregar una tarea")
	}
}

// cambiarHoras cambia las horas estimadas de una tarea pendiente y vuelve a la página (CA-25.3).
func (h tareas) cambiarHoras(w http.ResponseWriter, r *http.Request) {
	proyectoID, numero, tareaID, ok := idsDeTarea(w, r)
	if !ok {
		return
	}
	_, err := h.casos.CambiarEstimacion(r.Context(), proyectoID, numero, tareaID, r.PostFormValue("horas"))

	switch {
	case err == nil:
		http.Redirect(w, r, fmt.Sprintf("/proyectos/%d/backlog/%d/tareas?horas=1", proyectoID, numero), http.StatusSeeOther)
	case errors.Is(err, effort.ErrTareaCompletada):
		pagina, ok := h.paginaDespuesDeUnError(w, r, proyectoID, numero)
		if !ok {
			return
		}
		pagina.Aviso = err.Error()
		renderTareas(w, http.StatusConflict, pagina)
	case errors.Is(err, effort.ErrHorasInvalidas):
		pagina, ok := h.paginaDespuesDeUnError(w, r, proyectoID, numero)
		if !ok {
			return
		}
		pagina.Error = fmt.Sprintf("%s: %s", tituloDeTarea(pagina.Tareas, tareaID), err)
		renderTareas(w, http.StatusUnprocessableEntity, pagina)
	default:
		respondeErrorDeTareas(w, err, "cambiar las horas de una tarea")
	}
}

// marcar deja la tarea completada (completada=1) o pendiente (completada=0) y vuelve a la página (RN7).
func (h tareas) marcar(w http.ResponseWriter, r *http.Request) {
	proyectoID, numero, tareaID, ok := idsDeTarea(w, r)
	if !ok {
		return
	}
	valor := r.PostFormValue("completada")
	if valor != "1" && valor != "0" {
		http.Error(w, "Valor inválido.", http.StatusBadRequest)
		return
	}
	completada := valor == "1"
	if _, err := h.casos.Marcar(r.Context(), proyectoID, numero, tareaID, completada); respondeErrorDeTareas(w, err, "marcar una tarea") {
		return
	}
	resultado := "pendiente"
	if completada {
		resultado = "completada"
	}
	http.Redirect(w, r, fmt.Sprintf("/proyectos/%d/backlog/%d/tareas?marcada=%s", proyectoID, numero, resultado), http.StatusSeeOther)
}

// paginaDespuesDeUnError vuelve a buscar las tareas para mostrarlas junto al error. Si no puede, ya respondió.
func (h tareas) paginaDespuesDeUnError(w http.ResponseWriter, r *http.Request, proyectoID int64, numero int) (paginaTareas, bool) {
	vista, err := h.casos.Ver(r.Context(), proyectoID, numero)
	if respondeErrorDeTareas(w, err, "volver a mostrar las tareas") {
		return paginaTareas{}, false
	}
	return paginaTareas{VistaTareas: vista}, true
}

// tituloDeTarea devuelve el título de la tarea con ese ID, para decir de cuál es un error.
func tituloDeTarea(lista []effort.Tarea, id int64) string {
	for _, t := range lista {
		if t.ID == id {
			return t.Titulo
		}
	}
	return "La tarea"
}

// respondeErrorDeTareas responde 404 (proyecto, historia o tarea que no existen) o 500 si err no es nil, y dice si ya
// respondió. Con err nil no hace nada y devuelve false.
func respondeErrorDeTareas(w http.ResponseWriter, err error, accion string) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, project.ErrNoEncontrado):
		proyectoNoEncontrado(w)
	case errors.Is(err, backlog.ErrHistoriaNoEncontrada):
		historiaNoEncontrada(w)
	case errors.Is(err, effort.ErrTareaNoEncontrada):
		tareaNoEncontrada(w)
	default:
		log.Printf("%s: %v", accion, err)
		http.Error(w, "No se pudo completar la operación. Probá de nuevo en unos minutos.", http.StatusInternalServerError)
	}
	return true
}

func renderTareas(w http.ResponseWriter, codigo int, pagina paginaTareas) {
	render(w, codigo, "tareas.html", Pagina{Titulo: "Tareas de " + pagina.Historia.Identificador(), Datos: pagina})
}

// idsDeTarea lee el proyecto, la historia y el {tareaID} de la URL. Si la tarea no es un número responde 404, igual
// que si no existiera.
func idsDeTarea(w http.ResponseWriter, r *http.Request) (proyectoID int64, numero int, tareaID int64, ok bool) {
	proyectoID, numero, ok = idsDeHistoria(w, r)
	if !ok {
		return 0, 0, 0, false
	}
	tareaID, err := strconv.ParseInt(chi.URLParam(r, "tareaID"), 10, 64)
	if err != nil {
		tareaNoEncontrada(w)
		return 0, 0, 0, false
	}
	return proyectoID, numero, tareaID, true
}

func tareaNoEncontrada(w http.ResponseWriter) {
	http.Error(w, "No encontramos esa tarea.", http.StatusNotFound)
}
