package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
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
		`SELECT id, proyecto_id, numero, titulo, descripcion, prioridad, estado, story_points, sprint_id, completada_en
		 FROM historias WHERE proyecto_id = $1 ORDER BY numero`, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("listar las historias del proyecto: %w", err)
	}
	defer filas.Close()

	var lista []backlog.Historia
	for filas.Next() {
		var h backlog.Historia
		var prioridad, estado string
		if err := filas.Scan(&h.ID, &h.ProyectoID, &h.Numero, &h.Titulo, &h.Descripcion, &prioridad, &estado, &h.StoryPoints, &h.SprintID, &h.CompletadaEn); err != nil {
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
		`SELECT id, proyecto_id, numero, titulo, descripcion, prioridad, estado, story_points, sprint_id, completada_en
		 FROM historias WHERE proyecto_id = $1 AND numero = $2`, proyectoID, numero,
	).Scan(&h.ID, &h.ProyectoID, &h.Numero, &h.Titulo, &h.Descripcion, &prioridad, &estado, &h.StoryPoints, &h.SprintID, &h.CompletadaEn)
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
// donde se escribe la estimación (HU-16): HU-16b lo va a extender para anotar también el historial.
func (r *Historias) ActualizarStoryPoints(ctx context.Context, historiaID int64, puntos int) error {
	etiqueta, err := r.pool.Exec(ctx, `UPDATE historias SET story_points = $1 WHERE id = $2`, puntos, historiaID)
	if err != nil {
		return fmt.Errorf("actualizar los story points de la historia %d: %w", historiaID, err)
	}
	if etiqueta.RowsAffected() == 0 {
		return backlog.ErrHistoriaNoEncontrada
	}
	return nil
}
