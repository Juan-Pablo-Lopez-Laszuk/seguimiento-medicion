package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/defect"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// escenarioPanel tiene los repositorios en memoria y el panel armado con un reloj fijo en el 14/10 (Sprint 2 activo).
type escenarioPanel struct {
	proyectos *memory.Proyectos
	sprints   *memory.Sprints
	historias *memory.Historias
	tareas    *memory.Tareas
	registros *memory.Registros
	defectos  *service.Defectos
	panel     *service.Panel
	proyecto  int64
}

func nuevoEscenarioPanel(t *testing.T) escenarioPanel {
	t.Helper()
	e := escenarioPanel{proyectos: memory.NuevoProyectos(), sprints: memory.NuevoSprints(), historias: memory.NuevoHistorias(),
		tareas: memory.NuevoTareas(), registros: memory.NuevoRegistros()}
	reloj := func() time.Time { return time.Date(2026, 10, 14, 15, 0, 0, 0, time.UTC) }
	p, err := service.NuevoProyectos(e.proyectos, relojFijo).Crear(context.Background(), datosValidos())
	if err != nil {
		t.Fatalf("crear el proyecto: %v", err)
	}
	e.proyecto = p.ID
	e.defectos = service.NuevoDefectos(e.proyectos, e.sprints, e.historias, memory.NuevoDefectos())
	e.panel = service.NuevoPanel(
		service.NuevoEstadoProyectos(e.proyectos, e.sprints, e.historias, reloj),
		e.sprints,
		service.NuevoHoras(e.proyectos, e.sprints, e.historias, e.tareas, e.registros),
		e.defectos,
	)
	return e
}

func valorDe(t *testing.T, v service.VistaPanel, titulo string) string {
	t.Helper()
	for _, tarjeta := range v.Tarjetas() {
		if tarjeta.Titulo == titulo {
			return tarjeta.Valor
		}
	}
	t.Fatalf("el panel no tiene la tarjeta %q", titulo)
	return ""
}

// HU-35 · CA-35.2 · un proyecto sin datos muestra "sin datos" en cada tarjeta y no da error
func TestPanel_SinDatos(t *testing.T) {
	e := nuevoEscenarioPanel(t)
	v, err := e.panel.Ver(context.Background(), e.proyecto)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if len(v.Tarjetas()) != 6 || valorDe(t, v, "Estado") == "" {
		t.Errorf("tarjetas = %+v", v.Tarjetas())
	}
	for _, titulo := range []string{"Sprint activo", "Velocidad", "% completado", "Desviación de horas", "Defectos abiertos"} {
		if got := valorDe(t, v, titulo); got != "sin datos" {
			t.Errorf("%s = %q, se esperaba \"sin datos\"", titulo, got)
		}
	}
}

// HU-35 · CA-35.1 · cada tarjeta muestra su métrica
func TestPanel_ConDatos(t *testing.T) {
	e := nuevoEscenarioPanel(t)
	ctx := context.Background()
	casosSprints := service.NuevoSprints(e.proyectos, e.sprints)
	s1, err := casosSprints.Crear(ctx, e.proyecto, datosDeSprint("2026-10-05", "2026-10-11"))
	if err != nil {
		t.Fatalf("crear el Sprint 1: %v", err)
	}
	if _, err := casosSprints.Crear(ctx, e.proyecto, datosDeSprint("2026-10-12", "2026-10-18")); err != nil {
		t.Fatalf("crear el Sprint 2: %v", err)
	}
	ocho := 8
	s1.SPCompletados = &ocho
	if err := memory.NuevoCierres(e.sprints, e.historias).Cerrar(ctx, sprint.Cierre{Sprint: s1}); err != nil {
		t.Fatalf("cerrar el Sprint 1: %v", err)
	}
	casosHistorias := service.NuevoHistorias(e.proyectos, e.historias)
	for i, titulo := range []string{"Hecha", "Pendiente"} {
		h, err := casosHistorias.Crear(ctx, e.proyecto, datosDeHistoria(titulo))
		if err != nil {
			t.Fatalf("crear la historia: %v", err)
		}
		if err := e.historias.ActualizarStoryPoints(ctx, h.ID, 5); err != nil {
			t.Fatalf("estimar: %v", err)
		}
		if i == 0 {
			if err := e.historias.ActualizarCompletada(ctx, h.ID, backlog.EstadoHecho, nil); err != nil {
				t.Fatalf("completar: %v", err)
			}
			if _, err := e.tareas.Guardar(ctx, effort.Tarea{HistoriaID: h.ID, Titulo: "Tarea", HorasEstimadas: 10}); err != nil {
				t.Fatalf("tarea: %v", err)
			}
			if _, err := e.registros.Guardar(ctx, effort.Registro{HistoriaID: h.ID, Horas: 12, Actividad: "Trabajo"}); err != nil {
				t.Fatalf("registro: %v", err)
			}
		}
	}
	for i := 0; i < 3; i++ {
		d, err := e.defectos.Registrar(ctx, e.proyecto, defect.Datos{Descripcion: "Falla", Severidad: defect.SeveridadMedia, SprintID: s1.ID})
		if err != nil {
			t.Fatalf("registrar defecto: %v", err)
		}
		if i == 0 {
			for _, paso := range []defect.Estado{defect.EstadoEnProgreso, defect.EstadoResuelto} {
				if _, err := e.defectos.CambiarEstado(ctx, e.proyecto, d.ID, paso, s1.ID); err != nil {
					t.Fatalf("a %s: %v", paso, err)
				}
			}
		}
	}

	v, err := e.panel.Ver(ctx, e.proyecto)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	for titulo, esperado := range map[string]string{
		"Estado": "En curso", "Sprint activo": "Sprint 2", "Velocidad": "8 SP por sprint", "% completado": "50,0 %",
		"Desviación de horas": "+2 h", "Defectos abiertos": "2",
	} {
		if got := valorDe(t, v, titulo); got != esperado {
			t.Errorf("%s = %q, se esperaba %q", titulo, got, esperado)
		}
	}
}

func TestPanel_ProyectoInexistente(t *testing.T) {
	e := nuevoEscenarioPanel(t)
	if _, err := e.panel.Ver(context.Background(), 99); !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("se esperaba %v y se obtuvo %v", project.ErrNoEncontrado, err)
	}
}
