package service

import (
	"context"
	"fmt"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
)

// RepositorioEstimaciones es lo que el caso de uso necesita del repositorio de historias: buscar la historia y guardar
// sus story points. Lo implementan internal/store/memory (tests) e internal/store/postgres (Supabase).
type RepositorioEstimaciones interface {
	BuscadorHistorias
	// ActualizarStoryPoints guarda los story points de la historia sin tocar los demás datos, o devuelve
	// backlog.ErrHistoriaNoEncontrada si no existe. Es el único lugar donde se escribe la estimación.
	ActualizarStoryPoints(ctx context.Context, historiaID int64, puntos int) error
}

// Estimaciones es el caso de uso de estimar una historia con story points (HU-16). Las pantallas de Planning Poker
// pueden usarlo para guardar el valor acordado: pasa por la misma regla, backlog.Historia.Estimar.
type Estimaciones struct {
	proyectos BuscadorProyectos
	historias RepositorioEstimaciones
}

// NuevoEstimaciones arma el caso de uso.
func NuevoEstimaciones(proyectos BuscadorProyectos, historias RepositorioEstimaciones) *Estimaciones {
	return &Estimaciones{proyectos: proyectos, historias: historias}
}

// Estimar deja la historia HU-N del proyecto con los story points del texto, si valen, y los guarda; si ya tenía una
// estimación, la reemplaza. Un valor inválido se devuelve en domain.ErroresValidacion; en una historia Hecha,
// backlog.ErrEstimacionDeHistoriaHecha; si el proyecto o la historia no existen, project.ErrNoEncontrado o
// backlog.ErrHistoriaNoEncontrada; cualquier otro error viene de los repositorios. Si no vale, no guarda nada.
func (s *Estimaciones) Estimar(ctx context.Context, proyectoID int64, numero int, texto string) (backlog.Historia, error) {
	if _, err := s.proyectos.BuscarPorID(ctx, proyectoID); err != nil {
		return backlog.Historia{}, fmt.Errorf("buscar el proyecto %d: %w", proyectoID, err)
	}
	h, err := s.historias.BuscarPorNumero(ctx, proyectoID, numero)
	if err != nil {
		return backlog.Historia{}, fmt.Errorf("buscar la historia HU-%d del proyecto %d: %w", numero, proyectoID, err)
	}
	estimada, err := h.Estimar(texto)
	if err != nil {
		return backlog.Historia{}, err
	}
	if err := s.historias.ActualizarStoryPoints(ctx, h.ID, *estimada.StoryPoints); err != nil {
		return backlog.Historia{}, fmt.Errorf("guardar los story points de la historia %d: %w", h.ID, err)
	}
	return estimada, nil
}
