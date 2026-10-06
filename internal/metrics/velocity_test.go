package metrics_test

import (
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/metrics"
)

// cerrados arma sprints cerrados numerados 1, 2, 3... con los story points completados que se pasan.
func cerrados(sp ...float64) []metrics.Sprint {
	sprints := make([]metrics.Sprint, 0, len(sp))
	for i, completados := range sp {
		sprints = append(sprints, metrics.Sprint{
			Numero:        i + 1,
			Estado:        metrics.SprintCerrado,
			SPCompletados: completados,
		})
	}
	return sprints
}

// CA-31.1 y CA-31.2: promedio de los últimos N sprints cerrados; si hay menos, promedia los que haya.
func TestCalcularVelocidad_PromediaLosUltimosSprintsCerrados(t *testing.T) {
	casos := []struct {
		nombre  string
		sprints []metrics.Sprint
		ventana int
		quiero  float64
	}{
		{"más sprints que la ventana: toma los últimos 3", cerrados(10, 20, 30, 40), 3, 30},
		{"exactamente la ventana", cerrados(10, 20, 30), 3, 20},
		{"menos sprints que la ventana: promedia los que hay", cerrados(10, 20), 3, 15},
		{"ventana de 1: solo el último sprint", cerrados(12, 8), 1, 8},
		{"un sprint cerrado con 0 SP cuenta como 0", cerrados(10, 0), 3, 5},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			velocidad, err := metrics.CalcularVelocidad(c.sprints, c.ventana)
			if err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
			if velocidad.Valor != c.quiero {
				t.Errorf("velocidad = %v, se esperaba %v", velocidad.Valor, c.quiero)
			}
		})
	}
}

// RN4: el orden en que llegan los sprints no cambia el resultado.
func TestCalcularVelocidad_NoDependeDelOrdenDeLosSprints(t *testing.T) {
	desordenados := []metrics.Sprint{
		{Numero: 4, Estado: metrics.SprintCerrado, SPCompletados: 40},
		{Numero: 1, Estado: metrics.SprintCerrado, SPCompletados: 10},
		{Numero: 3, Estado: metrics.SprintCerrado, SPCompletados: 30},
		{Numero: 2, Estado: metrics.SprintCerrado, SPCompletados: 20},
	}

	velocidad, err := metrics.CalcularVelocidad(desordenados, 3)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if velocidad.Valor != 30 {
		t.Errorf("velocidad = %v, se esperaba 30", velocidad.Valor)
	}
}

// CA-31.4 (RN1): los sprints activos o planificados no cuentan.
func TestCalcularVelocidad_IgnoraSprintsActivosYPlanificados(t *testing.T) {
	sprints := append(cerrados(10, 20),
		metrics.Sprint{Numero: 3, Estado: metrics.SprintActivo, SPCompletados: 50},
		metrics.Sprint{Numero: 4, Estado: metrics.SprintPlanificado, SPCompletados: 0},
	)

	velocidad, err := metrics.CalcularVelocidad(sprints, 3)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if velocidad.Valor != 15 {
		t.Errorf("velocidad = %v, se esperaba 15", velocidad.Valor)
	}
}

// CA-31.3: sin sprints cerrados la velocidad es 0 y se avisa (nunca se divide por cero).
func TestCalcularVelocidad_SinSprintsCerradosDevuelveCeroYAvisa(t *testing.T) {
	casos := []struct {
		nombre  string
		sprints []metrics.Sprint
	}{
		{"proyecto sin sprints", nil},
		{"solo sprints activos o planificados", []metrics.Sprint{
			{Numero: 1, Estado: metrics.SprintActivo, SPCompletados: 50},
			{Numero: 2, Estado: metrics.SprintPlanificado},
		}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			velocidad, err := metrics.CalcularVelocidad(c.sprints, 3)
			if err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
			if velocidad.Valor != 0 {
				t.Errorf("velocidad = %v, se esperaba 0", velocidad.Valor)
			}
			if velocidad.Mensaje() != "Aún no hay sprints cerrados" {
				t.Errorf("mensaje = %q, se esperaba el aviso de que no hay sprints cerrados", velocidad.Mensaje())
			}
		})
	}
}

// Cuando hay sprints cerrados no hay nada que avisar.
func TestCalcularVelocidad_ConSprintsCerradosNoMuestraMensaje(t *testing.T) {
	velocidad, err := metrics.CalcularVelocidad(cerrados(10, 20), 3)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if velocidad.Mensaje() != "" {
		t.Errorf("mensaje = %q, no se esperaba ninguno", velocidad.Mensaje())
	}
}
