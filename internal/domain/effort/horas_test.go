package effort_test

import (
	"errors"
	"math"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
)

// CA-23.1: las horas estimadas son mayores a 0 y aceptan decimales.
func TestValidarHoras_MayoresACeroConDecimales(t *testing.T) {
	for _, h := range []float64{8, 1.5, 0.25, 40} {
		if err := effort.ValidarHoras(h); err != nil {
			t.Errorf("ValidarHoras(%v) = %v, se esperaba que sea válida", h, err)
		}
	}
	for _, h := range []float64{0, -1, -0.5, math.NaN(), math.Inf(1)} {
		if err := effort.ValidarHoras(h); !errors.Is(err, effort.ErrHorasInvalidas) {
			t.Errorf("ValidarHoras(%v) = %v, se esperaba ErrHorasInvalidas", h, err)
		}
	}
	if got := effort.ErrHorasInvalidas.Error(); got != "las horas estimadas deben ser mayores a 0" {
		t.Errorf("mensaje del error = %q", got)
	}
}

// CA-23.1: en el formulario las horas se escriben con coma o con punto decimal ("1,5" o "1.5").
func TestParsearHoras_AceptaComaOPuntoDecimal(t *testing.T) {
	casos := []struct {
		texto  string
		quiero float64
	}{
		{"8", 8},
		{"1,5", 1.5},
		{"1.5", 1.5},
		{" 0,25 ", 0.25},
	}
	for _, c := range casos {
		got, err := effort.ParsearHoras(c.texto)
		if err != nil {
			t.Errorf("ParsearHoras(%q): error inesperado %v", c.texto, err)
			continue
		}
		if got != c.quiero {
			t.Errorf("ParsearHoras(%q) = %v, se esperaba %v", c.texto, got, c.quiero)
		}
	}
}

// Un texto vacío, que no es un número, o un número que no es mayor a 0 se rechaza con ErrHorasInvalidas.
func TestParsearHoras_RechazaTextosInvalidos(t *testing.T) {
	for _, texto := range []string{"", "   ", "abc", "1,5,2", "0", "-2", "NaN", "Inf"} {
		if _, err := effort.ParsearHoras(texto); !errors.Is(err, effort.ErrHorasInvalidas) {
			t.Errorf("ParsearHoras(%q) = %v, se esperaba ErrHorasInvalidas", texto, err)
		}
	}
}

// CA-23.2: si la historia tiene tareas, sus horas estimadas son la suma de las de sus tareas;
// si no tiene, se usan las horas cargadas en la historia.
func TestHorasEstimadasDeHistoria_SumaLasTareas(t *testing.T) {
	casos := []struct {
		nombre        string
		horasHistoria float64
		tareas        []effort.Tarea
		quiero        float64
	}{
		{"sin tareas: usa las horas de la historia", 8, nil, 8},
		{"con tareas: suma las tareas", 0, []effort.Tarea{{HorasEstimadas: 2}, {HorasEstimadas: 3.5}}, 5.5},
		{"con tareas: ignora las horas cargadas en la historia", 10, []effort.Tarea{{HorasEstimadas: 2}, {HorasEstimadas: 3}}, 5},
		{"sin tareas y sin estimar: da 0", 0, nil, 0},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			got, err := effort.HorasEstimadasDeHistoria(c.horasHistoria, c.tareas)
			if err != nil {
				t.Fatalf("error inesperado: %v", err)
			}
			if got != c.quiero {
				t.Errorf("horas = %v, se esperaba %v", got, c.quiero)
			}
		})
	}
}
