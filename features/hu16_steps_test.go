package features

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// hu16 guarda el estado de un escenario de HU-16 (estimar una historia con story points), con sus propios repositorios
// en memoria. "La historia" de todos los pasos es la del Antecedentes.
type hu16 struct {
	repoHistorias *memory.Historias
	historias     *service.Historias
	estimaciones  *service.Estimaciones
	proyectoID    int64
	historia      backlog.Historia
	err           error
}

// registrarPasosHU16 conecta cada frase de features/hu-16-estimar-historia.feature con una función de Go.
func registrarPasosHU16(sc *godog.ScenarioContext) {
	repoProyectos := memory.NuevoProyectos()
	e := &hu16{repoHistorias: memory.NuevoHistorias()}
	e.historias = service.NuevoHistorias(repoProyectos, e.repoHistorias)
	e.estimaciones = service.NuevoEstimaciones(repoProyectos, e.repoHistorias)
	proyectos := service.NuevoProyectos(repoProyectos, time.Now)

	sc.Step(`^el proyecto "([^"]*)" con la historia "([^"]*)" para estimar$`, func(nombre, titulo string) error {
		return e.elProyectoConLaHistoria(proyectos, nombre, titulo)
	})
	sc.Step(`^que la historia ya está estimada en "([^"]*)" story points$`, e.queLaHistoriaYaEstaEstimadaEn)
	sc.Step(`^que la historia a estimar está "([^"]*)"$`, e.queLaHistoriaAEstimarEsta)
	sc.Step(`^estimo la historia en "([^"]*)" story points$`, e.estimoLaHistoriaEn)
	sc.Step(`^la estimación de la historia es "([^"]*)"$`, e.laEstimacionDeLaHistoriaEs)
	sc.Step(`^la estimación da como resultado "([^"]*)"$`, e.laEstimacionDaComoResultado)
}

func (e *hu16) elProyectoConLaHistoria(proyectos *service.Proyectos, nombre, titulo string) error {
	p, err := proyectos.Crear(context.Background(), project.Datos{
		Nombre:      nombre,
		FechaInicio: fechaDeEscenario("2026-10-05"),
		FechaFin:    fechaDeEscenario("2026-11-01"),
	})
	if err != nil {
		return fmt.Errorf("no se pudo crear el proyecto %q: %w", nombre, err)
	}
	e.proyectoID = p.ID
	e.historia, err = e.historias.Crear(context.Background(), p.ID, backlog.Datos{Titulo: titulo, Prioridad: backlog.PrioridadAlta})
	if err != nil {
		return fmt.Errorf("no se pudo crear la historia %q: %w", titulo, err)
	}
	return nil
}

func (e *hu16) queLaHistoriaYaEstaEstimadaEn(puntos string) error {
	_, err := e.estimaciones.Estimar(context.Background(), e.proyectoID, e.historia.Numero, puntos)
	return err
}

// queLaHistoriaAEstimarEsta deja la historia en ese estado directo en el repositorio: pasar a En sprint o a Hecho es de
// HU-12 y HU-13.
func (e *hu16) queLaHistoriaAEstimarEsta(estado string) error {
	return e.repoHistorias.ActualizarEstado(context.Background(), e.historia.ID, backlog.Estado(estado))
}

func (e *hu16) estimoLaHistoriaEn(puntos string) error {
	_, e.err = e.estimaciones.Estimar(context.Background(), e.proyectoID, e.historia.Numero, puntos)
	return nil
}

// laEstimacionDeLaHistoriaEs acepta el número de story points o "sin estimar".
func (e *hu16) laEstimacionDeLaHistoriaEs(esperada string) error {
	h, err := e.repoHistorias.BuscarPorNumero(context.Background(), e.proyectoID, e.historia.Numero)
	if err != nil {
		return err
	}
	obtenida := "sin estimar"
	if h.StoryPoints != nil {
		obtenida = fmt.Sprint(*h.StoryPoints)
	}
	if obtenida != esperada {
		return fmt.Errorf("estimación: se esperaba %q y se obtuvo %q (último error: %v)", esperada, obtenida, e.err)
	}
	return nil
}

// laEstimacionDaComoResultado acepta "ok", "campo: mensaje" (error de validación) o el mensaje de otro error.
func (e *hu16) laEstimacionDaComoResultado(resultado string) error {
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
