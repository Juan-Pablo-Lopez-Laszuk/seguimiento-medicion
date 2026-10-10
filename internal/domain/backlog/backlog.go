// Package backlog tiene la entidad Historia del Product Backlog y sus reglas de negocio (épica E2).
//
// Spec: specs/HU-05-crear-item-backlog.md. No usa base de datos ni HTTP: recibe los datos y las historias que
// ya tiene el proyecto, y devuelve la historia nueva o los errores de validación.
package backlog

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
)

// Prioridad es la importancia de una historia. Los valores son los mismos que guarda la base.
type Prioridad string

// Prioridades válidas (RN3).
const (
	PrioridadAlta  Prioridad = "Alta"
	PrioridadMedia Prioridad = "Media"
	PrioridadBaja  Prioridad = "Baja"
)

// Prioridades es la lista de prioridades válidas, en el orden en que se ofrecen en el formulario.
var Prioridades = []Prioridad{PrioridadAlta, PrioridadMedia, PrioridadBaja}

func (p Prioridad) valida() bool {
	return p == PrioridadAlta || p == PrioridadMedia || p == PrioridadBaja
}

// Estado es la etapa en la que está una historia.
type Estado string

// EstadoPendiente es el estado con el que nace toda historia (RN6). Los demás llegan con HU-08.
const EstadoPendiente Estado = "Pendiente"

// Nombres de los campos, tal como se informan en los errores y en el formulario.
const (
	CampoTitulo      = "titulo"
	CampoDescripcion = "descripcion"
	CampoPrioridad   = "prioridad"
)

// Límites, en caracteres: el del título (RN1) y el de la descripción (RN2).
const (
	TituloMax      = 120
	DescripcionMax = 2000
)

// Errores de validación (sección 7 de la spec).
var (
	ErrTituloObligatorio = errors.New("el título es obligatorio")
	ErrTituloLargo       = errors.New("el título no puede superar los 120 caracteres")
	ErrDescripcionLarga  = errors.New("la descripción no puede superar los 2000 caracteres")
	ErrPrioridadInvalida = errors.New("la prioridad tiene que ser Alta, Media o Baja")
)

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
// Si algún dato es inválido, devuelve domain.ErroresValidacion con todos los campos que fallaron (RN7).
func Nueva(d Datos, proyectoID int64, _ []Historia) (Historia, error) {
	titulo, descripcion := strings.TrimSpace(d.Titulo), strings.TrimSpace(d.Descripcion)

	errs := domain.ErroresValidacion{}
	errs.Agregar(CampoTitulo, validarTitulo(titulo))
	errs.Agregar(CampoDescripcion, validarDescripcion(descripcion))
	errs.Agregar(CampoPrioridad, validarPrioridad(d.Prioridad))
	if len(errs) > 0 {
		return Historia{}, errs
	}
	return Historia{
		ProyectoID:  proyectoID,
		Numero:      1,
		Titulo:      titulo,
		Descripcion: descripcion,
		Prioridad:   d.Prioridad,
		Estado:      EstadoPendiente,
	}, nil
}

func validarPrioridad(p Prioridad) error {
	if !p.valida() {
		return ErrPrioridadInvalida
	}
	return nil
}

func validarDescripcion(descripcion string) error {
	if utf8.RuneCountInString(descripcion) > DescripcionMax {
		return ErrDescripcionLarga
	}
	return nil
}

func validarTitulo(titulo string) error {
	switch largo := utf8.RuneCountInString(titulo); {
	case largo == 0:
		return ErrTituloObligatorio
	case largo > TituloMax:
		return ErrTituloLargo
	}
	return nil
}
