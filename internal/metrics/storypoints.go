package metrics

import "fmt"

// Avisos que se muestran cuando hay historias sin estimar (CA-30.2).
const (
	advertenciaUnaSinEstimar    = "hay 1 historia sin estimar"
	advertenciaVariasSinEstimar = "hay %d historias sin estimar"
)

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
		return advertenciaUnaSinEstimar
	default:
		return fmt.Sprintf(advertenciaVariasSinEstimar, s.SinEstimar)
	}
}

// CalcularStoryPoints calcula los story points de un sprint a partir de sus historias
// (specs/HU-30-sp-planificados-completados.md):
//
//   - RN1: planificados = suma de SP de todas las historias del sprint.
//   - RN2: completados = suma de SP de las historias Hechas.
//   - RN3: una historia sin estimar suma 0 en los dos y se cuenta en SinEstimar.
//
// Si alguna historia estimada tiene SP negativos devuelve ErrSPNegativos y ningún resultado.
func CalcularStoryPoints(historias []Historia) (StoryPoints, error) {
	var sp StoryPoints
	for _, h := range historias {
		if h.SinEstimar {
			sp.SinEstimar++
			continue
		}
		if h.SP < 0 {
			return StoryPoints{}, ErrSPNegativos
		}
		sp.Planificados += h.SP
		if h.Hecha {
			sp.Completados += h.SP
		}
	}
	return sp, nil
}

// SumarStoryPoints devuelve el total del proyecto: la suma de los story points de todos sus sprints (RN4).
func SumarStoryPoints(porSprint []StoryPoints) StoryPoints {
	var total StoryPoints
	for _, sp := range porSprint {
		total = total.mas(sp)
	}
	return total
}

// mas devuelve la suma de dos resultados, campo por campo.
func (s StoryPoints) mas(otro StoryPoints) StoryPoints {
	return StoryPoints{
		Planificados: s.Planificados + otro.Planificados,
		Completados:  s.Completados + otro.Completados,
		SinEstimar:   s.SinEstimar + otro.SinEstimar,
	}
}
