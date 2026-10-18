package backlog_test

import (
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
)

// Sprints del proyecto de ejemplo de la spec de HU-12 (sección 6): el 1 cerrado, el 2 activo y el 3 futuro.
var (
	sprint1 = backlog.DatosSprint{ID: 11, Numero: 1, Cerrado: true}
	sprint2 = backlog.DatosSprint{ID: 12, Numero: 2}
	sprint3 = backlog.DatosSprint{ID: 13, Numero: 3}
	todos   = []backlog.DatosSprint{sprint1, sprint2, sprint3}
)

// historiaLista es una historia que se puede asignar: Pendiente, sin sprint y estimada en 5 SP.
func historiaLista() backlog.Historia {
	h := historiaPendiente()
	puntos := 5
	h.StoryPoints = &puntos
	return h
}

// conCriterios es el progreso de una historia con 2 criterios, ninguno cumplido: alcanza para asignarla.
func conCriterios() backlog.Progreso {
	return backlog.ProgresoDe(criteriosConCumplidos(2, 0))
}

// enSprint devuelve la historia ya asignada al sprint, como queda después de AsignarA.
func enSprint(h backlog.Historia, s backlog.DatosSprint, estado backlog.Estado) backlog.Historia {
	id := s.ID
	h.SprintID, h.Estado = &id, estado
	return h
}

// CA-12.1 · una historia lista pasa a En sprint y queda en ese sprint; la original no cambia
func TestAsignarA_HistoriaLista(t *testing.T) {
	original := historiaLista()

	asignada, err := original.AsignarA(sprint2, conCriterios(), todos)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if asignada.Estado != backlog.EstadoEnSprint {
		t.Errorf("estado = %q, se esperaba %q", asignada.Estado, backlog.EstadoEnSprint)
	}
	if asignada.SprintID == nil || *asignada.SprintID != sprint2.ID {
		t.Errorf("SprintID = %v, se esperaba %d", asignada.SprintID, sprint2.ID)
	}
	if original.Estado != backlog.EstadoPendiente || original.SprintID != nil {
		t.Errorf("la historia original cambió: %+v", original)
	}
}

// CA-12.1 · 0 story points es una estimación; RN2 · se puede planificar un sprint que todavía no empezó
func TestAsignarA_CeroStoryPointsYSprintFuturo(t *testing.T) {
	h := historiaLista()
	cero := 0
	h.StoryPoints = &cero

	asignada, err := h.AsignarA(sprint3, conCriterios(), todos)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if asignada.SprintID == nil || *asignada.SprintID != sprint3.ID {
		t.Errorf("SprintID = %v, se esperaba %d", asignada.SprintID, sprint3.ID)
	}
}

// Sección 7 · cada motivo por el que no se asigna, con su error y su mensaje
func TestAsignarA_Errores(t *testing.T) {
	sinEstimar := historiaLista()
	sinEstimar.Numero, sinEstimar.StoryPoints = 2, nil
	hecha := historiaLista()
	hecha.Numero, hecha.Estado = 5, backlog.EstadoHecho

	casos := []struct {
		nombre   string
		historia backlog.Historia
		destino  backlog.DatosSprint
		progreso backlog.Progreso
		err      error
		mensaje  string
	}{
		{"sin estimar", sinEstimar, sprint2, conCriterios(), backlog.ErrSinEstimar,
			"HU-2 no se puede asignar: no está estimada"},
		{"sin criterios", historiaLista(), sprint2, backlog.Progreso{}, backlog.ErrSinCriterios,
			"HU-1 no se puede asignar: no tiene criterios de aceptación"},
		{"hecha", hecha, sprint2, conCriterios(), backlog.ErrYaHecha,
			"HU-5 no se puede asignar: ya está Hecha"},
		{"en otro sprint abierto", enSprint(historiaLista(), sprint2, backlog.EstadoEnSprint), sprint3, conCriterios(),
			backlog.ErrEnOtroSprint, "HU-1 ya está en el Sprint 2"},
		{"ya en progreso en el mismo sprint", enSprint(historiaLista(), sprint2, backlog.EstadoEnProgreso), sprint2,
			conCriterios(), backlog.ErrEnOtroSprint, "HU-1 ya está en el Sprint 2"},
		{"sprint cerrado", historiaLista(), sprint1, conCriterios(), backlog.ErrSprintCerrado,
			"el Sprint 1 está cerrado: no se le pueden asignar historias"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := c.historia.AsignarA(c.destino, c.progreso, todos)
			if !errors.Is(err, c.err) {
				t.Fatalf("error = %v, se esperaba %v", err, c.err)
			}
			if err.Error() != c.mensaje {
				t.Errorf("mensaje = %q, se esperaba %q", err.Error(), c.mensaje)
			}
		})
	}
}

// RN6 · el cambio de estado pasa por la regla de HU-08: una historia En sprint sin sprint (un dato roto) no se asigna
func TestAsignarA_UsaLasTransicionesDeHU08(t *testing.T) {
	h := historiaLista()
	h.Estado = backlog.EstadoEnSprint

	if _, err := h.AsignarA(sprint2, conCriterios(), todos); !errors.Is(err, backlog.ErrYaEstaEnEseEstado) {
		t.Errorf("error = %v, se esperaba backlog.ErrYaEstaEnEseEstado", err)
	}
}

// RN4 · una historia En sprint se quita: vuelve a Pendiente y sin sprint; la original no cambia
func TestQuitarDelSprint_EnSprint(t *testing.T) {
	original := enSprint(historiaLista(), sprint2, backlog.EstadoEnSprint)

	quitada, err := original.QuitarDelSprint()
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if quitada.Estado != backlog.EstadoPendiente || quitada.SprintID != nil {
		t.Errorf("quedó en %q con sprint %v, se esperaba Pendiente y sin sprint", quitada.Estado, quitada.SprintID)
	}
	if original.SprintID == nil {
		t.Error("la historia original cambió")
	}
}

// RN4 · ya En progreso no se quita; una Pendiente o una Hecha tampoco
func TestQuitarDelSprint_Errores(t *testing.T) {
	casos := []struct {
		nombre   string
		historia backlog.Historia
		err      error
		mensaje  string
	}{
		{"en progreso", enSprint(historiaLista(), sprint2, backlog.EstadoEnProgreso), backlog.ErrYaEmpezada,
			"HU-1 ya está En progreso: no se puede quitar del sprint"},
		{"hecha", enSprint(historiaLista(), sprint2, backlog.EstadoHecho), backlog.ErrYaEmpezada,
			"HU-1 ya está Hecho: no se puede quitar del sprint"},
		{"pendiente", historiaLista(), backlog.ErrNoEstaEnUnSprint, "HU-1 no está en ningún sprint"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := c.historia.QuitarDelSprint()
			if !errors.Is(err, c.err) {
				t.Fatalf("error = %v, se esperaba %v", err, c.err)
			}
			if err.Error() != c.mensaje {
				t.Errorf("mensaje = %q, se esperaba %q", err.Error(), c.mensaje)
			}
		})
	}
}
