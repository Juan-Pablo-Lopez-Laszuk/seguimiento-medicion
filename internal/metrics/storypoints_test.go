package metrics_test

import (
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/metrics"
)

// CA-30.1: planificados = suma de SP del sprint; completados = suma de SP de las historias Hechas.
func TestCalcularStoryPoints_PlanificadosYCompletados(t *testing.T) {
	casos := []struct {
		nombre       string
		historias    []metrics.Historia
		planificados float64
		completados  float64
	}{
		{"sprint con historias hechas y sin terminar", []metrics.Historia{
			{SP: 5, Hecha: true},
			{SP: 3, Hecha: true},
			{SP: 8},
		}, 16, 8},
		{"ninguna historia hecha", []metrics.Historia{{SP: 5}, {SP: 3}}, 8, 0},
		{"todas las historias hechas", []metrics.Historia{{SP: 5, Hecha: true}, {SP: 3, Hecha: true}}, 8, 8},
		{"sprint sin historias", nil, 0, 0},
		{"una historia de 0 SP está estimada y suma 0", []metrics.Historia{{SP: 0, Hecha: true}, {SP: 2}}, 2, 0},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			sp, err := metrics.CalcularStoryPoints(c.historias)
			if err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
			if sp.Planificados != c.planificados {
				t.Errorf("planificados = %v, se esperaba %v", sp.Planificados, c.planificados)
			}
			if sp.Completados != c.completados {
				t.Errorf("completados = %v, se esperaba %v", sp.Completados, c.completados)
			}
		})
	}
}
