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

// Tarea tiene los datos de una tarea que hacen falta para sumar las horas estimadas.
type Tarea struct {
	HorasEstimadas float64 // 0 si la tarea todavía no está estimada
}

// HorasEstimadasDeHistoria devuelve las horas estimadas de una historia (CA-23.2): si tiene tareas, la
// suma de las horas de sus tareas; si no tiene, las horas cargadas en la historia.
func HorasEstimadasDeHistoria(horasHistoria float64, tareas []Tarea) (float64, error) {
	if len(tareas) == 0 {
		return horasHistoria, nil
	}
	total := 0.0
	for _, t := range tareas {
		total += t.HorasEstimadas
	}
	return total, nil
}
