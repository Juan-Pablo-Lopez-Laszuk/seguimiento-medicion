package service

import (
	"context"
	"fmt"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
)

// RepositorioIntegrantes es lo que los casos de uso necesitan de la base de datos. Lo implementan
// internal/store/memory (tests) e internal/store/postgres (Supabase).
type RepositorioIntegrantes interface {
	// ListarPorProyecto devuelve los integrantes del proyecto (activos y dados de baja) en el orden en
	// que se registraron.
	ListarPorProyecto(ctx context.Context, proyectoID int64) ([]member.Integrante, error)
	// Guardar guarda un integrante nuevo y lo devuelve con su ID.
	Guardar(ctx context.Context, i member.Integrante) (member.Integrante, error)
	// BuscarPorID devuelve el integrante dentro del proyecto, o member.ErrNoEncontrado.
	BuscarPorID(ctx context.Context, proyectoID, id int64) (member.Integrante, error)
	// Actualizar guarda los cambios de un integrante que ya existe.
	Actualizar(ctx context.Context, i member.Integrante) error
}

// Integrantes agrupa los casos de uso de los integrantes de un proyecto (HU-03).
type Integrantes struct {
	proyectos BuscadorProyectos
	repo      RepositorioIntegrantes
}

// NuevoIntegrantes arma el caso de uso.
func NuevoIntegrantes(proyectos BuscadorProyectos, repo RepositorioIntegrantes) *Integrantes {
	return &Integrantes{proyectos: proyectos, repo: repo}
}

// Registrar valida los datos contra los integrantes que ya tiene el proyecto y guarda el integrante.
// Los datos inválidos se devuelven juntos en domain.ErroresValidacion; si el proyecto no existe, el
// error es project.ErrNoEncontrado; cualquier otro error viene de los repositorios.
func (s *Integrantes) Registrar(ctx context.Context, proyectoID int64, d member.Datos) (member.Integrante, error) {
	_, existentes, err := s.Listar(ctx, proyectoID)
	if err != nil {
		return member.Integrante{}, err
	}
	nuevo, err := member.Nuevo(d, proyectoID, existentes)
	if err != nil {
		return member.Integrante{}, err
	}
	guardado, err := s.repo.Guardar(ctx, nuevo)
	if err != nil {
		return member.Integrante{}, fmt.Errorf("guardar el integrante: %w", err)
	}
	return guardado, nil
}

// Listar devuelve el proyecto y todos sus integrantes, activos y dados de baja.
func (s *Integrantes) Listar(ctx context.Context, proyectoID int64) (project.Proyecto, []member.Integrante, error) {
	p, err := s.proyectos.BuscarPorID(ctx, proyectoID)
	if err != nil {
		return project.Proyecto{}, nil, fmt.Errorf("buscar el proyecto %d: %w", proyectoID, err)
	}
	lista, err := s.repo.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return project.Proyecto{}, nil, fmt.Errorf("listar los integrantes del proyecto %d: %w", proyectoID, err)
	}
	return p, lista, nil
}

// DarDeBaja deja inactivo al integrante (nunca se borra, CA-03.3). Si no existe en el proyecto devuelve
// member.ErrNoEncontrado, y si ya estaba dado de baja, member.ErrYaDadoDeBaja.
func (s *Integrantes) DarDeBaja(ctx context.Context, proyectoID, id int64) (member.Integrante, error) {
	i, err := s.repo.BuscarPorID(ctx, proyectoID, id)
	if err != nil {
		return member.Integrante{}, fmt.Errorf("buscar el integrante %d: %w", id, err)
	}
	baja, err := i.DarDeBaja()
	if err != nil {
		return member.Integrante{}, err
	}
	if err := s.repo.Actualizar(ctx, baja); err != nil {
		return member.Integrante{}, fmt.Errorf("dar de baja al integrante %d: %w", id, err)
	}
	return baja, nil
}
