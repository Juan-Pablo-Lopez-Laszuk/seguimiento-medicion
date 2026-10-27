package memory

import (
	"context"
	"slices"

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

// Votos devuelve los votos de la ronda con su valor, ordenados por integrante (HU-19 CA-19.2). HU-18 solo dice quiénes
// votaron; los valores se muestran recién con la ronda revelada (RevelarPoker).
func (r *Poker) Votos(_ context.Context, rondaID int64) ([]poker.Voto, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var votos []poker.Voto
	for _, v := range r.votos {
		if v.RondaID == rondaID {
			votos = append(votos, v)
		}
	}
	slices.SortFunc(votos, func(a, b poker.Voto) int { return int(a.IntegranteID - b.IntegranteID) })
	return votos, nil
}
