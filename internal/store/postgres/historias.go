package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Historias guarda las historias de todos los proyectos en la tabla historias.
type Historias struct {
	pool  *pgxpool.Pool
	ahora func() time.Time
}

// NuevoHistorias crea el repositorio sobre un pool de conexiones ya abierto.
func NuevoHistorias(pool *pgxpool.Pool) *Historias {
	return &Historias{pool: pool, ahora: time.Now}
}

// ConReloj cambia el reloj con el que se anota la hora de cada re-estimación (HU-16b): en los tests, uno fijo. Sin
// llamarlo se usa la hora actual.
func (r *Historias) ConReloj(ahora func() time.Time) *Historias {
	r.ahora = ahora
	return r
}

// Guardar inserta la historia, con su orden (HU-09), y la devuelve con el ID que le asignó la base. Los story points
// sin estimar se guardan como NULL. La base rechaza un número repetido dentro del proyecto (restricción única), un proyecto que
// no existe (clave foránea) y una prioridad o unos story points fuera de lo permitido (CHECK).
func (r *Historias) Guardar(ctx context.Context, h backlog.Historia) (backlog.Historia, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO historias (proyecto_id, numero, titulo, descripcion, prioridad, estado, story_points, orden)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		h.ProyectoID, h.Numero, h.Titulo, h.Descripcion, string(h.Prioridad), string(h.Estado), h.StoryPoints, h.Orden,
	).Scan(&h.ID)
	if err != nil {
		return backlog.Historia{}, fmt.Errorf("guardar la historia: %w", err)
	}
	return h, nil
}

// ListarPorProyecto devuelve las historias del proyecto en el orden en que se crearon.
func (r *Historias) ListarPorProyecto(ctx context.Context, proyectoID int64) ([]backlog.Historia, error) {
	filas, err := r.pool.Query(ctx,
		`SELECT id, proyecto_id, numero, titulo, descripcion, prioridad, estado, story_points, sprint_id, completada_en, orden
		 FROM historias WHERE proyecto_id = $1 ORDER BY numero`, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("listar las historias del proyecto: %w", err)
	}
	defer filas.Close()

	var lista []backlog.Historia
	for filas.Next() {
		var h backlog.Historia
		var prioridad, estado string
		if err := filas.Scan(&h.ID, &h.ProyectoID, &h.Numero, &h.Titulo, &h.Descripcion, &prioridad, &estado, &h.StoryPoints, &h.SprintID, &h.CompletadaEn, &h.Orden); err != nil {
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
		`SELECT id, proyecto_id, numero, titulo, descripcion, prioridad, estado, story_points, sprint_id, completada_en, orden
		 FROM historias WHERE proyecto_id = $1 AND numero = $2`, proyectoID, numero,
	).Scan(&h.ID, &h.ProyectoID, &h.Numero, &h.Titulo, &h.Descripcion, &prioridad, &estado, &h.StoryPoints, &h.SprintID, &h.CompletadaEn, &h.Orden)
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

// ActualizarSprint guarda juntos, en una sola sentencia, el estado y el sprint de la historia (nil = sin sprint) sin
// tocar los demás datos, o devuelve backlog.ErrHistoriaNoEncontrada si no existe. Así una historia nunca queda En
// sprint sin sprint (HU-12). La base rechaza un sprint que no existe (clave foránea).
func (r *Historias) ActualizarSprint(ctx context.Context, historiaID int64, estado backlog.Estado, sprintID *int64) error {
	etiqueta, err := r.pool.Exec(ctx, `UPDATE historias SET estado = $1, sprint_id = $2 WHERE id = $3`,
		string(estado), sprintID, historiaID)
	if err != nil {
		return fmt.Errorf("actualizar el estado y el sprint de la historia %d: %w", historiaID, err)
	}
	if etiqueta.RowsAffected() == 0 {
		return backlog.ErrHistoriaNoEncontrada
	}
	return nil
}

// ActualizarCompletada guarda juntos, en una sola sentencia, el estado y la fecha de completado de la historia sin
// tocar los demás datos (HU-13), o devuelve backlog.ErrHistoriaNoEncontrada si no existe.
func (r *Historias) ActualizarCompletada(ctx context.Context, historiaID int64, estado backlog.Estado, completadaEn *time.Time) error {
	etiqueta, err := r.pool.Exec(ctx, `UPDATE historias SET estado = $1, completada_en = $2 WHERE id = $3`,
		string(estado), completadaEn, historiaID)
	if err != nil {
		return fmt.Errorf("actualizar el estado y la fecha de completado de la historia %d: %w", historiaID, err)
	}
	if etiqueta.RowsAffected() == 0 {
		return backlog.ErrHistoriaNoEncontrada
	}
	return nil
}

// ActualizarStoryPoints guarda los story points de la historia sin tocar los demás datos, o devuelve
// backlog.ErrHistoriaNoEncontrada si no existe. La base rechaza un valor fuera de la escala (CHECK). Es el único lugar
// donde se escribe la estimación (HU-16), junto con el cierre de una sesión de Planning Poker (HU-22), que usa la misma
// guardarStoryPoints: si la historia ya tenía una y el valor es distinto, anota el cambio en el historial (HU-16b,
// CA-16.3). Todo ocurre en una transacción que bloquea la fila de la historia: si falla una cosa no queda la otra, y
// dos re-estimaciones a la vez no pierden un valor (cada una parte de donde terminó la anterior).
func (r *Historias) ActualizarStoryPoints(ctx context.Context, historiaID int64, puntos int) error {
	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		return guardarStoryPoints(ctx, tx, historiaID, puntos, r.ahora())
	})
	if errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		return backlog.ErrHistoriaNoEncontrada
	}
	if err != nil {
		return fmt.Errorf("actualizar los story points de la historia %d: %w", historiaID, err)
	}
	return nil
}

// guardarStoryPoints escribe los story points de la historia dentro de la transacción tx, bloqueando su fila, y anota
// el cambio en el historial si corresponde (HU-16b). Devuelve backlog.ErrHistoriaNoEncontrada si no existe. La usan
// ActualizarStoryPoints (HU-16) y Poker.Cerrar (HU-22).
func guardarStoryPoints(ctx context.Context, tx pgx.Tx, historiaID int64, puntos int, ahora time.Time) error {
	var actual *int
	err := tx.QueryRow(ctx, `SELECT story_points FROM historias WHERE id = $1 FOR UPDATE`, historiaID).Scan(&actual)
	if errors.Is(err, pgx.ErrNoRows) {
		return backlog.ErrHistoriaNoEncontrada
	}
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE historias SET story_points = $1 WHERE id = $2`, puntos, historiaID); err != nil {
		return err
	}
	cambio, anota := backlog.NuevaReestimacion(actual, puntos, ahora)
	if !anota {
		return nil
	}
	_, err = tx.Exec(ctx,
		`INSERT INTO estimaciones_historia (historia_id, valor_anterior, valor_nuevo, cambiada_en) VALUES ($1, $2, $3, $4)`,
		historiaID, cambio.Anterior, cambio.Nuevo, cambio.CambiadaEn)
	return err
}

// ListarEstimaciones devuelve el historial de re-estimaciones de la historia, la más reciente primero (HU-16b). Una
// historia que nunca se re-estimó, o que no existe, tiene el historial vacío.
func (r *Historias) ListarEstimaciones(ctx context.Context, historiaID int64) ([]backlog.Reestimacion, error) {
	filas, err := r.pool.Query(ctx,
		`SELECT valor_anterior, valor_nuevo, cambiada_en FROM estimaciones_historia
		 WHERE historia_id = $1 ORDER BY id DESC`, historiaID)
	if err != nil {
		return nil, fmt.Errorf("listar el historial de estimaciones de la historia %d: %w", historiaID, err)
	}
	defer filas.Close()
	var lista []backlog.Reestimacion
	for filas.Next() {
		var e backlog.Reestimacion
		if err := filas.Scan(&e.Anterior, &e.Nuevo, &e.CambiadaEn); err != nil {
			return nil, fmt.Errorf("leer una entrada del historial de estimaciones: %w", err)
		}
		lista = append(lista, e)
	}
	return lista, filas.Err()
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
		if _, err := tx.Exec(ctx, `DELETE FROM estimaciones_historia WHERE historia_id = $1`, historiaID); err != nil {
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

// ActualizarOrden guarda el orden de cada historia (por ID) sin tocar los demás datos (HU-09, RN5), todo en una
// transacción: si alguna no existe en el proyecto devuelve backlog.ErrHistoriaNoEncontrada y no se guarda ninguno.
func (r *Historias) ActualizarOrden(ctx context.Context, proyectoID int64, ordenes map[int64]int) error {
	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		for id, orden := range ordenes {
			etiqueta, err := tx.Exec(ctx, `UPDATE historias SET orden = $1 WHERE id = $2 AND proyecto_id = $3`, orden, id, proyectoID)
			if err != nil {
				return err
			}
			if etiqueta.RowsAffected() == 0 {
				return backlog.ErrHistoriaNoEncontrada
			}
		}
		return nil
	})
	if errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		return backlog.ErrHistoriaNoEncontrada
	}
	if err != nil {
		return fmt.Errorf("actualizar el orden de las historias del proyecto %d: %w", proyectoID, err)
	}
	return nil
}

// EnSprintCerrado dice si la historia es de un sprint cerrado (HU-06, RN6): si ese es su sprint actual
// (historias.sprint_id) o si aparece en la foto de cierre de ese sprint (sprint_historias), aunque no se haya terminado.
func (r *Historias) EnSprintCerrado(ctx context.Context, historiaID int64) (bool, error) {
	var cerrado bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM historias h JOIN sprints s ON s.id = h.sprint_id
		                WHERE h.id = $1 AND s.estado = $2)
		     OR EXISTS (SELECT 1 FROM sprint_historias sh JOIN sprints s ON s.id = sh.sprint_id
		                WHERE sh.historia_id = $1 AND s.estado = $2)`,
		historiaID, string(sprint.EstadoCerrado),
	).Scan(&cerrado)
	if err != nil {
		return false, fmt.Errorf("ver si la historia %d es de un sprint cerrado: %w", historiaID, err)
	}
	return cerrado, nil
}
