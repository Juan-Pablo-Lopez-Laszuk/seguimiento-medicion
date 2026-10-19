package memory_test

import (
	"context"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/defect"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// HU-27 · cada proyecto lista solo sus defectos, el más nuevo primero
func TestDefectos_GuardarYListarPorProyecto(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoDefectos()
	a, err := repo.Guardar(ctx, defect.Defecto{ProyectoID: 1, Descripcion: "A", Severidad: defect.SeveridadAlta, Estado: defect.EstadoAbierto, SprintDeteccionID: 1})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	_, _ = repo.Guardar(ctx, defect.Defecto{ProyectoID: 2, Descripcion: "De otro proyecto"})
	c, _ := repo.Guardar(ctx, defect.Defecto{ProyectoID: 1, HistoriaID: 4, Descripcion: "C", Severidad: defect.SeveridadBaja, Estado: defect.EstadoAbierto, SprintDeteccionID: 1})

	lista, err := repo.ListarPorProyecto(ctx, 1)
	if err != nil || len(lista) != 2 || lista[0] != c || lista[1] != a || a.ID != 1 {
		t.Errorf("se esperaba [c, a] del proyecto 1 y se obtuvo %+v, %v", lista, err)
	}
}
