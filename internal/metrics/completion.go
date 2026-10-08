package metrics

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
	if p.Planificadas == 0 {
		return p
	}
	p.Valor = float64(p.Hechas) / float64(p.Planificadas) * 100
	return p
}
