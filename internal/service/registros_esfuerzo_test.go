package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// el21deOctubre son las 15:00 de Argentina del 21/10/2026: "hoy" de los escenarios de HU-26.
func el21deOctubre() time.Time { return time.Date(2026, 10, 21, 18, 0, 0, 0, time.UTC) }

func octubre(d int) time.Time { return time.Date(2026, 10, d, 0, 0, 0, 0, time.UTC) }

// escenarioRegistros es el de features/hu-26-registros-esfuerzo.feature: Sprint 1 (05/10 al 18/10) y Sprint 2
// (19/10 al 01/11), HU-1 con 6 h estimadas y HU-2 con 4 h, Juan Pablo de Agile Enabler, Carolina y Mariano, y los
// registros "Pantalla" (Carolina, HU-1, 10/10, 3 h), "Validaciones" (Mariano, HU-1, 12/10, 2 h) y "Primer reporte"
// (Carolina, HU-2, 20/10, 4 h).
type escenarioRegistros struct {
	casos                   *service.RegistrosEsfuerzo
	proyectoID              int64
	hu1, hu2                backlog.Historia
	juan, carolina, mariano member.Integrante
	pantalla, validaciones  effort.Registro
	primerReporte           effort.Registro
	tareaHU1                effort.Tarea
}

func nuevoEscenarioRegistros(t *testing.T) escenarioRegistros {
	t.Helper()
	ctx := context.Background()
	proyectos, sprints, historias := memory.NuevoProyectos(), memory.NuevoSprints(), memory.NuevoHistorias()
	integrantes, tareas, registros := memory.NuevoIntegrantes(), memory.NuevoTareas(), memory.NuevoRegistros()
	p, err := service.NuevoProyectos(proyectos, relojFijo).Crear(ctx, datosValidos())
	if err != nil {
		t.Fatalf("crear el proyecto: %v", err)
	}
	e := escenarioRegistros{proyectoID: p.ID}
	casosSprints := service.NuevoSprints(proyectos, sprints)
	for _, d := range []sprint.Datos{{Objetivo: "Uno", FechaInicio: octubre(5), FechaFin: octubre(18)}, {Objetivo: "Dos", FechaInicio: octubre(19), FechaFin: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)}} {
		if _, err := casosSprints.Crear(ctx, p.ID, d); err != nil {
			t.Fatalf("crear el sprint %q: %v", d.Objetivo, err)
		}
	}
	casosHistorias := service.NuevoHistorias(proyectos, historias)
	if e.hu1, err = casosHistorias.Crear(ctx, p.ID, datosDeHistoria("Login")); err != nil {
		t.Fatalf("crear la HU-1: %v", err)
	}
	if e.hu2, err = casosHistorias.Crear(ctx, p.ID, datosDeHistoria("Reportes")); err != nil {
		t.Fatalf("crear la HU-2: %v", err)
	}
	casosIntegrantes := service.NuevoIntegrantes(proyectos, integrantes)
	registrar := func(nombre string, rol member.Rol) member.Integrante {
		i, err := casosIntegrantes.Registrar(ctx, p.ID, member.Datos{Nombre: nombre, Email: strings.ReplaceAll(strings.ToLower(nombre), " ", "") + "@mail.com", Rol: rol})
		if err != nil {
			t.Fatalf("registrar a %s: %v", nombre, err)
		}
		return i
	}
	e.juan, e.carolina, e.mariano = registrar("Juan Pablo", member.RolAgileEnabler), registrar("Carolina", member.RolProductBuilder), registrar("Mariano", member.RolProductBuilder)
	casosTareas := service.NuevoTareas(proyectos, historias, integrantes, tareas)
	if e.tareaHU1, err = casosTareas.Agregar(ctx, p.ID, e.hu1.Numero, effort.DatosTarea{Titulo: "Armar el login", ResponsableID: e.carolina.ID, Horas: "6"}); err != nil {
		t.Fatalf("agregar la tarea de la HU-1: %v", err)
	}
	if _, err = casosTareas.Agregar(ctx, p.ID, e.hu2.Numero, effort.DatosTarea{Titulo: "Armar el reporte", ResponsableID: e.carolina.ID, Horas: "4"}); err != nil {
		t.Fatalf("agregar la tarea de la HU-2: %v", err)
	}
	esfuerzo := service.NuevoEsfuerzo(proyectos, historias, integrantes, tareas, registros, el21deOctubre)
	cargar := func(i member.Integrante, h backlog.Historia, fecha, actividad, horas string) effort.Registro {
		r, err := esfuerzo.Registrar(ctx, p.ID, h.Numero, effort.DatosRegistro{IntegranteID: i.ID, Fecha: fecha, Actividad: actividad, Horas: horas})
		if err != nil {
			t.Fatalf("registrar %q: %v", actividad, err)
		}
		return r
	}
	e.pantalla = cargar(e.carolina, e.hu1, "2026-10-10", "Pantalla", "3")
	e.validaciones = cargar(e.mariano, e.hu1, "2026-10-12", "Validaciones", "2")
	e.primerReporte = cargar(e.carolina, e.hu2, "2026-10-20", "Primer reporte", "4")
	horas := service.NuevoHoras(proyectos, sprints, historias, tareas, registros)
	e.casos = service.NuevoRegistrosEsfuerzo(proyectos, sprints, historias, integrantes, tareas, registros, horas, el21deOctubre)
	return e
}

func actividades(filas []service.FilaRegistro) string {
	nombres := make([]string, len(filas))
	for i, f := range filas {
		nombres[i] = f.Actividad
	}
	return strings.Join(nombres, ", ")
}

// CA-26.1 · RN1 a RN3 · los filtros se combinan, el sprint va por fechas y lo que no existe se ignora
func TestConsultarRegistros_Filtros(t *testing.T) {
	e := nuevoEscenarioRegistros(t)
	casos := []struct {
		nombre    string
		filtro    service.FiltroRegistros
		esperadas string
		total     float64
		aplicado  service.FiltroRegistros
	}{
		{"sin filtros", service.FiltroRegistros{}, "Primer reporte, Validaciones, Pantalla", 9, service.FiltroRegistros{}},
		{"integrante y sprint", service.FiltroRegistros{IntegranteID: e.carolina.ID, Sprint: 1}, "Pantalla", 3, service.FiltroRegistros{IntegranteID: e.carolina.ID, Sprint: 1}},
		{"historia", service.FiltroRegistros{Historia: 1}, "Validaciones, Pantalla", 5, service.FiltroRegistros{Historia: 1}},
		{"sprint 2", service.FiltroRegistros{Sprint: 2}, "Primer reporte", 4, service.FiltroRegistros{Sprint: 2}},
		{"nada coincide", service.FiltroRegistros{IntegranteID: e.juan.ID}, "", 0, service.FiltroRegistros{IntegranteID: e.juan.ID}},
		{"valores que no existen", service.FiltroRegistros{IntegranteID: 99, Historia: 9, Sprint: 9}, "Primer reporte, Validaciones, Pantalla", 9, service.FiltroRegistros{}},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			v, err := e.casos.Consultar(context.Background(), e.proyectoID, c.filtro)
			if err != nil {
				t.Fatalf("no se esperaba error: %v", err)
			}
			if got := actividades(v.Registros); got != c.esperadas {
				t.Errorf("registros = %q, se esperaba %q", got, c.esperadas)
			}
			if v.Total != c.total || v.Filtro != c.aplicado {
				t.Errorf("total = %v y filtro = %+v, se esperaba %v y %+v", v.Total, v.Filtro, c.total, c.aplicado)
			}
			if len(v.Integrantes) != 3 || len(v.Historias) != 2 || len(v.Sprints) != 2 {
				t.Errorf("las opciones de los filtros: %d integrantes, %d historias y %d sprints", len(v.Integrantes), len(v.Historias), len(v.Sprints))
			}
		})
	}
}

// CA-26.3 · cada fila dice quién, en qué historia y tarea; y hay una fila de horas por historia de la lista
func TestConsultarRegistros_FilasYHorasPorHistoria(t *testing.T) {
	e := nuevoEscenarioRegistros(t)
	ctx := context.Background()
	correccion := effort.DatosRegistro{TareaID: e.tareaHU1.ID, Fecha: "2026-10-10", Actividad: "Pantalla", Horas: "3"}
	if _, err := e.casos.Corregir(ctx, e.proyectoID, e.pantalla.ID, e.carolina.ID, correccion); err != nil {
		t.Fatalf("poner la tarea al registro: %v", err)
	}

	v, err := e.casos.Consultar(ctx, e.proyectoID, service.FiltroRegistros{Historia: 1})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	pantalla := v.Registros[1]
	if pantalla.Integrante != "Carolina" || pantalla.Historia != "HU-1 · Login" || pantalla.Tarea != "Armar el login" || v.Registros[0].Tarea != "" {
		t.Errorf("filas = %+v", v.Registros)
	}
	if len(v.PorHistoria) != 1 || v.PorHistoria[0].Nombre != "HU-1 · Login" || v.PorHistoria[0].Estimadas != 6 || v.PorHistoria[0].Reales != 5 || v.PorHistoria[0].Horas != -1 {
		t.Errorf("por historia = %+v, se esperaba HU-1 con 6 h estimadas y 5 h reales", v.PorHistoria)
	}

	todas, _ := e.casos.Consultar(ctx, e.proyectoID, service.FiltroRegistros{})
	if len(todas.PorHistoria) != 2 || todas.PorHistoria[0].Nombre != "HU-1 · Login" || todas.PorHistoria[1].Reales != 4 {
		t.Errorf("por historia sin filtros = %+v, se esperaban HU-1 y HU-2 en el orden del backlog", todas.PorHistoria)
	}
}

func TestConsultarRegistros_ProyectoInexistente(t *testing.T) {
	e := nuevoEscenarioRegistros(t)
	if _, err := e.casos.Consultar(context.Background(), 99, service.FiltroRegistros{}); !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("error = %v, se esperaba project.ErrNoEncontrado", err)
	}
}

// CA-26.2 · la página de corrección trae el registro, su historia, sus tareas y quién puede corregir
func TestVerCorreccion(t *testing.T) {
	e := nuevoEscenarioRegistros(t)
	v, err := e.casos.VerCorreccion(context.Background(), e.proyectoID, e.validaciones.ID)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if v.Registro != e.validaciones || v.Historia.ID != e.hu1.ID || len(v.Tareas) != 1 || len(v.Activos()) != 3 || v.NombreDe(v.Registro.IntegranteID) != "Mariano" {
		t.Errorf("vista = %+v", v)
	}
	if !v.Hoy.Equal(octubre(21)) {
		t.Errorf("hoy = %v, se esperaba el 21/10/2026", v.Hoy)
	}
}

// CA-26.2 · RN4 a RN6 · quién puede corregir y qué queda guardado
func TestCorregirRegistro(t *testing.T) {
	ctx := context.Background()
	casos := []struct {
		nombre string
		quien  func(escenarioRegistros) member.Integrante
		de     func(escenarioRegistros) effort.Registro
		horas  string
		campo  string
		err    error
	}{
		{"su propio registro", func(e escenarioRegistros) member.Integrante { return e.carolina }, func(e escenarioRegistros) effort.Registro { return e.pantalla }, "3,5", "", nil},
		{"el Agile Enabler, el de otro", func(e escenarioRegistros) member.Integrante { return e.juan }, func(e escenarioRegistros) effort.Registro { return e.validaciones }, "1", "", nil},
		{"el registro de otro", func(e escenarioRegistros) member.Integrante { return e.carolina }, func(e escenarioRegistros) effort.Registro { return e.validaciones }, "1", effort.CampoQuien, effort.ErrNoPuedeCorregir},
		{"25 h", func(e escenarioRegistros) member.Integrante { return e.carolina }, func(e escenarioRegistros) effort.Registro { return e.pantalla }, "25", effort.CampoHoras, effort.ErrHorasTrabajadasInvalidas},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			e := nuevoEscenarioRegistros(t)
			r := c.de(e)
			d := effort.DatosRegistro{Fecha: r.Fecha.Format(time.DateOnly), Actividad: r.Actividad, Horas: c.horas}

			corregido, err := e.casos.Corregir(ctx, e.proyectoID, r.ID, c.quien(e).ID, d)

			if c.err != nil {
				var errs domain.ErroresValidacion
				if !errors.As(err, &errs) || !errors.Is(errs[c.campo], c.err) {
					t.Fatalf("error = %v, se esperaba %q en %q", err, c.err, c.campo)
				}
				return
			}
			if err != nil {
				t.Fatalf("no se esperaba error: %v", err)
			}
			v, _ := e.casos.VerCorreccion(ctx, e.proyectoID, r.ID)
			if v.Registro != corregido || corregido.IntegranteID != r.IntegranteID || effort.FormatearHoras(corregido.Horas) != c.horas {
				t.Errorf("guardado = %+v, corregido = %+v", v.Registro, corregido)
			}
		})
	}
}

// Un registro que no existe no se encuentra; tampoco un proyecto que no existe.
func TestCorregirRegistro_NoEncontrado(t *testing.T) {
	e := nuevoEscenarioRegistros(t)
	ctx := context.Background()
	d := effort.DatosRegistro{Fecha: "2026-10-10", Actividad: "Pantalla", Horas: "1"}

	if _, err := e.casos.Corregir(ctx, e.proyectoID, 99, e.carolina.ID, d); !errors.Is(err, effort.ErrRegistroNoEncontrado) {
		t.Errorf("registro inexistente: error = %v", err)
	}
	if _, err := e.casos.VerCorreccion(ctx, 99, e.pantalla.ID); !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("proyecto inexistente: error = %v", err)
	}
}
