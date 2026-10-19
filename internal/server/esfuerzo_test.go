package server_test

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// registroValido carga 3 h de Carolina (integrante 2) el 13/10/2026, el "hoy" de los tests de pantalla.
func registroValido() url.Values {
	return url.Values{"integrante": {"2"}, "tarea": {""}, "fecha": {"2026-10-13"}, "actividad": {"Pantalla de login"}, "horas": {"3"}}
}

// A la página de esfuerzo se llega desde la historia y desde sus tareas.
func TestHistoriaYTareas_EnlazanAlEsfuerzo(t *testing.T) {
	h := conHistoriaEIntegrantes(t)
	verificarContiene(t, get(h, "/proyectos/1/backlog/1").Body.String(), `href="/proyectos/1/backlog/1/esfuerzo"`)
	verificarContiene(t, get(h, "/proyectos/1/backlog/1/tareas").Body.String(), `href="/proyectos/1/backlog/1/esfuerzo"`)
}

// HU-24 · la página muestra la historia, que no tiene registros, y el formulario con integrantes, tareas y fecha máxima
func TestPaginaDeEsfuerzo_SinRegistros_MuestraElFormulario(t *testing.T) {
	h := conHistoriaEIntegrantes(t)
	agregarTarea(t, h)

	rec := get(h, "/proyectos/1/backlog/1/esfuerzo")

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	verificarContiene(t, rec.Body.String(),
		"<title>Esfuerzo de HU-1 · Software Metrics</title>", "Sin horas registradas todavía", "0 h",
		`action="/proyectos/1/backlog/1/esfuerzo"`, `method="post"`,
		`<option value="1">Juan Pablo</option>`, `<option value="2">Carolina</option>`,
		`<option value="1">Pantalla de login</option>`, "Toda la historia",
		`name="fecha"`, `max="2026-10-13"`, `min="2026-10-05"`, `name="actividad"`, `name="horas"`,
	)
}

// CA-24.1 y M5 · después de registrar se vuelve a la página con el registro y las horas reales
func TestRegistrarEsfuerzo_VuelveALaPaginaConElRegistro(t *testing.T) {
	h := conHistoriaEIntegrantes(t)
	agregarTarea(t, h)
	conTarea := registroValido()
	conTarea.Set("tarea", "1")
	conTarea.Set("horas", "1,5")

	rec := post(h, "/proyectos/1/backlog/1/esfuerzo", registroValido())
	post(h, "/proyectos/1/backlog/1/esfuerzo", conTarea)

	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/proyectos/1/backlog/1/esfuerzo?registrado=1" {
		t.Fatalf("código = %d y destino %q, se esperaba 303 a ?registrado=1", rec.Code, rec.Header().Get("Location"))
	}
	verificarContiene(t, get(h, "/proyectos/1/backlog/1/esfuerzo?registrado=1").Body.String(),
		"Horas registradas.", "13/10/2026", "Carolina", "Pantalla de login", "3 h", "1,5 h",
		"Horas reales de la historia", "4,5 h",
	)
}

// CA-24.1 a CA-24.3 · los errores aparecen debajo de cada campo, se conserva lo escrito y no se guarda nada
func TestRegistrarEsfuerzo_DatosInvalidos(t *testing.T) {
	h := conHistoriaEIntegrantes(t)
	malo := url.Values{"integrante": {"2"}, "fecha": {"2026-10-14"}, "actividad": {"  "}, "horas": {"25"}}

	rec := post(h, "/proyectos/1/backlog/1/esfuerzo", malo)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusUnprocessableEntity)
	}
	verificarContiene(t, rec.Body.String(),
		"la fecha no puede ser futura", "la actividad es obligatoria", "las horas tienen que ser mayores a 0 y hasta 24",
		"is-invalid", `value="25"`, `value="2026-10-14"`, `<option value="2" selected>Carolina</option>`,
	)
	if !strings.Contains(get(h, "/proyectos/1/backlog/1/esfuerzo").Body.String(), "Sin horas registradas todavía") {
		t.Error("no se debería haber guardado ningún registro")
	}
}

// Sección 7 · proyecto o historia que no existen son un 404
func TestEsfuerzo_ProyectoOHistoriaInexistente_Responde404(t *testing.T) {
	h := conHistoriaEIntegrantes(t)
	for nombre, codigo := range map[string]int{
		"proyecto":  get(h, "/proyectos/99/backlog/1/esfuerzo").Code,
		"historia":  get(h, "/proyectos/1/backlog/99/esfuerzo").Code,
		"registrar": post(h, "/proyectos/1/backlog/99/esfuerzo", registroValido()).Code,
	} {
		if codigo != http.StatusNotFound {
			t.Errorf("%s: código = %d, se esperaba %d", nombre, codigo, http.StatusNotFound)
		}
	}
}
