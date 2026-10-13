package backlog_test

import (
	"strings"
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

// RN1 · la descripción es obligatoria: vacía o de solo espacios es un error del campo "descripcion"
func TestNuevoCriterio_DescripcionVaciaOSoloEspacios_EsObligatoria(t *testing.T) {
	for _, descripcion := range []string{"", "   ", "\t\n "} {
		_, err := backlog.NuevoCriterio(descripcion, historiaPendiente())

		verificarErrorDeCampo(t, err, backlog.CampoDescripcion, backlog.ErrDescripcionCriterioObligatoria)
	}
}

// RN1 · hasta 300 caracteres (no bytes) valen; 301 no
func TestNuevoCriterio_LargoDeLaDescripcion(t *testing.T) {
	casos := []struct {
		nombre      string
		descripcion string
		esperado    error
	}{
		{"300 caracteres", strings.Repeat("a", 300), nil},
		{"301 caracteres", strings.Repeat("a", 301), backlog.ErrDescripcionCriterioLarga},
		{"300 caracteres con tildes (más de 300 bytes)", strings.Repeat("ó", 300), nil},
		{"301 caracteres con tildes", strings.Repeat("ó", 301), backlog.ErrDescripcionCriterioLarga},
		{"300 caracteres con espacios sobrantes", "  " + strings.Repeat("a", 300) + "  ", nil},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := backlog.NuevoCriterio(c.descripcion, historiaPendiente())

			if c.esperado == nil {
				if err != nil {
					t.Errorf("no se esperaba error y se obtuvo %v", err)
				}
				return
			}
			verificarErrorDeCampo(t, err, backlog.CampoDescripcion, c.esperado)
		})
	}
}
