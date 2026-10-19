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
