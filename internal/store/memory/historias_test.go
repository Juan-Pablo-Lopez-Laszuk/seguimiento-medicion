package memory_test

import (
	"context"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

func TestHistorias_GuardarAsignaIDsYListarPorProyecto(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoHistorias()

	a, err := repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 1, Titulo: "Primera"})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	_, _ = repo.Guardar(ctx, backlog.Historia{ProyectoID: 2, Numero: 1, Titulo: "De otro proyecto"})
	c, _ := repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 2, Titulo: "Segunda"})

	if a.ID != 1 || c.ID != 3 {
		t.Errorf("IDs: se esperaba 1 y 3 y se obtuvo %d y %d", a.ID, c.ID)
	}
	lista, err := repo.ListarPorProyecto(ctx, 1)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if len(lista) != 2 || lista[0].Titulo != "Primera" || lista[1].Titulo != "Segunda" {
		t.Errorf("se esperaban solo las historias del proyecto 1, en orden, y se obtuvo %+v", lista)
	}
}

func TestHistorias_ProyectoSinHistorias_DevuelveListaVacia(t *testing.T) {
	lista, err := memory.NuevoHistorias().ListarPorProyecto(context.Background(), 1)
	if err != nil || len(lista) != 0 {
		t.Errorf("se esperaba una lista vacía sin error y se obtuvo %+v, %v", lista, err)
	}
}

// Los story points son un puntero: lo que guarda el repositorio no tiene que cambiar si quien lo usa modifica
// el valor que recibió o que le devolvieron.
func TestHistorias_LosStoryPointsGuardadosNoSeComparten(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoHistorias()
	puntos := 5
	guardada, _ := repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 1, StoryPoints: &puntos})

	puntos = 8
	*guardada.StoryPoints = 13
	lista, _ := repo.ListarPorProyecto(ctx, 1)
	*lista[0].StoryPoints = 21
	otraVez, _ := repo.ListarPorProyecto(ctx, 1)

	if otraVez[0].StoryPoints == nil || *otraVez[0].StoryPoints != 5 {
		t.Errorf("se esperaban 5 story points guardados y se obtuvo %v", otraVez[0].StoryPoints)
	}
}
