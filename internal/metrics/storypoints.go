package metrics

// Historia tiene los datos de una historia de usuario que hacen falta para calcular métricas.
type Historia struct {
	SP    float64 // story points estimados
	Hecha bool    // true si la historia está en estado Hecho
}

// StoryPoints es el resultado de las métricas M1 (planificados) y M2 (completados).
type StoryPoints struct {
	Planificados float64 // suma de SP de todas las historias
	Completados  float64 // suma de SP de las historias Hechas
}

// CalcularStoryPoints suma los story points planificados y completados de las historias de un sprint.
func CalcularStoryPoints(historias []Historia) (StoryPoints, error) {
	var sp StoryPoints
	for _, h := range historias {
		sp.Planificados += h.SP
		if h.Hecha {
			sp.Completados += h.SP
		}
	}
	return sp, nil
}
