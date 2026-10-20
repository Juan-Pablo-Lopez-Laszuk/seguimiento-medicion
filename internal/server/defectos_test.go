package server_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// conSprintEHistoria arma la aplicación con el proyecto de ejemplo, su Sprint 1 (ID 1) y su HU-1 (ID 1).
func conSprintEHistoria(t *testing.T) http.Handler {
	t.Helper()
	h := conHistoria(t)
	if rec := post(h, "/proyectos/1/sprints", sprintValido()); rec.Code != http.StatusSeeOther {
		t.Fatalf("no se pudo crear el sprint de ejemplo: código %d", rec.Code)
	}
	return h
}

func defectoValido() url.Values {
	return url.Values{"descripcion": {"El login falla con Google"}, "severidad": {"Alta"}, "sprint": {"1"}, "historia": {"1"}}
}

// A los defectos se llega desde la vista del proyecto y desde la lista de proyectos.
func TestProyecto_EnlazaASusDefectos(t *testing.T) {
	h := conProyecto(t)
	verificarContiene(t, get(h, "/proyectos").Body.String(), `href="/proyectos/1/defectos"`)
	verificarContiene(t, get(h, "/proyectos/1").Body.String(), `href="/proyectos/1/defectos"`)
}

// HU-27 · la página muestra que no hay defectos y el formulario con severidades, sprints e historias
func TestPaginaDeDefectos_SinDefectos_MuestraElFormulario(t *testing.T) {
	rec := get(conSprintEHistoria(t), "/proyectos/1/defectos")

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	verificarContiene(t, rec.Body.String(),
		"<title>Defectos · Software Metrics</title>", "Sin defectos registrados",
		`action="/proyectos/1/defectos"`, `method="post"`, `name="descripcion"`,
		`<option value="Crítica">Crítica</option>`, `<option value="Baja">Baja</option>`,
		`<option value="1">Sprint 1</option>`, `<option value="1">HU-1 · Iniciar sesión con Google</option>`, "Ninguna",
	)
}

// Sin sprints no se puede registrar un defecto: la página avisa y enlaza a crear uno.
func TestPaginaDeDefectos_SinSprints_Avisa(t *testing.T) {
	html := get(conProyecto(t), "/proyectos/1/defectos").Body.String()
	verificarContiene(t, html, "Creá un sprint", `href="/proyectos/1/sprints/nuevo"`)
	if strings.Contains(html, `name="descripcion"`) {
		t.Error("sin sprints no debería mostrarse el formulario")
	}
}

// CA-27.1 a CA-27.3 · después de registrar se vuelve a la lista con el defecto Abierto
func TestRegistrarDefecto_VuelveALaListaConElDefectoAbierto(t *testing.T) {
	h := conSprintEHistoria(t)

	rec := post(h, "/proyectos/1/defectos", defectoValido())

	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/proyectos/1/defectos?registrado=1" {
		t.Fatalf("código = %d y destino %q, se esperaba 303 a ?registrado=1", rec.Code, rec.Header().Get("Location"))
	}
	verificarContiene(t, get(h, "/proyectos/1/defectos?registrado=1").Body.String(),
		"Defecto registrado.", "El login falla con Google", "Alta", "Abierto", "Sprint 1", "HU-1 · Iniciar sesión con Google")
}

// CA-27.1 · los errores aparecen debajo de cada campo, se conserva lo escrito y no se guarda nada
func TestRegistrarDefecto_DatosInvalidos(t *testing.T) {
	h := conSprintEHistoria(t)

	rec := post(h, "/proyectos/1/defectos", url.Values{"descripcion": {"  "}, "severidad": {"Urgente"}, "sprint": {"1"}, "historia": {"1"}})

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusUnprocessableEntity)
	}
	verificarContiene(t, rec.Body.String(),
		"la descripción del defecto es obligatoria", "la severidad tiene que ser Crítica, Alta, Media o Baja",
		"is-invalid", `<option value="1" selected>Sprint 1</option>`)
	if !strings.Contains(get(h, "/proyectos/1/defectos").Body.String(), "Sin defectos registrados") {
		t.Error("no se debería haber guardado el defecto")
	}
}

func TestDefectos_ProyectoInexistente_Responde404(t *testing.T) {
	h := conProyecto(t)
	for nombre, codigo := range map[string]int{
		"ver":       get(h, "/proyectos/99/defectos").Code,
		"registrar": post(h, "/proyectos/99/defectos", defectoValido()).Code,
	} {
		if codigo != http.StatusNotFound {
			t.Errorf("%s: código = %d, se esperaba %d", nombre, codigo, http.StatusNotFound)
		}
	}
}

// conDefecto arma la aplicación con el proyecto de ejemplo, su Sprint 1 y un defecto Abierto (ID 1).
func conDefecto(t *testing.T) http.Handler {
	t.Helper()
	h := conSprintEHistoria(t)
	if rec := post(h, "/proyectos/1/defectos", defectoValido()); rec.Code != http.StatusSeeOther {
		t.Fatalf("no se pudo registrar el defecto de ejemplo: código %d", rec.Code)
	}
	return h
}

// HU-28 · cada defecto ofrece los estados a los que puede pasar
func TestPaginaDeDefectos_OfreceLosEstadosSiguientes(t *testing.T) {
	verificarContiene(t, get(conDefecto(t), "/proyectos/1/defectos").Body.String(),
		`action="/proyectos/1/defectos/1/estado"`, `value="En progreso"`, "Resuelto en")
}

// HU-28 · CA-28.1 y CA-28.2 · el defecto pasa a En progreso y se resuelve en el Sprint 1
func TestCambiarEstadoDeDefecto_VuelveALaListaConElEstadoNuevo(t *testing.T) {
	h := conDefecto(t)
	rec := post(h, "/proyectos/1/defectos/1/estado", url.Values{"estado": {"En progreso"}})
	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/proyectos/1/defectos?estado=en-progreso" {
		t.Fatalf("código = %d y destino %q, se esperaba 303 a ?estado=en-progreso", rec.Code, rec.Header().Get("Location"))
	}
	verificarContiene(t, get(h, "/proyectos/1/defectos?estado=en-progreso").Body.String(), "Defecto en estado En progreso.", `value="Resuelto"`, `name="sprint"`)

	if rec := post(h, "/proyectos/1/defectos/1/estado", url.Values{"estado": {"Resuelto"}, "sprint": {"1"}}); rec.Code != http.StatusSeeOther {
		t.Fatalf("resolver: código = %d, se esperaba %d", rec.Code, http.StatusSeeOther)
	}
	html := get(h, "/proyectos/1/defectos?estado=resuelto").Body.String()
	verificarContiene(t, html, "Defecto en estado Resuelto.", `value="Cerrado"`, `value="Reabierto"`)
	if strings.Count(html, "Sprint 1</td>") != 2 {
		t.Errorf("la tabla debería mostrar el Sprint 1 como sprint de detección y de resolución")
	}
}

// HU-28 · un valor de ?estado= que no es conocido no muestra ningún mensaje
func TestPaginaDeDefectos_EstadoDesconocidoEnLaURL_NoMuestraMensaje(t *testing.T) {
	if strings.Contains(get(conDefecto(t), "/proyectos/1/defectos?estado=hackeado").Body.String(), "Defecto en estado") {
		t.Error("no debería mostrar un mensaje para un estado desconocido")
	}
}

// HU-28 · CA-28.1 y CA-28.2 · las transiciones y los sprints que no valen se rechazan con el motivo (409)
func TestCambiarEstadoDeDefecto_CambioNoPermitido_Responde409(t *testing.T) {
	h := conDefecto(t)
	rec := post(h, "/proyectos/1/defectos/1/estado", url.Values{"estado": {"Cerrado"}})
	if rec.Code != http.StatusConflict {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusConflict)
	}
	verificarContiene(t, rec.Body.String(), "un defecto en estado Abierto no puede pasar a Cerrado", "El login falla con Google")

	post(h, "/proyectos/1/defectos/1/estado", url.Values{"estado": {"En progreso"}})
	rec = post(h, "/proyectos/1/defectos/1/estado", url.Values{"estado": {"Resuelto"}, "sprint": {"x"}})
	if rec.Code != http.StatusConflict {
		t.Fatalf("sin sprint: código = %d, se esperaba %d", rec.Code, http.StatusConflict)
	}
	verificarContiene(t, rec.Body.String(), "el sprint de resolución tiene que ser un sprint del proyecto")
}

// HU-28 · un estado que no existe es un pedido inválido; un defecto o proyecto que no existe, 404
func TestCambiarEstadoDeDefecto_PedidosInvalidos(t *testing.T) {
	h := conDefecto(t)
	for ruta, esperado := range map[string]int{
		"/proyectos/1/defectos/1/estado":  http.StatusBadRequest,
		"/proyectos/1/defectos/99/estado": http.StatusNotFound,
		"/proyectos/1/defectos/x/estado":  http.StatusNotFound,
		"/proyectos/99/defectos/1/estado": http.StatusNotFound,
	} {
		estado := "En progreso"
		if esperado == http.StatusBadRequest {
			estado = "Hecho"
		}
		if rec := post(h, ruta, url.Values{"estado": {estado}}); rec.Code != esperado {
			t.Errorf("%s (%s): código = %d, se esperaba %d", ruta, estado, rec.Code, esperado)
		}
	}
}

// HU-34 · CA-34.1 y CA-34.2 · la página de defectos muestra el resumen por sprint y el total, con filtro por severidad
func TestPaginaDeDefectos_MuestraElResumen(t *testing.T) {
	h := conDefecto(t) // un defecto Alta detectado en el Sprint 1
	verificarContiene(t, get(h, "/proyectos/1/defectos").Body.String(),
		"Resumen", `id="filtro-severidad"`, `<option value="" selected>Todas</option>`,
		`<td class="text-end" data-detectados="Sprint 1">1</td>`, `<td class="text-end" data-resueltos="Sprint 1">0</td>`,
		`<strong id="total-detectados">1</strong>`, `<strong id="total-resueltos">0</strong>`, `<strong id="total-abiertos">1</strong>`,
	)
	verificarContiene(t, get(h, "/proyectos/1/defectos?severidad=Baja").Body.String(),
		`<option value="Baja" selected>Baja</option>`, `<strong id="total-detectados">0</strong>`,
		"El login falla con Google", // la lista de defectos no se filtra
	)
}
