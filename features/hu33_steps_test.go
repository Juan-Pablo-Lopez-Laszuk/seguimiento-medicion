package features

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// hu33 guarda el estado de un escenario de HU-33 (desviación de horas), con sus propios repositorios en memoria.
type hu33 struct {
	proyectos *memory.Proyectos
	sprints   *memory.Sprints
	historias *memory.Historias
	tareas    *memory.Tareas
	registros *memory.Registros
	horas     *service.Horas
	proyecto  int64
	sprint    int64
	ultima    int64 // la última historia creada
}

// registrarPasosHU33 conecta cada frase de features/hu-33-desviacion.feature con una función de Go.
func registrarPasosHU33(sc *godog.ScenarioContext) {
	e := &hu33{proyectos: memory.NuevoProyectos(), sprints: memory.NuevoSprints(), historias: memory.NuevoHistorias(),
		tareas: memory.NuevoTareas(), registros: memory.NuevoRegistros()}
	e.horas = service.NuevoHoras(e.proyectos, e.sprints, e.historias, e.tareas, e.registros)
	sc.Step(`^una historia con tareas estimadas en (\d+) h y (\d+) h$`, func(a, b float64) error { return e.historia(false, a, b) })
	sc.Step(`^una historia sin tareas estimadas$`, func() error { return e.historia(false) })
	sc.Step(`^se registran (\d+) h de trabajo en la historia$`, e.registrar)
	sc.Step(`^en el Sprint 1 una historia estimada en (\d+) h con (\d+) h trabajadas$`, func(est, reales float64) error {
		return e.historiaConHoras(true, est, reales)
	})
	sc.Step(`^sin sprint una historia estimada en (\d+) h con (\d+) h trabajadas$`, func(est, reales float64) error {
		return e.historiaConHoras(false, est, reales)
	})
	sc.Step(`^la desviación (de la historia|del Sprint 1|del proyecto) es "([^"]*)" y "([^"]*)"$`, e.laDesviacion)
}

// preparar crea el proyecto con su Sprint 1 la primera vez que el escenario lo necesita.
func (e *hu33) preparar(ctx context.Context) error {
	if e.proyecto != 0 {
		return nil
	}
	p, err := service.NuevoProyectos(e.proyectos, time.Now).Crear(ctx, project.Datos{Nombre: "Tienda", FechaInicio: fechaDeEscenario("2026-10-05"), FechaFin: fechaDeEscenario("2026-11-01")})
	if err != nil {
		return fmt.Errorf("no se pudo crear el proyecto: %w", err)
	}
	s, err := service.NuevoSprints(e.proyectos, e.sprints).Crear(ctx, p.ID, sprint.Datos{Objetivo: "MVP", FechaInicio: fechaDeEscenario("2026-10-05"), FechaFin: fechaDeEscenario("2026-10-11")})
	if err != nil {
		return fmt.Errorf("no se pudo crear el sprint: %w", err)
	}
	e.proyecto, e.sprint = p.ID, s.ID
	return nil
}

// historia crea una historia (en el Sprint 1 si enSprint) con una tarea por cada estimación.
func (e *hu33) historia(enSprint bool, estimaciones ...float64) error {
	ctx := context.Background()
	if err := e.preparar(ctx); err != nil {
		return err
	}
	h, err := service.NuevoHistorias(e.proyectos, e.historias).Crear(ctx, e.proyecto, backlog.Datos{Titulo: "Historia", Prioridad: backlog.PrioridadAlta})
	if err != nil {
		return fmt.Errorf("no se pudo crear la historia: %w", err)
	}
	e.ultima = h.ID
	if enSprint {
		if err := e.historias.ActualizarSprint(ctx, h.ID, backlog.EstadoEnSprint, &e.sprint); err != nil {
			return err
		}
	}
	for _, horas := range estimaciones {
		if _, err := e.tareas.Guardar(ctx, effort.Tarea{HistoriaID: h.ID, Titulo: "Tarea", HorasEstimadas: horas}); err != nil {
			return err
		}
	}
	return nil
}

func (e *hu33) historiaConHoras(enSprint bool, estimadas, reales float64) error {
	if err := e.historia(enSprint, estimadas); err != nil {
		return err
	}
	return e.registrar(reales)
}

func (e *hu33) registrar(horas float64) error {
	_, err := e.registros.Guardar(context.Background(), effort.Registro{HistoriaID: e.ultima, Horas: horas, Actividad: "Trabajo"})
	return err
}

func (e *hu33) laDesviacion(de, horas, porcentaje string) error {
	v, err := e.horas.Ver(context.Background(), e.proyecto)
	if err != nil {
		return err
	}
	var got service.FilaHoras
	switch de {
	case "de la historia":
		got = v.Historias[len(v.Historias)-1]
	case "del Sprint 1":
		got = v.Sprints[0]
	default:
		got = service.FilaHoras{Nombre: "proyecto", Desviacion: v.Total}
	}
	if got.HorasTexto() != horas || got.PorcentajeTexto() != porcentaje {
		return fmt.Errorf("la desviación %s es %q y %q, se esperaba %q y %q", de, got.HorasTexto(), got.PorcentajeTexto(), horas, porcentaje)
	}
	return nil
}
