package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
)

// Proyectos guarda los proyectos en la tabla proyectos.
type Proyectos struct {
	pool *pgxpool.Pool
}

// NuevoProyectos crea el repositorio sobre un pool de conexiones ya abierto.
func NuevoProyectos(pool *pgxpool.Pool) *Proyectos {
	return &Proyectos{pool: pool}
}

const columnasProyecto = "id, nombre, descripcion, fecha_inicio, fecha_fin, estado, creado_en"

// Guardar inserta el proyecto y lo devuelve con el ID que le asignó la base. Si el nombre ya existe (sin
// distinguir mayúsculas), el índice único de la base lo rechaza.
func (r *Proyectos) Guardar(ctx context.Context, p project.Proyecto) (project.Proyecto, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO proyectos (nombre, descripcion, fecha_inicio, fecha_fin, estado, creado_en)
		 VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		p.Nombre, p.Descripcion, p.FechaInicio, p.FechaFin, string(p.Estado), p.CreadoEn,
	).Scan(&p.ID)
	if err != nil {
		return project.Proyecto{}, fmt.Errorf("guardar el proyecto: %w", err)
	}
	return p, nil
}

// Listar devuelve todos los proyectos en el orden en que se crearon.
func (r *Proyectos) Listar(ctx context.Context) ([]project.Proyecto, error) {
	filas, err := r.pool.Query(ctx, `SELECT `+columnasProyecto+` FROM proyectos ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("listar los proyectos: %w", err)
	}
	defer filas.Close()

	var lista []project.Proyecto
	for filas.Next() {
		p, err := leerProyecto(filas)
		if err != nil {
			return nil, fmt.Errorf("leer un proyecto: %w", err)
		}
		lista = append(lista, p)
	}
	if err := filas.Err(); err != nil {
		return nil, fmt.Errorf("listar los proyectos: %w", err)
	}
	return lista, nil
}

// ExisteNombre informa si ya hay un proyecto con ese nombre, sin distinguir mayúsculas (RN3). Compara igual
// que el índice único de la base: lower(nombre).
func (r *Proyectos) ExisteNombre(ctx context.Context, nombre string) (bool, error) {
	var existe bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM proyectos WHERE lower(nombre) = lower($1))`, nombre,
	).Scan(&existe)
	if err != nil {
		return false, fmt.Errorf("buscar el nombre del proyecto: %w", err)
	}
	return existe, nil
}

// BuscarPorID devuelve el proyecto con ese ID, o project.ErrNoEncontrado si no existe.
func (r *Proyectos) BuscarPorID(ctx context.Context, id int64) (project.Proyecto, error) {
	p, err := leerProyecto(r.pool.QueryRow(ctx, `SELECT `+columnasProyecto+` FROM proyectos WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return project.Proyecto{}, project.ErrNoEncontrado
	}
	if err != nil {
		return project.Proyecto{}, fmt.Errorf("buscar el proyecto: %w", err)
	}
	return p, nil
}

// leerProyecto lee una fila con las columnas de columnasProyecto, en ese orden.
func leerProyecto(fila pgx.Row) (project.Proyecto, error) {
	var p project.Proyecto
	var estado string
	if err := fila.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.FechaInicio, &p.FechaFin, &estado, &p.CreadoEn); err != nil {
		return project.Proyecto{}, err
	}
	p.Estado = project.Estado(estado)
	return p, nil
}
