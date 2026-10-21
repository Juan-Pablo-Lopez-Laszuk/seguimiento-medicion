package features

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/defect"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// hu35 guarda el estado de un escenario de HU-35 (panel), con sus propios repositorios en memoria y hoy fijo en el
// 14/10/2026 (el Sprint 2 del escenario está activo).
type hu35 struct {
	proyectos *memory.Proyectos
	sprints   *memory.Sprints
	historias *memory.Historias
	tareas    *memory.Tareas
	registros *memory.Registros
	defectos  *service.Defectos
	panel     *service.Panel
	proyecto  int64
	sprint1   int64
	vista     service.VistaPanel
}

var nombresDeTarjetas = regexp.MustCompile(`"([^"]*)"`)

// registrarPasosHU35 conecta cada frase de features/hu-35-panel.feature con una función de Go.
func registrarPasosHU35(sc *godog.ScenarioContext) {
	e := &hu35{proyectos: memory.NuevoProyectos(), sprints: memory.NuevoSprints(), historias: memory.NuevoHistorias(),
		tareas: memory.NuevoTareas(), registros: memory.NuevoRegistros()}
	hoy := func() time.Time { return time.Date(2026, 10, 14, 15, 0, 0, 0, time.UTC) }
	e.defectos = service.NuevoDefectos(e.proyectos, e.sprints, e.historias, memory.NuevoDefectos())
	e.panel = service.NuevoPanel(service.NuevoEstadoProyectos(e.proyectos, e.sprints, e.historias, hoy), e.sprints,
		service.NuevoHoras(e.proyectos, e.sprints, e.historias, e.tareas, e.registros), e.defectos)

	sc.Step(`^un proyecto recién creado$`, e.crearProyecto)
	sc.Step(`^un proyecto en curso con el Sprint 1 cerrado con (\d+) SP completados y el Sprint 2 activo$`, e.conSprints)
	sc.Step(`^una historia de (\d+) SP Hecha y otra de (\d+) SP pendiente$`, e.conHistorias)
	sc.Step(`^una historia con (\d+) h estimadas y (\d+) h trabajadas$`, e.conHoras)
	sc.Step(`^(\d+) defectos detectados y (\d+) resuelto$`, e.conDefectos)
	sc.Step(`^abro el panel$`, e.abrir)
	sc.Step(`^la tarjeta "([^"]*)" muestra "([^"]*)"$`, e.laTarjetaMuestra)
	sc.Step(`^las tarjetas (.+) muestran "([^"]*)"$`, func(nombres, valor string) error {
		for _, m := range nombresDeTarjetas.FindAllStringSubmatch(nombres, -1) {
			if err := e.laTarjetaMuestra(m[1], valor); err != nil {
				return err
			}
		}
		return nil
	})
}

func (e *hu35) crearProyecto() error {
	p, err := service.NuevoProyectos(e.proyectos, time.Now).Crear(context.Background(),
		project.Datos{Nombre: "Tienda", FechaInicio: fechaDeEscenario("2026-10-05"), FechaFin: fechaDeEscenario("2026-11-01")})
	if err != nil {
		return fmt.Errorf("no se pudo crear el proyecto: %w", err)
	}
	e.proyecto = p.ID
	return nil
}

func (e *hu35) conSprints(sp int) error {
	ctx := context.Background()
	if err := e.crearProyecto(); err != nil {
		return err
	}
	casos := service.NuevoSprints(e.proyectos, e.sprints)
	s1, err := casos.Crear(ctx, e.proyecto, sprint.Datos{Objetivo: "MVP", FechaInicio: fechaDeEscenario("2026-10-05"), FechaFin: fechaDeEscenario("2026-10-11")})
	if err != nil {
		return err
	}
	if _, err := casos.Crear(ctx, e.proyecto, sprint.Datos{Objetivo: "Métricas", FechaInicio: fechaDeEscenario("2026-10-12"), FechaFin: fechaDeEscenario("2026-10-18")}); err != nil {
		return err
	}
	e.sprint1, s1.SPCompletados = s1.ID, &sp
	return memory.NuevoCierres(e.sprints, e.historias).Cerrar(ctx, sprint.Cierre{Sprint: s1})
}

// historia crea una historia del proyecto con esos story points.
func (e *hu35) historia(sp int) (backlog.Historia, error) {
	ctx := context.Background()
	h, err := service.NuevoHistorias(e.proyectos, e.historias).Crear(ctx, e.proyecto, backlog.Datos{Titulo: "Historia", Prioridad: backlog.PrioridadAlta})
	if err != nil {
		return h, err
	}
	if sp > 0 {
		err = e.historias.ActualizarStoryPoints(ctx, h.ID, sp)
	}
	return h, err
}

func (e *hu35) conHistorias(hecha, pendiente int) error {
	h, err := e.historia(hecha)
	if err != nil {
		return err
	}
	if err := e.historias.ActualizarCompletada(context.Background(), h.ID, backlog.EstadoHecho, nil); err != nil {
		return err
	}
	_, err = e.historia(pendiente)
	return err
}

func (e *hu35) conHoras(estimadas, reales float64) error {
	ctx := context.Background()
	h, err := e.historia(0)
	if err != nil {
		return err
	}
	if _, err := e.tareas.Guardar(ctx, effort.Tarea{HistoriaID: h.ID, Titulo: "Tarea", HorasEstimadas: estimadas}); err != nil {
		return err
	}
	_, err = e.registros.Guardar(ctx, effort.Registro{HistoriaID: h.ID, Horas: reales, Actividad: "Trabajo"})
	return err
}

func (e *hu35) conDefectos(detectados, resueltos int) error {
	ctx := context.Background()
	for i := 0; i < detectados; i++ {
		d, err := e.defectos.Registrar(ctx, e.proyecto, defect.Datos{Descripcion: "Falla", Severidad: defect.SeveridadMedia, SprintID: e.sprint1})
		if err != nil {
			return err
		}
		if i >= resueltos {
			continue
		}
		for _, paso := range []defect.Estado{defect.EstadoEnProgreso, defect.EstadoResuelto} {
			if _, err := e.defectos.CambiarEstado(ctx, e.proyecto, d.ID, paso, e.sprint1); err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *hu35) abrir() (err error) {
	e.vista, err = e.panel.Ver(context.Background(), e.proyecto)
	return err
}

func (e *hu35) laTarjetaMuestra(titulo, valor string) error {
	for _, t := range e.vista.Tarjetas() {
		if t.Titulo == titulo {
			if t.Valor != valor {
				return fmt.Errorf("la tarjeta %q muestra %q, se esperaba %q", titulo, t.Valor, valor)
			}
			return nil
		}
	}
	return fmt.Errorf("el panel no tiene la tarjeta %q", titulo)
}
