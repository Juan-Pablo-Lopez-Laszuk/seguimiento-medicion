package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// Los tests de HU-37 usan el ejemplo de la spec: el Sprint 1 (12/10 al 18/10) tiene HU-1 Login (5 SP, Hecha, 6 h
// estimadas y 8 h reales), HU-2 Registro (3 SP, En progreso, 4 h y 3 h) y HU-3 Perfil (5 SP, sin horas). El Sprint 2
// (19/10 al 25/10) no tiene historias.
type escenarioReporte struct {
	proyectos  *memory.Proyectos
	sprints    *memory.Sprints
	historias  *memory.Historias
	tareas     *memory.Tareas
	registros  *memory.Registros
	cierres    *memory.Cierres
	proyectoID int64
	sprintIDs  map[int]int64 // ID de cada sprint, por número
}

func nuevoEscenarioReporte(t *testing.T) *escenarioReporte {
	t.Helper()
	ctx := context.Background()
	e := &escenarioReporte{
		proyectos: memory.NuevoProyectos(), sprints: memory.NuevoSprints(), historias: memory.NuevoHistorias(),
		tareas: memory.NuevoTareas(), registros: memory.NuevoRegistros(), sprintIDs: map[int]int64{},
	}
	e.cierres = memory.NuevoCierres(e.sprints, e.historias)
	p, err := e.proyectos.Guardar(ctx, project.Proyecto{Nombre: "Software Metrics", FechaInicio: fechaISO(t, "2026-10-05"), FechaFin: fechaISO(t, "2026-11-30")})
	if err != nil {
		t.Fatalf("no se pudo guardar el proyecto: %v", err)
	}
	e.proyectoID = p.ID
	for i, inicio := range []string{"2026-10-12", "2026-10-19"} {
		numero, desde := i+1, fechaISO(t, inicio)
		s, err := e.sprints.Guardar(ctx, sprint.Sprint{ProyectoID: p.ID, Numero: numero, Objetivo: "Objetivo", FechaInicio: desde, FechaFin: desde.AddDate(0, 0, 6), Estado: sprint.EstadoPlanificado})
		if err != nil {
			t.Fatalf("no se pudo guardar el Sprint %d: %v", numero, err)
		}
		e.sprintIDs[numero] = s.ID
	}
	e.historiaDelReporte(t, 1, "Login", puntos(5), backlog.EstadoHecho, 6, 8)
	e.historiaDelReporte(t, 2, "Registro", puntos(3), backlog.EstadoEnProgreso, 4, 3)
	e.historiaDelReporte(t, 3, "Perfil", puntos(5), backlog.EstadoEnSprint, 0, 0)
	return e
}

// historiaDelReporte guarda la HU-numero en el Sprint 1, con una tarea de esas horas estimadas y un registro de esas
// horas reales (ninguno si son 0).
func (e *escenarioReporte) historiaDelReporte(t *testing.T, numero int, titulo string, sp *int, estado backlog.Estado, estimadas, reales float64) {
	t.Helper()
	ctx := context.Background()
	sprintID := e.sprintIDs[1]
	h, err := e.historias.Guardar(ctx, backlog.Historia{ProyectoID: e.proyectoID, Numero: numero, Titulo: titulo, Prioridad: backlog.PrioridadAlta, Estado: estado, StoryPoints: sp, SprintID: &sprintID})
	if err != nil {
		t.Fatalf("no se pudo guardar la HU-%d: %v", numero, err)
	}
	if estimadas > 0 {
		if _, err := e.tareas.Guardar(ctx, effort.Tarea{HistoriaID: h.ID, Titulo: "Desarrollo", HorasEstimadas: estimadas}); err != nil {
			t.Fatalf("no se pudo guardar la tarea de la HU-%d: %v", numero, err)
		}
	}
	if reales > 0 {
		if _, err := e.registros.Guardar(ctx, effort.Registro{HistoriaID: h.ID, Fecha: fechaISO(t, "2026-10-15"), Actividad: "Desarrollo", Horas: reales}); err != nil {
			t.Fatalf("no se pudo guardar el registro de la HU-%d: %v", numero, err)
		}
	}
}

// cerrar cierra el Sprint N con el caso de uso de HU-14, el día de su fecha de fin.
func (e *escenarioReporte) cerrar(t *testing.T, numero int, fin string) {
	t.Helper()
	ahora := fechaISO(t, fin).Add(18 * time.Hour) // 15:00 de Argentina
	cierres := service.NuevoCierres(e.proyectos, e.sprints, e.historias, e.cierres, func() time.Time { return ahora })
	if _, err := cierres.Cerrar(context.Background(), e.proyectoID, numero); err != nil {
		t.Fatalf("no se pudo cerrar el Sprint %d: %v", numero, err)
	}
}

// moverAlSprint2 re-estima la HU-numero y la pone en el Sprint 2.
func (e *escenarioReporte) moverAlSprint2(t *testing.T, numero, sp int) {
	t.Helper()
	ctx := context.Background()
	h, err := e.historias.BuscarPorNumero(ctx, e.proyectoID, numero)
	if err != nil {
		t.Fatalf("no está la HU-%d: %v", numero, err)
	}
	sprintID := e.sprintIDs[2]
	if err := e.historias.ActualizarStoryPoints(ctx, h.ID, sp); err != nil {
		t.Fatalf("no se pudo re-estimar la HU-%d: %v", numero, err)
	}
	if err := e.historias.ActualizarSprint(ctx, h.ID, backlog.EstadoEnSprint, &sprintID); err != nil {
		t.Fatalf("no se pudo mover la HU-%d al Sprint 2: %v", numero, err)
	}
}

// reportes arma el caso de uso con los repositorios del escenario y el de historias que se le pasa.
func (e *escenarioReporte) reportes(historias service.ListadorHistorias) *service.Reportes {
	cierres := service.NuevoCierres(e.proyectos, e.sprints, e.historias, e.cierres, time.Now)
	horas := service.NuevoHoras(e.proyectos, e.sprints, e.historias, e.tareas, e.registros)
	return service.NuevoReportes(e.proyectos, e.sprints, historias, cierres, horas)
}

func (e *escenarioReporte) reporte(t *testing.T, numero int) service.ReporteSprint {
	t.Helper()
	r, err := e.reportes(e.historias).Sprint(context.Background(), e.proyectoID, numero)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	return r
}

// metricasDelReporte son los textos de las tarjetas, en el orden de la sección 3 de la spec.
func metricasDelReporte(r service.ReporteSprint) []string {
	return []string{r.SPTexto(), r.CumplimientoTexto(), r.CompletadasTexto(), r.VelocidadTexto(), r.HorasTexto(), r.DesviacionTexto()}
}

func verificarMetricas(t *testing.T, r service.ReporteSprint, quiero ...string) {
	t.Helper()
	for i, obtenido := range metricasDelReporte(r) {
		if obtenido != quiero[i] {
			t.Errorf("métrica %d: se esperaba %q y se obtuvo %q", i+1, quiero[i], obtenido)
		}
	}
}

// CA-37.1 y CA-37.2 · el reporte de un sprint cerrado tiene las métricas de internal/metrics y las historias de la
// foto con sus horas (RN1, RN3, RN4 y RN5)
func TestReporteDeSprint_Cerrado(t *testing.T) {
	e := nuevoEscenarioReporte(t)
	e.cerrar(t, 1, "2026-10-18")

	r := e.reporte(t, 1)

	if r.Proyecto.Nombre != "Software Metrics" || r.Sprint.Numero != 1 || !r.Cerrado() {
		t.Errorf("se esperaba el Sprint 1 cerrado de Software Metrics y se obtuvo %+v", r.Sprint)
	}
	verificarMetricas(t, r, "5 de 13 SP", "38,5 %", "1 de 3 (33,3 %)", "5,0 SP por sprint (1 sprint cerrado)",
		"10 h estimadas · 11 h reales", "+1 h (+10,0 %)")
	quiero := []struct {
		id, titulo, resultado, estimadas, reales, desviacion string
		sp                                                   int
	}{
		{"HU-1", "Login", "Completada", "6 h", "8 h", "+2 h (+33,3 %)", 5},
		{"HU-2", "Registro", "No completada", "4 h", "3 h", "-1 h (-25,0 %)", 3},
		{"HU-3", "Perfil", "No completada", "0 h", "0 h", "0 h (no aplicable)", 5},
	}
	if len(r.Historias) != len(quiero) {
		t.Fatalf("se esperaban %d historias y se obtuvieron %d", len(quiero), len(r.Historias))
	}
	for i, q := range quiero {
		f := r.Historias[i]
		obtenido := []string{f.Identificador(), f.Titulo, f.Resultado(), f.EstimadasTexto(), f.RealesTexto(), f.DesviacionTexto()}
		esperado := []string{q.id, q.titulo, q.resultado, q.estimadas, q.reales, q.desviacion}
		for j := range esperado {
			if obtenido[j] != esperado[j] {
				t.Errorf("fila %d, dato %d: se esperaba %q y se obtuvo %q", i+1, j+1, esperado[j], obtenido[j])
			}
		}
		if *f.StoryPoints != q.sp {
			t.Errorf("fila %d: se esperaban %d SP y se obtuvieron %d", i+1, q.sp, *f.StoryPoints)
		}
	}
}

// RN1 · el reporte de un sprint cerrado no cambia aunque una historia de la foto se re-estime y entre a otro sprint
func TestReporteDeSprint_ConservaLaFoto(t *testing.T) {
	e := nuevoEscenarioReporte(t)
	e.cerrar(t, 1, "2026-10-18")
	e.moverAlSprint2(t, 2, 8)

	r := e.reporte(t, 1)

	verificarMetricas(t, r, "5 de 13 SP", "38,5 %", "1 de 3 (33,3 %)", "5,0 SP por sprint (1 sprint cerrado)",
		"10 h estimadas · 11 h reales", "+1 h (+10,0 %)")
	if f := r.Historias[1]; f.Identificador() != "HU-2" || *f.StoryPoints != 3 || f.Completada {
		t.Errorf("se esperaba HU-2 con 3 SP y no completada, y se obtuvo %+v", f.FilaFoto)
	}
}

// RN2 y RN4 · un sprint abierto usa sus historias de hoy, avisa las que no tienen estimación y la velocidad cuenta los
// sprints cerrados hasta ese
func TestReporteDeSprint_Abierto(t *testing.T) {
	e := nuevoEscenarioReporte(t)
	e.cerrar(t, 1, "2026-10-18")
	e.moverAlSprint2(t, 2, 8)
	sprintID := e.sprintIDs[2]
	if _, err := e.historias.Guardar(context.Background(), backlog.Historia{ProyectoID: e.proyectoID, Numero: 4, Titulo: "Ayuda", Prioridad: backlog.PrioridadBaja, Estado: backlog.EstadoEnSprint, SprintID: &sprintID}); err != nil {
		t.Fatalf("no se pudo guardar la HU-4: %v", err)
	}

	r := e.reporte(t, 2)

	if r.Cerrado() {
		t.Error("el Sprint 2 no está cerrado")
	}
	verificarMetricas(t, r, "0 de 8 SP", "0,0 %", "0 de 2 (0,0 %)", "5,0 SP por sprint (1 sprint cerrado)",
		"4 h estimadas · 3 h reales", "-1 h (-25,0 %)")
	if r.Advertencia() != "hay 1 historia sin estimar" {
		t.Errorf("se esperaba la advertencia de HU-30 y se obtuvo %q", r.Advertencia())
	}
	if len(r.Historias) != 2 || r.Historias[0].Identificador() != "HU-2" || *r.Historias[0].StoryPoints != 8 || r.Historias[1].StoryPoints != nil {
		t.Errorf("se esperaban HU-2 con 8 SP y HU-4 sin estimar, y se obtuvo %+v", r.Historias)
	}
}

// RN4 · la velocidad de un sprint viejo no cuenta los sprints que se cerraron después; sin cerrados, avisa
func TestReporteDeSprint_VelocidadHastaEseSprint(t *testing.T) {
	e := nuevoEscenarioReporte(t)
	if v := e.reporte(t, 1).VelocidadTexto(); v != "Aún no hay sprints cerrados" {
		t.Errorf("sin sprints cerrados se esperaba el aviso de HU-31 y se obtuvo %q", v)
	}
	e.cerrar(t, 1, "2026-10-18")
	e.moverAlSprint2(t, 2, 8)
	h, _ := e.historias.BuscarPorNumero(context.Background(), e.proyectoID, 2)
	if err := e.historias.ActualizarEstado(context.Background(), h.ID, backlog.EstadoHecho); err != nil {
		t.Fatalf("no se pudo completar la HU-2: %v", err)
	}
	e.cerrar(t, 2, "2026-10-25")

	if v := e.reporte(t, 1).VelocidadTexto(); v != "5,0 SP por sprint (1 sprint cerrado)" {
		t.Errorf("el Sprint 1 solo cuenta su cierre y se obtuvo %q", v)
	}
	if v := e.reporte(t, 2).VelocidadTexto(); v != "6,5 SP por sprint (2 sprints cerrados)" {
		t.Errorf("el Sprint 2 promedia los dos cierres y se obtuvo %q", v)
	}
}

// RN3 y RN5 · un sprint sin historias no divide por cero
func TestReporteDeSprint_SinHistorias(t *testing.T) {
	e := nuevoEscenarioReporte(t)

	r := e.reporte(t, 2)

	verificarMetricas(t, r, "0 de 0 SP", "—", "0 de 0 (0,0 %)", "Aún no hay sprints cerrados",
		"0 h estimadas · 0 h reales", "0 h (no aplicable)")
	if len(r.Historias) != 0 {
		t.Errorf("el Sprint 2 no tiene historias y se obtuvieron %d", len(r.Historias))
	}
}

// CA-37.3 y sección 7 · un proyecto o un sprint que no existen, o una falla de la base, son errores
func TestReporteDeSprint_Errores(t *testing.T) {
	e := nuevoEscenarioReporte(t)
	e.cerrar(t, 1, "2026-10-18")
	casos := []struct {
		nombre     string
		historias  service.ListadorHistorias
		proyectoID int64
		numero     int
		quiero     error
	}{
		{"el proyecto no existe", e.historias, 99, 1, project.ErrNoEncontrado},
		{"el sprint no existe", e.historias, e.proyectoID, 9, sprint.ErrNoEncontrado},
		{"falla la base en un sprint cerrado", historiasQueFallan{errListar: errRepo}, e.proyectoID, 1, errRepo},
		{"falla la base en un sprint abierto", historiasQueFallan{errListar: errRepo}, e.proyectoID, 2, errRepo},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := e.reportes(c.historias).Sprint(context.Background(), c.proyectoID, c.numero)
			if !errors.Is(err, c.quiero) {
				t.Errorf("se esperaba %v y se obtuvo %v", c.quiero, err)
			}
		})
	}
}
