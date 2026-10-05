package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// ahoraFijo hace que la fecha de creación sea siempre la misma, así el test no depende del reloj.
var ahoraFijo = time.Date(2026, 10, 5, 14, 30, 0, 0, time.UTC)

func relojFijo() time.Time { return ahoraFijo }

func datosValidos() project.Datos {
	return project.Datos{
		Nombre:      "Software Metrics",
		Descripcion: "TPI de ICS",
		FechaInicio: time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC),
		FechaFin:    time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
	}
}

func cantidadDeProyectos(t *testing.T, repo *memory.Proyectos) int {
	t.Helper()
	lista, err := repo.Listar(context.Background())
	if err != nil {
		t.Fatalf("no se esperaba error al listar: %v", err)
	}
	return len(lista)
}

func errorDeCampo(t *testing.T, err error, campo string) error {
	t.Helper()
	var errs project.ErroresValidacion
	if !errors.As(err, &errs) {
		t.Fatalf("se esperaba project.ErroresValidacion y se obtuvo: %v", err)
	}
	return errs[campo]
}

// CA-01.3 · el proyecto queda guardado, con ID, fecha de creación y estado Planificado
func TestCrear_DatosValidos_GuardaElProyecto(t *testing.T) {
	repo := memory.NuevoProyectos()
	s := service.NuevoProyectos(repo, relojFijo)

	p, err := s.Crear(context.Background(), datosValidos())

	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if p.ID == 0 {
		t.Error("se esperaba que el proyecto tenga ID")
	}
	if !p.CreadoEn.Equal(ahoraFijo) {
		t.Errorf("creado en: se esperaba %v y se obtuvo %v", ahoraFijo, p.CreadoEn)
	}
	if p.Estado != project.EstadoPlanificado {
		t.Errorf("estado: se esperaba %q y se obtuvo %q", project.EstadoPlanificado, p.Estado)
	}
	if n := cantidadDeProyectos(t, repo); n != 1 {
		t.Errorf("se esperaba 1 proyecto guardado y hay %d", n)
	}
}

// CA-01.1 · RN3 · nombre repetido sin distinguir mayúsculas ni espacios alrededor
func TestCrear_NombreRepetido_NoGuardaEInformaElCampo(t *testing.T) {
	repo := memory.NuevoProyectos()
	s := service.NuevoProyectos(repo, relojFijo)
	if _, err := s.Crear(context.Background(), datosValidos()); err != nil {
		t.Fatalf("no se esperaba error al crear el primero: %v", err)
	}

	d := datosValidos()
	d.Nombre = "  software METRICS  "
	_, err := s.Crear(context.Background(), d)

	if got := errorDeCampo(t, err, project.CampoNombre); !errors.Is(got, project.ErrNombreRepetido) {
		t.Errorf("nombre: se esperaba %v y se obtuvo %v", project.ErrNombreRepetido, got)
	}
	if n := cantidadDeProyectos(t, repo); n != 1 {
		t.Errorf("se esperaba que siga habiendo 1 proyecto y hay %d", n)
	}
}

// CA-01.4 · RN7 · el nombre repetido se informa junto con los demás errores y no se guarda nada
func TestCrear_DatosInvalidosYNombreRepetido_InformaTodoJunto(t *testing.T) {
	repo := memory.NuevoProyectos()
	s := service.NuevoProyectos(repo, relojFijo)
	_, _ = s.Crear(context.Background(), datosValidos())

	d := datosValidos()
	d.FechaFin = d.FechaInicio // inválida: el mismo día
	_, err := s.Crear(context.Background(), d)

	if got := errorDeCampo(t, err, project.CampoNombre); !errors.Is(got, project.ErrNombreRepetido) {
		t.Errorf("nombre: se esperaba %v y se obtuvo %v", project.ErrNombreRepetido, got)
	}
	if got := errorDeCampo(t, err, project.CampoFechaFin); !errors.Is(got, project.ErrFechasInvalidas) {
		t.Errorf("fecha_fin: se esperaba %v y se obtuvo %v", project.ErrFechasInvalidas, got)
	}
	if n := cantidadDeProyectos(t, repo); n != 1 {
		t.Errorf("se esperaba que siga habiendo 1 proyecto y hay %d", n)
	}
}

// repoQueFalla simula una base de datos caída: falla al buscar el nombre, al guardar o en ambos.
type repoQueFalla struct{ errBuscar, errGuardar error }

func (r repoQueFalla) ExisteNombre(context.Context, string) (bool, error) { return false, r.errBuscar }
func (r repoQueFalla) Guardar(_ context.Context, p project.Proyecto) (project.Proyecto, error) {
	return p, r.errGuardar
}

// Un error del repositorio no es un error de validación: se devuelve tal cual para que la pantalla
// muestre un error general.
func TestCrear_ErrorDelRepositorio_SeDevuelve(t *testing.T) {
	falla := errors.New("base caída")
	casos := map[string]repoQueFalla{
		"al buscar el nombre": {errBuscar: falla},
		"al guardar":          {errGuardar: falla},
	}
	for nombre, repo := range casos {
		s := service.NuevoProyectos(repo, relojFijo)

		_, err := s.Crear(context.Background(), datosValidos())

		if !errors.Is(err, falla) {
			t.Errorf("%s: se esperaba el error del repositorio y se obtuvo: %v", nombre, err)
		}
	}
}
