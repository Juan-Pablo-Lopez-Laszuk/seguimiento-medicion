package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// relojDeLaVista fija "hoy" en el 13/10/2026 a las 15:00 de Argentina, así la vista del proyecto (HU-04) no
// depende del día en que corren los tests.
func relojDeLaVista() time.Time { return time.Date(2026, 10, 13, 18, 0, 0, 0, time.UTC) }

// nuevoRouter arma la aplicación con repositorios en memoria vacíos y "hoy" en el 13/10/2026.
func nuevoRouter() http.Handler {
	return nuevoRouterConReloj(relojDeLaVista)
}

// nuevoRouterConReloj arma la aplicación con repositorios en memoria vacíos; ahora fija "hoy" para la vista del
// proyecto (HU-04) y para la modificación (HU-02).
func nuevoRouterConReloj(ahora func() time.Time) http.Handler {
	proyectos, sprints, historias := memory.NuevoProyectos(), memory.NuevoSprints(), memory.NuevoHistorias()
	integrantes := memory.NuevoIntegrantes()
	repoCriterios, repoTareas := memory.NuevoCriterios(), memory.NuevoTareas()
	criterios := service.NuevoCriterios(proyectos, historias, repoCriterios)
	return server.NewRouter(server.Dependencias{
		Proyectos:   service.NuevoProyectos(proyectos, time.Now),
		Sprints:     service.NuevoSprints(proyectos, sprints),
		Integrantes: service.NuevoIntegrantes(proyectos, integrantes),
		Historias:   service.NuevoHistorias(proyectos, historias),
		Criterios:   criterios,
		// la página de la historia pasa a En progreso (HU-08), por ejemplo en los tests de HU-13
		CambioDeEstado: service.NuevoEstados(proyectos, historias, repoCriterios),
		Asignaciones:   service.NuevoAsignaciones(proyectos, sprints, historias, criterios, ahora),
		Cierres:        service.NuevoCierres(proyectos, sprints, historias, memory.NuevoCierres(sprints, historias), ahora),
		Tareas:         service.NuevoTareas(proyectos, historias, integrantes, repoTareas),
		Esfuerzo:       service.NuevoEsfuerzo(proyectos, historias, integrantes, repoTareas, memory.NuevoRegistros(), ahora),
		Defectos:       service.NuevoDefectos(proyectos, sprints, historias, memory.NuevoDefectos()),
		Estado:         service.NuevoEstadoProyectos(proyectos, sprints, historias, ahora),
		Edicion:        service.NuevoEdicionProyectos(proyectos, sprints, ahora),
		Estimaciones:   service.NuevoEstimaciones(proyectos, historias),
	})
}

// pedir simula un pedido GET a la aplicación y devuelve la respuesta grabada.
func pedir(t *testing.T, ruta string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	nuevoRouter().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, ruta, nil))
	return rec
}

func TestHealth_RespondeOKEnJSON(t *testing.T) {
	rec := pedir(t, "/health")

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Content-Type = %q, se esperaba application/json", ct)
	}
	var cuerpo map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &cuerpo); err != nil {
		t.Fatalf("la respuesta no es JSON válido: %v", err)
	}
	if cuerpo["status"] != "ok" {
		t.Errorf("status = %q, se esperaba \"ok\"", cuerpo["status"])
	}
}

func TestInicio_MuestraLayoutBase(t *testing.T) {
	rec := pedir(t, "/")

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("Content-Type = %q, se esperaba text/html", ct)
	}
	html := rec.Body.String()
	for _, esperado := range []string{
		"<title>Inicio · Software Metrics</title>", // título armado por el layout base
		"bootstrap.min.css",                        // estilos
		"htmx.min.js",                              // pantallas dinámicas
		"Software Metrics &amp; Estimation",        // contenido propio de la página de inicio
	} {
		if !strings.Contains(html, esperado) {
			t.Errorf("la página no contiene %q", esperado)
		}
	}
}

func TestRutaInexistente_Responde404(t *testing.T) {
	if rec := pedir(t, "/no-existe"); rec.Code != http.StatusNotFound {
		t.Errorf("código = %d, se esperaba %d", rec.Code, http.StatusNotFound)
	}
}
