package server_test

import (
	"net/http"
	"testing"
)

// Al panel se llega desde la vista del proyecto.
func TestProyecto_EnlazaAlPanel(t *testing.T) {
	verificarContiene(t, get(conProyecto(t), "/proyectos/1").Body.String(), `href="/proyectos/1/panel"`)
}

// HU-35 · CA-35.1 y CA-35.2 · el panel muestra las seis tarjetas; sin datos dice "sin datos" y no da error
func TestPanel_SinDatos_MuestraLasSeisTarjetas(t *testing.T) {
	rec := get(conProyecto(t), "/proyectos/1/panel")
	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	verificarContiene(t, rec.Body.String(),
		"<title>Panel · Software Metrics</title>", `data-tarjeta="Estado"`, `data-tarjeta="Sprint activo">sin datos<`,
		`data-tarjeta="Velocidad">sin datos<`, `data-tarjeta="% completado">sin datos<`,
		`data-tarjeta="Desviación de horas">sin datos<`, `data-tarjeta="Defectos abiertos">sin datos<`,
	)
}

// HU-35 · CA-35.1 · con un defecto abierto, la tarjeta lo cuenta
func TestPanel_ConUnDefecto(t *testing.T) {
	verificarContiene(t, get(conDefecto(t), "/proyectos/1/panel").Body.String(),
		`data-tarjeta="Defectos abiertos">1<`, "De 1 detectados")
}

func TestPanel_ProyectoInexistente_Responde404(t *testing.T) {
	if rec := get(conProyecto(t), "/proyectos/99/panel"); rec.Code != http.StatusNotFound {
		t.Errorf("código = %d, se esperaba %d", rec.Code, http.StatusNotFound)
	}
}
