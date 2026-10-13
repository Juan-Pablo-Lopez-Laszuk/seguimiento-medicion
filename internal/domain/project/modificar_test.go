package project_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
)

// proyectoGuardado es el proyecto de la spec de HU-02 tal como está guardado: del 05/10/2026 al 01/11/2026.
func proyectoGuardado(t *testing.T) project.Proyecto {
	return project.Proyecto{
		ID:          7,
		Nombre:      "Software Metrics",
		FechaInicio: fecha(t, "2026-10-05"),
		FechaFin:    fecha(t, "2026-11-01"),
		Estado:      project.EstadoPlanificado,
		CreadoEn:    time.Date(2026, 10, 5, 14, 30, 0, 0, time.UTC),
	}
}

// conDosSprints son los sprints de la spec: el Sprint 1 empieza el 05/10 y el Sprint 2 termina el 18/10.
func conDosSprints(t *testing.T) project.LimitesSprints {
	return project.LimitesSprints{
		PrimerNumero: 1, PrimerInicio: fecha(t, "2026-10-05"),
		UltimoNumero: 2, UltimoFin: fecha(t, "2026-10-18"),
	}
}

// cambios devuelve los datos actuales del proyecto con las fechas indicadas, para que cada test cambie solo eso.
func cambios(t *testing.T, inicio, fin string) project.Datos {
	return project.Datos{Nombre: "Software Metrics", FechaInicio: fecha(t, inicio), FechaFin: fecha(t, fin)}
}

// CA-02.1 · cambiar la descripción guarda los datos nuevos y no toca el ID, la creación ni el estado guardado
func TestModificar_CambiarDescripcion(t *testing.T) {
	actual := proyectoGuardado(t)
	d := cambios(t, "2026-10-05", "2026-11-01")
	d.Descripcion = "TPI de ICS 2026"

	p, err := project.Modificar(actual, d, conDosSprints(t), fecha(t, "2026-10-13"))
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if p.Descripcion != "TPI de ICS 2026" || p.Nombre != "Software Metrics" {
		t.Errorf("se esperaban los datos nuevos y se obtuvo %+v", p)
	}
	if p.ID != actual.ID || !p.CreadoEn.Equal(actual.CreadoEn) || p.Estado != actual.Estado {
		t.Errorf("no tenían que cambiar el ID, la creación ni el estado: %+v", p)
	}
}

// RN1 · el nombre se guarda sin los espacios de alrededor, igual que al crear
func TestModificar_NormalizaElNombre(t *testing.T) {
	d := cambios(t, "2026-10-05", "2026-11-01")
	d.Nombre = "  Gestión Ñandú  "

	p, err := project.Modificar(proyectoGuardado(t), d, conDosSprints(t), fecha(t, "2026-10-13"))
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if p.Nombre != "Gestión Ñandú" {
		t.Errorf("se esperaba \"Gestión Ñandú\" y se obtuvo %q", p.Nombre)
	}
}

// CA-02.1 · aplican las validaciones de la creación, todas juntas
func TestModificar_ValidacionesDeLaCreacion(t *testing.T) {
	d := cambios(t, "2026-10-05", "2026-10-05")
	d.Nombre = "ab"

	_, err := project.Modificar(proyectoGuardado(t), d, conDosSprints(t), fecha(t, "2026-10-13"))

	verificarErrorDeCampo(t, err, project.CampoNombre, project.ErrNombreLargo)
	verificarErrorDeCampo(t, err, project.CampoFechaFin, project.ErrFechasInvalidas)
}

// CA-02.2 y RN4 · las fechas no pueden dejar sprints afuera; los bordes valen
func TestModificar_FechasContraLosSprints(t *testing.T) {
	casos := []struct {
		nombre      string
		inicio, fin string
		campo       string
		esperado    error
		detalle     string
	}{
		{"fin el día que termina el último sprint", "2026-10-05", "2026-10-18", "", nil, ""},
		{"fin antes del último sprint", "2026-10-05", "2026-10-17", project.CampoFechaFin,
			project.ErrFinAntesDelUltimoSprint, "(el Sprint 2 termina el 18/10/2026)"},
		{"inicio el día que empieza el primer sprint", "2026-10-05", "2026-11-01", "", nil, ""},
		{"inicio antes del primer sprint", "2026-10-01", "2026-11-01", "", nil, ""},
		{"inicio después del primer sprint", "2026-10-06", "2026-11-01", project.CampoFechaInicio,
			project.ErrInicioDespuesDelPrimerSprint, "(el Sprint 1 empieza el 05/10/2026)"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := project.Modificar(proyectoGuardado(t), cambios(t, c.inicio, c.fin), conDosSprints(t), fecha(t, "2026-10-13"))
			if c.esperado == nil {
				if err != nil {
					t.Fatalf("no se esperaba error: %v", err)
				}
				return
			}
			verificarErrorDeCampo(t, err, c.campo, c.esperado)
			if !strings.Contains(err.Error(), c.detalle) {
				t.Errorf("el mensaje %q tendría que decir %q", err.Error(), c.detalle)
			}
		})
	}
}

// Un solo error por campo: si la fecha de fin ya es inválida por HU-01, no se agrega el error de los sprints
func TestModificar_UnErrorPorCampo(t *testing.T) {
	_, err := project.Modificar(proyectoGuardado(t), cambios(t, "2026-10-05", "2026-10-01"), conDosSprints(t), fecha(t, "2026-10-13"))
	verificarErrorDeCampo(t, err, project.CampoFechaFin, project.ErrFechasInvalidas)
}

// Sección 6 · sin sprints vale cualquier fecha válida
func TestModificar_SinSprints(t *testing.T) {
	_, err := project.Modificar(proyectoGuardado(t), cambios(t, "2026-10-20", "2026-10-25"), project.LimitesSprints{}, fecha(t, "2026-10-13"))
	if err != nil {
		t.Errorf("no se esperaba error: %v", err)
	}
}

// CA-02.3 · un proyecto Finalizado no se modifica, aunque los datos nuevos sean válidos
func TestModificar_ProyectoFinalizado(t *testing.T) {
	_, err := project.Modificar(proyectoGuardado(t), cambios(t, "2026-10-05", "2026-11-01"), conDosSprints(t), fecha(t, "2026-11-02"))

	if !errors.Is(err, project.ErrProyectoFinalizado) {
		t.Fatalf("se esperaba ErrProyectoFinalizado y se obtuvo %v", err)
	}
	var errs domain.ErroresValidacion
	if errors.As(err, &errs) {
		t.Error("un proyecto Finalizado no es un error de un campo del formulario")
	}
}

// RN5 · el estado se calcula con las fechas guardadas: sin sprints sigue Planificado y se puede modificar
func TestModificar_SinSprintsYFechaPasada_SePuedeModificar(t *testing.T) {
	_, err := project.Modificar(proyectoGuardado(t), cambios(t, "2026-10-05", "2026-11-20"), project.LimitesSprints{}, fecha(t, "2026-11-10"))
	if err != nil {
		t.Errorf("no se esperaba error: %v", err)
	}
}

// RN5 · PuedeModificarse le dice a la pantalla si muestra el formulario
func TestPuedeModificarse(t *testing.T) {
	p := proyectoGuardado(t)
	if !project.PuedeModificarse(p, conDosSprints(t), fecha(t, "2026-11-01")) {
		t.Error("el último día del proyecto todavía se puede modificar")
	}
	if project.PuedeModificarse(p, conDosSprints(t), fecha(t, "2026-11-02")) {
		t.Error("un proyecto Finalizado no se puede modificar")
	}
}
