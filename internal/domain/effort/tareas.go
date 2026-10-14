package effort

import (
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
)

// TituloTareaMax es el largo máximo del título de una tarea, en caracteres (RN1 de HU-25), el mismo que el de una
// historia.
const TituloTareaMax = 120

// Nombres de los campos de una tarea, tal como se informan en los errores y en el formulario.
const (
	CampoTitulo      = "titulo"
	CampoResponsable = "responsable"
	CampoHoras       = "horas"
)

// Errores de validación de una tarea (sección 7 de la spec de HU-25). Las horas usan ErrHorasInvalidas de HU-23.
var (
	ErrTituloTareaObligatorio = errors.New("el título de la tarea es obligatorio")
	ErrTituloTareaLargo       = errors.New("el título de la tarea no puede superar los 120 caracteres")
	ErrResponsableObligatorio = errors.New("el responsable es obligatorio")
	ErrResponsableInvalido    = errors.New("el responsable tiene que ser un integrante activo del proyecto")
)

// ErrTareaNoEncontrada lo devuelven los repositorios cuando no hay una tarea con ese ID en la historia (CA-25.2).
var ErrTareaNoEncontrada = errors.New("no existe la tarea")

// ErrTareaCompletada se devuelve al querer cambiar la estimación de una tarea completada (CA-25.3).
var ErrTareaCompletada = errors.New("la tarea está completada: no se puede cambiar su estimación")

// DatosTarea son los campos que carga el usuario para agregar una tarea a una historia.
type DatosTarea struct {
	Titulo        string
	ResponsableID int64 // 0 si no eligió ninguno
	Horas         string
}

// Tarea es una parte del trabajo de una historia, con su responsable y sus horas estimadas.
//
// Spec: specs/HU-25-descomponer-tareas.md. HU-23 ya la usaba para sumar las horas de una historia (CA-23.2).
type Tarea struct {
	ID             int64
	HistoriaID     int64
	Titulo         string
	ResponsableID  int64
	HorasEstimadas float64 // 0 si la tarea todavía no está estimada
	Completada     bool
}

// NuevaTarea valida los datos y devuelve la tarea lista para guardar en la historia historiaID (CA-25.2: la historia
// se fija al crearla). integrantes son los del proyecto de la historia, activos y dados de baja. Toda tarea nace
// pendiente (RN5) y se guarda sin los espacios del inicio y del final del título (RN1). Si algún dato es inválido,
// devuelve domain.ErroresValidacion con todos los campos que fallaron.
func NuevaTarea(d DatosTarea, historiaID int64, integrantes []member.Integrante) (Tarea, error) {
	titulo := strings.TrimSpace(d.Titulo)
	horas, errHoras := ParsearHoras(d.Horas)

	errs := domain.ErroresValidacion{}
	errs.Agregar(CampoTitulo, validarTituloTarea(titulo))
	errs.Agregar(CampoResponsable, validarResponsable(d.ResponsableID, integrantes))
	errs.Agregar(CampoHoras, errHoras)
	if len(errs) > 0 {
		return Tarea{}, errs
	}
	return Tarea{HistoriaID: historiaID, Titulo: titulo, ResponsableID: d.ResponsableID, HorasEstimadas: horas}, nil
}

func validarTituloTarea(titulo string) error {
	switch largo := utf8.RuneCountInString(titulo); {
	case largo == 0:
		return ErrTituloTareaObligatorio
	case largo > TituloTareaMax:
		return ErrTituloTareaLargo
	}
	return nil
}

// validarResponsable controla que se haya elegido un responsable y que sea un integrante activo del proyecto (RN2).
func validarResponsable(id int64, integrantes []member.Integrante) error {
	if id == 0 {
		return ErrResponsableObligatorio
	}
	for _, i := range integrantes {
		if i.ID == id && i.Activo {
			return nil
		}
	}
	return ErrResponsableInvalido
}

// CambiarEstimacion devuelve la tarea con las horas estimadas nuevas, con la regla de HU-23 (RN3); la original no
// cambia. Una tarea completada no cambia su estimación (CA-25.3): devuelve ErrTareaCompletada, y eso se revisa antes
// que las horas.
func (t Tarea) CambiarEstimacion(horas string) (Tarea, error) {
	if t.Completada {
		return Tarea{}, ErrTareaCompletada
	}
	h, err := ParsearHoras(horas)
	if err != nil {
		return Tarea{}, err
	}
	t.HorasEstimadas = h
	return t, nil
}

// Marcar devuelve la tarea completada o pendiente, según completada; la original no cambia. Es explícito y no un
// "alternar" (RN7): marcarla dos veces igual la deja igual. Volver a pendiente sirve para corregir un error, y recién
// ahí se le pueden cambiar las horas. Las horas no cambian al completarla.
func (t Tarea) Marcar(completada bool) Tarea {
	t.Completada = completada
	return t
}
