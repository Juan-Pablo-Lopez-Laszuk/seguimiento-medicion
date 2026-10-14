package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/postgres"
)

// historiasConResponsable guarda dos historias y una integrante del mismo proyecto, y devuelve sus IDs: una tarea
// necesita una historia y un responsable que existan.
func historiasConResponsable(t *testing.T, pool *pgxpool.Pool) (historiaA, historiaB, responsable int64) {
	t.Helper()
	ctx := context.Background()
	proyectoA, _ := dosProyectos(t, pool)
	historias := postgres.NuevoHistorias(pool)
	a, err := historias.Guardar(ctx, historiaDeEjemplo(proyectoA, 1, "Primera"))
	if err != nil {
		t.Fatalf("guardar la primera historia: %v", err)
	}
	b, err := historias.Guardar(ctx, historiaDeEjemplo(proyectoA, 2, "Segunda"))
	if err != nil {
		t.Fatalf("guardar la segunda historia: %v", err)
	}
	i, err := postgres.NuevoIntegrantes(pool).Guardar(ctx, integranteDeEjemplo(proyectoA, "Carolina", "caro@mail.com", member.RolProductBuilder))
	if err != nil {
		t.Fatalf("guardar la integrante: %v", err)
	}
	return a.ID, b.ID, i.ID
}

// RN4 y RN9 · cada historia lista solo sus tareas, en el orden en que se cargaron
func TestTareas_GuardarAsignaIDsYListarPorHistoriaEnOrden(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	historiaA, historiaB, responsable := historiasConResponsable(t, pool)
	repo := postgres.NuevoTareas(pool)

	a, err := repo.Guardar(ctx, effort.Tarea{HistoriaID: historiaA, Titulo: "Primera", ResponsableID: responsable, HorasEstimadas: 3})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if _, err := repo.Guardar(ctx, effort.Tarea{HistoriaID: historiaB, Titulo: "De otra historia", ResponsableID: responsable, HorasEstimadas: 1}); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	c, err := repo.Guardar(ctx, effort.Tarea{HistoriaID: historiaA, Titulo: "Segunda", ResponsableID: responsable, HorasEstimadas: 2.5})
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
	if len(lista) != 2 || lista[0] != a || lista[1] != c {
		t.Errorf("se esperaban las dos tareas de la historia A, en orden y con todos sus campos, y se obtuvo %+v", lista)
	}
}

func TestTareas_HistoriaSinTareas_DevuelveListaVacia(t *testing.T) {
	pool := baseMigrada(t)
	historiaA, _, _ := historiasConResponsable(t, pool)

	lista, err := postgres.NuevoTareas(pool).ListarPorHistoria(context.Background(), historiaA)
	if err != nil || len(lista) != 0 {
		t.Errorf("se esperaba una lista vacía sin error y se obtuvo %+v, %v", lista, err)
	}
}

// CA-25.2 · una tarea se busca dentro de su historia: con el id de otra historia es un "no encontrado"
func TestTareas_BuscarPorID(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	historiaA, historiaB, responsable := historiasConResponsable(t, pool)
	repo := postgres.NuevoTareas(pool)
	guardada, err := repo.Guardar(ctx, effort.Tarea{HistoriaID: historiaA, Titulo: "Primera", ResponsableID: responsable, HorasEstimadas: 0.25})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	tarea, err := repo.BuscarPorID(ctx, historiaA, guardada.ID)
	if err != nil || tarea != guardada {
		t.Errorf("se esperaba %+v y se obtuvo %+v, %v", guardada, tarea, err)
	}
	if _, err := repo.BuscarPorID(ctx, historiaB, guardada.ID); !errors.Is(err, effort.ErrTareaNoEncontrada) {
		t.Errorf("tarea de otra historia: se esperaba effort.ErrTareaNoEncontrada y se obtuvo %v", err)
	}
	if _, err := repo.BuscarPorID(ctx, historiaA, guardada.ID+99); !errors.Is(err, effort.ErrTareaNoEncontrada) {
		t.Errorf("id que no existe: se esperaba effort.ErrTareaNoEncontrada y se obtuvo %v", err)
	}
}

// Actualizar guarda las horas y si está completada, sin poder moverla a otra historia (CA-25.2).
func TestTareas_Actualizar(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	historiaA, historiaB, responsable := historiasConResponsable(t, pool)
	repo := postgres.NuevoTareas(pool)
	guardada, err := repo.Guardar(ctx, effort.Tarea{HistoriaID: historiaA, Titulo: "Primera", ResponsableID: responsable, HorasEstimadas: 3})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	guardada.HorasEstimadas, guardada.Completada = 4.5, true
	if err := repo.Actualizar(ctx, guardada); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if tarea, _ := repo.BuscarPorID(ctx, historiaA, guardada.ID); tarea != guardada {
		t.Errorf("se esperaba %+v y quedó %+v", guardada, tarea)
	}
	otraHistoria := guardada
	otraHistoria.HistoriaID = historiaB
	if err := repo.Actualizar(ctx, otraHistoria); !errors.Is(err, effort.ErrTareaNoEncontrada) {
		t.Errorf("con otra historia: se esperaba effort.ErrTareaNoEncontrada y se obtuvo %v", err)
	}
}

// La base controla lo mismo que el dominio: la historia y el responsable tienen que existir, y las horas son > 0.
func TestTareas_LaBaseRechazaDatosInvalidos(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	historiaA, _, responsable := historiasConResponsable(t, pool)
	repo := postgres.NuevoTareas(pool)

	casos := map[string]effort.Tarea{
		"historia que no existe":    {HistoriaID: historiaA + 99, Titulo: "T", ResponsableID: responsable, HorasEstimadas: 1},
		"responsable que no existe": {HistoriaID: historiaA, Titulo: "T", ResponsableID: responsable + 99, HorasEstimadas: 1},
		"horas en cero":             {HistoriaID: historiaA, Titulo: "T", ResponsableID: responsable, HorasEstimadas: 0},
	}
	for nombre, tarea := range casos {
		if _, err := repo.Guardar(ctx, tarea); err == nil {
			t.Errorf("%s: se esperaba que la base la rechace", nombre)
		}
	}
}
