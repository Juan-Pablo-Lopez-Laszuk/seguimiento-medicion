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

// escenarioEstados arma los repositorios en memoria con el proyecto de ejemplo y su HU-1 (Pendiente) ya creados, y
// devuelve el caso de uso, los repositorios de historias y de criterios, el ID del proyecto y la HU-1.
func escenarioEstados(t *testing.T) (*service.Estados, *memory.Historias, *memory.Criterios, int64, backlog.Historia) {
	t.Helper()
	proyectos := memory.NuevoProyectos()
	p, err := service.NuevoProyectos(proyectos, relojFijo).Crear(context.Background(), datosValidos())
	if err != nil {
		t.Fatalf("no se pudo crear el proyecto de ejemplo: %v", err)
	}
	historias, criterios := memory.NuevoHistorias(), memory.NuevoCriterios()
	h, err := service.NuevoHistorias(proyectos, historias).Crear(context.Background(), p.ID, datosDeHistoria("Iniciar sesión con Google"))
	if err != nil {
		t.Fatalf("no se pudo crear la historia de ejemplo: %v", err)
	}
	return service.NuevoEstados(proyectos, historias, criterios), historias, criterios, p.ID, h
}

// ponerEn deja la historia en ese estado directo en el repositorio (los cambios de estado de HU-12 y HU-13 todavía no
// están hechos, y esos estados son el punto de partida de varias pruebas).
func ponerEn(t *testing.T, historias *memory.Historias, historiaID int64, estado backlog.Estado) {
	t.Helper()
	if err := historias.ActualizarEstado(context.Background(), historiaID, estado); err != nil {
		t.Fatalf("no se pudo dejar la historia en %q: %v", estado, err)
	}
}

// conCriterios guarda total criterios en la historia, los primeros cumplidos, directo en el repositorio.
func conCriterios(t *testing.T, criterios *memory.Criterios, historiaID int64, total, cumplidos int) {
	t.Helper()
	for i := 1; i <= total; i++ {
		if _, err := criterios.Guardar(context.Background(), backlog.Criterio{HistoriaID: historiaID, Descripcion: "Criterio", Cumplido: i <= cumplidos}); err != nil {
			t.Fatalf("no se pudo guardar el criterio %d: %v", i, err)
		}
	}
}

func estadoGuardado(t *testing.T, historias *memory.Historias, proyectoID int64, numero int) backlog.Estado {
	t.Helper()
	h, err := historias.BuscarPorNumero(context.Background(), proyectoID, numero)
	if err != nil {
		t.Fatalf("no se esperaba error al buscar la historia: %v", err)
	}
	return h.Estado
}

// CA-08.1 · el cambio válido queda guardado y el caso de uso devuelve la historia con el estado nuevo
func TestCambiarEstado_Valido_GuardaElEstadoNuevo(t *testing.T) {
	s, historias, _, proyectoID, h := escenarioEstados(t)

	nueva, err := s.Cambiar(context.Background(), proyectoID, h.Numero, backlog.EstadoEnSprint)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if nueva.Estado != backlog.EstadoEnSprint || nueva.ID != h.ID || nueva.Titulo != h.Titulo {
		t.Errorf("se esperaba la misma historia En sprint y se obtuvo %+v", nueva)
	}
	if guardado := estadoGuardado(t, historias, proyectoID, h.Numero); guardado != backlog.EstadoEnSprint {
		t.Errorf("estado guardado = %q, se esperaba En sprint", guardado)
	}
}

// CA-08.1 y CA-08.2 · el camino completo hasta Hecho, con todos los criterios cumplidos
func TestCambiarEstado_ElCaminoCompletoHastaHecho(t *testing.T) {
	s, historias, criterios, proyectoID, h := escenarioEstados(t)
	conCriterios(t, criterios, h.ID, 2, 2)

	for _, estado := range []backlog.Estado{backlog.EstadoEnSprint, backlog.EstadoEnProgreso, backlog.EstadoHecho} {
		if _, err := s.Cambiar(context.Background(), proyectoID, h.Numero, estado); err != nil {
			t.Fatalf("pasar a %q: no se esperaba error: %v", estado, err)
		}
	}

	if guardado := estadoGuardado(t, historias, proyectoID, h.Numero); guardado != backlog.EstadoHecho {
		t.Errorf("estado guardado = %q, se esperaba Hecho", guardado)
	}
}

// CA-08.2 y RN7 · con criterios sin cumplir no pasa a Hecho, y la historia queda como estaba
func TestCambiarEstado_AHechoConCriteriosSinCumplir_NoCambiaNada(t *testing.T) {
	s, historias, criterios, proyectoID, h := escenarioEstados(t)
	ponerEn(t, historias, h.ID, backlog.EstadoEnProgreso)
	conCriterios(t, criterios, h.ID, 3, 2)

	_, err := s.Cambiar(context.Background(), proyectoID, h.Numero, backlog.EstadoHecho)

	if !errors.Is(err, backlog.ErrCriteriosSinCumplir) {
		t.Fatalf("se esperaba backlog.ErrCriteriosSinCumplir y se obtuvo: %v", err)
	}
	if guardado := estadoGuardado(t, historias, proyectoID, h.Numero); guardado != backlog.EstadoEnProgreso {
		t.Errorf("estado guardado = %q, se esperaba que siga En progreso", guardado)
	}
}

// RN4 · una historia sin criterios tampoco pasa a Hecho
func TestCambiarEstado_AHechoSinCriterios_NoPasa(t *testing.T) {
	s, historias, _, proyectoID, h := escenarioEstados(t)
	ponerEn(t, historias, h.ID, backlog.EstadoEnProgreso)

	_, err := s.Cambiar(context.Background(), proyectoID, h.Numero, backlog.EstadoHecho)

	if !errors.Is(err, backlog.ErrCriteriosSinCumplir) {
		t.Errorf("se esperaba backlog.ErrCriteriosSinCumplir y se obtuvo: %v", err)
	}
}

// RN8 · los criterios de otra historia no cuentan: cada historia se evalúa con los suyos
func TestCambiarEstado_AHecho_UsaSoloLosCriteriosDeLaHistoria(t *testing.T) {
	s, historias, criterios, proyectoID, h := escenarioEstados(t)
	ponerEn(t, historias, h.ID, backlog.EstadoEnProgreso)
	conCriterios(t, criterios, h.ID+100, 3, 3) // cumplidos, pero de otra historia

	_, err := s.Cambiar(context.Background(), proyectoID, h.Numero, backlog.EstadoHecho)

	if !errors.Is(err, backlog.ErrCriteriosSinCumplir) {
		t.Errorf("se esperaba backlog.ErrCriteriosSinCumplir y se obtuvo: %v", err)
	}
}

// CA-08.3 · una transición inválida da un error descriptivo y no cambia nada
func TestCambiarEstado_TransicionInvalida_NoCambiaNada(t *testing.T) {
	s, historias, _, proyectoID, h := escenarioEstados(t)

	_, err := s.Cambiar(context.Background(), proyectoID, h.Numero, backlog.EstadoHecho)

	if !errors.Is(err, backlog.ErrTransicionInvalida) {
		t.Fatalf("se esperaba backlog.ErrTransicionInvalida y se obtuvo: %v", err)
	}
	if guardado := estadoGuardado(t, historias, proyectoID, h.Numero); guardado != backlog.EstadoPendiente {
		t.Errorf("estado guardado = %q, se esperaba que siga Pendiente", guardado)
	}
}

// El proyecto y la historia tienen que existir
func TestCambiarEstado_ProyectoOHistoriaInexistente_DevuelveNoEncontrado(t *testing.T) {
	s, _, _, proyectoID, _ := escenarioEstados(t)

	if _, err := s.Cambiar(context.Background(), 99, 1, backlog.EstadoEnSprint); !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("proyecto inexistente: se esperaba project.ErrNoEncontrado y se obtuvo %v", err)
	}
	if _, err := s.Cambiar(context.Background(), proyectoID, 99, backlog.EstadoEnSprint); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("historia inexistente: se esperaba backlog.ErrHistoriaNoEncontrada y se obtuvo %v", err)
	}
}

// historiasQueFallanAlCambiarEstado simula una base de datos caída para buscar la historia y para guardar su estado.
type historiasQueFallanAlCambiarEstado struct{ errBuscar, errGuardar error }

func (h historiasQueFallanAlCambiarEstado) BuscarPorNumero(_ context.Context, proyectoID int64, numero int) (backlog.Historia, error) {
	return backlog.Historia{ID: 4, ProyectoID: proyectoID, Numero: numero, Estado: backlog.EstadoPendiente}, h.errBuscar
}
func (h historiasQueFallanAlCambiarEstado) ActualizarEstado(context.Context, int64, backlog.Estado) error {
	return h.errGuardar
}

// Un error del repositorio no es de validación ni de "no encontrado": se devuelve para mostrar un error general.
func TestCambiarEstado_ErrorDelRepositorio_SeDevuelve(t *testing.T) {
	falla := errors.New("base caída")
	ctx := context.Background()
	casos := map[string]func() error{
		"al buscar el proyecto": func() error {
			_, err := service.NuevoEstados(proyectosQueFallan{err: falla}, historiasQueFallanAlCambiarEstado{}, criteriosQueFallan{}).Cambiar(ctx, 1, 1, backlog.EstadoEnSprint)
			return err
		},
		"al buscar la historia": func() error {
			_, err := service.NuevoEstados(proyectoDeEjemplo{}, historiasQueFallanAlCambiarEstado{errBuscar: falla}, criteriosQueFallan{}).Cambiar(ctx, 1, 1, backlog.EstadoEnSprint)
			return err
		},
		"al guardar el estado": func() error {
			_, err := service.NuevoEstados(proyectoDeEjemplo{}, historiasQueFallanAlCambiarEstado{errGuardar: falla}, criteriosQueFallan{}).Cambiar(ctx, 1, 1, backlog.EstadoEnSprint)
			return err
		},
		"al listar los criterios": func() error {
			h := historiasQueFallanAlCambiarEstado{}
			// la historia de ejemplo está Pendiente: la dejamos En progreso para que haga falta mirar los criterios
			_, err := service.NuevoEstados(proyectoDeEjemplo{}, historiaEnProgreso{h}, criteriosQueFallan{errListar: falla}).Cambiar(ctx, 1, 1, backlog.EstadoHecho)
			return err
		},
	}
	for nombre, accion := range casos {
		if err := accion(); !errors.Is(err, falla) {
			t.Errorf("%s: se esperaba el error del repositorio y se obtuvo: %v", nombre, err)
		}
	}
}

// historiaEnProgreso hace que la historia de ejemplo esté En progreso.
type historiaEnProgreso struct {
	historiasQueFallanAlCambiarEstado
}

func (h historiaEnProgreso) BuscarPorNumero(ctx context.Context, proyectoID int64, numero int) (backlog.Historia, error) {
	historia, err := h.historiasQueFallanAlCambiarEstado.BuscarPorNumero(ctx, proyectoID, numero)
	historia.Estado = backlog.EstadoEnProgreso
	return historia, err
}
