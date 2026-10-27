package memory

import (
	"context"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
)

// Revelar marca la ronda como revelada (HU-19 CA-19.1). Si la ronda no existe devuelve poker.ErrSesionNoEncontrada.
func (r *Poker) Revelar(_ context.Context, rondaID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.rondas {
		if r.rondas[i].ID == rondaID {
			r.rondas[i].Revelada = true
			return nil
		}
	}
	return poker.ErrSesionNoEncontrada
}
