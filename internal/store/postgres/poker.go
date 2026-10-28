package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
)

// codigoUnico es el código de PostgreSQL de una violación de unicidad (unique_violation).
const codigoUnico = "23505"

// Poker guarda las sesiones de Planning Poker y sus rondas en las tablas sesiones_poker y rondas_poker (épica E4).
// Cada historia de la épica le agrega sus métodos en su propio archivo (spec de HU-17, "Modelo común de Planning
// Poker").
type Poker struct {
	pool *pgxpool.Pool
}

// NuevoPoker crea el repositorio sobre un pool de conexiones ya abierto.
func NuevoPoker(pool *pgxpool.Pool) *Poker {
	return &Poker{pool: pool}
}

// Abrir guarda una sesión Abierta de la historia con su Ronda 1, en una transacción (HU-17, RN3). Si la historia ya
// tiene una sesión abierta, el índice único sesiones_poker_una_abierta_uq la rechaza (también si se abren dos a la
// vez): devuelve poker.ErrSesionYaAbierta y no guarda nada (CA-17.2).
func (r *Poker) Abrir(ctx context.Context, historiaID int64) (poker.Sesion, poker.Ronda, error) {
	s := poker.Sesion{HistoriaID: historiaID, Estado: poker.EstadoAbierta}
	ronda := poker.Ronda{Numero: 1}
	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx,
			`INSERT INTO sesiones_poker (historia_id, estado) VALUES ($1, $2) RETURNING id`,
			historiaID, string(poker.EstadoAbierta),
		).Scan(&s.ID); err != nil {
			return err
		}
		ronda.SesionID = s.ID
		return tx.QueryRow(ctx,
			`INSERT INTO rondas_poker (sesion_id, numero) VALUES ($1, 1) RETURNING id`, s.ID,
		).Scan(&ronda.ID)
	})

	var errBase *pgconn.PgError
	switch {
	case err == nil:
		return s, ronda, nil
	case errors.As(err, &errBase) && errBase.Code == codigoUnico:
		return poker.Sesion{}, poker.Ronda{}, poker.ErrSesionYaAbierta
	default:
		return poker.Sesion{}, poker.Ronda{}, fmt.Errorf("abrir la sesión de la historia %d: %w", historiaID, err)
	}
}

// SesionAbierta devuelve la sesión abierta de la historia y su ronda actual (la de número más alto), o
// poker.ErrSesionNoEncontrada.
func (r *Poker) SesionAbierta(ctx context.Context, historiaID int64) (poker.Sesion, poker.Ronda, error) {
	var (
		s      poker.Sesion
		ronda  poker.Ronda
		estado string
	)
	err := r.pool.QueryRow(ctx,
		`SELECT s.id, s.historia_id, s.estado, s.valor_acordado, r.id, r.sesion_id, r.numero, r.revelada
		 FROM sesiones_poker s JOIN rondas_poker r ON r.sesion_id = s.id
		 WHERE s.historia_id = $1 AND s.estado = $2
		 ORDER BY r.numero DESC LIMIT 1`,
		historiaID, string(poker.EstadoAbierta),
	).Scan(&s.ID, &s.HistoriaID, &estado, &s.ValorAcordado, &ronda.ID, &ronda.SesionID, &ronda.Numero, &ronda.Revelada)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		return poker.Sesion{}, poker.Ronda{}, poker.ErrSesionNoEncontrada
	case err != nil:
		return poker.Sesion{}, poker.Ronda{}, fmt.Errorf("buscar la sesión abierta de la historia %d: %w", historiaID, err)
	}
	s.Estado = poker.EstadoSesion(estado)
	return s, ronda, nil
}
