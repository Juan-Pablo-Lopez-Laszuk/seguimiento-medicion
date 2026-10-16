package backlog

import (
	"strings"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
)

// Cambios son los datos que se corrigen de una historia que ya existe (HU-06): los mismos que se cargan al crearla,
// menos los story points (HU-16) y el estado (HU-08), que tienen su propia pantalla.
type Cambios struct {
	Titulo      string
	Descripcion string
	Prioridad   Prioridad
}

// Editar devuelve la historia con los datos corregidos. No cambia el identificador (CA-06.1), el estado, los story
// points ni lo que tenga asociado (RN2), y se puede en cualquier estado (RN3). La historia original no se modifica.
// Con datos inválidos devuelve domain.ErroresValidacion con todos los campos que fallaron (RN1).
func (h Historia) Editar(c Cambios) (Historia, error) {
	titulo, descripcion := strings.TrimSpace(c.Titulo), strings.TrimSpace(c.Descripcion)

	errs := domain.ErroresValidacion{}
	errs.Agregar(CampoTitulo, validarTitulo(titulo))
	errs.Agregar(CampoDescripcion, validarDescripcion(descripcion))
	errs.Agregar(CampoPrioridad, validarPrioridad(c.Prioridad))
	if len(errs) > 0 {
		return Historia{}, errs
	}
	h.Titulo, h.Descripcion, h.Prioridad = titulo, descripcion, c.Prioridad
	return h, nil
}
