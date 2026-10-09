package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
)

// Integrantes guarda los integrantes de todos los proyectos en la tabla integrantes.
type Integrantes struct {
	pool *pgxpool.Pool
}

// NuevoIntegrantes crea el repositorio sobre un pool de conexiones ya abierto.
func NuevoIntegrantes(pool *pgxpool.Pool) *Integrantes {
	return &Integrantes{pool: pool}
}

const columnasIntegrante = "id, proyecto_id, nombre, email, rol, activo"

// Guardar inserta el integrante y lo devuelve con el ID que le asignó la base. La base rechaza el email repetido
// dentro del proyecto (también el de alguien dado de baja), un segundo Agile Enabler activo y un proyecto que no existe.
func (r *Integrantes) Guardar(ctx context.Context, i member.Integrante) (member.Integrante, error) {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO integrantes (proyecto_id, nombre, email, rol, activo)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		i.ProyectoID, i.Nombre, i.Email, string(i.Rol), i.Activo,
	).Scan(&i.ID)
	if err != nil {
		return member.Integrante{}, fmt.Errorf("guardar el integrante: %w", err)
	}
	return i, nil
}

// ListarPorProyecto devuelve los integrantes del proyecto, activos y dados de baja, en el orden en que se registraron.
func (r *Integrantes) ListarPorProyecto(ctx context.Context, proyectoID int64) ([]member.Integrante, error) {
	filas, err := r.pool.Query(ctx,
		`SELECT `+columnasIntegrante+` FROM integrantes WHERE proyecto_id = $1 ORDER BY id`, proyectoID)
	if err != nil {
		return nil, fmt.Errorf("listar los integrantes del proyecto: %w", err)
	}
	defer filas.Close()

	var lista []member.Integrante
	for filas.Next() {
		i, err := leerIntegrante(filas)
		if err != nil {
			return nil, fmt.Errorf("leer un integrante: %w", err)
		}
		lista = append(lista, i)
	}
	if err := filas.Err(); err != nil {
		return nil, fmt.Errorf("listar los integrantes del proyecto: %w", err)
	}
	return lista, nil
}

// BuscarPorID devuelve el integrante con ese ID dentro del proyecto, o member.ErrNoEncontrado.
func (r *Integrantes) BuscarPorID(ctx context.Context, proyectoID, id int64) (member.Integrante, error) {
	i, err := leerIntegrante(r.pool.QueryRow(ctx,
		`SELECT `+columnasIntegrante+` FROM integrantes WHERE proyecto_id = $1 AND id = $2`, proyectoID, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return member.Integrante{}, member.ErrNoEncontrado
	}
	if err != nil {
		return member.Integrante{}, fmt.Errorf("buscar el integrante: %w", err)
	}
	return i, nil
}

// Actualizar guarda los cambios del integrante con el mismo ID y proyecto, o devuelve member.ErrNoEncontrado.
func (r *Integrantes) Actualizar(ctx context.Context, i member.Integrante) error {
	res, err := r.pool.Exec(ctx,
		`UPDATE integrantes SET nombre = $1, email = $2, rol = $3, activo = $4
		 WHERE id = $5 AND proyecto_id = $6`,
		i.Nombre, i.Email, string(i.Rol), i.Activo, i.ID, i.ProyectoID)
	if err != nil {
		return fmt.Errorf("actualizar el integrante: %w", err)
	}
	if res.RowsAffected() == 0 {
		return member.ErrNoEncontrado
	}
	return nil
}

// leerIntegrante lee una fila con las columnas de columnasIntegrante, en ese orden.
func leerIntegrante(fila pgx.Row) (member.Integrante, error) {
	var i member.Integrante
	var rol string
	if err := fila.Scan(&i.ID, &i.ProyectoID, &i.Nombre, &i.Email, &rol, &i.Activo); err != nil {
		return member.Integrante{}, err
	}
	i.Rol = member.Rol(rol)
	return i, nil
}
