package app_test

import (
	"net/http"
	neturl "net/url"
	"strings"
	"testing"
)

// corregirEnLaApp crea el proyecto 1 con la HU-1 y Carolina, le registra 2,5 h el primer día del proyecto (ya pasó y
// está dentro de sus fechas) y Carolina lo corrige a 3 h (HU-26).
func corregirEnLaApp(t *testing.T, h http.Handler) {
	t.Helper()
	crearProyecto(h, "Con registros")
	post(h, "/proyectos/1/backlog", neturl.Values{"titulo": {"Historia con horas"}, "prioridad": {"Alta"}})
	post(h, "/proyectos/1/integrantes", neturl.Values{"nombre": {"Carolina"}, "email": {"caro@mail.com"}, "rol": {"ProductBuilder"}})
	campos := neturl.Values{"integrante": {"1"}, "fecha": {"2026-10-05"}, "actividad": {"Registro a corregir"}, "horas": {"2,5"}}
	if rec := post(h, "/proyectos/1/backlog/1/esfuerzo", campos); rec.Code != http.StatusSeeOther {
		t.Fatalf("registrar esfuerzo: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}
	if lista := get(h, "/proyectos/1/esfuerzo?integrante=1"); !strings.Contains(lista, "Registro a corregir") {
		t.Fatalf("la lista de registros debería mostrar el registro:\n%s", lista)
	}

	correccion := neturl.Values{"quien": {"1"}, "fecha": {"2026-10-05"}, "actividad": {"Registro corregido"}, "horas": {"3"}}
	if rec := post(h, "/proyectos/1/esfuerzo/1/corregir", correccion); rec.Code != http.StatusSeeOther {
		t.Fatalf("corregir: se esperaba %d y se obtuvo %d:\n%s", http.StatusSeeOther, rec.Code, rec.Body.String())
	}
}

func verificarCorregido(t *testing.T, h http.Handler) {
	t.Helper()
	lista := get(h, "/proyectos/1/esfuerzo")
	for _, esperado := range []string{"Registro corregido", "Carolina", "Total de la lista: <strong>3 h</strong>"} {
		if !strings.Contains(lista, esperado) {
			t.Errorf("la lista de registros debería mostrar %q", esperado)
		}
	}
}

// HU-26: sin base, la consulta y la corrección de registros funcionan en memoria.
func TestNewHandlerConBase_SinURLSeCorrigeUnRegistro(t *testing.T) {
	h := nuevaApp(t, "")
	corregirEnLaApp(t, h)
	verificarCorregido(t, h)
}

// HU-26: la corrección queda en la base y sobrevive a un reinicio.
func TestNewHandlerConBase_ConURLLaCorreccionSobreviveAUnReinicio(t *testing.T) {
	url := urlConEsquemaMigrado(t)
	corregirEnLaApp(t, nuevaApp(t, url))
	verificarCorregido(t, nuevaApp(t, url))
}
