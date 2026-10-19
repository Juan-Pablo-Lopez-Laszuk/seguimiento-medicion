package features

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/defect"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// proyectoConSprintEHistoria es un proyecto del escenario con su Sprint 1 y una historia.
type proyectoConSprintEHistoria struct {
	id       int64
	sprint   sprint.Sprint
	historia backlog.Historia
}

// hu27 guarda el estado de un escenario de HU-27 (registrar defecto), con sus propios repositorios en memoria.
type hu27 struct {
	proyectos *service.Proyectos
	sprints   *service.Sprints
	historias *service.Historias
	defectos  *service.Defectos
	principal proyectoConSprintEHistoria
	otro      proyectoConSprintEHistoria
	ultimo    defect.Defecto
	err       error
}

// registrarPasosHU27 conecta cada frase de features/hu-27-registrar-defecto.feature con una función de Go.
func registrarPasosHU27(sc *godog.ScenarioContext) {
	repoProyectos, repoSprints, repoHistorias := memory.NuevoProyectos(), memory.NuevoSprints(), memory.NuevoHistorias()
	e := &hu27{
		proyectos: service.NuevoProyectos(repoProyectos, time.Now),
		sprints:   service.NuevoSprints(repoProyectos, repoSprints),
		historias: service.NuevoHistorias(repoProyectos, repoHistorias),
		defectos:  service.NuevoDefectos(repoProyectos, repoSprints, repoHistorias, memory.NuevoDefectos()),
	}
	sc.Step(`^el proyecto "([^"]*)" con el Sprint 1 y la historia "([^"]*)"$`, func(nombre, titulo string) (err error) {
		e.principal, err = e.crear(nombre, titulo)
		return err
	})
	sc.Step(`^otro proyecto "([^"]*)" con su Sprint 1 y la historia "([^"]*)"$`, func(nombre, titulo string) (err error) {
		e.otro, err = e.crear(nombre, titulo)
		return err
	})
	sc.Step(`^registro el defecto "([^"]*)" de severidad "([^"]*)" en el Sprint 1 para la historia$`, func(desc, sev string) error {
		return e.registrar(desc, sev, e.principal.sprint.ID, e.principal.historia.ID)
	})
	sc.Step(`^registro el defecto "([^"]*)" de severidad "([^"]*)" en el Sprint 1$`, func(desc, sev string) error {
		return e.registrar(desc, sev, e.principal.sprint.ID, 0)
	})
	sc.Step(`^registro un defecto con el sprint y la historia del otro proyecto$`, func() error {
		return e.registrar("Defecto", string(defect.SeveridadAlta), e.otro.sprint.ID, e.otro.historia.ID)
	})
	sc.Step(`^el defecto queda "([^"]*)" con severidad "([^"]*)"$`, e.elDefectoQueda)
	sc.Step(`^el proyecto tiene (\d+) defectos?$`, e.elProyectoTiene)
	sc.Step(`^el defecto se rechaza con el error "([^"]*)"$`, e.seRechaza)
}

func (e *hu27) crear(nombre, titulo string) (proyectoConSprintEHistoria, error) {
	ctx := context.Background()
	p, err := e.proyectos.Crear(ctx, project.Datos{Nombre: nombre, FechaInicio: fechaDeEscenario("2026-10-05"), FechaFin: fechaDeEscenario("2026-11-01")})
	if err != nil {
		return proyectoConSprintEHistoria{}, fmt.Errorf("no se pudo crear el proyecto %q: %w", nombre, err)
	}
	r := proyectoConSprintEHistoria{id: p.ID}
	if r.sprint, err = e.sprints.Crear(ctx, p.ID, sprint.Datos{Objetivo: "MVP", FechaInicio: fechaDeEscenario("2026-10-05"), FechaFin: fechaDeEscenario("2026-10-11")}); err != nil {
		return r, fmt.Errorf("no se pudo crear el sprint: %w", err)
	}
	if r.historia, err = e.historias.Crear(ctx, p.ID, backlog.Datos{Titulo: titulo, Prioridad: backlog.PrioridadAlta}); err != nil {
		return r, fmt.Errorf("no se pudo crear la historia: %w", err)
	}
	return r, nil
}

func (e *hu27) registrar(descripcion, severidad string, sprintID, historiaID int64) error {
	e.ultimo, e.err = e.defectos.Registrar(context.Background(), e.principal.id,
		defect.Datos{Descripcion: descripcion, Severidad: defect.Severidad(severidad), SprintID: sprintID, HistoriaID: historiaID})
	return nil
}

func (e *hu27) elDefectoQueda(estado, severidad string) error {
	if e.err != nil {
		return fmt.Errorf("no se esperaba error: %w", e.err)
	}
	if string(e.ultimo.Estado) != estado || string(e.ultimo.Severidad) != severidad {
		return fmt.Errorf("el defecto quedó %q con severidad %q", e.ultimo.Estado, e.ultimo.Severidad)
	}
	return nil
}

func (e *hu27) elProyectoTiene(n int) error {
	v, err := e.defectos.Ver(context.Background(), e.principal.id)
	if err != nil {
		return err
	}
	if len(v.Defectos) != n {
		return fmt.Errorf("el proyecto tiene %d defectos, se esperaban %d", len(v.Defectos), n)
	}
	return nil
}

func (e *hu27) seRechaza(mensaje string) error {
	var errs domain.ErroresValidacion
	if !errors.As(e.err, &errs) {
		return fmt.Errorf("se esperaba el error %q y se obtuvo %v", mensaje, e.err)
	}
	for _, err := range errs {
		if err.Error() == mensaje {
			return nil
		}
	}
	return fmt.Errorf("errores = %v, se esperaba %q", errs, mensaje)
}
