package service

import (
	"context"
	"fmt"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
)

// RepositorioEdicionHistorias es lo que el caso de uso necesita del repositorio de historias: buscar la historia y
// guardar sus datos corregidos. Lo implementan internal/store/memory (tests) e internal/store/postgres (Supabase).
type RepositorioEdicionHistorias interface {
	BuscadorHistorias
	// ActualizarDatos guarda el título, la descripción y la prioridad de la historia sin tocar lo demás, o devuelve
	// backlog.ErrHistoriaNoEncontrada si no existe.
	ActualizarDatos(ctx context.Context, h backlog.Historia) error
}

// EdicionHistorias agrupa los casos de uso de editar y eliminar una historia del Product Backlog (HU-06).
type EdicionHistorias struct {
	proyectos BuscadorProyectos
	historias RepositorioEdicionHistorias
	criterios RepositorioCriterios
	tareas    RepositorioTareas
}

// NuevoEdicionHistorias arma el caso de uso.
func NuevoEdicionHistorias(proyectos BuscadorProyectos, historias RepositorioEdicionHistorias, criterios RepositorioCriterios, tareas RepositorioTareas) *EdicionHistorias {
	return &EdicionHistorias{proyectos: proyectos, historias: historias, criterios: criterios, tareas: tareas}
}

// Buscar devuelve el proyecto y su historia HU-N. Si el proyecto no existe, el error es project.ErrNoEncontrado; si la
// historia no existe en el proyecto, backlog.ErrHistoriaNoEncontrada.
func (s *EdicionHistorias) Buscar(ctx context.Context, proyectoID int64, numero int) (project.Proyecto, backlog.Historia, error) {
	p, err := s.proyectos.BuscarPorID(ctx, proyectoID)
	if err != nil {
		return project.Proyecto{}, backlog.Historia{}, fmt.Errorf("buscar el proyecto %d: %w", proyectoID, err)
	}
	h, err := s.historias.BuscarPorNumero(ctx, proyectoID, numero)
	if err != nil {
		return project.Proyecto{}, backlog.Historia{}, fmt.Errorf("buscar la historia HU-%d del proyecto %d: %w", numero, proyectoID, err)
	}
	return p, h, nil
}

// Editar corrige el título, la descripción y la prioridad de la historia HU-N (HU-06) y la guarda. Con datos inválidos
// devuelve domain.ErroresValidacion y no guarda nada; si el proyecto o la historia no existen, project.ErrNoEncontrado
// o backlog.ErrHistoriaNoEncontrada.
func (s *EdicionHistorias) Editar(ctx context.Context, proyectoID int64, numero int, c backlog.Cambios) (backlog.Historia, error) {
	_, h, err := s.Buscar(ctx, proyectoID, numero)
	if err != nil {
		return backlog.Historia{}, err
	}
	editada, err := h.Editar(c)
	if err != nil {
		return backlog.Historia{}, err
	}
	if err := s.historias.ActualizarDatos(ctx, editada); err != nil {
		return backlog.Historia{}, fmt.Errorf("guardar la historia %d: %w", h.ID, err)
	}
	return editada, nil
}
