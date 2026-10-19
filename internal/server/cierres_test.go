package server_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// Los tests de HU-14 usan el reloj de nuevoRouter (hoy es el 13/10/2026) y conHistoriaEnElSprint: el Sprint 1, del
// 12/10 al 18/10, es el activo y tiene la HU-1 (5 SP, un criterio sin cumplir) En sprint.

const rutaDeCierre = "/proyectos/1/sprints/1/cerrar"

// conHistoriaCompletada deja la HU-1 Hecha en el Sprint 1.
func conHistoriaCompletada(t *testing.T) http.Handler {
	t.Helper()
	h := conHistoriaEnElSprint(t)
	accionDelSprint(h, "empezar")
	cumplirElCriterio(t, h)
	if rec := accionDelSprint(h, "completar"); rec.Code != http.StatusSeeOther {
		t.Fatalf("no se pudo completar la HU-1: código %d", rec.Code)
	}
	return h
}

// conSprint2 crea el Sprint 2, del 19/10 al 25/10.
func conSprint2(t *testing.T, h http.Handler) {
	t.Helper()
	futuro := url.Values{"objetivo": {"Segundo"}, "fecha_inicio": {"2026-10-19"}, "fecha_fin": {"2026-10-25"}}
	if rec := post(h, "/proyectos/1/sprints", futuro); rec.Code != http.StatusSeeOther {
		t.Fatalf("no se pudo crear el Sprint 2: código %d", rec.Code)
	}
}

// Sección 3 · el sprint activo tiene el botón Cerrar sprint; un sprint futuro, no
func TestCerrarSprint_BotonSoloEnElSprintActivo(t *testing.T) {
	h := conHistoriaEnElSprint(t)
	conSprint2(t, h)

	verificarContiene(t, get(h, "/proyectos/1/sprints/1").Body.String(), `href="`+rutaDeCierre+`"`, "Cerrar sprint")
	if strings.Contains(get(h, "/proyectos/1/sprints/2").Body.String(), "/sprints/2/cerrar") {
		t.Error("un sprint que no es el activo no tiene que tener el botón Cerrar sprint")
	}
}

// Sección 3 · la confirmación dice qué va a pasar y no cierra nada
func TestConfirmacionDeCierre_MuestraLoQueVaAPasar(t *testing.T) {
	h := conHistoriaEnElSprint(t)
	accionDelSprint(h, "empezar")

	rec := get(h, rutaDeCierre)

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	verificarContiene(t, rec.Body.String(),
		"<title>Cerrar el Sprint 1 · Software Metrics</title>",
		"¿Cerrar el Sprint 1?", "0 de 5 SP", "0 de 1 historias",
		"Vuelven al backlog como Pendientes", "HU-1", "En progreso",
		"El sprint termina el 18/10/2026: lo estás cerrando 5 días antes",
		"Un sprint cerrado no se puede reabrir ni modificar",
		`action="`+rutaDeCierre+`" method="post"`, `href="/proyectos/1/sprints/1"`)
	verificarContiene(t, get(h, "/proyectos/1/sprints/1").Body.String(), `href="`+rutaDeCierre+`"`)
}

// CA-14.3 · cerrar vuelve a la página del sprint, que queda Cerrado con los SP del cierre y sin acciones
func TestCerrarSprint_VuelveALaPaginaDelSprintCerrado(t *testing.T) {
	h := conHistoriaCompletada(t)

	rec := post(h, rutaDeCierre, nil)

	destino := "/proyectos/1/sprints/1?cerrado=1&devueltas=0"
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != destino {
		t.Fatalf("se esperaba volver a %s y se obtuvo %d %q", destino, rec.Code, rec.Header().Get("Location"))
	}
	pagina := get(h, destino).Body.String()
	verificarContiene(t, pagina, "Sprint 1 cerrado: 5 de 5 SP completados.", "Cerrado", "5 de 5 SP completados")
	for _, accion := range []string{"/cerrar", "/empezar", "/completar", "/quitar"} {
		if strings.Contains(pagina, accion) {
			t.Errorf("un sprint cerrado no tiene que ofrecer %s", accion)
		}
	}
	verificarContiene(t, get(h, "/proyectos/1/sprints").Body.String(), "Cerrado")
}

// CA-14.2 · la historia no completada vuelve al backlog Pendiente, y el aviso lo dice
func TestCerrarSprint_LaNoCompletadaVuelveAlBacklog(t *testing.T) {
	h := conHistoriaEnElSprint(t)

	rec := post(h, rutaDeCierre, nil)

	destino := "/proyectos/1/sprints/1?cerrado=1&devueltas=1"
	if rec.Header().Get("Location") != destino {
		t.Fatalf("se esperaba volver a %s y se obtuvo %d %q", destino, rec.Code, rec.Header().Get("Location"))
	}
	verificarContiene(t, get(h, destino).Body.String(), "Sprint 1 cerrado: 0 de 5 SP completados. 1 historia volvió al backlog.")
	verificarContiene(t, get(h, "/proyectos/1/backlog/1").Body.String(), "Pendiente")
}

// CA-14.1 y CA-14.4 · sección 7 · un sprint que no es el activo o que ya está cerrado muestra el motivo sin el botón
func TestCerrarSprint_Errores(t *testing.T) {
	t.Run("sprint futuro", func(t *testing.T) {
		h := conHistoriaEnElSprint(t)
		conSprint2(t, h)
		for metodo, rec := range map[string]*httptest.ResponseRecorder{
			"GET":  get(h, "/proyectos/1/sprints/2/cerrar"),
			"POST": post(h, "/proyectos/1/sprints/2/cerrar", nil),
		} {
			if rec.Code != http.StatusConflict {
				t.Errorf("%s: código = %d, se esperaba %d", metodo, rec.Code, http.StatusConflict)
			}
			verificarContiene(t, rec.Body.String(), "solo se cierra el sprint activo (el Sprint 1)")
			if strings.Contains(rec.Body.String(), `action="/proyectos/1/sprints/2/cerrar"`) {
				t.Errorf("%s: si no se puede cerrar, la confirmación no tiene que ofrecer el botón", metodo)
			}
		}
	})

	t.Run("sprint ya cerrado", func(t *testing.T) {
		h := conHistoriaCompletada(t)
		post(h, rutaDeCierre, nil)
		rec := post(h, rutaDeCierre, nil)
		if rec.Code != http.StatusConflict {
			t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusConflict)
		}
		verificarContiene(t, rec.Body.String(), "el Sprint 1 ya está cerrado")
	})

	t.Run("sprint que no existe", func(t *testing.T) {
		h := conHistoriaEnElSprint(t)
		for _, ruta := range []string{"/proyectos/1/sprints/9/cerrar", "/proyectos/1/sprints/x/cerrar", "/proyectos/9/sprints/1/cerrar"} {
			if rec := get(h, ruta); rec.Code != http.StatusNotFound {
				t.Errorf("%s: código = %d, se esperaba %d", ruta, rec.Code, http.StatusNotFound)
			}
		}
	})
}

// Sección 3 · el aviso de cierre solo acepta números: cualquier otro valor en la URL se ignora
func TestCerrarSprint_AvisoSoloConNumeros(t *testing.T) {
	h := conHistoriaCompletada(t)
	post(h, rutaDeCierre, nil)

	pagina := get(h, "/proyectos/1/sprints/1?cerrado=1&devueltas=<script>").Body.String()

	verificarContiene(t, pagina, "Sprint 1 cerrado: 5 de 5 SP completados.")
	if strings.Contains(pagina, "volvi") {
		t.Error("con devueltas que no es un número no tiene que decir nada de las historias devueltas")
	}
}
