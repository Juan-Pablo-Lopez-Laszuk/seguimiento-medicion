package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// pedirSalud arma la app con la base indicada (nil = sin base, modo memoria), pide /health y devuelve el
// código y el JSON de la respuesta.
func pedirSalud(t *testing.T, base server.VerificadorBase) (int, map[string]string) {
	t.Helper()
	proyectos := memory.NuevoProyectos()
	h := server.NewRouter(server.Dependencias{
		Proyectos:   service.NuevoProyectos(proyectos, time.Now),
		Sprints:     service.NuevoSprints(proyectos, memory.NuevoSprints()),
		Integrantes: service.NuevoIntegrantes(proyectos, memory.NuevoIntegrantes()),
		Base:        base,
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))

	var cuerpo map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &cuerpo); err != nil {
		t.Fatalf("la respuesta no es JSON válido: %v\n%s", err, rec.Body.String())
	}
	return rec.Code, cuerpo
}

// RN1: sin DATABASE_URL la app guarda en memoria; /health sigue en 200 pero lo avisa.
func TestHealth_SinBaseInformaMemoria(t *testing.T) {
	codigo, cuerpo := pedirSalud(t, nil)

	if codigo != http.StatusOK {
		t.Errorf("código = %d, se esperaba %d", codigo, http.StatusOK)
	}
	if cuerpo["status"] != "ok" || cuerpo["base"] != "memoria" {
		t.Errorf("respuesta = %v, se esperaba status ok y base memoria", cuerpo)
	}
}
