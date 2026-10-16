package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
)

// Historias guarda las historias de todos los proyectos en la tabla historias.
type Historias struct {
	pool *pgxpool.Pool
}

// NuevoHistorias crea el repositorio sobre un pool de conexiones ya abierto.
func NuevoHistorias(pool *pgxpool.Pool) *Historias {
	return &Historias{pool: pool}
}

// Guardar inserta la historia y la devuelve con el ID que le asignó la base. Los story points sin estimar se
// guardan como NULL. La base rechaza un número repetido dentro del proyecto (restricción única), un proyecto que
// no existe (clave foránea) y una prioridad o unos story points fuera de lo permitido (CHECK).
func (r *Historias) Guardar(ctx context.Context, h backlog.Historia) (backlog.Historia, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO historias (proyecto_id, numero, titulo, descripcion, prioridad, estado, story_points)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		h.ProyectoID, h.Numero, h.Titulo, h.Descripcion, string(h.Prioridad), string(h.Estado), h.StoryPoints,
	).Scan(&h.ID)
	if err != nil {
		return backlog.Historia{}, fmt.Errorf("guardar la historia: %w", err)
	}
	return h, nil
}

// ListarPorProyecto devuelve las historias del proyecto en el orden en que se crearon.
func (r *Historias) ListarPorProyecto(ctx context.Context, proyectoID int64) ([]backlog.Historia, error) {
	filas, err := r.pool.Query(ctx,
		`SELECT id, proyecto_id, numero, titulo, descripcion, prioridad, estado, story_points
		 FROM historias WHERE proyecto_id = $1 ORDER BY numero`, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("listar las historias del proyecto: %w", err)
	}
	defer filas.Close()

	var lista []backlog.Historia
	for filas.Next() {
		var h backlog.Historia
		var prioridad, estado string
		if err := filas.Scan(&h.ID, &h.ProyectoID, &h.Numero, &h.Titulo, &h.Descripcion, &prioridad, &estado, &h.StoryPoints); err != nil {
			return nil, fmt.Errorf("leer una historia: %w", err)
		}
		h.Prioridad, h.Estado = backlog.Prioridad(prioridad), backlog.Estado(estado)
		lista = append(lista, h)
	}
	if err := filas.Err(); err != nil {
		return nil, fmt.Errorf("listar las historias del proyecto: %w", err)
	}
	return lista, nil
}

// BuscarPorNumero devuelve la historia con ese número (HU-N) dentro del proyecto, o backlog.ErrHistoriaNoEncontrada.
func (r *Historias) BuscarPorNumero(ctx context.Context, proyectoID int64, numero int) (backlog.Historia, error) {
	var h backlog.Historia
	var prioridad, estado string
	err := r.pool.QueryRow(ctx,
		`SELECT id, proyecto_id, numero, titulo, descripcion, prioridad, estado, story_points
		 FROM historias WHERE proyecto_id = $1 AND numero = $2`, proyectoID, numero,
	).Scan(&h.ID, &h.ProyectoID, &h.Numero, &h.Titulo, &h.Descripcion, &prioridad, &estado, &h.StoryPoints)
	if errors.Is(err, pgx.ErrNoRows) {
		return backlog.Historia{}, backlog.ErrHistoriaNoEncontrada
	}
	if err != nil {
		return backlog.Historia{}, fmt.Errorf("buscar la historia HU-%d: %w", numero, err)
	}
	h.Prioridad, h.Estado = backlog.Prioridad(prioridad), backlog.Estado(estado)
	return h, nil
}

// ActualizarEstado guarda el estado de la historia sin tocar los demás datos, o devuelve
// backlog.ErrHistoriaNoEncontrada si no existe. La base rechaza un estado que no sea uno de los cuatro (CHECK).
func (r *Historias) ActualizarEstado(ctx context.Context, historiaID int64, estado backlog.Estado) error {
	etiqueta, err := r.pool.Exec(ctx, `UPDATE historias SET estado = $1 WHERE id = $2`, string(estado), historiaID)
	if err != nil {
		return fmt.Errorf("actualizar el estado de la historia %d: %w", historiaID, err)
	}
	if etiqueta.RowsAffected() == 0 {
		return backlog.ErrHistoriaNoEncontrada
	}
	return nil
}

// ActualizarDatos guarda el título, la descripción y la prioridad de la historia (HU-06) sin tocar el número, el
// estado ni los story points, o devuelve backlog.ErrHistoriaNoEncontrada si no existe. La base rechaza una prioridad
// fuera de las tres (CHECK).
func (r *Historias) ActualizarDatos(ctx context.Context, h backlog.Historia) error {
	etiqueta, err := r.pool.Exec(ctx,
		`UPDATE historias SET titulo = $1, descripcion = $2, prioridad = $3 WHERE id = $4`,
		h.Titulo, h.Descripcion, string(h.Prioridad), h.ID)
	if err != nil {
		return fmt.Errorf("actualizar los datos de la historia %d: %w", h.ID, err)
	}
	if etiqueta.RowsAffected() == 0 {
		return backlog.ErrHistoriaNoEncontrada
	}
	return nil
}

// codigoClaveForanea es el código de PostgreSQL de una violación de clave foránea (foreign_key_violation).
const codigoClaveForanea = "23503"

// Eliminar borra la historia junto con sus criterios y sus tareas (HU-06, RN9), todo en una transacción: o se borra
// todo o no se borra nada. Devuelve backlog.ErrHistoriaNoEncontrada si no existe, y backlog.ErrHistoriaConRegistros si
// la base la rechaza porque tiene otros registros asociados (sesiones de Planning Poker, esfuerzo, defectos o la foto
// de un sprint cerrado), que no se borran (RN10).
func (r *Historias) Eliminar(ctx context.Context, historiaID int64) error {
	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `DELETE FROM criterios_aceptacion WHERE historia_id = $1`, historiaID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `DELETE FROM tareas WHERE historia_id = $1`, historiaID); err != nil {
			return err
		}
		etiqueta, err := tx.Exec(ctx, `DELETE FROM historias WHERE id = $1`, historiaID)
		if err != nil {
			return err
		}
		if etiqueta.RowsAffected() == 0 {
			return backlog.ErrHistoriaNoEncontrada
		}
		return nil
	})

	var errBase *pgconn.PgError
	switch {
	case err == nil:
		return nil
	case errors.Is(err, backlog.ErrHistoriaNoEncontrada):
		return backlog.ErrHistoriaNoEncontrada
	case errors.As(err, &errBase) && errBase.Code == codigoClaveForanea:
		return backlog.ErrHistoriaConRegistros
	default:
		return fmt.Errorf("eliminar la historia %d: %w", historiaID, err)
	}
}
