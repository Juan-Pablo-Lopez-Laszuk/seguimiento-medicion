// Package project tiene la entidad Proyecto y sus reglas de negocio (épica E1).
//
// Spec: specs/HU-01-crear-proyecto.md. No usa base de datos ni HTTP: recibe datos y devuelve un
// proyecto o los errores de validación.
package project

import (
	"errors"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

// Estado es la etapa en la que está un proyecto.
type Estado string

// EstadoPlanificado es el estado con el que nace todo proyecto (RN6).
const EstadoPlanificado Estado = "Planificado"

// Nombres de los campos, tal como se informan en los errores y en el formulario.
const (
	CampoNombre      = "nombre"
	CampoDescripcion = "descripcion"
	CampoFechaInicio = "fecha_inicio"
	CampoFechaFin    = "fecha_fin"
)

// Límites del nombre, en caracteres (RN2).
const (
	NombreMin = 3
	NombreMax = 100
	// DescripcionMax es el largo máximo de la descripción, en caracteres.
	DescripcionMax = 1000
)

// Errores de validación (sección 7 de la spec).
var (
	ErrNombreObligatorio = errors.New("el nombre es obligatorio")
	ErrNombreLargo       = errors.New("el nombre debe tener entre 3 y 100 caracteres")
	ErrDescripcionLarga  = errors.New("la descripción no puede superar los 1000 caracteres")
	ErrFechaInicio       = errors.New("la fecha de inicio es obligatoria y debe ser válida")
	ErrFechaFin          = errors.New("la fecha de finalización es obligatoria y debe ser válida")
	ErrFechasInvalidas   = errors.New("la fecha de finalización debe ser posterior a la de inicio")
)

// ErroresValidacion junta los errores de todos los campos inválidos, por nombre de campo (RN7).
type ErroresValidacion map[string]error

// Error arma un texto con todos los errores, ordenados por campo: "nombre: el nombre es obligatorio".
func (e ErroresValidacion) Error() string {
	campos := make([]string, 0, len(e))
	for campo := range e {
		campos = append(campos, campo)
	}
	sort.Strings(campos)
	partes := make([]string, 0, len(campos))
	for _, campo := range campos {
		partes = append(partes, campo+": "+e[campo].Error())
	}
	return strings.Join(partes, "; ")
}

// Datos son los campos que carga el usuario para crear un proyecto.
type Datos struct {
	Nombre      string
	Descripcion string
	FechaInicio time.Time
	FechaFin    time.Time
}

// Proyecto es un proyecto ya validado.
type Proyecto struct {
	ID          int64
	Nombre      string
	Descripcion string
	FechaInicio time.Time
	FechaFin    time.Time
	Estado      Estado
	CreadoEn    time.Time
}

// Nuevo valida los datos y devuelve el proyecto listo para guardar. Si algún dato es inválido,
// devuelve ErroresValidacion con todos los campos que fallaron.
func Nuevo(d Datos) (Proyecto, error) {
	errs := ErroresValidacion{}
	nombre := strings.TrimSpace(d.Nombre)   // RN1
	largo := utf8.RuneCountInString(nombre) // RN2: caracteres, no bytes
	switch {
	case nombre == "":
		errs[CampoNombre] = ErrNombreObligatorio
	case largo < NombreMin || largo > NombreMax:
		errs[CampoNombre] = ErrNombreLargo
	}
	if utf8.RuneCountInString(d.Descripcion) > DescripcionMax {
		errs[CampoDescripcion] = ErrDescripcionLarga
	}
	inicio, fin := soloFecha(d.FechaInicio), soloFecha(d.FechaFin)
	switch {
	case inicio.IsZero():
		errs[CampoFechaInicio] = ErrFechaInicio
	}
	switch {
	case fin.IsZero():
		errs[CampoFechaFin] = ErrFechaFin
	case !inicio.IsZero() && !fin.After(inicio): // RN4
		errs[CampoFechaFin] = ErrFechasInvalidas
	}
	if len(errs) > 0 {
		return Proyecto{}, errs
	}
	return Proyecto{
		Nombre:      nombre,
		Descripcion: d.Descripcion,
		FechaInicio: inicio,
		FechaFin:    fin,
		Estado:      EstadoPlanificado,
	}, nil
}

// soloFecha deja el día de t a medianoche UTC, así la comparación no depende de la hora ni de la zona horaria.
// La fecha cero (campo vacío) se mantiene en cero.
func soloFecha(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
