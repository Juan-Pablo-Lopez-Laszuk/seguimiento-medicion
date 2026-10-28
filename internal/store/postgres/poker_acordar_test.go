package postgres_test

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/postgres"
)

// historiaConSesion guarda la historia Login del proyecto A, con los story points que se pidan (nil: sin estimar), y
// le abre una sesión de Planning Poker.
func historiaConSesion(t *testing.T, pool *pgxpool.Pool, puntos *int) (backlog.Historia, poker.Sesion) {
	t.Helper()
	ctx := context.Background()
	proyectoA, _ := dosProyectos(t, pool)
	historias := postgres.NuevoHistorias(pool)
	h, err := historias.Guardar(ctx, historiaDeEjemplo(proyectoA, 1, "Login"))
	if err != nil {
		t.Fatalf("guardar la historia: %v", err)
	}
	if puntos != nil {
		if err := historias.ActualizarStoryPoints(ctx, h.ID, *puntos); err != nil {
			t.Fatalf("estimar la historia: %v", err)
		}
	}
	sesion, _, err := postgres.NuevoPoker(pool).Abrir(ctx, h.ID)
	if err != nil {
		t.Fatalf("abrir la sesión: %v", err)
	}
	return h, sesion
}

func storyPointsDe(t *testing.T, pool *pgxpool.Pool, h backlog.Historia) *int {
	t.Helper()
	guardada, err := postgres.NuevoHistorias(pool).BuscarPorNumero(context.Background(), h.ProyectoID, h.Numero)
	if err != nil {
		t.Fatalf("buscar la historia: %v", err)
	}
	return guardada.StoryPoints
}

// CA-22.2 CA-22.3 · cerrar deja la sesión Cerrada con su valor y estima la historia; la primera estimación no se anota
func TestPoker_Cerrar_CierraLaSesionYEstimaLaHistoria(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	h, sesion := historiaConSesion(t, pool, nil)
	repo := postgres.NuevoPoker(pool)

	if err := repo.Cerrar(ctx, sesion.ID, h.ID, 5); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if _, _, err := repo.SesionAbierta(ctx, h.ID); !errors.Is(err, poker.ErrSesionNoEncontrada) {
		t.Errorf("SesionAbierta: error = %v, se esperaba ErrSesionNoEncontrada", err)
	}
	if n := cuantos(t, pool, "sesiones_poker", "id = $1 AND estado = 'Cerrada' AND valor_acordado = 5", sesion.ID); n != 1 {
		t.Error("la sesión tenía que quedar Cerrada con el valor 5")
	}
	if sp := storyPointsDe(t, pool, h); sp == nil || *sp != 5 {
		t.Errorf("story points = %v, se esperaba 5", sp)
	}
	if n := cuantos(t, pool, "estimaciones_historia", "historia_id = $1", h.ID); n != 0 {
		t.Errorf("la primera estimación no se anota y hay %d entradas", n)
	}
}

// CA-22.2 RN2 · re-estimar con Planning Poker anota el cambio en el historial de HU-16b; el mismo valor no anota nada
func TestPoker_Cerrar_AnotaLaReestimacion(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	tres := 3
	h, sesion := historiaConSesion(t, pool, &tres)
	repo := postgres.NuevoPoker(pool)

	if err := repo.Cerrar(ctx, sesion.ID, h.ID, 8); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	historial, err := postgres.NuevoHistorias(pool).ListarEstimaciones(ctx, h.ID)
	if err != nil || len(historial) != 1 || historial[0].Anterior != 3 || historial[0].Nuevo != 8 {
		t.Fatalf("historial = %+v, %v; se esperaba el cambio de 3 a 8", historial, err)
	}

	otra, _, _ := repo.Abrir(ctx, h.ID)
	if err := repo.Cerrar(ctx, otra.ID, h.ID, 8); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if n := cuantos(t, pool, "estimaciones_historia", "historia_id = $1", h.ID); n != 1 {
		t.Errorf("acordar el mismo valor no anota nada y hay %d entradas", n)
	}
}

// RN5 · una sesión cerrada no se vuelve a cerrar
func TestPoker_Cerrar_SesionYaCerrada(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	h, sesion := historiaConSesion(t, pool, nil)
	repo := postgres.NuevoPoker(pool)
	if err := repo.Cerrar(ctx, sesion.ID, h.ID, 5); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if err := repo.Cerrar(ctx, sesion.ID, h.ID, 8); !errors.Is(err, poker.ErrSesionCerrada) {
		t.Errorf("error = %v, se esperaba ErrSesionCerrada", err)
	}
	if sp := storyPointsDe(t, pool, h); *sp != 5 {
		t.Errorf("story points = %d, se esperaba 5 (el segundo cierre no cuenta)", *sp)
	}
}

// RN5 · si dos pedidos cierran la misma sesión a la vez, se registra uno solo
func TestPoker_CerrarDosVecesALaVez(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	h, sesion := historiaConSesion(t, pool, nil)
	repo := postgres.NuevoPoker(pool)

	valores := []int{5, 8}
	errs := make([]error, len(valores))
	var wg sync.WaitGroup
	for i, valor := range valores {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = repo.Cerrar(ctx, sesion.ID, h.ID, valor)
		}()
	}
	wg.Wait()

	ganador := -1
	for i, err := range errs {
		switch {
		case err == nil:
			ganador = i
		case !errors.Is(err, poker.ErrSesionCerrada):
			t.Errorf("error inesperado: %v", err)
		}
	}
	if ganador == -1 || errs[1-ganador] == nil {
		t.Fatalf("se esperaba un solo cierre y se obtuvo %v", errs)
	}
	if sp := storyPointsDe(t, pool, h); *sp != valores[ganador] {
		t.Errorf("story points = %d, se esperaba %d, el del cierre que se registró", *sp, valores[ganador])
	}
}

// RN6 · todo o nada: si no se pueden guardar los story points, la sesión sigue abierta
func TestPoker_Cerrar_LoQueFallaNoCierraLaSesion(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	h, sesion := historiaConSesion(t, pool, nil)
	repo := postgres.NuevoPoker(pool)

	if err := repo.Cerrar(ctx, sesion.ID, h.ID+1000, 5); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Fatalf("error = %v, se esperaba ErrHistoriaNoEncontrada", err)
	}
	if n := cuantos(t, pool, "sesiones_poker", "id = $1 AND estado = 'Abierta' AND valor_acordado IS NULL", sesion.ID); n != 1 {
		t.Error("la sesión tenía que seguir abierta y sin valor")
	}
}

// CA-22.3 · las sesiones cerradas se consultan, la más reciente primero, con su número, sus rondas y su valor
func TestPoker_SesionesCerradas(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	h, primera := historiaConSesion(t, pool, nil)
	repo := postgres.NuevoPoker(pool)
	otraHistoria, _ := postgres.NuevoHistorias(pool).Guardar(ctx, historiaDeEjemplo(h.ProyectoID, 2, "Registro"))

	if cerradas, err := repo.SesionesCerradas(ctx, h.ID); err != nil || len(cerradas) != 0 {
		t.Errorf("sin cerrar: %+v, %v; se esperaba la lista vacía", cerradas, err)
	}
	if err := repo.Cerrar(ctx, primera.ID, h.ID, 5); err != nil {
		t.Fatal(err)
	}
	segunda, _, _ := repo.Abrir(ctx, h.ID)
	if _, err := pool.Exec(ctx, `INSERT INTO rondas_poker (sesion_id, numero) VALUES ($1, 2)`, segunda.ID); err != nil {
		t.Fatalf("agregar la Ronda 2: %v", err)
	}
	if err := repo.Cerrar(ctx, segunda.ID, h.ID, 8); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repo.Abrir(ctx, h.ID); err != nil { // la tercera queda abierta: no se lista
		t.Fatal(err)
	}
	ajena, _, _ := repo.Abrir(ctx, otraHistoria.ID)
	if err := repo.Cerrar(ctx, ajena.ID, otraHistoria.ID, 13); err != nil {
		t.Fatal(err)
	}

	cerradas, err := repo.SesionesCerradas(ctx, h.ID)
	quiero := []poker.SesionCerrada{{Numero: 2, Rondas: 2, ValorAcordado: 8}, {Numero: 1, Rondas: 1, ValorAcordado: 5}}
	if err != nil || !slices.Equal(cerradas, quiero) {
		t.Errorf("SesionesCerradas = %+v, %v; se esperaba %+v", cerradas, err, quiero)
	}
}
