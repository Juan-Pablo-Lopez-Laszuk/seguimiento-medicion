package backlog_test

import (
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
)

// historiaEn devuelve la historia de ejemplo en el estado indicado.
func historiaEn(estado backlog.Estado) backlog.Historia {
	h := historiaPendiente()
	h.Estado = estado
	return h
}

// progresoCompleto es el de una historia con 2 criterios, los 2 cumplidos: nada la frena para pasar a Hecho.
func progresoCompleto() backlog.Progreso {
	return backlog.ProgresoDe(criteriosConCumplidos(2, 2))
}

// RN1 · los estados, en el orden del camino de una historia
func TestEstados_EnElOrdenDelCamino(t *testing.T) {
	esperados := []backlog.Estado{backlog.EstadoPendiente, backlog.EstadoEnSprint, backlog.EstadoEnProgreso, backlog.EstadoHecho}
	if len(backlog.Estados) != len(esperados) {
		t.Fatalf("Estados = %v, se esperaba %v", backlog.Estados, esperados)
	}
	for i, e := range esperados {
		if backlog.Estados[i] != e {
			t.Errorf("Estados[%d] = %q, se esperaba %q", i, backlog.Estados[i], e)
		}
	}
	// los textos son los que guarda la base (CHECK de historias.estado)
	if backlog.EstadoEnSprint != "En sprint" || backlog.EstadoEnProgreso != "En progreso" {
		t.Errorf("los textos de los estados no son los de la base: %q y %q", backlog.EstadoEnSprint, backlog.EstadoEnProgreso)
	}
}

// CA-08.1 y RN2 · las transiciones válidas; la historia nueva queda en el estado pedido y la original no cambia
func TestCambiarEstado_TransicionesValidas(t *testing.T) {
	casos := []struct{ desde, hacia backlog.Estado }{
		{backlog.EstadoPendiente, backlog.EstadoEnSprint},
		{backlog.EstadoEnSprint, backlog.EstadoEnProgreso},
		{backlog.EstadoEnProgreso, backlog.EstadoHecho},
		{backlog.EstadoEnSprint, backlog.EstadoPendiente},
		{backlog.EstadoEnProgreso, backlog.EstadoPendiente},
	}
	for _, c := range casos {
		original := historiaEn(c.desde)

		nueva, err := original.CambiarEstado(c.hacia, progresoCompleto())
		if err != nil {
			t.Errorf("%s → %s: no se esperaba error: %v", c.desde, c.hacia, err)
			continue
		}

		if nueva.Estado != c.hacia {
			t.Errorf("%s → %s: la historia quedó en %q", c.desde, c.hacia, nueva.Estado)
		}
		if original.Estado != c.desde {
			t.Errorf("%s → %s: la historia original cambió a %q", c.desde, c.hacia, original.Estado)
		}
		if nueva.ID != original.ID || nueva.Titulo != original.Titulo || nueva.Numero != original.Numero {
			t.Errorf("%s → %s: cambió algo más que el estado: %+v", c.desde, c.hacia, nueva)
		}
	}
}
