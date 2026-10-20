package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
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
	filas, err := r.pool.Query(ctx, `SELECT `+columnasDefecto+` FROM defectos WHERE proyecto_id = $1 ORDER BY id DESC`, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("listar los defectos: %w", err)
	}
	defer filas.Close()
	var lista []defect.Defecto
	for filas.Next() {
		d, err := escanearDefecto(filas)
		if err != nil {
			return nil, fmt.Errorf("leer un defecto: %w", err)
		}
		lista = append(lista, d)
	}
	if err := filas.Err(); err != nil {
		return nil, fmt.Errorf("listar los defectos: %w", err)
	}
	return lista, nil
}

// columnasDefecto son las columnas que lee escanearDefecto, en su orden.
const columnasDefecto = `id, proyecto_id, COALESCE(historia_id, 0), descripcion, severidad, estado, sprint_deteccion_id,
	COALESCE(sprint_resolucion_id, 0)`

func escanearDefecto(fila pgx.Row) (defect.Defecto, error) {
	var d defect.Defecto
	var severidad, estado string
	err := fila.Scan(&d.ID, &d.ProyectoID, &d.HistoriaID, &d.Descripcion, &severidad, &estado, &d.SprintDeteccionID, &d.SprintResolucionID)
	d.Severidad, d.Estado = defect.Severidad(severidad), defect.Estado(estado)
	return d, err
}

// BuscarPorID devuelve el defecto dentro de su proyecto, o defect.ErrDefectoNoEncontrado.
func (r *Defectos) BuscarPorID(ctx context.Context, proyectoID, id int64) (defect.Defecto, error) {
	d, err := escanearDefecto(r.pool.QueryRow(ctx,
		`SELECT `+columnasDefecto+` FROM defectos WHERE id = $1 AND proyecto_id = $2`, id, proyectoID))
	if errors.Is(err, pgx.ErrNoRows) {
		return defect.Defecto{}, defect.ErrDefectoNoEncontrado
	}
	if err != nil {
		return defect.Defecto{}, fmt.Errorf("buscar el defecto %d: %w", id, err)
	}
	return d, nil
}

// Actualizar guarda el estado y el sprint de resolución (NULL si es 0) de un defecto que ya existe (HU-28), o devuelve
// defect.ErrDefectoNoEncontrado.
func (r *Defectos) Actualizar(ctx context.Context, d defect.Defecto) error {
	etiqueta, err := r.pool.Exec(ctx,
		`UPDATE defectos SET estado = $1, sprint_resolucion_id = $2 WHERE id = $3 AND proyecto_id = $4`,
		string(d.Estado), opcional(d.SprintResolucionID), d.ID, d.ProyectoID)
	if err != nil {
		return fmt.Errorf("actualizar el defecto %d: %w", d.ID, err)
	}
	if etiqueta.RowsAffected() == 0 {
		return defect.ErrDefectoNoEncontrado
	}
	return nil
}

// ListarFiltrados devuelve los defectos del proyecto que cumplen el filtro, el más nuevo primero (HU-29). Los valores
// del filtro van como parámetros: un campo vacío (o 0) no filtra.
func (r *Defectos) ListarFiltrados(ctx context.Context, proyectoID int64, f defect.Filtro) ([]defect.Defecto, error) {
	filas, err := r.pool.Query(ctx, `SELECT `+columnasDefecto+` FROM defectos
		WHERE proyecto_id = $1 AND ($2 = '' OR severidad = $2) AND ($3 = '' OR estado = $3) AND ($4 = 0 OR sprint_deteccion_id = $4)
		ORDER BY id DESC`, proyectoID, string(f.Severidad), string(f.Estado), f.SprintID)
	if err != nil {
		return nil, fmt.Errorf("listar los defectos filtrados del proyecto %d: %w", proyectoID, err)
	}
	defer filas.Close()
	var lista []defect.Defecto
	for filas.Next() {
		d, err := escanearDefecto(filas)
		if err != nil {
			return nil, err
		}
		lista = append(lista, d)
	}
	return lista, filas.Err()
}
