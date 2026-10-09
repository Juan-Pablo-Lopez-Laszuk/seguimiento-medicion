package metrics

import (
	"fmt"
	"strings"
)

// Porcentaje es el resultado de la métrica M7: qué parte de las historias planificadas se terminó.
type Porcentaje struct {
	Hechas       int     // historias en estado Hecho
	Planificadas int     // historias del sprint (o del proyecto)
	Valor        float64 // Hechas / Planificadas × 100
}

// CalcularPorcentajeCompletadas cuenta las historias Hechas de un sprint sobre sus historias planificadas.
// Un sprint sin historias da 0 %, sin error: nunca se divide por cero.
func CalcularPorcentajeCompletadas(historias []Historia) Porcentaje {
	p := Porcentaje{Planificadas: len(historias)}
	for _, h := range historias {
		if h.Hecha {
			p.Hechas++
		}
	}
	p.Valor = porcentaje(p.Hechas, p.Planificadas)
	return p
}

// SumarPorcentajes devuelve el porcentaje del proyecto: suma las historias Hechas y planificadas de todos
// sus sprints y recién ahí calcula el porcentaje (no promedia los porcentajes de cada sprint).
func SumarPorcentajes(porSprint []Porcentaje) Porcentaje {
	var total Porcentaje
	for _, p := range porSprint {
		total.Hechas += p.Hechas
		total.Planificadas += p.Planificadas
	}
	total.Valor = porcentaje(total.Hechas, total.Planificadas)
	return total
}

// Texto devuelve el porcentaje como se muestra en pantalla: con 1 decimal y coma decimal ("75,0 %").
func (p Porcentaje) Texto() string {
	return strings.Replace(fmt.Sprintf("%.1f %%", p.Valor), ".", ",", 1)
}

// porcentaje calcula hechas / planificadas × 100 y devuelve 0 si no hay historias (CA-32.2).
func porcentaje(hechas, planificadas int) float64 {
	if planificadas == 0 {
		return 0
	}
	return float64(hechas) / float64(planificadas) * 100
}
