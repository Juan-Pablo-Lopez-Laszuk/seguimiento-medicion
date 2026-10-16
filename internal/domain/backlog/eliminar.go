package backlog

import "errors"

// Motivos por los que una historia no se elimina (HU-06). Los dos primeros los decide el dominio; el tercero lo
// informa el repositorio cuando la base rechaza el borrado.
var (
	// ErrEliminarHistoriaHecha: una historia Hecha ya es parte de lo que el equipo terminó (CA-06.2).
	ErrEliminarHistoriaHecha = errors.New("la historia está Hecha: no se puede eliminar")
	// ErrEliminarEnSprintCerrado: la historia participó de un sprint que ya cerró, y sus métricas dependen de ella (CA-06.2).
	ErrEliminarEnSprintCerrado = errors.New("la historia es de un sprint cerrado: no se puede eliminar")
	// ErrHistoriaConRegistros: la historia tiene otros registros asociados (sesiones de Planning Poker, esfuerzo o
	// defectos) y la base no deja borrarla (RN10).
	ErrHistoriaConRegistros = errors.New("la historia tiene registros asociados (Planning Poker, esfuerzo o defectos): no se puede eliminar")
)

// PuedeEliminarse devuelve nil si la historia se puede eliminar, o el motivo por el que no. enSprintCerrado dice si la
// historia es de un sprint cerrado: su sprint actual o la foto de cierre lo dicen, y lo sabe el repositorio, no la
// historia. Una Hecha no se elimina (RN5), y eso se informa antes que lo del sprint; una de un sprint cerrado tampoco
// (RN6). Pendiente, En sprint y En progreso sí (RN7).
func (h Historia) PuedeEliminarse(enSprintCerrado bool) error {
	switch {
	case h.Estado == EstadoHecho:
		return ErrEliminarHistoriaHecha
	case enSprintCerrado:
		return ErrEliminarEnSprintCerrado
	}
	return nil
}
