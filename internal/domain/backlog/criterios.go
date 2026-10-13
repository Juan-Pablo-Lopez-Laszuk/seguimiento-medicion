package backlog

import "strings"

// Criterio es un criterio de aceptación de una historia: algo que se verifica para dar la historia por terminada.
//
// Spec: specs/HU-07-criterios-aceptacion.md.
type Criterio struct {
	ID          int64
	HistoriaID  int64
	Descripcion string
	Cumplido    bool
}

// NuevoCriterio devuelve el criterio listo para guardar en la historia. Todo criterio nace no cumplido (RN2) y se
// guarda sin los espacios del inicio y del final de la descripción (RN1).
func NuevoCriterio(descripcion string, h Historia) (Criterio, error) {
	return Criterio{HistoriaID: h.ID, Descripcion: strings.TrimSpace(descripcion)}, nil
}
