package server_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// Los tests de HU-13 usan el reloj de nuevoRouter (hoy es el 13/10/2026) y conSprintYBacklog: el Sprint 1, del 12/10
// al 18/10, es el activo, y la HU-1 (5 SP, un criterio sin cumplir) se puede asignar.

// conHistoriaEnElSprint asigna la HU-1 al Sprint 1.
func conHistoriaEnElSprint(t *testing.T) http.Handler {
	t.Helper()
	h := conSprintYBacklog(t)
	if rec := asignar(h, "1"); rec.Code != http.StatusSeeOther {
		t.Fatalf("no se pudo asignar la HU-1: código %d", rec.Code)
	}
	return h
}

func accionDelSprint(h http.Handler, accion string) *httptest.ResponseRecorder {
	return post(h, "/proyectos/1/sprints/1/historias/1/"+accion, nil)
}

func cumplirElCriterio(t *testing.T, h http.Handler) {
	t.Helper()
	if rec := post(h, "/proyectos/1/backlog/1/criterios/1/cumplido", url.Values{"cumplido": {"1"}}); rec.Code != http.StatusSeeOther {
		t.Fatalf("no se pudo cumplir el criterio: código %d", rec.Code)
	}
}

// RN2 · en el sprint activo, una historia En sprint tiene el botón Empezar y pasa a En progreso
func TestEmpezar_DesdeLaPaginaDelSprint(t *testing.T) {
	h := conHistoriaEnElSprint(t)
	verificarContiene(t, get(h, "/proyectos/1/sprints/1").Body.String(),
		`action="/proyectos/1/sprints/1/historias/1/empezar"`, "Empezar", "0 de 5 SP completados")

	rec := accionDelSprint(h, "empezar")

	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/proyectos/1/sprints/1?empezada=1" {
		t.Fatalf("se esperaba volver a /proyectos/1/sprints/1?empezada=1 y se obtuvo %d %q", rec.Code, rec.Header().Get("Location"))
	}
	verificarContiene(t, get(h, "/proyectos/1/sprints/1?empezada=1").Body.String(), "HU-1 pasó a En progreso", "En progreso")
}

// CA-13.3 · una historia En progreso con criterios sin cumplir muestra cuántos faltan en lugar del botón Completar
func TestCompletar_FaltanCriterios_MuestraCuantos(t *testing.T) {
	h := conHistoriaEnElSprint(t)
	accionDelSprint(h, "empezar")

	pagina := get(h, "/proyectos/1/sprints/1").Body.String()
	verificarContiene(t, pagina, "faltan 1 de 1 criterios")
	if strings.Contains(pagina, `/historias/1/completar"`) {
		t.Error("con criterios sin cumplir no tiene que estar el botón Completar")
	}

	rec := accionDelSprint(h, "completar")

	if rec.Code != http.StatusConflict {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusConflict)
	}
	verificarContiene(t, rec.Body.String(), "HU-1 no se puede completar: falta cumplir 1 de 1 criterios")
}

// CA-13.1 y CA-13.2 · completar vuelve a la página con el aviso, la historia Hecha con su fecha y el avance
func TestCompletar_HistoriaLista_VuelveALaPaginaDelSprint(t *testing.T) {
	h := conHistoriaEnElSprint(t)
	accionDelSprint(h, "empezar")
	cumplirElCriterio(t, h)
	verificarContiene(t, get(h, "/proyectos/1/sprints/1").Body.String(), `action="/proyectos/1/sprints/1/historias/1/completar"`)

	rec := accionDelSprint(h, "completar")

	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/proyectos/1/sprints/1?completada=1" {
		t.Fatalf("se esperaba volver a /proyectos/1/sprints/1?completada=1 y se obtuvo %d %q", rec.Code, rec.Header().Get("Location"))
	}
	verificarContiene(t, get(h, "/proyectos/1/sprints/1?completada=1").Body.String(),
		"HU-1 completada", "Hecho", "Completada el 13/10/2026", "5 de 5 SP completados")
}

// RN2 · una historia En sprint no se completa: primero hay que empezarla
func TestCompletar_SinEmpezar_MuestraElMotivo(t *testing.T) {
	h := conHistoriaEnElSprint(t)
	cumplirElCriterio(t, h)

	rec := accionDelSprint(h, "completar")

	if rec.Code != http.StatusConflict {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusConflict)
	}
	verificarContiene(t, rec.Body.String(), "HU-1 todavía no empezó: primero hay que pasarla a En progreso")
}

// CA-13.1 · en un sprint que no es el activo no hay botones para empezar ni completar, y completar se rechaza
func TestCompletar_SprintQueNoEsElActivo(t *testing.T) {
	h := conSprintYBacklog(t)
	futuro := url.Values{"objetivo": {"Segundo"}, "fecha_inicio": {"2026-10-19"}, "fecha_fin": {"2026-10-25"}}
	if rec := post(h, "/proyectos/1/sprints", futuro); rec.Code != http.StatusSeeOther {
		t.Fatalf("no se pudo crear el Sprint 2: código %d", rec.Code)
	}
	post(h, "/proyectos/1/sprints/2/historias", url.Values{"historia": {"1"}})
	post(h, "/proyectos/1/backlog/1/estado", url.Values{"estado": {"En progreso"}}) // desde la página de la historia (HU-08)
	cumplirElCriterio(t, h)

	pagina := get(h, "/proyectos/1/sprints/2").Body.String()
	if strings.Contains(pagina, "/historias/1/completar") || strings.Contains(pagina, "/historias/1/empezar") {
		t.Error("un sprint que no es el activo no tiene que tener los botones Empezar ni Completar")
	}
	rec := post(h, "/proyectos/1/sprints/2/historias/1/completar", nil)
	if rec.Code != http.StatusConflict {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusConflict)
	}
	verificarContiene(t, rec.Body.String(), "solo se completan historias del sprint activo")
}

// Sección 7 · una historia que no está en ese sprint da 404
func TestCompletarYEmpezar_HistoriaQueNoEstaEnElSprint(t *testing.T) {
	h := conSprintYBacklog(t)
	for _, accion := range []string{"empezar", "completar"} {
		if rec := accionDelSprint(h, accion); rec.Code != http.StatusNotFound {
			t.Errorf("%s: código = %d, se esperaba %d", accion, rec.Code, http.StatusNotFound)
		}
	}
}
