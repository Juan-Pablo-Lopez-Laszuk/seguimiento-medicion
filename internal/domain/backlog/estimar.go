package backlog

import (
	"errors"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
)

// ErrEstimacionDeHistoriaHecha se devuelve al querer estimar una historia Hecha (RN6): los story points completados de
// su sprint cambiarían después de terminada.
var ErrEstimacionDeHistoriaHecha = errors.New("la historia está Hecha: no se puede cambiar su estimación")

// Estimar devuelve la historia con los story points nuevos, si el valor vale; la original no cambia. Si ya tenía una
// estimación, la reemplaza (RN5). El valor es obligatorio (RN3) y tiene que ser de la escala (RN1): si no, devuelve
// domain.ErroresValidacion con el error del campo CampoStoryPoints. Una historia Hecha no se re-estima (RN6): devuelve
// ErrEstimacionDeHistoriaHecha, y eso se revisa antes que el valor. Lo usan la página de la historia (HU-16) y, más
// adelante, las sesiones de Planning Poker para guardar el valor acordado.
//
// Spec: specs/HU-16-estimar-historia.md.
func (h Historia) Estimar(texto string) (Historia, error) {
	if h.Estado == EstadoHecho {
		return Historia{}, ErrEstimacionDeHistoriaHecha
	}
	puntos, err := parsearStoryPoints(texto)
	if err == nil && puntos == nil {
		err = ErrStoryPointsInvalidos // vacío: esta historia estima, no deja una historia "sin estimar"
	}
	errs := domain.ErroresValidacion{}
	errs.Agregar(CampoStoryPoints, err)
	if len(errs) > 0 {
		return Historia{}, errs
	}
	h.StoryPoints = puntos
	return h, nil
}
