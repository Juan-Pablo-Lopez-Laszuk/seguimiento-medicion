package server_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

const tituloDelHistorial = "Historial de estimaciones"

// conHistorialDeEstimaciones arma la aplicación con el proyecto de ejemplo (ID 1) y sus dos historias, HU-1 y HU-2, sin
// estimar. El reloj de los repositorios avanza una hora por cada estimación, desde las 18:00 (UTC) del 20/10/2026.
func conHistorialDeEstimaciones(t *testing.T) (http.Handler, *memory.Historias) {
	t.Helper()
	ctx := context.Background()
	hora := time.Date(2026, 10, 20, 17, 0, 0, 0, time.UTC)
	proyectos, criterios := memory.NuevoProyectos(), memory.NuevoCriterios()
	historias := memory.NuevoHistorias().ConReloj(func() time.Time {
		hora = hora.Add(time.Hour)
		return hora
	})
	servicioProyectos := service.NuevoProyectos(proyectos, time.Now)
	p, err := servicioProyectos.Crear(ctx, project.Datos{
		Nombre:      "Software Metrics",
		FechaInicio: time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC),
		FechaFin:    time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("no se pudo crear el proyecto: %v", err)
	}
	for numero := 1; numero <= 2; numero++ {
		if _, err := historias.Guardar(ctx, backlog.Historia{
			ProyectoID: p.ID, Numero: numero, Titulo: "Historia", Prioridad: backlog.PrioridadAlta, Estado: backlog.EstadoPendiente,
		}); err != nil {
			t.Fatalf("no se pudo guardar la historia %d: %v", numero, err)
		}
	}
	return server.NewRouter(server.Dependencias{
		Proyectos:    servicioProyectos,
		Historias:    service.NuevoHistorias(proyectos, historias),
		Criterios:    service.NuevoCriterios(proyectos, historias, criterios).ConHistorial(historias),
		Estimaciones: service.NuevoEstimaciones(proyectos, historias),
	}), historias
}

// CA-16.3 · una historia que nunca se re-estimó lo dice, y la primera estimación no se anota
func TestPaginaDeLaHistoria_SinReestimar_DiceQueNoSeReestimo(t *testing.T) {
	h, _ := conHistorialDeEstimaciones(t)
	post(h, "/proyectos/1/backlog/1/estimacion", puntos("5"))

	html := get(h, "/proyectos/1/backlog/1").Body.String()

	verificarContiene(t, html, tituloDelHistorial, "Todavía no se re-estimó esta historia.")
	if strings.Contains(html, "→") {
		t.Error("la primera estimación no tiene valor anterior: no tendría que haber ninguna entrada")
	}
}

// CA-16.3 · re-estimar deja el valor anterior en el historial, con la hora de Argentina y el cambio más reciente primero
func TestPaginaDeLaHistoria_MuestraElHistorialDeEstimaciones(t *testing.T) {
	h, _ := conHistorialDeEstimaciones(t)
	for _, valor := range []string{"5", "8", "3"} { // las horas 18:00, 19:00 y 20:00 UTC son las 15:00, 16:00 y 17:00 en Argentina
		post(h, "/proyectos/1/backlog/1/estimacion", puntos(valor))
	}

	html := get(h, "/proyectos/1/backlog/1").Body.String()

	verificarContiene(t, html, tituloDelHistorial, "20/10/2026 16:00", "5 → 8", "20/10/2026 17:00", "8 → 3")
	if strings.Contains(html, "Todavía no se re-estimó esta historia.") {
		t.Error("una historia re-estimada no tendría que decir que nunca se re-estimó")
	}
	if strings.Index(html, "8 → 3") > strings.Index(html, "5 → 8") {
		t.Error("el cambio más reciente (8 → 3) tiene que ir antes que el más antiguo (5 → 8)")
	}
}

// RN7 · el historial es de cada historia
func TestPaginaDeLaHistoria_ElHistorialEsDeCadaHistoria(t *testing.T) {
	h, _ := conHistorialDeEstimaciones(t)
	post(h, "/proyectos/1/backlog/1/estimacion", puntos("5"))
	post(h, "/proyectos/1/backlog/1/estimacion", puntos("8"))

	deLaSegunda := get(h, "/proyectos/1/backlog/2").Body.String()

	verificarContiene(t, deLaSegunda, "Todavía no se re-estimó esta historia.")
	if strings.Contains(deLaSegunda, "5 → 8") {
		t.Error("la HU-2 no tiene que mostrar el historial de la HU-1")
	}
}

// RN5 · un valor inválido no anota nada, y la página que muestra el error conserva el historial
func TestEstimarHistoria_ValorInvalido_ConservaElHistorial(t *testing.T) {
	h, _ := conHistorialDeEstimaciones(t)
	post(h, "/proyectos/1/backlog/1/estimacion", puntos("5"))
	post(h, "/proyectos/1/backlog/1/estimacion", puntos("8"))

	rec := post(h, "/proyectos/1/backlog/1/estimacion", puntos("4"))

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusUnprocessableEntity)
	}
	html := rec.Body.String()
	verificarContiene(t, html, tituloDelHistorial, "5 → 8")
	if strings.Count(html, "→") != 1 {
		t.Error("el intento inválido no tendría que anotar nada: tendría que haber una sola entrada")
	}
}

// RN5 y RN6 · una historia Hecha no se re-estima, pero su historial se sigue viendo
func TestPaginaDeLaHistoria_Hecha_MuestraElHistorial(t *testing.T) {
	h, historias := conHistorialDeEstimaciones(t)
	post(h, "/proyectos/1/backlog/1/estimacion", puntos("5"))
	post(h, "/proyectos/1/backlog/1/estimacion", puntos("8"))
	if err := historias.ActualizarEstado(context.Background(), 1, backlog.EstadoHecho); err != nil {
		t.Fatalf("no se pudo dejar la historia Hecha: %v", err)
	}

	html := get(h, "/proyectos/1/backlog/1").Body.String()

	verificarContiene(t, html, tituloDelHistorial, "5 → 8", "La estimación no se puede cambiar: la historia está Hecha.")
}
