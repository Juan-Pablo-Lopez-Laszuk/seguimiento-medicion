package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
)

// Criterios guarda los criterios de aceptación de todas las historias en la tabla criterios_aceptacion.
type Criterios struct {
	pool *pgxpool.Pool
}

// NuevoCriterios crea el repositorio sobre un pool de conexiones ya abierto.
func NuevoCriterios(pool *pgxpool.Pool) *Criterios {
	return &Criterios{pool: pool}
}

// Guardar inserta el criterio y lo devuelve con el ID que le asignó la base. La base rechaza un criterio de una
// historia que no existe (clave foránea).
func (r *Criterios) Guardar(ctx context.Context, c backlog.Criterio) (backlog.Criterio, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO criterios_aceptacion (historia_id, descripcion, cumplido)
		 VALUES ($1, $2, $3) RETURNING id`,
		c.HistoriaID, c.Descripcion, c.Cumplido,
	).Scan(&c.ID)
	if err != nil {
		return backlog.Criterio{}, fmt.Errorf("guardar el criterio: %w", err)
	}
	return c, nil
}

// ListarPorHistoria devuelve los criterios de la historia en el orden en que se cargaron.
func (r *Criterios) ListarPorHistoria(ctx context.Context, historiaID int64) ([]backlog.Criterio, error) {
	return r.ListarPorHistorias(ctx, []int64{historiaID})
}

// ListarPorHistorias devuelve los criterios de todas esas historias juntos, en el orden en que se cargaron, con una
// sola consulta.
func (r *Criterios) ListarPorHistorias(ctx context.Context, historiaIDs []int64) ([]backlog.Criterio, error) {
	if len(historiaIDs) == 0 {
		return nil, nil
	}
	filas, err := r.pool.Query(ctx,
		`SELECT id, historia_id, descripcion, cumplido
		 FROM criterios_aceptacion WHERE historia_id = ANY($1) ORDER BY id`, historiaIDs)
	if err != nil {
		return nil, fmt.Errorf("listar los criterios: %w", err)
	}
	defer filas.Close()

	var lista []backlog.Criterio
	for filas.Next() {
		var c backlog.Criterio
		if err := filas.Scan(&c.ID, &c.HistoriaID, &c.Descripcion, &c.Cumplido); err != nil {
			return nil, fmt.Errorf("leer un criterio: %w", err)
		}
		lista = append(lista, c)
	}
	if err := filas.Err(); err != nil {
		return nil, fmt.Errorf("listar los criterios: %w", err)
	}
	return lista, nil
}

// BuscarPorID devuelve el criterio dentro de su historia, o backlog.ErrCriterioNoEncontrado. Con el id de un
// criterio de otra historia también es "no encontrado" (RN8).
func (r *Criterios) BuscarPorID(ctx context.Context, historiaID, id int64) (backlog.Criterio, error) {
	var c backlog.Criterio
	err := r.pool.QueryRow(ctx,
		`SELECT id, historia_id, descripcion, cumplido
		 FROM criterios_aceptacion WHERE id = $1 AND historia_id = $2`, id, historiaID,
	).Scan(&c.ID, &c.HistoriaID, &c.Descripcion, &c.Cumplido)
	if errors.Is(err, pgx.ErrNoRows) {
		return backlog.Criterio{}, backlog.ErrCriterioNoEncontrado
	}
	if err != nil {
		return backlog.Criterio{}, fmt.Errorf("buscar el criterio %d: %w", id, err)
	}
	return c, nil
}

// Actualizar guarda si el criterio está cumplido, o devuelve backlog.ErrCriterioNoEncontrado si no existe en esa
// historia. La descripción no se modifica: editar criterios no es parte de HU-07.
func (r *Criterios) Actualizar(ctx context.Context, c backlog.Criterio) error {
	etiqueta, err := r.pool.Exec(ctx,
		`UPDATE criterios_aceptacion SET cumplido = $1 WHERE id = $2 AND historia_id = $3`,
		c.Cumplido, c.ID, c.HistoriaID)
	if err != nil {
		return fmt.Errorf("actualizar el criterio %d: %w", c.ID, err)
	}
	if etiqueta.RowsAffected() == 0 {
		return backlog.ErrCriterioNoEncontrado
	}
	return nil
}
