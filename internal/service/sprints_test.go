package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// escenarioSprints arma los repositorios en memoria con el proyecto de ejemplo ya creado
// (del 05/10/2026 al 01/11/2026) y devuelve el caso de uso y el ID del proyecto.
func escenarioSprints(t *testing.T) (*service.Sprints, *memory.Sprints, int64) {
	t.Helper()
	proyectos := memory.NuevoProyectos()
	p, err := service.NuevoProyectos(proyectos, relojFijo).Crear(context.Background(), datosValidos())
	if err != nil {
		t.Fatalf("no se pudo crear el proyecto de ejemplo: %v", err)
	}
	sprints := memory.NuevoSprints()
	return service.NuevoSprints(proyectos, sprints), sprints, p.ID
}

func datosDeSprint(inicio, fin string) sprint.Datos {
	i, _ := time.Parse(time.DateOnly, inicio)
	f, _ := time.Parse(time.DateOnly, fin)
	return sprint.Datos{Objetivo: "Primer MVP", FechaInicio: i, FechaFin: f}
}

func cantidadDeSprints(t *testing.T, repo *memory.Sprints, proyectoID int64) int {
	t.Helper()
	lista, err := repo.ListarPorProyecto(context.Background(), proyectoID)
	if err != nil {
		t.Fatalf("no se esperaba error al listar: %v", err)
	}
	return len(lista)
}

// CA-11.3 · el sprint queda guardado con ID, número correlativo y estado Planificado
func TestCrearSprint_DatosValidos_GuardaConNumeroCorrelativo(t *testing.T) {
	s, repo, proyectoID := escenarioSprints(t)

	primero, err := s.Crear(context.Background(), proyectoID, datosDeSprint("2026-10-05", "2026-10-11"))
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	segundo, err := s.Crear(context.Background(), proyectoID, datosDeSprint("2026-10-12", "2026-10-18"))
	if err != nil {
		t.Fatalf("no se esperaba error en el segundo: %v", err)
	}

	if primero.ID == 0 || primero.Numero != 1 || segundo.Numero != 2 {
		t.Errorf("se esperaban los sprints 1 y 2 con ID y se obtuvo %+v y %+v", primero, segundo)
	}
	if primero.ProyectoID != proyectoID || primero.Estado != sprint.EstadoPlanificado {
		t.Errorf("se esperaba proyecto %d y estado Planificado y se obtuvo %+v", proyectoID, primero)
	}
	if n := cantidadDeSprints(t, repo, proyectoID); n != 2 {
		t.Errorf("se esperaban 2 sprints guardados y hay %d", n)
	}
}

// CA-11.2 · el caso de uso le pasa al dominio los sprints que ya tiene el proyecto
func TestCrearSprint_SuperpuestoConElUltimo_NoGuarda(t *testing.T) {
	s, repo, proyectoID := escenarioSprints(t)
	_, _ = s.Crear(context.Background(), proyectoID, datosDeSprint("2026-10-05", "2026-10-11"))

	_, err := s.Crear(context.Background(), proyectoID, datosDeSprint("2026-10-11", "2026-10-18"))

	if got := errorDeCampo(t, err, sprint.CampoFechaInicio); !errors.Is(got, sprint.ErrSuperpuesto) {
		t.Errorf("fecha_inicio: se esperaba %v y se obtuvo %v", sprint.ErrSuperpuesto, got)
	}
	if n := cantidadDeSprints(t, repo, proyectoID); n != 1 {
		t.Errorf("se esperaba que siga habiendo 1 sprint y hay %d", n)
	}
}

// Si el proyecto no existe se avisa con project.ErrNoEncontrado (la pantalla muestra 404).
func TestCrearSprint_ProyectoInexistente_DevuelveNoEncontrado(t *testing.T) {
	s, repo, _ := escenarioSprints(t)

	_, err := s.Crear(context.Background(), 99, datosDeSprint("2026-10-05", "2026-10-11"))

	if !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("se esperaba project.ErrNoEncontrado y se obtuvo: %v", err)
	}
	if n := cantidadDeSprints(t, repo, 99); n != 0 {
		t.Errorf("no se esperaba ningún sprint y hay %d", n)
	}
}

// La lista trae el proyecto (para el título y su rango de fechas) y sus sprints.
func TestListarSprints_DevuelveElProyectoYSusSprints(t *testing.T) {
	s, _, proyectoID := escenarioSprints(t)
	_, _ = s.Crear(context.Background(), proyectoID, datosDeSprint("2026-10-05", "2026-10-11"))

	p, lista, err := s.Listar(context.Background(), proyectoID)

	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if p.ID != proyectoID || len(lista) != 1 || lista[0].Numero != 1 {
		t.Errorf("se esperaba el proyecto %d con el Sprint 1 y se obtuvo %+v, %+v", proyectoID, p, lista)
	}
}

func TestListarSprints_ProyectoInexistente_DevuelveNoEncontrado(t *testing.T) {
	s, _, _ := escenarioSprints(t)

	_, _, err := s.Listar(context.Background(), 99)

	if !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("se esperaba project.ErrNoEncontrado y se obtuvo: %v", err)
	}
}

// proyectosQueFallan y sprintsQueFallan simulan una base de datos caída.
type proyectosQueFallan struct{ err error }

func (r proyectosQueFallan) BuscarPorID(context.Context, int64) (project.Proyecto, error) {
	return project.Proyecto{}, r.err
}

type sprintsQueFallan struct{ errListar, errGuardar error }

func (r sprintsQueFallan) ListarPorProyecto(context.Context, int64) ([]sprint.Sprint, error) {
	return nil, r.errListar
}
func (r sprintsQueFallan) Guardar(_ context.Context, s sprint.Sprint) (sprint.Sprint, error) {
	return s, r.errGuardar
}

// proyectoDeEjemplo cumple la interfaz de búsqueda devolviendo siempre el proyecto de ejemplo.
type proyectoDeEjemplo struct{}

func (proyectoDeEjemplo) BuscarPorID(context.Context, int64) (project.Proyecto, error) {
	d := datosValidos()
	return project.Proyecto{ID: 1, Nombre: d.Nombre, FechaInicio: d.FechaInicio, FechaFin: d.FechaFin}, nil
}

// Un error del repositorio no es de validación: se devuelve para que la pantalla muestre un error general.
func TestCrearSprint_ErrorDelRepositorio_SeDevuelve(t *testing.T) {
	falla := errors.New("base caída")
	casos := map[string]*service.Sprints{
		"al buscar el proyecto": service.NuevoSprints(proyectosQueFallan{err: falla}, sprintsQueFallan{}),
		"al listar los sprints": service.NuevoSprints(proyectoDeEjemplo{}, sprintsQueFallan{errListar: falla}),
		"al guardar":            service.NuevoSprints(proyectoDeEjemplo{}, sprintsQueFallan{errGuardar: falla}),
	}
	for nombre, s := range casos {
		_, err := s.Crear(context.Background(), 1, datosDeSprint("2026-10-05", "2026-10-11"))

		if !errors.Is(err, falla) {
			t.Errorf("%s: se esperaba el error del repositorio y se obtuvo: %v", nombre, err)
		}
	}
}

func TestListarSprints_ErrorDelRepositorio_SeDevuelve(t *testing.T) {
	falla := errors.New("base caída")
	casos := map[string]*service.Sprints{
		"al buscar el proyecto": service.NuevoSprints(proyectosQueFallan{err: falla}, sprintsQueFallan{}),
		"al listar los sprints": service.NuevoSprints(proyectoDeEjemplo{}, sprintsQueFallan{errListar: falla}),
	}
	for nombre, s := range casos {
		_, _, err := s.Listar(context.Background(), 1)

		if !errors.Is(err, falla) {
			t.Errorf("%s: se esperaba el error del repositorio y se obtuvo: %v", nombre, err)
		}
	}
}
