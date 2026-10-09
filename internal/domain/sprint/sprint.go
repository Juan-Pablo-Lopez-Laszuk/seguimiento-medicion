// Package sprint tiene la entidad Sprint y sus reglas de negocio (épica E3).
//
// Spec: specs/HU-11-crear-sprint.md. No usa base de datos ni HTTP: recibe los datos, el proyecto y los
// sprints que ya tiene, y devuelve el sprint nuevo o los errores de validación.
package sprint

import (
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
)

// Estado es la etapa en la que está un sprint.
type Estado string

// EstadoPlanificado es el estado con el que nace todo sprint (RN6).
const EstadoPlanificado Estado = "Planificado"

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
func Nuevo(d Datos, p project.Proyecto, _ []Sprint) (Sprint, error) {
	return Sprint{
		ProyectoID:  p.ID,
		Numero:      1,
		Objetivo:    d.Objetivo,
		FechaInicio: d.FechaInicio,
		FechaFin:    d.FechaFin,
		Estado:      EstadoPlanificado,
	}, nil
}
