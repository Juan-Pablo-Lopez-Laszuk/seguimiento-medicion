package postgres_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

// relojDeEjemplo devuelve un reloj que avanza una hora por cada vez que se lo consulta, empezando el 20/10/2026 a las
// 18:00, para saber qué hora tiene cada cambio.
func relojDeEjemplo() func() time.Time {
	var mu sync.Mutex
	hora := time.Date(2026, 10, 20, 17, 0, 0, 0, time.UTC)
	return func() time.Time {
		mu.Lock()
		defer mu.Unlock()
		hora = hora.Add(time.Hour)
		return hora
	}
}

// historiaParaReestimar guarda en la base una historia sin estimar y devuelve el repositorio, con el reloj de ejemplo.
func historiaParaReestimar(t *testing.T, pool *pgxpool.Pool, proyectoID int64, numero int) (*postgres.Historias, backlog.Historia) {
	t.Helper()
	repo := postgres.NuevoHistorias(pool).ConReloj(relojDeEjemplo())
	h, err := repo.Guardar(context.Background(), historiaDeEjemplo(proyectoID, numero, "Historia"))
	if err != nil {
		t.Fatalf("no se esperaba error al guardar: %v", err)
	}
	return repo, h
}

// CA-16.3 y RN1 · re-estimar anota el valor anterior, el nuevo y la hora; la lista va de la más reciente a la más antigua
func TestHistorias_Reestimar_AnotaElValorAnterior(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyecto, _ := dosProyectos(t, pool)
	repo, h := historiaParaReestimar(t, pool, proyecto, 1)
	for _, puntos := range []int{5, 8, 3} {
		if err := repo.ActualizarStoryPoints(ctx, h.ID, puntos); err != nil {
			t.Fatalf("%d story points: no se esperaba error: %v", puntos, err)
		}
	}

	historial, err := repo.ListarEstimaciones(ctx, h.ID)

	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	// el reloj se consulta una vez por cada estimación: 18:00, 19:00 y 20:00
	esperado := []backlog.Reestimacion{
		{Anterior: 8, Nuevo: 3, CambiadaEn: time.Date(2026, 10, 20, 20, 0, 0, 0, time.UTC)},
		{Anterior: 5, Nuevo: 8, CambiadaEn: time.Date(2026, 10, 20, 19, 0, 0, 0, time.UTC)},
	}
	if len(historial) != len(esperado) {
		t.Fatalf("se esperaban %d entradas y se obtuvieron %d: %+v", len(esperado), len(historial), historial)
	}
	for i := range esperado {
		if historial[i].Anterior != esperado[i].Anterior || historial[i].Nuevo != esperado[i].Nuevo || !historial[i].CambiadaEn.Equal(esperado[i].CambiadaEn) {
			t.Errorf("entrada %d: se esperaba %+v y se obtuvo %+v", i, esperado[i], historial[i])
		}
	}
	if h, _ := repo.BuscarPorNumero(ctx, proyecto, 1); h.StoryPoints == nil || *h.StoryPoints != 3 {
		t.Errorf("la historia tenía que quedar con 3 story points y quedó con %v", h.StoryPoints)
	}
}

// RN2, RN3 y RN4 · la primera estimación y pedir el mismo valor no se anotan; el 0 sí es un valor
func TestHistorias_Reestimar_SoloAnotaLosCambios(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyecto, _ := dosProyectos(t, pool)
	repo, h := historiaParaReestimar(t, pool, proyecto, 1)

	for _, puntos := range []int{0, 0, 5, 5, 0} { // primera (0), igual (0), 0 → 5, igual (5), 5 → 0
		if err := repo.ActualizarStoryPoints(ctx, h.ID, puntos); err != nil {
			t.Fatalf("%d story points: no se esperaba error: %v", puntos, err)
		}
	}

	historial, _ := repo.ListarEstimaciones(ctx, h.ID)
	if len(historial) != 2 || historial[0].Anterior != 5 || historial[0].Nuevo != 0 || historial[1].Anterior != 0 || historial[1].Nuevo != 5 {
		t.Errorf("se esperaban 5 → 0 y 0 → 5, y se obtuvo %+v", historial)
	}
}

// RN7 · el historial es de cada historia
func TestHistorias_Reestimar_ElHistorialEsDeCadaHistoria(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyecto, _ := dosProyectos(t, pool)
	repo, primera := historiaParaReestimar(t, pool, proyecto, 1)
	segunda, _ := repo.Guardar(ctx, historiaDeEjemplo(proyecto, 2, "Otra"))
	for _, puntos := range []int{5, 8} {
		_ = repo.ActualizarStoryPoints(ctx, primera.ID, puntos)
	}
	_ = repo.ActualizarStoryPoints(ctx, segunda.ID, 2)

	deLaPrimera, _ := repo.ListarEstimaciones(ctx, primera.ID)
	deLaSegunda, err := repo.ListarEstimaciones(ctx, segunda.ID)

	if len(deLaPrimera) != 1 {
		t.Errorf("la primera historia tenía que tener 1 entrada y tiene %d", len(deLaPrimera))
	}
	if err != nil || len(deLaSegunda) != 0 {
		t.Errorf("la segunda historia no tenía que tener entradas y tiene %+v, %v", deLaSegunda, err)
	}
	if historial, err := repo.ListarEstimaciones(ctx, 99999); err != nil || len(historial) != 0 {
		t.Errorf("una historia que no existe tiene el historial vacío, sin error, y se obtuvo %+v, %v", historial, err)
	}
}

// RN8 · todo o nada: si la historia no existe no se anota nada, y un valor que la base rechaza no deja una entrada suelta
func TestHistorias_Reestimar_LoQueFallaNoDejaHistorial(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyecto, _ := dosProyectos(t, pool)
	repo, h := historiaParaReestimar(t, pool, proyecto, 1)
	_ = repo.ActualizarStoryPoints(ctx, h.ID, 5)

	if err := repo.ActualizarStoryPoints(ctx, h.ID, 4); err == nil { // fuera de la escala: la base lo rechaza
		t.Fatal("la base aceptó 4 story points, que no están en la escala")
	}

	if historial, _ := repo.ListarEstimaciones(ctx, h.ID); len(historial) != 0 {
		t.Errorf("el intento fallido no tenía que anotar nada y quedó %+v", historial)
	}
	if h, _ := repo.BuscarPorNumero(ctx, proyecto, 1); h.StoryPoints == nil || *h.StoryPoints != 5 {
		t.Errorf("la historia tenía que seguir con 5 story points y quedó con %v", h.StoryPoints)
	}
}

// RN9 · al eliminar la historia se elimina su historial: una historia re-estimada se puede eliminar
func TestHistorias_Eliminar_ConHistorialDeEstimaciones(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyecto, _ := dosProyectos(t, pool)
	repo, primera := historiaParaReestimar(t, pool, proyecto, 1)
	segunda, _ := repo.Guardar(ctx, historiaDeEjemplo(proyecto, 2, "Otra"))
	for _, id := range []int64{primera.ID, segunda.ID} {
		_ = repo.ActualizarStoryPoints(ctx, id, 5)
		_ = repo.ActualizarStoryPoints(ctx, id, 8)
	}

	if err := repo.Eliminar(ctx, primera.ID); err != nil {
		t.Fatalf("una historia con historial de estimaciones se tiene que poder eliminar: %v", err)
	}

	var deLaEliminada, deLaOtra int
	_ = pool.QueryRow(ctx, `SELECT count(*) FROM estimaciones_historia WHERE historia_id = $1`, primera.ID).Scan(&deLaEliminada)
	_ = pool.QueryRow(ctx, `SELECT count(*) FROM estimaciones_historia WHERE historia_id = $1`, segunda.ID).Scan(&deLaOtra)
	if deLaEliminada != 0 || deLaOtra != 1 {
		t.Errorf("tenía que desaparecer el historial de la eliminada (quedan %d) y quedar el de la otra (hay %d)", deLaEliminada, deLaOtra)
	}
}

// RN8 · dos re-estimaciones a la vez no pierden ningún valor: cada entrada parte de donde terminó la anterior
func TestHistorias_Reestimar_ALaVez_NoPierdeValores(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyecto, _ := dosProyectos(t, pool)
	repo, h := historiaParaReestimar(t, pool, proyecto, 1)
	_ = repo.ActualizarStoryPoints(ctx, h.ID, 1)

	valores := []int{0, 2, 3, 5, 8, 13, 21} // distintos entre sí y del valor inicial: cada uno cambia la estimación
	var espera sync.WaitGroup
	for _, v := range valores {
		espera.Add(1)
		go func() {
			defer espera.Done()
			if err := repo.ActualizarStoryPoints(ctx, h.ID, v); err != nil {
				t.Errorf("%d story points: no se esperaba error: %v", v, err)
			}
		}()
	}
	espera.Wait()

	historial, _ := repo.ListarEstimaciones(ctx, h.ID) // la más reciente primero
	if len(historial) != len(valores) {
		t.Fatalf("se esperaban %d entradas y hay %d: %+v", len(valores), len(historial), historial)
	}
	for i := len(historial) - 1; i > 0; i-- { // de la más antigua a la más reciente
		if historial[i].Nuevo != historial[i-1].Anterior {
			t.Errorf("la entrada %d termina en %d y la siguiente parte de %d", i, historial[i].Nuevo, historial[i-1].Anterior)
		}
	}
	if historial[len(historial)-1].Anterior != 1 {
		t.Errorf("la primera entrada tenía que partir de la estimación inicial (1) y parte de %d", historial[len(historial)-1].Anterior)
	}
	if h, _ := repo.BuscarPorNumero(ctx, proyecto, 1); h.StoryPoints == nil || *h.StoryPoints != historial[0].Nuevo {
		t.Errorf("la historia tenía que quedar con el valor de la última entrada (%d) y quedó con %v", historial[0].Nuevo, h.StoryPoints)
	}
}
