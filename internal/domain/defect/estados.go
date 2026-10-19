package defect

import (
	"errors"
	"fmt"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
)

// Errores de un cambio de estado (sección 7 de la spec de HU-28).
var (
	ErrTransicionInvalida       = errors.New("transición de estado no permitida")
	ErrEstadoInvalido           = errors.New("el estado tiene que ser En progreso, Resuelto, Cerrado o Reabierto")
	ErrSprintResolucionInvalido = errors.New("el sprint de resolución tiene que ser un sprint del proyecto")
	ErrSprintResolucionAnterior = errors.New("el sprint de resolución no puede ser anterior al de detección")
	// ErrDefectoNoEncontrado lo devuelven los repositorios cuando no hay un defecto con ese ID en el proyecto.
	ErrDefectoNoEncontrado = errors.New("no existe el defecto")
)

// transiciones son los cambios de estado permitidos (RN1 de HU-28), en el orden en que se ofrecen.
var transiciones = map[Estado][]Estado{
	EstadoAbierto:    {EstadoEnProgreso},
	EstadoEnProgreso: {EstadoResuelto},
	EstadoResuelto:   {EstadoCerrado, EstadoReabierto},
	EstadoReabierto:  {EstadoEnProgreso},
}

// Siguientes devuelve los estados a los que puede pasar un defecto desde el estado dado. Cerrado es final.
func Siguientes(e Estado) []Estado {
	return transiciones[e]
}

// errTransicion dice de qué estado a qué estado no se puede pasar, y se reconoce con errors.Is(err, ErrTransicionInvalida).
type errTransicion struct{ de, a Estado }

func (e errTransicion) Error() string {
	return fmt.Sprintf("un defecto en estado %s no puede pasar a %s", e.de, e.a)
}

func (e errTransicion) Is(objetivo error) bool { return objetivo == ErrTransicionInvalida }

// CambiarEstado devuelve el defecto en el estado nuevo; el original no cambia. Al pasar a Resuelto, sprintResolucion
// tiene que ser un sprint del proyecto (de sprints) no anterior al de detección (CA-28.2); al reabrir se borra el
// sprint de resolución (CA-28.3). En las demás transiciones sprintResolucion no se usa.
func (d Defecto) CambiarEstado(nuevo Estado, sprintResolucion int64, sprints []sprint.Sprint) (Defecto, error) {
	if nuevo != EstadoAbierto && !esEstado(nuevo) {
		return Defecto{}, ErrEstadoInvalido
	}
	if !permitida(d.Estado, nuevo) {
		return Defecto{}, errTransicion{de: d.Estado, a: nuevo}
	}
	switch nuevo {
	case EstadoResuelto:
		resolucion, ok := sprintDelProyecto(sprintResolucion, d.ProyectoID, sprints)
		if !ok {
			return Defecto{}, ErrSprintResolucionInvalido
		}
		if deteccion, ok := sprintDelProyecto(d.SprintDeteccionID, d.ProyectoID, sprints); ok && resolucion.Numero < deteccion.Numero {
			return Defecto{}, ErrSprintResolucionAnterior
		}
		d.SprintResolucionID = resolucion.ID
	case EstadoReabierto:
		d.SprintResolucionID = 0
	}
	d.Estado = nuevo
	return d, nil
}

func esEstado(e Estado) bool {
	for _, siguientes := range transiciones {
		for _, s := range siguientes {
			if s == e {
				return true
			}
		}
	}
	return false
}

func permitida(de, a Estado) bool {
	for _, s := range transiciones[de] {
		if s == a {
			return true
		}
	}
	return false
}

func sprintDelProyecto(id, proyectoID int64, sprints []sprint.Sprint) (sprint.Sprint, bool) {
	for _, s := range sprints {
		if s.ID == id && s.ProyectoID == proyectoID {
			return s, true
		}
	}
	return sprint.Sprint{}, false
}
