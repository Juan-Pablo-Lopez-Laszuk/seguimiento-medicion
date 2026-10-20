package server_test

import (
	"net/http"
	"strings"
	"testing"
)

const rutaDelReporte = "/proyectos/1/sprints/1/reporte"

// CA-37.1 · el reporte del Sprint 1 cerrado (HU-1 completada con 5 SP y HU-3 sin completar con 3 SP) muestra las
// métricas y las historias de la foto, con enlace a cada una
func TestReporteDeSprint_MuestraElReporte(t *testing.T) {
	rec := get(conSprintCerrado(t), rutaDelReporte)

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	verificarContiene(t, rec.Body.String(),
		"<title>Reporte del Sprint 1 · Software Metrics</title>",
		"Reporte del Sprint 1", "Cerrado",
		"SP planificados y completados", "5 de 8 SP",
		"Cumplimiento", "62,5 %",
		"Historias completadas", "1 de 2 (50,0 %)",
		"Velocidad del equipo", "5,0 SP por sprint (1 sprint cerrado)",
		"Horas estimadas y reales", "0 h estimadas · 0 h reales",
		"Desviación", "0 h (no aplicable)",
		`href="/proyectos/1/backlog/1"`, "Iniciar sesión con Google", "Completada",
		`href="/proyectos/1/backlog/3"`, "Editar el perfil", "No completada")
}

// Sección 3 · el reporte se imprime sin la navegación ni los botones
func TestReporteDeSprint_SeImprime(t *testing.T) {
	pagina := get(conSprintCerrado(t), rutaDelReporte).Body.String()

	verificarContiene(t, pagina, `onclick="window.print()"`, "Imprimir")
	if strings.Count(pagina, "d-print-none") < 3 {
		t.Error("la barra de navegación, las migas y el botón Imprimir tienen que ocultarse al imprimir (d-print-none)")
	}
	if strings.Contains(pagina, "El sprint no está cerrado") {
		t.Error("el Sprint 1 está cerrado: el reporte no es parcial")
	}
}

// Sección 2 · la página del sprint tiene el botón para ver su reporte
func TestPaginaDelSprint_EnlazaAlReporte(t *testing.T) {
	verificarContiene(t, get(conSprintCerrado(t), "/proyectos/1/sprints/2").Body.String(),
		`href="/proyectos/1/sprints/2/reporte"`, "Reporte del sprint")
}

// RN2 · el reporte de un sprint abierto avisa que muestra cómo va hasta hoy
func TestReporteDeSprintAbierto_AvisaQueEsParcial(t *testing.T) {
	verificarContiene(t, get(conSprintCerrado(t), "/proyectos/1/sprints/2/reporte").Body.String(),
		"El sprint no está cerrado: el reporte muestra cómo va hasta hoy.", "0 de 0 SP")
}

// CA-37.3 y sección 7 · un sprint o un proyecto que no existen dan 404
func TestReporteDeSprint_NoExiste(t *testing.T) {
	h := conSprintCerrado(t)
	casos := map[string]string{
		"/proyectos/1/sprints/9/reporte":   "No encontramos ese sprint.",
		"/proyectos/1/sprints/abc/reporte": "No encontramos ese sprint.",
		"/proyectos/9/sprints/1/reporte":   "No encontramos ese proyecto.",
	}
	for ruta, mensaje := range casos {
		rec := get(h, ruta)
		if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), mensaje) {
			t.Errorf("%s: se esperaba 404 con %q y se obtuvo %d con %q", ruta, mensaje, rec.Code, rec.Body.String())
		}
	}
}

// CA-37.1 · el reporte muestra los defectos del sprint con su conteo, y un sprint sin defectos lo dice
func TestReporteDeSprint_MuestraLosDefectos(t *testing.T) {
	h := conSprintCerrado(t)
	if rec := post(h, "/proyectos/1/defectos", defectoValido()); rec.Code != http.StatusSeeOther {
		t.Fatalf("no se pudo registrar el defecto: código %d", rec.Code)
	}

	verificarContiene(t, get(h, rutaDelReporte).Body.String(),
		"Defectos", "1 detectado · 0 resueltos", "Defectos del sprint",
		"El login falla con Google", "Alta", "Abierto", "HU-1 · Iniciar sesión con Google", "Sprint 1", "—")
	verificarContiene(t, get(h, "/proyectos/1/sprints/2/reporte").Body.String(),
		"0 detectados · 0 resueltos", "No hubo defectos en este sprint.")
}
