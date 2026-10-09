package sprint_test

import (
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
)

// fecha arma una fecha sin hora, como las que llegan del formulario.
func fecha(t *testing.T, texto string) time.Time {
	t.Helper()
	f, err := time.Parse(time.DateOnly, texto)
	if err != nil {
		t.Fatalf("fecha de prueba inválida %q: %v", texto, err)
	}
	return f
}

// proyecto devuelve el proyecto de ejemplo de la spec: del 05/10/2026 al 01/11/2026.
func proyecto(t *testing.T) project.Proyecto {
	return project.Proyecto{
		ID:          7,
		Nombre:      "Software Metrics",
		FechaInicio: fecha(t, "2026-10-05"),
		FechaFin:    fecha(t, "2026-11-01"),
		Estado:      project.EstadoPlanificado,
	}
}

// datosValidos devuelve datos que cumplen todas las reglas; cada test cambia solo lo que quiere probar.
func datosValidos(t *testing.T) sprint.Datos {
	return sprint.Datos{
		Objetivo:    "Primer MVP",
		FechaInicio: fecha(t, "2026-10-05"),
		FechaFin:    fecha(t, "2026-10-11"),
	}
}

// CA-11.3 · el primer sprint del proyecto es el 1 y nace Planificado
func TestNuevo_PrimerSprint_TieneElNumero1YEstadoPlanificado(t *testing.T) {
	d := datosValidos(t)
	p := proyecto(t)

	s, err := sprint.Nuevo(d, p, nil)

	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo: %v", err)
	}
	if s.Numero != 1 {
		t.Errorf("número: se esperaba 1 y se obtuvo %d", s.Numero)
	}
	if s.Estado != sprint.EstadoPlanificado {
		t.Errorf("estado: se esperaba %q y se obtuvo %q", sprint.EstadoPlanificado, s.Estado)
	}
	if s.ProyectoID != p.ID || s.Objetivo != d.Objetivo {
		t.Errorf("datos: se esperaba proyecto %d y goal %q y se obtuvo %d y %q", p.ID, d.Objetivo, s.ProyectoID, s.Objetivo)
	}
	if !s.FechaInicio.Equal(d.FechaInicio) || !s.FechaFin.Equal(d.FechaFin) {
		t.Errorf("fechas: se esperaba %v → %v y se obtuvo %v → %v", d.FechaInicio, d.FechaFin, s.FechaInicio, s.FechaFin)
	}
}
