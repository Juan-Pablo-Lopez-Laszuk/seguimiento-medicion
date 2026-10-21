package server_test

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// conRegistros arma la aplicación con el proyecto 1 (del 05/10 al 01/11/2026, hoy 21/10), el Sprint 1 (05/10 al
// 18/10), la HU-1 "Login" con una tarea de 6 h, Juan Pablo (1) de Agile Enabler, Carolina (2) y Mariano (3), y los
// registros "Pantalla" (1: Carolina, 10/10, 3 h) y "Validaciones" (2: Mariano, 12/10, 2 h).
func conRegistros(t *testing.T) http.Handler {
	t.Helper()
	ctx := context.Background()
	ahora := func() time.Time { return time.Date(2026, 10, 21, 18, 0, 0, 0, time.UTC) }
	proyectos, sprints, historias := memory.NuevoProyectos(), memory.NuevoSprints(), memory.NuevoHistorias()
	integrantes, tareas, registros := memory.NuevoIntegrantes(), memory.NuevoTareas(), memory.NuevoRegistros()
	casosProyectos := service.NuevoProyectos(proyectos, ahora)
	casosSprints := service.NuevoSprints(proyectos, sprints)
	casosIntegrantes := service.NuevoIntegrantes(proyectos, integrantes)
	casosHistorias := service.NuevoHistorias(proyectos, historias)
	casosTareas := service.NuevoTareas(proyectos, historias, integrantes, tareas)
	esfuerzo := service.NuevoEsfuerzo(proyectos, historias, integrantes, tareas, registros, ahora)

	p, err := casosProyectos.Crear(ctx, project.Datos{Nombre: "Software Metrics", FechaInicio: time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC), FechaFin: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("crear el proyecto: %v", err)
	}
	if _, err := casosSprints.Crear(ctx, p.ID, sprint.Datos{Objetivo: "Uno", FechaInicio: time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC), FechaFin: time.Date(2026, 10, 18, 0, 0, 0, 0, time.UTC)}); err != nil {
		t.Fatalf("crear el sprint: %v", err)
	}
	for _, d := range []member.Datos{
		{Nombre: "Juan Pablo", Email: "juan@mail.com", Rol: member.RolAgileEnabler},
		{Nombre: "Carolina", Email: "caro@mail.com", Rol: member.RolProductBuilder},
		{Nombre: "Mariano", Email: "mariano@mail.com", Rol: member.RolProductBuilder},
	} {
		if _, err := casosIntegrantes.Registrar(ctx, p.ID, d); err != nil {
			t.Fatalf("registrar a %s: %v", d.Nombre, err)
		}
	}
	h, err := casosHistorias.Crear(ctx, p.ID, backlog.Datos{Titulo: "Login", Prioridad: backlog.PrioridadAlta})
	if err != nil {
		t.Fatalf("crear la historia: %v", err)
	}
	if _, err := casosTareas.Agregar(ctx, p.ID, h.Numero, effort.DatosTarea{Titulo: "Armar el login", ResponsableID: 2, Horas: "6"}); err != nil {
		t.Fatalf("agregar la tarea: %v", err)
	}
	for _, d := range []effort.DatosRegistro{
		{IntegranteID: 2, Fecha: "2026-10-10", Actividad: "Pantalla", Horas: "3"},
		{IntegranteID: 3, Fecha: "2026-10-12", Actividad: "Validaciones", Horas: "2"},
	} {
		if _, err := esfuerzo.Registrar(ctx, p.ID, h.Numero, d); err != nil {
			t.Fatalf("registrar %q: %v", d.Actividad, err)
		}
	}
	return server.NewRouter(server.Dependencias{
		Proyectos: casosProyectos, Sprints: casosSprints, Integrantes: casosIntegrantes, Historias: casosHistorias,
		Tareas: casosTareas, Esfuerzo: esfuerzo,
		RegistrosEsfuerzo: service.NuevoRegistrosEsfuerzo(proyectos, sprints, historias, integrantes, tareas, registros,
			service.NuevoHoras(proyectos, sprints, historias, tareas, registros), ahora),
	})
}

// A la página de registros de horas se llega desde la vista del proyecto.
func TestProyecto_EnlazaASusRegistrosDeHoras(t *testing.T) {
	verificarContiene(t, get(conProyecto(t), "/proyectos/1").Body.String(), `href="/proyectos/1/esfuerzo"`, "Registros de horas")
}

// HU-26 · CA-26.1 y CA-26.3 · los filtros, la lista con el enlace Corregir, el total y las horas por historia
func TestPaginaDeRegistros_FiltrosListaYHorasPorHistoria(t *testing.T) {
	rec := get(conRegistros(t), "/proyectos/1/esfuerzo?integrante=2&historia=1&sprint=1")

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	verificarContiene(t, rec.Body.String(),
		"<title>Registros de horas · Software Metrics</title>",
		`<option value="2" selected>Carolina</option>`, `<option value="1" selected>HU-1 · Login</option>`, `<option value="1" selected>Sprint 1</option>`,
		"10/10/2026", "Pantalla", "Toda la historia", `href="/proyectos/1/esfuerzo/1/corregir"`,
		"Total de la lista: <strong>3 h</strong>", `href="/proyectos/1/esfuerzo"`, // quitar filtros
		"Horas estimadas vs reales por historia", "6 h", "5 h",
	)
	if strings.Contains(rec.Body.String(), "Validaciones") {
		t.Error("la lista filtrada no debería tener el registro de Mariano")
	}
}

// HU-26 · si nada coincide se avisa y se ofrece quitar los filtros
func TestPaginaDeRegistros_NingunoCumpleLosFiltros(t *testing.T) {
	verificarContiene(t, get(conRegistros(t), "/proyectos/1/esfuerzo?integrante=1").Body.String(),
		"Ningún registro con esos filtros.", `href="/proyectos/1/esfuerzo"`)
}

// HU-26 · CA-26.2 · el formulario de corrección trae los datos del registro y quién puede corregir
func TestCorregirRegistro_Formulario(t *testing.T) {
	rec := get(conRegistros(t), "/proyectos/1/esfuerzo/2/corregir")

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	verificarContiene(t, rec.Body.String(),
		"Corregir registro de Mariano", "HU-1 · Login", `name="quien"`, "Juan Pablo", `value="2026-10-12"`,
		`value="Validaciones"`, `value="2"`, "Armar el login", `max="2026-10-21"`,
		"Cada integrante corrige solo sus registros; el Agile Enabler, todos.",
	)
}

// HU-26 · CA-26.2 · una corrección válida vuelve a la lista con el mensaje
func TestCorregirRegistro_Valida_VuelveALaLista(t *testing.T) {
	h := conRegistros(t)
	campos := url.Values{"quien": {"3"}, "tarea": {""}, "fecha": {"2026-10-12"}, "actividad": {"Validaciones"}, "horas": {"1,5"}}

	rec := post(h, "/proyectos/1/esfuerzo/2/corregir", campos)

	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/proyectos/1/esfuerzo?corregido=1" {
		t.Fatalf("código = %d y Location = %q", rec.Code, rec.Header().Get("Location"))
	}
	verificarContiene(t, get(h, "/proyectos/1/esfuerzo?corregido=1").Body.String(), "Registro corregido.", "1,5 h")
}

// HU-26 · CA-26.2 · Carolina no corrige el registro de Mariano: el error queda debajo de "Quién corrige"
func TestCorregirRegistro_DeOtro_MuestraElError(t *testing.T) {
	campos := url.Values{"quien": {"2"}, "tarea": {""}, "fecha": {"2026-10-12"}, "actividad": {"Validaciones"}, "horas": {"1"}}

	rec := post(conRegistros(t), "/proyectos/1/esfuerzo/2/corregir", campos)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusUnprocessableEntity)
	}
	verificarContiene(t, rec.Body.String(), "cada integrante corrige solo sus registros (el Agile Enabler, todos)", `value="1"`)
}

func TestRegistros_NoEncontrados_Responden404(t *testing.T) {
	h := conRegistros(t)
	for _, ruta := range []string{"/proyectos/99/esfuerzo", "/proyectos/1/esfuerzo/99/corregir", "/proyectos/1/esfuerzo/abc/corregir", "/proyectos/99/esfuerzo/1/corregir"} {
		if rec := get(h, ruta); rec.Code != http.StatusNotFound {
			t.Errorf("%s: código = %d, se esperaba %d", ruta, rec.Code, http.StatusNotFound)
		}
	}
	if rec := post(h, "/proyectos/1/esfuerzo/99/corregir", url.Values{"quien": {"1"}}); rec.Code != http.StatusNotFound {
		t.Errorf("POST de un registro que no existe: código = %d, se esperaba %d", rec.Code, http.StatusNotFound)
	}
}
