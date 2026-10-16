package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// estados agrupa el handler del cambio de estado de una historia desde su página (HU-08). La página se arma con el
// caso de uso de criterios, que trae la historia con sus criterios y su progreso.
type estados struct {
	casos *service.Estados
	vista *service.Criterios
}

// avisoDeLaPantalla explica por qué esta pantalla solo pasa a En progreso (RN8 de la spec de HU-08).
const avisoDeLaPantalla = "Desde esta pantalla solo se puede pasar a En progreso. A En sprint se pasa al asignar la historia a un sprint y a Hecho desde el sprint activo."

// cambiar pasa la historia a En progreso. Si el cambio vale vuelve a la página (redirección 303); si no, muestra la
// página con el motivo (409). Un estado que esta pantalla no ofrece es un pedido inválido (400): lo hacen HU-12 y HU-13.
func (h estados) cambiar(w http.ResponseWriter, r *http.Request) {
	proyectoID, numero, ok := idsDeHistoria(w, r)
	if !ok {
		return
	}
	if backlog.Estado(r.PostFormValue("estado")) != backlog.EstadoEnProgreso {
		http.Error(w, avisoDeLaPantalla, http.StatusBadRequest)
		return
	}
	_, err := h.casos.Cambiar(r.Context(), proyectoID, numero, backlog.EstadoEnProgreso)

	switch {
	case err == nil:
		http.Redirect(w, r, fmt.Sprintf("/proyectos/%d/backlog/%d?estado=en-progreso", proyectoID, numero), http.StatusSeeOther)
	case errors.Is(err, backlog.ErrTransicionInvalida), errors.Is(err, backlog.ErrYaEstaEnEseEstado),
		errors.Is(err, backlog.ErrEstadoDesconocido), errors.Is(err, backlog.ErrCriteriosSinCumplir):
		pagina, ok := paginaDespuesDeUnError(w, r, h.vista, proyectoID, numero)
		if !ok {
			return
		}
		pagina.Aviso = err.Error()
		renderHistoria(w, http.StatusConflict, pagina)
	case errors.Is(err, project.ErrNoEncontrado):
		proyectoNoEncontrado(w)
	case errors.Is(err, backlog.ErrHistoriaNoEncontrada):
		historiaNoEncontrada(w)
	default:
		log.Printf("cambiar el estado de la historia HU-%d del proyecto %d: %v", numero, proyectoID, err)
		http.Error(w, "No se pudo cambiar el estado. Probá de nuevo en unos minutos.", http.StatusInternalServerError)
	}
}
