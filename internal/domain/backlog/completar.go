package backlog

import (
	"errors"
	"fmt"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
)

// ErrNoEsElSprintActivo se devuelve al completar una historia de un sprint que no es el activo (HU-13 RN1).
var ErrNoEsElSprintActivo = errors.New("solo se completan historias del sprint activo")

// Completar devuelve la historia Hecha y con la fecha de completado (solo la fecha de hoy, sin la hora), si se puede
// completar (HU-13); la original no cambia y sigue en su sprint. delSprintActivo dice si la historia es del sprint
// activo (RN1). La transición y los criterios cumplidos son la regla de HU-08 (CambiarEstado): acá solo se le pone a
// cada error un mensaje que nombra la historia, y sigue siendo reconocible con errors.Is.
func (h Historia) Completar(delSprintActivo bool, progreso Progreso, hoy time.Time) (Historia, error) {
	if !delSprintActivo {
		return Historia{}, ErrNoEsElSprintActivo
	}
	hecha, err := h.CambiarEstado(EstadoHecho, progreso)
	switch {
	case errors.Is(err, ErrYaEstaEnEseEstado):
		return Historia{}, errorDeEstado{base: err, mensaje: h.Identificador() + " ya está Hecha"}
	case errors.Is(err, ErrTransicionInvalida) && h.Estado == EstadoEnSprint:
		return Historia{}, errorDeEstado{base: err,
			mensaje: h.Identificador() + " todavía no empezó: primero hay que pasarla a En progreso"}
	case errors.Is(err, ErrCriteriosSinCumplir):
		return Historia{}, errorDeEstado{base: err, mensaje: h.Identificador() + " no se puede completar: " + criteriosQueFaltan(progreso)}
	case err != nil:
		return Historia{}, err
	}
	fecha := domain.SoloFecha(hoy)
	hecha.CompletadaEn = &fecha
	return hecha, nil
}

// criteriosQueFaltan explica qué falta para completar: "falta cumplir 1 de 2 criterios", o que no tiene criterios.
func criteriosQueFaltan(p Progreso) string {
	if !p.TieneCriterios() {
		return "no tiene criterios de aceptación"
	}
	return fmt.Sprintf("falta cumplir %d de %d criterios", p.Total-p.Cumplidos, p.Total)
}
