package service_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/defect"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// nuevoEscenarioReporteDeProyecto arma el ejemplo de la sección 6 de la spec de HU-38 sobre el de HU-37: el Sprint 1
// se cierra con HU-1 completada (5 de 13 SP); HU-2 y HU-3 pasan al Sprint 2, que se cierra con HU-2 completada (3 de
// 8 SP); HU-3 queda en el Sprint 3, abierto. HU-3 tiene 2 h estimadas. Defectos: #1 detectado y resuelto en el Sprint
// 1, #2 detectado en el Sprint 1 y resuelto en el 2, y #3 detectado en el Sprint 2 y abierto.
func nuevoEscenarioReporteDeProyecto(t *testing.T) *escenarioReporte {
	t.Helper()
	ctx := context.Background()
	e := nuevoEscenarioReporte(t)
	desde := fechaISO(t, "2026-10-26")
	s, err := e.sprints.Guardar(ctx, sprint.Sprint{ProyectoID: e.proyectoID, Numero: 3, Objetivo: "Objetivo", FechaInicio: desde, FechaFin: desde.AddDate(0, 0, 6), Estado: sprint.EstadoPlanificado})
	if err != nil {
		t.Fatalf("no se pudo guardar el Sprint 3: %v", err)
	}
	e.sprintIDs[3] = s.ID
	perfil, err := e.historias.BuscarPorNumero(ctx, e.proyectoID, 3)
	if err != nil {
		t.Fatalf("no está la HU-3: %v", err)
	}
	if _, err := e.tareas.Guardar(ctx, effort.Tarea{HistoriaID: perfil.ID, Titulo: "Desarrollo", HorasEstimadas: 2}); err != nil {
		t.Fatalf("no se pudo guardar la tarea de la HU-3: %v", err)
	}
	e.defecto(t, "El login no valida el correo", defect.SeveridadCritica, 1, 1, 1)
	e.defecto(t, "El registro tarda mucho", defect.SeveridadMedia, 2, 1, 2)
	e.defecto(t, "El perfil no carga la foto", defect.SeveridadBaja, 3, 2, 0)

	e.cerrar(t, 1, "2026-10-18")
	e.ponerEnSprint(t, 2, 2, backlog.EstadoHecho)
	e.ponerEnSprint(t, 3, 2, backlog.EstadoEnSprint)
	e.cerrar(t, 2, "2026-10-25")
	e.ponerEnSprint(t, 3, 3, backlog.EstadoEnSprint)
	return e
}

// ponerEnSprint deja la HU-numero en el Sprint sprintNumero con ese estado.
func (e *escenarioReporte) ponerEnSprint(t *testing.T, numero, sprintNumero int, estado backlog.Estado) {
	t.Helper()
	ctx := context.Background()
	h, err := e.historias.BuscarPorNumero(ctx, e.proyectoID, numero)
	if err != nil {
		t.Fatalf("no está la HU-%d: %v", numero, err)
	}
	sprintID := e.sprintIDs[sprintNumero]
	if err := e.historias.ActualizarSprint(ctx, h.ID, estado, &sprintID); err != nil {
		t.Fatalf("no se pudo poner la HU-%d en el Sprint %d: %v", numero, sprintNumero, err)
	}
}

func (e *escenarioReporte) reporteDeProyecto(t *testing.T, proyectoID int64) service.ReporteProyecto {
	t.Helper()
	r, err := e.reportes(e.historias).Proyecto(context.Background(), proyectoID)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	return r
}

// resumenDelReporte son los textos de las tarjetas del resumen, en el orden de la sección 3 de la spec.
func resumenDelReporte(r service.ReporteProyecto) []string {
	return []string{r.SprintsTexto(), r.SPTexto(), r.VelocidadTexto(), r.HorasTexto(), r.DesviacionTexto(), r.DefectosTexto()}
}

func verificarResumen(t *testing.T, r service.ReporteProyecto, quiero ...string) {
	t.Helper()
	for i, obtenido := range resumenDelReporte(r) {
		if obtenido != quiero[i] {
			t.Errorf("tarjeta %d: se esperaba %q y se obtuvo %q", i+1, quiero[i], obtenido)
		}
	}
}

// CA-38.1 y RN2 · el resumen tiene los totales del proyecto, como el panel: HU-2 y HU-3 aparecen en dos sprints pero
// sus horas se cuentan una vez (12 h y 11 h, no 20 h y 14 h)
func TestReporteDeProyecto_Resumen(t *testing.T) {
	e := nuevoEscenarioReporteDeProyecto(t)

	r := e.reporteDeProyecto(t, e.proyectoID)

	if r.Resumen.Proyecto.Nombre != "Software Metrics" {
		t.Errorf("se esperaba el proyecto Software Metrics y se obtuvo %q", r.Resumen.Proyecto.Nombre)
	}
	verificarResumen(t, r, "2 cerrados de 3", "8 de 13 SP (61,5 %)", "4,0 SP por sprint (2 sprints cerrados)",
		"12 h estimadas · 11 h reales", "-1 h (-8,3 %)", "3 detectados · 2 resueltos · 1 abierto")
}

// CA-38.1 y RN1 · un bloque por sprint, ordenados por número, igual al reporte de ese sprint (HU-37)
func TestReporteDeProyecto_UnBloquePorSprint(t *testing.T) {
	e := nuevoEscenarioReporteDeProyecto(t)

	r := e.reporteDeProyecto(t, e.proyectoID)

	if len(r.Sprints) != 3 {
		t.Fatalf("se esperaban 3 bloques y se obtuvieron %d", len(r.Sprints))
	}
	for i, bloque := range r.Sprints {
		numero := i + 1
		if quiero := e.reporte(t, numero); !reflect.DeepEqual(bloque, quiero) {
			t.Errorf("el bloque %d no es igual al reporte del Sprint %d:\n%+v\n%+v", numero, numero, bloque, quiero)
		}
	}
	if r.Sprints[2].Cerrado() {
		t.Error("el Sprint 3 está abierto y su bloque dice que está cerrado")
	}
}

// CA-38.2 y RN3 · la evolución de la velocidad tiene una fila por sprint cerrado, con los SP del cierre, el
// cumplimiento y la velocidad hasta ese sprint
func TestReporteDeProyecto_EvolucionDeLaVelocidad(t *testing.T) {
	e := nuevoEscenarioReporteDeProyecto(t)

	evolucion := e.reporteDeProyecto(t, e.proyectoID).Evolucion()

	quiero := []struct {
		numero                    int
		planificados, completados float64
		cumplimiento, velocidad   string
	}{
		{1, 13, 5, "38,5 %", "5,0 SP por sprint"},
		{2, 8, 3, "37,5 %", "4,0 SP por sprint"},
	}
	if len(evolucion) != len(quiero) {
		t.Fatalf("se esperaban %d filas y se obtuvieron %d", len(quiero), len(evolucion))
	}
	for i, q := range quiero {
		f := evolucion[i]
		if f.Sprint.Numero != q.numero || f.SP.Planificados != q.planificados || f.SP.Completados != q.completados ||
			f.CumplimientoTexto() != q.cumplimiento || f.VelocidadPorSprint() != q.velocidad {
			t.Errorf("fila %d: se esperaba %+v y se obtuvo Sprint %d, %g de %g SP, %s, %s", i+1, q, f.Sprint.Numero,
				f.SP.Completados, f.SP.Planificados, f.CumplimientoTexto(), f.VelocidadPorSprint())
		}
	}
}

// RN1 y RN3 · re-estimar una historia después del cierre no cambia el bloque ni la fila del sprint cerrado
func TestReporteDeProyecto_ConservaLasFotos(t *testing.T) {
	e := nuevoEscenarioReporteDeProyecto(t)
	perfil, err := e.historias.BuscarPorNumero(context.Background(), e.proyectoID, 3)
	if err != nil {
		t.Fatalf("no está la HU-3: %v", err)
	}
	if err := e.historias.ActualizarStoryPoints(context.Background(), perfil.ID, 8); err != nil {
		t.Fatalf("no se pudo re-estimar la HU-3: %v", err)
	}

	r := e.reporteDeProyecto(t, e.proyectoID)

	if f := r.Evolucion()[1]; f.SP.Planificados != 8 || f.SP.Completados != 3 {
		t.Errorf("la fila del Sprint 2 debía seguir con 3 de 8 SP y tiene %g de %g", f.SP.Completados, f.SP.Planificados)
	}
	for _, h := range r.Sprints[1].Historias {
		if h.Numero == 3 && *h.StoryPoints != 5 {
			t.Errorf("en el bloque del Sprint 2, HU-3 debía seguir con 5 SP y tiene %d", *h.StoryPoints)
		}
	}
	if texto := r.SPTexto(); texto != "8 de 16 SP (50,0 %)" {
		t.Errorf("el resumen usa los SP de hoy: se esperaba %q y se obtuvo %q", "8 de 16 SP (50,0 %)", texto)
	}
}

// RN5 · un proyecto sin sprints no da error: el resumen dice "sin datos" y no hay bloques ni evolución
func TestReporteDeProyecto_SinSprints(t *testing.T) {
	e := nuevoEscenarioReporte(t)
	nuevo, err := e.proyectos.Guardar(context.Background(), project.Proyecto{Nombre: "Proyecto nuevo", FechaInicio: fechaISO(t, "2026-10-05"), FechaFin: fechaISO(t, "2026-11-30")})
	if err != nil {
		t.Fatalf("no se pudo guardar el proyecto: %v", err)
	}

	r := e.reporteDeProyecto(t, nuevo.ID)

	verificarResumen(t, r, "sin datos", "sin datos", "sin datos", "sin datos", "sin datos", "sin datos")
	if len(r.Sprints) != 0 || len(r.Evolucion()) != 0 {
		t.Errorf("no se esperaban bloques ni evolución y se obtuvieron %d y %d", len(r.Sprints), len(r.Evolucion()))
	}
}

// Sección 6 · con sprints pero ninguno cerrado, hay bloques (parciales) y la evolución está vacía
func TestReporteDeProyecto_SinSprintsCerrados(t *testing.T) {
	e := nuevoEscenarioReporte(t)

	r := e.reporteDeProyecto(t, e.proyectoID)

	if len(r.Sprints) != 2 || r.Sprints[0].Cerrado() || len(r.Evolucion()) != 0 {
		t.Errorf("se esperaban 2 bloques abiertos y ninguna fila en la evolución; hay %d bloques y %d filas", len(r.Sprints), len(r.Evolucion()))
	}
	if r.SprintsTexto() != "0 cerrados de 2" || r.VelocidadTexto() != "sin datos" {
		t.Errorf("se esperaba %q y %q y se obtuvo %q y %q", "0 cerrados de 2", "sin datos", r.SprintsTexto(), r.VelocidadTexto())
	}
}

// Sección 7 · un proyecto que no existe, o una falla de la base, son errores
func TestReporteDeProyecto_Errores(t *testing.T) {
	e := nuevoEscenarioReporteDeProyecto(t)
	casos := []struct {
		nombre     string
		reportes   *service.Reportes
		proyectoID int64
		quiero     error
	}{
		{"el proyecto no existe", e.reportes(e.historias), 99, project.ErrNoEncontrado},
		{"fallan las historias", e.reportes(historiasQueFallan{errListar: errRepo}), e.proyectoID, errRepo},
		{"fallan los defectos", e.reportesCon(e.historias, defectosQueFallan{}), e.proyectoID, errRepo},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if _, err := c.reportes.Proyecto(context.Background(), c.proyectoID); !errors.Is(err, c.quiero) {
				t.Errorf("se esperaba %v y se obtuvo %v", c.quiero, err)
			}
		})
	}
}
