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

// agregar registra el error del campo solo si hay error.
func (e ErroresValidacion) agregar(campo string, err error) {
	if err != nil {
		e[campo] = err
	}
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
	nombre := NormalizarNombre(d.Nombre)
	inicio, fin := soloFecha(d.FechaInicio), soloFecha(d.FechaFin)

	errs := ErroresValidacion{}
	errs.agregar(CampoNombre, validarNombre(nombre))
	errs.agregar(CampoDescripcion, validarDescripcion(d.Descripcion))
	errInicio, errFin := validarFechas(inicio, fin)
	errs.agregar(CampoFechaInicio, errInicio)
	errs.agregar(CampoFechaFin, errFin)
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

// NormalizarNombre quita los espacios del inicio y del final (RN1). La usa también el caso de uso para
// comparar el nombre con los proyectos existentes.
func NormalizarNombre(nombre string) string {
	return strings.TrimSpace(nombre)
}

func validarNombre(nombre string) error {
	largo := utf8.RuneCountInString(nombre) // RN2: caracteres, no bytes
	switch {
	case largo == 0:
		return ErrNombreObligatorio
	case largo < NombreMin || largo > NombreMax:
		return ErrNombreLargo
	}
	return nil
}

func validarDescripcion(descripcion string) error {
	if utf8.RuneCountInString(descripcion) > DescripcionMax {
		return ErrDescripcionLarga
	}
	return nil
}

// validarFechas devuelve el error de cada fecha (nil si está bien).
func validarFechas(inicio, fin time.Time) (errInicio, errFin error) {
	if inicio.IsZero() {
		errInicio = ErrFechaInicio
	}
	switch {
	case fin.IsZero():
		errFin = ErrFechaFin
	case !inicio.IsZero() && !fin.After(inicio): // RN4
		errFin = ErrFechasInvalidas
	}
	return errInicio, errFin
}

// soloFecha deja el día de t a medianoche UTC, así la comparación no depende de la hora ni de la zona horaria.
// La fecha cero (campo vacío) se mantiene en cero.
func soloFecha(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
