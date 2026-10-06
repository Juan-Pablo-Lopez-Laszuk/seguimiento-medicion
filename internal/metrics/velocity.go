package metrics

import "sort"

// EstadoSprint es el estado de un sprint: Planificado, Activo o Cerrado.
type EstadoSprint string

// Estados posibles de un sprint (docs/metricas.md, sección 3).
const (
	SprintPlanificado EstadoSprint = "Planificado"
	SprintActivo      EstadoSprint = "Activo"
	SprintCerrado     EstadoSprint = "Cerrado"
)

// Sprint tiene los datos de un sprint que hacen falta para calcular métricas.
type Sprint struct {
	Numero        int          // número de sprint dentro del proyecto (1, 2, 3...)
	Estado        EstadoSprint // Planificado, Activo o Cerrado
	SPCompletados float64      // story points de las historias Hechas
}

// Velocidad es el resultado del cálculo de velocidad del equipo (métrica M3).
type Velocidad struct {
	Valor float64 // promedio de story points completados por sprint
}

// CalcularVelocidad devuelve el promedio de story points completados de los últimos `ventana` sprints.
// Si hay menos sprints que la ventana, promedia los que haya.
func CalcularVelocidad(sprints []Sprint, ventana int) (Velocidad, error) {
	ordenados := make([]Sprint, len(sprints))
	copy(ordenados, sprints)
	sort.Slice(ordenados, func(i, j int) bool { return ordenados[i].Numero < ordenados[j].Numero })

	if len(ordenados) > ventana {
		ordenados = ordenados[len(ordenados)-ventana:]
	}

	suma := 0.0
	for _, s := range ordenados {
		suma += s.SPCompletados
	}
	return Velocidad{Valor: suma / float64(len(ordenados))}, nil
}
