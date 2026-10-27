package memory

import (
	"context"
	"errors"
	"slices"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
)

// EstimadorDeHistorias guarda los story points de una historia (HU-16, con su historial de HU-16b). Lo implementa
// Historias; el cierre de una sesión de Planning Poker lo usa (HU-22).
type EstimadorDeHistorias interface {
	ActualizarStoryPoints(ctx context.Context, historiaID int64, puntos int) error
}

// ConHistorias conecta el repositorio de historias, que guarda los story points al cerrar una sesión (HU-22). Devuelve
// el mismo repositorio para armarlo en una línea.
func (r *Poker) ConHistorias(historias EstimadorDeHistorias) *Poker {
	r.historias = historias
	return r
}

// Cerrar deja la sesión Cerrada con el valor acordado y guarda los story points de la historia (HU-22). Si la sesión ya
// no está abierta devuelve poker.ErrSesionCerrada y no guarda nada (RN5). La sesión se cierra primero, con el candado,
// así de dos pedidos a la vez se registra uno solo; si después no se pueden guardar los story points, se vuelve a
// abrir (RN6, todo o nada). Los story points se guardan sin el candado de Poker: Historias, al eliminar, pregunta a
// Poker con su propio candado tomado.
func (r *Poker) Cerrar(ctx context.Context, sesionID, historiaID int64, valor int) error {
	if r.historias == nil {
		return errors.New("el repositorio de Planning Poker no tiene conectado el de historias (ConHistorias)")
	}
	r.mu.Lock()
	i := slices.IndexFunc(r.sesiones, func(s poker.Sesion) bool { return s.ID == sesionID && s.Estado == poker.EstadoAbierta })
	if i == -1 {
		r.mu.Unlock()
		return poker.ErrSesionCerrada
	}
	r.sesiones[i].Estado, r.sesiones[i].ValorAcordado = poker.EstadoCerrada, &valor
	r.mu.Unlock()

	if err := r.historias.ActualizarStoryPoints(ctx, historiaID, valor); err != nil {
		r.mu.Lock()
		r.sesiones[i].Estado, r.sesiones[i].ValorAcordado = poker.EstadoAbierta, nil
		r.mu.Unlock()
		return err
	}
	return nil
}

// SesionesCerradas devuelve las sesiones cerradas de la historia, la más reciente primero (CA-22.3), con su número
// entre todas las sesiones de la historia y su cantidad de rondas.
func (r *Poker) SesionesCerradas(_ context.Context, historiaID int64) ([]poker.SesionCerrada, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cerradas := []poker.SesionCerrada{}
	numero := 0
	for _, s := range r.sesiones {
		if s.HistoriaID != historiaID {
			continue
		}
		numero++
		if s.Estado != poker.EstadoCerrada {
			continue
		}
		rondas := 0
		for _, ronda := range r.rondas {
			if ronda.SesionID == s.ID {
				rondas++
			}
		}
		cerradas = append(cerradas, poker.SesionCerrada{Numero: numero, Rondas: rondas, ValorAcordado: *s.ValorAcordado})
	}
	slices.Reverse(cerradas)
	return cerradas, nil
}
