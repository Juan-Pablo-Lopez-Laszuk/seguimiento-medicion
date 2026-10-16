package backlog

import "strings"

// Cambios son los datos que se corrigen de una historia que ya existe (HU-06): los mismos que se cargan al crearla,
// menos los story points (HU-16) y el estado (HU-08), que tienen su propia pantalla.
type Cambios struct {
	Titulo      string
	Descripcion string
	Prioridad   Prioridad
}

// Editar devuelve la historia con los datos corregidos. No cambia el identificador (CA-06.1), el estado, los story
// points ni lo que tenga asociado (RN2), y se puede en cualquier estado (RN3). La historia original no se modifica.
func (h Historia) Editar(c Cambios) (Historia, error) {
	h.Titulo = strings.TrimSpace(c.Titulo)
	h.Descripcion = strings.TrimSpace(c.Descripcion)
	h.Prioridad = c.Prioridad
	return h, nil
}
