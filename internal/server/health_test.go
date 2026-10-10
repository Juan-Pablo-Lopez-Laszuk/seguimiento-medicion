package server_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// baseFalsa simula la base de datos: responde con err (nil = responde bien) y tarda demora en contestar.
// Para no esperar de verdad, si el pedido tiene un plazo menor que la demora devuelve el error de plazo vencido.
type baseFalsa struct {
	err    error
	demora time.Duration
}

func (b baseFalsa) Ping(ctx context.Context) error {
	if limite, ok := ctx.Deadline(); ok && time.Until(limite) < b.demora {
		return context.DeadlineExceeded
	}
	return b.err
}

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

// CA-04.3: con base, /health le hace ping y si responde informa ok.
func TestHealth_ConBaseQueRespondeInformaOK(t *testing.T) {
	codigo, cuerpo := pedirSalud(t, baseFalsa{})

	if codigo != http.StatusOK {
		t.Errorf("código = %d, se esperaba %d", codigo, http.StatusOK)
	}
	if cuerpo["status"] != "ok" || cuerpo["base"] != "ok" {
		t.Errorf("respuesta = %v, se esperaba status ok y base ok", cuerpo)
	}
}
