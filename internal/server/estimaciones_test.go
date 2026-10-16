package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

func puntos(valor string) url.Values {
	return url.Values{"story_points": {valor}}
}

// sinEstimar arma la aplicación con el proyecto de ejemplo y su HU-1, Pendiente y sin estimar (conHistoria crea la
// historia ya estimada en 5).
func sinEstimar(t *testing.T) http.Handler {
	t.Helper()
	return conLaHistoriaEn(t, backlog.EstadoPendiente, 0, 0)
}

const avisoDeHistoriaHechaAlEstimar = "la historia está Hecha: no se puede cambiar su estimación"

// CA-16.1 · la página de la historia ofrece la escala para estimar, y la historia sin estimar no tiene ningún valor elegido
func TestPaginaDeLaHistoria_OfreceEstimarConLaEscala(t *testing.T) {
	html := get(sinEstimar(t), "/proyectos/1/backlog/1").Body.String()

	verificarContiene(t, html,
		`action="/proyectos/1/backlog/1/estimacion"`, `name="story_points"`, "Guardar estimación",
		`value="0"`, `value="1"`, `value="2"`, `value="3"`, `value="5"`, `value="8"`, `value="13"`, `value="21"`, // la escala acordada
	)
	if strings.Contains(html, `selected>`) {
		t.Error("una historia sin estimar no tiene que tener ningún valor elegido")
	}
}

// CA-16.1 · estimar guarda el valor, vuelve a la página con el mensaje y deja elegido el valor nuevo
func TestEstimarHistoria_Valido_VuelveALaPaginaConLaEstimacion(t *testing.T) {
	h := sinEstimar(t)

	rec := post(h, "/proyectos/1/backlog/1/estimacion", puntos("5"))

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("código = %d, se esperaba %d (redirección)", rec.Code, http.StatusSeeOther)
	}
	destino := rec.Header().Get("Location")
	if destino != "/proyectos/1/backlog/1?estimada=1" {
		t.Fatalf("redirige a %q, se esperaba /proyectos/1/backlog/1?estimada=1", destino)
	}
	verificarContiene(t, get(h, destino).Body.String(), "Estimación guardada.", "Story points: 5", `<option value="5" selected>`)
}

// RN5 · re-estimar reemplaza el valor, y 0 cuenta como estimada
func TestEstimarHistoria_Reestimar_ReemplazaElValorAnterior(t *testing.T) {
	h := sinEstimar(t)
	post(h, "/proyectos/1/backlog/1/estimacion", puntos("5"))

	post(h, "/proyectos/1/backlog/1/estimacion", puntos("8"))
	verificarContiene(t, get(h, "/proyectos/1/backlog/1").Body.String(), "Story points: 8", `<option value="8" selected>`)

	post(h, "/proyectos/1/backlog/1/estimacion", puntos("0"))
	html := get(h, "/proyectos/1/backlog/1").Body.String()
	verificarContiene(t, html, "Story points: 0", `<option value="0" selected>`)
	if strings.Contains(html, "Sin estimar") {
		t.Error("una historia estimada en 0 no está sin estimar")
	}
}

// CA-16.2 y RN7 · un valor fuera de la escala o vacío se informa debajo del campo y se conserva la estimación anterior
func TestEstimarHistoria_ValorInvalido_MuestraElErrorYConservaLaEstimacion(t *testing.T) {
	h := sinEstimar(t)
	post(h, "/proyectos/1/backlog/1/estimacion", puntos("5"))

	for _, valor := range []string{"4", "22", "-1", "abc", ""} {
		rec := post(h, "/proyectos/1/backlog/1/estimacion", puntos(valor))

		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("valor %q: código = %d, se esperaba %d", valor, rec.Code, http.StatusUnprocessableEntity)
		}
		verificarContiene(t, rec.Body.String(), "los story points tienen que ser 0, 1, 2, 3, 5, 8, 13 o 21", "is-invalid", "Story points: 5")
	}
	verificarContiene(t, get(h, "/proyectos/1/backlog/1").Body.String(), "Story points: 5", `<option value="5" selected>`)
}

// RN6 · una historia Hecha no se re-estima: la página no ofrece estimar y el servidor tampoco
func TestEstimarHistoria_HistoriaHecha_NoSeReestima(t *testing.T) {
	h := conLaHistoriaEn(t, backlog.EstadoHecho, 1, 1)

	pagina := get(h, "/proyectos/1/backlog/1").Body.String()
	verificarContiene(t, pagina, "La estimación no se puede cambiar")
	if strings.Contains(pagina, `action="/proyectos/1/backlog/1/estimacion"`) {
		t.Error("la página de una historia Hecha no tiene que ofrecer estimar")
	}

	rec := post(h, "/proyectos/1/backlog/1/estimacion", puntos("8"))

	if rec.Code != http.StatusConflict {
		t.Errorf("código = %d, se esperaba %d", rec.Code, http.StatusConflict)
	}
	verificarContiene(t, rec.Body.String(), avisoDeHistoriaHechaAlEstimar)
	verificarContiene(t, get(h, "/proyectos/1/backlog/1").Body.String(), "Story points: Sin estimar")
}

// RN6 · en Pendiente, En sprint y En progreso se puede estimar
func TestEstimarHistoria_PendienteEnSprintOEnProgreso_Admiten(t *testing.T) {
	for _, estado := range []backlog.Estado{backlog.EstadoPendiente, backlog.EstadoEnSprint, backlog.EstadoEnProgreso} {
		h := conLaHistoriaEn(t, estado, 0, 0)

		if rec := post(h, "/proyectos/1/backlog/1/estimacion", puntos("13")); rec.Code != http.StatusSeeOther {
			t.Errorf("%s: código = %d, se esperaba %d", estado, rec.Code, http.StatusSeeOther)
		}
		verificarContiene(t, get(h, "/proyectos/1/backlog/1").Body.String(), "Story points: 13")
	}
}

// Los parámetros del mensaje de la página solo muestran el texto conocido, nunca lo que venga en la URL
func TestPaginaDeLaHistoria_ElMensajeDeLaEstimacionSoloMuestraElTextoConocido(t *testing.T) {
	html := get(conHistoria(t), "/proyectos/1/backlog/1?estimada=%3Cscript%3E").Body.String()

	if strings.Contains(html, "<script>") || strings.Contains(html, "Estimación guardada") {
		t.Error("la página mostró un mensaje que no corresponde a un valor conocido")
	}
}

// El proyecto o la historia que no existen (o un ID que no es un número) dan 404
func TestEstimarHistoria_ProyectoOHistoriaInexistente_Responde404(t *testing.T) {
	h := conHistoria(t)
	pedidos := map[string]*httptest.ResponseRecorder{
		"proyecto que no existe":                 post(h, "/proyectos/99/backlog/1/estimacion", puntos("5")),
		"historia que no existe":                 post(h, "/proyectos/1/backlog/99/estimacion", puntos("5")),
		"número de historia que no es un número": post(h, "/proyectos/1/backlog/abc/estimacion", puntos("5")),
		"número de proyecto que no es un número": post(h, "/proyectos/abc/backlog/1/estimacion", puntos("5")),
	}
	for nombre, rec := range pedidos {
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s: código = %d, se esperaba %d", nombre, rec.Code, http.StatusNotFound)
		}
	}
}

// historiasQueFallanAlGuardarLaEstimacion simula la base de datos caída: la historia existe pero no se pueden guardar
// sus story points.
type historiasQueFallanAlGuardarLaEstimacion struct{}

func (historiasQueFallanAlGuardarLaEstimacion) BuscarPorNumero(_ context.Context, proyectoID int64, numero int) (backlog.Historia, error) {
	return backlog.Historia{ID: 4, ProyectoID: proyectoID, Numero: numero, Titulo: "Iniciar sesión", Estado: backlog.EstadoPendiente}, nil
}
func (historiasQueFallanAlGuardarLaEstimacion) ActualizarStoryPoints(context.Context, int64, int) error {
	return errBaseCaida
}

// Si falla la base se muestra un mensaje general, sin detalles internos.
func TestEstimarHistoria_ErrorDeLaBase_MuestraUnMensajeGeneral(t *testing.T) {
	historias := historiasQueFallanAlGuardarLaEstimacion{}
	h := server.NewRouter(server.Dependencias{
		Proyectos:    service.NuevoProyectos(repoCaido{}, nil),
		Criterios:    service.NuevoCriterios(sprintsCaidos{}, historias, criteriosVacios{}),
		Estimaciones: service.NuevoEstimaciones(sprintsCaidos{}, historias),
	})

	rec := post(h, "/proyectos/1/backlog/1/estimacion", puntos("5"))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("código = %d, se esperaba %d", rec.Code, http.StatusInternalServerError)
	}
	if strings.Contains(rec.Body.String(), "detalle interno") {
		t.Error("la página muestra el error interno")
	}
}
