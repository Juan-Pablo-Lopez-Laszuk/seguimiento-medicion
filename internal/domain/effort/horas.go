package effort

import (
	"errors"
	"math"
)

// ErrHorasInvalidas se devuelve cuando las horas estimadas no son un número mayor a 0 (CA-23.1).
var ErrHorasInvalidas = errors.New("las horas estimadas deben ser mayores a 0")

// ValidarHoras controla que las horas estimadas sean un número finito mayor a 0. Acepta decimales.
func ValidarHoras(horas float64) error {
	if math.IsNaN(horas) || math.IsInf(horas, 0) || horas <= 0 {
		return ErrHorasInvalidas
	}
	return nil
}
