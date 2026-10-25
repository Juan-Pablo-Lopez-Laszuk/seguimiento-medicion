package server_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/defect"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// conGraficos arma la aplicación con el proyecto 1 (hoy 15/10/2026) y, si conSprint, el Sprint 1 (12/10 al 18/10)
// con la HU-1 (5 SP, completada el 13/10) y la HU-2 (3 SP, sin completar), y un defecto Alto abierto.
func conGraficos(t *testing.T, conSprint bool) http.Handler {
	t.Helper()
	ctx := context.Background()
	ahora := func() time.Time { return time.Date(2026, 10, 15, 18, 0, 0, 0, time.UTC) }
	proyectos, sprints, historias := memory.NuevoProyectos(), memory.NuevoSprints(), memory.NuevoHistorias()
	tareas, registros, repoDefectos := memory.NuevoTareas(), memory.NuevoRegistros(), memory.NuevoDefectos()
	cierres := service.NuevoCierres(proyectos, sprints, historias, memory.NuevoCierres(sprints, historias), ahora)
	horas := service.NuevoHoras(proyectos, sprints, historias, tareas, registros)
	defectos := service.NuevoDefectos(proyectos, sprints, historias, repoDefectos)
	estado := service.NuevoEstadoProyectos(proyectos, sprints, historias, ahora)
	panel := service.NuevoPanel(estado, sprints, horas, defectos)
	reportes := service.NuevoReportes(proyectos, sprints, historias, cierres, horas, repoDefectos, panel)

	p, err := proyectos.Guardar(ctx, project.Proyecto{Nombre: "Software Metrics", FechaInicio: time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC), FechaFin: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("guardar el proyecto: %v", err)
	}
	if conSprint {
		s, err := sprints.Guardar(ctx, sprint.Sprint{ProyectoID: p.ID, Numero: 1, Objetivo: "MVP", FechaInicio: time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC), FechaFin: time.Date(2026, 10, 18, 0, 0, 0, 0, time.UTC), Estado: sprint.EstadoPlanificado})
		if err != nil {
			t.Fatalf("guardar el sprint: %v", err)
		}
		cinco, tres, completada := 5, 3, time.Date(2026, 10, 13, 0, 0, 0, 0, time.UTC)
		for _, h := range []backlog.Historia{
			{Numero: 1, Titulo: "Login", Estado: backlog.EstadoHecho, StoryPoints: &cinco, CompletadaEn: &completada},
			{Numero: 2, Titulo: "Menú", Estado: backlog.EstadoEnSprint, StoryPoints: &tres},
		} {
			h.ProyectoID, h.Prioridad, h.SprintID = p.ID, backlog.PrioridadAlta, &s.ID
			if _, err := historias.Guardar(ctx, h); err != nil {
				t.Fatalf("guardar la HU-%d: %v", h.Numero, err)
			}
		}
		if _, err := repoDefectos.Guardar(ctx, defect.Defecto{ProyectoID: p.ID, Descripcion: "Login lento", Severidad: defect.SeveridadAlta, Estado: defect.EstadoAbierto, SprintDeteccionID: s.ID}); err != nil {
			t.Fatalf("guardar el defecto: %v", err)
		}
	}
	return server.NewRouter(server.Dependencias{
		Proyectos: service.NuevoProyectos(proyectos, ahora), Estado: estado, Panel: panel, Reportes: reportes,
		Graficos: service.NuevoGraficos(reportes, defectos, historias, ahora),
	})
}

// HU-36 · CA-36.2 · el endpoint devuelve en JSON los datos de los cinco gráficos
func TestDatosDeGraficos_RespondeJSON(t *testing.T) {
	rec := get(conGraficos(t, true), "/proyectos/1/graficos/datos")

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q", ct)
	}
	var d struct {
		Velocidad   map[string][]any `json:"velocidad"`
		StoryPoints struct {
			Sprints      []string  `json:"sprints"`
			Planificados []float64 `json:"planificados"`
			Completados  []float64 `json:"completados"`
		} `json:"storyPoints"`
		Horas    map[string][]any `json:"horas"`
		Defectos struct {
			Severidades []string `json:"severidades"`
			Detectados  []int    `json:"detectados"`
		} `json:"defectos"`
		Burndown struct {
			Sprint     string     `json:"sprint"`
			Dias       []string   `json:"dias"`
			Ideal      []float64  `json:"ideal"`
			Pendientes []*float64 `json:"pendientes"`
		} `json:"burndown"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &d); err != nil {
		t.Fatalf("la respuesta no es JSON válido: %v\n%s", err, rec.Body.String())
	}
	if len(d.StoryPoints.Sprints) != 1 || d.StoryPoints.Planificados[0] != 8 || d.StoryPoints.Completados[0] != 5 {
		t.Errorf("storyPoints = %+v, se esperaba el Sprint 1 con 8 planificados y 5 completados", d.StoryPoints)
	}
	if d.Velocidad == nil || len(d.Velocidad["sprints"]) != 0 || d.Horas == nil {
		t.Errorf("velocidad = %v y horas = %v: se esperaban series vacías (sin sprints cerrados)", d.Velocidad, d.Horas)
	}
	if d.Defectos.Severidades[1] != "Alta" || d.Defectos.Detectados[1] != 1 {
		t.Errorf("defectos = %+v", d.Defectos)
	}
	b := d.Burndown
	if b.Sprint != "Sprint 1" || len(b.Dias) != 7 || b.Dias[0] != "12/10" || b.Ideal[0] != 8 || b.Pendientes[1] == nil || *b.Pendientes[1] != 3 || b.Pendientes[4] != nil {
		t.Errorf("burndown = %+v, se esperaban 7 días desde el 12/10, 3 pendientes el 13/10 y sin dato el 16/10", b)
	}
}

// HU-36 · RN5 · un proyecto sin sprints responde sin burndown
func TestDatosDeGraficos_SinSprints_BurndownNulo(t *testing.T) {
	rec := get(conGraficos(t, false), "/proyectos/1/graficos/datos")

	var d map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &d); err != nil || rec.Code != http.StatusOK {
		t.Fatalf("código %d, %v", rec.Code, err)
	}
	if b, ok := d["burndown"]; !ok || b != nil {
		t.Errorf("burndown = %v, se esperaba null", b)
	}
}

// HU-36 · sección 7 · un proyecto que no existe responde 404 en JSON
func TestDatosDeGraficos_ProyectoInexistente_404EnJSON(t *testing.T) {
	for _, ruta := range []string{"/proyectos/99/graficos/datos", "/proyectos/abc/graficos/datos"} {
		rec := get(conGraficos(t, false), ruta)

		var d map[string]string
		if rec.Code != http.StatusNotFound || json.Unmarshal(rec.Body.Bytes(), &d) != nil || d["error"] != "No encontramos ese proyecto." {
			t.Errorf("%s: código %d y cuerpo %q, se esperaba 404 con el error en JSON", ruta, rec.Code, rec.Body.String())
		}
	}
}
