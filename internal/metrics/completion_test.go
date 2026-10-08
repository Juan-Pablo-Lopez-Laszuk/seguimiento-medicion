package metrics_test

import (
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/metrics"
)

// historias arma `planificadas` historias, de las cuales las primeras `hechas` están en estado Hecho.
func historias(planificadas, hechas int) []metrics.Historia {
	hs := make([]metrics.Historia, planificadas)
	for i := 0; i < hechas; i++ {
		hs[i].Hecha = true
	}
	return hs
}

// CA-32.1: % completadas = historias Hechas / historias planificadas × 100 (se cuentan historias, no SP).
func TestCalcularPorcentajeCompletadas_HechasSobrePlanificadas(t *testing.T) {
	casos := []struct {
		nombre       string
		historias    []metrics.Historia
		hechas       int
		planificadas int
		quiero       float64
	}{
		{"3 de 4 hechas", historias(4, 3), 3, 4, 75},
		{"todas hechas", historias(3, 3), 3, 3, 100},
		{"ninguna hecha", historias(5, 0), 0, 5, 0},
		{"cuenta historias y no story points", []metrics.Historia{{SP: 13, Hecha: true}, {SP: 1}}, 1, 2, 50},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			p := metrics.CalcularPorcentajeCompletadas(c.historias)
			if p.Hechas != c.hechas || p.Planificadas != c.planificadas {
				t.Errorf("hechas/planificadas = %d/%d, se esperaba %d/%d", p.Hechas, p.Planificadas, c.hechas, c.planificadas)
			}
			if p.Valor != c.quiero {
				t.Errorf("porcentaje = %v, se esperaba %v", p.Valor, c.quiero)
			}
		})
	}
}
