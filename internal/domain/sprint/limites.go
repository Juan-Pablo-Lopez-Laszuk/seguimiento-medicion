package sprint

import "github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"

// Limites devuelve cuándo empieza el primer sprint (el de menor número) y cuándo termina el último (el que termina
// más tarde), o el valor cero si no hay sprints. Es lo que necesita project.Modificar (HU-02 RN3 y RN4).
func Limites(sprints []Sprint) project.LimitesSprints {
	ordenados := porNumero(sprints)
	ultimo, hay := ultimoSprint(sprints)
	if !hay {
		return project.LimitesSprints{}
	}
	return project.LimitesSprints{
		PrimerNumero: ordenados[0].Numero, PrimerInicio: ordenados[0].FechaInicio,
		UltimoNumero: ultimo.Numero, UltimoFin: ultimo.FechaFin,
	}
}
