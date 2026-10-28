package poker_test

import (
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
)

var (
	sesionAbierta  = poker.Sesion{ID: 1, HistoriaID: 7, Estado: poker.EstadoAbierta}
	rondaRevelada  = poker.Ronda{ID: 1, SesionID: 1, Numero: 1, Revelada: true}
	rondaSinVotar  = poker.Ronda{ID: 1, SesionID: 1, Numero: 1}
	historiaLogin  = backlog.Historia{ID: 7, Numero: 1, Titulo: "Login", Estado: backlog.EstadoPendiente}
	tresStoryPoint = 3
)

// CA-22.2 · el valor acordado queda como story points de la historia
func TestAcordar_ConLaRondaRevelada_EstimaLaHistoria(t *testing.T) {
	h, err := poker.Acordar(historiaLogin, sesionAbierta, rondaRevelada, juan, "5")
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if h.StoryPoints == nil || *h.StoryPoints != 5 {
		t.Errorf("story points = %v, se esperaba 5", h.StoryPoints)
	}
	if historiaLogin.StoryPoints != nil {
		t.Error("la historia original no tenía que cambiar")
	}
}

// CA-22.1 · el valor pertenece a la escala; 0 es una estimación
func TestAcordar_ElValorTieneQueSerDeLaEscala(t *testing.T) {
	for _, valor := range []string{"0", "1", "2", "3", "5", "8", "13", "21"} {
		if _, err := poker.Acordar(historiaLogin, sesionAbierta, rondaRevelada, juan, valor); err != nil {
			t.Errorf("Acordar(%q): no se esperaba error: %v", valor, err)
		}
	}
	for _, valor := range []string{"4", "34", "-1", "", "cinco"} {
		_, err := poker.Acordar(historiaLogin, sesionAbierta, rondaRevelada, juan, valor)
		var errs domain.ErroresValidacion
		if !errors.As(err, &errs) || !errors.Is(errs[backlog.CampoStoryPoints], backlog.ErrStoryPointsInvalidos) {
			t.Errorf("Acordar(%q): error = %v, se esperaba ErrStoryPointsInvalidos en story_points", valor, err)
		}
	}
}

// CA-22.2 RN2 · re-estimar con Planning Poker reemplaza la estimación anterior
func TestAcordar_HistoriaYaEstimada_ReemplazaLaEstimacion(t *testing.T) {
	estimada := historiaLogin
	estimada.StoryPoints = &tresStoryPoint
	h, err := poker.Acordar(estimada, sesionAbierta, rondaRevelada, juan, "8")
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if *h.StoryPoints != 8 {
		t.Errorf("story points = %d, se esperaba 8", *h.StoryPoints)
	}
}

// RN3 · solo el Agile Enabler activo registra la estimación
func TestAcordar_SoloElAgileEnabler(t *testing.T) {
	juanDeBaja := juan
	juanDeBaja.Activo = false
	casos := []struct {
		nombre string
		quien  member.Integrante
	}{
		{"un Product Builder", carolina},
		{"un Agile Enabler dado de baja", juanDeBaja},
		{"nadie (no se eligió quién)", member.Integrante{}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := poker.Acordar(historiaLogin, sesionAbierta, rondaRevelada, c.quien, "5")
			if !errors.Is(err, poker.ErrSoloAgileEnablerAcuerda) {
				t.Errorf("error = %v, se esperaba ErrSoloAgileEnablerAcuerda", err)
			}
		})
	}
}

// RN4 · el valor se acuerda después de revelar la ronda
func TestAcordar_ConLaRondaSinRevelar_NoSeAcuerda(t *testing.T) {
	_, err := poker.Acordar(historiaLogin, sesionAbierta, rondaSinVotar, juan, "5")
	if !errors.Is(err, poker.ErrAcordarSinRevelar) {
		t.Errorf("error = %v, se esperaba ErrAcordarSinRevelar", err)
	}
}

// RN5 · una sesión cerrada no se vuelve a cerrar
func TestAcordar_SesionCerrada_NoSeAcuerda(t *testing.T) {
	cerrada := sesionAbierta
	cerrada.Estado = poker.EstadoCerrada
	_, err := poker.Acordar(historiaLogin, cerrada, rondaRevelada, juan, "5")
	if !errors.Is(err, poker.ErrSesionCerrada) {
		t.Errorf("error = %v, se esperaba ErrSesionCerrada", err)
	}
}

// RN7 · una historia Hecha no se estima, aunque tenga la sesión abierta
func TestAcordar_HistoriaHecha_NoSeEstima(t *testing.T) {
	hecha := historiaLogin
	hecha.Estado = backlog.EstadoHecho
	hecha.StoryPoints = &tresStoryPoint
	_, err := poker.Acordar(hecha, sesionAbierta, rondaRevelada, juan, "5")
	if !errors.Is(err, backlog.ErrEstimacionDeHistoriaHecha) {
		t.Errorf("error = %v, se esperaba ErrEstimacionDeHistoriaHecha", err)
	}
}

// Orden de las reglas: la sesión, quién registra, la ronda y por último el valor
func TestAcordar_OrdenDeLasReglas(t *testing.T) {
	cerrada := sesionAbierta
	cerrada.Estado = poker.EstadoCerrada
	if _, err := poker.Acordar(historiaLogin, cerrada, rondaSinVotar, carolina, "4"); !errors.Is(err, poker.ErrSesionCerrada) {
		t.Errorf("sesión cerrada primero: error = %v", err)
	}
	if _, err := poker.Acordar(historiaLogin, sesionAbierta, rondaSinVotar, carolina, "4"); !errors.Is(err, poker.ErrSoloAgileEnablerAcuerda) {
		t.Errorf("después quién registra: error = %v", err)
	}
	if _, err := poker.Acordar(historiaLogin, sesionAbierta, rondaSinVotar, juan, "4"); !errors.Is(err, poker.ErrAcordarSinRevelar) {
		t.Errorf("después la ronda: error = %v", err)
	}
}
