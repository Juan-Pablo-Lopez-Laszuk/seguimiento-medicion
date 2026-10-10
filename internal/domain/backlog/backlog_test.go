package backlog_test

import (
	"errors"
	"strconv"
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

// CA-05.2 · la prioridad es Alta, Media o Baja
func TestNueva_PrioridadesValidas(t *testing.T) {
	for _, p := range []backlog.Prioridad{backlog.PrioridadAlta, backlog.PrioridadMedia, backlog.PrioridadBaja} {
		d := datosValidos()
		d.Prioridad = p

		h, err := backlog.Nueva(d, 7, nil)
		if err != nil {
			t.Fatalf("prioridad %q: no se esperaba error: %v", p, err)
		}

		if h.Prioridad != p {
			t.Errorf("prioridad = %q, se esperaba %q", h.Prioridad, p)
		}
	}
}

// RN3 · sin prioridad, o con una que no existe, es un error: no se elige una por el usuario
func TestNueva_PrioridadInvalida(t *testing.T) {
	for _, p := range []backlog.Prioridad{"", "Urgente", "alta", "Alta "} {
		d := datosValidos()
		d.Prioridad = p

		_, err := backlog.Nueva(d, 7, nil)

		verificarErrorDeCampo(t, err, backlog.CampoPrioridad, backlog.ErrPrioridadInvalida)
	}
}

// El formulario ofrece las prioridades en este orden, de la más importante a la menos
func TestPrioridades_EnElOrdenDelFormulario(t *testing.T) {
	esperadas := []backlog.Prioridad{backlog.PrioridadAlta, backlog.PrioridadMedia, backlog.PrioridadBaja}
	if len(backlog.Prioridades) != len(esperadas) {
		t.Fatalf("Prioridades = %v, se esperaba %v", backlog.Prioridades, esperadas)
	}
	for i, p := range esperadas {
		if backlog.Prioridades[i] != p {
			t.Errorf("Prioridades[%d] = %q, se esperaba %q", i, backlog.Prioridades[i], p)
		}
	}
}

// CA-05.3 y RN4 · los story points son opcionales: vacío es "sin estimar", que no es lo mismo que 0
func TestNueva_StoryPointsVacios_QuedaSinEstimar(t *testing.T) {
	for _, texto := range []string{"", "   "} {
		d := datosValidos()
		d.StoryPoints = texto

		h, err := backlog.Nueva(d, 7, nil)
		if err != nil {
			t.Fatalf("story points %q: no se esperaba error: %v", texto, err)
		}

		if h.StoryPoints != nil {
			t.Errorf("story points %q: tendría que quedar sin estimar y quedó en %d", texto, *h.StoryPoints)
		}
	}
}

// RN4 · todos los valores de la escala valen, incluidos los extremos 0 y 21
func TestNueva_StoryPointsDeLaEscala(t *testing.T) {
	for _, puntos := range []int{0, 1, 2, 3, 5, 8, 13, 21} {
		d := datosValidos()
		d.StoryPoints = strconv.Itoa(puntos)

		h, err := backlog.Nueva(d, 7, nil)
		if err != nil {
			t.Fatalf("story points %d: no se esperaba error: %v", puntos, err)
		}

		if h.StoryPoints == nil || *h.StoryPoints != puntos {
			t.Errorf("story points %d: se guardó %v", puntos, h.StoryPoints)
		}
	}
}

// RN4 · los espacios alrededor del número no importan
func TestNueva_StoryPointsConEspaciosAlrededor(t *testing.T) {
	d := datosValidos()
	d.StoryPoints = " 8 "

	h, err := backlog.Nueva(d, 7, nil)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if h.StoryPoints == nil || *h.StoryPoints != 8 {
		t.Errorf("se esperaban 8 story points y se guardó %v", h.StoryPoints)
	}
}

// RN4 · fuera de la escala, o algo que no es un número entero, es un error
func TestNueva_StoryPointsInvalidos(t *testing.T) {
	for _, texto := range []string{"4", "22", "-1", "3,5", "3.5", "abc", "5 puntos", "99999999999999999999"} {
		d := datosValidos()
		d.StoryPoints = texto

		_, err := backlog.Nueva(d, 7, nil)

		verificarErrorDeCampo(t, err, backlog.CampoStoryPoints, backlog.ErrStoryPointsInvalidos)
	}
}

// El formulario ofrece la escala de menor a mayor
func TestEscalaStoryPoints_EsLaAcordada(t *testing.T) {
	esperada := []int{0, 1, 2, 3, 5, 8, 13, 21}
	if len(backlog.EscalaStoryPoints) != len(esperada) {
		t.Fatalf("EscalaStoryPoints = %v, se esperaba %v", backlog.EscalaStoryPoints, esperada)
	}
	for i, v := range esperada {
		if backlog.EscalaStoryPoints[i] != v {
			t.Errorf("EscalaStoryPoints[%d] = %d, se esperaba %d", i, backlog.EscalaStoryPoints[i], v)
		}
	}
}
