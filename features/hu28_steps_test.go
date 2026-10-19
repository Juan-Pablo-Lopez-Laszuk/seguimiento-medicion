package features

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/defect"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// hu28 guarda el estado de un escenario de HU-28 (estado de un defecto), con sus propios repositorios en memoria.
type hu28 struct {
	proyectos *service.Proyectos
	sprints   *service.Sprints
	defectos  *service.Defectos
	proyecto  int64
	numeros   map[int]int64 // número de sprint → ID
	defecto   defect.Defecto
	err       error
}

// registrarPasosHU28 conecta cada frase de features/hu-28-estado-defecto.feature con una función de Go.
func registrarPasosHU28(sc *godog.ScenarioContext) {
	repoProyectos, repoSprints := memory.NuevoProyectos(), memory.NuevoSprints()
	e := &hu28{
		proyectos: service.NuevoProyectos(repoProyectos, time.Now),
		sprints:   service.NuevoSprints(repoProyectos, repoSprints),
		defectos:  service.NuevoDefectos(repoProyectos, repoSprints, memory.NuevoHistorias(), memory.NuevoDefectos()),
		numeros:   map[int]int64{},
	}
	sc.Step(`^el proyecto con los sprints 1 y 2 y un defecto Abierto detectado en el Sprint 1$`, e.preparar)
	sc.Step(`^paso el defecto a "([^"]*)"$`, func(estado string) error {
		e.cambiar(estado, 0)
		return nil
	})
	sc.Step(`^lo resuelvo en el Sprint (\d+)$`, func(n int) error {
		e.cambiar(string(defect.EstadoResuelto), e.numeros[n])
		return nil
	})
	sc.Step(`^que el defecto está "En progreso"$`, func() error {
		e.cambiar(string(defect.EstadoEnProgreso), 0)
		if e.err != nil {
			return fmt.Errorf("no se pudo pasar a En progreso: %w", e.err)
		}
		return nil
	})
	sc.Step(`^el defecto queda "([^"]*)" con sprint de resolución "Sprint (\d+)"$`, func(estado string, n int) error {
		return e.queda(estado, e.numeros[n])
	})
	sc.Step(`^el defecto queda "([^"]*)" sin sprint de resolución$`, func(estado string) error { return e.queda(estado, 0) })
	sc.Step(`^el cambio se rechaza con el error "([^"]*)"$`, func(esperado string) error {
		if got := mensajeDelError(e.err); got != esperado {
			return fmt.Errorf("error = %q, se esperaba %q", got, esperado)
		}
		return nil
	})
}

func (e *hu28) preparar() error {
	ctx := context.Background()
	p, err := e.proyectos.Crear(ctx, project.Datos{Nombre: "Tienda", FechaInicio: fechaDeEscenario("2026-10-05"), FechaFin: fechaDeEscenario("2026-11-01")})
	if err != nil {
		return fmt.Errorf("no se pudo crear el proyecto: %w", err)
	}
	e.proyecto = p.ID
	for _, f := range [][2]string{{"2026-10-05", "2026-10-11"}, {"2026-10-12", "2026-10-18"}} {
		s, err := e.sprints.Crear(ctx, p.ID, sprint.Datos{Objetivo: "MVP", FechaInicio: fechaDeEscenario(f[0]), FechaFin: fechaDeEscenario(f[1])})
		if err != nil {
			return fmt.Errorf("no se pudo crear el sprint: %w", err)
		}
		e.numeros[s.Numero] = s.ID
	}
	e.defecto, err = e.defectos.Registrar(ctx, p.ID, defect.Datos{Descripcion: "Falla", Severidad: defect.SeveridadAlta, SprintID: e.numeros[1]})
	return err
}

// cambiar guarda el error en e.err para que lo revise el paso Entonces.
func (e *hu28) cambiar(estado string, sprintResolucion int64) {
	d, err := e.defectos.CambiarEstado(context.Background(), e.proyecto, e.defecto.ID, defect.Estado(estado), sprintResolucion)
	if e.err = err; err == nil {
		e.defecto = d
	}
}

// queda revisa el defecto guardado, no el que devolvió el último cambio.
func (e *hu28) queda(estado string, sprintResolucion int64) error {
	v, err := e.defectos.Ver(context.Background(), e.proyecto)
	if err != nil {
		return err
	}
	d := v.Defectos[0]
	if string(d.Estado) != estado || d.SprintResolucionID != sprintResolucion {
		return fmt.Errorf("el defecto quedó %q con sprint de resolución %d (se esperaba %q y %d)", d.Estado, d.SprintResolucionID, estado, sprintResolucion)
	}
	return nil
}
