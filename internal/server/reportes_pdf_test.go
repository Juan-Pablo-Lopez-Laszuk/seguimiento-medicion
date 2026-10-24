package server_test

import (
	"net/http"
	"strings"
	"testing"
)

// verificarPDF controla que la respuesta sea un PDF que el navegador descarga con el nombre esperado.
func verificarPDF(t *testing.T, h http.Handler, ruta, archivo string) {
	t.Helper()
	rec := get(h, ruta)

	if rec.Code != http.StatusOK {
		t.Fatalf("%s: código = %d, se esperaba %d (%s)", ruta, rec.Code, http.StatusOK, rec.Body.String())
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/pdf" {
		t.Errorf("%s: Content-Type = %q, se esperaba application/pdf", ruta, ct)
	}
	if cd, esperado := rec.Header().Get("Content-Disposition"), `attachment; filename="`+archivo+`"`; cd != esperado {
		t.Errorf("%s: Content-Disposition = %q, se esperaba %q", ruta, cd, esperado)
	}
	if !strings.HasPrefix(rec.Body.String(), "%PDF-") {
		t.Errorf("%s: la respuesta no es un PDF", ruta)
	}
}

// CA-39.1 · el PDF del reporte de un sprint se descarga como proyecto-sprint-fecha.pdf, con la fecha de hoy
// (13/10/2026 en los tests)
func TestPDFDelSprint_SeDescarga(t *testing.T) {
	h := conSprintCerrado(t)

	verificarPDF(t, h, "/proyectos/1/sprints/1/reporte.pdf", "software-metrics-sprint-1-2026-10-13.pdf")
	verificarPDF(t, h, "/proyectos/1/sprints/2/reporte.pdf", "software-metrics-sprint-2-2026-10-13.pdf")
}

// CA-39.1 · el PDF del reporte del proyecto no lleva el sprint en el nombre
func TestPDFDelProyecto_SeDescarga(t *testing.T) {
	verificarPDF(t, conSprintCerrado(t), "/proyectos/1/reporte.pdf", "software-metrics-2026-10-13.pdf")
}

// RN5 de HU-38 · el PDF de un proyecto sin sprints también se genera
func TestPDFDelProyecto_SinSprints(t *testing.T) {
	verificarPDF(t, conProyecto(t), "/proyectos/1/reporte.pdf", "software-metrics-2026-10-13.pdf")
}

// Sección 7 · un proyecto o un sprint que no existe da 404 con el mensaje del reporte
func TestPDF_NoExiste(t *testing.T) {
	h := conSprintCerrado(t)
	casos := map[string]string{
		"/proyectos/9/reporte.pdf":             "No encontramos ese proyecto.",
		"/proyectos/abc/reporte.pdf":           "No encontramos ese proyecto.",
		"/proyectos/9/sprints/1/reporte.pdf":   "No encontramos ese proyecto.",
		"/proyectos/1/sprints/9/reporte.pdf":   "No encontramos ese sprint.",
		"/proyectos/1/sprints/abc/reporte.pdf": "No encontramos ese sprint.",
	}
	for ruta, mensaje := range casos {
		rec := get(h, ruta)
		if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), mensaje) {
			t.Errorf("%s: se esperaba 404 con %q y se obtuvo %d con %q", ruta, mensaje, rec.Code, rec.Body.String())
		}
	}
}

// Sección 2 · las dos páginas de reporte tienen el botón Descargar PDF al lado de Imprimir, que no se imprime
func TestReportes_EnlazanAlPDF(t *testing.T) {
	h := conSprintCerrado(t)
	casos := map[string]string{
		"/proyectos/1/sprints/1/reporte": `href="/proyectos/1/sprints/1/reporte.pdf"`,
		"/proyectos/1/reporte":           `href="/proyectos/1/reporte.pdf"`,
	}
	for pagina, enlace := range casos {
		html := get(h, pagina).Body.String()
		i := strings.Index(html, enlace)
		if i < 0 {
			t.Errorf("%s: no tiene el enlace %s", pagina, enlace)
			continue
		}
		boton := html[strings.LastIndex(html[:i], "<a "):]
		boton = boton[:strings.Index(boton, "</a>")]
		verificarContiene(t, boton, "Descargar PDF", "d-print-none")
	}
}
