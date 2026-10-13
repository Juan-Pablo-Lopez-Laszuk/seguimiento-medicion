package backlog_test

import (
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
)

// historiaPendiente devuelve la historia de ejemplo de la spec de HU-07: la HU-1, Pendiente.
func historiaPendiente() backlog.Historia {
	return backlog.Historia{
		ID:         4,
		ProyectoID: 7,
		Numero:     1,
		Titulo:     "Iniciar sesión con Google",
		Prioridad:  backlog.PrioridadAlta,
		Estado:     backlog.EstadoPendiente,
	}
}

// CA-07.2 · todo criterio nace no cumplido, es de la historia y se guarda sin los espacios sobrantes
func TestNuevoCriterio_NaceNoCumplidoYSinEspaciosSobrantes(t *testing.T) {
	c, err := backlog.NuevoCriterio("  Entra con su cuenta de Google \n", historiaPendiente())
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if c.Descripcion != "Entra con su cuenta de Google" {
		t.Errorf("descripción = %q, se esperaba %q", c.Descripcion, "Entra con su cuenta de Google")
	}
	if c.Cumplido {
		t.Error("un criterio nuevo tiene que nacer no cumplido")
	}
	if c.HistoriaID != 4 {
		t.Errorf("historia = %d, se esperaba 4", c.HistoriaID)
	}
}
