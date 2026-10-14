package service

import (
	"context"
	"fmt"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
)

// RepositorioEstados es lo que el caso de uso necesita del repositorio de historias: buscar la historia y guardar su
// estado. Lo implementan internal/store/memory (tests) e internal/store/postgres (Supabase).
type RepositorioEstados interface {
	BuscadorHistorias
	// ActualizarEstado guarda el estado de la historia sin tocar los demás datos, o devuelve
	// backlog.ErrHistoriaNoEncontrada si no existe.
	ActualizarEstado(ctx context.Context, historiaID int64, estado backlog.Estado) error
}

// Estados es el caso de uso de cambiar el estado de una historia (HU-08). Lo usan la página de la historia (En sprint
// a En progreso), HU-12 (a En sprint), HU-13 (a Hecho) y HU-14 (de vuelta a Pendiente): todas pasan por la misma regla
// de backlog.Historia.CambiarEstado.
type Estados struct {
	proyectos BuscadorProyectos
	historias RepositorioEstados
	criterios RepositorioCriterios
}

// NuevoEstados arma el caso de uso.
func NuevoEstados(proyectos BuscadorProyectos, historias RepositorioEstados, criterios RepositorioCriterios) *Estados {
	return &Estados{proyectos: proyectos, historias: historias, criterios: criterios}
}

// Cambiar pasa la historia HU-N del proyecto al estado nuevo, si el cambio vale, y lo guarda. Si no vale devuelve el
// error de la regla (backlog.ErrTransicionInvalida, ErrYaEstaEnEseEstado, ErrEstadoDesconocido o
// ErrCriteriosSinCumplir) y no guarda nada; si el proyecto o la historia no existen, project.ErrNoEncontrado o
// backlog.ErrHistoriaNoEncontrada; cualquier otro error viene de los repositorios.
func (s *Estados) Cambiar(ctx context.Context, proyectoID int64, numero int, nuevo backlog.Estado) (backlog.Historia, error) {
	if _, err := s.proyectos.BuscarPorID(ctx, proyectoID); err != nil {
		return backlog.Historia{}, fmt.Errorf("buscar el proyecto %d: %w", proyectoID, err)
	}
	h, err := s.historias.BuscarPorNumero(ctx, proyectoID, numero)
	if err != nil {
		return backlog.Historia{}, fmt.Errorf("buscar la historia HU-%d del proyecto %d: %w", numero, proyectoID, err)
	}
	criterios, err := s.criterios.ListarPorHistoria(ctx, h.ID)
	if err != nil {
		return backlog.Historia{}, fmt.Errorf("listar los criterios de la historia %d: %w", h.ID, err)
	}
	cambiada, err := h.CambiarEstado(nuevo, backlog.ProgresoDe(criterios))
	if err != nil {
		return backlog.Historia{}, err
	}
	if err := s.historias.ActualizarEstado(ctx, h.ID, cambiada.Estado); err != nil {
		return backlog.Historia{}, fmt.Errorf("guardar el estado de la historia %d: %w", h.ID, err)
	}
	return cambiada, nil
}
