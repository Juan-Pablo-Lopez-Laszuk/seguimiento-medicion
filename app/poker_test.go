package app_test

import (
	"net/http"
	neturl "net/url"
	"strings"
	"testing"
)

// abrirPokerEnLaApp crea el proyecto 1 con su Agile Enabler y la historia HU-1, y abre la sesión de Planning Poker de la
// historia (HU-17).
func abrirPokerEnLaApp(t *testing.T, h http.Handler) {
	t.Helper()
	crearProyecto(h, "Con Planning Poker")
	if rec := post(h, "/proyectos/1/integrantes", neturl.Values{"nombre": {"Juan"}, "email": {"juan@mail.com"}, "rol": {"AgileEnabler"}}); rec.Code != http.StatusSeeOther {
		t.Fatalf("registrar el integrante: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}
	if rec := post(h, "/proyectos/1/backlog", neturl.Values{"titulo": {"Login"}, "prioridad": {"Alta"}}); rec.Code != http.StatusSeeOther {
		t.Fatalf("crear la historia: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}
	if rec := post(h, "/proyectos/1/backlog/1/poker", neturl.Values{"quien": {"1"}}); rec.Code != http.StatusSeeOther {
		t.Fatalf("abrir la sesión: se esperaba %d y se obtuvo %d:\n%s", http.StatusSeeOther, rec.Code, rec.Body.String())
	}
}

// verificarPoker comprueba que la sesión quedó abierta en la Ronda 1 y que la historia ya no se puede eliminar
// (HU-06, RN10: la sesión es un registro asociado).
func verificarPoker(t *testing.T, h http.Handler) {
	t.Helper()
	if pagina := get(h, "/proyectos/1/backlog/1/poker"); !strings.Contains(pagina, "Ronda 1") || !strings.Contains(pagina, "Juan") {
		t.Errorf("la página de la sesión tendría que mostrar la Ronda 1 y a Juan:\n%s", pagina)
	}
	if rec := post(h, "/proyectos/1/backlog/1/poker", neturl.Values{"quien": {"1"}}); rec.Code != http.StatusConflict {
		t.Errorf("abrir otra vez: se esperaba %d y se obtuvo %d", http.StatusConflict, rec.Code)
	}
	if rec := post(h, "/proyectos/1/backlog/1/eliminar", neturl.Values{}); rec.Code != http.StatusConflict {
		t.Errorf("eliminar la historia con una sesión: se esperaba %d y se obtuvo %d", http.StatusConflict, rec.Code)
	}
}

// HU-17: sin base, Planning Poker funciona en memoria.
func TestNewHandlerConBase_SinURLAbrePlanningPoker(t *testing.T) {
	h := nuevaApp(t, "")
	abrirPokerEnLaApp(t, h)
	verificarPoker(t, h)
}

// HU-17: con PostgreSQL, la sesión se guarda en la base.
func TestNewHandlerConBase_ConURLAbrePlanningPoker(t *testing.T) {
	h := nuevaApp(t, urlConEsquemaMigrado(t))
	abrirPokerEnLaApp(t, h)
	verificarPoker(t, h)
}

// verificarVotos vota con Juan y comprueba que la página dice que votó, sin mostrar el valor (HU-18).
func verificarVotos(t *testing.T, h http.Handler) {
	t.Helper()
	if rec := post(h, "/proyectos/1/backlog/1/poker/votos", neturl.Values{"quien": {"1"}, "valor": {"13"}}); rec.Code != http.StatusSeeOther {
		t.Fatalf("votar: se esperaba %d y se obtuvo %d:\n%s", http.StatusSeeOther, rec.Code, rec.Body.String())
	}
	if fragmento := get(h, "/proyectos/1/backlog/1/poker/votos"); !strings.Contains(fragmento, "1 de 1 votaron") || strings.Contains(fragmento, "13") {
		t.Errorf("quién votó tendría que decir que Juan votó, sin el valor:\n%s", fragmento)
	}
}

// HU-18: sin base, se vota en memoria.
func TestNewHandlerConBase_SinURLVotaEnPlanningPoker(t *testing.T) {
	h := nuevaApp(t, "")
	abrirPokerEnLaApp(t, h)
	verificarVotos(t, h)
}

// HU-18: con PostgreSQL, el voto se guarda en la base.
func TestNewHandlerConBase_ConURLVotaEnPlanningPoker(t *testing.T) {
	h := nuevaApp(t, urlConEsquemaMigrado(t))
	abrirPokerEnLaApp(t, h)
	verificarVotos(t, h)
}
