// Package sprint tiene la entidad Sprint y sus reglas de negocio (épica E3).
//
// Spec: specs/HU-11-crear-sprint.md. No usa base de datos ni HTTP: recibe los datos, el proyecto y los
// sprints que ya tiene, y devuelve el sprint nuevo o los errores de validación.
package sprint

import (
	"errors"
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
)

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
func Nuevo(d Datos, p project.Proyecto, _ []Sprint) (Sprint, error) {
	objetivo := strings.TrimSpace(d.Objetivo)
	inicio, fin := domain.SoloFecha(d.FechaInicio), domain.SoloFecha(d.FechaFin)

	errs := domain.ErroresValidacion{}
	errs.Agregar(CampoObjetivo, validarObjetivo(objetivo))
	if inicio.IsZero() {
		errs.Agregar(CampoFechaInicio, ErrFechaInicio)
	}
	switch {
	case fin.IsZero():
		errs.Agregar(CampoFechaFin, ErrFechaFin)
	case !inicio.IsZero() && !fin.After(inicio): // RN2
		errs.Agregar(CampoFechaFin, ErrFechasInvalidas)
	}
	if len(errs) > 0 {
		return Sprint{}, errs
	}
	return Sprint{
		ProyectoID:  p.ID,
		Numero:      1,
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
