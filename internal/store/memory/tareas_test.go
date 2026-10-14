package memory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// RN4 y RN9 · cada historia lista solo sus tareas, en el orden en que se cargaron
func TestTareas_GuardarAsignaIDsYListarPorHistoriaEnOrden(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoTareas()

	a, err := repo.Guardar(ctx, effort.Tarea{HistoriaID: 1, Titulo: "Primera", ResponsableID: 2, HorasEstimadas: 3})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	_, _ = repo.Guardar(ctx, effort.Tarea{HistoriaID: 2, Titulo: "De otra historia", ResponsableID: 2, HorasEstimadas: 1})
	c, _ := repo.Guardar(ctx, effort.Tarea{HistoriaID: 1, Titulo: "Segunda", ResponsableID: 3, HorasEstimadas: 2.5})

	if a.ID != 1 || c.ID != 3 {
		t.Errorf("IDs: se esperaba 1 y 3 y se obtuvo %d y %d", a.ID, c.ID)
	}
	lista, err := repo.ListarPorHistoria(ctx, 1)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if len(lista) != 2 || lista[0].Titulo != "Primera" || lista[1].Titulo != "Segunda" {
		t.Errorf("se esperaban solo las tareas de la historia 1, en orden, y se obtuvo %+v", lista)
	}
	if lista[1] != c {
		t.Errorf("la tarea guardada = %+v, se esperaba %+v con todos sus campos", lista[1], c)
	}
}

func TestTareas_HistoriaSinTareas_DevuelveListaVacia(t *testing.T) {
	lista, err := memory.NuevoTareas().ListarPorHistoria(context.Background(), 1)
	if err != nil || len(lista) != 0 {
		t.Errorf("se esperaba una lista vacía sin error y se obtuvo %+v, %v", lista, err)
	}
}

// CA-25.2 · una tarea se busca dentro de su historia: con el id de otra historia es un "no encontrado"
func TestTareas_BuscarPorID(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoTareas()
	guardada, _ := repo.Guardar(ctx, effort.Tarea{HistoriaID: 1, Titulo: "Primera", ResponsableID: 2, HorasEstimadas: 3})

	tarea, err := repo.BuscarPorID(ctx, 1, guardada.ID)
	if err != nil || tarea != guardada {
		t.Errorf("se esperaba la tarea y se obtuvo %+v, %v", tarea, err)
	}
	if _, err := repo.BuscarPorID(ctx, 2, guardada.ID); !errors.Is(err, effort.ErrTareaNoEncontrada) {
		t.Errorf("tarea de otra historia: se esperaba effort.ErrTareaNoEncontrada y se obtuvo %v", err)
	}
	if _, err := repo.BuscarPorID(ctx, 1, 99); !errors.Is(err, effort.ErrTareaNoEncontrada) {
		t.Errorf("id que no existe: se esperaba effort.ErrTareaNoEncontrada y se obtuvo %v", err)
	}
}

// Actualizar guarda las horas y si está completada, sin poder moverla a otra historia (CA-25.2).
func TestTareas_Actualizar(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoTareas()
	guardada, _ := repo.Guardar(ctx, effort.Tarea{HistoriaID: 1, Titulo: "Primera", ResponsableID: 2, HorasEstimadas: 3})

	guardada.HorasEstimadas, guardada.Completada = 4, true
	if err := repo.Actualizar(ctx, guardada); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	lista, _ := repo.ListarPorHistoria(ctx, 1)
	if len(lista) != 1 || lista[0] != guardada {
		t.Errorf("se esperaba la tarea con 4 horas y completada, y se obtuvo %+v", lista)
	}
	otraHistoria := guardada
	otraHistoria.HistoriaID = 2
	if err := repo.Actualizar(ctx, otraHistoria); !errors.Is(err, effort.ErrTareaNoEncontrada) {
		t.Errorf("con otra historia: se esperaba effort.ErrTareaNoEncontrada y se obtuvo %v", err)
	}
	if err := repo.Actualizar(ctx, effort.Tarea{ID: 99, HistoriaID: 1}); !errors.Is(err, effort.ErrTareaNoEncontrada) {
		t.Errorf("tarea que no existe: se esperaba effort.ErrTareaNoEncontrada y se obtuvo %v", err)
	}
}
