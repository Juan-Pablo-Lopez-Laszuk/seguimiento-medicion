package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/postgres"
)

func fecha(anio int, mes time.Month, dia int) time.Time {
	return time.Date(anio, mes, dia, 0, 0, 0, 0, time.UTC)
}

func proyectoDeEjemplo(nombre string) project.Proyecto {
	return project.Proyecto{
		Nombre:      nombre,
		Descripcion: "Seguimiento de métricas",
		FechaInicio: fecha(2026, time.September, 28),
		FechaFin:    fecha(2026, time.November, 1),
		Estado:      project.EstadoPlanificado,
		CreadoEn:    time.Date(2026, time.October, 9, 15, 30, 0, 0, time.UTC),
	}
}

func TestProyectos_GuardarAsignaIDsCorrelativosYListar(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NuevoProyectos(baseMigrada(t))

	a, err := repo.Guardar(ctx, proyectoDeEjemplo("Primero"))
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	b, err := repo.Guardar(ctx, proyectoDeEjemplo("Segundo"))
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if a.ID == 0 || b.ID != a.ID+1 {
		t.Errorf("IDs: se esperaban dos correlativos y se obtuvo %d y %d", a.ID, b.ID)
	}
	lista, err := repo.Listar(ctx)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if len(lista) != 2 || lista[0].Nombre != "Primero" || lista[1].Nombre != "Segundo" {
		t.Errorf("se esperaba [Primero Segundo] y se obtuvo %+v", lista)
	}
}

func TestProyectos_GuardaYDevuelveTodosLosCampos(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NuevoProyectos(baseMigrada(t))
	original := proyectoDeEjemplo("Gestión Ñandú")

	guardado, err := repo.Guardar(ctx, original)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	leido, err := repo.BuscarPorID(ctx, guardado.ID)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if leido.ID != guardado.ID || leido.Nombre != original.Nombre || leido.Descripcion != original.Descripcion || leido.Estado != original.Estado {
		t.Errorf("campos de texto distintos: se guardó %+v y se leyó %+v", original, leido)
	}
	// Las fechas son sin hora y a medianoche UTC, como las maneja el dominio (SoloFecha).
	if !leido.FechaInicio.Equal(original.FechaInicio) || !leido.FechaFin.Equal(original.FechaFin) {
		t.Errorf("fechas distintas: se guardó %v-%v y se leyó %v-%v", original.FechaInicio, original.FechaFin, leido.FechaInicio, leido.FechaFin)
	}
	if !leido.CreadoEn.Equal(original.CreadoEn) {
		t.Errorf("creado_en: se guardó %v y se leyó %v", original.CreadoEn, leido.CreadoEn)
	}
}

// RN3 · el nombre se compara sin distinguir mayúsculas de minúsculas (también con tildes y ñ)
func TestProyectos_ExisteNombreSinDistinguirMayusculas(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NuevoProyectos(baseMigrada(t))
	if _, err := repo.Guardar(ctx, proyectoDeEjemplo("Gestión Ñandú")); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

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
	repo := postgres.NuevoProyectos(baseMigrada(t))
	guardado, _ := repo.Guardar(ctx, proyectoDeEjemplo("Software Metrics"))

	p, err := repo.BuscarPorID(ctx, guardado.ID)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if p.Nombre != "Software Metrics" {
		t.Errorf("se esperaba el proyecto %q y se obtuvo %+v", "Software Metrics", p)
	}

	_, err = repo.BuscarPorID(ctx, guardado.ID+99)
	if !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("id inexistente: se esperaba project.ErrNoEncontrado y se obtuvo %v", err)
	}
}

func TestProyectos_ListarSinProyectosDevuelveListaVacia(t *testing.T) {
	lista, err := postgres.NuevoProyectos(baseMigrada(t)).Listar(context.Background())
	if err != nil || len(lista) != 0 {
		t.Errorf("se esperaba una lista vacía sin error y se obtuvo %+v, %v", lista, err)
	}
}

// Dos pedidos a la vez pueden pasar el ExisteNombre del caso de uso: el índice único de la base es la última defensa.
func TestProyectos_LaBaseRechazaUnNombreRepetidoAunqueSeSalteExisteNombre(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NuevoProyectos(baseMigrada(t))
	if _, err := repo.Guardar(ctx, proyectoDeEjemplo("Repetido")); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if _, err := repo.Guardar(ctx, proyectoDeEjemplo("REPETIDO")); err == nil {
		t.Error("se esperaba un error por el nombre repetido y se guardó")
	}
	if lista, _ := repo.Listar(ctx); len(lista) != 1 {
		t.Errorf("tenía que quedar un solo proyecto y hay %d", len(lista))
	}
}
