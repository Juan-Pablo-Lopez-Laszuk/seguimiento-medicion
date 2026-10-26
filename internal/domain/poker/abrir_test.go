package poker_test

import (
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
)

var (
	juan     = member.Integrante{ID: 1, Nombre: "Juan", Rol: member.RolAgileEnabler, Activo: true}
	carolina = member.Integrante{ID: 2, Nombre: "Carolina", Rol: member.RolProductBuilder, Activo: true}
	mariano  = member.Integrante{ID: 3, Nombre: "Mariano", Rol: member.RolProductBuilder, Activo: true}
	lucia    = member.Integrante{ID: 4, Nombre: "Lucía", Rol: member.RolProductBuilder}
)

// HU-17 RN4 · solo el Agile Enabler activo conduce la sesión
func TestPuedeConducir(t *testing.T) {
	juanDeBaja := juan
	juanDeBaja.Activo = false
	casos := []struct {
		nombre string
		quien  member.Integrante
		err    error
	}{
		{"el Agile Enabler activo", juan, nil},
		{"un Product Builder", carolina, poker.ErrSoloAgileEnabler},
		{"un Agile Enabler dado de baja", juanDeBaja, poker.ErrSoloAgileEnabler},
		{"nadie (no se eligió quién)", member.Integrante{}, poker.ErrSoloAgileEnabler},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if err := poker.PuedeConducir(c.quien); !errors.Is(err, c.err) {
				t.Errorf("error = %v, se esperaba %v", err, c.err)
			}
		})
	}
}

// CA-17.1 y RN1 · se abre para una historia que no está Hecha; RN4 · la abre el Agile Enabler
func TestPuedeAbrir(t *testing.T) {
	casos := []struct {
		nombre string
		estado backlog.Estado
		quien  member.Integrante
		err    error
	}{
		{"Pendiente", backlog.EstadoPendiente, juan, nil},
		{"En sprint", backlog.EstadoEnSprint, juan, nil},
		{"En progreso", backlog.EstadoEnProgreso, juan, nil},
		{"Hecha", backlog.EstadoHecho, juan, poker.ErrHistoriaHecha},
		{"la abre un Product Builder", backlog.EstadoPendiente, carolina, poker.ErrSoloAgileEnabler},
		{"Hecha y la abre un Product Builder: primero la historia", backlog.EstadoHecho, carolina, poker.ErrHistoriaHecha},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			h := backlog.Historia{Numero: 1, Titulo: "Login", Estado: c.estado}
			if err := poker.PuedeAbrir(h, c.quien); !errors.Is(err, c.err) {
				t.Errorf("error = %v, se esperaba %v", err, c.err)
			}
		})
	}
}

// RN1 · una historia que ya tiene story points se puede volver a estimar
func TestPuedeAbrir_HistoriaYaEstimada(t *testing.T) {
	sp := 5
	h := backlog.Historia{Numero: 1, Estado: backlog.EstadoPendiente, StoryPoints: &sp}
	if err := poker.PuedeAbrir(h, juan); err != nil {
		t.Errorf("no se esperaba error: %v", err)
	}
}

// CA-17.3 y RN5 · participan los integrantes activos, en el orden en que se registraron
func TestParticipantes(t *testing.T) {
	got := poker.Participantes([]member.Integrante{juan, lucia, carolina, mariano})

	if len(got) != 3 || got[0].ID != juan.ID || got[1].ID != carolina.ID || got[2].ID != mariano.ID {
		t.Errorf("participantes = %+v, se esperaban Juan, Carolina y Mariano", got)
	}
}

func TestParticipantes_SinActivos(t *testing.T) {
	if got := poker.Participantes([]member.Integrante{lucia}); len(got) != 0 {
		t.Errorf("participantes = %+v, no se esperaba ninguno", got)
	}
}

// Los mensajes de los errores son los de la spec (sección 7)
func TestErroresDeAbrir_Mensajes(t *testing.T) {
	casos := map[error]string{
		poker.ErrHistoriaHecha:    "la historia está Hecha: ya no se estima",
		poker.ErrSesionYaAbierta:  "la historia ya tiene una sesión de Planning Poker abierta",
		poker.ErrSoloAgileEnabler: "solo el Agile Enabler del proyecto abre una sesión de Planning Poker",
	}
	for err, quiero := range casos {
		if err.Error() != quiero {
			t.Errorf("mensaje = %q, se esperaba %q", err.Error(), quiero)
		}
	}
}
