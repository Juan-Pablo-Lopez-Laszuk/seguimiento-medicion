package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
)

// pedir simula un pedido GET a la aplicación y devuelve la respuesta grabada.
func pedir(t *testing.T, ruta string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	server.NewRouter().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, ruta, nil))
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
