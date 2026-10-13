package postgres

import (
	"context"
	"fmt"

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
