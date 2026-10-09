package effort

import (
	"errors"
	"math"
	"strconv"
	"strings"
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

// ParsearHoras convierte lo que escribe el usuario en horas. Acepta coma o punto decimal ("1,5" o
// "1.5") y espacios alrededor. Si el texto no es un número mayor a 0 devuelve ErrHorasInvalidas.
func ParsearHoras(texto string) (float64, error) {
	normalizado := strings.Replace(strings.TrimSpace(texto), ",", ".", 1)
	horas, err := strconv.ParseFloat(normalizado, 64)
	if err != nil {
		return 0, ErrHorasInvalidas
	}
	if err := ValidarHoras(horas); err != nil {
		return 0, err
	}
	return horas, nil
}
