package features

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// baseSimulada hace de base de datos para los escenarios de /health: responde con err y tarda demora.
// No espera de verdad: si el pedido tiene un plazo menor que la demora, devuelve el error de plazo vencido.
type baseSimulada struct {
	err    error
	demora time.Duration
}

func (b baseSimulada) Ping(ctx context.Context) error {
	if limite, ok := ctx.Deadline(); ok && time.Until(limite) < b.demora {
		return context.DeadlineExceeded
	}
	return b.err
}

// tec04 guarda el estado de un escenario de TEC-04 (spec specs/TEC-04-deploy-vercel.md).
type tec04 struct {
	base      server.VerificadorBase
	codigo    int
	respuesta map[string]string
}

func registrarPasosTEC04(sc *godog.ScenarioContext) {
	e := &tec04{}
	sc.Step(`^que la app corre sin base de datos$`, e.sinBase)
	sc.Step(`^que la app tiene una base de datos que responde$`, e.baseQueResponde)
	sc.Step(`^que la app tiene una base de datos que no responde$`, e.baseQueNoResponde)
	sc.Step(`^que la app tiene una base de datos que tarda (\d+) segundos en responder$`, e.baseQueTarda)
	sc.Step(`^consulto el estado de la app$`, e.consultarEstado)
	sc.Step(`^la respuesta de salud tiene el código (\d+)$`, e.codigoEs)
	sc.Step(`^el estado general es "([^"]*)"$`, e.campoEs("status"))
	sc.Step(`^el estado de la base es "([^"]*)"$`, e.campoEs("base"))
}

func (e *tec04) sinBase() error {
	e.base = nil
	return nil
}

func (e *tec04) baseQueResponde() error {
	e.base = baseSimulada{}
	return nil
}

func (e *tec04) baseQueNoResponde() error {
	e.base = baseSimulada{err: errors.New("connection refused")}
	return nil
}

func (e *tec04) baseQueTarda(segundos int) error {
	e.base = baseSimulada{demora: time.Duration(segundos) * time.Second}
	return nil
}

func (e *tec04) consultarEstado() error {
	proyectos := memory.NuevoProyectos()
	h := server.NewRouter(server.Dependencias{
		Proyectos:   service.NuevoProyectos(proyectos, time.Now),
		Sprints:     service.NuevoSprints(proyectos, memory.NuevoSprints()),
		Integrantes: service.NuevoIntegrantes(proyectos, memory.NuevoIntegrantes()),
		Base:        e.base,
	})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	e.codigo = rec.Code
	e.respuesta = map[string]string{}
	if err := json.Unmarshal(rec.Body.Bytes(), &e.respuesta); err != nil {
		return fmt.Errorf("/health no respondió JSON: %w", err)
	}
	return nil
}

func (e *tec04) codigoEs(esperado string) error {
	if strconv.Itoa(e.codigo) != esperado {
		return fmt.Errorf("se esperaba el código %s y se obtuvo %d", esperado, e.codigo)
	}
	return nil
}

func (e *tec04) campoEs(campo string) func(string) error {
	return func(esperado string) error {
		if e.respuesta[campo] != esperado {
			return fmt.Errorf("se esperaba %s %q y se obtuvo %q", campo, esperado, e.respuesta[campo])
		}
		return nil
	}
}
