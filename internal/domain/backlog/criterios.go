package backlog

import (
	"errors"
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
// es inválida, devuelve domain.ErroresValidacion con el error del campo CampoDescripcion.
func NuevoCriterio(descripcion string, h Historia) (Criterio, error) {
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
