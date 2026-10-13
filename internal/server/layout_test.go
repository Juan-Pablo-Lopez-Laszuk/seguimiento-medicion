package server_test

import "testing"

// TEC-07 · RN1: en el celular el menú se junta detrás del botón ☰ y se abre con el JavaScript de Bootstrap.
func TestLayout_MenuSeColapsaEnElCelular(t *testing.T) {
	html := pedir(t, "/").Body.String()

	verificarContiene(t, html,
		`navbar-expand-md`,           // en pantallas medianas o más el menú se ve completo
		`class="navbar-toggler"`,     // en el celular aparece el botón ☰
		`data-bs-toggle="collapse"`,  // que abre y cierra el menú
		`data-bs-target="#menu"`,     // el menú que abre
		`id="menu"`,                  // el menú
		`aria-label="Abrir el menú"`, // el botón tiene un nombre para lectores de pantalla
		`bootstrap.bundle.min.js`,    // el JavaScript de Bootstrap que hace funcionar el botón
	)
}

// TEC-07 · CA-07.2: quien navega con teclado puede saltar el menú e ir directo al contenido; el pie es igual en todas.
func TestLayout_EnlaceParaIrAlContenidoYPie(t *testing.T) {
	html := pedir(t, "/proyectos").Body.String()

	verificarContiene(t, html,
		`href="#contenido"`, "Ir al contenido",
		`<main id="contenido"`,
		"<footer",
	)
}
