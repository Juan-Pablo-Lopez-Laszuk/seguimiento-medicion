package metrics

import "testing"

// defectosDeEjemplo: Sprint 1 (ID 10) y Sprint 2 (ID 20), como en features/hu-34-defectos.feature.
var defectosDeEjemplo = []Defecto{
	{Severidad: "Alta", SprintDeteccion: 10, SprintResolucion: 10, Resuelto: true},
	{Severidad: "Alta", SprintDeteccion: 10, SprintResolucion: 20, Resuelto: true},
	{Severidad: "Baja", SprintDeteccion: 10},
	{Severidad: "Crítica", SprintDeteccion: 20}, // reabierto: sin sprint de resolución
}

// M8 y M9 · CA-34.1 y CA-34.2 · detectados por sprint de detección, resueltos por sprint de resolución y total
func TestContarDefectos_PorSprintYTotal(t *testing.T) {
	c := ContarDefectos(defectosDeEjemplo, "")
	if c.PorSprint[10] != (ConteoSprint{Detectados: 3, Resueltos: 1}) || c.PorSprint[20] != (ConteoSprint{Detectados: 1, Resueltos: 1}) {
		t.Errorf("por sprint = %+v", c.PorSprint)
	}
	if c.Detectados != 4 || c.Resueltos != 2 || c.Abiertos != 2 {
		t.Errorf("total = %+v", c)
	}
}

// CA-34.2 · el filtro por severidad se aplica a los sprints y al total
func TestContarDefectos_FiltraPorSeveridad(t *testing.T) {
	c := ContarDefectos(defectosDeEjemplo, "Alta")
	if c.PorSprint[10] != (ConteoSprint{Detectados: 2, Resueltos: 1}) || c.PorSprint[20] != (ConteoSprint{Resueltos: 1}) {
		t.Errorf("por sprint = %+v", c.PorSprint)
	}
	if c.Detectados != 2 || c.Resueltos != 2 || c.Abiertos != 0 {
		t.Errorf("total = %+v", c)
	}
}

// Un defecto marcado resuelto pero sin sprint de resolución no cuenta (no debería pasar: HU-28 lo exige)
func TestContarDefectos_ResueltoSinSprintNoCuenta(t *testing.T) {
	c := ContarDefectos([]Defecto{{SprintDeteccion: 10, Resuelto: true}}, "")
	if c.Resueltos != 0 || c.Abiertos != 1 || c.PorSprint[10].Resueltos != 0 {
		t.Errorf("conteo = %+v", c)
	}
}

func TestContarDefectos_SinDefectos(t *testing.T) {
	if c := ContarDefectos(nil, ""); c.Detectados != 0 || c.Resueltos != 0 || c.Abiertos != 0 || len(c.PorSprint) != 0 {
		t.Errorf("conteo = %+v", c)
	}
}
