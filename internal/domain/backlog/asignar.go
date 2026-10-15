package backlog

import (
	"errors"
	"fmt"
)

// Errores de asignar una historia a un sprint y de quitarla (sección 7 de specs/HU-12-asignar-historias.md). Se
// devuelven con un mensaje para el usuario que nombra la historia o el sprint, y se reconocen con errors.Is.
var (
	ErrSinEstimar       = errors.New("la historia no está estimada")
	ErrSinCriterios     = errors.New("la historia no tiene criterios de aceptación")
	ErrYaHecha          = errors.New("la historia ya está Hecha")
	ErrEnOtroSprint     = errors.New("la historia ya está en un sprint abierto")
	ErrSprintCerrado    = errors.New("el sprint está cerrado")
	ErrYaEmpezada       = errors.New("la historia ya empezó")
	ErrNoEstaEnUnSprint = errors.New("la historia no está en ningún sprint")
)

// DatosSprint es lo que la regla de asignación necesita saber de un sprint: su ID, su número (para los mensajes) y si
// está cerrado. Así backlog no depende del paquete sprint.
type DatosSprint struct {
	ID      int64
	Numero  int
	Cerrado bool
}

// AsignarA devuelve la historia En sprint y en el sprint destino, si se puede asignar (HU-12); la original no cambia.
// sprints son los del proyecto, para nombrar el sprint en el que ya está. Rechaza un sprint cerrado (RN2), una
// historia Hecha, una que ya está en un sprint abierto (RN3), una sin estimar o una sin criterios (RN1). El cambio de
// estado pasa por CambiarEstado, la regla de HU-08 (RN6).
func (h Historia) AsignarA(destino DatosSprint, progreso Progreso, sprints []DatosSprint) (Historia, error) {
	switch {
	case destino.Cerrado:
		return Historia{}, errorDeEstado{base: ErrSprintCerrado,
			mensaje: fmt.Sprintf("el Sprint %d está cerrado: no se le pueden asignar historias", destino.Numero)}
	case h.Estado == EstadoHecho:
		return Historia{}, h.noSePuedeAsignar(ErrYaHecha, "ya está Hecha")
	case h.SprintID != nil:
		return Historia{}, errorDeEstado{base: ErrEnOtroSprint, mensaje: h.Identificador() + " ya está en " + nombreDelSprint(*h.SprintID, sprints)}
	case h.StoryPoints == nil:
		return Historia{}, h.noSePuedeAsignar(ErrSinEstimar, "no está estimada")
	case !progreso.TieneCriterios():
		return Historia{}, h.noSePuedeAsignar(ErrSinCriterios, "no tiene criterios de aceptación")
	}
	asignada, err := h.CambiarEstado(EstadoEnSprint, progreso)
	if err != nil {
		return Historia{}, err
	}
	id := destino.ID
	asignada.SprintID = &id
	return asignada, nil
}

// QuitarDelSprint devuelve la historia de vuelta en Pendiente y sin sprint (RN4); la original no cambia. Solo se
// quita mientras está En sprint: una que ya empezó (En progreso o Hecho) sigue en su sprint.
func (h Historia) QuitarDelSprint() (Historia, error) {
	switch {
	case h.SprintID == nil:
		return Historia{}, errorDeEstado{base: ErrNoEstaEnUnSprint, mensaje: h.Identificador() + " no está en ningún sprint"}
	case h.Estado != EstadoEnSprint:
		return Historia{}, errorDeEstado{base: ErrYaEmpezada,
			mensaje: fmt.Sprintf("%s ya está %s: no se puede quitar del sprint", h.Identificador(), h.Estado)}
	}
	quitada, err := h.CambiarEstado(EstadoPendiente, Progreso{})
	if err != nil {
		return Historia{}, err
	}
	quitada.SprintID = nil
	return quitada, nil
}

// noSePuedeAsignar arma el error de una historia que no está lista: "HU-2 no se puede asignar: no está estimada".
func (h Historia) noSePuedeAsignar(base error, motivo string) error {
	return errorDeEstado{base: base, mensaje: h.Identificador() + " no se puede asignar: " + motivo}
}

// nombreDelSprint devuelve "el Sprint N" para el sprint con ese ID, u "otro sprint" si no está en la lista.
func nombreDelSprint(id int64, sprints []DatosSprint) string {
	for _, s := range sprints {
		if s.ID == id {
			return fmt.Sprintf("el Sprint %d", s.Numero)
		}
	}
	return "otro sprint"
}
