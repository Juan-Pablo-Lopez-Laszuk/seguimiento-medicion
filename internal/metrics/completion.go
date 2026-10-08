package metrics

// Porcentaje es el resultado de la métrica M7: qué parte de las historias planificadas se terminó.
type Porcentaje struct {
	Hechas       int     // historias en estado Hecho
	Planificadas int     // historias del sprint (o del proyecto)
	Valor        float64 // Hechas / Planificadas × 100
}

// CalcularPorcentajeCompletadas cuenta las historias Hechas de un sprint sobre sus historias planificadas.
func CalcularPorcentajeCompletadas(historias []Historia) Porcentaje {
	p := Porcentaje{Planificadas: len(historias)}
	for _, h := range historias {
		if h.Hecha {
			p.Hechas++
		}
	}
	p.Valor = float64(p.Hechas) / float64(p.Planificadas) * 100
	return p
}
