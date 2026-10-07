package metrics

import "fmt"

// Historia tiene los datos de una historia de usuario que hacen falta para calcular métricas.
type Historia struct {
	SP         float64 // story points estimados (no se usan si SinEstimar es true)
	Hecha      bool    // true si la historia está en estado Hecho
	SinEstimar bool    // true si todavía no tiene estimación
}

// StoryPoints es el resultado de las métricas M1 (planificados) y M2 (completados).
type StoryPoints struct {
	Planificados float64 // suma de SP de todas las historias
	Completados  float64 // suma de SP de las historias Hechas
	SinEstimar   int     // cantidad de historias sin estimar (suman 0)
}

// Advertencia devuelve el aviso para el usuario cuando hay historias sin estimar, o "" si no hay.
func (s StoryPoints) Advertencia() string {
	switch s.SinEstimar {
	case 0:
		return ""
	case 1:
		return "hay 1 historia sin estimar"
	default:
		return fmt.Sprintf("hay %d historias sin estimar", s.SinEstimar)
	}
}

// CalcularStoryPoints suma los story points planificados y completados de las historias de un sprint.
// Las historias sin estimar suman 0 y se cuentan aparte.
func CalcularStoryPoints(historias []Historia) (StoryPoints, error) {
	var sp StoryPoints
	for _, h := range historias {
		if h.SinEstimar {
			sp.SinEstimar++
			continue
		}
		sp.Planificados += h.SP
		if h.Hecha {
			sp.Completados += h.SP
		}
	}
	return sp, nil
}
