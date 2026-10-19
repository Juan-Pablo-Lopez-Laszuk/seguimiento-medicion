package app_test

import (
	"net/http"
	neturl "net/url"
	"strings"
	"testing"
	"time"
)

// conSprintActivoHoy crea un proyecto con un sprint que es el activo el día en que corre el test (la app usa la hora
// real), y le asigna la HU-1 (5 SP, con un criterio sin cumplir).
func conSprintActivoHoy(t *testing.T, h http.Handler) {
	t.Helper()
	hoy := time.Now()
	dia := func(d int) string { return hoy.AddDate(0, 0, d).Format(time.DateOnly) }
	pasos := []struct {
		ruta   string
		campos neturl.Values
	}{
		{"/proyectos", neturl.Values{"nombre": {"Con cierre"}, "descripcion": {"HU-14"}, "fecha_inicio": {dia(-30)}, "fecha_fin": {dia(30)}}},
		{"/proyectos/1/sprints", neturl.Values{"objetivo": {"El que se cierra"}, "fecha_inicio": {dia(-2)}, "fecha_fin": {dia(5)}}},
		{"/proyectos/1/backlog", neturl.Values{"titulo": {"Historia del sprint"}, "prioridad": {"Alta"}, "story_points": {"5"}}},
		{"/proyectos/1/backlog/1/criterios", neturl.Values{"descripcion": {"Un criterio"}}},
		{"/proyectos/1/sprints/1/historias", neturl.Values{"historia": {"1"}}},
	}
	for _, p := range pasos {
		if rec := post(h, p.ruta, p.campos); rec.Code != http.StatusSeeOther {
			t.Fatalf("%s: se esperaba %d y se obtuvo %d:\n%s", p.ruta, http.StatusSeeOther, rec.Code, rec.Body.String())
		}
	}
}

// verificarCerrado revisa que el Sprint 1 esté cerrado con 0 de 5 SP, la HU-1 haya vuelto al backlog y ya no se pueda
// eliminar, porque está en la foto del cierre (HU-06 RN6).
func verificarCerrado(t *testing.T, h http.Handler) {
	t.Helper()
	sprint := get(h, "/proyectos/1/sprints/1")
	if !strings.Contains(sprint, "Cerrado") || !strings.Contains(sprint, "0 de 5 SP completados") {
		t.Errorf("el Sprint 1 tendría que estar Cerrado con 0 de 5 SP completados:\n%s", sprint)
	}
	if historia := get(h, "/proyectos/1/backlog/1"); !strings.Contains(historia, "Pendiente") {
		t.Error("la HU-1 tendría que haber vuelto al backlog como Pendiente")
	}
	if eliminar := get(h, "/proyectos/1/backlog/1/eliminar"); !strings.Contains(eliminar, "es de un sprint cerrado") {
		t.Error("la HU-1 está en la foto del sprint cerrado y no se tendría que poder eliminar")
	}
}

// HU-14: sin base, cerrar el sprint también funciona en memoria, con la foto que impide eliminar la historia devuelta.
func TestNewHandlerConBase_SinURLSeCierraElSprint(t *testing.T) {
	h := nuevaApp(t, "")
	conSprintActivoHoy(t, h)

	if rec := post(h, "/proyectos/1/sprints/1/cerrar", nil); rec.Code != http.StatusSeeOther {
		t.Fatalf("cerrar: se esperaba %d y se obtuvo %d:\n%s", http.StatusSeeOther, rec.Code, rec.Body.String())
	}

	verificarCerrado(t, h)
}

// HU-14: el cierre queda en la base (el sprint, la foto y la historia devuelta) y sobrevive a un reinicio.
func TestNewHandlerConBase_ConURLElCierreSobreviveAUnReinicio(t *testing.T) {
	url := urlConEsquemaMigrado(t)
	primera := nuevaApp(t, url)
	conSprintActivoHoy(t, primera)
	if rec := post(primera, "/proyectos/1/sprints/1/cerrar", nil); rec.Code != http.StatusSeeOther {
		t.Fatalf("cerrar: se esperaba %d y se obtuvo %d:\n%s", http.StatusSeeOther, rec.Code, rec.Body.String())
	}

	reiniciada := nuevaApp(t, url)
	verificarCerrado(t, reiniciada)
	if rec := post(reiniciada, "/proyectos/1/sprints/1/cerrar", nil); rec.Code != http.StatusConflict {
		t.Errorf("cerrar otra vez después de reiniciar: se esperaba %d y se obtuvo %d", http.StatusConflict, rec.Code)
	}
}
