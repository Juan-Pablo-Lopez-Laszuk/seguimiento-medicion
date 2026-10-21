package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
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
	filas, err := r.pool.Query(ctx, `SELECT `+columnasRegistro+` FROM registros_esfuerzo WHERE `+columna+` = $1 ORDER BY fecha DESC, id DESC`, id)
	if err != nil {
		return nil, fmt.Errorf("listar los registros de esfuerzo: %w", err)
	}
	return leerRegistros(filas)
}

// columnasRegistro son las columnas de un registro, en el orden en que las lee escanearRegistro.
const columnasRegistro = `id, integrante_id, COALESCE(historia_id, 0), COALESCE(tarea_id, 0), fecha, actividad, horas::float8`

// ListarFiltrados devuelve los registros de esas historias que cumplen el filtro, la fecha más reciente primero
// (HU-26). Cada valor del filtro va como parámetro; un 0 o una fecha cero no filtran.
func (r *Registros) ListarFiltrados(ctx context.Context, historiaIDs []int64, f effort.Filtro) ([]effort.Registro, error) {
	filas, err := r.pool.Query(ctx,
		`SELECT `+columnasRegistro+` FROM registros_esfuerzo
		 WHERE historia_id = ANY($1)
		   AND ($2 = 0 OR integrante_id = $2)
		   AND ($3 = 0 OR historia_id = $3)
		   AND ($4::date IS NULL OR fecha >= $4)
		   AND ($5::date IS NULL OR fecha <= $5)
		 ORDER BY fecha DESC, id DESC`,
		historiaIDs, f.IntegranteID, f.HistoriaID, fechaONula(f.Desde), fechaONula(f.Hasta))
	if err != nil {
		return nil, fmt.Errorf("listar los registros de esfuerzo filtrados: %w", err)
	}
	return leerRegistros(filas)
}

// BuscarPorID devuelve el registro con ese ID, o effort.ErrRegistroNoEncontrado (HU-26).
func (r *Registros) BuscarPorID(ctx context.Context, id int64) (effort.Registro, error) {
	var reg effort.Registro
	err := escanearRegistro(r.pool.QueryRow(ctx, `SELECT `+columnasRegistro+` FROM registros_esfuerzo WHERE id = $1`, id), &reg)
	if errors.Is(err, pgx.ErrNoRows) {
		return effort.Registro{}, effort.ErrRegistroNoEncontrado
	}
	if err != nil {
		return effort.Registro{}, fmt.Errorf("buscar el registro de esfuerzo %d: %w", id, err)
	}
	return reg, nil
}

// Actualizar guarda la tarea, la fecha, la actividad y las horas corregidas (HU-26, RN5), o devuelve
// effort.ErrRegistroNoEncontrado. La base rechaza horas fuera de (0, 24].
func (r *Registros) Actualizar(ctx context.Context, reg effort.Registro) error {
	var tarea *int64
	if reg.TareaID != 0 {
		tarea = &reg.TareaID
	}
	res, err := r.pool.Exec(ctx,
		`UPDATE registros_esfuerzo SET tarea_id = $2, fecha = $3, actividad = $4, horas = $5 WHERE id = $1`,
		reg.ID, tarea, reg.Fecha, reg.Actividad, reg.Horas)
	if err != nil {
		return fmt.Errorf("actualizar el registro de esfuerzo %d: %w", reg.ID, err)
	}
	if res.RowsAffected() == 0 {
		return effort.ErrRegistroNoEncontrado
	}
	return nil
}

// fechaONula devuelve nil si la fecha es cero, para que el filtro no la use.
func fechaONula(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// escanearRegistro lee una fila con las columnas de columnasRegistro.
func escanearRegistro(fila pgx.Row, reg *effort.Registro) error {
	return fila.Scan(&reg.ID, &reg.IntegranteID, &reg.HistoriaID, &reg.TareaID, &reg.Fecha, &reg.Actividad, &reg.Horas)
}

// leerRegistros lee todas las filas y las cierra.
func leerRegistros(filas pgx.Rows) ([]effort.Registro, error) {
	defer filas.Close()
	var lista []effort.Registro
	for filas.Next() {
		var reg effort.Registro
		if err := escanearRegistro(filas, &reg); err != nil {
			return nil, fmt.Errorf("leer un registro de esfuerzo: %w", err)
		}
		lista = append(lista, reg)
	}
	if err := filas.Err(); err != nil {
		return nil, fmt.Errorf("listar los registros de esfuerzo: %w", err)
	}
	return lista, nil
}
