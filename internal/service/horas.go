package service

import (
	"context"
	"fmt"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/metrics"
)

// ListadorRegistros es lo que la desviación necesita del repositorio de registros de esfuerzo (HU-33).
type ListadorRegistros interface {
	ListarPorHistoria(ctx context.Context, historiaID int64) ([]effort.Registro, error)
}

// Horas arma la comparación de horas estimadas y reales de un proyecto (HU-33). Solo lee: no cambia nada.
type Horas struct {
	proyectos BuscadorProyectos
	sprints   ListadorSprints
	historias ListadorHistorias
	tareas    ListadorTareas
	registros ListadorRegistros
}

// NuevoHoras arma el caso de uso.
func NuevoHoras(proyectos BuscadorProyectos, sprints ListadorSprints, historias ListadorHistorias, tareas ListadorTareas, registros ListadorRegistros) *Horas {
	return &Horas{proyectos: proyectos, sprints: sprints, historias: historias, tareas: tareas, registros: registros}
}

// FilaHoras es una fila de la página de horas: una historia o un sprint con su desviación (M6).
type FilaHoras struct {
	Nombre string
	metrics.Desviacion
}

// VistaHoras es lo que muestra la página de horas del proyecto.
type VistaHoras struct {
	Proyecto  project.Proyecto
	Total     metrics.Desviacion // del proyecto: suma de todas sus historias (CA-33.3)
	Sprints   []FilaHoras        // en el orden de los sprints
	Historias []FilaHoras        // en el orden del backlog
}

// Ver calcula la desviación de cada historia, de cada sprint (sus historias asignadas) y del proyecto.
func (s *Horas) Ver(ctx context.Context, proyectoID int64) (VistaHoras, error) {
	p, err := s.proyectos.BuscarPorID(ctx, proyectoID)
	if err != nil {
		return VistaHoras{}, fmt.Errorf("buscar el proyecto %d: %w", proyectoID, err)
	}
	sprints, err := s.sprints.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return VistaHoras{}, fmt.Errorf("listar los sprints del proyecto %d: %w", proyectoID, err)
	}
	historias, err := s.historias.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return VistaHoras{}, fmt.Errorf("listar las historias del proyecto %d: %w", proyectoID, err)
	}

	v := VistaHoras{Proyecto: p}
	todas := make([]metrics.Desviacion, 0, len(historias))
	porSprint := map[int64][]metrics.Desviacion{}
	for _, h := range historias {
		d, err := s.desviacionDe(ctx, h.ID)
		if err != nil {
			return VistaHoras{}, err
		}
		v.Historias = append(v.Historias, FilaHoras{Nombre: h.Identificador() + " · " + h.Titulo, Desviacion: d})
		todas = append(todas, d)
		if h.SprintID != nil {
			porSprint[*h.SprintID] = append(porSprint[*h.SprintID], d)
		}
	}
	for _, sp := range sprints {
		v.Sprints = append(v.Sprints, FilaHoras{Nombre: fmt.Sprintf("Sprint %d", sp.Numero), Desviacion: metrics.SumarDesviaciones(porSprint[sp.ID])})
	}
	v.Total = metrics.SumarDesviaciones(todas)
	return v, nil
}

// desviacionDe compara las horas estimadas de la historia (suma de sus tareas, M4) con las reales (sus registros, M5).
func (s *Horas) desviacionDe(ctx context.Context, historiaID int64) (metrics.Desviacion, error) {
	tareas, err := s.tareas.ListarPorHistoria(ctx, historiaID)
	if err != nil {
		return metrics.Desviacion{}, fmt.Errorf("listar las tareas de la historia %d: %w", historiaID, err)
	}
	estimadas, err := effort.HorasEstimadasDeHistoria(0, tareas)
	if err != nil {
		return metrics.Desviacion{}, fmt.Errorf("horas estimadas de la historia %d: %w", historiaID, err)
	}
	registros, err := s.registros.ListarPorHistoria(ctx, historiaID)
	if err != nil {
		return metrics.Desviacion{}, fmt.Errorf("listar los registros de la historia %d: %w", historiaID, err)
	}
	return metrics.CalcularDesviacion(estimadas, effort.HorasReales(registros)), nil
}
