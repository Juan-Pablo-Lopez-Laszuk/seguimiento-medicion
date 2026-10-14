package backlog_test

import (
	"errors"
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

// CA-08.1 y CA-08.3 · las transiciones que no están en la tabla son un error descriptivo: dice de qué estado a cuál
// se quiso pasar y a cuáles se puede pasar desde ahí
func TestCambiarEstado_TransicionesInvalidas_DanUnErrorDescriptivo(t *testing.T) {
	casos := []struct {
		desde, hacia backlog.Estado
		mensaje      string
	}{
		{backlog.EstadoPendiente, backlog.EstadoEnProgreso, "no se puede pasar de Pendiente a En progreso: desde Pendiente solo se puede pasar a En sprint"},
		{backlog.EstadoPendiente, backlog.EstadoHecho, "no se puede pasar de Pendiente a Hecho: desde Pendiente solo se puede pasar a En sprint"},
		{backlog.EstadoEnSprint, backlog.EstadoHecho, "no se puede pasar de En sprint a Hecho: desde En sprint solo se puede pasar a En progreso o Pendiente"},
		{backlog.EstadoEnProgreso, backlog.EstadoEnSprint, "no se puede pasar de En progreso a En sprint: desde En progreso solo se puede pasar a Hecho o Pendiente"},
		{backlog.EstadoHecho, backlog.EstadoEnProgreso, "no se puede pasar de Hecho a En progreso: una historia Hecha no cambia de estado"},
		{backlog.EstadoHecho, backlog.EstadoEnSprint, "no se puede pasar de Hecho a En sprint: una historia Hecha no cambia de estado"},
		{backlog.EstadoHecho, backlog.EstadoPendiente, "no se puede pasar de Hecho a Pendiente: una historia Hecha no cambia de estado"},
	}
	for _, c := range casos {
		_, err := historiaEn(c.desde).CambiarEstado(c.hacia, progresoCompleto())

		if !errors.Is(err, backlog.ErrTransicionInvalida) {
			t.Errorf("%s → %s: se esperaba backlog.ErrTransicionInvalida y se obtuvo: %v", c.desde, c.hacia, err)
			continue
		}
		if err.Error() != c.mensaje {
			t.Errorf("%s → %s: mensaje = %q, se esperaba %q", c.desde, c.hacia, err.Error(), c.mensaje)
		}
	}
}

// RN5 · pedir el estado en el que ya está la historia es un error, no un cambio que no hace nada
func TestCambiarEstado_ElMismoEstado_EsUnError(t *testing.T) {
	for _, estado := range backlog.Estados {
		_, err := historiaEn(estado).CambiarEstado(estado, progresoCompleto())

		if !errors.Is(err, backlog.ErrYaEstaEnEseEstado) {
			t.Errorf("%s → %s: se esperaba backlog.ErrYaEstaEnEseEstado y se obtuvo: %v", estado, estado, err)
			continue
		}
		if esperado := "la historia ya está " + string(estado); err.Error() != esperado {
			t.Errorf("mensaje = %q, se esperaba %q", err.Error(), esperado)
		}
	}
}

// RN1 · un estado que no existe es un error que lista los que sí existen
func TestCambiarEstado_UnEstadoQueNoExiste_EsUnError(t *testing.T) {
	for _, estado := range []backlog.Estado{"Terminada", "", "en sprint", "Hecha"} {
		_, err := historiaEn(backlog.EstadoEnSprint).CambiarEstado(estado, progresoCompleto())

		if !errors.Is(err, backlog.ErrEstadoDesconocido) {
			t.Errorf("estado %q: se esperaba backlog.ErrEstadoDesconocido y se obtuvo: %v", estado, err)
			continue
		}
		esperado := "el estado " + string(estado) + " no existe: los estados son Pendiente, En sprint, En progreso y Hecho"
		if err.Error() != esperado {
			t.Errorf("estado %q: mensaje = %q, se esperaba %q", estado, err.Error(), esperado)
		}
	}
}
