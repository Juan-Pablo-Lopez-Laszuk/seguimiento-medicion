package server_test

import (
	"strings"
	"testing"
)

// TEC-07 · RN2 y RN3: el mensaje de éxito es verde, lo anuncia un lector de pantalla y se puede cerrar.
func TestMensajeDeExito_SeAnunciaYSePuedeCerrar(t *testing.T) {
	html := get(conProyecto(t), "/proyectos?creado=1").Body.String()

	verificarContiene(t, html,
		`alert alert-success alert-dismissible`,
		`role="status"`,
		"Proyecto creado.",
		`class="btn-close"`, `data-bs-dismiss="alert"`, `aria-label="Cerrar"`,
	)
}

// TEC-07 · RN2: un aviso no se puede cerrar, para que no se pierda.
func TestAviso_NoSePuedeCerrar(t *testing.T) {
	h := conProyecto(t)
	post(h, "/proyectos/1/integrantes", integranteValido())
	post(h, "/proyectos/1/integrantes/1/baja", nil)
	html := post(h, "/proyectos/1/integrantes/1/baja", nil).Body.String()

	verificarContiene(t, html, `alert alert-warning`, `role="alert"`, "el integrante ya está dado de baja")
	if strings.Contains(html, `data-bs-dismiss="alert"`) {
		t.Error("el aviso no debería poder cerrarse")
	}
}

// TEC-07 · RN4: dar de baja pide confirmación en una ventana de la página, no con el cuadro del navegador.
func TestDarDeBaja_PideConfirmacionEnUnaVentana(t *testing.T) {
	h := conProyecto(t)
	post(h, "/proyectos/1/integrantes", integranteValido())
	html := get(h, "/proyectos/1/integrantes").Body.String()

	verificarContiene(t, html,
		`data-bs-toggle="modal"`, `data-bs-target="#confirmar-baja-1"`, // el botón abre la ventana
		`class="modal fade" id="confirmar-baja-1"`,                  // la ventana
		"¿Dar de baja a Juan Pablo?",                                // qué se confirma
		`action="/proyectos/1/integrantes/1/baja"`, `method="post"`, // el formulario se envía recién al confirmar
		"Cancelar",
	)
	if strings.Contains(html, "confirm(") {
		t.Error("la baja no debería usar el cuadro de confirmación del navegador")
	}
}

// TEC-07b · CA-07b.1 y RN3: el formulario de nuevo integrante usa el componente campo: etiqueta unida al campo, ayuda
// unida con aria-describedby y, sin datos, ningún "<no value>" (RN1).
func TestCampo_EtiquetaYAyudaUnidasAlCampo(t *testing.T) {
	html := get(conProyecto(t), "/proyectos/1/integrantes/nuevo").Body.String()

	verificarContiene(t, html,
		`<label class="form-label" for="nombre">Nombre</label>`,
		`<label class="form-label" for="email">Email</label>`,
		`aria-describedby="email-ayuda"`, `id="email-ayuda"`, "No puede repetirse dentro del proyecto.",
		`value=""`,
	)
	if strings.Contains(html, "no value") {
		t.Error("un dato que no se pasa al campo no tiene que mostrar <no value>")
	}
	if strings.Contains(html, `aria-describedby="nombre-ayuda"`) {
		t.Error("el campo Nombre no tiene ayuda: no tiene que llevar aria-describedby")
	}
}

// TEC-07b · RN2: con error, el campo queda en rojo con el mensaje debajo y conserva lo escrito.
func TestCampo_ConErrorMarcaElCampoYMuestraElMensaje(t *testing.T) {
	campos := integranteValido()
	campos.Set("email", "sin-arroba")

	html := post(conProyecto(t), "/proyectos/1/integrantes", campos).Body.String()

	verificarContiene(t, html, `class="form-control is-invalid" type="email" id="email" name="email" value="sin-arroba"`, `<div class="invalid-feedback">`)
}
