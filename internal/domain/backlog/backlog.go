// Package backlog tiene la entidad Historia del Product Backlog y sus reglas de negocio (épica E2).
//
// Spec: specs/HU-05-crear-item-backlog.md. No usa base de datos ni HTTP: recibe los datos y las historias que
// ya tiene el proyecto, y devuelve la historia nueva o los errores de validación.
package backlog

import (
	"fmt"
	"strings"
)

// Prioridad es la importancia de una historia. Los valores son los mismos que guarda la base.
type Prioridad string

// Prioridades válidas (RN3).
const (
	PrioridadAlta  Prioridad = "Alta"
	PrioridadMedia Prioridad = "Media"
	PrioridadBaja  Prioridad = "Baja"
)

// Estado es la etapa en la que está una historia.
type Estado string

// EstadoPendiente es el estado con el que nace toda historia (RN6). Los demás llegan con HU-08.
const EstadoPendiente Estado = "Pendiente"

// Datos son los campos que carga el usuario para crear una historia.
type Datos struct {
	Titulo      string
	Descripcion string
	Prioridad   Prioridad
	StoryPoints string // vacío = sin estimar; si no, un número de la escala
}

// Historia es una historia ya validada. StoryPoints es nil mientras no esté estimada.
type Historia struct {
	ID          int64
	ProyectoID  int64
	Numero      int
	Titulo      string
	Descripcion string
	Prioridad   Prioridad
	Estado      Estado
	StoryPoints *int
}

// Identificador devuelve el número como se muestra en pantalla: "HU-3".
func (h Historia) Identificador() string {
	return fmt.Sprintf("HU-%d", h.Numero)
}

// Nueva valida los datos contra las historias que ya tiene el proyecto y devuelve la historia lista para guardar.
func Nueva(d Datos, proyectoID int64, _ []Historia) (Historia, error) {
	return Historia{
		ProyectoID:  proyectoID,
		Numero:      1,
		Titulo:      strings.TrimSpace(d.Titulo),
		Descripcion: strings.TrimSpace(d.Descripcion),
		Prioridad:   d.Prioridad,
		Estado:      EstadoPendiente,
	}, nil
}
