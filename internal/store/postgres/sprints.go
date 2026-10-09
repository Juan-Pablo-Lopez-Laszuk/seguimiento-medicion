package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
)

// Sprints guarda los sprints de todos los proyectos en la tabla sprints.
type Sprints struct {
	pool *pgxpool.Pool
}

// NuevoSprints crea el repositorio sobre un pool de conexiones ya abierto.
func NuevoSprints(pool *pgxpool.Pool) *Sprints {
	return &Sprints{pool: pool}
}

// Guardar inserta el sprint y lo devuelve con el ID que le asignó la base. La base rechaza un número repetido
// dentro del proyecto (restricción única) y un proyecto que no existe (clave foránea).
func (r *Sprints) Guardar(ctx context.Context, s sprint.Sprint) (sprint.Sprint, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO sprints (proyecto_id, numero, objetivo, fecha_inicio, fecha_fin, estado)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		s.ProyectoID, s.Numero, s.Objetivo, s.FechaInicio, s.FechaFin, string(s.Estado),
	).Scan(&s.ID)
	if err != nil {
		return sprint.Sprint{}, fmt.Errorf("guardar el sprint: %w", err)
	}
	return s, nil
}

// ListarPorProyecto devuelve los sprints del proyecto en el orden en que se crearon.
func (r *Sprints) ListarPorProyecto(ctx context.Context, proyectoID int64) ([]sprint.Sprint, error) {
	filas, err := r.pool.Query(ctx,
		`SELECT id, proyecto_id, numero, objetivo, fecha_inicio, fecha_fin, estado
		 FROM sprints WHERE proyecto_id = $1 ORDER BY id`, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("listar los sprints del proyecto: %w", err)
	}
	defer filas.Close()

	var lista []sprint.Sprint
	for filas.Next() {
		var s sprint.Sprint
		var estado string
		if err := filas.Scan(&s.ID, &s.ProyectoID, &s.Numero, &s.Objetivo, &s.FechaInicio, &s.FechaFin, &estado); err != nil {
			return nil, fmt.Errorf("leer un sprint: %w", err)
		}
		s.Estado = sprint.Estado(estado)
		lista = append(lista, s)
	}
	if err := filas.Err(); err != nil {
		return nil, fmt.Errorf("listar los sprints del proyecto: %w", err)
	}
	return lista, nil
}
