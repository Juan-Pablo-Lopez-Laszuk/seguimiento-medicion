package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// estimaciones agrupa el handler de estimar una historia desde su página (HU-16). La página se arma con el caso de uso
// de criterios, que trae la historia con sus criterios y su progreso.
type estimaciones struct {
	casos *service.Estimaciones
	vista *service.Criterios
}

// estimar guarda los story points de la historia. Si el valor vale vuelve a la página (redirección 303); si no, muestra
// la página con el error debajo del campo (422); y si la historia está Hecha, con el aviso (409).
func (h estimaciones) estimar(w http.ResponseWriter, r *http.Request) {
	proyectoID, numero, ok := idsDeHistoria(w, r)
	if !ok {
		return
	}
	_, err := h.casos.Estimar(r.Context(), proyectoID, numero, r.PostFormValue("story_points"))

	var errs domain.ErroresValidacion
	switch {
	case err == nil:
		http.Redirect(w, r, fmt.Sprintf("/proyectos/%d/backlog/%d?estimada=1", proyectoID, numero), http.StatusSeeOther)
	case errors.As(err, &errs):
		pagina, ok := h.paginaDespuesDeUnError(w, r, proyectoID, numero)
		if !ok {
			return
		}
		pagina.Errores = make(map[string]string, len(errs))
		for campo, e := range errs {
			pagina.Errores[campo] = e.Error()
		}
		renderHistoria(w, http.StatusUnprocessableEntity, pagina)
	case errors.Is(err, backlog.ErrEstimacionDeHistoriaHecha):
		pagina, ok := h.paginaDespuesDeUnError(w, r, proyectoID, numero)
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
		log.Printf("estimar la historia HU-%d del proyecto %d: %v", numero, proyectoID, err)
		http.Error(w, "No se pudo guardar la estimación. Probá de nuevo en unos minutos.", http.StatusInternalServerError)
	}
}

// paginaDespuesDeUnError vuelve a buscar la historia para mostrarla junto al error. Si no puede, ya respondió.
func (h estimaciones) paginaDespuesDeUnError(w http.ResponseWriter, r *http.Request, proyectoID int64, numero int) (paginaHistoria, bool) {
	vista, err := h.vista.Ver(r.Context(), proyectoID, numero)
	if respondeError(w, err, "volver a mostrar la historia") {
		return paginaHistoria{}, false
	}
	return nuevaPagina(vista), true
}
