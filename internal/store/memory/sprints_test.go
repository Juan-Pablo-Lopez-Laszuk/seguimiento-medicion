package memory_test

import (
	"context"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

func TestSprints_GuardarAsignaIDsYListarPorProyecto(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoSprints()

	a, err := repo.Guardar(ctx, sprint.Sprint{ProyectoID: 1, Numero: 1, Objetivo: "Primero"})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	_, _ = repo.Guardar(ctx, sprint.Sprint{ProyectoID: 2, Numero: 1, Objetivo: "De otro proyecto"})
	c, _ := repo.Guardar(ctx, sprint.Sprint{ProyectoID: 1, Numero: 2, Objetivo: "Segundo"})

	if a.ID != 1 || c.ID != 3 {
		t.Errorf("IDs: se esperaba 1 y 3 y se obtuvo %d y %d", a.ID, c.ID)
	}
	lista, err := repo.ListarPorProyecto(ctx, 1)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if len(lista) != 2 || lista[0].Objetivo != "Primero" || lista[1].Objetivo != "Segundo" {
		t.Errorf("se esperaban solo los sprints del proyecto 1, en orden, y se obtuvo %+v", lista)
	}
}

func TestSprints_ProyectoSinSprints_DevuelveListaVacia(t *testing.T) {
	lista, err := memory.NuevoSprints().ListarPorProyecto(context.Background(), 1)
	if err != nil || len(lista) != 0 {
		t.Errorf("se esperaba una lista vacía sin error y se obtuvo %+v, %v", lista, err)
	}
}
