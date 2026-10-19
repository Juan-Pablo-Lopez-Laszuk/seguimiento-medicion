package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// HU-33 · CA-33.1 a CA-33.3 · desviación por historia, por sprint y del proyecto
func TestVerHoras_DesviacionPorHistoriaSprintYProyecto(t *testing.T) {
	ctx := context.Background()
	proyectos, sprints, historias := memory.NuevoProyectos(), memory.NuevoSprints(), memory.NuevoHistorias()
	tareas, registros := memory.NuevoTareas(), memory.NuevoRegistros()
	p, err := service.NuevoProyectos(proyectos, relojFijo).Crear(ctx, datosValidos())
	if err != nil {
		t.Fatalf("crear el proyecto: %v", err)
	}
	s, err := service.NuevoSprints(proyectos, sprints).Crear(ctx, p.ID, datosDeSprint("2026-10-05", "2026-10-11"))
	if err != nil {
		t.Fatalf("crear el sprint: %v", err)
	}
	casosHistorias := service.NuevoHistorias(proyectos, historias)
	// Dos historias en el Sprint 1 (10 h estimadas y 12 reales; 5 y 4) y una sin sprint (sin tareas y 2 reales).
	for _, c := range []struct {
		titulo   string
		enSprint bool
		tareas   []float64
		reales   float64
	}{{"Login", true, []float64{4, 6}, 12}, {"Logout", true, []float64{5}, 4}, {"Perfil", false, nil, 2}} {
		h, err := casosHistorias.Crear(ctx, p.ID, datosDeHistoria(c.titulo))
		if err != nil {
			t.Fatalf("crear la historia: %v", err)
		}
		if c.enSprint {
			if err := historias.ActualizarSprint(ctx, h.ID, backlog.EstadoEnSprint, &s.ID); err != nil {
				t.Fatalf("asignar al sprint: %v", err)
			}
		}
		for _, horas := range c.tareas {
			if _, err := tareas.Guardar(ctx, effort.Tarea{HistoriaID: h.ID, Titulo: "Tarea", HorasEstimadas: horas}); err != nil {
				t.Fatalf("guardar la tarea: %v", err)
			}
		}
		if _, err := registros.Guardar(ctx, effort.Registro{HistoriaID: h.ID, Horas: c.reales, Actividad: "Trabajo"}); err != nil {
			t.Fatalf("guardar el registro: %v", err)
		}
	}

	v, err := service.NuevoHoras(proyectos, sprints, historias, tareas, registros).Ver(ctx, p.ID)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if v.Proyecto.ID != p.ID || len(v.Historias) != 3 || len(v.Sprints) != 1 {
		t.Fatalf("vista = %+v", v)
	}
	for _, c := range []struct {
		nombre, horas, porcentaje string
		got                       interface {
			HorasTexto() string
			PorcentajeTexto() string
		}
	}{
		{"HU-1", "+2 h", "+20,0 %", v.Historias[0]},
		{"HU-2", "-1 h", "-20,0 %", v.Historias[1]},
		{"HU-3", "+2 h", "no aplicable", v.Historias[2]},
		{"Sprint 1", "+1 h", "+6,7 %", v.Sprints[0]},
		{"proyecto", "+3 h", "+20,0 %", v.Total},
	} {
		if c.got.HorasTexto() != c.horas || c.got.PorcentajeTexto() != c.porcentaje {
			t.Errorf("%s: %s y %s, se esperaba %s y %s", c.nombre, c.got.HorasTexto(), c.got.PorcentajeTexto(), c.horas, c.porcentaje)
		}
	}
	if v.Historias[0].Nombre != "HU-1 · Login" || v.Sprints[0].Nombre != "Sprint 1" {
		t.Errorf("nombres = %q y %q", v.Historias[0].Nombre, v.Sprints[0].Nombre)
	}
}

func TestVerHoras_ProyectoInexistente(t *testing.T) {
	casos := service.NuevoHoras(memory.NuevoProyectos(), memory.NuevoSprints(), memory.NuevoHistorias(), memory.NuevoTareas(), memory.NuevoRegistros())
	if _, err := casos.Ver(context.Background(), 99); !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("se esperaba %v y se obtuvo %v", project.ErrNoEncontrado, err)
	}
}
