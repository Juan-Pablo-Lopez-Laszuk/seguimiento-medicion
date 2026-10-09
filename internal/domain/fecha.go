package domain

import "time"

// SoloFecha deja el día de t a medianoche UTC, así las comparaciones entre fechas no dependen de la
// hora ni de la zona horaria. La fecha cero (campo vacío) se mantiene en cero.
func SoloFecha(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
