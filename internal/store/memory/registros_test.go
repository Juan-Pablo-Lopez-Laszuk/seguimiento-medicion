package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

func fechaRegistro(d int) time.Time { return time.Date(2026, 10, d, 0, 0, 0, 0, time.UTC) }

// HU-24 · se listan los registros de una historia (fecha más reciente primero) y los de un integrante
func TestRegistros_GuardarYListar(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoRegistros()
	a, err := repo.Guardar(ctx, effort.Registro{IntegranteID: 2, HistoriaID: 1, Fecha: fechaRegistro(16), Actividad: "A", Horas: 3})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	b, _ := repo.Guardar(ctx, effort.Registro{IntegranteID: 3, HistoriaID: 1, TareaID: 5, Fecha: fechaRegistro(17), Actividad: "B", Horas: 2})
	c, _ := repo.Guardar(ctx, effort.Registro{IntegranteID: 2, HistoriaID: 2, Fecha: fechaRegistro(17), Actividad: "C", Horas: 1})

	if a.ID != 1 || c.ID != 3 {
		t.Errorf("IDs: se esperaba 1 y 3 y se obtuvo %d y %d", a.ID, c.ID)
	}
	deLaHistoria, err := repo.ListarPorHistoria(ctx, 1)
	if err != nil || len(deLaHistoria) != 2 || deLaHistoria[0] != b || deLaHistoria[1] != a {
		t.Errorf("de la historia 1 se esperaba [b, a] y se obtuvo %+v, %v", deLaHistoria, err)
	}
	delIntegrante, err := repo.ListarPorIntegrante(ctx, 2)
	if err != nil || len(delIntegrante) != 2 {
		t.Errorf("de la integrante 2 se esperaban 2 registros y se obtuvo %+v, %v", delIntegrante, err)
	}
}
