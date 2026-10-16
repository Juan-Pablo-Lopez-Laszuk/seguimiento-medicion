package service

import (
	"context"
	"fmt"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
)

// RepositorioEdicionHistorias es lo que el caso de uso necesita del repositorio de historias: buscar la historia,
// guardar sus datos corregidos y eliminarla. Lo implementan internal/store/memory (tests) e internal/store/postgres
// (Supabase).
type RepositorioEdicionHistorias interface {
	BuscadorHistorias
	// ActualizarDatos guarda el título, la descripción y la prioridad de la historia sin tocar lo demás, o devuelve
	// backlog.ErrHistoriaNoEncontrada si no existe.
	ActualizarDatos(ctx context.Context, h backlog.Historia) error
	// EnSprintCerrado dice si la historia es de un sprint cerrado: su sprint actual o la foto de cierre lo dicen.
	EnSprintCerrado(ctx context.Context, historiaID int64) (bool, error)
	// Eliminar borra la historia con sus criterios y sus tareas, todo junto, o devuelve
	// backlog.ErrHistoriaNoEncontrada si no existe y backlog.ErrHistoriaConRegistros si tiene otros registros asociados.
	Eliminar(ctx context.Context, historiaID int64) error
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

// VistaEliminacion es lo que muestra la pantalla que pide confirmar la eliminación de una historia: qué se elimina
// con ella y, si no se puede eliminar, por qué.
type VistaEliminacion struct {
	Proyecto  project.Proyecto
	Historia  backlog.Historia
	Criterios int // cuántos criterios de aceptación se eliminan con la historia
	Tareas    int // cuántas tareas se eliminan con la historia
	// Impedimento es el motivo por el que no se puede eliminar (backlog.ErrEliminarHistoriaHecha o
	// backlog.ErrEliminarEnSprintCerrado); nil si se puede.
	Impedimento error
}

// PedirEliminacion arma la pantalla de confirmación de la historia HU-N (CA-06.3): cuenta lo que se eliminaría con ella
// y dice si se puede. No elimina nada. Si el proyecto o la historia no existen, project.ErrNoEncontrado o
// backlog.ErrHistoriaNoEncontrada.
func (s *EdicionHistorias) PedirEliminacion(ctx context.Context, proyectoID int64, numero int) (VistaEliminacion, error) {
	p, h, err := s.Buscar(ctx, proyectoID, numero)
	if err != nil {
		return VistaEliminacion{}, err
	}
	enSprintCerrado, err := s.historias.EnSprintCerrado(ctx, h.ID)
	if err != nil {
		return VistaEliminacion{}, fmt.Errorf("ver si la historia %d es de un sprint cerrado: %w", h.ID, err)
	}
	criterios, err := s.criterios.ListarPorHistoria(ctx, h.ID)
	if err != nil {
		return VistaEliminacion{}, fmt.Errorf("listar los criterios de la historia %d: %w", h.ID, err)
	}
	tareas, err := s.tareas.ListarPorHistoria(ctx, h.ID)
	if err != nil {
		return VistaEliminacion{}, fmt.Errorf("listar las tareas de la historia %d: %w", h.ID, err)
	}
	return VistaEliminacion{
		Proyecto: p, Historia: h, Criterios: len(criterios), Tareas: len(tareas),
		Impedimento: h.PuedeEliminarse(enSprintCerrado),
	}, nil
}

// Eliminar borra la historia HU-N con sus criterios y sus tareas (CA-06.3: la confirmación es el POST que llega acá).
// Una historia Hecha o de un sprint cerrado no se elimina (CA-06.2): devuelve backlog.ErrEliminarHistoriaHecha o
// backlog.ErrEliminarEnSprintCerrado y no borra nada. Si la base la rechaza por otros registros asociados, el error es
// backlog.ErrHistoriaConRegistros. Devuelve la historia eliminada.
func (s *EdicionHistorias) Eliminar(ctx context.Context, proyectoID int64, numero int) (backlog.Historia, error) {
	_, h, err := s.Buscar(ctx, proyectoID, numero)
	if err != nil {
		return backlog.Historia{}, err
	}
	enSprintCerrado, err := s.historias.EnSprintCerrado(ctx, h.ID)
	if err != nil {
		return backlog.Historia{}, fmt.Errorf("ver si la historia %d es de un sprint cerrado: %w", h.ID, err)
	}
	if err := h.PuedeEliminarse(enSprintCerrado); err != nil {
		return backlog.Historia{}, err
	}
	if err := s.historias.Eliminar(ctx, h.ID); err != nil {
		return backlog.Historia{}, fmt.Errorf("eliminar la historia %d: %w", h.ID, err)
	}
	return h, nil
}
