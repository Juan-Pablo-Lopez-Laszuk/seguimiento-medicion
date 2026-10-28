package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
)

// Cerrar deja la sesión Cerrada con el valor acordado y guarda los story points de la historia, con su entrada del
// historial de HU-16b si cambian, todo en una transacción (HU-22, RN6). La sesión solo se actualiza si sigue Abierta:
// si ya estaba cerrada (también si dos pedidos la cierran a la vez, porque el segundo espera al primero y después no
// encuentra la fila) devuelve poker.ErrSesionCerrada y no guarda nada (RN5). Si la historia no existe devuelve
// backlog.ErrHistoriaNoEncontrada y la sesión sigue abierta.
func (r *Poker) Cerrar(ctx context.Context, sesionID, historiaID int64, valor int) error {
	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		etiqueta, err := tx.Exec(ctx,
			`UPDATE sesiones_poker SET estado = $1, valor_acordado = $2 WHERE id = $3 AND estado = $4`,
			string(poker.EstadoCerrada), valor, sesionID, string(poker.EstadoAbierta))
		if err != nil {
			return err
		}
		if etiqueta.RowsAffected() == 0 {
			return poker.ErrSesionCerrada
		}
		return guardarStoryPoints(ctx, tx, historiaID, valor, time.Now())
	})
	switch {
	case err == nil:
		return nil
	case errors.Is(err, poker.ErrSesionCerrada), errors.Is(err, backlog.ErrHistoriaNoEncontrada):
		return err
	default:
		return fmt.Errorf("cerrar la sesión %d de la historia %d: %w", sesionID, historiaID, err)
	}
}

// SesionesCerradas devuelve las sesiones cerradas de la historia, la más reciente primero (CA-22.3). Cada una lleva su
// número entre todas las sesiones de la historia, en el orden en que se abrieron, y su cantidad de rondas.
func (r *Poker) SesionesCerradas(ctx context.Context, historiaID int64) ([]poker.SesionCerrada, error) {
	filas, err := r.pool.Query(ctx,
		`SELECT numero, rondas, valor_acordado FROM (
		     SELECT s.estado, s.valor_acordado,
		            row_number() OVER (ORDER BY s.id) AS numero,
		            (SELECT count(*) FROM rondas_poker r WHERE r.sesion_id = s.id) AS rondas
		     FROM sesiones_poker s WHERE s.historia_id = $1
		 ) t
		 WHERE estado = $2 ORDER BY numero DESC`,
		historiaID, string(poker.EstadoCerrada))
	if err != nil {
		return nil, fmt.Errorf("listar las sesiones cerradas de la historia %d: %w", historiaID, err)
	}
	defer filas.Close()
	cerradas := []poker.SesionCerrada{}
	for filas.Next() {
		var c poker.SesionCerrada
		if err := filas.Scan(&c.Numero, &c.Rondas, &c.ValorAcordado); err != nil {
			return nil, fmt.Errorf("leer una sesión cerrada: %w", err)
		}
		cerradas = append(cerradas, c)
	}
	if err := filas.Err(); err != nil {
		return nil, fmt.Errorf("listar las sesiones cerradas de la historia %d: %w", historiaID, err)
	}
	return cerradas, nil
}
