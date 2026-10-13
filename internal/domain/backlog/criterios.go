package backlog

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
)

// DescripcionCriterioMax es el largo máximo de la descripción de un criterio, en caracteres (RN1).
const DescripcionCriterioMax = 300

// Errores de validación de un criterio (sección 7 de la spec de HU-07). Se informan en el campo CampoDescripcion.
var (
	ErrDescripcionCriterioObligatoria = errors.New("la descripción del criterio es obligatoria")
	ErrDescripcionCriterioLarga       = errors.New("la descripción del criterio no puede superar los 300 caracteres")
)

// ErrHistoriaHecha se devuelve al querer cambiar los criterios de una historia Hecha (RN7): agregar uno o desmarcarlo
// rompería que una historia Hecha tiene todos sus criterios cumplidos.
var ErrHistoriaHecha = errors.New("la historia está Hecha: no se pueden cambiar sus criterios")

// Errores que no son de un campo del formulario: los devuelven los repositorios cuando no hay lo que se pidió.
var (
	ErrHistoriaNoEncontrada = errors.New("no existe la historia")
	ErrCriterioNoEncontrado = errors.New("no existe el criterio")
)

// Criterio es un criterio de aceptación de una historia: algo que se verifica para dar la historia por terminada.
//
// Spec: specs/HU-07-criterios-aceptacion.md.
type Criterio struct {
	ID          int64
	HistoriaID  int64
	Descripcion string
	Cumplido    bool
}

// NuevoCriterio valida la descripción y devuelve el criterio listo para guardar en la historia. Todo criterio nace
// no cumplido (RN2) y se guarda sin los espacios del inicio y del final de la descripción (RN1). Si la descripción
// es inválida, devuelve domain.ErroresValidacion con el error del campo CampoDescripcion. En una historia Hecha no se
// agregan criterios (RN7): devuelve ErrHistoriaHecha, y eso se revisa antes que la descripción.
func NuevoCriterio(descripcion string, h Historia) (Criterio, error) {
	if h.Estado == EstadoHecho {
		return Criterio{}, ErrHistoriaHecha
	}
	descripcion = strings.TrimSpace(descripcion)

	errs := domain.ErroresValidacion{}
	errs.Agregar(CampoDescripcion, validarDescripcionCriterio(descripcion))
	if len(errs) > 0 {
		return Criterio{}, errs
	}
	return Criterio{HistoriaID: h.ID, Descripcion: descripcion}, nil
}

func validarDescripcionCriterio(descripcion string) error {
	switch largo := utf8.RuneCountInString(descripcion); {
	case largo == 0:
		return ErrDescripcionCriterioObligatoria
	case largo > DescripcionCriterioMax:
		return ErrDescripcionCriterioLarga
	}
	return nil
}

// Marcar devuelve el criterio marcado como cumplido o no cumplido, según cumplido; el original no cambia. Es
// explícito y no un "alternar" (RN3): marcar un criterio que ya está en ese estado lo deja igual. En una historia
// Hecha no se desmarca ningún criterio (RN7): devuelve ErrHistoriaHecha.
func (c Criterio) Marcar(cumplido bool, h Historia) (Criterio, error) {
	if h.Estado == EstadoHecho && !cumplido {
		return Criterio{}, ErrHistoriaHecha
	}
	c.Cumplido = cumplido
	return c, nil
}

// Progreso es cuántos criterios de una historia están cumplidos sobre el total (RN4). Es lo que se muestra como
// "2/3" y lo que HU-12 y HU-08 le preguntan a la historia.
type Progreso struct {
	Cumplidos int
	Total     int
}

// ProgresoDe cuenta los criterios cumplidos y el total de la lista. Sin criterios es 0/0.
func ProgresoDe(criterios []Criterio) Progreso {
	p := Progreso{Total: len(criterios)}
	for _, c := range criterios {
		if c.Cumplido {
			p.Cumplidos++
		}
	}
	return p
}

// Texto devuelve el progreso como se muestra en pantalla: "2/3".
func (p Progreso) Texto() string {
	return fmt.Sprintf("%d/%d", p.Cumplidos, p.Total)
}

// TieneCriterios dice si la historia tiene al menos un criterio (RN5). Es la regla de CA-07.1: HU-12 la usa para no
// asignar al sprint una historia sin criterios.
func (p Progreso) TieneCriterios() bool {
	return p.Total > 0
}

// Completo dice si la historia tiene criterios y todos están cumplidos (RN6). Con 0/0 no está completa: que no haya
// nada que cumplir no es lo mismo que haberlo cumplido. HU-08 la usa para pasar a Hecho (CA-08.2).
func (p Progreso) Completo() bool {
	return p.Total > 0 && p.Cumplidos == p.Total
}
