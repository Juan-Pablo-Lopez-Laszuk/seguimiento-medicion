package server_test

import (
	"net/http"
	"strings"
	"testing"
)

const rutaDelReporteDelProyecto = "/proyectos/1/reporte"

// CA-38.1 y CA-38.2 · con el Sprint 1 cerrado (5 de 8 SP) y el Sprint 2 abierto, el reporte del proyecto muestra el
// resumen, la evolución de la velocidad con el Sprint 1 y un bloque por sprint con su enlace al reporte de ese sprint
func TestReporteDeProyecto_MuestraElReporte(t *testing.T) {
	rec := get(conSprintCerrado(t), rutaDelReporteDelProyecto)

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	pagina := rec.Body.String()
	verificarContiene(t, pagina,
		"<title>Reporte del proyecto · Software Metrics</title>",
		"Resumen del proyecto", "Sprints", "1 cerrado de 2", "SP completados", "Velocidad del equipo",
		"5,0 SP por sprint (1 sprint cerrado)", "Horas estimadas y reales", "Desviación", "Defectos",
		"Evolución de la velocidad",
		`href="/proyectos/1/sprints/1/reporte"`, `href="/proyectos/1/sprints/2/reporte"`,
		"El sprint no está cerrado: el reporte muestra cómo va hasta hoy.")
	evolucion := pagina[strings.Index(pagina, "Evolución de la velocidad"):]
	evolucion = evolucion[:strings.Index(evolucion, "</table>")]
	verificarContiene(t, evolucion, "Sprint 1", "8", "5", "62,5 %", "5,0 SP por sprint")
	if strings.Contains(evolucion, "Sprint 2") {
		t.Error("el Sprint 2 está abierto: no va en la evolución de la velocidad")
	}
	uno, dos := strings.Index(pagina, `id="sprint-1"`), strings.Index(pagina, `id="sprint-2"`)
	if uno < 0 || dos < uno {
		t.Error("tiene que haber un bloque por sprint, el Sprint 1 antes que el Sprint 2")
	}
}

// Sección 3 · el reporte se imprime sin la navegación ni los botones, y cada sprint arranca en una hoja nueva
func TestReporteDeProyecto_SeImprime(t *testing.T) {
	pagina := get(conSprintCerrado(t), rutaDelReporteDelProyecto).Body.String()

	verificarContiene(t, pagina, `onclick="window.print()"`, "Imprimir", "break-before: page")
	if strings.Count(pagina, "d-print-none") < 3 {
		t.Error("la barra de navegación, las migas y el botón Imprimir tienen que ocultarse al imprimir (d-print-none)")
	}
}

// Sección 2 · la página del proyecto tiene el botón para ver su reporte
func TestProyecto_EnlazaAlReporte(t *testing.T) {
	verificarContiene(t, get(conProyecto(t), "/proyectos/1").Body.String(),
		`href="/proyectos/1/reporte"`, "Reporte del proyecto")
}

// RN5 · un proyecto sin sprints muestra el reporte sin bloques ni evolución
func TestReporteDeProyecto_SinSprints(t *testing.T) {
	rec := get(conProyecto(t), rutaDelReporteDelProyecto)

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	verificarContiene(t, rec.Body.String(), "sin datos", "Aún no hay sprints cerrados", "Todavía no hay sprints en este proyecto.")
}

// Sección 7 · un proyecto que no existe da 404
func TestReporteDeProyecto_NoExiste(t *testing.T) {
	h := conProyecto(t)
	for _, ruta := range []string{"/proyectos/9/reporte", "/proyectos/abc/reporte"} {
		rec := get(h, ruta)
		if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), "No encontramos ese proyecto.") {
			t.Errorf("%s: se esperaba 404 con %q y se obtuvo %d con %q", ruta, "No encontramos ese proyecto.", rec.Code, rec.Body.String())
		}
	}
}
