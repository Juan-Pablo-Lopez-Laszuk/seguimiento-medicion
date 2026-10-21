package features

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// hu16b guarda el estado de un escenario de HU-16b (historial de estimaciones), con sus propios repositorios en
// memoria. Las historias del Antecedentes se buscan por su título.
type hu16b struct {
	repoHistorias *memory.Historias
	estimaciones  *service.Estimaciones
	edicion       *service.EdicionHistorias
	proyectoID    int64
	historias     map[string]backlog.Historia
	err           error
}

// registrarPasosHU16b conecta cada frase de features/hu-16b-historial-estimaciones.feature con una función de Go.
func registrarPasosHU16b(sc *godog.ScenarioContext) {
	repoProyectos := memory.NuevoProyectos()
	repoCriterios, repoTareas := memory.NuevoCriterios(), memory.NuevoTareas()
	e := &hu16b{repoHistorias: memory.NuevoHistorias(repoCriterios, repoTareas), historias: map[string]backlog.Historia{}}
	e.estimaciones = service.NuevoEstimaciones(repoProyectos, e.repoHistorias)
	e.edicion = service.NuevoEdicionHistorias(repoProyectos, e.repoHistorias, repoCriterios, repoTareas)
	proyectos := service.NuevoProyectos(repoProyectos, time.Now)
	historias := service.NuevoHistorias(repoProyectos, e.repoHistorias)

	sc.Step(`^el proyecto "([^"]*)" con las historias "([^"]*)" y "([^"]*)" para el historial$`, func(nombre, primera, segunda string) error {
		return e.elProyectoConLasHistorias(proyectos, historias, nombre, primera, segunda)
	})
	sc.Step(`^que "([^"]*)" está estimada en (\d+) story points$`, e.queEstaEstimadaEn)
	sc.Step(`^que "([^"]*)" se re-estimó a "([^"]*)"$`, e.queSeReestimoA)
	sc.Step(`^que "([^"]*)" está Hecha$`, e.queEstaHecha)
	sc.Step(`^cambio la estimación de "([^"]*)" a "([^"]*)"$`, e.cambioLaEstimacionDe)
	sc.Step(`^elimino la historia "([^"]*)"$`, e.eliminoLaHistoria)
	sc.Step(`^el historial de "([^"]*)" es:$`, e.elHistorialDeEs)
	sc.Step(`^el historial de "([^"]*)" está vacío$`, e.elHistorialDeEstaVacio)
	sc.Step(`^la estimación de "([^"]*)" es "([^"]*)"$`, e.laEstimacionDeEs)
	sc.Step(`^la re-estimación da como resultado "([^"]*)"$`, e.elCambioDaComoResultado)
	sc.Step(`^la historia "([^"]*)" ya no existe$`, e.laHistoriaYaNoExiste)
	sc.Step(`^no queda ningún historial de la historia eliminada$`, e.noQuedaNingunHistorial)
}

func (e *hu16b) elProyectoConLasHistorias(proyectos *service.Proyectos, historias *service.Historias, nombre, primera, segunda string) error {
	ctx := context.Background()
	p, err := proyectos.Crear(ctx, project.Datos{
		Nombre:      nombre,
		FechaInicio: fechaDeEscenario("2026-10-05"),
		FechaFin:    fechaDeEscenario("2026-11-01"),
	})
	if err != nil {
		return fmt.Errorf("no se pudo crear el proyecto %q: %w", nombre, err)
	}
	e.proyectoID = p.ID
	for _, titulo := range []string{primera, segunda} {
		h, err := historias.Crear(ctx, p.ID, backlog.Datos{Titulo: titulo, Prioridad: backlog.PrioridadAlta})
		if err != nil {
			return fmt.Errorf("no se pudo crear la historia %q: %w", titulo, err)
		}
		e.historias[titulo] = h
	}
	return nil
}

func (e *hu16b) historia(titulo string) (backlog.Historia, error) {
	h, ok := e.historias[titulo]
	if !ok {
		return backlog.Historia{}, fmt.Errorf("el escenario no tiene una historia %q", titulo)
	}
	return h, nil
}

func (e *hu16b) estimar(titulo, puntos string) error {
	h, err := e.historia(titulo)
	if err != nil {
		return err
	}
	_, e.err = e.estimaciones.Estimar(context.Background(), e.proyectoID, h.Numero, puntos)
	return nil
}

func (e *hu16b) queEstaEstimadaEn(titulo string, puntos int) error {
	if err := e.estimar(titulo, strconv.Itoa(puntos)); err != nil {
		return err
	}
	return e.err
}

func (e *hu16b) queSeReestimoA(titulo, puntos string) error {
	if err := e.estimar(titulo, puntos); err != nil {
		return err
	}
	return e.err
}

// queEstaHecha deja la historia Hecha directo en el repositorio: pasar a Hecho es de HU-13.
func (e *hu16b) queEstaHecha(titulo string) error {
	h, err := e.historia(titulo)
	if err != nil {
		return err
	}
	return e.repoHistorias.ActualizarEstado(context.Background(), h.ID, backlog.EstadoHecho)
}

func (e *hu16b) cambioLaEstimacionDe(titulo, puntos string) error {
	return e.estimar(titulo, puntos)
}

func (e *hu16b) eliminoLaHistoria(titulo string) error {
	h, err := e.historia(titulo)
	if err != nil {
		return err
	}
	_, e.err = e.edicion.Eliminar(context.Background(), e.proyectoID, h.Numero)
	return e.err
}

// elHistorialDeEs compara el historial, de la más reciente a la más antigua, con la tabla (columnas anterior y nuevo).
func (e *hu16b) elHistorialDeEs(titulo string, tabla *godog.Table) error {
	h, err := e.historia(titulo)
	if err != nil {
		return err
	}
	historial, err := e.repoHistorias.ListarEstimaciones(context.Background(), h.ID)
	if err != nil {
		return err
	}
	filas := tabla.Rows[1:] // la primera fila es el encabezado
	if len(historial) != len(filas) {
		return fmt.Errorf("historial de %q: se esperaban %d entradas y hay %d: %+v", titulo, len(filas), len(historial), historial)
	}
	for i, fila := range filas {
		anterior, errA := strconv.Atoi(fila.Cells[0].Value)
		nuevo, errN := strconv.Atoi(fila.Cells[1].Value)
		if errA != nil || errN != nil {
			return fmt.Errorf("fila %d de la tabla mal escrita: %v", i+1, fila.Cells)
		}
		if historial[i].Anterior != anterior || historial[i].Nuevo != nuevo {
			return fmt.Errorf("historial de %q, entrada %d: se esperaba %d → %d y se obtuvo %d → %d",
				titulo, i+1, anterior, nuevo, historial[i].Anterior, historial[i].Nuevo)
		}
	}
	return nil
}

func (e *hu16b) elHistorialDeEstaVacio(titulo string) error {
	h, err := e.historia(titulo)
	if err != nil {
		return err
	}
	historial, err := e.repoHistorias.ListarEstimaciones(context.Background(), h.ID)
	if err != nil {
		return err
	}
	if len(historial) != 0 {
		return fmt.Errorf("historial de %q: se esperaba vacío y se obtuvo %+v", titulo, historial)
	}
	return nil
}

func (e *hu16b) laEstimacionDeEs(titulo, esperada string) error {
	h, err := e.historia(titulo)
	if err != nil {
		return err
	}
	actual, err := e.repoHistorias.BuscarPorNumero(context.Background(), e.proyectoID, h.Numero)
	if err != nil {
		return err
	}
	obtenida := "sin estimar"
	if actual.StoryPoints != nil {
		obtenida = strconv.Itoa(*actual.StoryPoints)
	}
	if obtenida != esperada {
		return fmt.Errorf("estimación de %q: se esperaba %q y se obtuvo %q", titulo, esperada, obtenida)
	}
	return nil
}

// elCambioDaComoResultado acepta "ok", "campo: mensaje" (error de validación) o el mensaje de otro error.
func (e *hu16b) elCambioDaComoResultado(resultado string) error {
	if resultado == "ok" {
		if e.err != nil {
			return fmt.Errorf("se esperaba que valga y se obtuvo: %w", e.err)
		}
		return nil
	}
	var errs domain.ErroresValidacion
	if errors.As(e.err, &errs) {
		campo, mensaje, ok := strings.Cut(resultado, ": ")
		if !ok {
			return fmt.Errorf("resultado mal escrito en el escenario: %q (se espera \"campo: mensaje\")", resultado)
		}
		if errs[campo] == nil || errs[campo].Error() != mensaje {
			return fmt.Errorf("campo %q: se esperaba %q y se obtuvo %v", campo, mensaje, errs[campo])
		}
		return nil
	}
	if e.err == nil || e.err.Error() != resultado {
		return fmt.Errorf("se esperaba el error %q y se obtuvo: %v", resultado, e.err)
	}
	return nil
}

func (e *hu16b) laHistoriaYaNoExiste(titulo string) error {
	h, err := e.historia(titulo)
	if err != nil {
		return err
	}
	if _, err := e.repoHistorias.BuscarPorNumero(context.Background(), e.proyectoID, h.Numero); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		return fmt.Errorf("la historia %q tenía que haberse eliminado y se obtuvo: %v", titulo, err)
	}
	return nil
}

// noQuedaNingunHistorial mira el historial de la historia eliminada (la que se borró en el escenario).
func (e *hu16b) noQuedaNingunHistorial() error {
	for titulo, h := range e.historias {
		if _, err := e.repoHistorias.BuscarPorNumero(context.Background(), e.proyectoID, h.Numero); err == nil {
			continue // sigue existiendo: no es la eliminada
		}
		historial, err := e.repoHistorias.ListarEstimaciones(context.Background(), h.ID)
		if err != nil {
			return err
		}
		if len(historial) != 0 {
			return fmt.Errorf("el historial de la historia eliminada %q tenía que desaparecer y quedó %+v", titulo, historial)
		}
	}
	return nil
}
