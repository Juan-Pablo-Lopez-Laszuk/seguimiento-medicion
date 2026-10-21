package memory_test

import (
	"context"
	"errors"
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

// HU-28 · un defecto se busca dentro de su proyecto y se actualizan su estado y su sprint de resolución
func TestDefectos_BuscarPorIDYActualizar(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoDefectos()
	d, _ := repo.Guardar(ctx, defect.Defecto{ProyectoID: 1, Descripcion: "A", Severidad: defect.SeveridadAlta, Estado: defect.EstadoAbierto, SprintDeteccionID: 1})

	if got, err := repo.BuscarPorID(ctx, 1, d.ID); err != nil || got != d {
		t.Errorf("se esperaba %+v y se obtuvo %+v, %v", d, got, err)
	}
	if _, err := repo.BuscarPorID(ctx, 2, d.ID); !errors.Is(err, defect.ErrDefectoNoEncontrado) {
		t.Errorf("de otro proyecto: se esperaba %v y se obtuvo %v", defect.ErrDefectoNoEncontrado, err)
	}
	d.Estado, d.SprintResolucionID = defect.EstadoResuelto, 2
	if err := repo.Actualizar(ctx, d); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if got, _ := repo.BuscarPorID(ctx, 1, d.ID); got != d {
		t.Errorf("quedó %+v, se esperaba %+v", got, d)
	}
	if err := repo.Actualizar(ctx, defect.Defecto{ID: 99, ProyectoID: 1}); !errors.Is(err, defect.ErrDefectoNoEncontrado) {
		t.Errorf("inexistente: se esperaba %v y se obtuvo %v", defect.ErrDefectoNoEncontrado, err)
	}
}

// HU-29 · los filtros por severidad, estado y sprint se combinan, sin mezclar proyectos
func TestDefectos_ListarFiltrados(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoDefectos()
	a, _ := repo.Guardar(ctx, defect.Defecto{ProyectoID: 1, Descripcion: "A", Severidad: defect.SeveridadAlta, Estado: defect.EstadoAbierto, SprintDeteccionID: 1})
	b, _ := repo.Guardar(ctx, defect.Defecto{ProyectoID: 1, Descripcion: "B", Severidad: defect.SeveridadAlta, Estado: defect.EstadoEnProgreso, SprintDeteccionID: 2})
	_, _ = repo.Guardar(ctx, defect.Defecto{ProyectoID: 2, Descripcion: "Otro", Severidad: defect.SeveridadAlta, Estado: defect.EstadoAbierto, SprintDeteccionID: 1})

	for _, c := range []struct {
		filtro defect.Filtro
		want   []defect.Defecto
	}{
		{defect.Filtro{}, []defect.Defecto{b, a}},
		{defect.Filtro{Severidad: defect.SeveridadAlta, SprintID: 1}, []defect.Defecto{a}},
		{defect.Filtro{Estado: defect.EstadoEnProgreso}, []defect.Defecto{b}},
		{defect.Filtro{Severidad: defect.SeveridadBaja}, nil},
	} {
		got, err := repo.ListarFiltrados(ctx, 1, c.filtro)
		if err != nil || len(got) != len(c.want) {
			t.Fatalf("ListarFiltrados(%+v) = %+v, %v; se esperaba %+v", c.filtro, got, err, c.want)
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("ListarFiltrados(%+v)[%d] = %+v, se esperaba %+v", c.filtro, i, got[i], c.want[i])
			}
		}
	}
}
