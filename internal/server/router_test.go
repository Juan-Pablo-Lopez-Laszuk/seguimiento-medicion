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

// nuevoRouter arma la aplicación con repositorios en memoria vacíos.
func nuevoRouter() http.Handler {
	return server.NewRouter(server.Dependencias{
		Proyectos: service.NuevoProyectos(memory.NuevoProyectos(), time.Now),
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
