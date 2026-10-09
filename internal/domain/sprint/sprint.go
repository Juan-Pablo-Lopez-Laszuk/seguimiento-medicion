// Package sprint tiene la entidad Sprint y sus reglas de negocio (épica E3).
//
// Spec: specs/HU-11-crear-sprint.md. No usa base de datos ni HTTP: recibe los datos, el proyecto y los
// sprints que ya tiene, y devuelve el sprint nuevo o los errores de validación.
package sprint

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
)

// Estado es la etapa en la que está un sprint.
type Estado string

// EstadoPlanificado es el estado con el que nace todo sprint (RN6).
const EstadoPlanificado Estado = "Planificado"

// Nombres de los campos, tal como se informan en los errores y en el formulario.
const (
	CampoObjetivo    = "objetivo"
	CampoFechaInicio = "fecha_inicio"
	CampoFechaFin    = "fecha_fin"
)

// ObjetivoMax es el largo máximo del Sprint Goal, en caracteres (RN1).
const ObjetivoMax = 500

// Errores de validación (sección 7 de la spec).
var (
	ErrObjetivoObligatorio = errors.New("el Sprint Goal es obligatorio")
	ErrObjetivoLargo       = errors.New("el Sprint Goal no puede superar los 500 caracteres")
	// Los errores de fechas son los mismos de HU-01, así el usuario ve el mismo mensaje en las dos pantallas.
	ErrFechaInicio     = project.ErrFechaInicio
	ErrFechaFin        = project.ErrFechaFin
	ErrFechasInvalidas = project.ErrFechasInvalidas
	// ErrFueraDelProyecto se informa con el rango del proyecto: "... (del 05/10/2026 al 01/11/2026)".
	ErrFueraDelProyecto = errors.New("la fecha tiene que estar dentro del proyecto")
	// ErrSuperpuesto se informa con el último sprint: "... (el Sprint 1 termina el 11/10/2026)".
	ErrSuperpuesto = errors.New("el sprint tiene que empezar después del último")
)

// formatoFecha es como se muestran las fechas en los mensajes: 05/10/2026.
const formatoFecha = "02/01/2006"

// Datos son los campos que carga el usuario para crear un sprint.
type Datos struct {
	Objetivo    string // Sprint Goal
	FechaInicio time.Time
	FechaFin    time.Time
}

// Sprint es un sprint ya validado.
type Sprint struct {
	ID          int64
	ProyectoID  int64
	Numero      int
	Objetivo    string
	FechaInicio time.Time
	FechaFin    time.Time
	Estado      Estado
}

// Nuevo valida los datos contra el proyecto y sus sprints existentes y devuelve el sprint listo para guardar.
// Si algún dato es inválido, devuelve domain.ErroresValidacion con todos los campos que fallaron (RN7).
func Nuevo(d Datos, p project.Proyecto, existentes []Sprint) (Sprint, error) {
	objetivo := strings.TrimSpace(d.Objetivo)
	inicio, fin := domain.SoloFecha(d.FechaInicio), domain.SoloFecha(d.FechaFin)

	errs := domain.ErroresValidacion{}
	errs.Agregar(CampoObjetivo, validarObjetivo(objetivo))
	errs.Agregar(CampoFechaInicio, validarInicio(inicio, p, existentes))
	errs.Agregar(CampoFechaFin, validarFin(inicio, fin, p))
	if len(errs) > 0 {
		return Sprint{}, errs
	}
	return Sprint{
		ProyectoID:  p.ID,
		Numero:      siguienteNumero(existentes),
		Objetivo:    objetivo,
		FechaInicio: inicio,
		FechaFin:    fin,
		Estado:      EstadoPlanificado,
	}, nil
}

func validarObjetivo(objetivo string) error {
	switch largo := utf8.RuneCountInString(objetivo); {
	case largo == 0:
		return ErrObjetivoObligatorio
	case largo > ObjetivoMax:
		return ErrObjetivoLargo
	}
	return nil
}

// validarInicio devuelve el primer error de la fecha de inicio, en este orden: falta, queda fuera del
// proyecto (RN3) o no empieza después del último sprint (RN4). Así se informa un solo error por campo.
func validarInicio(inicio time.Time, p project.Proyecto, existentes []Sprint) error {
	ultimo, hay := ultimoSprint(existentes)
	switch {
	case inicio.IsZero():
		return ErrFechaInicio
	case fueraDelProyecto(inicio, p):
		return errFueraDelProyecto(p)
	case hay && !inicio.After(ultimo.FechaFin):
		return fmt.Errorf("%w (el Sprint %d termina el %s)", ErrSuperpuesto,
			ultimo.Numero, ultimo.FechaFin.Format(formatoFecha))
	}
	return nil
}

// validarFin devuelve el primer error de la fecha de fin: falta, no es posterior al inicio (RN2) o queda
// fuera del proyecto (RN3).
func validarFin(inicio, fin time.Time, p project.Proyecto) error {
	switch {
	case fin.IsZero():
		return ErrFechaFin
	case !inicio.IsZero() && !fin.After(inicio):
		return ErrFechasInvalidas
	case fueraDelProyecto(fin, p):
		return errFueraDelProyecto(p)
	}
	return nil
}

// fueraDelProyecto informa si f cae antes del inicio o después del fin del proyecto (los bordes valen).
func fueraDelProyecto(f time.Time, p project.Proyecto) bool {
	return f.Before(domain.SoloFecha(p.FechaInicio)) || f.After(domain.SoloFecha(p.FechaFin))
}

// errFueraDelProyecto agrega al error el rango del proyecto, para que el usuario sepa qué fechas valen.
func errFueraDelProyecto(p project.Proyecto) error {
	return fmt.Errorf("%w (del %s al %s)", ErrFueraDelProyecto,
		p.FechaInicio.Format(formatoFecha), p.FechaFin.Format(formatoFecha))
}

// ultimoSprint devuelve el sprint que termina más tarde; hay es false si el proyecto todavía no tiene sprints.
func ultimoSprint(existentes []Sprint) (ultimo Sprint, hay bool) {
	for _, s := range existentes {
		if !hay || s.FechaFin.After(ultimo.FechaFin) {
			ultimo, hay = s, true
		}
	}
	return ultimo, hay
}

// siguienteNumero devuelve el número más alto de los sprints existentes más uno (RN5). No reutiliza
// números que falten en el medio, así un número nunca nombra a dos sprints distintos.
func siguienteNumero(existentes []Sprint) int {
	maximo := 0
	for _, s := range existentes {
		maximo = max(maximo, s.Numero)
	}
	return maximo + 1
}
