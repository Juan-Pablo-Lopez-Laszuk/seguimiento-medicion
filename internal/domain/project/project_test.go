package project_test

import (
	"errors"
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

// verificarErrorDeCampo comprueba que err sea un error de validación con el error esperado en ese campo (CA-01.4).
func verificarErrorDeCampo(t *testing.T, err error, campo string, esperado error) {
	t.Helper()
	var errs project.ErroresValidacion
	if !errors.As(err, &errs) {
		t.Fatalf("se esperaba project.ErroresValidacion y se obtuvo: %v", err)
	}
	if !errors.Is(errs[campo], esperado) {
		t.Errorf("campo %q: se esperaba %v y se obtuvo %v (todos: %v)", campo, esperado, errs[campo], errs)
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

// CA-01.1 y CA-01.4
func TestNuevo_NombreVacioOSoloEspacios_InformaQueEsObligatorio(t *testing.T) {
	for _, nombre := range []string{"", "   "} {
		d := datosValidos(t)
		d.Nombre = nombre

		_, err := project.Nuevo(d)

		verificarErrorDeCampo(t, err, project.CampoNombre, project.ErrNombreObligatorio)
	}
}

// RN1
func TestNuevo_QuitaLosEspaciosAlrededorDelNombre(t *testing.T) {
	d := datosValidos(t)
	d.Nombre = "   Gestión Ñandú   "

	p, err := project.Nuevo(d)

	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo: %v", err)
	}
	if p.Nombre != "Gestión Ñandú" {
		t.Errorf("nombre: se esperaba %q y se obtuvo %q", "Gestión Ñandú", p.Nombre)
	}
}
