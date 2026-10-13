package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
)

// RepositorioEdicionProyectos es lo que la modificación necesita del repositorio de proyectos. Lo implementan
// internal/store/memory e internal/store/postgres.
type RepositorioEdicionProyectos interface {
	BuscadorProyectos
	// ExisteOtroConNombre informa si otro proyecto (no el de exceptoID) ya tiene ese nombre, sin distinguir mayúsculas.
	ExisteOtroConNombre(ctx context.Context, nombre string, exceptoID int64) (bool, error)
	// Actualizar guarda los datos del proyecto, o devuelve project.ErrNoEncontrado si no existe.
	Actualizar(ctx context.Context, p project.Proyecto) error
}

// EdicionProyectos agrupa los casos de uso de modificar un proyecto (HU-02).
type EdicionProyectos struct {
	repo    RepositorioEdicionProyectos
	sprints ListadorSprints
	ahora   func() time.Time
}

// NuevoEdicionProyectos arma el caso de uso. ahora devuelve la fecha y hora actual (en los tests, una fija).
func NuevoEdicionProyectos(repo RepositorioEdicionProyectos, sprints ListadorSprints, ahora func() time.Time) *EdicionProyectos {
	return &EdicionProyectos{repo: repo, sprints: sprints, ahora: ahora}
}

// Preparar devuelve el proyecto para cargar el formulario con sus datos actuales. Si está Finalizado devuelve
// project.ErrProyectoFinalizado y el formulario no se muestra (RN5); si no existe, project.ErrNoEncontrado.
func (s *EdicionProyectos) Preparar(ctx context.Context, id int64) (project.Proyecto, error) {
	p, limites, err := s.cargar(ctx, id)
	if err != nil {
		return project.Proyecto{}, err
	}
	if !project.PuedeModificarse(p, limites, domain.HoyEnArgentina(s.ahora())) {
		return project.Proyecto{}, project.ErrProyectoFinalizado
	}
	return p, nil
}

// Modificar valida los datos nuevos (spec: specs/HU-02-modificar-proyecto.md) y guarda el proyecto. Los datos
// inválidos vuelven juntos en domain.ErroresValidacion, incluido el nombre repetido (RN2); un proyecto Finalizado
// devuelve project.ErrProyectoFinalizado; cualquier otro error viene de los repositorios.
func (s *EdicionProyectos) Modificar(ctx context.Context, id int64, d project.Datos) (project.Proyecto, error) {
	actual, limites, err := s.cargar(ctx, id)
	if err != nil {
		return project.Proyecto{}, err
	}
	p, err := project.Modificar(actual, d, limites, domain.HoyEnArgentina(s.ahora()))
	errs := domain.ErroresValidacion{}
	if err != nil && !errors.As(err, &errs) {
		return project.Proyecto{}, err // Finalizado
	}

	// Igual que al crear: solo se busca el nombre si es válido; si no, ya tiene su error.
	if _, nombreInvalido := errs[project.CampoNombre]; !nombreInvalido {
		existe, err := s.repo.ExisteOtroConNombre(ctx, project.NormalizarNombre(d.Nombre), id)
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

	if err := s.repo.Actualizar(ctx, p); err != nil {
		return project.Proyecto{}, fmt.Errorf("guardar el proyecto %d: %w", id, err)
	}
	return p, nil
}

// cargar busca el proyecto y calcula los límites de sus sprints.
func (s *EdicionProyectos) cargar(ctx context.Context, id int64) (project.Proyecto, project.LimitesSprints, error) {
	p, err := s.repo.BuscarPorID(ctx, id)
	if err != nil {
		return project.Proyecto{}, project.LimitesSprints{}, fmt.Errorf("buscar el proyecto %d: %w", id, err)
	}
	sprints, err := s.sprints.ListarPorProyecto(ctx, id)
	if err != nil {
		return project.Proyecto{}, project.LimitesSprints{}, fmt.Errorf("listar los sprints del proyecto %d: %w", id, err)
	}
	return p, sprint.Limites(sprints), nil
}
