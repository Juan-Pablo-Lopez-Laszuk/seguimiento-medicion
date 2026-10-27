package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

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

// Votos devuelve los votos de la ronda con su valor, ordenados por integrante (HU-19 CA-19.2). HU-18 solo dice quiénes
// votaron; los valores se muestran recién con la ronda revelada (RevelarPoker).
func (r *Poker) Votos(ctx context.Context, rondaID int64) ([]poker.Voto, error) {
	filas, err := r.pool.Query(ctx,
		`SELECT id, ronda_id, integrante_id, valor FROM votos WHERE ronda_id = $1 ORDER BY integrante_id`, rondaID)
	if err != nil {
		return nil, fmt.Errorf("listar los votos de la ronda %d: %w", rondaID, err)
	}
	votos, err := pgx.CollectRows(filas, func(f pgx.CollectableRow) (poker.Voto, error) {
		var v poker.Voto
		err := f.Scan(&v.ID, &v.RondaID, &v.IntegranteID, &v.Valor)
		return v, err
	})
	if err != nil {
		return nil, fmt.Errorf("leer los votos de la ronda %d: %w", rondaID, err)
	}
	return votos, nil
}
