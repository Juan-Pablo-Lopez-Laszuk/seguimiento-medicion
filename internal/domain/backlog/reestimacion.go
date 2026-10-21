package backlog

import "time"

// Reestimacion es una entrada del historial de estimaciones de una historia (HU-16b): cuánto valía su estimación antes,
// cuánto vale ahora y cuándo cambió. El historial solo crece: las entradas no se editan (RN6).
type Reestimacion struct {
	Anterior   int       // story points que tenía la historia
	Nuevo      int       // story points que tiene ahora
	CambiadaEn time.Time // cuándo se re-estimó
}

// NuevaReestimacion devuelve la entrada del historial para un cambio de estimación, o false si el cambio no se anota.
// Se anota cuando la historia ya tenía una estimación y el valor nuevo es distinto (RN1); 0 cuenta como estimación
// (RN4). No se anota la primera estimación, porque no hay valor anterior que conservar (RN2), ni pedir el mismo valor
// (RN3). anterior es nil si la historia estaba sin estimar. Los dos repositorios la usan al guardar la estimación, así la
// regla está en un solo lugar.
//
// Spec: specs/HU-16b-historial-estimaciones.md.
func NuevaReestimacion(anterior *int, nuevo int, cuando time.Time) (Reestimacion, bool) {
	if anterior == nil || *anterior == nuevo {
		return Reestimacion{}, false
	}
	return Reestimacion{Anterior: *anterior, Nuevo: nuevo, CambiadaEn: cuando}, true
}
