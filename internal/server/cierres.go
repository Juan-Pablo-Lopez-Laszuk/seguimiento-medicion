package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// cierres agrupa los handlers de cerrar el sprint activo (HU-14): la pantalla de confirmación y el cierre.
type cierres struct {
	casos *service.Cierres
}

// paginaCierre son los datos de sprint_cerrar.html: lo que va a pasar al cerrar, o el motivo si no se puede.
type paginaCierre struct {
	service.VistaCierre
	Error string
}

// confirmar muestra qué va a pasar al cerrar el sprint, sin cambiar nada. Si no se puede cerrar, muestra el motivo
// sin el botón (409).
func (h cierres) confirmar(w http.ResponseWriter, r *http.Request) {
	proyectoID, numeroSprint, ok := idsDeSprint(w, r)
	if !ok {
		return
	}
	vista, err := h.casos.Preparar(r.Context(), proyectoID, numeroSprint)
	h.mostrar(w, vista, err, http.StatusOK)
}

// cerrar recibe el botón Cerrar sprint de la confirmación. Si se puede, vuelve a la página del sprint con el aviso
// (redirección 303); si no, muestra la confirmación con el motivo (409).
func (h cierres) cerrar(w http.ResponseWriter, r *http.Request) {
	proyectoID, numeroSprint, ok := idsDeSprint(w, r)
	if !ok {
		return
	}
	cierre, err := h.casos.Cerrar(r.Context(), proyectoID, numeroSprint)
	if err == nil {
		http.Redirect(w, r, fmt.Sprintf("/proyectos/%d/sprints/%d?cerrado=1&devueltas=%d", proyectoID, numeroSprint, len(cierre.Devueltas)), http.StatusSeeOther)
		return
	}
	if !esErrorDeCierre(err) {
		respondeErrorDeSprint(w, err, "cerrar el sprint")
		return
	}
	vista, errVista := h.casos.Preparar(r.Context(), proyectoID, numeroSprint)
	if errVista != nil && !esErrorDeCierre(errVista) {
		respondeErrorDeSprint(w, errVista, "volver a mostrar el cierre")
		return
	}
	renderCierre(w, http.StatusConflict, paginaCierre{VistaCierre: vista, Error: err.Error()})
}

// mostrar renderiza la confirmación con el código dado, o con el motivo (409) si el error es de la regla.
func (h cierres) mostrar(w http.ResponseWriter, vista service.VistaCierre, err error, codigo int) {
	switch {
	case esErrorDeCierre(err):
		renderCierre(w, http.StatusConflict, paginaCierre{VistaCierre: vista, Error: err.Error()})
	case respondeErrorDeSprint(w, err, "preparar el cierre del sprint"):
	default:
		renderCierre(w, codigo, paginaCierre{VistaCierre: vista})
	}
}

// esErrorDeCierre dice si err es un motivo de la regla de cierre, que se le muestra al usuario tal cual.
func esErrorDeCierre(err error) bool {
	return errors.Is(err, sprint.ErrYaCerrado) || errors.Is(err, sprint.ErrNoEsElActivo)
}

func renderCierre(w http.ResponseWriter, codigo int, pagina paginaCierre) {
	render(w, codigo, "sprint_cerrar.html", Pagina{Titulo: fmt.Sprintf("Cerrar el Sprint %d", pagina.Sprint.Numero), Datos: pagina})
}

// mensajeDeCierre arma el aviso después de cerrar, con los SP de la foto (HU-14). Solo acepta números en la URL:
// cualquier otro valor se ignora, así la página nunca muestra lo que venga en ella.
func mensajeDeCierre(consulta url.Values, v service.VistaSprint) string {
	if consulta.Get("cerrado") != "1" || v.Abierto {
		return ""
	}
	mensaje := fmt.Sprintf("Sprint %d cerrado: %d de %d SP completados.", v.Sprint.Numero, v.Completados, v.Comprometidos)
	switch devueltas, err := strconv.Atoi(consulta.Get("devueltas")); {
	case err != nil || devueltas <= 0:
	case devueltas == 1:
		mensaje += " 1 historia volvió al backlog."
	default:
		mensaje += fmt.Sprintf(" %d historias volvieron al backlog.", devueltas)
	}
	return mensaje
}
