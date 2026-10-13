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
