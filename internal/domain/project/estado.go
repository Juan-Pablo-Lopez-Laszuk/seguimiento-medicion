package project

import (
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
)

// EstadoActual calcula el estado del proyecto; no se guarda (spec: specs/HU-04-estado-proyecto.md, RN1).
// inicioPrimerSprint es la fecha de inicio del primer sprint, o la fecha cero si el proyecto no tiene sprints
// (el paquete sprint la calcula con sprint.InicioDelPrimero). hoy es la fecha de Argentina (RN6).
func EstadoActual(p Proyecto, inicioPrimerSprint, hoy time.Time) Estado {
	hoy = domain.SoloFecha(hoy)
	switch {
	case inicioPrimerSprint.IsZero(): // CA-04.3: sin sprints, sean cuales sean las fechas
		return EstadoPlanificado
	case hoy.Before(domain.SoloFecha(inicioPrimerSprint)):
		return EstadoPlanificado
	case hoy.After(p.FechaFin):
		return EstadoFinalizado
	default:
		return EstadoEnCurso
	}
}
