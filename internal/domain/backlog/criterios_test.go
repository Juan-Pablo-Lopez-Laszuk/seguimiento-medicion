package backlog_test

import (
	"errors"
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

// historiaHecha es la misma historia de ejemplo, ya Hecha.
func historiaHecha() backlog.Historia {
	h := historiaPendiente()
	h.Estado = backlog.EstadoHecho
	return h
}

// RN7 · una historia Hecha no admite criterios nuevos: agregarlos rompería que todos están cumplidos
func TestNuevoCriterio_HistoriaHecha_NoAdmiteCriteriosNuevos(t *testing.T) {
	_, err := backlog.NuevoCriterio("Un criterio más", historiaHecha())

	if !errors.Is(err, backlog.ErrHistoriaHecha) {
		t.Errorf("se esperaba backlog.ErrHistoriaHecha y se obtuvo: %v", err)
	}
}

// RN7 · el estado de la historia se revisa antes que la descripción: no tiene sentido corregir un texto que igual
// no se va a poder guardar
func TestNuevoCriterio_HistoriaHechaConDescripcionInvalida_InformaQueEstaHecha(t *testing.T) {
	_, err := backlog.NuevoCriterio("   ", historiaHecha())

	if !errors.Is(err, backlog.ErrHistoriaHecha) {
		t.Errorf("se esperaba backlog.ErrHistoriaHecha y se obtuvo: %v", err)
	}
}

// RN7 · en cualquier otro estado se pueden agregar criterios
func TestNuevoCriterio_HistoriaEnSprintOEnProgreso_AdmiteCriterios(t *testing.T) {
	for _, estado := range []backlog.Estado{backlog.EstadoPendiente, "En sprint", "En progreso"} {
		h := historiaPendiente()
		h.Estado = estado

		if _, err := backlog.NuevoCriterio("Un criterio", h); err != nil {
			t.Errorf("estado %q: no se esperaba error y se obtuvo %v", estado, err)
		}
	}
}

// CA-07.2 y RN3 · un criterio se marca cumplido o no cumplido de forma explícita, y el original no cambia
func TestMarcar_PasaDeUnEstadoAlOtro(t *testing.T) {
	original := backlog.Criterio{ID: 9, HistoriaID: 4, Descripcion: "Entra con su cuenta"}

	cumplido, err := original.Marcar(true, historiaPendiente())
	if err != nil {
		t.Fatalf("no se esperaba error al marcar cumplido: %v", err)
	}
	if !cumplido.Cumplido || cumplido.ID != 9 || cumplido.Descripcion != "Entra con su cuenta" {
		t.Errorf("se esperaba el mismo criterio ya cumplido y se obtuvo %+v", cumplido)
	}
	if original.Cumplido {
		t.Error("marcar no tiene que modificar el criterio original")
	}

	desmarcado, err := cumplido.Marcar(false, historiaPendiente())
	if err != nil {
		t.Fatalf("no se esperaba error al marcar no cumplido: %v", err)
	}
	if desmarcado.Cumplido {
		t.Error("el criterio tenía que quedar no cumplido")
	}
}

// RN3 · marcar un criterio que ya está en ese estado no es un error: queda igual
func TestMarcar_ElMismoEstadoDosVeces_QuedaIgual(t *testing.T) {
	c := backlog.Criterio{ID: 9, HistoriaID: 4, Descripcion: "Entra con su cuenta", Cumplido: true}

	otraVez, err := c.Marcar(true, historiaPendiente())

	if err != nil || !otraVez.Cumplido {
		t.Errorf("se esperaba que siga cumplido sin error y se obtuvo %+v, %v", otraVez, err)
	}
	sinCumplir := backlog.Criterio{ID: 10, HistoriaID: 4, Descripcion: "Muestra su nombre"}
	if sigue, err := sinCumplir.Marcar(false, historiaPendiente()); err != nil || sigue.Cumplido {
		t.Errorf("se esperaba que siga no cumplido sin error y se obtuvo %+v, %v", sigue, err)
	}
}

// RN7 · en una historia Hecha no se desmarca ningún criterio
func TestMarcar_HistoriaHecha_NoSeDesmarca(t *testing.T) {
	c := backlog.Criterio{ID: 9, HistoriaID: 4, Descripcion: "Entra con su cuenta", Cumplido: true}

	_, err := c.Marcar(false, historiaHecha())

	if !errors.Is(err, backlog.ErrHistoriaHecha) {
		t.Errorf("se esperaba backlog.ErrHistoriaHecha y se obtuvo: %v", err)
	}
}

// RN7 · marcar como cumplido en una historia Hecha no cambia nada (todos ya están cumplidos), así que no es un error
func TestMarcar_HistoriaHecha_MarcarCumplidoNoCambiaNada(t *testing.T) {
	c := backlog.Criterio{ID: 9, HistoriaID: 4, Descripcion: "Entra con su cuenta", Cumplido: true}

	otraVez, err := c.Marcar(true, historiaHecha())

	if err != nil || !otraVez.Cumplido {
		t.Errorf("se esperaba que siga cumplido sin error y se obtuvo %+v, %v", otraVez, err)
	}
}
