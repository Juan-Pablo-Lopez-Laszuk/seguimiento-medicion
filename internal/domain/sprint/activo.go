package sprint

import (
	"slices"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
)

// Activo devuelve el sprint activo: el no cerrado de menor número cuya fecha de inicio ya llegó
// (spec: specs/HU-04-estado-proyecto.md, RN2). Un sprint vencido sin cerrar sigue activo, así se puede
// cerrar tarde. ok es false si no hay ninguno. La usan también HU-13 y HU-14.
func Activo(sprints []Sprint, hoy time.Time) (activo Sprint, ok bool) {
	hoy = domain.SoloFecha(hoy)
	for _, s := range porNumero(sprints) {
		if s.Estado == EstadoCerrado {
			continue
		}
		// Es el primero sin cerrar: si todavía no empezó, los siguientes tampoco (HU-11 RN4).
		if hoy.Before(s.FechaInicio) {
			return Sprint{}, false
		}
		return s, true
	}
	return Sprint{}, false
}

// Proximo devuelve el primer sprint sin cerrar que todavía no empezó. Se muestra cuando no hay sprint activo.
func Proximo(sprints []Sprint, hoy time.Time) (proximo Sprint, ok bool) {
	hoy = domain.SoloFecha(hoy)
	for _, s := range porNumero(sprints) {
		if s.Estado != EstadoCerrado && hoy.Before(s.FechaInicio) {
			return s, true
		}
	}
	return Sprint{}, false
}

// Terminado informa si la fecha de fin del sprint ya pasó. Si sigue activo, falta cerrarlo (HU-04 RN3).
func (s Sprint) Terminado(hoy time.Time) bool {
	return domain.SoloFecha(hoy).After(s.FechaFin)
}

// InicioDelPrimero devuelve la fecha de inicio del sprint de menor número, o la fecha cero si no hay
// sprints. Es lo que necesita project.EstadoActual (HU-04 RN1).
func InicioDelPrimero(sprints []Sprint) time.Time {
	ordenados := porNumero(sprints)
	if len(ordenados) == 0 {
		return time.Time{}
	}
	return ordenados[0].FechaInicio
}

// porNumero devuelve una copia de los sprints ordenada por número, sin tocar la original.
func porNumero(sprints []Sprint) []Sprint {
	ordenados := slices.Clone(sprints)
	slices.SortFunc(ordenados, func(a, b Sprint) int { return a.Numero - b.Numero })
	return ordenados
}
