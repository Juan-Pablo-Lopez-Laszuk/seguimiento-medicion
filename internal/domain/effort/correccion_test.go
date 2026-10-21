package effort_test

import (
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
)

// CA-26.1 y RN2 · los filtros se combinan; el sprint se toma por las fechas (inclusive)
func TestFiltro_Cumple(t *testing.T) {
	r := effort.Registro{ID: 1, IntegranteID: 2, HistoriaID: 9, Fecha: dia(10), Horas: 3}
	casos := []struct {
		nombre string
		filtro effort.Filtro
		cumple bool
	}{
		{"sin filtros", effort.Filtro{}, true},
		{"su integrante", effort.Filtro{IntegranteID: 2}, true},
		{"otro integrante", effort.Filtro{IntegranteID: 3}, false},
		{"su historia", effort.Filtro{HistoriaID: 9}, true},
		{"otra historia", effort.Filtro{HistoriaID: 8}, false},
		{"fecha dentro del sprint", effort.Filtro{Desde: dia(5), Hasta: dia(18)}, true},
		{"fecha igual al inicio", effort.Filtro{Desde: dia(10), Hasta: dia(18)}, true},
		{"fecha igual al fin", effort.Filtro{Desde: dia(5), Hasta: dia(10)}, true},
		{"fecha antes del sprint", effort.Filtro{Desde: dia(11), Hasta: dia(18)}, false},
		{"fecha después del sprint", effort.Filtro{Desde: dia(1), Hasta: dia(9)}, false},
		{"todo combinado", effort.Filtro{IntegranteID: 2, HistoriaID: 9, Desde: dia(5), Hasta: dia(18)}, true},
		{"combinado con uno que no cumple", effort.Filtro{IntegranteID: 2, HistoriaID: 8, Desde: dia(5), Hasta: dia(18)}, false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := c.filtro.Cumple(r); got != c.cumple {
				t.Errorf("Cumple = %v, se esperaba %v", got, c.cumple)
			}
		})
	}
}

// CA-26.2 y RN4 · cada uno corrige sus registros y el Agile Enabler todos
func TestPuedeCorregir(t *testing.T) {
	deCarolina := effort.Registro{ID: 1, IntegranteID: 2}
	casos := []struct {
		nombre string
		quien  member.Integrante
		err    error
	}{
		{"la dueña del registro", member.Integrante{ID: 2, Rol: member.RolProductBuilder, Activo: true}, nil},
		{"el Agile Enabler", member.Integrante{ID: 1, Rol: member.RolAgileEnabler, Activo: true}, nil},
		{"otro Product Builder", member.Integrante{ID: 3, Rol: member.RolProductBuilder, Activo: true}, effort.ErrNoPuedeCorregir},
		{"la dueña dada de baja", member.Integrante{ID: 2, Rol: member.RolProductBuilder, Activo: false}, effort.ErrQuienInvalido},
		{"nadie elegido", member.Integrante{}, effort.ErrQuienInvalido},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if err := effort.PuedeCorregir(c.quien, deCarolina); !errors.Is(err, c.err) {
				t.Errorf("error = %v, se esperaba %v", err, c.err)
			}
		})
	}
}

// contextoDeCorreccion es el de la spec de HU-24 con Juan Pablo (1) de Agile Enabler.
func contextoDeCorreccion() effort.ContextoRegistro {
	c := contextoDeEjemplo()
	c.Integrantes = append([]member.Integrante{{ID: 1, ProyectoID: 7, Nombre: "Juan Pablo", Rol: member.RolAgileEnabler, Activo: true}}, c.Integrantes...)
	return c
}

func registroDeCarolina() effort.Registro {
	return effort.Registro{ID: 7, IntegranteID: 2, HistoriaID: 9, Fecha: dia(17), Actividad: "Pantalla", Horas: 20}
}

// RN5 · se cambian la tarea, la fecha, la actividad y las horas; el ID, el integrante y la historia no cambian
func TestCorregir_DatosValidos(t *testing.T) {
	d := effort.DatosRegistro{TareaID: 5, Fecha: "2026-10-18", Actividad: " Pantalla de login ", Horas: "3,5"}

	r, err := effort.Corregir(registroDeCarolina(), 2, d, contextoDeCorreccion())
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	esperado := effort.Registro{ID: 7, IntegranteID: 2, HistoriaID: 9, TareaID: 5, Fecha: dia(18), Actividad: "Pantalla de login", Horas: 3.5}
	if r != esperado {
		t.Errorf("registro = %+v, se esperaba %+v", r, esperado)
	}
}

// RN6 · para las 24 h del día no se cuenta el mismo registro que se corrige
func TestCorregir_NoCuentaDosVecesElMismoRegistro(t *testing.T) {
	c := contextoDeCorreccion()
	c.DelIntegrante = []effort.Registro{registroDeCarolina()}
	d := effort.DatosRegistro{Fecha: "2026-10-17", Actividad: "Pantalla", Horas: "24"}

	if _, err := effort.Corregir(registroDeCarolina(), 2, d, c); err != nil {
		t.Errorf("no se esperaba error: %v", err)
	}

	otro := effort.Registro{ID: 8, IntegranteID: 2, HistoriaID: 9, Fecha: dia(17), Actividad: "Otra", Horas: 1}
	c.DelIntegrante = append(c.DelIntegrante, otro)
	_, err := effort.Corregir(registroDeCarolina(), 2, d, c)
	if errs := erroresDe(t, err); !errors.Is(errs[effort.CampoHoras], effort.ErrMasDe24HorasEnElDia) {
		t.Errorf("errores = %v, se esperaba %q en horas", errs, effort.ErrMasDe24HorasEnElDia)
	}
}

// CA-26.2 y RN6 · cada error en su campo, como en HU-24, y quién corrige en el campo "quien"
func TestCorregir_Validaciones(t *testing.T) {
	valido := effort.DatosRegistro{Fecha: "2026-10-18", Actividad: "Pantalla", Horas: "3"}
	casos := []struct {
		nombre string
		quien  int64
		cambio func(*effort.DatosRegistro)
		campo  string
		err    error
	}{
		{"el Agile Enabler corrige el de otro", 1, func(*effort.DatosRegistro) {}, "", nil},
		{"otro integrante", 3, func(*effort.DatosRegistro) {}, effort.CampoQuien, effort.ErrNoPuedeCorregir},
		{"nadie elegido", 0, func(*effort.DatosRegistro) {}, effort.CampoQuien, effort.ErrQuienInvalido},
		{"una integrante dada de baja", 4, func(*effort.DatosRegistro) {}, effort.CampoQuien, effort.ErrQuienInvalido},
		{"alguien de otro proyecto", 99, func(*effort.DatosRegistro) {}, effort.CampoQuien, effort.ErrQuienInvalido},
		{"25 h", 2, func(d *effort.DatosRegistro) { d.Horas = "25" }, effort.CampoHoras, effort.ErrHorasTrabajadasInvalidas},
		{"fecha futura", 2, func(d *effort.DatosRegistro) { d.Fecha = "2026-10-19" }, effort.CampoFecha, effort.ErrFechaFutura},
		{"fecha fuera del proyecto", 2, func(d *effort.DatosRegistro) { d.Fecha = "2026-10-04" }, effort.CampoFecha, effort.ErrFechaFueraDelProyecto},
		{"tarea de otra historia", 2, func(d *effort.DatosRegistro) { d.TareaID = 99 }, effort.CampoTarea, effort.ErrTareaDeOtraHistoria},
		{"sin actividad", 2, func(d *effort.DatosRegistro) { d.Actividad = "  " }, effort.CampoActividad, effort.ErrActividadObligatoria},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d := valido
			c.cambio(&d)
			_, err := effort.Corregir(registroDeCarolina(), c.quien, d, contextoDeCorreccion())
			if c.err == nil {
				if err != nil {
					t.Fatalf("no se esperaba error: %v", err)
				}
				return
			}
			if errs := erroresDe(t, err); !errors.Is(errs[c.campo], c.err) {
				t.Errorf("errores = %v, se esperaba %q en %q", errs, c.err, c.campo)
			}
		})
	}
}

func erroresDe(t *testing.T, err error) domain.ErroresValidacion {
	t.Helper()
	var errs domain.ErroresValidacion
	if !errors.As(err, &errs) {
		t.Fatalf("error = %v, se esperaba domain.ErroresValidacion", err)
	}
	return errs
}
