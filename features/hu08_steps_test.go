package features

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// hu08 guarda el estado de un escenario de HU-08 (cambiar el estado de una historia), con sus propios repositorios en
// memoria. "La historia" de todos los pasos es la del Antecedentes.
type hu08 struct {
	repoHistorias *memory.Historias
	repoCriterios *memory.Criterios
	historias     *service.Historias
	estados       *service.Estados
	proyectoID    int64
	historia      backlog.Historia
	err           error
}

// registrarPasosHU08 conecta cada frase de features/hu-08-estado-historia.feature con una función de Go.
func registrarPasosHU08(sc *godog.ScenarioContext) {
	repoProyectos := memory.NuevoProyectos()
	e := &hu08{repoHistorias: memory.NuevoHistorias(), repoCriterios: memory.NuevoCriterios()}
	e.historias = service.NuevoHistorias(repoProyectos, e.repoHistorias)
	e.estados = service.NuevoEstados(repoProyectos, e.repoHistorias, e.repoCriterios)
	proyectos := service.NuevoProyectos(repoProyectos, time.Now)

	sc.Step(`^el proyecto "([^"]*)" con la historia "([^"]*)" para cambiar su estado$`, func(nombre, titulo string) error {
		return e.elProyectoConLaHistoria(proyectos, nombre, titulo)
	})
	sc.Step(`^que la historia está "([^"]*)"$`, func(estado string) error { return e.queLaHistoriaEsta(estado, 0, 0) })
	sc.Step(`^que la historia está "([^"]*)" y tiene (\d+) criterios?, (\d+) de ellos cumplidos$`, e.queLaHistoriaEsta)
	sc.Step(`^la paso a "([^"]*)"$`, e.laPasoA)
	sc.Step(`^el estado de la historia es "([^"]*)"$`, e.elEstadoDeLaHistoriaEs)
	sc.Step(`^el cambio da como resultado "([^"]*)"$`, e.elCambioDaComoResultado)
}

func (e *hu08) elProyectoConLaHistoria(proyectos *service.Proyectos, nombre, titulo string) error {
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

// queLaHistoriaEsta deja la historia en ese estado y con esos criterios, directo en los repositorios: los cambios de
// estado de HU-12 y HU-13 todavía no existen, y varios escenarios arrancan desde esos estados.
func (e *hu08) queLaHistoriaEsta(estado string, total, cumplidos int) error {
	ctx := context.Background()
	if err := e.repoHistorias.ActualizarEstado(ctx, e.historia.ID, backlog.Estado(estado)); err != nil {
		return fmt.Errorf("no se pudo dejar la historia en %q: %w", estado, err)
	}
	for i := 1; i <= total; i++ {
		c := backlog.Criterio{HistoriaID: e.historia.ID, Descripcion: fmt.Sprintf("Criterio %d", i), Cumplido: i <= cumplidos}
		if _, err := e.repoCriterios.Guardar(ctx, c); err != nil {
			return fmt.Errorf("no se pudo guardar el criterio %d: %w", i, err)
		}
	}
	return nil
}

func (e *hu08) laPasoA(estado string) error {
	_, e.err = e.estados.Cambiar(context.Background(), e.proyectoID, e.historia.Numero, backlog.Estado(estado))
	return nil
}

func (e *hu08) elEstadoDeLaHistoriaEs(esperado string) error {
	h, err := e.repoHistorias.BuscarPorNumero(context.Background(), e.proyectoID, e.historia.Numero)
	if err != nil {
		return err
	}
	if string(h.Estado) != esperado {
		return fmt.Errorf("estado: se esperaba %q y se obtuvo %q (último error: %v)", esperado, h.Estado, e.err)
	}
	return nil
}

// elCambioDaComoResultado acepta "ok" o el mensaje del error.
func (e *hu08) elCambioDaComoResultado(resultado string) error {
	if resultado == "ok" {
		if e.err != nil {
			return fmt.Errorf("se esperaba que el cambio valga y se obtuvo: %w", e.err)
		}
		return nil
	}
	if e.err == nil {
		return fmt.Errorf("se esperaba el error %q y el cambio valió", resultado)
	}
	if e.err.Error() != resultado {
		return fmt.Errorf("se esperaba el error %q y se obtuvo %q", resultado, e.err.Error())
	}
	return nil
}
