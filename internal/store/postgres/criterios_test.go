package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/postgres"
)

// dosHistorias guarda dos historias del mismo proyecto y devuelve sus IDs: los criterios necesitan una historia
// existente.
func dosHistorias(t *testing.T, pool *pgxpool.Pool) (int64, int64) {
	t.Helper()
	proyectoA, _ := dosProyectos(t, pool)
	repo := postgres.NuevoHistorias(pool)
	a, err := repo.Guardar(context.Background(), historiaDeEjemplo(proyectoA, 1, "Primera"))
	if err != nil {
		t.Fatalf("guardar la primera historia: %v", err)
	}
	b, err := repo.Guardar(context.Background(), historiaDeEjemplo(proyectoA, 2, "Segunda"))
	if err != nil {
		t.Fatalf("guardar la segunda historia: %v", err)
	}
	return a.ID, b.ID
}

func TestCriterios_GuardarAsignaIDsYListarPorHistoriaEnOrden(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	historiaA, historiaB := dosHistorias(t, pool)
	repo := postgres.NuevoCriterios(pool)

	a, err := repo.Guardar(ctx, backlog.Criterio{HistoriaID: historiaA, Descripcion: "Primero"})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if _, err := repo.Guardar(ctx, backlog.Criterio{HistoriaID: historiaB, Descripcion: "De otra historia"}); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	c, err := repo.Guardar(ctx, backlog.Criterio{HistoriaID: historiaA, Descripcion: "Segundo"})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if a.ID == 0 || c.ID != a.ID+2 {
		t.Errorf("IDs: se esperaban correlativos (el del medio es de otra historia) y se obtuvo %d y %d", a.ID, c.ID)
	}
	lista, err := repo.ListarPorHistoria(ctx, historiaA)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if len(lista) != 2 || lista[0].Descripcion != "Primero" || lista[1].Descripcion != "Segundo" {
		t.Errorf("se esperaban solo los criterios de la historia A, en orden, y se obtuvo %+v", lista)
	}
}

func TestCriterios_GuardaYDevuelveTodosLosCampos(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	historiaA, _ := dosHistorias(t, pool)
	repo := postgres.NuevoCriterios(pool)

	if _, err := repo.Guardar(ctx, backlog.Criterio{HistoriaID: historiaA, Descripcion: "Entra con su cuenta de Google"}); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if _, err := repo.Guardar(ctx, backlog.Criterio{HistoriaID: historiaA, Descripcion: "Muestra su nombre", Cumplido: true}); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	lista, err := repo.ListarPorHistoria(ctx, historiaA)
	if err != nil || len(lista) != 2 {
		t.Fatalf("se esperaban dos criterios y se obtuvo %+v, %v", lista, err)
	}

	if lista[0].HistoriaID != historiaA || lista[0].Descripcion != "Entra con su cuenta de Google" || lista[0].Cumplido {
		t.Errorf("el primero tenía que ser no cumplido y se obtuvo %+v", lista[0])
	}
	if lista[1].Descripcion != "Muestra su nombre" || !lista[1].Cumplido {
		t.Errorf("el segundo tenía que ser cumplido y se obtuvo %+v", lista[1])
	}
}

func TestCriterios_HistoriaSinCriterios_DevuelveListaVacia(t *testing.T) {
	pool := baseMigrada(t)
	historiaA, _ := dosHistorias(t, pool)
	lista, err := postgres.NuevoCriterios(pool).ListarPorHistoria(context.Background(), historiaA)
	if err != nil || len(lista) != 0 {
		t.Errorf("se esperaba una lista vacía sin error y se obtuvo %+v, %v", lista, err)
	}
}

// Para el progreso de toda la lista del backlog se piden los criterios de varias historias con una sola consulta
func TestCriterios_ListarPorHistorias(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	historiaA, historiaB := dosHistorias(t, pool)
	repo := postgres.NuevoCriterios(pool)
	_, _ = repo.Guardar(ctx, backlog.Criterio{HistoriaID: historiaA, Descripcion: "De la A"})
	_, _ = repo.Guardar(ctx, backlog.Criterio{HistoriaID: historiaB, Descripcion: "De la B"})

	todas, err := repo.ListarPorHistorias(ctx, []int64{historiaA, historiaB})
	if err != nil || len(todas) != 2 || todas[0].Descripcion != "De la A" || todas[1].Descripcion != "De la B" {
		t.Errorf("se esperaban los dos criterios en orden y se obtuvo %+v, %v", todas, err)
	}
	soloB, err := repo.ListarPorHistorias(ctx, []int64{historiaB})
	if err != nil || len(soloB) != 1 || soloB[0].HistoriaID != historiaB {
		t.Errorf("se esperaba solo el criterio de la B y se obtuvo %+v, %v", soloB, err)
	}
	if vacia, err := repo.ListarPorHistorias(ctx, nil); err != nil || len(vacia) != 0 {
		t.Errorf("sin historias se esperaba una lista vacía sin error y se obtuvo %+v, %v", vacia, err)
	}
}

// RN8 · un criterio se busca dentro de su historia: con el id de otra historia es un "no encontrado"
func TestCriterios_BuscarPorID(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	historiaA, historiaB := dosHistorias(t, pool)
	repo := postgres.NuevoCriterios(pool)
	guardado, _ := repo.Guardar(ctx, backlog.Criterio{HistoriaID: historiaA, Descripcion: "Primero"})

	c, err := repo.BuscarPorID(ctx, historiaA, guardado.ID)
	if err != nil || c.Descripcion != "Primero" || c.HistoriaID != historiaA {
		t.Errorf("se esperaba el criterio y se obtuvo %+v, %v", c, err)
	}
	if _, err := repo.BuscarPorID(ctx, historiaB, guardado.ID); !errors.Is(err, backlog.ErrCriterioNoEncontrado) {
		t.Errorf("criterio de otra historia: se esperaba backlog.ErrCriterioNoEncontrado y se obtuvo %v", err)
	}
	if _, err := repo.BuscarPorID(ctx, historiaA, 99999); !errors.Is(err, backlog.ErrCriterioNoEncontrado) {
		t.Errorf("id que no existe: se esperaba backlog.ErrCriterioNoEncontrado y se obtuvo %v", err)
	}
}

func TestCriterios_Actualizar(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	historiaA, historiaB := dosHistorias(t, pool)
	repo := postgres.NuevoCriterios(pool)
	guardado, _ := repo.Guardar(ctx, backlog.Criterio{HistoriaID: historiaA, Descripcion: "Primero"})

	guardado.Cumplido = true
	if err := repo.Actualizar(ctx, guardado); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	lista, _ := repo.ListarPorHistoria(ctx, historiaA)
	if len(lista) != 1 || !lista[0].Cumplido {
		t.Errorf("se esperaba el criterio cumplido y se obtuvo %+v", lista)
	}

	guardado.Cumplido = false
	if err := repo.Actualizar(ctx, guardado); err != nil {
		t.Fatalf("no se esperaba error al desmarcar: %v", err)
	}
	if lista, _ := repo.ListarPorHistoria(ctx, historiaA); len(lista) != 1 || lista[0].Cumplido {
		t.Errorf("se esperaba el criterio no cumplido y se obtuvo %+v", lista)
	}

	if err := repo.Actualizar(ctx, backlog.Criterio{ID: 99999, HistoriaID: historiaA}); !errors.Is(err, backlog.ErrCriterioNoEncontrado) {
		t.Errorf("criterio que no existe: se esperaba backlog.ErrCriterioNoEncontrado y se obtuvo %v", err)
	}
	if err := repo.Actualizar(ctx, backlog.Criterio{ID: guardado.ID, HistoriaID: historiaB}); !errors.Is(err, backlog.ErrCriterioNoEncontrado) {
		t.Errorf("criterio de otra historia: se esperaba backlog.ErrCriterioNoEncontrado y se obtuvo %v", err)
	}
}

func TestCriterios_NoSeGuardaUnCriterioDeUnaHistoriaQueNoExiste(t *testing.T) {
	repo := postgres.NuevoCriterios(baseMigrada(t))

	if _, err := repo.Guardar(context.Background(), backlog.Criterio{HistoriaID: 999, Descripcion: "Huérfano"}); err == nil {
		t.Error("se esperaba un error por la historia inexistente y se guardó")
	}
}
