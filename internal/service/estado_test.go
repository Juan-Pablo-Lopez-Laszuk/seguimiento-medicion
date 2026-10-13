package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// escenarioEstado tiene los repositorios en memoria con el proyecto de ejemplo ya creado
// (del 05/10/2026 al 01/11/2026). Los tests le agregan sprints e historias y después consultan.
type escenarioEstado struct {
	t          *testing.T
	proyectos  *memory.Proyectos
	sprints    *memory.Sprints
	historias  *memory.Historias
	proyectoID int64
}

func nuevoEscenarioEstado(t *testing.T) *escenarioEstado {
	t.Helper()
	e := &escenarioEstado{t: t, proyectos: memory.NuevoProyectos(), sprints: memory.NuevoSprints(), historias: memory.NuevoHistorias()}
	p, err := service.NuevoProyectos(e.proyectos, relojFijo).Crear(context.Background(), datosValidos())
	if err != nil {
		t.Fatalf("no se pudo crear el proyecto de ejemplo: %v", err)
	}
	e.proyectoID = p.ID
	return e
}

// fechaISO convierte "AAAA-MM-DD" en fecha.
func fechaISO(t *testing.T, texto string) time.Time {
	t.Helper()
	f, err := time.Parse(time.DateOnly, texto)
	if err != nil {
		t.Fatalf("fecha de prueba inválida %q: %v", texto, err)
	}
	return f
}

func (e *escenarioEstado) sprint(numero int, inicio, fin string, estado sprint.Estado) {
	e.t.Helper()
	_, err := e.sprints.Guardar(context.Background(), sprint.Sprint{
		ProyectoID: e.proyectoID, Numero: numero, Objetivo: "Goal",
		FechaInicio: fechaISO(e.t, inicio), FechaFin: fechaISO(e.t, fin), Estado: estado,
	})
	if err != nil {
		e.t.Fatalf("no se pudo guardar el sprint: %v", err)
	}
}

// historia guarda una historia con esos SP (nil = sin estimar) y ese estado.
func (e *escenarioEstado) historia(sp *int, estado backlog.Estado) {
	e.t.Helper()
	_, err := e.historias.Guardar(context.Background(), backlog.Historia{
		ProyectoID: e.proyectoID, Titulo: "Historia", Prioridad: backlog.PrioridadAlta, Estado: estado, StoryPoints: sp,
	})
	if err != nil {
		e.t.Fatalf("no se pudo guardar la historia: %v", err)
	}
}

// caso arma el caso de uso con el reloj en ese instante.
func (e *escenarioEstado) caso(ahora time.Time) *service.EstadoProyectos {
	return service.NuevoEstadoProyectos(e.proyectos, e.sprints, e.historias, func() time.Time { return ahora })
}

// consultar devuelve el estado del proyecto con "hoy" a las 15:00 de Argentina de ese día.
func (e *escenarioEstado) consultar(hoy string) service.EstadoProyecto {
	e.t.Helper()
	ahora := fechaISO(e.t, hoy).Add(18 * time.Hour) // 18:00 UTC = 15:00 de Argentina
	estado, err := e.caso(ahora).Consultar(context.Background(), e.proyectoID)
	if err != nil {
		e.t.Fatalf("no se esperaba error: %v", err)
	}
	return estado
}

func puntos(n int) *int { return &n }

// CA-04.3 · un proyecto sin sprints está Planificado, aunque ya haya pasado su fecha de fin
func TestConsultarEstado_SinSprints_EstaPlanificado(t *testing.T) {
	e := nuevoEscenarioEstado(t)
	for _, hoy := range []string{"2026-10-20", "2026-11-10"} {
		estado := e.consultar(hoy)
		if estado.Estado != project.EstadoPlanificado || estado.Activo != nil || estado.Proximo != nil {
			t.Errorf("hoy %s: se esperaba Planificado sin sprint activo ni próximo y se obtuvo %+v", hoy, estado)
		}
		if estado.Proyecto.Nombre != "Software Metrics" {
			t.Errorf("hoy %s: se esperaba el proyecto de ejemplo y se obtuvo %+v", hoy, estado.Proyecto)
		}
	}
}

// CA-04.1 · proyecto en curso con su sprint activo y los días que le quedan a los dos
func TestConsultarEstado_ConSprintActivo(t *testing.T) {
	e := nuevoEscenarioEstado(t)
	e.sprint(1, "2026-10-05", "2026-10-11", sprint.EstadoCerrado)
	e.sprint(2, "2026-10-12", "2026-10-18", sprint.EstadoPlanificado)

	estado := e.consultar("2026-10-13")

	if estado.Estado != project.EstadoEnCurso || estado.DiasRestantes != 19 {
		t.Errorf("se esperaba En curso con 19 días y se obtuvo %q con %d", estado.Estado, estado.DiasRestantes)
	}
	if estado.Activo == nil || estado.Activo.Numero != 2 || estado.Activo.DiasRestantes != 5 || estado.Activo.Terminado {
		t.Errorf("se esperaba el Sprint 2 activo con 5 días y se obtuvo %+v", estado.Activo)
	}
}

// CA-04.1 · un sprint vencido sin cerrar sigue activo, con 0 días y el aviso de que falta cerrarlo
func TestConsultarEstado_SprintVencidoSinCerrar(t *testing.T) {
	e := nuevoEscenarioEstado(t)
	e.sprint(1, "2026-10-05", "2026-10-11", sprint.EstadoPlanificado)
	e.sprint(2, "2026-10-12", "2026-10-18", sprint.EstadoPlanificado)

	estado := e.consultar("2026-10-13")

	if estado.Activo == nil || estado.Activo.Numero != 1 || estado.Activo.DiasRestantes != 0 || !estado.Activo.Terminado {
		t.Errorf("se esperaba el Sprint 1 activo, terminado y con 0 días, y se obtuvo %+v", estado.Activo)
	}
}

// Sección 3 · sin sprint activo se informa el próximo
func TestConsultarEstado_SinSprintActivo_InformaElProximo(t *testing.T) {
	e := nuevoEscenarioEstado(t)
	e.sprint(1, "2026-10-05", "2026-10-11", sprint.EstadoPlanificado)

	estado := e.consultar("2026-10-04")

	if estado.Estado != project.EstadoPlanificado || estado.Activo != nil {
		t.Errorf("se esperaba Planificado sin sprint activo y se obtuvo %+v", estado)
	}
	if estado.Proximo == nil || estado.Proximo.Numero != 1 {
		t.Errorf("se esperaba el Sprint 1 como próximo y se obtuvo %+v", estado.Proximo)
	}
}

// CA-04.2 · el estado depende de la fecha (último día en curso, el siguiente finalizado)
func TestConsultarEstado_FinDelProyecto(t *testing.T) {
	e := nuevoEscenarioEstado(t)
	e.sprint(1, "2026-10-05", "2026-10-11", sprint.EstadoCerrado)

	if estado := e.consultar("2026-11-01"); estado.Estado != project.EstadoEnCurso || estado.DiasRestantes != 0 {
		t.Errorf("01/11: se esperaba En curso con 0 días y se obtuvo %q con %d", estado.Estado, estado.DiasRestantes)
	}
	if estado := e.consultar("2026-11-02"); estado.Estado != project.EstadoFinalizado || estado.DiasRestantes != 0 {
		t.Errorf("02/11: se esperaba Finalizado con 0 días y se obtuvo %q con %d", estado.Estado, estado.DiasRestantes)
	}
}

// RN6 · a las 22:00 de Argentina del 01/11 (en UTC ya es 02/11) el proyecto sigue En curso
func TestConsultarEstado_HoyEsLaFechaDeArgentina(t *testing.T) {
	e := nuevoEscenarioEstado(t)
	e.sprint(1, "2026-10-05", "2026-10-11", sprint.EstadoCerrado)

	ahora := time.Date(2026, 11, 2, 1, 0, 0, 0, time.UTC)
	estado, err := e.caso(ahora).Consultar(context.Background(), e.proyectoID)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if estado.Estado != project.EstadoEnCurso {
		t.Errorf("se esperaba En curso y se obtuvo %q", estado.Estado)
	}
}

// RN4 · % de SP = SP Hechos / SP estimados; las historias sin estimar no suman y se cuentan aparte
func TestConsultarEstado_PorcentajeDeStoryPoints(t *testing.T) {
	e := nuevoEscenarioEstado(t)
	e.historia(puntos(3), backlog.EstadoHecho)
	e.historia(puntos(5), backlog.EstadoPendiente)
	e.historia(nil, backlog.EstadoPendiente)

	estado := e.consultar("2026-10-13")

	if got := estado.PorcentajeSP(); got != "37,5 %" {
		t.Errorf("se esperaba \"37,5 %%\" y se obtuvo %q", got)
	}
	if estado.StoryPoints.SinEstimar != 1 || estado.StoryPoints.Planificados != 8 || estado.StoryPoints.Completados != 3 {
		t.Errorf("se esperaban 8 SP planificados, 3 completados y 1 sin estimar, y se obtuvo %+v", estado.StoryPoints)
	}
}

// RN4 y RG-3 · sin SP estimados no se divide por cero: se muestra un texto en lugar de 0 %
func TestConsultarEstado_SinHistoriasEstimadas(t *testing.T) {
	e := nuevoEscenarioEstado(t)
	if got := e.consultar("2026-10-13").PorcentajeSP(); got != "Sin historias estimadas" {
		t.Errorf("sin historias se esperaba \"Sin historias estimadas\" y se obtuvo %q", got)
	}
	e.historia(nil, backlog.EstadoPendiente)
	if got := e.consultar("2026-10-13").PorcentajeSP(); got != "Sin historias estimadas" {
		t.Errorf("con historias sin estimar se esperaba \"Sin historias estimadas\" y se obtuvo %q", got)
	}
}

// Sección 7 · el proyecto no existe
func TestConsultarEstado_ProyectoInexistente(t *testing.T) {
	e := nuevoEscenarioEstado(t)
	_, err := e.caso(ahoraFijo).Consultar(context.Background(), 999)
	if !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("se esperaba project.ErrNoEncontrado y se obtuvo %v", err)
	}
}

// sprintsCaidos simula la base caída al leer los sprints.
type sprintsCaidos struct{}

var errSprintsCaidos = errors.New("base caída")

func (sprintsCaidos) ListarPorProyecto(context.Context, int64) ([]sprint.Sprint, error) {
	return nil, errSprintsCaidos
}

// Sección 7 · si falla la base, el error llega al handler para mostrar el mensaje general
func TestConsultarEstado_ErrorDeLaBase(t *testing.T) {
	e := nuevoEscenarioEstado(t)
	caso := service.NuevoEstadoProyectos(e.proyectos, sprintsCaidos{}, e.historias, relojFijo)

	if _, err := caso.Consultar(context.Background(), e.proyectoID); !errors.Is(err, errSprintsCaidos) {
		t.Errorf("Consultar: se esperaba el error de la base y se obtuvo %v", err)
	}
	if _, err := caso.Listar(context.Background()); !errors.Is(err, errSprintsCaidos) {
		t.Errorf("Listar: se esperaba el error de la base y se obtuvo %v", err)
	}
}

// RN1 · en la lista de proyectos también se muestra el estado calculado, no el guardado
func TestListarConEstado(t *testing.T) {
	e := nuevoEscenarioEstado(t)
	e.sprint(1, "2026-10-05", "2026-10-11", sprint.EstadoPlanificado)
	otro := datosValidos()
	otro.Nombre = "Otro proyecto"
	if _, err := service.NuevoProyectos(e.proyectos, relojFijo).Crear(context.Background(), otro); err != nil {
		t.Fatalf("no se pudo crear el segundo proyecto: %v", err)
	}

	lista, err := e.caso(fechaISO(t, "2026-10-13").Add(18*time.Hour)).Listar(context.Background())
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if len(lista) != 2 {
		t.Fatalf("se esperaban 2 proyectos y se obtuvieron %d", len(lista))
	}
	if lista[0].Proyecto.Nombre != "Software Metrics" || lista[0].Estado != project.EstadoEnCurso {
		t.Errorf("primero: se esperaba Software Metrics En curso y se obtuvo %+v", lista[0])
	}
	if lista[1].Proyecto.Nombre != "Otro proyecto" || lista[1].Estado != project.EstadoPlanificado {
		t.Errorf("segundo: se esperaba Otro proyecto Planificado (sin sprints) y se obtuvo %+v", lista[1])
	}
}
