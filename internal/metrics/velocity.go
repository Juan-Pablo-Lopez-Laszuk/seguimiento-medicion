package metrics

import (
	"errors"
	"sort"
)

// Errores que puede devolver el cálculo de velocidad.
var (
	ErrVentanaInvalida = errors.New("la ventana de sprints debe ser mayor a cero")
	ErrSPNegativos     = errors.New("los story points no pueden ser negativos")
)

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

// VentanaPorDefecto es la cantidad de sprints cerrados que se promedian si el proyecto no configura otra.
const VentanaPorDefecto = 3

// mensajeSinSprintsCerrados es el aviso que se muestra cuando todavía no hay datos para calcular.
const mensajeSinSprintsCerrados = "Aún no hay sprints cerrados"

// Velocidad es el resultado del cálculo de velocidad del equipo (métrica M3).
type Velocidad struct {
	Valor         float64 // promedio de story points completados por sprint
	SprintsUsados int     // cantidad de sprints cerrados que entraron en el promedio
}

// Mensaje devuelve el aviso para el usuario cuando no hay datos para calcular, o "" si los hay.
func (v Velocidad) Mensaje() string {
	if v.SprintsUsados == 0 {
		return mensajeSinSprintsCerrados
	}
	return ""
}

// CalcularVelocidad devuelve el promedio de story points completados de los últimos `ventana` sprints
// cerrados (specs/HU-31-velocidad.md):
//
//   - RN1: solo cuentan los sprints cerrados; los activos y planificados se ignoran.
//   - RN2: se toman los últimos, es decir, los de mayor número de sprint.
//   - RN3: se divide por la cantidad de sprints que se tomaron (si hay menos que la ventana, los que haya).
//   - RN4: el orden en que llegan los sprints no cambia el resultado.
//
// Sin sprints cerrados devuelve 0 y Mensaje() avisa; nunca se divide por cero.
func CalcularVelocidad(sprints []Sprint, ventana int) (Velocidad, error) {
	if ventana <= 0 {
		return Velocidad{}, ErrVentanaInvalida
	}
	for _, s := range sprints {
		if s.SPCompletados < 0 {
			return Velocidad{}, ErrSPNegativos
		}
	}

	ultimos := ultimosCerrados(sprints, ventana)
	if len(ultimos) == 0 {
		return Velocidad{}, nil
	}

	suma := 0.0
	for _, s := range ultimos {
		suma += s.SPCompletados
	}
	return Velocidad{Valor: suma / float64(len(ultimos)), SprintsUsados: len(ultimos)}, nil
}

// ultimosCerrados devuelve, ordenados por número, los últimos `ventana` sprints cerrados.
// No modifica la lista que recibe.
func ultimosCerrados(sprints []Sprint, ventana int) []Sprint {
	var cerrados []Sprint
	for _, s := range sprints {
		if s.Estado == SprintCerrado {
			cerrados = append(cerrados, s)
		}
	}
	sort.Slice(cerrados, func(i, j int) bool { return cerrados[i].Numero < cerrados[j].Numero })

	if len(cerrados) > ventana {
		cerrados = cerrados[len(cerrados)-ventana:]
	}
	return cerrados
}
