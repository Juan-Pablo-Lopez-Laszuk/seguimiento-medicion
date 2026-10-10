package backlog_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
)

// verificarErrorDeCampo comprueba que err sea un error de validación con el error esperado en ese campo.
func verificarErrorDeCampo(t *testing.T, err error, campo string, esperado error) {
	t.Helper()
	var errs domain.ErroresValidacion
	if !errors.As(err, &errs) {
		t.Fatalf("se esperaba domain.ErroresValidacion y se obtuvo: %v", err)
	}
	if !errors.Is(errs[campo], esperado) {
		t.Errorf("campo %q: se esperaba %v y se obtuvo %v (todos: %v)", campo, esperado, errs[campo], errs)
	}
}

// verificarSinErrorEnElCampo comprueba que, si hay errores de validación, ninguno sea del campo.
func verificarSinErrorEnElCampo(t *testing.T, err error, campo string) {
	t.Helper()
	var errs domain.ErroresValidacion
	if errors.As(err, &errs) && errs[campo] != nil {
		t.Errorf("campo %q: no se esperaba error y se obtuvo %v", campo, errs[campo])
	}
}

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

// CA-05.2 · el título es obligatorio: vacío o de solo espacios es un error
func TestNueva_TituloVacioOSoloEspacios_EsObligatorio(t *testing.T) {
	for _, titulo := range []string{"", "   ", "\t\n "} {
		d := datosValidos()
		d.Titulo = titulo

		_, err := backlog.Nueva(d, 7, nil)

		verificarErrorDeCampo(t, err, backlog.CampoTitulo, backlog.ErrTituloObligatorio)
	}
}

// RN1 · el título se guarda sin los espacios del inicio y del final
func TestNueva_QuitaLosEspaciosAlrededorDelTitulo(t *testing.T) {
	d := datosValidos()
	d.Titulo = "  Login con Google \n"

	h, err := backlog.Nueva(d, 7, nil)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if h.Titulo != "Login con Google" {
		t.Errorf("título = %q, se esperaba %q", h.Titulo, "Login con Google")
	}
}

// RN1 · hasta 120 caracteres (no bytes) valen; 121 no
func TestNueva_LargoDelTitulo(t *testing.T) {
	casos := []struct {
		nombre   string
		titulo   string
		esperado error
	}{
		{"120 caracteres", strings.Repeat("a", 120), nil},
		{"121 caracteres", strings.Repeat("a", 121), backlog.ErrTituloLargo},
		{"120 caracteres con tildes (más de 120 bytes)", strings.Repeat("á", 120), nil},
		{"121 caracteres con tildes", strings.Repeat("á", 121), backlog.ErrTituloLargo},
		{"120 caracteres con espacios sobrantes", "  " + strings.Repeat("a", 120) + "  ", nil},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d := datosValidos()
			d.Titulo = c.titulo

			_, err := backlog.Nueva(d, 7, nil)

			if c.esperado == nil {
				if err != nil {
					t.Errorf("no se esperaba error y se obtuvo %v", err)
				}
				return
			}
			verificarErrorDeCampo(t, err, backlog.CampoTitulo, c.esperado)
		})
	}
}

// RN2 · la descripción es opcional y se guarda sin los espacios del inicio y del final
func TestNueva_Descripcion_EsOpcionalYSeGuardaSinEspaciosSobrantes(t *testing.T) {
	casos := []struct{ cargada, guardada string }{
		{"", ""},
		{"   ", ""},
		{"  Con Google \n", "Con Google"},
	}
	for _, c := range casos {
		d := datosValidos()
		d.Descripcion = c.cargada

		h, err := backlog.Nueva(d, 7, nil)
		if err != nil {
			t.Fatalf("descripción %q: no se esperaba error: %v", c.cargada, err)
		}

		if h.Descripcion != c.guardada {
			t.Errorf("descripción %q: se guardó %q y se esperaba %q", c.cargada, h.Descripcion, c.guardada)
		}
	}
}

// RN2 · hasta 2000 caracteres (no bytes) valen; 2001 no
func TestNueva_LargoDeLaDescripcion(t *testing.T) {
	casos := []struct {
		nombre      string
		descripcion string
		esperado    error
	}{
		{"2000 caracteres", strings.Repeat("a", 2000), nil},
		{"2001 caracteres", strings.Repeat("a", 2001), backlog.ErrDescripcionLarga},
		{"2000 caracteres con tildes (más de 2000 bytes)", strings.Repeat("ñ", 2000), nil},
		{"2001 caracteres con tildes", strings.Repeat("ñ", 2001), backlog.ErrDescripcionLarga},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d := datosValidos()
			d.Descripcion = c.descripcion

			_, err := backlog.Nueva(d, 7, nil)

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
