// Package backlog tiene la entidad Historia del Product Backlog y sus reglas de negocio (épica E2).
//
// Spec: specs/HU-05-crear-item-backlog.md. No usa base de datos ni HTTP: recibe los datos y las historias que
// ya tiene el proyecto, y devuelve la historia nueva o los errores de validación.
package backlog

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
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

// EstadoPendiente es el estado con el que nace toda historia (RN6). Los demás llegan con HU-08; EstadoHecho
// ya está porque lo lee HU-04 para el porcentaje de story points completados.
const (
	EstadoPendiente Estado = "Pendiente"
	EstadoHecho     Estado = "Hecho"
)

// Nombres de los campos, tal como se informan en los errores y en el formulario.
const (
	CampoTitulo      = "titulo"
	CampoDescripcion = "descripcion"
	CampoPrioridad   = "prioridad"
	CampoStoryPoints = "story_points"
)

// EscalaStoryPoints es la escala acordada (Fibonacci) de story points, de menor a mayor. La misma que valida la
// base (historias_story_points_ck) y que usa HU-16 para estimar.
var EscalaStoryPoints = []int{0, 1, 2, 3, 5, 8, 13, 21}

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
	// ErrStoryPointsInvalidos cubre un valor fuera de la escala y un texto que no es un número entero.
	ErrStoryPointsInvalidos = errors.New("los story points tienen que ser 0, 1, 2, 3, 5, 8, 13 o 21")
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
func Nueva(d Datos, proyectoID int64, existentes []Historia) (Historia, error) {
	titulo, descripcion := strings.TrimSpace(d.Titulo), strings.TrimSpace(d.Descripcion)

	errs := domain.ErroresValidacion{}
	errs.Agregar(CampoTitulo, validarTitulo(titulo))
	errs.Agregar(CampoDescripcion, validarDescripcion(descripcion))
	errs.Agregar(CampoPrioridad, validarPrioridad(d.Prioridad))
	puntos, errPuntos := parsearStoryPoints(d.StoryPoints)
	errs.Agregar(CampoStoryPoints, errPuntos)
	if len(errs) > 0 {
		return Historia{}, errs
	}
	return Historia{
		ProyectoID:  proyectoID,
		Numero:      siguienteNumero(existentes),
		Titulo:      titulo,
		Descripcion: descripcion,
		Prioridad:   d.Prioridad,
		Estado:      EstadoPendiente,
		StoryPoints: puntos,
	}, nil
}

// siguienteNumero devuelve el número más alto de las historias del proyecto más uno (RN5). Con el más alto y no
// con la cantidad, un número que falte no se reutiliza.
func siguienteNumero(existentes []Historia) int {
	mayor := 0
	for _, h := range existentes {
		mayor = max(mayor, h.Numero)
	}
	return mayor + 1
}

// parsearStoryPoints convierte el texto del formulario. Vacío es "sin estimar" (nil, RN4); si no, tiene que ser
// un número entero de la escala.
func parsearStoryPoints(texto string) (*int, error) {
	texto = strings.TrimSpace(texto)
	if texto == "" {
		return nil, nil
	}
	n, err := strconv.Atoi(texto)
	// Atoi acepta un signo adelante ("+5"); acá solo valen dígitos.
	if err != nil || texto[0] < '0' || texto[0] > '9' || !slices.Contains(EscalaStoryPoints, n) {
		return nil, ErrStoryPointsInvalidos
	}
	return &n, nil
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
