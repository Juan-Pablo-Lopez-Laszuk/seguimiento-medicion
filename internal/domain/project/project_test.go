package project_test

import (
	"errors"
	"strings"
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

// CA-01.1 · casos límite del largo del nombre (RN2: se cuentan caracteres, no bytes)
func TestNuevo_LargoDelNombre(t *testing.T) {
	casos := []struct {
		nombre string
		valido bool
	}{
		{"ab", false},
		{"abc", true},
		{"ñú", false}, // 2 caracteres aunque ocupen 4 bytes
		{"Año", true}, // 3 caracteres aunque ocupen 4 bytes
		{strings.Repeat("a", 100), true},
		{strings.Repeat("ñ", 100), true}, // 100 caracteres, 200 bytes
		{strings.Repeat("a", 101), false},
		{"   ab   ", false}, // sin los espacios quedan 2
	}
	for _, c := range casos {
		d := datosValidos(t)
		d.Nombre = c.nombre

		_, err := project.Nuevo(d)

		if c.valido && err != nil {
			t.Errorf("nombre de %d caracteres: no se esperaba error, se obtuvo %v", len([]rune(c.nombre)), err)
		}
		if !c.valido {
			verificarErrorDeCampo(t, err, project.CampoNombre, project.ErrNombreLargo)
		}
	}
}

// Descripción opcional, hasta 1000 caracteres.
func TestNuevo_LargoDeLaDescripcion(t *testing.T) {
	casos := []struct {
		descripcion string
		valido      bool
	}{
		{"", true},
		{strings.Repeat("ñ", 1000), true},
		{strings.Repeat("a", 1001), false},
	}
	for _, c := range casos {
		d := datosValidos(t)
		d.Descripcion = c.descripcion

		_, err := project.Nuevo(d)

		if c.valido && err != nil {
			t.Errorf("descripción de %d caracteres: no se esperaba error, se obtuvo %v", len([]rune(c.descripcion)), err)
		}
		if !c.valido {
			verificarErrorDeCampo(t, err, project.CampoDescripcion, project.ErrDescripcionLarga)
		}
	}
}

// CA-01.2 y CA-01.4 · fechas faltantes (la fecha cero es la que llega cuando el campo está vacío o no es una fecha)
func TestNuevo_FechasObligatorias(t *testing.T) {
	d := datosValidos(t)
	d.FechaInicio = time.Time{}
	_, err := project.Nuevo(d)
	verificarErrorDeCampo(t, err, project.CampoFechaInicio, project.ErrFechaInicio)

	d = datosValidos(t)
	d.FechaFin = time.Time{}
	_, err = project.Nuevo(d)
	verificarErrorDeCampo(t, err, project.CampoFechaFin, project.ErrFechaFin)
}

// CA-01.2 · casos límite: la fecha de fin tiene que ser estrictamente posterior (RN4)
func TestNuevo_FechaDeFinPosteriorALaDeInicio(t *testing.T) {
	casos := []struct {
		inicio, fin string
		valido      bool
	}{
		{"2026-10-05", "2026-10-05", false}, // mismo día
		{"2026-10-05", "2026-10-06", true},  // un día después
		{"2026-10-05", "2026-10-04", false}, // anterior
		{"2020-01-01", "2020-12-31", true},  // RN5: el inicio puede estar en el pasado
	}
	for _, c := range casos {
		d := datosValidos(t)
		d.FechaInicio, d.FechaFin = fecha(t, c.inicio), fecha(t, c.fin)

		_, err := project.Nuevo(d)

		if c.valido && err != nil {
			t.Errorf("%s → %s: no se esperaba error, se obtuvo %v", c.inicio, c.fin, err)
		}
		if !c.valido {
			verificarErrorDeCampo(t, err, project.CampoFechaFin, project.ErrFechasInvalidas)
		}
	}
}

// Las fechas se comparan y se guardan sin hora: dos horarios del mismo día son el mismo día.
func TestNuevo_LasFechasSeTomanSinHora(t *testing.T) {
	d := datosValidos(t)
	d.FechaInicio = time.Date(2026, 10, 5, 10, 0, 0, 0, time.UTC)
	d.FechaFin = time.Date(2026, 10, 5, 18, 0, 0, 0, time.UTC)
	_, err := project.Nuevo(d)
	verificarErrorDeCampo(t, err, project.CampoFechaFin, project.ErrFechasInvalidas)

	d.FechaFin = time.Date(2026, 10, 6, 9, 30, 0, 0, time.UTC)
	p, err := project.Nuevo(d)
	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo: %v", err)
	}
	if !p.FechaInicio.Equal(fecha(t, "2026-10-05")) || !p.FechaFin.Equal(fecha(t, "2026-10-06")) {
		t.Errorf("se esperaban las fechas sin hora y se obtuvo %v → %v", p.FechaInicio, p.FechaFin)
	}
}

// RN7 · si hay varios datos inválidos se informan todos juntos
func TestNuevo_InformaTodosLosCamposInvalidosJuntos(t *testing.T) {
	d := datosValidos(t)
	d.Nombre = ""
	d.FechaInicio = time.Time{}

	_, err := project.Nuevo(d)

	verificarErrorDeCampo(t, err, project.CampoNombre, project.ErrNombreObligatorio)
	verificarErrorDeCampo(t, err, project.CampoFechaInicio, project.ErrFechaInicio)
}

// El texto del error junta todos los campos, ordenados, con su mensaje.
func TestErroresValidacion_TextoConTodosLosCamposOrdenados(t *testing.T) {
	errs := project.ErroresValidacion{
		project.CampoNombre:      project.ErrNombreObligatorio,
		project.CampoFechaInicio: project.ErrFechaInicio,
	}

	esperado := "fecha_inicio: la fecha de inicio es obligatoria y debe ser válida; nombre: el nombre es obligatorio"
	if errs.Error() != esperado {
		t.Errorf("se esperaba %q y se obtuvo %q", esperado, errs.Error())
	}
}
