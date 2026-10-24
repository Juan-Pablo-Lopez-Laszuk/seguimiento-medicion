package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/postgres"
)

// ordenesDe devuelve el orden de cada historia del proyecto, por número.
func ordenesDe(t *testing.T, repo *postgres.Historias, proyectoID int64) map[int]int {
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

// guardarConOrden guarda la HU-numero del proyecto con ese orden.
func guardarConOrden(t *testing.T, repo *postgres.Historias, proyectoID int64, numero, orden int) backlog.Historia {
	t.Helper()
	h := historiaDeEjemplo(proyectoID, numero, "Historia")
	h.Orden = orden
	guardada, err := repo.Guardar(context.Background(), h)
	if err != nil {
		t.Fatalf("guardar la HU-%d: %v", numero, err)
	}
	return guardada
}

// HU-09 · Guardar, ListarPorProyecto y BuscarPorNumero conservan el orden de la historia
func TestHistorias_GuardanYDevuelvenElOrden(t *testing.T) {
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	repo := postgres.NuevoHistorias(pool)
	guardarConOrden(t, repo, proyectoA, 1, 3)

	if got := ordenesDe(t, repo, proyectoA)[1]; got != 3 {
		t.Errorf("orden listado = %d, se esperaba 3", got)
	}
	h, err := repo.BuscarPorNumero(context.Background(), proyectoA, 1)
	if err != nil || h.Orden != 3 {
		t.Errorf("BuscarPorNumero: orden = %d (%v), se esperaba 3", h.Orden, err)
	}
}

// HU-09 RN5 · ActualizarOrden guarda el orden de varias historias del proyecto juntas y no toca las demás
func TestHistorias_ActualizarOrden(t *testing.T) {
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	repo := postgres.NuevoHistorias(pool)
	a, b := guardarConOrden(t, repo, proyectoA, 1, 0), guardarConOrden(t, repo, proyectoA, 2, 0)
	guardarConOrden(t, repo, proyectoA, 3, 5)

	if err := repo.ActualizarOrden(context.Background(), proyectoA, map[int64]int{a.ID: 2, b.ID: 1}); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if got, want := ordenesDe(t, repo, proyectoA), map[int]int{1: 2, 2: 1, 3: 5}; got[1] != want[1] || got[2] != want[2] || got[3] != want[3] {
		t.Errorf("órdenes = %v, se esperaba %v", got, want)
	}
}

// HU-09 RN5 · si una historia no es del proyecto no se guarda ningún orden: todo va en una transacción
func TestHistorias_ActualizarOrden_ConUnaHistoriaDeOtroProyecto_NoGuardaNada(t *testing.T) {
	pool := baseMigrada(t)
	proyectoA, proyectoB := dosProyectos(t, pool)
	repo := postgres.NuevoHistorias(pool)
	a := guardarConOrden(t, repo, proyectoA, 1, 0)
	otra := guardarConOrden(t, repo, proyectoB, 1, 0)

	err := repo.ActualizarOrden(context.Background(), proyectoA, map[int64]int{a.ID: 2, otra.ID: 1})

	if !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("error = %v, se esperaba ErrHistoriaNoEncontrada", err)
	}
	if ordenesDe(t, repo, proyectoA)[1] != 0 || ordenesDe(t, repo, proyectoB)[1] != 0 {
		t.Errorf("no se esperaba ningún orden guardado")
	}
}

// HU-09 RN7 · cambiar la prioridad (HU-06) conserva el orden
func TestHistorias_ActualizarDatos_ConservaElOrden(t *testing.T) {
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	repo := postgres.NuevoHistorias(pool)
	h := guardarConOrden(t, repo, proyectoA, 1, 4)

	h.Prioridad, h.Orden = backlog.PrioridadAlta, 0
	if err := repo.ActualizarDatos(context.Background(), h); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if got := ordenesDe(t, repo, proyectoA)[1]; got != 4 {
		t.Errorf("orden = %d, se esperaba que siga en 4", got)
	}
}
