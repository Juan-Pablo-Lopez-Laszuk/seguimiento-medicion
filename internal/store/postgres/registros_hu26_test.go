package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/postgres"
)

// HU-26 · CA-26.1 · los filtros van como parámetros en el WHERE, la fecha más reciente primero
func TestRegistros_ListarFiltrados(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	historiaA, historiaB, integrante := historiasConResponsable(t, pool)
	repo := postgres.NuevoRegistros(pool)
	guardar := func(historia int64, dia int, actividad string) effort.Registro {
		t.Helper()
		r, err := repo.Guardar(ctx, effort.Registro{IntegranteID: integrante, HistoriaID: historia, Fecha: fechaDeRegistro(dia), Actividad: actividad, Horas: 1})
		if err != nil {
			t.Fatalf("guardar el registro %s: %v", actividad, err)
		}
		return r
	}
	a, b, c := guardar(historiaA, 10, "A"), guardar(historiaA, 12, "B"), guardar(historiaB, 20, "C")

	casos := []struct {
		nombre    string
		historias []int64
		filtro    effort.Filtro
		esperados []effort.Registro
	}{
		{"sin filtros", []int64{historiaA, historiaB}, effort.Filtro{}, []effort.Registro{c, b, a}},
		{"por fechas del sprint", []int64{historiaA, historiaB}, effort.Filtro{Desde: fechaDeRegistro(11), Hasta: fechaDeRegistro(20)}, []effort.Registro{c, b}},
		{"por integrante e historia", []int64{historiaA, historiaB}, effort.Filtro{IntegranteID: integrante, HistoriaID: historiaB}, []effort.Registro{c}},
		{"otro integrante", []int64{historiaA, historiaB}, effort.Filtro{IntegranteID: integrante + 99}, nil},
		{"solo las historias pedidas", []int64{historiaA}, effort.Filtro{}, []effort.Registro{b, a}},
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

// HU-26 · CA-26.2 · se busca un registro por su ID y se guarda corregido, también con tarea
func TestRegistros_BuscarYActualizar(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	historia, _, integrante := historiasConResponsable(t, pool)
	tarea, err := postgres.NuevoTareas(pool).Guardar(ctx, effort.Tarea{HistoriaID: historia, Titulo: "T", ResponsableID: integrante, HorasEstimadas: 2})
	if err != nil {
		t.Fatalf("guardar la tarea: %v", err)
	}
	repo := postgres.NuevoRegistros(pool)
	a, err := repo.Guardar(ctx, effort.Registro{IntegranteID: integrante, HistoriaID: historia, TareaID: tarea.ID, Fecha: fechaDeRegistro(10), Actividad: "A", Horas: 3})
	if err != nil {
		t.Fatalf("guardar el registro: %v", err)
	}

	a.Horas, a.Actividad, a.TareaID, a.Fecha = 3.5, "A corregida", 0, fechaDeRegistro(11)
	if err := repo.Actualizar(ctx, a); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	guardado, err := repo.BuscarPorID(ctx, a.ID)
	if err != nil || guardado != a {
		t.Errorf("se esperaba %+v y se obtuvo %+v, %v", a, guardado, err)
	}

	if _, err := repo.BuscarPorID(ctx, a.ID+99); !errors.Is(err, effort.ErrRegistroNoEncontrado) {
		t.Errorf("buscar uno que no existe: error = %v, se esperaba ErrRegistroNoEncontrado", err)
	}
	if err := repo.Actualizar(ctx, effort.Registro{ID: a.ID + 99, Fecha: fechaDeRegistro(10), Horas: 1}); !errors.Is(err, effort.ErrRegistroNoEncontrado) {
		t.Errorf("actualizar uno que no existe: error = %v, se esperaba ErrRegistroNoEncontrado", err)
	}
	a.Horas = 25
	if err := repo.Actualizar(ctx, a); err == nil {
		t.Error("25 horas: se esperaba que la base lo rechace")
	}
}
