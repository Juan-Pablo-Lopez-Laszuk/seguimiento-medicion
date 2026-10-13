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

// HU-02 RN2 · el nombre repetido no cuenta al propio proyecto, sin distinguir mayúsculas
func TestProyectos_ExisteOtroConNombre(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoProyectos()
	propio, _ := repo.Guardar(ctx, project.Proyecto{Nombre: "Software Metrics"})
	_, _ = repo.Guardar(ctx, project.Proyecto{Nombre: "Otro proyecto"})

	casos := []struct {
		nombre   string
		esperado bool
	}{
		{"software METRICS", false}, // el propio, con otras mayúsculas
		{"OTRO PROYECTO", true},
		{"Nuevo nombre", false},
	}
	for _, c := range casos {
		existe, err := repo.ExisteOtroConNombre(ctx, c.nombre, propio.ID)
		if err != nil {
			t.Fatalf("no se esperaba error: %v", err)
		}
		if existe != c.esperado {
			t.Errorf("%q: se esperaba %v y se obtuvo %v", c.nombre, c.esperado, existe)
		}
	}
}

// HU-02 · Actualizar reemplaza los datos del proyecto con ese ID
func TestProyectos_Actualizar(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoProyectos()
	p, _ := repo.Guardar(ctx, project.Proyecto{Nombre: "Software Metrics"})
	otro, _ := repo.Guardar(ctx, project.Proyecto{Nombre: "Otro proyecto"})

	p.Nombre, p.Descripcion = "Software Metrics 2026", "TPI de ICS"
	if err := repo.Actualizar(ctx, p); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	guardado, _ := repo.BuscarPorID(ctx, p.ID)
	if guardado != p {
		t.Errorf("se esperaba %+v y se obtuvo %+v", p, guardado)
	}
	if sinTocar, _ := repo.BuscarPorID(ctx, otro.ID); sinTocar != otro {
		t.Errorf("el otro proyecto no tenía que cambiar: %+v", sinTocar)
	}
	if err := repo.Actualizar(ctx, project.Proyecto{ID: 99}); !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("con un ID que no existe se esperaba project.ErrNoEncontrado y se obtuvo %v", err)
	}
}
