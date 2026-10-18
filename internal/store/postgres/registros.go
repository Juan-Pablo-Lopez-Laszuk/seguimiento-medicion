package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
)

// Registros guarda los registros de esfuerzo en la tabla registros_esfuerzo (HU-24).
type Registros struct {
	pool *pgxpool.Pool
}

// NuevoRegistros crea el repositorio sobre un pool de conexiones ya abierto.
func NuevoRegistros(pool *pgxpool.Pool) *Registros {
	return &Registros{pool: pool}
}

// Guardar inserta el registro y lo devuelve con su ID. La tarea va en NULL si el registro es de toda la historia; la
// base rechaza horas fuera de (0, 24] y un integrante, historia o tarea que no existan.
func (r *Registros) Guardar(ctx context.Context, reg effort.Registro) (effort.Registro, error) {
	var tarea *int64
	if reg.TareaID != 0 {
		tarea = &reg.TareaID
	}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO registros_esfuerzo (integrante_id, historia_id, tarea_id, fecha, actividad, horas)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		reg.IntegranteID, reg.HistoriaID, tarea, reg.Fecha, reg.Actividad, reg.Horas,
	).Scan(&reg.ID)
	if err != nil {
		return effort.Registro{}, fmt.Errorf("guardar el registro de esfuerzo: %w", err)
	}
	return reg, nil
}

// ListarPorHistoria devuelve los registros de la historia, la fecha más reciente primero.
func (r *Registros) ListarPorHistoria(ctx context.Context, historiaID int64) ([]effort.Registro, error) {
	return r.listar(ctx, "historia_id", historiaID)
}

// ListarPorIntegrante devuelve los registros del integrante, la fecha más reciente primero.
func (r *Registros) ListarPorIntegrante(ctx context.Context, integranteID int64) ([]effort.Registro, error) {
	return r.listar(ctx, "integrante_id", integranteID)
}

// listar filtra por columna, que siempre es una de las dos constantes de arriba (nunca un dato del usuario).
func (r *Registros) listar(ctx context.Context, columna string, id int64) ([]effort.Registro, error) {
	filas, err := r.pool.Query(ctx,
		`SELECT id, integrante_id, COALESCE(historia_id, 0), COALESCE(tarea_id, 0), fecha, actividad, horas::float8
		 FROM registros_esfuerzo WHERE `+columna+` = $1 ORDER BY fecha DESC, id DESC`, id)
	if err != nil {
		return nil, fmt.Errorf("listar los registros de esfuerzo: %w", err)
	}
	defer filas.Close()
	var lista []effort.Registro
	for filas.Next() {
		var reg effort.Registro
		if err := filas.Scan(&reg.ID, &reg.IntegranteID, &reg.HistoriaID, &reg.TareaID, &reg.Fecha, &reg.Actividad, &reg.Horas); err != nil {
			return nil, fmt.Errorf("leer un registro de esfuerzo: %w", err)
		}
		lista = append(lista, reg)
	}
	if err := filas.Err(); err != nil {
		return nil, fmt.Errorf("listar los registros de esfuerzo: %w", err)
	}
	return lista, nil
}
