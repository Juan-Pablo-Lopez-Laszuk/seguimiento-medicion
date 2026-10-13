package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/metrics"
)

// ListadorProyectos es lo que el estado necesita del repositorio de proyectos: uno por ID para la vista y
// todos para la lista.
type ListadorProyectos interface {
	BuscadorProyectos
	Listar(ctx context.Context) ([]project.Proyecto, error)
}

// ListadorSprints es la parte de lectura de RepositorioSprints.
type ListadorSprints interface {
	ListarPorProyecto(ctx context.Context, proyectoID int64) ([]sprint.Sprint, error)
}

// ListadorHistorias es la parte de lectura de RepositorioHistorias.
type ListadorHistorias interface {
	ListarPorProyecto(ctx context.Context, proyectoID int64) ([]backlog.Historia, error)
}

// EstadoProyectos arma la vista del estado de un proyecto (HU-04). Solo lee: no cambia nada (RN5).
type EstadoProyectos struct {
	proyectos ListadorProyectos
	sprints   ListadorSprints
	historias ListadorHistorias
	ahora     func() time.Time
}

// NuevoEstadoProyectos arma el caso de uso. ahora devuelve la fecha y hora actual (en los tests, una fija).
func NuevoEstadoProyectos(proyectos ListadorProyectos, sprints ListadorSprints, historias ListadorHistorias, ahora func() time.Time) *EstadoProyectos {
	return &EstadoProyectos{proyectos: proyectos, sprints: sprints, historias: historias, ahora: ahora}
}

// EstadoProyecto es lo que muestra la vista del proyecto (spec: specs/HU-04-estado-proyecto.md, sección 3).
type EstadoProyecto struct {
	Proyecto      project.Proyecto
	Estado        project.Estado // calculado (RN1), no el guardado
	DiasRestantes int            // del proyecto (RN3)
	Activo        *SprintActivo  // nil si no hay sprint activo (RN2)
	Proximo       *sprint.Sprint // el que sigue, si no hay activo
	StoryPoints   metrics.StoryPoints
}

// SprintActivo es el sprint activo con los días que le quedan. Terminado indica que su fecha de fin ya pasó
// y falta cerrarlo.
type SprintActivo struct {
	sprint.Sprint
	DiasRestantes int
	Terminado     bool
}

// ResumenProyecto es una fila de la lista de proyectos, con el estado calculado.
type ResumenProyecto struct {
	Proyecto project.Proyecto
	Estado   project.Estado
}

// PorcentajeSP devuelve el porcentaje de story points completados como se muestra en pantalla: con 1 decimal y
// coma ("37,5 %"). Si no hay SP estimados devuelve "Sin historias estimadas": nunca se divide por cero (RN4).
func (e EstadoProyecto) PorcentajeSP() string {
	sp := e.StoryPoints
	if sp.Planificados == 0 {
		return "Sin historias estimadas"
	}
	return strings.Replace(fmt.Sprintf("%.1f %%", sp.Completados/sp.Planificados*100), ".", ",", 1)
}

// Consultar devuelve el estado del proyecto. Si no existe, el error es project.ErrNoEncontrado; cualquier
// otro error viene de los repositorios.
func (s *EstadoProyectos) Consultar(ctx context.Context, proyectoID int64) (EstadoProyecto, error) {
	p, err := s.proyectos.BuscarPorID(ctx, proyectoID)
	if err != nil {
		return EstadoProyecto{}, fmt.Errorf("buscar el proyecto %d: %w", proyectoID, err)
	}
	sprints, err := s.listarSprints(ctx, proyectoID)
	if err != nil {
		return EstadoProyecto{}, err
	}
	historias, err := s.historias.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return EstadoProyecto{}, fmt.Errorf("listar las historias del proyecto %d: %w", proyectoID, err)
	}
	sp, err := metrics.CalcularStoryPoints(paraMetricas(historias))
	if err != nil {
		return EstadoProyecto{}, fmt.Errorf("calcular los story points del proyecto %d: %w", proyectoID, err)
	}

	hoy := domain.HoyEnArgentina(s.ahora()) // RN6
	estado := EstadoProyecto{
		Proyecto:      p,
		Estado:        project.EstadoActual(p, sprint.InicioDelPrimero(sprints), hoy),
		DiasRestantes: domain.DiasRestantes(hoy, p.FechaFin),
		StoryPoints:   sp,
	}
	if activo, ok := sprint.Activo(sprints, hoy); ok {
		estado.Activo = &SprintActivo{
			Sprint:        activo,
			DiasRestantes: domain.DiasRestantes(hoy, activo.FechaFin),
			Terminado:     activo.Terminado(hoy),
		}
	} else if proximo, ok := sprint.Proximo(sprints, hoy); ok {
		estado.Proximo = &proximo
	}
	return estado, nil
}

// Listar devuelve todos los proyectos, en el orden en que se crearon, con su estado calculado (RN1).
func (s *EstadoProyectos) Listar(ctx context.Context) ([]ResumenProyecto, error) {
	proyectos, err := s.proyectos.Listar(ctx)
	if err != nil {
		return nil, fmt.Errorf("listar los proyectos: %w", err)
	}
	hoy := domain.HoyEnArgentina(s.ahora())
	lista := make([]ResumenProyecto, 0, len(proyectos))
	for _, p := range proyectos {
		sprints, err := s.listarSprints(ctx, p.ID)
		if err != nil {
			return nil, err
		}
		lista = append(lista, ResumenProyecto{Proyecto: p, Estado: project.EstadoActual(p, sprint.InicioDelPrimero(sprints), hoy)})
	}
	return lista, nil
}

func (s *EstadoProyectos) listarSprints(ctx context.Context, proyectoID int64) ([]sprint.Sprint, error) {
	sprints, err := s.sprints.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("listar los sprints del proyecto %d: %w", proyectoID, err)
	}
	return sprints, nil
}

// paraMetricas pasa las historias del backlog a los datos simples que reciben las métricas (decisión del #80).
func paraMetricas(historias []backlog.Historia) []metrics.Historia {
	lista := make([]metrics.Historia, 0, len(historias))
	for _, h := range historias {
		m := metrics.Historia{Hecha: h.Estado == backlog.EstadoHecho, SinEstimar: h.StoryPoints == nil}
		if h.StoryPoints != nil {
			m.SP = float64(*h.StoryPoints)
		}
		lista = append(lista, m)
	}
	return lista
}
