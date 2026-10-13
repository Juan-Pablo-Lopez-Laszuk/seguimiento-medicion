package project_test

import (
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
)

// HU-04 RN1 · el estado se calcula con las fechas del proyecto, el inicio del primer sprint y hoy
// (CA-04.2 y CA-04.3). Proyecto de la spec: del 05/10/2026 al 01/11/2026.
func TestEstadoActual(t *testing.T) {
	p := project.Proyecto{FechaInicio: fecha(t, "2026-10-05"), FechaFin: fecha(t, "2026-11-01")}
	primerSprint := fecha(t, "2026-10-05")
	sinSprints := time.Time{}

	casos := []struct {
		nombre       string
		inicioSprint time.Time
		hoy          string
		esperado     project.Estado
	}{
		{"sin sprints, en fecha", sinSprints, "2026-10-20", project.EstadoPlanificado},
		{"sin sprints, ya pasó el fin", sinSprints, "2026-11-10", project.EstadoPlanificado},
		{"antes del primer sprint", primerSprint, "2026-10-04", project.EstadoPlanificado},
		{"el día que empieza el primer sprint", primerSprint, "2026-10-05", project.EstadoEnCurso},
		{"el último día del proyecto", primerSprint, "2026-11-01", project.EstadoEnCurso},
		{"el día después del fin", primerSprint, "2026-11-02", project.EstadoFinalizado},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := project.EstadoActual(p, c.inicioSprint, fecha(t, c.hoy)); got != c.esperado {
				t.Errorf("hoy %s: se esperaba %q y se obtuvo %q", c.hoy, c.esperado, got)
			}
		})
	}
}

// HU-04 RN1 · un primer sprint que empieza después del inicio del proyecto: hasta ese día sigue Planificado
func TestEstadoActual_PrimerSprintPosteriorAlInicio(t *testing.T) {
	p := project.Proyecto{FechaInicio: fecha(t, "2026-10-05"), FechaFin: fecha(t, "2026-11-01")}
	primerSprint := fecha(t, "2026-10-12")
	if got := project.EstadoActual(p, primerSprint, fecha(t, "2026-10-08")); got != project.EstadoPlanificado {
		t.Errorf("se esperaba Planificado y se obtuvo %q", got)
	}
	if got := project.EstadoActual(p, primerSprint, fecha(t, "2026-10-12")); got != project.EstadoEnCurso {
		t.Errorf("se esperaba En curso y se obtuvo %q", got)
	}
}
