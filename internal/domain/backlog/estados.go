package backlog

// Los otros dos estados de una historia (EstadoPendiente y EstadoHecho están en backlog.go). Los textos son los que
// guarda la base (CHECK de historias.estado).
//
// Spec: specs/HU-08-estado-historia.md.
const (
	EstadoEnSprint   Estado = "En sprint"
	EstadoEnProgreso Estado = "En progreso"
)

// Estados son todos los estados de una historia, en el orden del camino: Pendiente, En sprint, En progreso, Hecho (RN1).
var Estados = []Estado{EstadoPendiente, EstadoEnSprint, EstadoEnProgreso, EstadoHecho}

// CambiarEstado devuelve la historia con el estado nuevo, si el cambio vale; la original no cambia. Es la única
// regla de transición: HU-12 la usa para pasar a En sprint, HU-13 para pasar a Hecho y HU-14 para volver a Pendiente.
// progreso son los criterios de la historia (HU-07), que hacen falta para pasar a Hecho.
func (h Historia) CambiarEstado(nuevo Estado, progreso Progreso) (Historia, error) {
	h.Estado = nuevo
	return h, nil
}
