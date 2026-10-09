package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
)

// RepositorioProyectos es lo que el caso de uso necesita de la base de datos. Lo implementan
// internal/store/memory (tests) e internal/store/postgres (Supabase).
type RepositorioProyectos interface {
	// ExisteNombre informa si ya hay un proyecto con ese nombre, sin distinguir mayúsculas.
	ExisteNombre(ctx context.Context, nombre string) (bool, error)
	// Guardar guarda el proyecto y lo devuelve con su ID.
	Guardar(ctx context.Context, p project.Proyecto) (project.Proyecto, error)
	// Listar devuelve todos los proyectos en el orden en que se crearon.
	Listar(ctx context.Context) ([]project.Proyecto, error)
}

// Proyectos agrupa los casos de uso de la épica E1.
type Proyectos struct {
	repo  RepositorioProyectos
	ahora func() time.Time
}

// NuevoProyectos arma el caso de uso. ahora devuelve la fecha y hora actual (en los tests, una fija).
func NuevoProyectos(repo RepositorioProyectos, ahora func() time.Time) *Proyectos {
	return &Proyectos{repo: repo, ahora: ahora}
}

// Crear valida los datos (HU-01), controla que el nombre no exista y guarda el proyecto.
// Los datos inválidos se devuelven juntos en domain.ErroresValidacion; cualquier otro error
// viene del repositorio.
func (s *Proyectos) Crear(ctx context.Context, d project.Datos) (project.Proyecto, error) {
	p, err := project.Nuevo(d)
	errs := domain.ErroresValidacion{}
	if err != nil && !errors.As(err, &errs) {
		return project.Proyecto{}, err
	}

	// RN3: solo se busca el nombre si es válido; si no, ya tiene su error.
	if _, nombreInvalido := errs[project.CampoNombre]; !nombreInvalido {
		existe, err := s.repo.ExisteNombre(ctx, project.NormalizarNombre(d.Nombre))
		if err != nil {
			return project.Proyecto{}, fmt.Errorf("verificar el nombre del proyecto: %w", err)
		}
		if existe {
			errs[project.CampoNombre] = project.ErrNombreRepetido
		}
	}
	if len(errs) > 0 {
		return project.Proyecto{}, errs
	}

	p.CreadoEn = s.ahora()
	guardado, err := s.repo.Guardar(ctx, p)
	if err != nil {
		return project.Proyecto{}, fmt.Errorf("guardar el proyecto: %w", err)
	}
	return guardado, nil
}

// Listar devuelve los proyectos en el orden en que se crearon.
func (s *Proyectos) Listar(ctx context.Context) ([]project.Proyecto, error) {
	lista, err := s.repo.Listar(ctx)
	if err != nil {
		return nil, fmt.Errorf("listar los proyectos: %w", err)
	}
	return lista, nil
}
