package app_test

import (
	"net/http"
	"strings"
	"testing"
)

// verificarReporte revisa que el reporte del Sprint 1 cerrado muestre sus métricas y la HU-1 de la foto (HU-37).
func verificarReporte(t *testing.T, h http.Handler) {
	t.Helper()
	reporte := get(h, "/proyectos/1/sprints/1/reporte")
	for _, esperado := range []string{"Reporte del Sprint 1", "0 de 5 SP", "0,0 %", "0 de 1 (0,0 %)",
		"0,0 SP por sprint (1 sprint cerrado)", "0 h estimadas · 0 h reales", "Historia del sprint", "No completada"} {
		if !strings.Contains(reporte, esperado) {
			t.Errorf("el reporte del Sprint 1 no muestra %q", esperado)
		}
	}
}

// HU-37: sin base, el reporte de un sprint cerrado funciona en memoria.
func TestNewHandlerConBase_SinURLSeVeElReporteDelSprint(t *testing.T) {
	h := nuevaApp(t, "")
	conSprintActivoHoy(t, h)
	if rec := post(h, "/proyectos/1/sprints/1/cerrar", nil); rec.Code != http.StatusSeeOther {
		t.Fatalf("cerrar: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}

	verificarReporte(t, h)
}

// HU-37: con base, el reporte se arma con lo que está en PostgreSQL después de un reinicio.
func TestNewHandlerConBase_ConURLElReporteSobreviveAUnReinicio(t *testing.T) {
	url := urlConEsquemaMigrado(t)
	primera := nuevaApp(t, url)
	conSprintActivoHoy(t, primera)
	if rec := post(primera, "/proyectos/1/sprints/1/cerrar", nil); rec.Code != http.StatusSeeOther {
		t.Fatalf("cerrar: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}

	verificarReporte(t, nuevaApp(t, url))
}

// verificarReporteDelProyecto revisa que el reporte del proyecto tenga el Sprint 1 cerrado en la evolución y su bloque
// con lo mismo que su reporte (HU-38).
func verificarReporteDelProyecto(t *testing.T, h http.Handler) {
	t.Helper()
	reporte := get(h, "/proyectos/1/reporte")
	for _, esperado := range []string{"Reporte del proyecto", "1 cerrado de 1", "0,0 SP por sprint (1 sprint cerrado)",
		"Evolución de la velocidad", "Sprint 1", `id="sprint-1"`, "0 de 5 SP", "No completada"} {
		if !strings.Contains(reporte, esperado) {
			t.Errorf("el reporte del proyecto no muestra %q", esperado)
		}
	}
}

// HU-38: sin base, el reporte del proyecto funciona en memoria.
func TestNewHandlerConBase_SinURLSeVeElReporteDelProyecto(t *testing.T) {
	h := nuevaApp(t, "")
	conSprintActivoHoy(t, h)
	if rec := post(h, "/proyectos/1/sprints/1/cerrar", nil); rec.Code != http.StatusSeeOther {
		t.Fatalf("cerrar: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}

	verificarReporteDelProyecto(t, h)
}

// HU-38: con base, el reporte del proyecto se arma con lo que está en PostgreSQL después de un reinicio.
func TestNewHandlerConBase_ConURLElReporteDelProyectoSobreviveAUnReinicio(t *testing.T) {
	url := urlConEsquemaMigrado(t)
	primera := nuevaApp(t, url)
	conSprintActivoHoy(t, primera)
	if rec := post(primera, "/proyectos/1/sprints/1/cerrar", nil); rec.Code != http.StatusSeeOther {
		t.Fatalf("cerrar: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}

	verificarReporteDelProyecto(t, nuevaApp(t, url))
}
