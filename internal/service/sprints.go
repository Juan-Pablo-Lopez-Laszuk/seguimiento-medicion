package service

import (
	"context"
	"fmt"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
)

// BuscadorProyectos es lo que los casos de uso de sprints necesitan del repositorio de proyectos.
type BuscadorProyectos interface {
	// BuscarPorID devuelve el proyecto, o project.ErrNoEncontrado si no existe.
	BuscarPorID(ctx context.Context, id int64) (project.Proyecto, error)
}

// RepositorioSprints es lo que el caso de uso necesita de la base de datos. Lo implementan
// internal/store/memory (tests) e internal/store/postgres (Supabase).
type RepositorioSprints interface {
	// ListarPorProyecto devuelve los sprints del proyecto en el orden en que se crearon.
	ListarPorProyecto(ctx context.Context, proyectoID int64) ([]sprint.Sprint, error)
	// Guardar guarda el sprint y lo devuelve con su ID.
	Guardar(ctx context.Context, s sprint.Sprint) (sprint.Sprint, error)
}

// Sprints agrupa los casos de uso de la épica E3.
type Sprints struct {
	proyectos BuscadorProyectos
	repo      RepositorioSprints
}

// NuevoSprints arma el caso de uso.
func NuevoSprints(proyectos BuscadorProyectos, repo RepositorioSprints) *Sprints {
	return &Sprints{proyectos: proyectos, repo: repo}
}

// Crear valida los datos contra el proyecto y sus sprints (HU-11) y guarda el sprint. Los datos
// inválidos se devuelven juntos en domain.ErroresValidacion; si el proyecto no existe, el error es
// project.ErrNoEncontrado; cualquier otro error viene de los repositorios.
func (s *Sprints) Crear(ctx context.Context, proyectoID int64, d sprint.Datos) (sprint.Sprint, error) {
	p, existentes, err := s.Listar(ctx, proyectoID)
	if err != nil {
		return sprint.Sprint{}, err
	}
	nuevo, err := sprint.Nuevo(d, p, existentes)
	if err != nil {
		return sprint.Sprint{}, err
	}
	guardado, err := s.repo.Guardar(ctx, nuevo)
	if err != nil {
		return sprint.Sprint{}, fmt.Errorf("guardar el sprint: %w", err)
	}
	return guardado, nil
}

// Listar devuelve el proyecto y sus sprints en el orden en que se crearon.
func (s *Sprints) Listar(ctx context.Context, proyectoID int64) (project.Proyecto, []sprint.Sprint, error) {
	p, err := s.proyectos.BuscarPorID(ctx, proyectoID)
	if err != nil {
		return project.Proyecto{}, nil, fmt.Errorf("buscar el proyecto %d: %w", proyectoID, err)
	}
	lista, err := s.repo.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return project.Proyecto{}, nil, fmt.Errorf("listar los sprints del proyecto %d: %w", proyectoID, err)
	}
	return p, lista, nil
}
