package effort

import (
	"errors"
	"strings"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
)

// CampoQuien es el campo del formulario de corrección donde se elige quién corrige (HU-26).
const CampoQuien = "quien"

// Errores de la consulta y la corrección de registros (sección 7 de la spec de HU-26).
var (
	ErrRegistroNoEncontrado = errors.New("no existe el registro de esfuerzo")
	ErrQuienInvalido        = errors.New("elegí quién corrige: un integrante activo del proyecto")
	ErrNoPuedeCorregir      = errors.New("cada integrante corrige solo sus registros (el Agile Enabler, todos)")
)

// Filtro elige registros por integrante, historia y fechas (HU-26). Un ID en 0 o una fecha cero no filtran.
type Filtro struct {
	IntegranteID int64
	HistoriaID   int64
	Desde        time.Time // inicio del sprint elegido (RN2), inclusive
	Hasta        time.Time // fin del sprint elegido, inclusive
}

// Cumple dice si el registro cumple todo lo que el filtro eligió (CA-26.1).
func (f Filtro) Cumple(r Registro) bool {
	fecha := domain.SoloFecha(r.Fecha)
	return (f.IntegranteID == 0 || r.IntegranteID == f.IntegranteID) &&
		(f.HistoriaID == 0 || r.HistoriaID == f.HistoriaID) &&
		(f.Desde.IsZero() || !fecha.Before(domain.SoloFecha(f.Desde))) &&
		(f.Hasta.IsZero() || !fecha.After(domain.SoloFecha(f.Hasta)))
}

// PuedeCorregir devuelve nil si quien puede corregir el registro: es su dueño o es el Agile Enabler (CA-26.2, RN4).
// Quien corrige tiene que ser un integrante activo.
func PuedeCorregir(quien member.Integrante, r Registro) error {
	switch {
	case quien.ID == 0 || !quien.Activo:
		return ErrQuienInvalido
	case quien.Rol == member.RolAgileEnabler || quien.ID == r.IntegranteID:
		return nil
	}
	return ErrNoPuedeCorregir
}

// Corregir valida la corrección que hace el integrante quienID y devuelve el registro corregido, con el mismo ID,
// integrante e historia (RN5). Valida lo mismo que HU-24, sin contar el mismo registro en las 24 h del día (RN6). Si
// algún dato es inválido, devuelve domain.ErroresValidacion con todos los campos que fallaron.
func Corregir(r Registro, quienID int64, d DatosRegistro, c ContextoRegistro) (Registro, error) {
	actividad := strings.TrimSpace(d.Actividad)
	fecha, errFecha := validarFecha(d.Fecha, c)
	horas, errHoras := parsearHorasTrabajadas(d.Horas)
	otros := make([]Registro, 0, len(c.DelIntegrante))
	for _, o := range c.DelIntegrante {
		if o.ID != r.ID {
			otros = append(otros, o)
		}
	}
	if errHoras == nil && errFecha == nil && horasDelDia(otros, r.IntegranteID, fecha)+horas > HorasPorDiaMax {
		errHoras = ErrMasDe24HorasEnElDia
	}

	errs := domain.ErroresValidacion{}
	errs.Agregar(CampoQuien, PuedeCorregir(buscarIntegrante(quienID, c.Integrantes), r))
	errs.Agregar(CampoTarea, validarTarea(d.TareaID, c))
	errs.Agregar(CampoFecha, errFecha)
	errs.Agregar(CampoActividad, validarActividad(actividad))
	errs.Agregar(CampoHoras, errHoras)
	if len(errs) > 0 {
		return Registro{}, errs
	}
	r.TareaID, r.Fecha, r.Actividad, r.Horas = d.TareaID, fecha, actividad, horas
	return r, nil
}

// buscarIntegrante devuelve el integrante con ese ID, o uno vacío si no es del proyecto.
func buscarIntegrante(id int64, integrantes []member.Integrante) member.Integrante {
	for _, i := range integrantes {
		if i.ID == id {
			return i
		}
	}
	return member.Integrante{}
}
