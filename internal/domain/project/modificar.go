package project

import (
	"errors"
	"fmt"
	"time"
)

// Errores de modificar un proyecto (spec: specs/HU-02-modificar-proyecto.md, sección 7).
var (
	// ErrFinAntesDelUltimoSprint se informa con el último sprint: "... (el Sprint 2 termina el 18/10/2026)".
	ErrFinAntesDelUltimoSprint = errors.New("la fecha de finalización no puede ser anterior al fin del último sprint")
	// ErrInicioDespuesDelPrimerSprint se informa con el primer sprint: "... (el Sprint 1 empieza el 05/10/2026)".
	ErrInicioDespuesDelPrimerSprint = errors.New("la fecha de inicio no puede ser posterior al inicio del primer sprint")
	// ErrProyectoFinalizado no es de un campo: el proyecto entero no se puede modificar (CA-02.3).
	ErrProyectoFinalizado = errors.New("un proyecto Finalizado no se puede modificar")
)

// formatoFecha es como se muestran las fechas en los mensajes: 05/10/2026.
const formatoFecha = "02/01/2006"

// LimitesSprints es lo que el proyecto necesita saber de sus sprints para modificarse: cuándo empieza el primero y
// cuándo termina el último. Lo arma sprint.Limites (este paquete no puede importar a sprint, que ya lo importa a él).
// El valor cero significa que el proyecto no tiene sprints.
type LimitesSprints struct {
	PrimerNumero int
	PrimerInicio time.Time
	UltimoNumero int
	UltimoFin    time.Time
}

// HaySprints informa si el proyecto tiene al menos un sprint.
func (l LimitesSprints) HaySprints() bool {
	return l.PrimerNumero != 0
}

// PuedeModificarse informa si el proyecto se puede modificar: no puede estar Finalizado (RN5). La pantalla lo usa
// para decidir si muestra el formulario.
func PuedeModificarse(p Proyecto, sprints LimitesSprints, hoy time.Time) bool {
	return EstadoActual(p, sprints.PrimerInicio, hoy) != EstadoFinalizado
}

// Modificar valida los datos nuevos del proyecto actual y lo devuelve modificado, con el mismo ID, la misma fecha de
// creación y el mismo estado guardado. Aplica las validaciones de la creación (RN1) y controla que las fechas no
// dejen sprints afuera (RN3 y RN4); los errores de los campos vienen juntos en domain.ErroresValidacion. Si el
// proyecto está Finalizado devuelve ErrProyectoFinalizado sin validar nada más (RN5). El nombre repetido lo controla
// el caso de uso, porque necesita los demás proyectos (RN2).
func Modificar(actual Proyecto, d Datos, sprints LimitesSprints, hoy time.Time) (Proyecto, error) {
	if !PuedeModificarse(actual, sprints, hoy) {
		return Proyecto{}, ErrProyectoFinalizado
	}
	nuevo, errs := validarComoAlCrear(d)
	if sprints.HaySprints() {
		// Un solo error por campo: los de los sprints se agregan solo si la fecha pasó las validaciones de HU-01.
		if _, invalido := errs[CampoFechaInicio]; !invalido && nuevo.FechaInicio.After(sprints.PrimerInicio) {
			errs.Agregar(CampoFechaInicio, fmt.Errorf("%w (el Sprint %d empieza el %s)", ErrInicioDespuesDelPrimerSprint,
				sprints.PrimerNumero, sprints.PrimerInicio.Format(formatoFecha)))
		}
		if _, invalido := errs[CampoFechaFin]; !invalido && nuevo.FechaFin.Before(sprints.UltimoFin) {
			errs.Agregar(CampoFechaFin, fmt.Errorf("%w (el Sprint %d termina el %s)", ErrFinAntesDelUltimoSprint,
				sprints.UltimoNumero, sprints.UltimoFin.Format(formatoFecha)))
		}
	}
	if len(errs) > 0 {
		return Proyecto{}, errs
	}
	actual.Nombre, actual.Descripcion = nuevo.Nombre, nuevo.Descripcion
	actual.FechaInicio, actual.FechaFin = nuevo.FechaInicio, nuevo.FechaFin
	return actual, nil
}
