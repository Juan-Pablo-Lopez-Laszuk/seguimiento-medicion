package features

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// hu04 guarda el estado de un escenario de HU-04 (consultar estado del proyecto), con sus propios repositorios
// en memoria. Los sprints se juntan en el escenario y se guardan recién al consultar, así un paso posterior
// puede marcar uno como cerrado (todavía no existe el cierre de HU-14).
type hu04 struct {
	proyectos  *memory.Proyectos
	sprints    *memory.Sprints
	historias  *memory.Historias
	proyectoID int64
	pendientes []sprint.Sprint
	hoy        time.Time
	estado     service.EstadoProyecto
}

// registrarPasosHU04 conecta cada frase de features/hu-04-estado-proyecto.feature con una función de Go.
func registrarPasosHU04(sc *godog.ScenarioContext) {
	e := &hu04{
		proyectos: memory.NuevoProyectos(),
		sprints:   memory.NuevoSprints(),
		historias: memory.NuevoHistorias(),
		hoy:       fechaDeEscenario("2026-10-13"), // si el escenario no dice qué día es
	}

	sc.Step(`^el proyecto "([^"]*)" del "([^"]*)" al "([^"]*)" para consultar su estado$`, e.elProyecto)
	sc.Step(`^que hoy es "([^"]*)"$`, e.queHoyEs)
	sc.Step(`^que el proyecto tiene el Sprint (\d+) del "([^"]*)" al "([^"]*)"$`, e.queElProyectoTieneElSprint)
	sc.Step(`^que el Sprint (\d+) está cerrado$`, e.queElSprintEstaCerrado)
	sc.Step(`^que el proyecto tiene historias con (\d+) SP Hechos de (\d+) SP estimados$`, e.queTieneHistorias)
	sc.Step(`^consulto el estado del proyecto$`, e.consultoElEstado)
	sc.Step(`^el estado es "([^"]*)"$`, e.elEstadoEs)
	sc.Step(`^no hay sprint activo$`, e.noHaySprintActivo)
	sc.Step(`^al proyecto le quedan (\d+) días$`, e.alProyectoLeQuedan)
	sc.Step(`^el sprint activo es el Sprint (\d+) y le quedan (\d+) días$`, e.elSprintActivoEs)
	sc.Step(`^se avisa que el sprint terminó y falta cerrarlo$`, e.seAvisaQueFaltaCerrarlo)
	sc.Step(`^el porcentaje de SP completados es "([^"]*)"$`, e.elPorcentajeDeSPEs)
}

func (e *hu04) elProyecto(nombre, inicio, fin string) error {
	p, err := service.NuevoProyectos(e.proyectos, time.Now).Crear(context.Background(), project.Datos{
		Nombre:      nombre,
		FechaInicio: fechaDeEscenario(inicio),
		FechaFin:    fechaDeEscenario(fin),
	})
	if err != nil {
		return fmt.Errorf("no se pudo crear el proyecto %q: %w", nombre, err)
	}
	e.proyectoID = p.ID
	return nil
}

func (e *hu04) queHoyEs(fecha string) error {
	e.hoy = fechaDeEscenario(fecha)
	return nil
}

func (e *hu04) queElProyectoTieneElSprint(numero int, inicio, fin string) error {
	e.pendientes = append(e.pendientes, sprint.Sprint{
		ProyectoID:  e.proyectoID,
		Numero:      numero,
		Objetivo:    fmt.Sprintf("Goal del Sprint %d", numero),
		FechaInicio: fechaDeEscenario(inicio),
		FechaFin:    fechaDeEscenario(fin),
		Estado:      sprint.EstadoPlanificado,
	})
	return nil
}

func (e *hu04) queElSprintEstaCerrado(numero int) error {
	for i := range e.pendientes {
		if e.pendientes[i].Numero == numero {
			e.pendientes[i].Estado = sprint.EstadoCerrado
			return nil
		}
	}
	return fmt.Errorf("el escenario no tiene el Sprint %d", numero)
}

// queTieneHistorias carga una historia Hecha con los SP hechos y otra Pendiente con el resto de los estimados.
func (e *hu04) queTieneHistorias(hechos, estimados int) error {
	for _, h := range []struct {
		sp     int
		estado backlog.Estado
	}{{hechos, backlog.EstadoHecho}, {estimados - hechos, backlog.EstadoPendiente}} {
		if h.sp == 0 {
			continue
		}
		sp := h.sp
		_, err := e.historias.Guardar(context.Background(), backlog.Historia{
			ProyectoID: e.proyectoID, Titulo: "Historia", Prioridad: backlog.PrioridadAlta, Estado: h.estado, StoryPoints: &sp,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *hu04) consultoElEstado() error {
	for _, s := range e.pendientes {
		if _, err := e.sprints.Guardar(context.Background(), s); err != nil {
			return err
		}
	}
	e.pendientes = nil
	// A las 15:00 de Argentina del día del escenario.
	ahora := func() time.Time { return e.hoy.Add(18 * time.Hour) }
	estado, err := service.NuevoEstadoProyectos(e.proyectos, e.sprints, e.historias, ahora).Consultar(context.Background(), e.proyectoID)
	e.estado = estado
	return err
}

func (e *hu04) elEstadoEs(esperado string) error {
	if string(e.estado.Estado) != esperado {
		return fmt.Errorf("se esperaba el estado %q y se obtuvo %q", esperado, e.estado.Estado)
	}
	return nil
}

func (e *hu04) noHaySprintActivo() error {
	if e.estado.Activo != nil {
		return fmt.Errorf("no se esperaba sprint activo y se obtuvo el Sprint %d", e.estado.Activo.Numero)
	}
	return nil
}

func (e *hu04) alProyectoLeQuedan(dias int) error {
	if e.estado.DiasRestantes != dias {
		return fmt.Errorf("se esperaban %d días y se obtuvieron %d", dias, e.estado.DiasRestantes)
	}
	return nil
}

func (e *hu04) elSprintActivoEs(numero, dias int) error {
	a := e.estado.Activo
	if a == nil {
		return fmt.Errorf("se esperaba el Sprint %d activo y no hay sprint activo", numero)
	}
	if a.Numero != numero || a.DiasRestantes != dias {
		return fmt.Errorf("se esperaba el Sprint %d con %d días y se obtuvo el Sprint %d con %d", numero, dias, a.Numero, a.DiasRestantes)
	}
	return nil
}

func (e *hu04) seAvisaQueFaltaCerrarlo() error {
	if e.estado.Activo == nil || !e.estado.Activo.Terminado {
		return fmt.Errorf("se esperaba un sprint activo terminado sin cerrar y se obtuvo %+v", e.estado.Activo)
	}
	return nil
}

func (e *hu04) elPorcentajeDeSPEs(esperado string) error {
	if got := e.estado.PorcentajeSP(); got != esperado {
		return fmt.Errorf("se esperaba %q y se obtuvo %q", esperado, got)
	}
	return nil
}
