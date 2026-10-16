package server_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// conSprintYBacklog arma la aplicación con el proyecto de ejemplo, el Sprint 1 (del 12/10 al 18/10) y tres historias:
// HU-1 lista (5 SP y un criterio), HU-2 sin estimar y HU-3 sin criterios.
func conSprintYBacklog(t *testing.T) http.Handler {
	t.Helper()
	h := conSprint(t, "2026-10-12", "2026-10-18")
	sinEstimar := historiaValida()
	sinEstimar.Set("titulo", "Registrarse")
	sinEstimar.Set("story_points", "")
	sinCriterios := historiaValida()
	sinCriterios.Set("titulo", "Editar el perfil")
	sinCriterios.Set("story_points", "3")
	for _, campos := range []url.Values{historiaValida(), sinEstimar, sinCriterios} {
		if rec := post(h, "/proyectos/1/backlog", campos); rec.Code != http.StatusSeeOther {
			t.Fatalf("no se pudo crear la historia: código %d", rec.Code)
		}
	}
	for _, numero := range []string{"1", "2"} {
		if rec := post(h, "/proyectos/1/backlog/"+numero+"/criterios", criterioValido("Entra con su cuenta")); rec.Code != http.StatusSeeOther {
			t.Fatalf("no se pudo agregar el criterio: código %d", rec.Code)
		}
	}
	return h
}

func asignar(h http.Handler, numero string) *httptest.ResponseRecorder {
	return post(h, "/proyectos/1/sprints/1/historias", url.Values{"historia": {numero}})
}

// Sección 3 · la página del sprint: su goal, lo comprometido y las candidatas, con el botón o el motivo
func TestPaginaDelSprint_MuestraElSprintYLasCandidatas(t *testing.T) {
	rec := get(conSprintYBacklog(t), "/proyectos/1/sprints/1")

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	verificarContiene(t, rec.Body.String(),
		"<title>Sprint 1 · Software Metrics</title>",
		"Primer MVP", "12/10/2026", "18/10/2026",
		"0 SP comprometidos",
		`action="/proyectos/1/sprints/1/historias"`, `name="historia" value="1"`, "Asignar",
		"HU-2", "sin estimar",
		"HU-3", "sin criterios de aceptación",
	)
	if strings.Contains(rec.Body.String(), `name="historia" value="2"`) {
		t.Error("una historia sin estimar no tiene que tener el botón Asignar")
	}
}

// CA-12.1 · asignar vuelve a la página con el aviso, la historia en el sprint y el botón para quitarla
func TestAsignar_HistoriaLista_VuelveALaPaginaDelSprint(t *testing.T) {
	h := conSprintYBacklog(t)

	rec := asignar(h, "1")

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("código = %d, se esperaba %d (redirección)", rec.Code, http.StatusSeeOther)
	}
	destino := rec.Header().Get("Location")
	if destino != "/proyectos/1/sprints/1?asignada=1" {
		t.Fatalf("redirige a %q, se esperaba /proyectos/1/sprints/1?asignada=1", destino)
	}
	verificarContiene(t, get(h, destino).Body.String(),
		"HU-1 asignada al Sprint 1", "5 SP comprometidos", "En sprint",
		`action="/proyectos/1/sprints/1/historias/1/quitar"`,
	)
	verificarContiene(t, get(h, "/proyectos/1/backlog/1").Body.String(), "En sprint")
}

// Sección 7 · lo que no se puede asignar se avisa en la página del sprint (409) y no cambia nada
func TestAsignar_HistoriaNoLista_MuestraElMotivo(t *testing.T) {
	h := conSprintYBacklog(t)

	rec := asignar(h, "2")

	if rec.Code != http.StatusConflict {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusConflict)
	}
	verificarContiene(t, rec.Body.String(), "HU-2 no se puede asignar: no está estimada", "0 SP comprometidos")
}

// RN4 · quitar vuelve a la página del sprint con el aviso y la historia otra vez entre las candidatas
func TestQuitar_VuelveALaPaginaDelSprint(t *testing.T) {
	h := conSprintYBacklog(t)
	asignar(h, "1")

	rec := post(h, "/proyectos/1/sprints/1/historias/1/quitar", nil)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("código = %d, se esperaba %d (redirección)", rec.Code, http.StatusSeeOther)
	}
	destino := rec.Header().Get("Location")
	if destino != "/proyectos/1/sprints/1?quitada=1" {
		t.Fatalf("redirige a %q, se esperaba /proyectos/1/sprints/1?quitada=1", destino)
	}
	verificarContiene(t, get(h, destino).Body.String(), "HU-1 quitada del Sprint 1", "0 SP comprometidos", `name="historia" value="1"`)
}

// RN4 · una historia que no está en ese sprint no se puede quitar de él
func TestQuitar_HistoriaQueNoEstaEnElSprint_Responde404(t *testing.T) {
	if rec := post(conSprintYBacklog(t), "/proyectos/1/sprints/1/historias/1/quitar", nil); rec.Code != http.StatusNotFound {
		t.Errorf("código = %d, se esperaba %d", rec.Code, http.StatusNotFound)
	}
}

// Sección 7 · proyecto, sprint o historia que no existen (o que no son un número) dan 404
func TestPaginaDelSprint_NoEncontrados(t *testing.T) {
	h := conSprintYBacklog(t)
	for _, ruta := range []string{"/proyectos/9/sprints/1", "/proyectos/1/sprints/9", "/proyectos/1/sprints/abc"} {
		if rec := get(h, ruta); rec.Code != http.StatusNotFound {
			t.Errorf("%s: código = %d, se esperaba %d", ruta, rec.Code, http.StatusNotFound)
		}
	}
	for _, numero := range []string{"99", "abc"} {
		if rec := asignar(h, numero); rec.Code != http.StatusNotFound {
			t.Errorf("asignar la historia %q: código = %d, se esperaba %d", numero, rec.Code, http.StatusNotFound)
		}
	}
}

// La página solo muestra los avisos que pone el servidor: lo que venga en la URL que no sea un número se ignora
func TestPaginaDelSprint_IgnoraAvisosInventados(t *testing.T) {
	rec := get(conSprintYBacklog(t), "/proyectos/1/sprints/1?asignada=%3Cb%3Ehola")

	if strings.Contains(rec.Body.String(), "hola") || strings.Contains(rec.Body.String(), "asignada al Sprint") {
		t.Error("la página muestra un aviso armado con lo que vino en la URL")
	}
}

// Sección 3 · en la lista de sprints, cada sprint enlaza a su página
func TestListaDeSprints_CadaSprintEnlazaASuPagina(t *testing.T) {
	verificarContiene(t, get(conSprintYBacklog(t), "/proyectos/1/sprints").Body.String(), `href="/proyectos/1/sprints/1"`)
}

// Si falla la base se muestra un mensaje general, sin detalles internos
func TestPaginaDelSprint_ErrorDeLaBase(t *testing.T) {
	proyectos, historias := memory.NuevoProyectos(), memory.NuevoHistorias()
	criterios := service.NuevoCriterios(proyectos, historias, memory.NuevoCriterios())
	h := server.NewRouter(server.Dependencias{
		Asignaciones: service.NuevoAsignaciones(sprintsCaidos{}, sprintsCaidos{}, historias, criterios, relojDeLaVista),
	})

	rec := get(h, "/proyectos/1/sprints/1")

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("código = %d, se esperaba %d", rec.Code, http.StatusInternalServerError)
	}
	if strings.Contains(rec.Body.String(), "detalle interno") {
		t.Error("la página muestra el error interno")
	}
}
