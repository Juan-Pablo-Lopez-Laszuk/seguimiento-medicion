// Package defect tiene la entidad Defecto y sus reglas de negocio (épica E6).
//
// Spec: specs/HU-27-registrar-defecto.md. No usa base de datos ni HTTP: recibe los datos, los sprints y las
// historias del proyecto, y devuelve el defecto o los errores de validación.
package defect

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
)

// Severidad es qué tan grave es un defecto. Los valores son los mismos que guarda la base.
type Severidad string

// Severidades válidas (RN2), de la más grave a la más leve, en el orden en que se ofrecen en el formulario.
const (
	SeveridadCritica Severidad = "Crítica"
	SeveridadAlta    Severidad = "Alta"
	SeveridadMedia   Severidad = "Media"
	SeveridadBaja    Severidad = "Baja"
)

// Severidades es la lista de severidades válidas.
var Severidades = []Severidad{SeveridadCritica, SeveridadAlta, SeveridadMedia, SeveridadBaja}

// Estado es la etapa en la que está un defecto. Los cambios de estado son de HU-28.
type Estado string

// EstadoAbierto es el estado con el que nace todo defecto (RN3).
const EstadoAbierto Estado = "Abierto"

// DescripcionMax es el largo máximo de la descripción, en caracteres (RN1).
const DescripcionMax = 1000

// Nombres de los campos, tal como se informan en los errores y en el formulario.
const (
	CampoDescripcion = "descripcion"
	CampoSeveridad   = "severidad"
	CampoSprint      = "sprint"
	CampoHistoria    = "historia"
)

// Errores de validación (sección 7 de la spec).
var (
	ErrDescripcionObligatoria = errors.New("la descripción del defecto es obligatoria")
	ErrDescripcionLarga       = errors.New("la descripción del defecto no puede superar los 1000 caracteres")
	ErrSeveridadInvalida      = errors.New("la severidad tiene que ser Crítica, Alta, Media o Baja")
	ErrSprintInvalido         = errors.New("el sprint de detección tiene que ser un sprint del proyecto")
	ErrHistoriaInvalida       = errors.New("la historia tiene que ser del proyecto")
)

// Datos son los campos que carga el usuario para registrar un defecto.
type Datos struct {
	Descripcion string
	Severidad   Severidad
	SprintID    int64 // sprint de detección
	HistoriaID  int64 // 0 si el defecto no es de una historia
}

// Defecto es un error encontrado en el producto, con su severidad y el sprint en que se detectó.
type Defecto struct {
	ID                 int64
	ProyectoID         int64
	HistoriaID         int64 // 0 si no es de una historia
	Descripcion        string
	Severidad          Severidad
	Estado             Estado
	SprintDeteccionID  int64
	SprintResolucionID int64 // 0 mientras no se resuelve (HU-28)
}

// Nuevo valida los datos contra los sprints y las historias del proyecto y devuelve el defecto Abierto (RN3), listo
// para guardar. Si algún dato es inválido, devuelve domain.ErroresValidacion con todos los campos que fallaron (RN6).
func Nuevo(d Datos, proyectoID int64, sprints []sprint.Sprint, historias []backlog.Historia) (Defecto, error) {
	descripcion := strings.TrimSpace(d.Descripcion)

	errs := domain.ErroresValidacion{}
	errs.Agregar(CampoDescripcion, validarDescripcion(descripcion))
	errs.Agregar(CampoSeveridad, validarSeveridad(d.Severidad))
	errs.Agregar(CampoSprint, validarSprint(d.SprintID, proyectoID, sprints))
	errs.Agregar(CampoHistoria, validarHistoria(d.HistoriaID, proyectoID, historias))
	if len(errs) > 0 {
		return Defecto{}, errs
	}
	return Defecto{
		ProyectoID: proyectoID, HistoriaID: d.HistoriaID, Descripcion: descripcion,
		Severidad: d.Severidad, Estado: EstadoAbierto, SprintDeteccionID: d.SprintID,
	}, nil
}

func validarDescripcion(descripcion string) error {
	switch largo := utf8.RuneCountInString(descripcion); {
	case largo == 0:
		return ErrDescripcionObligatoria
	case largo > DescripcionMax:
		return ErrDescripcionLarga
	}
	return nil
}

func validarSeveridad(s Severidad) error {
	for _, valida := range Severidades {
		if s == valida {
			return nil
		}
	}
	return ErrSeveridadInvalida
}

func validarSprint(id, proyectoID int64, sprints []sprint.Sprint) error {
	for _, s := range sprints {
		if s.ID == id && s.ProyectoID == proyectoID {
			return nil
		}
	}
	return ErrSprintInvalido
}

func validarHistoria(id, proyectoID int64, historias []backlog.Historia) error {
	if id == 0 {
		return nil
	}
	for _, h := range historias {
		if h.ID == id && h.ProyectoID == proyectoID {
			return nil
		}
	}
	return ErrHistoriaInvalida
}
