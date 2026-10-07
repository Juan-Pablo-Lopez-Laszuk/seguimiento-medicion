package features

import (
	"fmt"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/metrics"
)

// estadoHecho es el estado de una historia terminada (docs/modelo-datos.md).
const estadoHecho = "Hecho"

// hu30 guarda los datos de un escenario de HU-30 (story points planificados y completados).
// Se crea uno nuevo por escenario, así un escenario no le deja datos al siguiente.
type hu30 struct {
	sprints       [][]metrics.Historia // historias de cada sprint del escenario
	resultado     metrics.StoryPoints  // resultado informado
	errInformado  error                // error devuelto por el cálculo
	consultaHecha bool
}

// registrarPasosHU30 conecta cada frase de features/hu-30-sp.feature con una función de Go.
func registrarPasosHU30(sc *godog.ScenarioContext) {
	e := &hu30{}

	// (-?\d+) acepta también números negativos: el escenario de error prueba una historia de -3 SP.
	sc.Step(`^un sprint sin historias$`, e.unSprintSinHistorias)
	sc.Step(`^(?:un|otro) sprint con una historia de (-?\d+) story points en estado "([^"]*)"$`, e.unSprintConUnaHistoria)
	sc.Step(`^una historia de (-?\d+) story points en estado "([^"]*)"$`, e.unaHistoria)
	sc.Step(`^una historia sin estimar en estado "([^"]*)"$`, e.unaHistoriaSinEstimar)
	sc.Step(`^consulto los story points del sprint$`, e.consultoLosStoryPointsDelSprint)
	sc.Step(`^consulto los story points del proyecto$`, e.consultoLosStoryPointsDelProyecto)
	sc.Step(`^los story points planificados son (\d+)$`, e.losPlanificadosSon)
	sc.Step(`^los story points completados son (\d+)$`, e.losCompletadosSon)
	sc.Step(`^no se muestra ninguna advertencia de estimación$`, e.noSeMuestraAdvertencia)
	sc.Step(`^se muestra la advertencia "([^"]*)"$`, e.seMuestraLaAdvertencia)
	sc.Step(`^el cálculo de story points se rechaza con el error "([^"]*)"$`, e.seRechazaConElError)
}

func (e *hu30) unSprintSinHistorias() error {
	e.sprints = append(e.sprints, nil)
	return nil
}

func (e *hu30) unSprintConUnaHistoria(sp int, estado string) error {
	e.sprints = append(e.sprints, nil)
	return e.unaHistoria(sp, estado)
}

// agregar suma una historia al último sprint del escenario.
func (e *hu30) agregar(h metrics.Historia) error {
	if len(e.sprints) == 0 {
		return fmt.Errorf("el escenario no definió ningún sprint antes de la historia")
	}
	ultimo := len(e.sprints) - 1
	e.sprints[ultimo] = append(e.sprints[ultimo], h)
	return nil
}

func (e *hu30) unaHistoria(sp int, estado string) error {
	return e.agregar(metrics.Historia{SP: float64(sp), Hecha: estado == estadoHecho})
}

func (e *hu30) unaHistoriaSinEstimar(estado string) error {
	return e.agregar(metrics.Historia{SinEstimar: true, Hecha: estado == estadoHecho})
}

// consultoLosStoryPointsDelSprint llama al cálculo real de internal/metrics con el primer sprint.
func (e *hu30) consultoLosStoryPointsDelSprint() error {
	if len(e.sprints) != 1 {
		return fmt.Errorf("se esperaba 1 sprint en el escenario y hay %d", len(e.sprints))
	}
	e.resultado, e.errInformado = metrics.CalcularStoryPoints(e.sprints[0])
	e.consultaHecha = true
	return nil
}

// consultoLosStoryPointsDelProyecto calcula cada sprint y después suma el total del proyecto.
func (e *hu30) consultoLosStoryPointsDelProyecto() error {
	porSprint := make([]metrics.StoryPoints, 0, len(e.sprints))
	for _, historias := range e.sprints {
		sp, err := metrics.CalcularStoryPoints(historias)
		if err != nil {
			e.errInformado = err
			e.consultaHecha = true
			return nil
		}
		porSprint = append(porSprint, sp)
	}
	e.resultado = metrics.SumarStoryPoints(porSprint)
	e.consultaHecha = true
	return nil
}

// sinError falla si todavía no se consultó o si el cálculo devolvió un error que el escenario no esperaba.
func (e *hu30) sinError() error {
	if !e.consultaHecha {
		return godog.ErrPending
	}
	if e.errInformado != nil {
		return fmt.Errorf("error inesperado: %w", e.errInformado)
	}
	return nil
}

func (e *hu30) losPlanificadosSon(esperados int) error {
	if err := e.sinError(); err != nil {
		return err
	}
	if e.resultado.Planificados != float64(esperados) {
		return fmt.Errorf("planificados: se esperaba %d y se obtuvo %v", esperados, e.resultado.Planificados)
	}
	return nil
}

func (e *hu30) losCompletadosSon(esperados int) error {
	if err := e.sinError(); err != nil {
		return err
	}
	if e.resultado.Completados != float64(esperados) {
		return fmt.Errorf("completados: se esperaba %d y se obtuvo %v", esperados, e.resultado.Completados)
	}
	return nil
}

func (e *hu30) noSeMuestraAdvertencia() error {
	return e.seMuestraLaAdvertencia("")
}

func (e *hu30) seMuestraLaAdvertencia(esperada string) error {
	if err := e.sinError(); err != nil {
		return err
	}
	if e.resultado.Advertencia() != esperada {
		return fmt.Errorf("advertencia: se esperaba %q y se obtuvo %q", esperada, e.resultado.Advertencia())
	}
	return nil
}

func (e *hu30) seRechazaConElError(esperado string) error {
	if !e.consultaHecha {
		return godog.ErrPending
	}
	if e.errInformado == nil || e.errInformado.Error() != esperado {
		return fmt.Errorf("error: se esperaba %q y se obtuvo %v", esperado, e.errInformado)
	}
	return nil
}
