package features

import (
	"fmt"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/metrics"
)

// hu32 guarda los datos de un escenario de HU-32 (porcentaje de historias completadas).
// Se crea uno nuevo por escenario, así un escenario no le deja datos al siguiente.
type hu32 struct {
	sprints       [][]metrics.Historia // historias de cada sprint del escenario
	resultado     metrics.Porcentaje   // resultado informado
	consultaHecha bool
}

// registrarPasosHU32 conecta cada frase de features/hu-32-porcentaje.feature con una función de Go.
func registrarPasosHU32(sc *godog.ScenarioContext) {
	e := &hu32{}

	sc.Step(`^(?:un|otro) sprint con (\d+) historias de las cuales (\d+) están hechas$`, e.unSprintConHistorias)
	sc.Step(`^un sprint sin historias planificadas$`, e.unSprintSinHistoriasPlanificadas)
	sc.Step(`^consulto el porcentaje de completadas del sprint$`, e.consultoElPorcentajeDelSprint)
	sc.Step(`^consulto el porcentaje de completadas del proyecto$`, e.consultoElPorcentajeDelProyecto)
	sc.Step(`^el porcentaje de historias completadas es "([^"]*)"$`, e.elPorcentajeEs)
}

func (e *hu32) unSprintConHistorias(planificadas, hechas int) error {
	if hechas > planificadas {
		return fmt.Errorf("el escenario tiene más historias hechas (%d) que planificadas (%d)", hechas, planificadas)
	}
	historias := make([]metrics.Historia, planificadas)
	for i := 0; i < hechas; i++ {
		historias[i].Hecha = true
	}
	e.sprints = append(e.sprints, historias)
	return nil
}

func (e *hu32) unSprintSinHistoriasPlanificadas() error {
	e.sprints = append(e.sprints, nil)
	return nil
}

// consultoElPorcentajeDelSprint llama al cálculo real de internal/metrics con el único sprint del escenario.
func (e *hu32) consultoElPorcentajeDelSprint() error {
	if len(e.sprints) != 1 {
		return fmt.Errorf("se esperaba 1 sprint en el escenario y hay %d", len(e.sprints))
	}
	e.resultado = metrics.CalcularPorcentajeCompletadas(e.sprints[0])
	e.consultaHecha = true
	return nil
}

// consultoElPorcentajeDelProyecto calcula cada sprint y después el total del proyecto.
func (e *hu32) consultoElPorcentajeDelProyecto() error {
	porSprint := make([]metrics.Porcentaje, 0, len(e.sprints))
	for _, historias := range e.sprints {
		porSprint = append(porSprint, metrics.CalcularPorcentajeCompletadas(historias))
	}
	e.resultado = metrics.SumarPorcentajes(porSprint)
	e.consultaHecha = true
	return nil
}

func (e *hu32) elPorcentajeEs(esperado string) error {
	if !e.consultaHecha {
		return godog.ErrPending
	}
	if e.resultado.Texto() != esperado {
		return fmt.Errorf("porcentaje: se esperaba %q y se obtuvo %q", esperado, e.resultado.Texto())
	}
	return nil
}
