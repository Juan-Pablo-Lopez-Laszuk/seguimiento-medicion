package memory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

func TestIntegrantes_GuardarAsignaIDsYListarPorProyecto(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoIntegrantes()

	a, err := repo.Guardar(ctx, member.Integrante{ProyectoID: 1, Nombre: "Ana"})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	_, _ = repo.Guardar(ctx, member.Integrante{ProyectoID: 2, Nombre: "De otro proyecto"})
	c, _ := repo.Guardar(ctx, member.Integrante{ProyectoID: 1, Nombre: "Bea"})

	if a.ID != 1 || c.ID != 3 {
		t.Errorf("IDs: se esperaba 1 y 3 y se obtuvo %d y %d", a.ID, c.ID)
	}
	lista, err := repo.ListarPorProyecto(ctx, 1)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if len(lista) != 2 || lista[0].Nombre != "Ana" || lista[1].Nombre != "Bea" {
		t.Errorf("se esperaban solo los integrantes del proyecto 1, en orden, y se obtuvo %+v", lista)
	}
}

// Un integrante se busca dentro de su proyecto: con el ID de otro proyecto no se encuentra.
func TestIntegrantes_BuscarPorID(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoIntegrantes()
	guardado, _ := repo.Guardar(ctx, member.Integrante{ProyectoID: 1, Nombre: "Ana"})

	i, err := repo.BuscarPorID(ctx, 1, guardado.ID)
	if err != nil || i.Nombre != "Ana" {
		t.Errorf("se esperaba a Ana sin error y se obtuvo %+v, %v", i, err)
	}
	for nombre, ids := range map[string][2]int64{"ID inexistente": {1, 99}, "otro proyecto": {2, guardado.ID}} {
		if _, err := repo.BuscarPorID(ctx, ids[0], ids[1]); !errors.Is(err, member.ErrNoEncontrado) {
			t.Errorf("%s: se esperaba member.ErrNoEncontrado y se obtuvo %v", nombre, err)
		}
	}
}

// Actualizar reemplaza los datos guardados (la usa la baja).
func TestIntegrantes_Actualizar(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoIntegrantes()
	guardado, _ := repo.Guardar(ctx, member.Integrante{ProyectoID: 1, Nombre: "Ana", Activo: true})

	guardado.Activo = false
	if err := repo.Actualizar(ctx, guardado); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if i, _ := repo.BuscarPorID(ctx, 1, guardado.ID); i.Activo {
		t.Error("se esperaba que quede inactivo")
	}
	if err := repo.Actualizar(ctx, member.Integrante{ID: 99, ProyectoID: 1}); !errors.Is(err, member.ErrNoEncontrado) {
		t.Errorf("ID inexistente: se esperaba member.ErrNoEncontrado y se obtuvo %v", err)
	}
}

// La lista es una copia: cambiarla no cambia lo guardado.
func TestIntegrantes_ListarDevuelveUnaCopia(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoIntegrantes()
	_, _ = repo.Guardar(ctx, member.Integrante{ProyectoID: 1, Nombre: "Ana"})

	lista, _ := repo.ListarPorProyecto(ctx, 1)
	lista[0].Nombre = "Cambiado"

	if otra, _ := repo.ListarPorProyecto(ctx, 1); otra[0].Nombre != "Ana" {
		t.Errorf("se esperaba que lo guardado siga siendo Ana y es %q", otra[0].Nombre)
	}
}
