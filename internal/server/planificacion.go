package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// planificacion agrupa los handlers de la página de un sprint: ver su Sprint Backlog, asignarle historias y
// quitárselas (HU-12), y empezar y completar sus historias (HU-13).
type planificacion struct {
	casos *service.Asignaciones
}

// paginaSprint son los datos de la página sprint.html: el sprint con sus historias, el mensaje de éxito después de
// asignar o quitar y el motivo si algo no se pudo.
type paginaSprint struct {
	service.VistaSprint
	Mensaje string
	Error   string
}

// ver muestra el sprint con sus historias asignadas y las candidatas.
func (h planificacion) ver(w http.ResponseWriter, r *http.Request) {
	proyectoID, numeroSprint, ok := idsDeSprint(w, r)
	if !ok {
		return
	}
	vista, err := h.casos.Ver(r.Context(), proyectoID, numeroSprint)
	if respondeErrorDeSprint(w, err, "ver el sprint") {
		return
	}
	mensaje := mensajeDelSprint(r.URL.Query(), numeroSprint)
	if mensaje == "" {
		mensaje = mensajeDeCierre(r.URL.Query(), vista)
	}
	renderSprint(w, http.StatusOK, paginaSprint{VistaSprint: vista, Mensaje: mensaje})
}

// asignar recibe el botón Asignar. Si se puede, vuelve a la página del sprint (redirección 303); si no, la muestra con
// el motivo (409).
func (h planificacion) asignar(w http.ResponseWriter, r *http.Request) {
	proyectoID, numeroSprint, ok := idsDeSprint(w, r)
	if !ok {
		return
	}
	numero, err := strconv.Atoi(r.PostFormValue("historia"))
	if err != nil {
		historiaNoEncontrada(w)
		return
	}
	_, err = h.casos.Asignar(r.Context(), proyectoID, numeroSprint, numero)
	h.responder(w, r, err, proyectoID, numeroSprint, fmt.Sprintf("asignada=%d", numero))
}

// accionSobreHistoria es un caso de uso sobre una historia del sprint: Quitar (HU-12), Empezar o Completar (HU-13).
type accionSobreHistoria func(casos *service.Asignaciones, ctx context.Context, proyectoID int64, numeroSprint, numero int) (backlog.Historia, error)

// accion devuelve el handler del botón de una historia del sprint. Si la acción vale vuelve a la página del sprint
// con el parámetro de éxito (por ejemplo ?quitada=1); si no, la muestra con el motivo.
func (h planificacion) accion(hacer accionSobreHistoria, exito string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		proyectoID, numeroSprint, ok := idsDeSprint(w, r)
		if !ok {
			return
		}
		numero, err := strconv.Atoi(chi.URLParam(r, "historia"))
		if err != nil {
			historiaNoEncontrada(w)
			return
		}
		_, err = hacer(h.casos, r.Context(), proyectoID, numeroSprint, numero)
		h.responder(w, r, err, proyectoID, numeroSprint, fmt.Sprintf("%s=%d", exito, numero))
	}
}

// responder vuelve a la página del sprint con el aviso de éxito, o la muestra con el motivo de la regla (409).
func (h planificacion) responder(w http.ResponseWriter, r *http.Request, err error, proyectoID int64, numeroSprint int, exito string) {
	if err == nil {
		http.Redirect(w, r, fmt.Sprintf("/proyectos/%d/sprints/%d?%s", proyectoID, numeroSprint, exito), http.StatusSeeOther)
		return
	}
	if !esErrorDeLaRegla(err) {
		respondeErrorDeSprint(w, err, "asignar o quitar una historia del sprint")
		return
	}
	vista, errVista := h.casos.Ver(r.Context(), proyectoID, numeroSprint)
	if respondeErrorDeSprint(w, errVista, "volver a mostrar el sprint") {
		return
	}
	renderSprint(w, http.StatusConflict, paginaSprint{VistaSprint: vista, Error: err.Error()})
}

// esErrorDeLaRegla dice si err es un motivo de la regla de asignar o quitar (o de las transiciones de HU-08), que se
// le muestra al usuario tal cual.
func esErrorDeLaRegla(err error) bool {
	for _, regla := range []error{
		backlog.ErrSinEstimar, backlog.ErrSinCriterios, backlog.ErrYaHecha, backlog.ErrEnOtroSprint,
		backlog.ErrSprintCerrado, backlog.ErrYaEmpezada, backlog.ErrNoEstaEnUnSprint, backlog.ErrNoEsElSprintActivo,
		backlog.ErrTransicionInvalida, backlog.ErrYaEstaEnEseEstado, backlog.ErrCriteriosSinCumplir,
	} {
		if errors.Is(err, regla) {
			return true
		}
	}
	return false
}

// mensajeDelSprint arma el aviso de éxito con los parámetros de la redirección. Solo acepta un número de historia:
// cualquier otro valor se ignora, así la página nunca muestra lo que venga en la URL.
func mensajeDelSprint(consulta url.Values, numeroSprint int) string {
	for _, m := range []struct{ parametro, formato string }{
		{"asignada", "HU-%d asignada al Sprint %d"},
		{"quitada", "HU-%d quitada del Sprint %d"},
		{"empezada", "HU-%d pasó a En progreso"},
		{"completada", "HU-%d completada"},
	} {
		if numero, err := strconv.Atoi(consulta.Get(m.parametro)); err == nil && numero > 0 {
			if strings.Count(m.formato, "%d") == 1 {
				return fmt.Sprintf(m.formato, numero)
			}
			return fmt.Sprintf(m.formato, numero, numeroSprint)
		}
	}
	return ""
}

// respondeErrorDeSprint responde 404 (proyecto, sprint o historia que no existen) o 500 si err no es nil, y dice si ya
// respondió.
func respondeErrorDeSprint(w http.ResponseWriter, err error, accion string) bool {
	if errors.Is(err, sprint.ErrNoEncontrado) {
		http.Error(w, "No encontramos ese sprint.", http.StatusNotFound)
		return true
	}
	return respondeError(w, err, accion)
}

func renderSprint(w http.ResponseWriter, codigo int, pagina paginaSprint) {
	render(w, codigo, "sprint.html", Pagina{Titulo: fmt.Sprintf("Sprint %d", pagina.Sprint.Numero), Datos: pagina})
}

// idsDeSprint lee el {id} del proyecto y el {sprint} de la URL. Si alguno no es un número responde 404.
func idsDeSprint(w http.ResponseWriter, r *http.Request) (proyectoID int64, numeroSprint int, ok bool) {
	proyectoID, ok = idDeProyecto(w, r)
	if !ok {
		return 0, 0, false
	}
	numeroSprint, err := strconv.Atoi(chi.URLParam(r, "sprint"))
	if err != nil {
		http.Error(w, "No encontramos ese sprint.", http.StatusNotFound)
		return 0, 0, false
	}
	return proyectoID, numeroSprint, true
}
