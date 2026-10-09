package features

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
)

// hu23 guarda los datos de un escenario de HU-23 (registrar horas estimadas).
// Se crea uno nuevo por escenario, así un escenario no le deja datos al siguiente.
type hu23 struct {
	horasHistoria float64        // horas cargadas en la historia
	tareas        []effort.Tarea // tareas de la historia
	resultado     float64        // horas estimadas informadas
	errInformado  error          // error devuelto por la regla
	consultaHecha bool
}

// registrarPasosHU23 conecta cada frase de features/hu-23-horas-estimadas.feature con una función de Go.
func registrarPasosHU23(sc *godog.ScenarioContext) {
	e := &hu23{}

	sc.Step(`^cargo "([^"]*)" horas estimadas en una historia sin tareas$`, e.cargoHorasEnUnaHistoriaSinTareas)
	sc.Step(`^una historia estimada en "([^"]*)" horas$`, e.unaHistoriaEstimadaEn)
	sc.Step(`^una tarea estimada en "([^"]*)" horas$`, e.unaTareaEstimadaEn)
	sc.Step(`^consulto las horas estimadas de la historia$`, e.consultoLasHorasEstimadas)
	sc.Step(`^las horas estimadas de la historia son "([^"]*)"$`, e.lasHorasEstimadasSon)
	sc.Step(`^se rechaza la estimación con el error "([^"]*)"$`, e.seRechazaLaEstimacion)
}

// cargoHorasEnUnaHistoriaSinTareas usa la regla real: primero interpreta lo que escribe el usuario
// (ParsearHoras) y después calcula las horas de la historia.
func (e *hu23) cargoHorasEnUnaHistoriaSinTareas(texto string) error {
	horas, err := effort.ParsearHoras(texto)
	e.consultaHecha = true
	if err != nil {
		e.errInformado = err
		return nil
	}
	e.horasHistoria = horas
	return e.consultoLasHorasEstimadas()
}

func (e *hu23) unaHistoriaEstimadaEn(texto string) error {
	horas, err := effort.ParsearHoras(texto)
	if err != nil {
		return fmt.Errorf("el escenario usa horas inválidas para la historia %q: %w", texto, err)
	}
	e.horasHistoria = horas
	return nil
}

func (e *hu23) unaTareaEstimadaEn(texto string) error {
	horas, err := effort.ParsearHoras(texto)
	if err != nil {
		return fmt.Errorf("el escenario usa horas inválidas para la tarea %q: %w", texto, err)
	}
	e.tareas = append(e.tareas, effort.Tarea{HorasEstimadas: horas})
	return nil
}

func (e *hu23) consultoLasHorasEstimadas() error {
	e.resultado, e.errInformado = effort.HorasEstimadasDeHistoria(e.horasHistoria, e.tareas)
	e.consultaHecha = true
	return nil
}

// conComa muestra las horas como en pantalla: sin ceros de más y con coma decimal (5.5 → "5,5").
func conComa(horas float64) string {
	return strings.Replace(strconv.FormatFloat(horas, 'f', -1, 64), ".", ",", 1)
}

func (e *hu23) lasHorasEstimadasSon(esperadas string) error {
	if !e.consultaHecha {
		return godog.ErrPending
	}
	if e.errInformado != nil {
		return fmt.Errorf("error inesperado: %w", e.errInformado)
	}
	if conComa(e.resultado) != esperadas {
		return fmt.Errorf("horas estimadas: se esperaba %q y se obtuvo %q", esperadas, conComa(e.resultado))
	}
	return nil
}

func (e *hu23) seRechazaLaEstimacion(esperado string) error {
	if !e.consultaHecha {
		return godog.ErrPending
	}
	if e.errInformado == nil || e.errInformado.Error() != esperado {
		return fmt.Errorf("error: se esperaba %q y se obtuvo %v", esperado, e.errInformado)
	}
	return nil
}
