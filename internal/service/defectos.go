package service

import (
	"context"
	"fmt"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/defect"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
)

// Los sprints y las historias del proyecto se piden con ListadorSprints y ListadorHistorias (de HU-04).

// RepositorioDefectos es lo que el caso de uso necesita de la base de datos (HU-27).
type RepositorioDefectos interface {
	// Guardar guarda un defecto nuevo y lo devuelve con su ID.
	Guardar(ctx context.Context, d defect.Defecto) (defect.Defecto, error)
	// ListarPorProyecto devuelve los defectos del proyecto, el más nuevo primero.
	ListarPorProyecto(ctx context.Context, proyectoID int64) ([]defect.Defecto, error)
}

// Defectos agrupa los casos de uso de los defectos de un proyecto (HU-27).
type Defectos struct {
	proyectos BuscadorProyectos
	sprints   ListadorSprints
	historias ListadorHistorias
	repo      RepositorioDefectos
}

// NuevoDefectos arma el caso de uso.
func NuevoDefectos(proyectos BuscadorProyectos, sprints ListadorSprints, historias ListadorHistorias, repo RepositorioDefectos) *Defectos {
	return &Defectos{proyectos: proyectos, sprints: sprints, historias: historias, repo: repo}
}

// VistaDefectos es lo que muestra la página de defectos: el proyecto, sus defectos y sus sprints e historias.
type VistaDefectos struct {
	Proyecto  project.Proyecto
	Defectos  []defect.Defecto
	Sprints   []sprint.Sprint
	Historias []backlog.Historia
}

// SprintDe devuelve el nombre del sprint con ese ID ("Sprint 1"), o "" si no es del proyecto.
func (v VistaDefectos) SprintDe(id int64) string {
	for _, s := range v.Sprints {
		if s.ID == id {
			return fmt.Sprintf("Sprint %d", s.Numero)
		}
	}
	return ""
}

// HistoriaDe devuelve "HU-N · título" de la historia con ese ID, o "" si el defecto no es de una historia.
func (v VistaDefectos) HistoriaDe(id int64) string {
	for _, h := range v.Historias {
		if h.ID == id {
			return h.Identificador() + " · " + h.Titulo
		}
	}
	return ""
}

// Ver devuelve el proyecto con sus defectos, sprints e historias, o project.ErrNoEncontrado.
func (s *Defectos) Ver(ctx context.Context, proyectoID int64) (VistaDefectos, error) {
	v, err := s.contexto(ctx, proyectoID)
	if err != nil {
		return VistaDefectos{}, err
	}
	if v.Defectos, err = s.repo.ListarPorProyecto(ctx, proyectoID); err != nil {
		return VistaDefectos{}, fmt.Errorf("listar los defectos del proyecto %d: %w", proyectoID, err)
	}
	return v, nil
}

// Registrar valida los datos (HU-27) contra los sprints y las historias del proyecto y guarda el defecto Abierto. Los
// datos inválidos se devuelven en domain.ErroresValidacion; si el proyecto no existe, project.ErrNoEncontrado.
func (s *Defectos) Registrar(ctx context.Context, proyectoID int64, d defect.Datos) (defect.Defecto, error) {
	v, err := s.contexto(ctx, proyectoID)
	if err != nil {
		return defect.Defecto{}, err
	}
	nuevo, err := defect.Nuevo(d, proyectoID, v.Sprints, v.Historias)
	if err != nil {
		return defect.Defecto{}, err
	}
	guardado, err := s.repo.Guardar(ctx, nuevo)
	if err != nil {
		return defect.Defecto{}, fmt.Errorf("guardar el defecto: %w", err)
	}
	return guardado, nil
}

func (s *Defectos) contexto(ctx context.Context, proyectoID int64) (VistaDefectos, error) {
	p, err := s.proyectos.BuscarPorID(ctx, proyectoID)
	if err != nil {
		return VistaDefectos{}, fmt.Errorf("buscar el proyecto %d: %w", proyectoID, err)
	}
	sprints, err := s.sprints.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return VistaDefectos{}, fmt.Errorf("listar los sprints del proyecto %d: %w", proyectoID, err)
	}
	historias, err := s.historias.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return VistaDefectos{}, fmt.Errorf("listar las historias del proyecto %d: %w", proyectoID, err)
	}
	return VistaDefectos{Proyecto: p, Sprints: sprints, Historias: historias}, nil
}
