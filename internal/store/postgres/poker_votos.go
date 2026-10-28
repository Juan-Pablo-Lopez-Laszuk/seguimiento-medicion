package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
)

// Votar guarda el voto del integrante en la ronda; si ya había votado en esa ronda, lo reemplaza con un UPDATE sobre
// votos_uno_por_ronda_uq (HU-18, CA-18.1). Guardar y controlar que la ronda no esté revelada ni su sesión cerrada va
// en una sola sentencia, que bloquea la ronda mientras tanto: si se revela a la vez, el voto no se guarda (RN3). Si no
// se guardó, devuelve poker.ErrSesionNoEncontrada (la ronda no existe), poker.ErrSesionCerrada o
// poker.ErrRondaRevelada.
func (r *Poker) Votar(ctx context.Context, rondaID, integranteID int64, valor int) error {
	resultado, err := r.pool.Exec(ctx,
		`INSERT INTO votos (ronda_id, integrante_id, valor)
		 SELECT r.id, $2, $3
		 FROM rondas_poker r JOIN sesiones_poker s ON s.id = r.sesion_id
		 WHERE r.id = $1 AND NOT r.revelada AND s.estado = $4
		 FOR SHARE
		 ON CONFLICT (ronda_id, integrante_id) DO UPDATE SET valor = EXCLUDED.valor`,
		rondaID, integranteID, valor, string(poker.EstadoAbierta),
	)
	if err != nil {
		return fmt.Errorf("guardar el voto en la ronda %d: %w", rondaID, err)
	}
	if resultado.RowsAffected() == 1 {
		return nil
	}
	return r.porQueNoSeVota(ctx, rondaID)
}

// QuienesVotaron devuelve los IDs de los integrantes que votaron en la ronda, de menor a mayor. Nunca los valores
// (CA-18.3, RN5): leerlos es de HU-19, con la ronda revelada.
func (r *Poker) QuienesVotaron(ctx context.Context, rondaID int64) ([]int64, error) {
	filas, err := r.pool.Query(ctx,
		`SELECT integrante_id FROM votos WHERE ronda_id = $1 ORDER BY integrante_id`, rondaID)
	if err != nil {
		return nil, fmt.Errorf("listar quiénes votaron en la ronda %d: %w", rondaID, err)
	}
	ids, err := pgx.CollectRows(filas, pgx.RowTo[int64])
	if err != nil {
		return nil, fmt.Errorf("listar quiénes votaron en la ronda %d: %w", rondaID, err)
	}
	return ids, nil
}

// porQueNoSeVota devuelve el error de un voto que no se guardó: la ronda no existe, su sesión está cerrada o ya se
// reveló.
func (r *Poker) porQueNoSeVota(ctx context.Context, rondaID int64) error {
	var (
		revelada bool
		estado   string
	)
	err := r.pool.QueryRow(ctx,
		`SELECT r.revelada, s.estado FROM rondas_poker r JOIN sesiones_poker s ON s.id = r.sesion_id WHERE r.id = $1`,
		rondaID,
	).Scan(&revelada, &estado)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return poker.ErrSesionNoEncontrada
	case err != nil:
		return fmt.Errorf("revisar la ronda %d: %w", rondaID, err)
	case poker.EstadoSesion(estado) == poker.EstadoCerrada:
		return poker.ErrSesionCerrada
	case revelada:
		return poker.ErrRondaRevelada
	}
	return fmt.Errorf("el voto en la ronda %d no se guardó", rondaID)
}
