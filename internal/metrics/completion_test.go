package metrics_test

import (
	"math"
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

// CA-32.2: un sprint sin historias da 0 % y no es un error (la división por cero está controlada).
func TestCalcularPorcentajeCompletadas_SprintSinHistoriasDaCero(t *testing.T) {
	p := metrics.CalcularPorcentajeCompletadas(nil)
	if p.Valor != 0 {
		t.Errorf("porcentaje = %v, se esperaba 0", p.Valor)
	}
	if p.Hechas != 0 || p.Planificadas != 0 {
		t.Errorf("hechas/planificadas = %d/%d, se esperaba 0/0", p.Hechas, p.Planificadas)
	}
}

// CA-32.3: el porcentaje del proyecto se calcula con el total de historias de todos sus sprints,
// no promediando los porcentajes de cada sprint.
func TestSumarPorcentajes_TotalDelProyecto(t *testing.T) {
	porSprint := []metrics.Porcentaje{
		metrics.CalcularPorcentajeCompletadas(historias(4, 3)), // 75 %
		metrics.CalcularPorcentajeCompletadas(historias(3, 3)), // 100 %
	}

	total := metrics.SumarPorcentajes(porSprint)

	if total.Hechas != 6 || total.Planificadas != 7 {
		t.Errorf("hechas/planificadas = %d/%d, se esperaba 6/7", total.Hechas, total.Planificadas)
	}
	// 6 / 7 × 100 = 85,71…; el promedio de los porcentajes (87,5) sería incorrecto.
	if math.Abs(total.Valor-600.0/7) > 1e-9 {
		t.Errorf("porcentaje = %v, se esperaba 85,71…", total.Valor)
	}
}

// Un proyecto sin sprints, o con sprints sin historias, da 0 %.
func TestSumarPorcentajes_ProyectoSinHistorias(t *testing.T) {
	for _, porSprint := range [][]metrics.Porcentaje{nil, {metrics.CalcularPorcentajeCompletadas(nil)}} {
		if total := metrics.SumarPorcentajes(porSprint); total.Valor != 0 {
			t.Errorf("porcentaje = %v, se esperaba 0", total.Valor)
		}
	}
}
