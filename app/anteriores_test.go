package app_test

import (
	"net/http"
	"strings"
	"testing"
)

// verificarAnterior revisa que la lista de sprints muestre el resultado del Sprint 1 cerrado (0 de 5 SP) y que su
// página muestre la foto del cierre con la HU-1 no completada (HU-15).
func verificarAnterior(t *testing.T, h http.Handler) {
	t.Helper()
	if lista := get(h, "/proyectos/1/sprints"); !strings.Contains(lista, "0 de 5") || !strings.Contains(lista, "0,0 %") {
		t.Errorf("la lista tendría que mostrar el Sprint 1 con 0 de 5 SP y 0,0 %% de cumplimiento:\n%s", lista)
	}
	detalle := get(h, "/proyectos/1/sprints/1")
	for _, esperado := range []string{"0 de 5 SP completados · 0,0 % de cumplimiento · 0 de 1 historias", "Historias al cerrar", "Historia del sprint", "No completada"} {
		if !strings.Contains(detalle, esperado) {
			t.Errorf("la página del Sprint 1 no muestra %q", esperado)
		}
	}
}

// HU-15: sin base, la lista y el detalle de un sprint cerrado funcionan en memoria.
func TestNewHandlerConBase_SinURLSeConsultaElSprintAnterior(t *testing.T) {
	h := nuevaApp(t, "")
	conSprintActivoHoy(t, h)
	if rec := post(h, "/proyectos/1/sprints/1/cerrar", nil); rec.Code != http.StatusSeeOther {
		t.Fatalf("cerrar: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}

	verificarAnterior(t, h)
}

// HU-15: con base, el resultado y la foto se leen de PostgreSQL después de un reinicio.
func TestNewHandlerConBase_ConURLElSprintAnteriorSobreviveAUnReinicio(t *testing.T) {
	url := urlConEsquemaMigrado(t)
	primera := nuevaApp(t, url)
	conSprintActivoHoy(t, primera)
	if rec := post(primera, "/proyectos/1/sprints/1/cerrar", nil); rec.Code != http.StatusSeeOther {
		t.Fatalf("cerrar: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}

	verificarAnterior(t, nuevaApp(t, url))
}
