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

// escenarioEstimaciones arma los repositorios en memoria con el proyecto de ejemplo y su HU-1 (Pendiente, sin estimar)
// ya creados, y devuelve el caso de uso, el repositorio de historias, el ID del proyecto y la HU-1.
func escenarioEstimaciones(t *testing.T) (*service.Estimaciones, *memory.Historias, int64, backlog.Historia) {
	t.Helper()
	proyectos := memory.NuevoProyectos()
	p, err := service.NuevoProyectos(proyectos, relojFijo).Crear(context.Background(), datosValidos())
	if err != nil {
		t.Fatalf("no se pudo crear el proyecto de ejemplo: %v", err)
	}
	historias := memory.NuevoHistorias()
	h, err := service.NuevoHistorias(proyectos, historias).Crear(context.Background(), p.ID, datosDeHistoria("Iniciar sesión con Google"))
	if err != nil {
		t.Fatalf("no se pudo crear la historia de ejemplo: %v", err)
	}
	return service.NuevoEstimaciones(proyectos, historias), historias, p.ID, h
}

// puntosGuardados devuelve los story points que guarda el repositorio para la historia (nil si está sin estimar).
func puntosGuardados(t *testing.T, historias *memory.Historias, proyectoID int64, numero int) *int {
	t.Helper()
	h, err := historias.BuscarPorNumero(context.Background(), proyectoID, numero)
	if err != nil {
		t.Fatalf("no se esperaba error al buscar la historia: %v", err)
	}
	return h.StoryPoints
}

// CA-16.1 · estimar guarda los story points y devuelve la historia estimada
func TestEstimar_Valido_GuardaLosStoryPoints(t *testing.T) {
	s, historias, proyectoID, h := escenarioEstimaciones(t)

	estimada, err := s.Estimar(context.Background(), proyectoID, h.Numero, "5")
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if estimada.StoryPoints == nil || *estimada.StoryPoints != 5 || estimada.ID != h.ID {
		t.Errorf("se esperaba la misma historia estimada en 5 y se obtuvo %+v", estimada)
	}
	if guardados := puntosGuardados(t, historias, proyectoID, h.Numero); guardados == nil || *guardados != 5 {
		t.Errorf("story points guardados = %v, se esperaba 5", guardados)
	}
}

// RN5 y RN4 · re-estimar reemplaza el valor anterior, también con 0
func TestEstimar_ReestimarReemplazaElValorAnterior(t *testing.T) {
	s, historias, proyectoID, h := escenarioEstimaciones(t)

	for _, texto := range []string{"5", "8", "0", "21"} {
		if _, err := s.Estimar(context.Background(), proyectoID, h.Numero, texto); err != nil {
			t.Fatalf("estimar en %s: no se esperaba error: %v", texto, err)
		}
	}

	if guardados := puntosGuardados(t, historias, proyectoID, h.Numero); guardados == nil || *guardados != 21 {
		t.Errorf("story points guardados = %v, se esperaba 21", guardados)
	}
}

// CA-16.2 y RN7 · un valor inválido se informa en su campo y no cambia lo guardado
func TestEstimar_ValorInvalido_NoCambiaNada(t *testing.T) {
	s, historias, proyectoID, h := escenarioEstimaciones(t)
	_, _ = s.Estimar(context.Background(), proyectoID, h.Numero, "5")

	_, err := s.Estimar(context.Background(), proyectoID, h.Numero, "4")

	if got := errorDeCampo(t, err, backlog.CampoStoryPoints); !errors.Is(got, backlog.ErrStoryPointsInvalidos) {
		t.Errorf("story_points: se esperaba %v y se obtuvo %v", backlog.ErrStoryPointsInvalidos, got)
	}
	if guardados := puntosGuardados(t, historias, proyectoID, h.Numero); guardados == nil || *guardados != 5 {
		t.Errorf("story points guardados = %v, se esperaba que siga en 5", guardados)
	}
}

// RN6 · una historia Hecha no se re-estima y lo guardado no cambia
func TestEstimar_HistoriaHecha_NoSeReestima(t *testing.T) {
	s, historias, proyectoID, h := escenarioEstimaciones(t)
	_, _ = s.Estimar(context.Background(), proyectoID, h.Numero, "5")
	if err := historias.ActualizarEstado(context.Background(), h.ID, backlog.EstadoHecho); err != nil {
		t.Fatalf("no se pudo dejar la historia Hecha: %v", err)
	}

	_, err := s.Estimar(context.Background(), proyectoID, h.Numero, "8")

	if !errors.Is(err, backlog.ErrEstimacionDeHistoriaHecha) {
		t.Fatalf("se esperaba backlog.ErrEstimacionDeHistoriaHecha y se obtuvo: %v", err)
	}
	if guardados := puntosGuardados(t, historias, proyectoID, h.Numero); guardados == nil || *guardados != 5 {
		t.Errorf("story points guardados = %v, se esperaba que siga en 5", guardados)
	}
}

// El proyecto y la historia tienen que existir
func TestEstimar_ProyectoOHistoriaInexistente_DevuelveNoEncontrado(t *testing.T) {
	s, _, proyectoID, _ := escenarioEstimaciones(t)

	if _, err := s.Estimar(context.Background(), 99, 1, "5"); !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("proyecto inexistente: se esperaba project.ErrNoEncontrado y se obtuvo %v", err)
	}
	if _, err := s.Estimar(context.Background(), proyectoID, 99, "5"); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("historia inexistente: se esperaba backlog.ErrHistoriaNoEncontrada y se obtuvo %v", err)
	}
}

// historiasQueFallanAlEstimar simula una base de datos caída para buscar la historia y para guardar sus story points.
type historiasQueFallanAlEstimar struct{ errBuscar, errGuardar error }

func (h historiasQueFallanAlEstimar) BuscarPorNumero(_ context.Context, proyectoID int64, numero int) (backlog.Historia, error) {
	return backlog.Historia{ID: 4, ProyectoID: proyectoID, Numero: numero, Estado: backlog.EstadoPendiente}, h.errBuscar
}
func (h historiasQueFallanAlEstimar) ActualizarStoryPoints(context.Context, int64, int) error {
	return h.errGuardar
}

// Un error del repositorio no es de validación ni de "no encontrado": se devuelve para mostrar un error general.
func TestEstimar_ErrorDelRepositorio_SeDevuelve(t *testing.T) {
	falla := errors.New("base caída")
	ctx := context.Background()
	casos := map[string]func() error{
		"al buscar el proyecto": func() error {
			_, err := service.NuevoEstimaciones(proyectosQueFallan{err: falla}, historiasQueFallanAlEstimar{}).Estimar(ctx, 1, 1, "5")
			return err
		},
		"al buscar la historia": func() error {
			_, err := service.NuevoEstimaciones(proyectoDeEjemplo{}, historiasQueFallanAlEstimar{errBuscar: falla}).Estimar(ctx, 1, 1, "5")
			return err
		},
		"al guardar": func() error {
			_, err := service.NuevoEstimaciones(proyectoDeEjemplo{}, historiasQueFallanAlEstimar{errGuardar: falla}).Estimar(ctx, 1, 1, "5")
			return err
		},
	}
	for nombre, accion := range casos {
		if err := accion(); !errors.Is(err, falla) {
			t.Errorf("%s: se esperaba el error del repositorio y se obtuvo: %v", nombre, err)
		}
	}
}
