package service_test

import (
	"context"
	"errors"
	"reflect"
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

func dia36(d int) time.Time { return time.Date(2026, 10, d, 0, 0, 0, 0, time.UTC) }

// escenarioGraficos es el de features/hu-36-graficos.feature: Sprint 1 (05/10 al 11/10) cerrado con HU-1 (5 SP,
// completada el 08/10) y HU-2 (3 SP, el 10/10) de 13 SP; Sprint 2 (12/10 al 18/10) en curso con HU-4 (8 SP, el 13/10)
// y HU-5 (2 SP, el 15/10) de 13 SP; hoy es el 15/10. Horas: Sprint 1 10 h estimadas y 12 reales, Sprint 2 8 y 5.
// Defectos: una Alta resuelta en el Sprint 1 y una Crítica, una Alta y una Baja abiertas del Sprint 2.
type escenarioGraficos struct {
	proyectos *memory.Proyectos
	casos     *service.Graficos
	proyecto  int64
	ahora     time.Time
}

func nuevoEscenarioGraficos(t *testing.T) *escenarioGraficos {
	t.Helper()
	ctx := context.Background()
	proyectos, sprints, historias := memory.NuevoProyectos(), memory.NuevoSprints(), memory.NuevoHistorias()
	tareas, registros, defectos := memory.NuevoTareas(), memory.NuevoRegistros(), memory.NuevoDefectos()
	e := &escenarioGraficos{proyectos: proyectos}
	reloj := func() time.Time { return e.ahora }
	cierres := service.NuevoCierres(proyectos, sprints, historias, memory.NuevoCierres(sprints, historias), reloj)
	horas := service.NuevoHoras(proyectos, sprints, historias, tareas, registros)
	casosDefectos := service.NuevoDefectos(proyectos, sprints, historias, defectos)
	panel := service.NuevoPanel(service.NuevoEstadoProyectos(proyectos, sprints, historias, reloj), sprints, horas, casosDefectos)
	reportes := service.NuevoReportes(proyectos, sprints, historias, cierres, horas, defectos, panel)
	e.casos = service.NuevoGraficos(reportes, casosDefectos, historias, reloj)

	p, err := proyectos.Guardar(ctx, project.Proyecto{Nombre: "Tienda", FechaInicio: dia36(5), FechaFin: dia36(31)})
	if err != nil {
		t.Fatalf("guardar el proyecto: %v", err)
	}
	e.proyecto = p.ID
	ids := map[int]int64{}
	for n, fechas := range map[int][2]int{1: {5, 11}, 2: {12, 18}} {
		s, err := sprints.Guardar(ctx, sprint.Sprint{ProyectoID: p.ID, Numero: n, Objetivo: "Objetivo", FechaInicio: dia36(fechas[0]), FechaFin: dia36(fechas[1]), Estado: sprint.EstadoPlanificado})
		if err != nil {
			t.Fatalf("guardar el Sprint %d: %v", n, err)
		}
		ids[n] = s.ID
	}
	historiasPorNumero := map[int]int64{}
	for _, h := range []struct{ numero, sprint, sp, completada, estimadas, reales int }{
		{1, 1, 5, 8, 6, 7}, {2, 1, 3, 10, 4, 5}, {3, 1, 5, 0, 0, 0}, {4, 2, 8, 13, 8, 5}, {5, 2, 2, 15, 0, 0}, {6, 2, 3, 0, 0, 0},
	} {
		sp, sprintID := h.sp, ids[h.sprint]
		historia := backlog.Historia{ProyectoID: p.ID, Numero: h.numero, Titulo: "Historia", Prioridad: backlog.PrioridadAlta, Estado: backlog.EstadoEnSprint, StoryPoints: &sp, SprintID: &sprintID}
		if h.completada > 0 {
			fecha := dia36(h.completada)
			historia.Estado, historia.CompletadaEn = backlog.EstadoHecho, &fecha
		}
		guardada, err := historias.Guardar(ctx, historia)
		if err != nil {
			t.Fatalf("guardar la HU-%d: %v", h.numero, err)
		}
		historiasPorNumero[h.numero] = guardada.ID
		if h.estimadas > 0 {
			if _, err := tareas.Guardar(ctx, effort.Tarea{HistoriaID: guardada.ID, Titulo: "Desarrollo", HorasEstimadas: float64(h.estimadas)}); err != nil {
				t.Fatalf("guardar la tarea: %v", err)
			}
			if _, err := registros.Guardar(ctx, effort.Registro{HistoriaID: guardada.ID, Fecha: dia36(9), Actividad: "Desarrollo", Horas: float64(h.reales)}); err != nil {
				t.Fatalf("guardar el registro: %v", err)
			}
		}
	}
	e.ahora = dia36(11)
	if _, err := cierres.Cerrar(ctx, p.ID, 1); err != nil {
		t.Fatalf("cerrar el Sprint 1: %v", err)
	}
	for _, d := range []defect.Defecto{
		{Descripcion: "Login lento", Severidad: defect.SeveridadAlta, Estado: defect.EstadoResuelto, SprintDeteccionID: ids[1], SprintResolucionID: ids[1]},
		{Descripcion: "Botón roto", Severidad: defect.SeveridadCritica, Estado: defect.EstadoAbierto, SprintDeteccionID: ids[2]},
		{Descripcion: "Texto cortado", Severidad: defect.SeveridadBaja, Estado: defect.EstadoAbierto, SprintDeteccionID: ids[2]},
		{Descripcion: "Error de cálculo", Severidad: defect.SeveridadAlta, Estado: defect.EstadoEnProgreso, SprintDeteccionID: ids[2]},
	} {
		d.ProyectoID, d.HistoriaID = p.ID, historiasPorNumero[1]
		if _, err := defectos.Guardar(ctx, d); err != nil {
			t.Fatalf("guardar el defecto %q: %v", d.Descripcion, err)
		}
	}
	e.ahora = dia36(15).Add(18 * time.Hour) // las 15:00 de Argentina
	return e
}

func (e *escenarioGraficos) datos(t *testing.T, numeroSprint int) service.DatosGraficos {
	t.Helper()
	d, err := e.casos.Datos(context.Background(), e.proyecto, numeroSprint)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	return d
}

func verificarIgual(t *testing.T, nombre string, obtenido, quiero any) {
	t.Helper()
	if !reflect.DeepEqual(obtenido, quiero) {
		t.Errorf("%s = %#v, se esperaba %#v", nombre, obtenido, quiero)
	}
}

// CA-36.1 · RN1 · SP, velocidad y horas por sprint (el cerrado con su foto) y defectos por severidad
func TestDatosDeGraficos_SeriesPorSprint(t *testing.T) {
	d := nuevoEscenarioGraficos(t).datos(t, 0)

	dos := []string{"Sprint 1", "Sprint 2"}
	verificarIgual(t, "story points", d.StoryPoints, service.GraficoStoryPoints{Sprints: dos, Planificados: []float64{13, 13}, Completados: []float64{8, 10}})
	verificarIgual(t, "velocidad", d.Velocidad, service.GraficoVelocidad{Sprints: []string{"Sprint 1"}, Completados: []float64{8}, Velocidad: []float64{8}})
	verificarIgual(t, "horas", d.Horas, service.GraficoHoras{Sprints: dos, Estimadas: []float64{10, 8}, Reales: []float64{12, 5}})
	verificarIgual(t, "defectos", d.Defectos, service.GraficoDefectos{
		Severidades: []string{"Crítica", "Alta", "Media", "Baja"}, Detectados: []int{1, 2, 0, 1}, Abiertos: []int{1, 1, 0, 1},
	})
	verificarIgual(t, "sprints para elegir", d.Sprints, []int{1, 2})
	if d.Proyecto.Nombre != "Tienda" {
		t.Errorf("proyecto = %q", d.Proyecto.Nombre)
	}
}

func pendientes(valores ...float64) []*float64 {
	p := make([]*float64, len(valores))
	for i, v := range valores {
		if v >= 0 {
			p[i] = &v
		}
	}
	return p
}

// CA-36.1 · RN2 a RN5 · el burndown del sprint activo por defecto, el de otro sprint con ?sprint= y uno que no existe
// se ignora
func TestDatosDeGraficos_Burndown(t *testing.T) {
	e := nuevoEscenarioGraficos(t)
	activo := &service.GraficoBurndown{
		Sprint: "Sprint 2", Numero: 2, Dias: []string{"12/10", "13/10", "14/10", "15/10", "16/10", "17/10", "18/10"},
		Ideal: []float64{13, 10.8, 8.7, 6.5, 4.3, 2.2, 0}, Pendientes: pendientes(13, 5, 5, 3, -1, -1, -1),
	}

	verificarIgual(t, "burndown del activo", e.datos(t, 0).Burndown, activo)
	verificarIgual(t, "burndown de un sprint que no existe", e.datos(t, 9).Burndown, activo)
	verificarIgual(t, "burndown del Sprint 1", e.datos(t, 1).Burndown, &service.GraficoBurndown{
		Sprint: "Sprint 1", Numero: 1, Dias: []string{"05/10", "06/10", "07/10", "08/10", "09/10", "10/10", "11/10"},
		Ideal: []float64{13, 10.8, 8.7, 6.5, 4.3, 2.2, 0}, Pendientes: pendientes(13, 13, 13, 8, 8, 5, 5),
	})

	// sin sprint activo (después del último día), el del último sprint
	e.ahora = dia36(25)
	if b := e.datos(t, 0).Burndown; b == nil || b.Sprint != "Sprint 2" || b.Pendientes[6] == nil {
		t.Errorf("sin sprint activo se esperaba el burndown completo del Sprint 2 y se obtuvo %+v", b)
	}
}

// RN5 · un proyecto sin sprints tiene las series vacías (no nulas) y no tiene burndown
func TestDatosDeGraficos_SinSprints(t *testing.T) {
	e := nuevoEscenarioGraficos(t)
	p, err := e.proyectos.Guardar(context.Background(), project.Proyecto{Nombre: "Vacío", FechaInicio: dia36(5), FechaFin: dia36(31)})
	if err != nil {
		t.Fatalf("guardar el proyecto: %v", err)
	}

	d, err := e.casos.Datos(context.Background(), p.ID, 0)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if d.Burndown != nil || d.StoryPoints.Sprints == nil || len(d.StoryPoints.Sprints) != 0 || d.Velocidad.Velocidad == nil || d.Horas.Reales == nil {
		t.Errorf("se esperaban series vacías y sin burndown, y se obtuvo %+v", d)
	}
	verificarIgual(t, "defectos", d.Defectos.Detectados, []int{0, 0, 0, 0})
}

func TestDatosDeGraficos_ProyectoInexistente(t *testing.T) {
	if _, err := nuevoEscenarioGraficos(t).casos.Datos(context.Background(), 99, 0); !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("error = %v, se esperaba project.ErrNoEncontrado", err)
	}
}
