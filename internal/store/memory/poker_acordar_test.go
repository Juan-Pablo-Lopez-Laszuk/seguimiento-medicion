package memory_test

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// pokerConHistoria arma los repositorios de Planning Poker y de historias conectados, guarda la historia Login con los
// story points que se pidan (nil: sin estimar) y le abre una sesión.
func pokerConHistoria(t *testing.T, puntos *int) (*memory.Poker, *memory.Historias, backlog.Historia, poker.Sesion) {
	t.Helper()
	ctx := context.Background()
	historias := memory.NuevoHistorias()
	h, err := historias.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 1, Titulo: "Login", Prioridad: backlog.PrioridadAlta, Estado: backlog.EstadoPendiente})
	if err != nil {
		t.Fatalf("guardar la historia: %v", err)
	}
	if puntos != nil {
		if err := historias.ActualizarStoryPoints(ctx, h.ID, *puntos); err != nil {
			t.Fatalf("estimar la historia: %v", err)
		}
	}
	repo := memory.NuevoPoker().ConHistorias(historias)
	sesion, _, err := repo.Abrir(ctx, h.ID)
	if err != nil {
		t.Fatalf("abrir la sesión: %v", err)
	}
	return repo, historias, h, sesion
}

func puntosEnMemoria(t *testing.T, historias *memory.Historias, h backlog.Historia) *int {
	t.Helper()
	guardada, err := historias.BuscarPorNumero(context.Background(), h.ProyectoID, h.Numero)
	if err != nil {
		t.Fatalf("buscar la historia: %v", err)
	}
	return guardada.StoryPoints
}

// CA-22.2 CA-22.3 · cerrar deja la sesión Cerrada con su valor y estima la historia; la primera estimación no se anota
func TestPoker_Cerrar_CierraLaSesionYEstimaLaHistoria(t *testing.T) {
	ctx := context.Background()
	repo, historias, h, sesion := pokerConHistoria(t, nil)

	if err := repo.Cerrar(ctx, sesion.ID, h.ID, 5); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if _, _, err := repo.SesionAbierta(ctx, h.ID); !errors.Is(err, poker.ErrSesionNoEncontrada) {
		t.Errorf("SesionAbierta: error = %v, se esperaba ErrSesionNoEncontrada", err)
	}
	if sp := puntosEnMemoria(t, historias, h); sp == nil || *sp != 5 {
		t.Errorf("story points = %v, se esperaba 5", sp)
	}
	if historial, _ := historias.ListarEstimaciones(ctx, h.ID); len(historial) != 0 {
		t.Errorf("la primera estimación no se anota y el historial es %+v", historial)
	}
}

// CA-22.2 RN2 · re-estimar con Planning Poker anota el cambio en el historial de HU-16b
func TestPoker_Cerrar_AnotaLaReestimacion(t *testing.T) {
	ctx := context.Background()
	tres := 3
	repo, historias, h, sesion := pokerConHistoria(t, &tres)

	if err := repo.Cerrar(ctx, sesion.ID, h.ID, 8); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	historial, err := historias.ListarEstimaciones(ctx, h.ID)
	if err != nil || len(historial) != 1 || historial[0].Anterior != 3 || historial[0].Nuevo != 8 {
		t.Errorf("historial = %+v, %v; se esperaba el cambio de 3 a 8", historial, err)
	}
}

// RN5 · una sesión cerrada no se vuelve a cerrar
func TestPoker_Cerrar_SesionYaCerrada(t *testing.T) {
	ctx := context.Background()
	repo, historias, h, sesion := pokerConHistoria(t, nil)
	if err := repo.Cerrar(ctx, sesion.ID, h.ID, 5); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if err := repo.Cerrar(ctx, sesion.ID, h.ID, 8); !errors.Is(err, poker.ErrSesionCerrada) {
		t.Errorf("error = %v, se esperaba ErrSesionCerrada", err)
	}
	if sp := puntosEnMemoria(t, historias, h); *sp != 5 {
		t.Errorf("story points = %d, se esperaba 5", *sp)
	}
}

// RN5 · si dos pedidos cierran la misma sesión a la vez, se registra uno solo
func TestPoker_CerrarDosVecesALaVez(t *testing.T) {
	ctx := context.Background()
	repo, historias, h, sesion := pokerConHistoria(t, nil)

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

	ganador := slices.Index(errs, nil)
	if ganador == -1 || !errors.Is(errs[1-ganador], poker.ErrSesionCerrada) {
		t.Fatalf("se esperaba un solo cierre y se obtuvo %v", errs)
	}
	if sp := puntosEnMemoria(t, historias, h); *sp != valores[ganador] {
		t.Errorf("story points = %d, se esperaba %d", *sp, valores[ganador])
	}
}

// RN6 · todo o nada: si no se pueden guardar los story points, la sesión sigue abierta
func TestPoker_Cerrar_LoQueFallaNoCierraLaSesion(t *testing.T) {
	ctx := context.Background()
	repo, _, h, sesion := pokerConHistoria(t, nil)

	if err := repo.Cerrar(ctx, sesion.ID, h.ID+1000, 5); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Fatalf("error = %v, se esperaba ErrHistoriaNoEncontrada", err)
	}
	abierta, _, err := repo.SesionAbierta(ctx, h.ID)
	if err != nil || abierta.ID != sesion.ID || abierta.ValorAcordado != nil {
		t.Errorf("SesionAbierta = %+v, %v; la sesión tenía que seguir abierta y sin valor", abierta, err)
	}
}

// CA-22.3 · las sesiones cerradas se consultan, la más reciente primero, con su número, sus rondas y su valor
func TestPoker_SesionesCerradas(t *testing.T) {
	ctx := context.Background()
	repo, historias, h, primera := pokerConHistoria(t, nil)
	otra, _ := historias.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 2, Titulo: "Registro", Prioridad: backlog.PrioridadAlta, Estado: backlog.EstadoPendiente})

	if cerradas, err := repo.SesionesCerradas(ctx, h.ID); err != nil || len(cerradas) != 0 {
		t.Errorf("sin cerrar: %+v, %v; se esperaba la lista vacía", cerradas, err)
	}
	if err := repo.Cerrar(ctx, primera.ID, h.ID, 5); err != nil {
		t.Fatal(err)
	}
	ajena, _, _ := repo.Abrir(ctx, otra.ID)
	if err := repo.Cerrar(ctx, ajena.ID, otra.ID, 13); err != nil {
		t.Fatal(err)
	}
	segunda, _, _ := repo.Abrir(ctx, h.ID)
	if err := repo.Cerrar(ctx, segunda.ID, h.ID, 8); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repo.Abrir(ctx, h.ID); err != nil { // la tercera queda abierta: no se lista
		t.Fatal(err)
	}

	cerradas, err := repo.SesionesCerradas(ctx, h.ID)
	quiero := []poker.SesionCerrada{{Numero: 2, Rondas: 1, ValorAcordado: 8}, {Numero: 1, Rondas: 1, ValorAcordado: 5}}
	if err != nil || !slices.Equal(cerradas, quiero) {
		t.Errorf("SesionesCerradas = %+v, %v; se esperaba %+v", cerradas, err, quiero)
	}
}
