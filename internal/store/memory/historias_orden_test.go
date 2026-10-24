package memory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// ordenesDe devuelve el orden de cada historia del proyecto, por número.
func ordenesDe(t *testing.T, repo *memory.Historias, proyectoID int64) map[int]int {
	t.Helper()
	lista, err := repo.ListarPorProyecto(context.Background(), proyectoID)
	if err != nil {
		t.Fatalf("listar: %v", err)
	}
	ordenes := map[int]int{}
	for _, h := range lista {
		ordenes[h.Numero] = h.Orden
	}
	return ordenes
}

// HU-09 · Guardar, ListarPorProyecto y BuscarPorNumero conservan el orden de la historia
func TestHistorias_GuardanYDevuelvenElOrden(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoHistorias()
	if _, err := repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 1, Titulo: "Login", Orden: 3}); err != nil {
		t.Fatalf("guardar: %v", err)
	}

	if got := ordenesDe(t, repo, 1)[1]; got != 3 {
		t.Errorf("orden listado = %d, se esperaba 3", got)
	}
	h, err := repo.BuscarPorNumero(ctx, 1, 1)
	if err != nil || h.Orden != 3 {
		t.Errorf("BuscarPorNumero: orden = %d (%v), se esperaba 3", h.Orden, err)
	}
}

// HU-09 RN5 · ActualizarOrden guarda el orden de varias historias del proyecto juntas y no toca las demás
func TestHistorias_ActualizarOrden(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoHistorias()
	a, _ := repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 1, Titulo: "A"})
	b, _ := repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 2, Titulo: "B"})
	_, _ = repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 3, Titulo: "C", Orden: 5})

	if err := repo.ActualizarOrden(ctx, 1, map[int64]int{a.ID: 2, b.ID: 1}); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if got, want := ordenesDe(t, repo, 1), map[int]int{1: 2, 2: 1, 3: 5}; got[1] != want[1] || got[2] != want[2] || got[3] != want[3] {
		t.Errorf("órdenes = %v, se esperaba %v", got, want)
	}
}

// HU-09 RN5 · si una historia no es del proyecto no se guarda ningún orden
func TestHistorias_ActualizarOrden_ConUnaHistoriaDeOtroProyecto_NoGuardaNada(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoHistorias()
	a, _ := repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 1, Titulo: "A"})
	otra, _ := repo.Guardar(ctx, backlog.Historia{ProyectoID: 2, Numero: 1, Titulo: "De otro proyecto"})

	err := repo.ActualizarOrden(ctx, 1, map[int64]int{a.ID: 2, otra.ID: 1})

	if !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("error = %v, se esperaba ErrHistoriaNoEncontrada", err)
	}
	if ordenesDe(t, repo, 1)[1] != 0 || ordenesDe(t, repo, 2)[1] != 0 {
		t.Errorf("no se esperaba ningún orden guardado")
	}
}

// HU-09 RN7 · cambiar la prioridad (HU-06) conserva el orden
func TestHistorias_ActualizarDatos_ConservaElOrden(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoHistorias()
	h, _ := repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 1, Titulo: "A", Prioridad: backlog.PrioridadBaja, Orden: 4})

	h.Prioridad, h.Orden = backlog.PrioridadAlta, 0
	if err := repo.ActualizarDatos(ctx, h); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if got := ordenesDe(t, repo, 1)[1]; got != 4 {
		t.Errorf("orden = %d, se esperaba que siga en 4", got)
	}
}
