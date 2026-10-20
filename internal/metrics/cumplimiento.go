package metrics

import (
	"fmt"
	"strings"
)

// sinCumplimiento es lo que se muestra cuando no hay SP planificados (HU-15 RN3).
const sinCumplimiento = "—"

// Cumplimiento devuelve qué parte de los story points planificados se completó: SP completados ÷ SP planificados ×
// 100 (specs/HU-15-sprints-anteriores.md, RN2). Con 0 SP planificados no hay cumplimiento y devuelve 0, false: nunca
// se divide por cero (RN3).
func (s StoryPoints) Cumplimiento() (float64, bool) {
	if s.Planificados == 0 {
		return 0, false
	}
	return s.Completados / s.Planificados * 100, true
}

// CumplimientoTexto devuelve el cumplimiento como se muestra en pantalla: con 1 decimal y coma decimal ("38,5 %"), o
// "—" si no hay SP planificados.
func (s StoryPoints) CumplimientoTexto() string {
	valor, ok := s.Cumplimiento()
	if !ok {
		return sinCumplimiento
	}
	return strings.Replace(fmt.Sprintf("%.1f %%", valor), ".", ",", 1)
}
