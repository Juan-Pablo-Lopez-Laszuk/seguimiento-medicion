package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
)

// Cierres guarda el cierre de los sprints (HU-14) en las tablas sprints, sprint_historias e historias.
type Cierres struct {
	pool *pgxpool.Pool
}

// NuevoCierres crea el repositorio sobre un pool de conexiones ya abierto.
func NuevoCierres(pool *pgxpool.Pool) *Cierres {
	return &Cierres{pool: pool}
}

// Cerrar guarda todo el cierre en una transacción (HU-14 RN6): el sprint Cerrado con sus SP, la foto en
// sprint_historias y las historias que vuelven al backlog. Devuelve sprint.ErrNoEncontrado si el sprint no existe,
// sprint.ErrYaCerrado si ya está cerrado y backlog.ErrHistoriaNoEncontrada si falta alguna historia; en esos casos no
// guarda nada.
func (r *Cierres) Cerrar(ctx context.Context, c sprint.Cierre) error {
	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		if err := cerrarElSprint(ctx, tx, c.Sprint); err != nil {
			return err
		}
		for _, f := range c.Foto {
			if _, err := tx.Exec(ctx,
				`INSERT INTO sprint_historias (sprint_id, historia_id, story_points, completada) VALUES ($1, $2, $3, $4)`,
				c.Sprint.ID, f.HistoriaID, f.StoryPoints, f.Completada); err != nil {
				return fmt.Errorf("guardar la HU-%d en la foto: %w", f.Numero, err)
			}
		}
		for _, h := range c.Devueltas {
			etiqueta, err := tx.Exec(ctx, `UPDATE historias SET estado = $1, sprint_id = NULL WHERE id = $2`, string(h.Estado), h.ID)
			if err != nil {
				return fmt.Errorf("devolver la HU-%d al backlog: %w", h.Numero, err)
			}
			if etiqueta.RowsAffected() == 0 {
				return backlog.ErrHistoriaNoEncontrada
			}
		}
		return nil
	})
	switch {
	case err == nil:
		return nil
	case errors.Is(err, sprint.ErrNoEncontrado), errors.Is(err, sprint.ErrYaCerrado), errors.Is(err, backlog.ErrHistoriaNoEncontrada):
		return err
	default:
		return fmt.Errorf("cerrar el sprint %d: %w", c.Sprint.ID, err)
	}
}

// cerrarElSprint pone el sprint Cerrado con sus SP. Solo cambia un sprint que no está cerrado, así dos cierres al mismo
// tiempo no pisan la foto: el segundo recibe sprint.ErrYaCerrado.
func cerrarElSprint(ctx context.Context, tx pgx.Tx, s sprint.Sprint) error {
	etiqueta, err := tx.Exec(ctx,
		`UPDATE sprints SET estado = $1, sp_planificados = $2, sp_completados = $3 WHERE id = $4 AND estado <> $1`,
		string(sprint.EstadoCerrado), s.SPPlanificados, s.SPCompletados, s.ID)
	if err != nil {
		return fmt.Errorf("cerrar el sprint: %w", err)
	}
	if etiqueta.RowsAffected() == 1 {
		return nil
	}
	var existe bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM sprints WHERE id = $1)`, s.ID).Scan(&existe); err != nil {
		return fmt.Errorf("buscar el sprint: %w", err)
	}
	if existe {
		return sprint.ErrYaCerrado
	}
	return sprint.ErrNoEncontrado
}

// Foto devuelve la foto del cierre del sprint, ordenada por número de historia; vacía si no se cerró.
func (r *Cierres) Foto(ctx context.Context, sprintID int64) ([]sprint.FotoHistoria, error) {
	filas, err := r.pool.Query(ctx,
		`SELECT sh.historia_id, h.numero, sh.story_points, sh.completada
		 FROM sprint_historias sh JOIN historias h ON h.id = sh.historia_id
		 WHERE sh.sprint_id = $1 ORDER BY h.numero`, sprintID)
	if err != nil {
		return nil, fmt.Errorf("leer la foto del sprint %d: %w", sprintID, err)
	}
	foto, err := pgx.CollectRows(filas, func(fila pgx.CollectableRow) (sprint.FotoHistoria, error) {
		var f sprint.FotoHistoria
		err := fila.Scan(&f.HistoriaID, &f.Numero, &f.StoryPoints, &f.Completada)
		return f, err
	})
	if err != nil {
		return nil, fmt.Errorf("leer la foto del sprint %d: %w", sprintID, err)
	}
	return foto, nil
}
