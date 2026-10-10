package backlog_test

import (
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
)

// datosValidos devuelve datos que cumplen todas las reglas; cada test cambia solo lo que quiere probar.
func datosValidos() backlog.Datos {
	return backlog.Datos{
		Titulo:      "Iniciar sesión con Google",
		Descripcion: "Para no tener que recordar otra contraseña.",
		Prioridad:   backlog.PrioridadAlta,
	}
}

// CA-05.1 y CA-05.3 · la primera historia del proyecto es la HU-1 y nace Pendiente
func TestNueva_PrimeraHistoria_EsLaHU1YNacePendiente(t *testing.T) {
	h, err := backlog.Nueva(datosValidos(), 7, nil)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if h.Numero != 1 || h.Identificador() != "HU-1" {
		t.Errorf("se esperaba la HU-1 y se obtuvo número %d (%s)", h.Numero, h.Identificador())
	}
	if h.Estado != backlog.EstadoPendiente {
		t.Errorf("estado = %q, se esperaba %q", h.Estado, backlog.EstadoPendiente)
	}
	if h.ProyectoID != 7 {
		t.Errorf("proyecto = %d, se esperaba 7", h.ProyectoID)
	}
	if h.Titulo != "Iniciar sesión con Google" || h.Prioridad != backlog.PrioridadAlta {
		t.Errorf("se esperaba el título y la prioridad cargados y se obtuvo %+v", h)
	}
	if h.StoryPoints != nil {
		t.Errorf("sin story points cargados, la historia tiene que quedar sin estimar y quedó con %d", *h.StoryPoints)
	}
}
