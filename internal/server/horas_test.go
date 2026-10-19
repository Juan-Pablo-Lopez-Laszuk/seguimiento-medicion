package server_test

import (
	"net/http"
	"testing"
)

// A la página de horas se llega desde la vista del proyecto.
func TestProyecto_EnlazaASusHoras(t *testing.T) {
	verificarContiene(t, get(conProyecto(t), "/proyectos/1").Body.String(), `href="/proyectos/1/horas"`)
}

// HU-33 · CA-33.1 a CA-33.3 · la página muestra horas estimadas, reales y desviación del proyecto, de cada sprint y de
// cada historia
func TestPaginaDeHoras_MuestraLaDesviacion(t *testing.T) {
	h := conHistoriaEIntegrantes(t)
	agregarTarea(t, h) // 2,5 h estimadas
	if rec := post(h, "/proyectos/1/backlog/1/esfuerzo", registroValido()); rec.Code != http.StatusSeeOther {
		t.Fatalf("no se pudo registrar el esfuerzo: código %d", rec.Code)
	}

	rec := get(h, "/proyectos/1/horas")

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	verificarContiene(t, rec.Body.String(),
		"<title>Horas · Software Metrics</title>", "Total del proyecto", "HU-1 · ", "2,5 h", "3 h", "&#43;0,5 h", "&#43;20,0 %", // html/template escribe el + como &#43;
		"Sin sprints todavía",
	)
}

// HU-33 · CA-33.2 · sin horas estimadas el % dice "no aplicable"
func TestPaginaDeHoras_SinEstimadas_NoAplica(t *testing.T) {
	verificarContiene(t, get(conSprintEHistoria(t), "/proyectos/1/horas").Body.String(), "Sprint 1", "no aplicable", "0 h")
}

func TestHoras_ProyectoInexistente_Responde404(t *testing.T) {
	if rec := get(conProyecto(t), "/proyectos/99/horas"); rec.Code != http.StatusNotFound {
		t.Errorf("código = %d, se esperaba %d", rec.Code, http.StatusNotFound)
	}
}
