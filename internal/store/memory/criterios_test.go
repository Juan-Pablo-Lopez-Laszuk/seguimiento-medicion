package memory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

func TestCriterios_GuardarAsignaIDsYListarPorHistoriaEnOrden(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoCriterios()

	a, err := repo.Guardar(ctx, backlog.Criterio{HistoriaID: 1, Descripcion: "Primero"})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	_, _ = repo.Guardar(ctx, backlog.Criterio{HistoriaID: 2, Descripcion: "De otra historia"})
	c, _ := repo.Guardar(ctx, backlog.Criterio{HistoriaID: 1, Descripcion: "Segundo"})

	if a.ID != 1 || c.ID != 3 {
		t.Errorf("IDs: se esperaba 1 y 3 y se obtuvo %d y %d", a.ID, c.ID)
	}
	lista, err := repo.ListarPorHistoria(ctx, 1)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if len(lista) != 2 || lista[0].Descripcion != "Primero" || lista[1].Descripcion != "Segundo" {
		t.Errorf("se esperaban solo los criterios de la historia 1, en orden, y se obtuvo %+v", lista)
	}
}

func TestCriterios_HistoriaSinCriterios_DevuelveListaVacia(t *testing.T) {
	lista, err := memory.NuevoCriterios().ListarPorHistoria(context.Background(), 1)
	if err != nil || len(lista) != 0 {
		t.Errorf("se esperaba una lista vacía sin error y se obtuvo %+v, %v", lista, err)
	}
}

// Para el progreso de toda la lista del backlog se piden los criterios de varias historias de una vez
func TestCriterios_ListarPorHistorias(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoCriterios()
	_, _ = repo.Guardar(ctx, backlog.Criterio{HistoriaID: 1, Descripcion: "De la 1"})
	_, _ = repo.Guardar(ctx, backlog.Criterio{HistoriaID: 2, Descripcion: "De la 2"})
	_, _ = repo.Guardar(ctx, backlog.Criterio{HistoriaID: 3, Descripcion: "De la 3"})

	lista, err := repo.ListarPorHistorias(ctx, []int64{1, 3})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if len(lista) != 2 || lista[0].HistoriaID != 1 || lista[1].HistoriaID != 3 {
		t.Errorf("se esperaban solo los criterios de las historias 1 y 3 y se obtuvo %+v", lista)
	}
	if vacia, err := repo.ListarPorHistorias(ctx, nil); err != nil || len(vacia) != 0 {
		t.Errorf("sin historias se esperaba una lista vacía sin error y se obtuvo %+v, %v", vacia, err)
	}
}

// RN8 · un criterio se busca dentro de su historia: con el id de otra historia es un "no encontrado"
func TestCriterios_BuscarPorID(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoCriterios()
	guardado, _ := repo.Guardar(ctx, backlog.Criterio{HistoriaID: 1, Descripcion: "Primero"})

	c, err := repo.BuscarPorID(ctx, 1, guardado.ID)
	if err != nil || c.Descripcion != "Primero" {
		t.Errorf("se esperaba el criterio y se obtuvo %+v, %v", c, err)
	}
	if _, err := repo.BuscarPorID(ctx, 2, guardado.ID); !errors.Is(err, backlog.ErrCriterioNoEncontrado) {
		t.Errorf("criterio de otra historia: se esperaba backlog.ErrCriterioNoEncontrado y se obtuvo %v", err)
	}
	if _, err := repo.BuscarPorID(ctx, 1, 99); !errors.Is(err, backlog.ErrCriterioNoEncontrado) {
		t.Errorf("id que no existe: se esperaba backlog.ErrCriterioNoEncontrado y se obtuvo %v", err)
	}
}

func TestCriterios_Actualizar(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoCriterios()
	guardado, _ := repo.Guardar(ctx, backlog.Criterio{HistoriaID: 1, Descripcion: "Primero"})

	guardado.Cumplido = true
	if err := repo.Actualizar(ctx, guardado); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	lista, _ := repo.ListarPorHistoria(ctx, 1)
	if len(lista) != 1 || !lista[0].Cumplido {
		t.Errorf("se esperaba el criterio cumplido y se obtuvo %+v", lista)
	}
	if err := repo.Actualizar(ctx, backlog.Criterio{ID: 99, HistoriaID: 1}); !errors.Is(err, backlog.ErrCriterioNoEncontrado) {
		t.Errorf("criterio que no existe: se esperaba backlog.ErrCriterioNoEncontrado y se obtuvo %v", err)
	}
}
