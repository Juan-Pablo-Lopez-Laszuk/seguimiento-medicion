package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// escenarioHistorias arma los repositorios en memoria con el proyecto de ejemplo ya creado y devuelve el caso de
// uso, el repositorio de historias y el ID del proyecto.
func escenarioHistorias(t *testing.T) (*service.Historias, *memory.Historias, int64) {
	t.Helper()
	proyectos := memory.NuevoProyectos()
	p, err := service.NuevoProyectos(proyectos, relojFijo).Crear(context.Background(), datosValidos())
	if err != nil {
		t.Fatalf("no se pudo crear el proyecto de ejemplo: %v", err)
	}
	historias := memory.NuevoHistorias()
	return service.NuevoHistorias(proyectos, historias), historias, p.ID
}

func datosDeHistoria(titulo string) backlog.Datos {
	return backlog.Datos{Titulo: titulo, Prioridad: backlog.PrioridadAlta}
}

func cantidadDeHistorias(t *testing.T, repo *memory.Historias, proyectoID int64) int {
	t.Helper()
	lista, err := repo.ListarPorProyecto(context.Background(), proyectoID)
	if err != nil {
		t.Fatalf("no se esperaba error al listar: %v", err)
	}
	return len(lista)
}

// CA-05.1 y CA-05.3 · la historia queda guardada con ID, número correlativo y estado Pendiente
func TestCrearHistoria_DatosValidos_GuardaConNumeroCorrelativo(t *testing.T) {
	s, repo, proyectoID := escenarioHistorias(t)

	primera, err := s.Crear(context.Background(), proyectoID, datosDeHistoria("Primera"))
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	segunda, err := s.Crear(context.Background(), proyectoID, datosDeHistoria("Segunda"))
	if err != nil {
		t.Fatalf("no se esperaba error en la segunda: %v", err)
	}

	if primera.ID == 0 || primera.Numero != 1 || segunda.Numero != 2 {
		t.Errorf("se esperaban las historias 1 y 2 con ID y se obtuvo %+v y %+v", primera, segunda)
	}
	if primera.ProyectoID != proyectoID || primera.Estado != backlog.EstadoPendiente {
		t.Errorf("se esperaba proyecto %d y estado Pendiente y se obtuvo %+v", proyectoID, primera)
	}
	if n := cantidadDeHistorias(t, repo, proyectoID); n != 2 {
		t.Errorf("se esperaban 2 historias guardadas y hay %d", n)
	}
}

// CA-05.1 · el caso de uso le pasa al dominio las historias que ya tiene el proyecto
func TestCrearHistoria_CadaProyectoNumeraLasSuyas(t *testing.T) {
	proyectos := memory.NuevoProyectos()
	proyectosS := service.NuevoProyectos(proyectos, relojFijo)
	a, _ := proyectosS.Crear(context.Background(), datosValidos())
	otros := datosValidos()
	otros.Nombre = "Otro proyecto"
	b, _ := proyectosS.Crear(context.Background(), otros)
	s := service.NuevoHistorias(proyectos, memory.NuevoHistorias())
	_, _ = s.Crear(context.Background(), a.ID, datosDeHistoria("Del primer proyecto"))
	_, _ = s.Crear(context.Background(), a.ID, datosDeHistoria("Otra del primer proyecto"))

	primeraDelOtro, err := s.Crear(context.Background(), b.ID, datosDeHistoria("Del otro proyecto"))
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if primeraDelOtro.Numero != 1 {
		t.Errorf("número = %d, se esperaba 1 (cada proyecto numera las suyas)", primeraDelOtro.Numero)
	}
}

// CA-05.2 · los datos inválidos se devuelven juntos y no se guarda nada
func TestCrearHistoria_DatosInvalidos_NoGuarda(t *testing.T) {
	s, repo, proyectoID := escenarioHistorias(t)

	_, err := s.Crear(context.Background(), proyectoID, backlog.Datos{Titulo: "  ", Prioridad: "Urgente"})

	if got := errorDeCampo(t, err, backlog.CampoTitulo); !errors.Is(got, backlog.ErrTituloObligatorio) {
		t.Errorf("titulo: se esperaba %v y se obtuvo %v", backlog.ErrTituloObligatorio, got)
	}
	if got := errorDeCampo(t, err, backlog.CampoPrioridad); !errors.Is(got, backlog.ErrPrioridadInvalida) {
		t.Errorf("prioridad: se esperaba %v y se obtuvo %v", backlog.ErrPrioridadInvalida, got)
	}
	if n := cantidadDeHistorias(t, repo, proyectoID); n != 0 {
		t.Errorf("no se esperaba ninguna historia guardada y hay %d", n)
	}
}

// Si el proyecto no existe se avisa con project.ErrNoEncontrado (la pantalla muestra 404).
func TestCrearHistoria_ProyectoInexistente_DevuelveNoEncontrado(t *testing.T) {
	s, repo, _ := escenarioHistorias(t)

	_, err := s.Crear(context.Background(), 99, datosDeHistoria("Huérfana"))

	if !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("se esperaba project.ErrNoEncontrado y se obtuvo: %v", err)
	}
	if n := cantidadDeHistorias(t, repo, 99); n != 0 {
		t.Errorf("no se esperaba ninguna historia y hay %d", n)
	}
}

// La lista trae el proyecto (para el título de la pantalla) y sus historias, en orden de número.
func TestListarHistorias_DevuelveElProyectoYSusHistorias(t *testing.T) {
	s, _, proyectoID := escenarioHistorias(t)
	_, _ = s.Crear(context.Background(), proyectoID, datosDeHistoria("Primera"))

	p, lista, err := s.Listar(context.Background(), proyectoID)

	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if p.ID != proyectoID || len(lista) != 1 || lista[0].Numero != 1 {
		t.Errorf("se esperaba el proyecto %d con la HU-1 y se obtuvo %+v, %+v", proyectoID, p, lista)
	}
}

func TestListarHistorias_ProyectoInexistente_DevuelveNoEncontrado(t *testing.T) {
	s, _, _ := escenarioHistorias(t)

	_, _, err := s.Listar(context.Background(), 99)

	if !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("se esperaba project.ErrNoEncontrado y se obtuvo: %v", err)
	}
}

// historiasQueFallan simula una base de datos caída.
type historiasQueFallan struct{ errListar, errGuardar error }

func (r historiasQueFallan) ListarPorProyecto(context.Context, int64) ([]backlog.Historia, error) {
	return nil, r.errListar
}
func (r historiasQueFallan) Guardar(_ context.Context, h backlog.Historia) (backlog.Historia, error) {
	return h, r.errGuardar
}
func (r historiasQueFallan) ActualizarOrden(context.Context, int64, map[int64]int) error {
	return r.errGuardar
}

// Un error del repositorio no es de validación: se devuelve para que la pantalla muestre un error general.
func TestCrearHistoria_ErrorDelRepositorio_SeDevuelve(t *testing.T) {
	falla := errors.New("base caída")
	casos := map[string]*service.Historias{
		"al buscar el proyecto":   service.NuevoHistorias(proyectosQueFallan{err: falla}, historiasQueFallan{}),
		"al listar las historias": service.NuevoHistorias(proyectoDeEjemplo{}, historiasQueFallan{errListar: falla}),
		"al guardar":              service.NuevoHistorias(proyectoDeEjemplo{}, historiasQueFallan{errGuardar: falla}),
	}
	for nombre, s := range casos {
		_, err := s.Crear(context.Background(), 1, datosDeHistoria("Primera"))

		if !errors.Is(err, falla) {
			t.Errorf("%s: se esperaba el error del repositorio y se obtuvo: %v", nombre, err)
		}
	}
}

func TestListarHistorias_ErrorDelRepositorio_SeDevuelve(t *testing.T) {
	falla := errors.New("base caída")
	casos := map[string]*service.Historias{
		"al buscar el proyecto":   service.NuevoHistorias(proyectosQueFallan{err: falla}, historiasQueFallan{}),
		"al listar las historias": service.NuevoHistorias(proyectoDeEjemplo{}, historiasQueFallan{errListar: falla}),
	}
	for nombre, s := range casos {
		_, _, err := s.Listar(context.Background(), 1)

		if !errors.Is(err, falla) {
			t.Errorf("%s: se esperaba el error del repositorio y se obtuvo: %v", nombre, err)
		}
	}
}
