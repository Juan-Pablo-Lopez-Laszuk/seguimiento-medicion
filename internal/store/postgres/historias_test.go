package postgres_test

import (
	"context"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/postgres"
)

func historiaDeEjemplo(proyectoID int64, numero int, titulo string) backlog.Historia {
	return backlog.Historia{
		ProyectoID: proyectoID,
		Numero:     numero,
		Titulo:     titulo,
		Prioridad:  backlog.PrioridadMedia,
		Estado:     backlog.EstadoPendiente,
	}
}

func TestHistorias_GuardarAsignaIDsYListarPorProyecto(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, proyectoB := dosProyectos(t, pool)
	repo := postgres.NuevoHistorias(pool)

	a, err := repo.Guardar(ctx, historiaDeEjemplo(proyectoA, 1, "Primera"))
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if _, err := repo.Guardar(ctx, historiaDeEjemplo(proyectoB, 1, "De otro proyecto")); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	c, err := repo.Guardar(ctx, historiaDeEjemplo(proyectoA, 2, "Segunda"))
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
	if len(lista) != 2 || lista[0].Titulo != "Primera" || lista[1].Titulo != "Segunda" {
		t.Errorf("se esperaban solo las historias del proyecto A, en orden, y se obtuvo %+v", lista)
	}
}

func TestHistorias_GuardaYDevuelveTodosLosCampos(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	repo := postgres.NuevoHistorias(pool)
	puntos := 8
	original := backlog.Historia{
		ProyectoID:  proyectoA,
		Numero:      1,
		Titulo:      "Iniciar sesión con Google",
		Descripcion: "Para no tener que recordar otra contraseña.",
		Prioridad:   backlog.PrioridadAlta,
		Estado:      backlog.EstadoPendiente,
		StoryPoints: &puntos,
	}

	if _, err := repo.Guardar(ctx, original); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	lista, err := repo.ListarPorProyecto(ctx, proyectoA)
	if err != nil || len(lista) != 1 {
		t.Fatalf("se esperaba una historia y se obtuvo %+v, %v", lista, err)
	}
	leida := lista[0]

	if leida.ProyectoID != original.ProyectoID || leida.Numero != original.Numero || leida.Titulo != original.Titulo ||
		leida.Descripcion != original.Descripcion || leida.Prioridad != original.Prioridad || leida.Estado != original.Estado {
		t.Errorf("campos distintos: se guardó %+v y se leyó %+v", original, leida)
	}
	if leida.StoryPoints == nil || *leida.StoryPoints != 8 {
		t.Errorf("story points: se guardaron 8 y se leyó %v", leida.StoryPoints)
	}
}

// Sin estimar (NULL en la base) y estimada en 0 son cosas distintas, y las dos tienen que volver como se guardaron.
func TestHistorias_SinEstimarYEstimadaEnCero_SeDistinguen(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	repo := postgres.NuevoHistorias(pool)
	cero := 0
	sinEstimar := historiaDeEjemplo(proyectoA, 1, "Sin estimar")
	enCero := historiaDeEjemplo(proyectoA, 2, "Estimada en cero")
	enCero.StoryPoints = &cero
	for _, h := range []backlog.Historia{sinEstimar, enCero} {
		if _, err := repo.Guardar(ctx, h); err != nil {
			t.Fatalf("no se esperaba error: %v", err)
		}
	}

	lista, err := repo.ListarPorProyecto(ctx, proyectoA)
	if err != nil || len(lista) != 2 {
		t.Fatalf("se esperaban dos historias y se obtuvo %+v, %v", lista, err)
	}

	if lista[0].StoryPoints != nil {
		t.Errorf("la primera tenía que quedar sin estimar y quedó en %d", *lista[0].StoryPoints)
	}
	if lista[1].StoryPoints == nil || *lista[1].StoryPoints != 0 {
		t.Errorf("la segunda tenía que quedar en 0 y quedó %v", lista[1].StoryPoints)
	}
}

func TestHistorias_ProyectoSinHistorias_DevuelveListaVacia(t *testing.T) {
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	lista, err := postgres.NuevoHistorias(pool).ListarPorProyecto(context.Background(), proyectoA)
	if err != nil || len(lista) != 0 {
		t.Errorf("se esperaba una lista vacía sin error y se obtuvo %+v, %v", lista, err)
	}
}

// Si dos personas crean la historia a la vez, la restricción única (proyecto_id, numero) evita el duplicado.
func TestHistorias_LaBaseRechazaUnNumeroRepetidoEnElProyecto(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	repo := postgres.NuevoHistorias(pool)
	if _, err := repo.Guardar(ctx, historiaDeEjemplo(proyectoA, 1, "Primera")); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if _, err := repo.Guardar(ctx, historiaDeEjemplo(proyectoA, 1, "Repetida")); err == nil {
		t.Error("se esperaba un error por el número repetido y se guardó")
	}
	if lista, _ := repo.ListarPorProyecto(ctx, proyectoA); len(lista) != 1 {
		t.Errorf("tenía que quedar una sola historia y hay %d", len(lista))
	}
}

func TestHistorias_NoSeGuardaUnaHistoriaDeUnProyectoQueNoExiste(t *testing.T) {
	repo := postgres.NuevoHistorias(baseMigrada(t))

	if _, err := repo.Guardar(context.Background(), historiaDeEjemplo(999, 1, "Huérfana")); err == nil {
		t.Error("se esperaba un error por el proyecto inexistente y se guardó")
	}
}
