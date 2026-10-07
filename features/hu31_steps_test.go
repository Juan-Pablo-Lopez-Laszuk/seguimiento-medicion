package features

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/metrics"
)

// hu31 guarda los datos de un escenario de HU-31 (velocidad del equipo) mientras se ejecuta.
// Se crea uno nuevo por escenario, así un escenario no le deja datos al siguiente.
type hu31 struct {
	ventana       int       // cantidad de sprints que se promedian
	spCerrados    []float64 // SP completados de cada sprint cerrado, en orden
	spActivo      float64   // SP que lleva el sprint activo (no deben contar)
	velocidad     float64   // resultado informado
	mensaje       string    // aviso al usuario (por ejemplo, "Aún no hay sprints cerrados")
	errInformado  error     // error devuelto por el cálculo
	consultaHecha bool
}

// registrarPasosHU31 conecta cada frase de features/hu-31-velocidad.feature con una función de Go.
func registrarPasosHU31(sc *godog.ScenarioContext) {
	e := &hu31{}

	// (-?\d+) acepta también números negativos: el escenario de error prueba la ventana -1.
	sc.Step(`^un proyecto con una ventana de velocidad de (-?\d+) sprints$`, e.unProyectoConVentana)
	sc.Step(`^sprints cerrados que completaron "([^"]*)" story points$`, e.sprintsCerradosQueCompletaron)
	sc.Step(`^que el proyecto no tiene sprints cerrados$`, e.sinSprintsCerrados)
	sc.Step(`^un sprint activo que lleva completados (\d+) story points$`, e.unSprintActivo)
	sc.Step(`^consulto la velocidad del equipo$`, e.consultoLaVelocidad)
	sc.Step(`^la velocidad informada es (\d+(?:[.,]\d+)?)$`, e.laVelocidadInformadaEs)
	sc.Step(`^se muestra el mensaje "([^"]*)"$`, e.seMuestraElMensaje)
	sc.Step(`^se informa el error "([^"]*)"$`, e.seInformaElError)
}

func (e *hu31) unProyectoConVentana(ventana int) error {
	e.ventana = ventana
	return nil
}

func (e *hu31) sprintsCerradosQueCompletaron(lista string) error {
	e.spCerrados = nil
	for _, parte := range strings.Split(lista, ",") {
		sp, err := strconv.ParseFloat(strings.TrimSpace(parte), 64)
		if err != nil {
			return fmt.Errorf("story points inválidos %q: %w", parte, err)
		}
		e.spCerrados = append(e.spCerrados, sp)
	}
	return nil
}

func (e *hu31) sinSprintsCerrados() error {
	e.spCerrados = nil
	return nil
}

func (e *hu31) unSprintActivo(sp int) error {
	e.spActivo = float64(sp)
	return nil
}

// consultoLaVelocidad arma los sprints del escenario y llama al cálculo real de internal/metrics
// (specs/HU-31-velocidad.md). Guarda el resultado para que lo revisen los pasos "Entonces".
func (e *hu31) consultoLaVelocidad() error {
	var sprints []metrics.Sprint
	for i, sp := range e.spCerrados {
		sprints = append(sprints, metrics.Sprint{Numero: i + 1, Estado: metrics.SprintCerrado, SPCompletados: sp})
	}
	if e.spActivo > 0 {
		sprints = append(sprints, metrics.Sprint{Numero: len(sprints) + 1, Estado: metrics.SprintActivo, SPCompletados: e.spActivo})
	}

	velocidad, err := metrics.CalcularVelocidad(sprints, e.ventana)
	e.velocidad = velocidad.Valor
	e.mensaje = velocidad.Mensaje()
	e.errInformado = err
	e.consultaHecha = true
	return nil
}

func (e *hu31) laVelocidadInformadaEs(esperada string) error {
	if !e.consultaHecha {
		return godog.ErrPending
	}
	quiero, err := strconv.ParseFloat(strings.Replace(esperada, ",", ".", 1), 64)
	if err != nil {
		return fmt.Errorf("velocidad esperada inválida %q: %w", esperada, err)
	}
	if e.velocidad != quiero {
		return fmt.Errorf("velocidad: se esperaba %v y se obtuvo %v", quiero, e.velocidad)
	}
	return nil
}

func (e *hu31) seMuestraElMensaje(esperado string) error {
	if !e.consultaHecha {
		return godog.ErrPending
	}
	if e.mensaje != esperado {
		return fmt.Errorf("mensaje: se esperaba %q y se obtuvo %q", esperado, e.mensaje)
	}
	return nil
}

func (e *hu31) seInformaElError(esperado string) error {
	if !e.consultaHecha {
		return godog.ErrPending
	}
	if e.errInformado == nil || e.errInformado.Error() != esperado {
		return fmt.Errorf("error: se esperaba %q y se obtuvo %v", esperado, e.errInformado)
	}
	return nil
}
