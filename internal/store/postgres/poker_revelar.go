package postgres

import (
	"context"
	"fmt"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
)

// Revelar marca la ronda como revelada (HU-19 CA-19.1). Si la ronda no existe devuelve poker.ErrSesionNoEncontrada.
func (r *Poker) Revelar(ctx context.Context, rondaID int64) error {
	etiqueta, err := r.pool.Exec(ctx, `UPDATE rondas_poker SET revelada = true WHERE id = $1`, rondaID)
	if err != nil {
		return fmt.Errorf("revelar la ronda %d: %w", rondaID, err)
	}
	if etiqueta.RowsAffected() == 0 {
		return poker.ErrSesionNoEncontrada
	}
	return nil
}
