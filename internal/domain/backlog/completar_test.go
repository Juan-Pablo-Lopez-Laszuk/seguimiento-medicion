package backlog_test

import (
	"errors"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
)

// quinceDeOctubre es "hoy" en los tests de HU-13, con hora: Completar guarda solo la fecha.
var quinceDeOctubre = time.Date(2026, 10, 15, 21, 30, 0, 0, time.UTC)

// enProgreso devuelve la historia de ejemplo En progreso en el Sprint 2.
func enProgreso() backlog.Historia {
	return enSprint(historiaLista(), sprint2, backlog.EstadoEnProgreso)
}

// CA-13.1 y CA-13.2 · una historia En progreso con todos sus criterios pasa a Hecho con la fecha de hoy, sin la hora;
// la original no cambia
func TestCompletar_HistoriaLista(t *testing.T) {
	original := enProgreso()

	hecha, err := original.Completar(true, progresoCompleto(), quinceDeOctubre)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if hecha.Estado != backlog.EstadoHecho {
		t.Errorf("estado = %q, se esperaba %q", hecha.Estado, backlog.EstadoHecho)
	}
	if hecha.CompletadaEn == nil || !hecha.CompletadaEn.Equal(time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("CompletadaEn = %v, se esperaba el 15/10/2026 sin hora", hecha.CompletadaEn)
	}
	if hecha.SprintID == nil || *hecha.SprintID != sprint2.ID {
		t.Errorf("la historia Hecha sigue en su sprint (RN6) y quedó en %v", hecha.SprintID)
	}
	if original.Estado != backlog.EstadoEnProgreso || original.CompletadaEn != nil {
		t.Errorf("la historia original cambió: %+v", original)
	}
}

// Sección 7 · cada motivo por el que no se completa, con el error de HU-08 (o el del sprint activo) y su mensaje
func TestCompletar_Errores(t *testing.T) {
	sinCumplir := enProgreso()
	sinCumplir.Numero = 2
	sinEmpezar := enSprint(historiaLista(), sprint2, backlog.EstadoEnSprint)
	sinEmpezar.Numero = 3
	hecha := enSprint(historiaLista(), sprint2, backlog.EstadoHecho)

	casos := []struct {
		nombre   string
		historia backlog.Historia
		activo   bool
		progreso backlog.Progreso
		err      error
		mensaje  string
	}{
		{"faltan criterios", sinCumplir, true, backlog.ProgresoDe(criteriosConCumplidos(2, 1)), backlog.ErrCriteriosSinCumplir,
			"HU-2 no se puede completar: falta cumplir 1 de 2 criterios"},
		{"sin criterios", sinCumplir, true, backlog.Progreso{}, backlog.ErrCriteriosSinCumplir,
			"HU-2 no se puede completar: no tiene criterios de aceptación"},
		{"todavía no empezó", sinEmpezar, true, progresoCompleto(), backlog.ErrTransicionInvalida,
			"HU-3 todavía no empezó: primero hay que pasarla a En progreso"},
		{"ya está hecha", hecha, true, progresoCompleto(), backlog.ErrYaEstaEnEseEstado,
			"HU-1 ya está Hecha"},
		{"no es del sprint activo", enProgreso(), false, progresoCompleto(), backlog.ErrNoEsElSprintActivo,
			"solo se completan historias del sprint activo"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := c.historia.Completar(c.activo, c.progreso, quinceDeOctubre)
			if !errors.Is(err, c.err) {
				t.Fatalf("error = %v, se esperaba %v", err, c.err)
			}
			if err.Error() != c.mensaje {
				t.Errorf("mensaje = %q, se esperaba %q", err.Error(), c.mensaje)
			}
		})
	}
}
