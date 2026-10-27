package app_test

import (
	"context"
	"net/http"
	neturl "net/url"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// verificarAcordarSinRevelar comprueba que registrar la estimación está conectado en la app: con la Ronda 1 sin revelar
// (revelar es de HU-19) la página lo avisa y registrar da 409, no 500; y la página ya tiene la tabla de sesiones
// cerradas (HU-22).
func verificarAcordarSinRevelar(t *testing.T, h http.Handler) {
	t.Helper()
	pagina := get(h, "/proyectos/1/backlog/1/poker")
	for _, texto := range []string{"Para registrar la estimación, primero revelá la ronda.", "Sesiones cerradas", "Todavía no se cerró ninguna sesión."} {
		if !strings.Contains(pagina, texto) {
			t.Errorf("la página de la sesión tendría que mostrar %q", texto)
		}
	}
	if rec := post(h, "/proyectos/1/backlog/1/poker/acordar", neturl.Values{"quien": {"1"}, "valor": {"5"}}); rec.Code != http.StatusConflict {
		t.Errorf("registrar sin revelar: se esperaba %d y se obtuvo %d:\n%s", http.StatusConflict, rec.Code, rec.Body.String())
	}
}

// HU-22: sin base, registrar la estimación está conectado en memoria.
func TestNewHandlerConBase_SinURLRegistrarEstimacion(t *testing.T) {
	h := nuevaApp(t, "")
	abrirPokerEnLaApp(t, h)
	verificarAcordarSinRevelar(t, h)
}

// HU-22: con PostgreSQL, con la ronda revelada (directo en la base, hasta que llegue HU-19) el Agile Enabler registra
// el valor: la historia queda con esos story points y la sesión cerrada, también después de un reinicio.
func TestNewHandlerConBase_ConURLRegistrarEstimacion(t *testing.T) {
	url := urlConEsquemaMigrado(t)
	h := nuevaApp(t, url)
	abrirPokerEnLaApp(t, h)
	verificarAcordarSinRevelar(t, h)

	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if _, err := pool.Exec(context.Background(), `UPDATE rondas_poker SET revelada = true`); err != nil {
		t.Fatalf("revelar la ronda: %v", err)
	}

	rec := post(h, "/proyectos/1/backlog/1/poker/acordar", neturl.Values{"quien": {"1"}, "valor": {"5"}})
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("registrar: se esperaba %d y se obtuvo %d:\n%s", http.StatusSeeOther, rec.Code, rec.Body.String())
	}
	reiniciada := nuevaApp(t, url)
	pagina := get(reiniciada, "/proyectos/1/backlog/1/poker")
	for _, texto := range []string{"Story points actuales: 5", "Sesión 1", "5 SP", "Abrir sesión"} {
		if !strings.Contains(pagina, texto) {
			t.Errorf("después de registrar y reiniciar, la página tendría que mostrar %q", texto)
		}
	}
}
