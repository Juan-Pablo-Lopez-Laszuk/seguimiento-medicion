package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/postgres"
)

// dosProyectos guarda dos proyectos en la base y devuelve sus IDs: los sprints necesitan un proyecto existente.
func dosProyectos(t *testing.T, pool *pgxpool.Pool) (int64, int64) {
	t.Helper()
	repo := postgres.NuevoProyectos(pool)
	a, err := repo.Guardar(context.Background(), proyectoDeEjemplo("Proyecto A"))
	if err != nil {
		t.Fatalf("guardar el proyecto A: %v", err)
	}
	b, err := repo.Guardar(context.Background(), proyectoDeEjemplo("Proyecto B"))
	if err != nil {
		t.Fatalf("guardar el proyecto B: %v", err)
	}
	return a.ID, b.ID
}

func sprintDeEjemplo(proyectoID int64, numero int, objetivo string) sprint.Sprint {
	inicio := fecha(2026, time.October, 5).AddDate(0, 0, 7*(numero-1))
	return sprint.Sprint{
		ProyectoID:  proyectoID,
		Numero:      numero,
		Objetivo:    objetivo,
		FechaInicio: inicio,
		FechaFin:    inicio.AddDate(0, 0, 6),
		Estado:      sprint.EstadoPlanificado,
	}
}

func TestSprints_GuardarAsignaIDsYListarPorProyecto(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, proyectoB := dosProyectos(t, pool)
	repo := postgres.NuevoSprints(pool)

	a, err := repo.Guardar(ctx, sprintDeEjemplo(proyectoA, 1, "Primero"))
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if _, err := repo.Guardar(ctx, sprintDeEjemplo(proyectoB, 1, "De otro proyecto")); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	c, err := repo.Guardar(ctx, sprintDeEjemplo(proyectoA, 2, "Segundo"))
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if a.ID == 0 || c.ID != a.ID+2 {
		t.Errorf("IDs: se esperaban correlativos (el del medio es del otro proyecto) y se obtuvo %d y %d", a.ID, c.ID)
	}
	lista, err := repo.ListarPorProyecto(ctx, proyectoA)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if len(lista) != 2 || lista[0].Objetivo != "Primero" || lista[1].Objetivo != "Segundo" {
		t.Errorf("se esperaban solo los sprints del proyecto A, en orden, y se obtuvo %+v", lista)
	}
}

func TestSprints_GuardaYDevuelveTodosLosCampos(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	repo := postgres.NuevoSprints(pool)
	original := sprintDeEjemplo(proyectoA, 1, "Primer MVP: crear proyectos y sprints")

	if _, err := repo.Guardar(ctx, original); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	lista, err := repo.ListarPorProyecto(ctx, proyectoA)
	if err != nil || len(lista) != 1 {
		t.Fatalf("se esperaba un sprint y se obtuvo %+v, %v", lista, err)
	}
	leido := lista[0]

	if leido.ProyectoID != original.ProyectoID || leido.Numero != original.Numero || leido.Objetivo != original.Objetivo || leido.Estado != original.Estado {
		t.Errorf("campos distintos: se guardó %+v y se leyó %+v", original, leido)
	}
	if !leido.FechaInicio.Equal(original.FechaInicio) || !leido.FechaFin.Equal(original.FechaFin) {
		t.Errorf("fechas distintas: se guardó %v-%v y se leyó %v-%v", original.FechaInicio, original.FechaFin, leido.FechaInicio, leido.FechaFin)
	}
}

func TestSprints_ProyectoSinSprints_DevuelveListaVacia(t *testing.T) {
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	lista, err := postgres.NuevoSprints(pool).ListarPorProyecto(context.Background(), proyectoA)
	if err != nil || len(lista) != 0 {
		t.Errorf("se esperaba una lista vacía sin error y se obtuvo %+v, %v", lista, err)
	}
}

// Si dos personas crean el sprint a la vez, la restricción única (proyecto_id, numero) evita el duplicado.
func TestSprints_LaBaseRechazaUnNumeroRepetidoEnElProyecto(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	repo := postgres.NuevoSprints(pool)
	if _, err := repo.Guardar(ctx, sprintDeEjemplo(proyectoA, 1, "Primero")); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if _, err := repo.Guardar(ctx, sprintDeEjemplo(proyectoA, 1, "Repetido")); err == nil {
		t.Error("se esperaba un error por el número repetido y se guardó")
	}
	if lista, _ := repo.ListarPorProyecto(ctx, proyectoA); len(lista) != 1 {
		t.Errorf("tenía que quedar un solo sprint y hay %d", len(lista))
	}
}

func TestSprints_NoSeGuardaUnSprintDeUnProyectoQueNoExiste(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NuevoSprints(baseMigrada(t))

	if _, err := repo.Guardar(ctx, sprintDeEjemplo(999, 1, "Huérfano")); err == nil {
		t.Error("se esperaba un error por el proyecto inexistente y se guardó")
	}
}
