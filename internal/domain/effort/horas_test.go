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
