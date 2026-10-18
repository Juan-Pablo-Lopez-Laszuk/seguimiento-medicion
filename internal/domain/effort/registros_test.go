package effort_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
)

func dia(d int) time.Time { return time.Date(2026, 10, d, 0, 0, 0, 0, time.UTC) }

// contextoDeEjemplo es el de la spec de HU-24: proyecto del 05/10 al 01/11/2026, hoy 18/10, la historia 9 con la
// tarea 5, Carolina (2) activa y Ana (4) dada de baja.
func contextoDeEjemplo() effort.ContextoRegistro {
	return effort.ContextoRegistro{
		HistoriaID:     9,
		InicioProyecto: dia(5),
		FinProyecto:    time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
		Hoy:            dia(18),
		Integrantes:    integrantesDelProyecto(),
		Tareas:         []effort.Tarea{{ID: 5, HistoriaID: 9, Titulo: "Pantalla de login"}},
	}
}

func registroValido() effort.DatosRegistro {
	return effort.DatosRegistro{IntegranteID: 2, Fecha: "2026-10-18", Actividad: " Pantalla de login ", Horas: "3"}
}

// CA-24.1 y CA-24.3 · el registro guarda integrante, historia, fecha, actividad sin espacios sobrantes y horas
func TestNuevoRegistro_DatosValidos(t *testing.T) {
	d := registroValido()
	d.TareaID = 5

	r, err := effort.NuevoRegistro(d, contextoDeEjemplo())
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	esperado := effort.Registro{IntegranteID: 2, HistoriaID: 9, TareaID: 5, Fecha: dia(18), Actividad: "Pantalla de login", Horas: 3}
	if r != esperado {
		t.Errorf("registro = %+v, se esperaba %+v", r, esperado)
	}
}

// Cada regla de la spec se informa en su campo (sección 7).
func TestNuevoRegistro_Validaciones(t *testing.T) {
	casos := []struct {
		nombre string
		cambio func(*effort.DatosRegistro, *effort.ContextoRegistro)
		campo  string
		err    error
	}{
		{"24 h es el máximo", func(d *effort.DatosRegistro, _ *effort.ContextoRegistro) { d.Horas = "24" }, "", nil},
		{"0 h", func(d *effort.DatosRegistro, _ *effort.ContextoRegistro) { d.Horas = "0" }, effort.CampoHoras, effort.ErrHorasTrabajadasInvalidas},
		{"24,5 h", func(d *effort.DatosRegistro, _ *effort.ContextoRegistro) { d.Horas = "24,5" }, effort.CampoHoras, effort.ErrHorasTrabajadasInvalidas},
		{"horas que no son número", func(d *effort.DatosRegistro, _ *effort.ContextoRegistro) { d.Horas = "abc" }, effort.CampoHoras, effort.ErrHorasTrabajadasInvalidas},
		{"fecha vacía", func(d *effort.DatosRegistro, _ *effort.ContextoRegistro) { d.Fecha = "" }, effort.CampoFecha, effort.ErrFechaInvalida},
		{"fecha futura", func(d *effort.DatosRegistro, _ *effort.ContextoRegistro) { d.Fecha = "2026-10-19" }, effort.CampoFecha, effort.ErrFechaFutura},
		{"antes del proyecto", func(d *effort.DatosRegistro, _ *effort.ContextoRegistro) { d.Fecha = "2026-10-04" }, effort.CampoFecha, effort.ErrFechaFueraDelProyecto},
		{"primer día del proyecto", func(d *effort.DatosRegistro, _ *effort.ContextoRegistro) { d.Fecha = "2026-10-05" }, "", nil},
		{"después del proyecto", func(d *effort.DatosRegistro, c *effort.ContextoRegistro) {
			d.Fecha = "2026-11-02"
			c.Hoy = dia(30).AddDate(0, 0, 5)
		}, effort.CampoFecha, effort.ErrFechaFueraDelProyecto},
		{"sin integrante", func(d *effort.DatosRegistro, _ *effort.ContextoRegistro) { d.IntegranteID = 0 }, effort.CampoIntegrante, effort.ErrIntegranteInvalido},
		{"integrante dado de baja", func(d *effort.DatosRegistro, _ *effort.ContextoRegistro) { d.IntegranteID = 4 }, effort.CampoIntegrante, effort.ErrIntegranteInvalido},
		{"integrante de otro proyecto", func(d *effort.DatosRegistro, _ *effort.ContextoRegistro) { d.IntegranteID = 99 }, effort.CampoIntegrante, effort.ErrIntegranteInvalido},
		{"tarea de otra historia", func(d *effort.DatosRegistro, _ *effort.ContextoRegistro) { d.TareaID = 77 }, effort.CampoTarea, effort.ErrTareaDeOtraHistoria},
		{"actividad vacía", func(d *effort.DatosRegistro, _ *effort.ContextoRegistro) { d.Actividad = "  " }, effort.CampoActividad, effort.ErrActividadObligatoria},
		{"actividad de 200", func(d *effort.DatosRegistro, _ *effort.ContextoRegistro) { d.Actividad = strings.Repeat("ñ", 200) }, "", nil},
		{"actividad de 201", func(d *effort.DatosRegistro, _ *effort.ContextoRegistro) { d.Actividad = strings.Repeat("a", 201) }, effort.CampoActividad, effort.ErrActividadLarga},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d, ctx := registroValido(), contextoDeEjemplo()
			c.cambio(&d, &ctx)

			_, err := effort.NuevoRegistro(d, ctx)

			if c.err == nil {
				if err != nil {
					t.Fatalf("no se esperaba error: %v", err)
				}
				return
			}
			if got := errorDelCampo(t, err, c.campo); !errors.Is(got, c.err) {
				t.Errorf("error de %s = %v, se esperaba %v", c.campo, got, c.err)
			}
		})
	}
}

// CA-24.1 (RN2) · un integrante no suma más de 24 h en un día; los registros de otros días u otros integrantes no cuentan
func TestNuevoRegistro_TotalDelDia(t *testing.T) {
	ctx := contextoDeEjemplo()
	ctx.DelIntegrante = []effort.Registro{
		{IntegranteID: 2, Fecha: dia(17), Horas: 15},
		{IntegranteID: 2, Fecha: dia(17), Horas: 5},
		{IntegranteID: 2, Fecha: dia(16), Horas: 10},
	}
	d := registroValido()
	d.Fecha = "2026-10-17"

	d.Horas = "4"
	if _, err := effort.NuevoRegistro(d, ctx); err != nil {
		t.Errorf("20 + 4 = 24 h: no se esperaba error y llegó %v", err)
	}
	d.Horas = "4,5"
	if _, err := effort.NuevoRegistro(d, ctx); !errors.Is(errorDelCampo(t, err, effort.CampoHoras), effort.ErrMasDe24HorasEnElDia) {
		t.Errorf("20 + 4,5 h: se esperaba %v y llegó %v", effort.ErrMasDe24HorasEnElDia, err)
	}
	d.Fecha, d.Horas = "2026-10-18", "24"
	if _, err := effort.NuevoRegistro(d, ctx); err != nil {
		t.Errorf("otro día: no se esperaba error y llegó %v", err)
	}
}

// RN8 · los errores de todos los campos se informan juntos
func TestNuevoRegistro_InformaTodosLosErrores(t *testing.T) {
	_, err := effort.NuevoRegistro(effort.DatosRegistro{TareaID: 77}, contextoDeEjemplo())

	for _, campo := range []string{effort.CampoIntegrante, effort.CampoTarea, effort.CampoFecha, effort.CampoActividad, effort.CampoHoras} {
		if errorDelCampo(t, err, campo) == nil {
			t.Errorf("falta el error del campo %q", campo)
		}
	}
}
