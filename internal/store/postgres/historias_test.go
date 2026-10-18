package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

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

// docs/modelo-datos.md pide una sola constante de la escala en Go y un test que compruebe que coincide con el
// CHECK de la migración: la base acepta exactamente los valores de la escala, ni uno más ni uno menos.
func TestHistorias_LaEscalaDelCodigoCoincideConElCheckDeLaMigracion(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	enLaEscala := map[int]bool{}
	for _, v := range backlog.EscalaStoryPoints {
		enLaEscala[v] = true
	}

	for puntos := -2; puntos <= 25; puntos++ {
		_, err := pool.Exec(ctx,
			`INSERT INTO historias (proyecto_id, numero, titulo, prioridad, story_points) VALUES ($1, $2, 'x', 'Alta', $3)`,
			proyectoA, puntos+100, puntos)

		if aceptada := err == nil; aceptada != enLaEscala[puntos] {
			t.Errorf("story points %d: la escala del código dice que vale=%v y la base lo aceptó=%v", puntos, enLaEscala[puntos], aceptada)
		}
	}
}

// Lo mismo para las prioridades: la base acepta las del código y rechaza cualquier otra.
func TestHistorias_LasPrioridadesDelCodigoCoincidenConElCheckDeLaMigracion(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)

	for i, p := range backlog.Prioridades {
		if _, err := pool.Exec(ctx,
			`INSERT INTO historias (proyecto_id, numero, titulo, prioridad) VALUES ($1, $2, 'x', $3)`, proyectoA, i+1, string(p)); err != nil {
			t.Errorf("la base rechazó la prioridad %q del código: %v", p, err)
		}
	}
	if _, err := pool.Exec(ctx,
		`INSERT INTO historias (proyecto_id, numero, titulo, prioridad) VALUES ($1, 99, 'x', 'Urgente')`, proyectoA); err == nil {
		t.Error("la base aceptó una prioridad que no está en el código")
	}
}

// HU-07: la página de una historia se pide por su número dentro del proyecto (HU-N)
func TestHistorias_BuscarPorNumero(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, proyectoB := dosProyectos(t, pool)
	repo := postgres.NuevoHistorias(pool)
	puntos := 5
	conPuntos := historiaDeEjemplo(proyectoB, 1, "HU-1 del proyecto B")
	conPuntos.StoryPoints = &puntos
	if _, err := repo.Guardar(ctx, historiaDeEjemplo(proyectoA, 1, "HU-1 del proyecto A")); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	guardada, err := repo.Guardar(ctx, conPuntos)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	h, err := repo.BuscarPorNumero(ctx, proyectoB, 1)
	if err != nil || h.ID != guardada.ID || h.Titulo != "HU-1 del proyecto B" || h.Estado != backlog.EstadoPendiente {
		t.Errorf("se esperaba la HU-1 del proyecto B y se obtuvo %+v, %v", h, err)
	}
	if h.StoryPoints == nil || *h.StoryPoints != 5 {
		t.Errorf("story points: se guardaron 5 y se leyó %v", h.StoryPoints)
	}
	if _, err := repo.BuscarPorNumero(ctx, proyectoA, 2); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("número que no existe: se esperaba backlog.ErrHistoriaNoEncontrada y se obtuvo %v", err)
	}
	if _, err := repo.BuscarPorNumero(ctx, 999, 1); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("proyecto que no existe: se esperaba backlog.ErrHistoriaNoEncontrada y se obtuvo %v", err)
	}
}

// HU-08: el estado de una historia se guarda sin tocar los demás datos
func TestHistorias_ActualizarEstado(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	repo := postgres.NuevoHistorias(pool)
	puntos := 8
	primera := historiaDeEjemplo(proyectoA, 1, "Primera")
	primera.StoryPoints = &puntos
	guardada, err := repo.Guardar(ctx, primera)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	otra, _ := repo.Guardar(ctx, historiaDeEjemplo(proyectoA, 2, "Segunda"))

	if err := repo.ActualizarEstado(ctx, guardada.ID, backlog.EstadoEnSprint); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	h, err := repo.BuscarPorNumero(ctx, proyectoA, 1)
	if err != nil || h.Estado != backlog.EstadoEnSprint || h.Titulo != "Primera" || h.StoryPoints == nil || *h.StoryPoints != 8 {
		t.Errorf("se esperaba la HU-1 En sprint con el resto de sus datos y se obtuvo %+v, %v", h, err)
	}
	if o, _ := repo.BuscarPorNumero(ctx, proyectoA, 2); o.ID != otra.ID || o.Estado != backlog.EstadoPendiente {
		t.Errorf("la HU-2 no tenía que cambiar y se obtuvo %+v", o)
	}
	if err := repo.ActualizarEstado(ctx, 99999, backlog.EstadoEnSprint); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("historia que no existe: se esperaba backlog.ErrHistoriaNoEncontrada y se obtuvo %v", err)
	}
}

// Lo mismo que con la escala y las prioridades: los estados del código son exactamente los que acepta el CHECK de la
// migración (historias_estado_ck).
func TestHistorias_LosEstadosDelCodigoCoincidenConElCheckDeLaMigracion(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)

	for i, e := range backlog.Estados {
		if _, err := pool.Exec(ctx,
			`INSERT INTO historias (proyecto_id, numero, titulo, prioridad, estado) VALUES ($1, $2, 'x', 'Alta', $3)`, proyectoA, i+1, string(e)); err != nil {
			t.Errorf("la base rechazó el estado %q del código: %v", e, err)
		}
	}
	for _, malo := range []string{"Terminada", "en sprint", "Hecha", ""} {
		if _, err := pool.Exec(ctx,
			`INSERT INTO historias (proyecto_id, numero, titulo, prioridad, estado) VALUES ($1, 99, 'x', 'Alta', $2)`, proyectoA, malo); err == nil {
			t.Errorf("la base aceptó el estado %q, que no está en el código", malo)
		}
	}
}

// HU-12: al asignar o quitar, el estado y el sprint se guardan juntos, y la historia se lee con su sprint
func TestHistorias_ActualizarSprint(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	s, err := postgres.NuevoSprints(pool).Guardar(ctx, sprintDeEjemplo(proyectoA, 1, "Primero"))
	if err != nil {
		t.Fatalf("guardar el sprint: %v", err)
	}
	repo := postgres.NuevoHistorias(pool)
	guardada, err := repo.Guardar(ctx, historiaDeEjemplo(proyectoA, 1, "Primera"))
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if guardada.SprintID != nil {
		t.Errorf("una historia nueva no tiene sprint y se obtuvo %d", *guardada.SprintID)
	}

	if err := repo.ActualizarSprint(ctx, guardada.ID, backlog.EstadoEnSprint, &s.ID); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	h, err := repo.BuscarPorNumero(ctx, proyectoA, 1)
	if err != nil || h.Estado != backlog.EstadoEnSprint || h.SprintID == nil || *h.SprintID != s.ID || h.Titulo != "Primera" {
		t.Errorf("se esperaba la HU-1 En sprint en el sprint %d y se obtuvo %+v, %v", s.ID, h, err)
	}
	if lista, _ := repo.ListarPorProyecto(ctx, proyectoA); len(lista) != 1 || lista[0].SprintID == nil || *lista[0].SprintID != s.ID {
		t.Errorf("la lista tiene que traer el sprint de la historia y se obtuvo %+v", lista)
	}

	if err := repo.ActualizarSprint(ctx, guardada.ID, backlog.EstadoPendiente, nil); err != nil {
		t.Fatalf("no se esperaba error al quitar: %v", err)
	}
	if h, _ := repo.BuscarPorNumero(ctx, proyectoA, 1); h.Estado != backlog.EstadoPendiente || h.SprintID != nil {
		t.Errorf("se esperaba la HU-1 Pendiente y sin sprint y se obtuvo %+v", h)
	}
	if err := repo.ActualizarSprint(ctx, 99999, backlog.EstadoEnSprint, &s.ID); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("historia que no existe: se esperaba backlog.ErrHistoriaNoEncontrada y se obtuvo %v", err)
	}
}

// HU-13: al completar, el estado y la fecha de completado se guardan juntos, y la historia se lee con su fecha
func TestHistorias_ActualizarCompletada(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	repo := postgres.NuevoHistorias(pool)
	guardada, err := repo.Guardar(ctx, historiaDeEjemplo(proyectoA, 1, "Primera"))
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if guardada.CompletadaEn != nil {
		t.Errorf("una historia nueva no está completada y se obtuvo %v", guardada.CompletadaEn)
	}
	completada := fecha(2026, time.October, 15)

	if err := repo.ActualizarCompletada(ctx, guardada.ID, backlog.EstadoHecho, &completada); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	h, err := repo.BuscarPorNumero(ctx, proyectoA, 1)
	if err != nil || h.Estado != backlog.EstadoHecho || h.CompletadaEn == nil || !h.CompletadaEn.Equal(completada) || h.Titulo != "Primera" {
		t.Errorf("se esperaba la HU-1 Hecha el 15/10/2026 y se obtuvo %+v, %v", h, err)
	}
	if lista, _ := repo.ListarPorProyecto(ctx, proyectoA); len(lista) != 1 || lista[0].CompletadaEn == nil || !lista[0].CompletadaEn.Equal(completada) {
		t.Errorf("la lista tiene que traer la fecha de completado y se obtuvo %+v", lista)
	}
	if err := repo.ActualizarCompletada(ctx, 99999, backlog.EstadoHecho, &completada); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("historia que no existe: se esperaba backlog.ErrHistoriaNoEncontrada y se obtuvo %v", err)
	}
}
