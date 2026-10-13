package service

import (
	"context"
	"fmt"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
)

// RepositorioHistorias es lo que el caso de uso necesita de la base de datos. Lo implementan
// internal/store/memory (tests) e internal/store/postgres (Supabase).
type RepositorioHistorias interface {
	// ListarPorProyecto devuelve las historias del proyecto en el orden en que se crearon.
	ListarPorProyecto(ctx context.Context, proyectoID int64) ([]backlog.Historia, error)
	// Guardar guarda la historia y la devuelve con su ID.
	Guardar(ctx context.Context, h backlog.Historia) (backlog.Historia, error)
}

// Historias agrupa los casos de uso del Product Backlog (épica E2).
type Historias struct {
	proyectos BuscadorProyectos
	repo      RepositorioHistorias
}

// NuevoHistorias arma el caso de uso.
func NuevoHistorias(proyectos BuscadorProyectos, repo RepositorioHistorias) *Historias {
	return &Historias{proyectos: proyectos, repo: repo}
}

// Crear valida los datos contra las historias del proyecto (HU-05) y guarda la historia. Los datos inválidos se
// devuelven juntos en domain.ErroresValidacion; si el proyecto no existe, el error es project.ErrNoEncontrado;
// cualquier otro error viene de los repositorios.
func (s *Historias) Crear(ctx context.Context, proyectoID int64, d backlog.Datos) (backlog.Historia, error) {
	_, existentes, err := s.Listar(ctx, proyectoID)
	if err != nil {
		return backlog.Historia{}, err
	}
	nueva, err := backlog.Nueva(d, proyectoID, existentes)
	if err != nil {
		return backlog.Historia{}, err
	}
	guardada, err := s.repo.Guardar(ctx, nueva)
	if err != nil {
		return backlog.Historia{}, fmt.Errorf("guardar la historia: %w", err)
	}
	return guardada, nil
}

// Listar devuelve el proyecto y sus historias en el orden en que se crearon.
func (s *Historias) Listar(ctx context.Context, proyectoID int64) (project.Proyecto, []backlog.Historia, error) {
	p, err := s.proyectos.BuscarPorID(ctx, proyectoID)
	if err != nil {
		return project.Proyecto{}, nil, fmt.Errorf("buscar el proyecto %d: %w", proyectoID, err)
	}
	lista, err := s.repo.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return project.Proyecto{}, nil, fmt.Errorf("listar las historias del proyecto %d: %w", proyectoID, err)
	}
	return p, lista, nil
}
