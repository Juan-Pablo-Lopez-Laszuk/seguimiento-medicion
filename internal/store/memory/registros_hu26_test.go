package memory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// HU-26 · CA-26.1 · solo los registros de las historias pedidas que cumplen el filtro, la fecha más reciente primero
func TestRegistros_ListarFiltrados(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoRegistros()
	a, _ := repo.Guardar(ctx, effort.Registro{IntegranteID: 2, HistoriaID: 1, Fecha: fechaRegistro(10), Actividad: "A", Horas: 3})
	b, _ := repo.Guardar(ctx, effort.Registro{IntegranteID: 3, HistoriaID: 1, Fecha: fechaRegistro(12), Actividad: "B", Horas: 2})
	c, _ := repo.Guardar(ctx, effort.Registro{IntegranteID: 2, HistoriaID: 2, Fecha: fechaRegistro(20), Actividad: "C", Horas: 4})
	_, _ = repo.Guardar(ctx, effort.Registro{IntegranteID: 2, HistoriaID: 9, Fecha: fechaRegistro(11), Actividad: "De otro proyecto", Horas: 1})

	casos := []struct {
		nombre    string
		historias []int64
		filtro    effort.Filtro
		esperados []effort.Registro
	}{
		{"sin filtros", []int64{1, 2}, effort.Filtro{}, []effort.Registro{c, b, a}},
		{"por integrante y fechas", []int64{1, 2}, effort.Filtro{IntegranteID: 2, Desde: fechaRegistro(5), Hasta: fechaRegistro(18)}, []effort.Registro{a}},
		{"por historia", []int64{1, 2}, effort.Filtro{HistoriaID: 1}, []effort.Registro{b, a}},
		{"ninguna historia", nil, effort.Filtro{}, nil},
	}
	for _, caso := range casos {
		t.Run(caso.nombre, func(t *testing.T) {
			lista, err := repo.ListarFiltrados(ctx, caso.historias, caso.filtro)
			if err != nil || len(lista) != len(caso.esperados) {
				t.Fatalf("se esperaban %d registros y se obtuvo %+v, %v", len(caso.esperados), lista, err)
			}
			for i := range lista {
				if lista[i] != caso.esperados[i] {
					t.Errorf("registro %d = %+v, se esperaba %+v", i, lista[i], caso.esperados[i])
				}
			}
		})
	}
}

// HU-26 · CA-26.2 · se busca un registro por su ID y se guarda corregido
func TestRegistros_BuscarYActualizar(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoRegistros()
	a, _ := repo.Guardar(ctx, effort.Registro{IntegranteID: 2, HistoriaID: 1, Fecha: fechaRegistro(10), Actividad: "A", Horas: 3})

	a.Horas, a.Actividad, a.TareaID, a.Fecha = 3.5, "A corregida", 5, fechaRegistro(11)
	if err := repo.Actualizar(ctx, a); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	guardado, err := repo.BuscarPorID(ctx, a.ID)
	if err != nil || guardado != a {
		t.Errorf("se esperaba %+v y se obtuvo %+v, %v", a, guardado, err)
	}

	if _, err := repo.BuscarPorID(ctx, 99); !errors.Is(err, effort.ErrRegistroNoEncontrado) {
		t.Errorf("buscar uno que no existe: error = %v, se esperaba ErrRegistroNoEncontrado", err)
	}
	if err := repo.Actualizar(ctx, effort.Registro{ID: 99}); !errors.Is(err, effort.ErrRegistroNoEncontrado) {
		t.Errorf("actualizar uno que no existe: error = %v, se esperaba ErrRegistroNoEncontrado", err)
	}
}
