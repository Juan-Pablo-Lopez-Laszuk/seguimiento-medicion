package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
)

// Tareas guarda las tareas de todas las historias en la tabla tareas (HU-25).
type Tareas struct {
	pool *pgxpool.Pool
}

// NuevoTareas crea el repositorio sobre un pool de conexiones ya abierto.
func NuevoTareas(pool *pgxpool.Pool) *Tareas {
	return &Tareas{pool: pool}
}

// columnasTarea son las columnas que se leen de una tarea, en el orden en que las espera escanearTarea.
const columnasTarea = `id, historia_id, titulo, COALESCE(responsable_id, 0), COALESCE(horas_estimadas, 0)::float8, completada`

// escanearTarea lee una fila con columnasTarea. Una tarea sin responsable o sin estimar (la tabla lo permite) queda
// con 0.
func escanearTarea(fila pgx.Row) (effort.Tarea, error) {
	var t effort.Tarea
	err := fila.Scan(&t.ID, &t.HistoriaID, &t.Titulo, &t.ResponsableID, &t.HorasEstimadas, &t.Completada)
	return t, err
}

// Guardar inserta la tarea y la devuelve con el ID que le asignó la base. La base rechaza una tarea de una historia o
// con un responsable que no existen (claves foráneas) y unas horas que no sean mayores a 0 (tareas_horas_ck).
func (r *Tareas) Guardar(ctx context.Context, t effort.Tarea) (effort.Tarea, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO tareas (historia_id, titulo, responsable_id, horas_estimadas, completada)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		t.HistoriaID, t.Titulo, t.ResponsableID, t.HorasEstimadas, t.Completada,
	).Scan(&t.ID)
	if err != nil {
		return effort.Tarea{}, fmt.Errorf("guardar la tarea: %w", err)
	}
	return t, nil
}

// ListarPorHistoria devuelve las tareas de la historia en el orden en que se cargaron (RN9).
func (r *Tareas) ListarPorHistoria(ctx context.Context, historiaID int64) ([]effort.Tarea, error) {
	filas, err := r.pool.Query(ctx,
		`SELECT `+columnasTarea+` FROM tareas WHERE historia_id = $1 ORDER BY id`, historiaID)
	if err != nil {
		return nil, fmt.Errorf("listar las tareas: %w", err)
	}
	defer filas.Close()

	var lista []effort.Tarea
	for filas.Next() {
		t, err := escanearTarea(filas)
		if err != nil {
			return nil, fmt.Errorf("leer una tarea: %w", err)
		}
		lista = append(lista, t)
	}
	if err := filas.Err(); err != nil {
		return nil, fmt.Errorf("listar las tareas: %w", err)
	}
	return lista, nil
}

// BuscarPorID devuelve la tarea dentro de su historia, o effort.ErrTareaNoEncontrada. Con el id de una tarea de otra
// historia también es "no encontrado" (CA-25.2).
func (r *Tareas) BuscarPorID(ctx context.Context, historiaID, id int64) (effort.Tarea, error) {
	t, err := escanearTarea(r.pool.QueryRow(ctx,
		`SELECT `+columnasTarea+` FROM tareas WHERE id = $1 AND historia_id = $2`, id, historiaID))
	if errors.Is(err, pgx.ErrNoRows) {
		return effort.Tarea{}, effort.ErrTareaNoEncontrada
	}
	if err != nil {
		return effort.Tarea{}, fmt.Errorf("buscar la tarea %d: %w", id, err)
	}
	return t, nil
}

// Actualizar guarda las horas y si la tarea está completada, o devuelve effort.ErrTareaNoEncontrada si no existe en
// esa historia. El título, el responsable y la historia no se modifican: editarlos no es parte de HU-25.
func (r *Tareas) Actualizar(ctx context.Context, t effort.Tarea) error {
	etiqueta, err := r.pool.Exec(ctx,
		`UPDATE tareas SET horas_estimadas = $1, completada = $2 WHERE id = $3 AND historia_id = $4`,
		t.HorasEstimadas, t.Completada, t.ID, t.HistoriaID)
	if err != nil {
		return fmt.Errorf("actualizar la tarea %d: %w", t.ID, err)
	}
	if etiqueta.RowsAffected() == 0 {
		return effort.ErrTareaNoEncontrada
	}
	return nil
}
