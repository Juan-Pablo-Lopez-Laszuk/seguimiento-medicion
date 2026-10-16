package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
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

// HU-06 · ActualizarDatos cambia título, descripción y prioridad de esa historia, y deja el número, el estado y los
// story points
func TestHistorias_ActualizarDatos(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	repo := postgres.NuevoHistorias(pool)
	puntos := 5
	original := historiaDeEjemplo(proyectoA, 1, "Login")
	original.Descripcion, original.Prioridad = "Vieja", backlog.PrioridadAlta
	original.Estado, original.StoryPoints = backlog.EstadoEnProgreso, &puntos
	guardada, err := repo.Guardar(ctx, original)
	if err != nil {
		t.Fatalf("guardar la historia: %v", err)
	}
	otra, err := repo.Guardar(ctx, historiaDeEjemplo(proyectoA, 2, "Otra"))
	if err != nil {
		t.Fatalf("guardar la otra historia: %v", err)
	}

	err = repo.ActualizarDatos(ctx, backlog.Historia{
		ID: guardada.ID, Titulo: "Login con Google", Descripcion: "Nueva", Prioridad: backlog.PrioridadBaja,
	})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	h, err := repo.BuscarPorNumero(ctx, proyectoA, 1)
	if err != nil {
		t.Fatalf("buscar la historia: %v", err)
	}
	if h.Titulo != "Login con Google" || h.Descripcion != "Nueva" || h.Prioridad != backlog.PrioridadBaja {
		t.Errorf("la historia quedó con %q / %q / %q", h.Titulo, h.Descripcion, h.Prioridad)
	}
	if h.Numero != 1 || h.Estado != backlog.EstadoEnProgreso || h.StoryPoints == nil || *h.StoryPoints != 5 {
		t.Errorf("cambió lo que no se editaba: HU-%d, estado %q, story points %v", h.Numero, h.Estado, h.StoryPoints)
	}
	o, _ := repo.BuscarPorNumero(ctx, proyectoA, 2)
	if o.ID != otra.ID || o.Titulo != "Otra" || o.Prioridad != backlog.PrioridadMedia {
		t.Errorf("la otra historia cambió: %+v", o)
	}
}

func TestHistorias_ActualizarDatos_HistoriaInexistente(t *testing.T) {
	pool := baseMigrada(t)

	err := postgres.NuevoHistorias(pool).ActualizarDatos(context.Background(), backlog.Historia{
		ID: 99, Titulo: "X", Prioridad: backlog.PrioridadAlta,
	})

	if !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("se esperaba ErrHistoriaNoEncontrada y se obtuvo: %v", err)
	}
}

// criteriosYTareasDe guarda 2 criterios y 1 tarea en la historia, para probar qué pasa con ellos al eliminarla.
func criteriosYTareasDe(t *testing.T, pool *pgxpool.Pool, proyectoID, historiaID int64) {
	t.Helper()
	ctx := context.Background()
	for _, descripcion := range []string{"Primer criterio", "Segundo criterio"} {
		if _, err := postgres.NuevoCriterios(pool).Guardar(ctx, backlog.Criterio{HistoriaID: historiaID, Descripcion: descripcion}); err != nil {
			t.Fatalf("guardar un criterio: %v", err)
		}
	}
	responsable, err := postgres.NuevoIntegrantes(pool).Guardar(ctx, integranteDeEjemplo(proyectoID, "Carolina", "caro@mail.com", member.RolProductBuilder))
	if err != nil {
		t.Fatalf("guardar la integrante: %v", err)
	}
	if _, err := postgres.NuevoTareas(pool).Guardar(ctx, effort.Tarea{HistoriaID: historiaID, Titulo: "Una tarea", ResponsableID: responsable.ID, HorasEstimadas: 2}); err != nil {
		t.Fatalf("guardar una tarea: %v", err)
	}
}

// cuantos cuenta las filas de la tabla que cumplen la condición.
func cuantos(t *testing.T, pool *pgxpool.Pool, tabla, condicion string, args ...any) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(context.Background(), "SELECT count(*) FROM "+tabla+" WHERE "+condicion, args...).Scan(&n); err != nil {
		t.Fatalf("contar en %s: %v", tabla, err)
	}
	return n
}

// HU-06 · Eliminar borra la historia, sus criterios y sus tareas, y las demás historias conservan su número (RN9, RN11)
func TestHistorias_Eliminar_BorraLaHistoriaConSusCriteriosYSusTareas(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	repo := postgres.NuevoHistorias(pool)
	a, _ := repo.Guardar(ctx, historiaDeEjemplo(proyectoA, 1, "Primera"))
	b, _ := repo.Guardar(ctx, historiaDeEjemplo(proyectoA, 2, "Segunda"))
	c, _ := repo.Guardar(ctx, historiaDeEjemplo(proyectoA, 3, "Tercera"))
	criteriosYTareasDe(t, pool, proyectoA, b.ID)

	if err := repo.Eliminar(ctx, b.ID); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	lista, _ := repo.ListarPorProyecto(ctx, proyectoA)
	if len(lista) != 2 || lista[0].ID != a.ID || lista[0].Numero != 1 || lista[1].ID != c.ID || lista[1].Numero != 3 {
		t.Errorf("se esperaban HU-1 y HU-3, y se obtuvo %+v", lista)
	}
	if n := cuantos(t, pool, "criterios_aceptacion", "historia_id = $1", b.ID); n != 0 {
		t.Errorf("quedaron %d criterios de la historia eliminada", n)
	}
	if n := cuantos(t, pool, "tareas", "historia_id = $1", b.ID); n != 0 {
		t.Errorf("quedaron %d tareas de la historia eliminada", n)
	}
}

func TestHistorias_Eliminar_NoTocaLosCriteriosNiLasTareasDeLasDemas(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	repo := postgres.NuevoHistorias(pool)
	a, _ := repo.Guardar(ctx, historiaDeEjemplo(proyectoA, 1, "A eliminar"))
	b, _ := repo.Guardar(ctx, historiaDeEjemplo(proyectoA, 2, "Se queda"))
	criteriosYTareasDe(t, pool, proyectoA, b.ID)

	if err := repo.Eliminar(ctx, a.ID); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if n := cuantos(t, pool, "criterios_aceptacion", "historia_id = $1", b.ID); n != 2 {
		t.Errorf("la otra historia tenía que conservar sus 2 criterios y tiene %d", n)
	}
	if n := cuantos(t, pool, "tareas", "historia_id = $1", b.ID); n != 1 {
		t.Errorf("la otra historia tenía que conservar su tarea y tiene %d", n)
	}
}

func TestHistorias_Eliminar_HistoriaInexistente(t *testing.T) {
	pool := baseMigrada(t)

	err := postgres.NuevoHistorias(pool).Eliminar(context.Background(), 99)

	if !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("se esperaba ErrHistoriaNoEncontrada y se obtuvo: %v", err)
	}
}

// HU-06 (RN10) · si la base rechaza el borrado por otros registros asociados (acá, una sesión de Planning Poker), no se
// borra nada: ni la historia ni sus criterios ni sus tareas
func TestHistorias_Eliminar_ConRegistrosAsociados_NoBorraNada(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	repo := postgres.NuevoHistorias(pool)
	h, _ := repo.Guardar(ctx, historiaDeEjemplo(proyectoA, 1, "Con una sesión de poker"))
	criteriosYTareasDe(t, pool, proyectoA, h.ID)
	if _, err := pool.Exec(ctx, `INSERT INTO sesiones_poker (historia_id, estado) VALUES ($1, 'Abierta')`, h.ID); err != nil {
		t.Fatalf("guardar la sesión de poker: %v", err)
	}

	err := repo.Eliminar(ctx, h.ID)

	if !errors.Is(err, backlog.ErrHistoriaConRegistros) {
		t.Fatalf("se esperaba ErrHistoriaConRegistros y se obtuvo: %v", err)
	}
	if _, err := repo.BuscarPorNumero(ctx, proyectoA, 1); err != nil {
		t.Errorf("la historia tenía que seguir: %v", err)
	}
	if n := cuantos(t, pool, "criterios_aceptacion", "historia_id = $1", h.ID); n != 2 {
		t.Errorf("los criterios tenían que seguir (2) y hay %d", n)
	}
	if n := cuantos(t, pool, "tareas", "historia_id = $1", h.ID); n != 1 {
		t.Errorf("la tarea tenía que seguir (1) y hay %d", n)
	}
}
