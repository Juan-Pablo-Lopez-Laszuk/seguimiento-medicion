package app_test

import (
	"net/http"
	"strings"
	"testing"

	neturl "net/url"
)

// verificarRevelar vota con Juan, el único participante, y comprueba que al ver la sesión la ronda se revela sola con
// su voto y que ya no se puede cerrar (HU-19, CA-19.1, CA-19.2 y RN5).
func verificarRevelar(t *testing.T, h http.Handler) {
	t.Helper()
	if rec := post(h, "/proyectos/1/backlog/1/poker/votos", neturl.Values{"quien": {"1"}, "valor": {"13"}}); rec.Code != http.StatusSeeOther {
		t.Fatalf("votar: se esperaba %d y se obtuvo %d:\n%s", http.StatusSeeOther, rec.Code, rec.Body.String())
	}
	pagina := get(h, "/proyectos/1/backlog/1/poker")
	for _, texto := range []string{"Ronda revelada.", "Votos de la ronda 1", "<td>Juan</td><td>13</td>"} {
		if !strings.Contains(pagina, texto) {
			t.Errorf("la página de la sesión tendría que mostrar %q:\n%s", texto, pagina)
		}
	}
	if rec := post(h, "/proyectos/1/backlog/1/poker/cerrar", neturl.Values{"quien": {"1"}}); rec.Code != http.StatusConflict {
		t.Errorf("cerrar una ronda revelada: se esperaba %d y se obtuvo %d", http.StatusConflict, rec.Code)
	}
}

// HU-19: sin base, la ronda se revela en memoria.
func TestNewHandlerConBase_SinURLRevelaLaRonda(t *testing.T) {
	h := nuevaApp(t, "")
	abrirPokerEnLaApp(t, h)
	verificarRevelar(t, h)
}

// HU-19: con PostgreSQL, la ronda revelada y los votos salen de la base.
func TestNewHandlerConBase_ConURLRevelaLaRonda(t *testing.T) {
	h := nuevaApp(t, urlConEsquemaMigrado(t))
	abrirPokerEnLaApp(t, h)
	verificarRevelar(t, h)
}
