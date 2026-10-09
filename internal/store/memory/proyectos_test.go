package memory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

func TestProyectos_GuardarAsignaIDsCorrelativosYListar(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoProyectos()

	a, err := repo.Guardar(ctx, project.Proyecto{Nombre: "Primero"})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	b, _ := repo.Guardar(ctx, project.Proyecto{Nombre: "Segundo"})

	if a.ID != 1 || b.ID != 2 {
		t.Errorf("IDs: se esperaba 1 y 2 y se obtuvo %d y %d", a.ID, b.ID)
	}
	lista, err := repo.Listar(ctx)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if len(lista) != 2 || lista[0].Nombre != "Primero" || lista[1].Nombre != "Segundo" {
		t.Errorf("se esperaba [Primero Segundo] y se obtuvo %+v", lista)
	}
}

// RN3 · el nombre se compara sin distinguir mayúsculas de minúsculas (también con tildes y ñ)
func TestProyectos_ExisteNombreSinDistinguirMayusculas(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoProyectos()
	_, _ = repo.Guardar(ctx, project.Proyecto{Nombre: "Gestión Ñandú"})

	casos := map[string]bool{
		"Gestión Ñandú": true,
		"GESTIÓN ÑANDÚ": true,
		"gestión ñandú": true,
		"Gestion Nandu": false, // sin tilde ni ñ es otro nombre
		"Otro":          false,
	}
	for nombre, esperado := range casos {
		existe, err := repo.ExisteNombre(ctx, nombre)
		if err != nil {
			t.Fatalf("no se esperaba error: %v", err)
		}
		if existe != esperado {
			t.Errorf("ExisteNombre(%q): se esperaba %v y se obtuvo %v", nombre, esperado, existe)
		}
	}
}

// HU-11 · para crear un sprint hay que buscar su proyecto
func TestProyectos_BuscarPorID(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoProyectos()
	guardado, _ := repo.Guardar(ctx, project.Proyecto{Nombre: "Software Metrics"})

	p, err := repo.BuscarPorID(ctx, guardado.ID)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if p.Nombre != "Software Metrics" {
		t.Errorf("se esperaba el proyecto %q y se obtuvo %+v", "Software Metrics", p)
	}

	_, err = repo.BuscarPorID(ctx, 99)
	if !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("id inexistente: se esperaba project.ErrNoEncontrado y se obtuvo %v", err)
	}
}
