package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/defect"
)

// Defectos guarda los defectos en la tabla defectos (HU-27).
type Defectos struct {
	pool *pgxpool.Pool
}

// NuevoDefectos crea el repositorio sobre un pool de conexiones ya abierto.
func NuevoDefectos(pool *pgxpool.Pool) *Defectos {
	return &Defectos{pool: pool}
}

// opcional pasa un id 0 a NULL.
func opcional(id int64) *int64 {
	if id == 0 {
		return nil
	}
	return &id
}

// Guardar inserta el defecto y lo devuelve con su ID. La base rechaza un proyecto, sprint o historia que no existan.
func (r *Defectos) Guardar(ctx context.Context, d defect.Defecto) (defect.Defecto, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO defectos (proyecto_id, historia_id, descripcion, severidad, estado, sprint_deteccion_id, sprint_resolucion_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		d.ProyectoID, opcional(d.HistoriaID), d.Descripcion, string(d.Severidad), string(d.Estado), d.SprintDeteccionID, opcional(d.SprintResolucionID),
	).Scan(&d.ID)
	if err != nil {
		return defect.Defecto{}, fmt.Errorf("guardar el defecto: %w", err)
	}
	return d, nil
}

// ListarPorProyecto devuelve los defectos del proyecto, el más nuevo primero.
func (r *Defectos) ListarPorProyecto(ctx context.Context, proyectoID int64) ([]defect.Defecto, error) {
	filas, err := r.pool.Query(ctx,
		`SELECT id, proyecto_id, COALESCE(historia_id, 0), descripcion, severidad, estado, sprint_deteccion_id,
		        COALESCE(sprint_resolucion_id, 0)
		 FROM defectos WHERE proyecto_id = $1 ORDER BY id DESC`, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("listar los defectos: %w", err)
	}
	defer filas.Close()
	var lista []defect.Defecto
	for filas.Next() {
		var d defect.Defecto
		var severidad, estado string
		if err := filas.Scan(&d.ID, &d.ProyectoID, &d.HistoriaID, &d.Descripcion, &severidad, &estado, &d.SprintDeteccionID, &d.SprintResolucionID); err != nil {
			return nil, fmt.Errorf("leer un defecto: %w", err)
		}
		d.Severidad, d.Estado = defect.Severidad(severidad), defect.Estado(estado)
		lista = append(lista, d)
	}
	if err := filas.Err(); err != nil {
		return nil, fmt.Errorf("listar los defectos: %w", err)
	}
	return lista, nil
}
