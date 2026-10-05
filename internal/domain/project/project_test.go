package project_test

import (
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
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

// datosValidos devuelve datos que cumplen todas las reglas; cada test cambia solo lo que quiere probar.
func datosValidos(t *testing.T) project.Datos {
	return project.Datos{
		Nombre:      "Software Metrics",
		Descripcion: "TPI de ICS",
		FechaInicio: fecha(t, "2026-10-05"),
		FechaFin:    fecha(t, "2026-11-01"),
	}
}

// CA-01.3
func TestNuevo_DatosValidos_CreaElProyectoEnEstadoPlanificado(t *testing.T) {
	d := datosValidos(t)

	p, err := project.Nuevo(d)

	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo: %v", err)
	}
	if p.Nombre != d.Nombre || p.Descripcion != d.Descripcion {
		t.Errorf("datos: se esperaba %q / %q y se obtuvo %q / %q", d.Nombre, d.Descripcion, p.Nombre, p.Descripcion)
	}
	if !p.FechaInicio.Equal(d.FechaInicio) || !p.FechaFin.Equal(d.FechaFin) {
		t.Errorf("fechas: se esperaba %v → %v y se obtuvo %v → %v", d.FechaInicio, d.FechaFin, p.FechaInicio, p.FechaFin)
	}
	if p.Estado != project.EstadoPlanificado {
		t.Errorf("estado: se esperaba %q y se obtuvo %q", project.EstadoPlanificado, p.Estado)
	}
}
