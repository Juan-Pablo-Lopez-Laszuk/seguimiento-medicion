package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// conHistoriasEn arma la aplicación con el proyecto de ejemplo (ID 1) y una historia por cada estado de la lista (la
// primera es la HU-1, y así), guardadas directo en los repositorios: los cambios de estado de HU-12 y HU-13 todavía no
// están hechos. La HU-1 tiene total criterios, los primeros cumplidos.
func conHistoriasEn(t *testing.T, estados []backlog.Estado, total, cumplidos int) http.Handler {
	t.Helper()
	ctx := context.Background()
	proyectos, historias, criterios := memory.NuevoProyectos(), memory.NuevoHistorias(), memory.NuevoCriterios()
	servicioProyectos := service.NuevoProyectos(proyectos, time.Now)
	p, err := servicioProyectos.Crear(ctx, project.Datos{
		Nombre:      "Software Metrics",
		FechaInicio: time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC),
		FechaFin:    time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("no se pudo crear el proyecto: %v", err)
	}
	for i, estado := range estados {
		h, err := historias.Guardar(ctx, backlog.Historia{
			ProyectoID: p.ID, Numero: i + 1, Titulo: "Historia " + string(estado),
			Prioridad: backlog.PrioridadAlta, Estado: estado,
		})
		if err != nil {
			t.Fatalf("no se pudo guardar la historia %d: %v", i+1, err)
		}
		if i != 0 {
			continue
		}
		for c := 1; c <= total; c++ {
			if _, err := criterios.Guardar(ctx, backlog.Criterio{HistoriaID: h.ID, Descripcion: "Un criterio", Cumplido: c <= cumplidos}); err != nil {
				t.Fatalf("no se pudo guardar el criterio %d: %v", c, err)
			}
		}
	}
	return server.NewRouter(server.Dependencias{
		Proyectos:      servicioProyectos,
		Historias:      service.NuevoHistorias(proyectos, historias),
		Criterios:      service.NuevoCriterios(proyectos, historias, criterios),
		CambioDeEstado: service.NuevoEstados(proyectos, historias, criterios),
		Estimaciones:   service.NuevoEstimaciones(proyectos, historias),
	})
}

func conLaHistoriaEn(t *testing.T, estado backlog.Estado, total, cumplidos int) http.Handler {
	t.Helper()
	return conHistoriasEn(t, []backlog.Estado{estado}, total, cumplidos)
}

func estado(valor string) url.Values {
	return url.Values{"estado": {valor}}
}

const avisoDeLaPantalla = "Desde esta pantalla solo se puede pasar a En progreso."

// CA-08.1 · la página de una historia En sprint ofrece empezar a trabajarla
func TestPaginaDeLaHistoria_EnSprint_OfreceEmpezarATrabajarla(t *testing.T) {
	html := get(conLaHistoriaEn(t, backlog.EstadoEnSprint, 0, 0), "/proyectos/1/backlog/1").Body.String()

	verificarContiene(t, html,
		`action="/proyectos/1/backlog/1/estado"`, `name="estado" value="En progreso"`, "Empezar a trabajarla", "En sprint",
	)
}

// CA-08.1 · después de empezar se vuelve a la página con el estado nuevo, y ya no se ofrece empezar
func TestCambiarEstado_EnSprintAEnProgreso_VuelveALaPaginaConElEstadoNuevo(t *testing.T) {
	h := conLaHistoriaEn(t, backlog.EstadoEnSprint, 0, 0)

	rec := post(h, "/proyectos/1/backlog/1/estado", estado("En progreso"))

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("código = %d, se esperaba %d (redirección)", rec.Code, http.StatusSeeOther)
	}
	destino := rec.Header().Get("Location")
	if destino != "/proyectos/1/backlog/1?estado=en-progreso" {
		t.Fatalf("redirige a %q, se esperaba /proyectos/1/backlog/1?estado=en-progreso", destino)
	}
	html := get(h, destino).Body.String()
	verificarContiene(t, html, "La historia pasó a En progreso.", "En progreso")
	if strings.Contains(html, "Empezar a trabajarla") {
		t.Error("una historia que ya está En progreso no tiene que ofrecer empezarla")
	}
}

// La página dice qué sigue en cada estado, y solo En sprint tiene el botón
func TestPaginaDeLaHistoria_DiceQueSigueEnCadaEstado(t *testing.T) {
	casos := []struct {
		estado   backlog.Estado
		esperado string
	}{
		{backlog.EstadoPendiente, "cuando se asigna a un sprint"},
		{backlog.EstadoEnProgreso, "desde el sprint activo"},
		{backlog.EstadoHecho, "Historia terminada"},
	}
	for _, c := range casos {
		html := get(conLaHistoriaEn(t, c.estado, 1, 1), "/proyectos/1/backlog/1").Body.String()

		verificarContiene(t, html, c.esperado)
		if strings.Contains(html, `action="/proyectos/1/backlog/1/estado"`) {
			t.Errorf("%s: la página no tiene que ofrecer cambiar el estado desde ahí", c.estado)
		}
	}
}

// CA-08.3 y RN7 · un cambio que no vale se informa en la página, con su motivo, y la historia queda como estaba
func TestCambiarEstado_Invalido_MuestraElAvisoYNoCambiaNada(t *testing.T) {
	casos := map[string]struct {
		estado  backlog.Estado
		mensaje string
	}{
		"desde Pendiente": {backlog.EstadoPendiente, "no se puede pasar de Pendiente a En progreso: desde Pendiente solo se puede pasar a En sprint"},
		"desde Hecho":     {backlog.EstadoHecho, "no se puede pasar de Hecho a En progreso: una historia Hecha no cambia de estado"},
		"el mismo estado": {backlog.EstadoEnProgreso, "la historia ya está En progreso"},
	}
	for nombre, c := range casos {
		h := conLaHistoriaEn(t, c.estado, 1, 1)

		rec := post(h, "/proyectos/1/backlog/1/estado", estado("En progreso"))

		if rec.Code != http.StatusConflict {
			t.Errorf("%s: código = %d, se esperaba %d", nombre, rec.Code, http.StatusConflict)
		}
		verificarContiene(t, rec.Body.String(), c.mensaje)
		verificarContiene(t, get(h, "/proyectos/1/backlog/1").Body.String(), string(c.estado))
	}
}

// RN8 · desde esta pantalla solo se pasa a En progreso: lo demás lo hacen HU-12 (En sprint) y HU-13 (Hecho), así que
// el servidor lo rechaza aunque la regla de la historia lo permitiría
func TestCambiarEstado_AUnEstadoQueLaPantallaNoOfrece_Responde400YNoCambiaNada(t *testing.T) {
	h := conLaHistoriaEn(t, backlog.EstadoEnProgreso, 2, 2) // con todo cumplido: la regla dejaría pasar a Hecho

	for _, valor := range []string{"Hecho", "En sprint", "Pendiente", "Terminada", ""} {
		rec := post(h, "/proyectos/1/backlog/1/estado", estado(valor))

		if rec.Code != http.StatusBadRequest {
			t.Errorf("estado %q: código = %d, se esperaba %d", valor, rec.Code, http.StatusBadRequest)
		}
		verificarContiene(t, rec.Body.String(), avisoDeLaPantalla)
	}
	verificarContiene(t, get(h, "/proyectos/1/backlog/1").Body.String(), `text-bg-primary">En progreso`) // sigue En progreso
}

// Los parámetros del mensaje de la página solo muestran el texto conocido, nunca lo que venga en la URL
func TestPaginaDeLaHistoria_ElMensajeDelEstadoSoloMuestraElTextoConocido(t *testing.T) {
	h := conLaHistoriaEn(t, backlog.EstadoEnSprint, 0, 0)

	html := get(h, "/proyectos/1/backlog/1?estado=%3Cscript%3E").Body.String()

	if strings.Contains(html, "<script>") || strings.Contains(html, "La historia pasó a") {
		t.Error("la página mostró un mensaje que no corresponde a un valor conocido")
	}
}

// La lista del backlog muestra el estado de cada historia con un color distinto
func TestListaDelBacklog_MuestraElEstadoConUnColorPorEstado(t *testing.T) {
	h := conHistoriasEn(t, backlog.Estados, 0, 0)

	html := get(h, "/proyectos/1/backlog").Body.String()

	verificarContiene(t, html,
		`text-bg-secondary">Pendiente`, `text-bg-info">En sprint`, `text-bg-primary">En progreso`, `text-bg-success">Hecho`,
	)
}

// El proyecto o la historia que no existen (o un ID que no es un número) dan 404
func TestCambiarEstado_ProyectoOHistoriaInexistente_Responde404(t *testing.T) {
	h := conLaHistoriaEn(t, backlog.EstadoEnSprint, 0, 0)
	pedidos := map[string]*httptest.ResponseRecorder{
		"proyecto que no existe":                 post(h, "/proyectos/99/backlog/1/estado", estado("En progreso")),
		"historia que no existe":                 post(h, "/proyectos/1/backlog/99/estado", estado("En progreso")),
		"número de historia que no es número":    post(h, "/proyectos/1/backlog/abc/estado", estado("En progreso")),
		"número de proyecto que no es un número": post(h, "/proyectos/abc/backlog/1/estado", estado("En progreso")),
	}
	for nombre, rec := range pedidos {
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s: código = %d, se esperaba %d", nombre, rec.Code, http.StatusNotFound)
		}
	}
}

// historiasQueFallanAlCambiarElEstado simula la base de datos caída: la historia existe y está En sprint, pero no se
// puede guardar su estado.
type historiasQueFallanAlCambiarElEstado struct{}

func (historiasQueFallanAlCambiarElEstado) BuscarPorNumero(_ context.Context, proyectoID int64, numero int) (backlog.Historia, error) {
	return backlog.Historia{ID: 4, ProyectoID: proyectoID, Numero: numero, Titulo: "Iniciar sesión", Estado: backlog.EstadoEnSprint}, nil
}
func (historiasQueFallanAlCambiarElEstado) ActualizarEstado(context.Context, int64, backlog.Estado) error {
	return errBaseCaida
}

// Si falla la base se muestra un mensaje general, sin detalles internos.
func TestCambiarEstado_ErrorDeLaBase_MuestraUnMensajeGeneral(t *testing.T) {
	historias := historiasQueFallanAlCambiarElEstado{}
	h := server.NewRouter(server.Dependencias{
		Proyectos:      service.NuevoProyectos(repoCaido{}, nil),
		Criterios:      service.NuevoCriterios(sprintsCaidos{}, historias, criteriosVacios{}),
		CambioDeEstado: service.NuevoEstados(sprintsCaidos{}, historias, criteriosVacios{}),
	})

	rec := post(h, "/proyectos/1/backlog/1/estado", estado("En progreso"))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("código = %d, se esperaba %d", rec.Code, http.StatusInternalServerError)
	}
	if strings.Contains(rec.Body.String(), "detalle interno") {
		t.Error("la página muestra el error interno")
	}
}

// criteriosVacios es un repositorio de criterios que funciona pero no tiene ninguno.
type criteriosVacios struct{ criteriosCaidos }

func (criteriosVacios) ListarPorHistoria(context.Context, int64) ([]backlog.Criterio, error) {
	return nil, nil
}
func (criteriosVacios) ListarPorHistorias(context.Context, []int64) ([]backlog.Criterio, error) {
	return nil, nil
}
