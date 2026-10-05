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
)

// Estado es la etapa en la que está un proyecto.
type Estado string

// EstadoPlanificado es el estado con el que nace todo proyecto (RN6).
const EstadoPlanificado Estado = "Planificado"

// Nombres de los campos, tal como se informan en los errores y en el formulario.
const (
	CampoNombre = "nombre"
)

// Errores de validación (sección 7 de la spec).
var (
	ErrNombreObligatorio = errors.New("el nombre es obligatorio")
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
	nombre := strings.TrimSpace(d.Nombre) // RN1
	if nombre == "" {
		errs[CampoNombre] = ErrNombreObligatorio
	}
	if len(errs) > 0 {
		return Proyecto{}, errs
	}
	return Proyecto{
		Nombre:      nombre,
		Descripcion: d.Descripcion,
		FechaInicio: d.FechaInicio,
		FechaFin:    d.FechaFin,
		Estado:      EstadoPlanificado,
	}, nil
}
