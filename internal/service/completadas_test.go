package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
)

// enSprintCon guarda una historia ya asignada al sprint, en ese estado y con esos criterios (los primeros cumplidos),
// directo en los repositorios, y devuelve su número.
func (e *escenarioAsignacion) enSprintCon(t *testing.T, numeroSprint, sp int, estado backlog.Estado, total, cumplidos int) int {
	t.Helper()
	numero := e.historia(t, "En el sprint", puntos(sp), 0, backlog.EstadoPendiente)
	h := e.guardada(t, numero)
	conCriterios(t, e.criterios, h.ID, total, cumplidos)
	id := e.ids[numeroSprint]
	if err := e.historias.ActualizarSprint(context.Background(), h.ID, estado, &id); err != nil {
		t.Fatalf("no se pudo asignar la historia: %v", err)
	}
	return numero
}

// conReloj vuelve a armar el caso de uso con "hoy" en ese momento.
func (e *escenarioAsignacion) conReloj(ahora time.Time) {
	e.ahora = func() time.Time { return ahora }
	e.casos = e.armar(e.historias)
}

func fechaDeCompletado(t *testing.T, h backlog.Historia) string {
	t.Helper()
	if h.CompletadaEn == nil {
		return "sin fecha"
	}
	return h.CompletadaEn.Format(time.DateOnly)
}

// CA-13.1 y CA-13.2 · la historia del sprint activo queda guardada Hecha con la fecha de hoy y sigue en su sprint
func TestCompletar_HistoriaDelSprintActivo(t *testing.T) {
	e := nuevoEscenarioAsignacion(t, true)
	numero := e.enSprintCon(t, 2, 5, backlog.EstadoEnProgreso, 2, 2)

	h, err := e.casos.Completar(context.Background(), e.proyectoID, 2, numero)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	guardada := e.guardada(t, numero)
	if h.Estado != backlog.EstadoHecho || guardada.Estado != backlog.EstadoHecho || fechaDeCompletado(t, guardada) != "2026-10-15" {
		t.Errorf("se esperaba Hecha el 2026-10-15 y quedó %q el %s", guardada.Estado, fechaDeCompletado(t, guardada))
	}
	if guardada.SprintID == nil || *guardada.SprintID != e.ids[2] {
		t.Errorf("la historia Hecha tiene que seguir en el Sprint 2 y quedó en %v", guardada.SprintID)
	}
}

// CA-13.2 · la fecha es la de Argentina: a las 22:00 del 18/10 en Argentina, en UTC ya es el 19/10
func TestCompletar_LaFechaEsLaDeArgentina(t *testing.T) {
	e := nuevoEscenarioAsignacion(t, true)
	numero := e.enSprintCon(t, 2, 5, backlog.EstadoEnProgreso, 1, 1)
	e.conReloj(time.Date(2026, 10, 19, 1, 0, 0, 0, time.UTC))

	if _, err := e.casos.Completar(context.Background(), e.proyectoID, 2, numero); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if f := fechaDeCompletado(t, e.guardada(t, numero)); f != "2026-10-18" {
		t.Errorf("fecha de completado = %s, se esperaba 2026-10-18", f)
	}
}

// CA-13.1 y HU-04 RN2 · un sprint vencido sin cerrar sigue activo: se completa con la fecha de hoy
func TestCompletar_SprintVencidoSinCerrar(t *testing.T) {
	e := nuevoEscenarioAsignacion(t, true)
	numero := e.enSprintCon(t, 2, 5, backlog.EstadoEnProgreso, 1, 1)
	e.conReloj(time.Date(2026, 10, 20, 15, 0, 0, 0, time.UTC))

	if _, err := e.casos.Completar(context.Background(), e.proyectoID, 2, numero); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if f := fechaDeCompletado(t, e.guardada(t, numero)); f != "2026-10-20" {
		t.Errorf("fecha de completado = %s, se esperaba 2026-10-20", f)
	}
}

// Sección 7 · lo que no se completa no cambia nada en el repositorio
func TestCompletar_Rechazos_NoGuardanNada(t *testing.T) {
	casos := []struct {
		nombre           string
		sprint           int
		estado           backlog.Estado
		total, cumplidos int
		err              error
	}{
		{"faltan criterios", 2, backlog.EstadoEnProgreso, 2, 1, backlog.ErrCriteriosSinCumplir},
		{"todavía no empezó", 2, backlog.EstadoEnSprint, 1, 1, backlog.ErrTransicionInvalida},
		{"no es del sprint activo", 3, backlog.EstadoEnProgreso, 1, 1, backlog.ErrNoEsElSprintActivo},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			e := nuevoEscenarioAsignacion(t, true)
			numero := e.enSprintCon(t, c.sprint, 5, c.estado, c.total, c.cumplidos)

			_, err := e.casos.Completar(context.Background(), e.proyectoID, c.sprint, numero)

			if !errors.Is(err, c.err) {
				t.Fatalf("error = %v, se esperaba %v", err, c.err)
			}
			if h := e.guardada(t, numero); h.Estado != c.estado || h.CompletadaEn != nil {
				t.Errorf("la historia cambió: quedó %q con fecha %s", h.Estado, fechaDeCompletado(t, h))
			}
		})
	}
}

// Sección 7 · una historia que no está en ese sprint no se completa ni se empieza desde él
func TestCompletarYEmpezar_HistoriaDeOtroSprint(t *testing.T) {
	e := nuevoEscenarioAsignacion(t, true)
	numero := e.enSprintCon(t, 3, 5, backlog.EstadoEnSprint, 1, 1)
	ctx := context.Background()

	if _, err := e.casos.Completar(ctx, e.proyectoID, 2, numero); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("completar: error = %v, se esperaba backlog.ErrHistoriaNoEncontrada", err)
	}
	if _, err := e.casos.Empezar(ctx, e.proyectoID, 2, numero); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("empezar: error = %v, se esperaba backlog.ErrHistoriaNoEncontrada", err)
	}
}

// RN2 · empezar pasa una historia En sprint a En progreso (la transición de HU-08) y sigue en su sprint
func TestEmpezar_HistoriaEnSprint(t *testing.T) {
	e := nuevoEscenarioAsignacion(t, true)
	numero := e.enSprintCon(t, 2, 5, backlog.EstadoEnSprint, 1, 0)

	if _, err := e.casos.Empezar(context.Background(), e.proyectoID, 2, numero); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if h := e.guardada(t, numero); h.Estado != backlog.EstadoEnProgreso || h.SprintID == nil || *h.SprintID != e.ids[2] {
		t.Errorf("se esperaba En progreso en el Sprint 2 y quedó %q en %v", h.Estado, h.SprintID)
	}
	if _, err := e.casos.Empezar(context.Background(), e.proyectoID, 2, numero); !errors.Is(err, backlog.ErrYaEstaEnEseEstado) {
		t.Errorf("empezarla de nuevo: error = %v, se esperaba backlog.ErrYaEstaEnEseEstado", err)
	}
}

// Sección 3 · la página dice si el sprint es el activo, el avance en SP y el progreso de criterios de cada historia
func TestVerSprint_AvanceYProgresoDeCriterios(t *testing.T) {
	e := nuevoEscenarioAsignacion(t, true)
	hecha := e.enSprintCon(t, 2, 5, backlog.EstadoEnProgreso, 2, 2)
	aMedias := e.enSprintCon(t, 2, 8, backlog.EstadoEnProgreso, 3, 1)
	if _, err := e.casos.Completar(context.Background(), e.proyectoID, 2, hecha); err != nil {
		t.Fatalf("no se pudo completar: %v", err)
	}

	v, err := e.casos.Ver(context.Background(), e.proyectoID, 2)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if !v.Activo || v.Completados != 5 || v.Comprometidos != 13 {
		t.Errorf("se esperaba el sprint activo con 5 de 13 SP y se obtuvo activo=%v, %d de %d", v.Activo, v.Completados, v.Comprometidos)
	}
	for _, a := range v.Asignadas {
		if a.Numero == aMedias && (a.Progreso.Cumplidos != 1 || a.Progreso.Total != 3) {
			t.Errorf("HU-%d: progreso = %+v, se esperaba 1 de 3", a.Numero, a.Progreso)
		}
	}
	futuro, err := e.casos.Ver(context.Background(), e.proyectoID, 3)
	if err != nil || futuro.Activo {
		t.Errorf("el Sprint 3 todavía no empezó y no es el activo: se obtuvo %v, %v", futuro.Activo, err)
	}
}
