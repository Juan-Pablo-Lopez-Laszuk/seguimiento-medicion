package sprint_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
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

// verificarErrorDeCampo comprueba que err sea un error de validación con el error esperado en ese campo.
func verificarErrorDeCampo(t *testing.T, err error, campo string, esperado error) {
	t.Helper()
	var errs domain.ErroresValidacion
	if !errors.As(err, &errs) {
		t.Fatalf("se esperaba domain.ErroresValidacion y se obtuvo: %v", err)
	}
	if !errors.Is(errs[campo], esperado) {
		t.Errorf("campo %q: se esperaba %v y se obtuvo %v (todos: %v)", campo, esperado, errs[campo], errs)
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

// CA-11.1 · RN1: el Sprint Goal es obligatorio; solo espacios cuenta como vacío
func TestNuevo_SprintGoalVacioOSoloEspacios_InformaQueEsObligatorio(t *testing.T) {
	for _, objetivo := range []string{"", "   "} {
		d := datosValidos(t)
		d.Objetivo = objetivo

		_, err := sprint.Nuevo(d, proyecto(t), nil)

		verificarErrorDeCampo(t, err, sprint.CampoObjetivo, sprint.ErrObjetivoObligatorio)
	}
}

// RN1 · el Sprint Goal se guarda sin los espacios de alrededor
func TestNuevo_QuitaLosEspaciosAlrededorDelSprintGoal(t *testing.T) {
	d := datosValidos(t)
	d.Objetivo = "  Primer MVP con métricas  "

	s, err := sprint.Nuevo(d, proyecto(t), nil)

	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo: %v", err)
	}
	if s.Objetivo != "Primer MVP con métricas" {
		t.Errorf("goal: se esperaba %q y se obtuvo %q", "Primer MVP con métricas", s.Objetivo)
	}
}

// RN1 · casos límite del largo (se cuentan caracteres, no bytes)
func TestNuevo_LargoDelSprintGoal(t *testing.T) {
	casos := []struct {
		objetivo string
		valido   bool
	}{
		{"a", true},
		{strings.Repeat("ñ", sprint.ObjetivoMax), true}, // 500 caracteres, 1000 bytes
		{strings.Repeat("a", sprint.ObjetivoMax+1), false},
	}
	for _, c := range casos {
		d := datosValidos(t)
		d.Objetivo = c.objetivo

		_, err := sprint.Nuevo(d, proyecto(t), nil)

		if c.valido && err != nil {
			t.Errorf("goal de %d caracteres: no se esperaba error, se obtuvo %v", len([]rune(c.objetivo)), err)
		}
		if !c.valido {
			verificarErrorDeCampo(t, err, sprint.CampoObjetivo, sprint.ErrObjetivoLargo)
		}
	}
}

// Fechas faltantes (la fecha cero es la que llega cuando el campo está vacío o no es una fecha).
func TestNuevo_FechasObligatorias(t *testing.T) {
	d := datosValidos(t)
	d.FechaInicio = time.Time{}
	_, err := sprint.Nuevo(d, proyecto(t), nil)
	verificarErrorDeCampo(t, err, sprint.CampoFechaInicio, sprint.ErrFechaInicio)

	d = datosValidos(t)
	d.FechaFin = time.Time{}
	_, err = sprint.Nuevo(d, proyecto(t), nil)
	verificarErrorDeCampo(t, err, sprint.CampoFechaFin, sprint.ErrFechaFin)
}

// RN2 · la fecha de fin tiene que ser estrictamente posterior a la de inicio
func TestNuevo_FechaDeFinPosteriorALaDeInicio(t *testing.T) {
	casos := []struct {
		inicio, fin string
		valido      bool
	}{
		{"2026-10-05", "2026-10-05", false}, // mismo día
		{"2026-10-05", "2026-10-06", true},  // un día después
		{"2026-10-08", "2026-10-06", false}, // anterior
	}
	for _, c := range casos {
		d := datosValidos(t)
		d.FechaInicio, d.FechaFin = fecha(t, c.inicio), fecha(t, c.fin)

		_, err := sprint.Nuevo(d, proyecto(t), nil)

		if c.valido && err != nil {
			t.Errorf("%s → %s: no se esperaba error, se obtuvo %v", c.inicio, c.fin, err)
		}
		if !c.valido {
			verificarErrorDeCampo(t, err, sprint.CampoFechaFin, sprint.ErrFechasInvalidas)
		}
	}
}

// Las fechas se guardan sin hora, como en HU-01.
func TestNuevo_LasFechasSeTomanSinHora(t *testing.T) {
	d := datosValidos(t)
	d.FechaInicio = time.Date(2026, 10, 5, 10, 0, 0, 0, time.UTC)
	d.FechaFin = time.Date(2026, 10, 11, 18, 30, 0, 0, time.UTC)

	s, err := sprint.Nuevo(d, proyecto(t), nil)

	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo: %v", err)
	}
	if !s.FechaInicio.Equal(fecha(t, "2026-10-05")) || !s.FechaFin.Equal(fecha(t, "2026-10-11")) {
		t.Errorf("se esperaban las fechas sin hora y se obtuvo %v → %v", s.FechaInicio, s.FechaFin)
	}
}

// RN7 · si hay varios datos inválidos se informan todos juntos
func TestNuevo_InformaTodosLosCamposInvalidosJuntos(t *testing.T) {
	d := datosValidos(t)
	d.Objetivo = "   "
	d.FechaInicio = time.Time{}

	_, err := sprint.Nuevo(d, proyecto(t), nil)

	verificarErrorDeCampo(t, err, sprint.CampoObjetivo, sprint.ErrObjetivoObligatorio)
	verificarErrorDeCampo(t, err, sprint.CampoFechaInicio, sprint.ErrFechaInicio)
}

// CA-11.2 · RN3: el sprint queda dentro del proyecto (del 05/10 al 01/11); los bordes valen
func TestNuevo_FechasDentroDelProyecto(t *testing.T) {
	casos := []struct {
		inicio, fin string
		campo       string // campo con error; vacío si es válido
	}{
		{"2026-10-05", "2026-11-01", ""},                      // justo los bordes del proyecto
		{"2026-10-04", "2026-10-11", sprint.CampoFechaInicio}, // un día antes del inicio
		{"2026-10-26", "2026-11-02", sprint.CampoFechaFin},    // un día después del fin
		{"2026-11-02", "2026-11-08", sprint.CampoFechaInicio}, // empieza cuando el proyecto ya terminó
		{"2026-09-01", "2026-09-07", sprint.CampoFechaInicio}, // termina antes de que empiece el proyecto
	}
	for _, c := range casos {
		d := datosValidos(t)
		d.FechaInicio, d.FechaFin = fecha(t, c.inicio), fecha(t, c.fin)

		_, err := sprint.Nuevo(d, proyecto(t), nil)

		if c.campo == "" && err != nil {
			t.Errorf("%s → %s: no se esperaba error, se obtuvo %v", c.inicio, c.fin, err)
		}
		if c.campo != "" {
			verificarErrorDeCampo(t, err, c.campo, sprint.ErrFueraDelProyecto)
		}
	}
}

// El mensaje dice las fechas del proyecto, para que el usuario sepa qué rango tiene.
func TestNuevo_FueraDelProyecto_ElMensajeIncluyeElRango(t *testing.T) {
	d := datosValidos(t)
	d.FechaInicio = fecha(t, "2026-10-04")

	_, err := sprint.Nuevo(d, proyecto(t), nil)

	esperado := "fecha_inicio: la fecha tiene que estar dentro del proyecto (del 05/10/2026 al 01/11/2026)"
	if err == nil || err.Error() != esperado {
		t.Errorf("se esperaba %q y se obtuvo %v", esperado, err)
	}
}

// sprintExistente arma un sprint ya guardado del proyecto de ejemplo.
func sprintExistente(t *testing.T, numero int, inicio, fin string) sprint.Sprint {
	return sprint.Sprint{
		ID:          int64(numero),
		ProyectoID:  proyecto(t).ID,
		Numero:      numero,
		Objetivo:    "Sprint anterior",
		FechaInicio: fecha(t, inicio),
		FechaFin:    fecha(t, fin),
		Estado:      sprint.EstadoPlanificado,
	}
}

// CA-11.2 · RN4: el sprint nuevo empieza después del día en que termina el último
func TestNuevo_EmpiezaDespuesDelUltimoSprint(t *testing.T) {
	existentes := []sprint.Sprint{
		sprintExistente(t, 1, "2026-10-05", "2026-10-11"),
		sprintExistente(t, 2, "2026-10-12", "2026-10-18"),
	}
	casos := []struct {
		inicio, fin string
		valido      bool
	}{
		{"2026-10-18", "2026-10-25", false}, // empieza el día que termina el último
		{"2026-10-15", "2026-10-17", false}, // queda adentro del último
		{"2026-10-05", "2026-10-08", false}, // queda antes del último (en el lugar del Sprint 1)
		{"2026-10-19", "2026-10-25", true},  // el día siguiente
		{"2026-10-22", "2026-10-28", true},  // con días libres en el medio
	}
	for _, c := range casos {
		d := datosValidos(t)
		d.FechaInicio, d.FechaFin = fecha(t, c.inicio), fecha(t, c.fin)

		_, err := sprint.Nuevo(d, proyecto(t), existentes)

		if c.valido && err != nil {
			t.Errorf("%s → %s: no se esperaba error, se obtuvo %v", c.inicio, c.fin, err)
		}
		if !c.valido {
			verificarErrorDeCampo(t, err, sprint.CampoFechaInicio, sprint.ErrSuperpuesto)
		}
	}
}

// El mensaje dice cuál es el último sprint y cuándo termina.
func TestNuevo_Superpuesto_ElMensajeDiceCuandoTerminaElUltimo(t *testing.T) {
	existentes := []sprint.Sprint{sprintExistente(t, 1, "2026-10-05", "2026-10-11")}
	d := datosValidos(t)
	d.FechaInicio, d.FechaFin = fecha(t, "2026-10-11"), fecha(t, "2026-10-18")

	_, err := sprint.Nuevo(d, proyecto(t), existentes)

	esperado := "fecha_inicio: el sprint tiene que empezar después del último (el Sprint 1 termina el 11/10/2026)"
	if err == nil || err.Error() != esperado {
		t.Errorf("se esperaba %q y se obtuvo %v", esperado, err)
	}
}
