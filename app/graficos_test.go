package app_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	neturl "net/url"
	"strings"
	"testing"
)

// graficosDeLaApp crea el proyecto 1 con un sprint y pide los datos de sus gráficos en JSON (HU-36, CA-36.2).
func graficosDeLaApp(t *testing.T, h http.Handler) map[string]any {
	t.Helper()
	crearProyecto(h, "Con gráficos")
	if rec := post(h, "/proyectos/1/sprints", neturl.Values{"objetivo": {"Primer MVP"}, "fecha_inicio": {"2026-10-05"}, "fecha_fin": {"2026-10-11"}}); rec.Code != http.StatusSeeOther {
		t.Fatalf("crear el sprint: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/proyectos/1/graficos/datos", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("datos de los gráficos: se esperaba %d y se obtuvo %d:\n%s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var d map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &d); err != nil {
		t.Fatalf("la respuesta no es JSON: %v", err)
	}
	return d
}

func verificarGraficos(t *testing.T, h http.Handler, d map[string]any) {
	t.Helper()
	b, ok := d["burndown"].(map[string]any)
	if !ok || b["sprint"] != "Sprint 1" {
		t.Errorf("se esperaba el burndown del Sprint 1 y se obtuvo %v", d["burndown"])
	}
	if pagina := get(h, "/proyectos/1/graficos"); !strings.Contains(pagina, `id="grafico-burndown"`) {
		t.Error("la página de gráficos tendría que tener el burndown")
	}
}

// HU-36: sin base, los gráficos funcionan en memoria.
func TestNewHandlerConBase_SinURLMuestraLosGraficos(t *testing.T) {
	h := nuevaApp(t, "")
	verificarGraficos(t, h, graficosDeLaApp(t, h))
}

// HU-36: con PostgreSQL, los gráficos leen la base.
func TestNewHandlerConBase_ConURLLosGraficosLeenLaBase(t *testing.T) {
	h := nuevaApp(t, urlConEsquemaMigrado(t))
	verificarGraficos(t, h, graficosDeLaApp(t, h))
}
