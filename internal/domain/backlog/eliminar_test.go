package backlog_test

import (
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
)

// RN7 · una historia Pendiente, En sprint o En progreso se puede eliminar si no es de un sprint cerrado
func TestPuedeEliminarse_PendienteEnSprintOEnProgreso_Vale(t *testing.T) {
	for _, estado := range []backlog.Estado{backlog.EstadoPendiente, backlog.EstadoEnSprint, backlog.EstadoEnProgreso} {
		if err := historiaEn(estado).PuedeEliminarse(false); err != nil {
			t.Errorf("estado %q: no se esperaba error: %v", estado, err)
		}
	}
}

// CA-06.2 y RN5 · una historia Hecha no se elimina
func TestPuedeEliminarse_HistoriaHecha_NoSeElimina(t *testing.T) {
	err := historiaHecha().PuedeEliminarse(false)

	if !errors.Is(err, backlog.ErrEliminarHistoriaHecha) {
		t.Fatalf("se esperaba ErrEliminarHistoriaHecha y se obtuvo: %v", err)
	}
	if err.Error() != "la historia está Hecha: no se puede eliminar" {
		t.Errorf("mensaje = %q", err.Error())
	}
}

// CA-06.2 y RN6 · una historia de un sprint cerrado no se elimina, aunque no esté Hecha
func TestPuedeEliminarse_DeUnSprintCerrado_NoSeElimina(t *testing.T) {
	for _, estado := range []backlog.Estado{backlog.EstadoPendiente, backlog.EstadoEnSprint, backlog.EstadoEnProgreso} {
		err := historiaEn(estado).PuedeEliminarse(true)

		if !errors.Is(err, backlog.ErrEliminarEnSprintCerrado) {
			t.Errorf("estado %q: se esperaba ErrEliminarEnSprintCerrado y se obtuvo: %v", estado, err)
		}
	}
	if msg := backlog.ErrEliminarEnSprintCerrado.Error(); msg != "la historia es de un sprint cerrado: no se puede eliminar" {
		t.Errorf("mensaje = %q", msg)
	}
}

// Si una historia es Hecha y además de un sprint cerrado, el motivo que se informa es que está Hecha
func TestPuedeEliminarse_HechaYDeUnSprintCerrado_InformaQueEstaHecha(t *testing.T) {
	err := historiaHecha().PuedeEliminarse(true)

	if !errors.Is(err, backlog.ErrEliminarHistoriaHecha) {
		t.Errorf("se esperaba ErrEliminarHistoriaHecha y se obtuvo: %v", err)
	}
}

// RN10 · el mensaje de una historia con registros asociados
func TestErrHistoriaConRegistros_Mensaje(t *testing.T) {
	esperado := "la historia tiene registros asociados (Planning Poker, esfuerzo o defectos): no se puede eliminar"
	if msg := backlog.ErrHistoriaConRegistros.Error(); msg != esperado {
		t.Errorf("mensaje = %q, se esperaba %q", msg, esperado)
	}
}
