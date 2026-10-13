package server_test

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// Los tests de la vista del proyecto usan el reloj de nuevoRouter: hoy es el 13/10/2026 en Argentina.

func conSprint(t *testing.T, inicio, fin string) http.Handler {
	t.Helper()
	h := conProyecto(t)
	campos := url.Values{"objetivo": {"Primer MVP"}, "fecha_inicio": {inicio}, "fecha_fin": {fin}}
	if rec := post(h, "/proyectos/1/sprints", campos); rec.Code != http.StatusSeeOther {
		t.Fatalf("no se pudo crear el sprint: código %d", rec.Code)
	}
	return h
}

// CA-04.3 y sección 3 · la vista muestra el proyecto, su estado, los días y los enlaces a sus secciones
func TestVistaDelProyecto_SinSprints(t *testing.T) {
	rec := get(conProyecto(t), "/proyectos/1")

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	verificarContiene(t, rec.Body.String(),
		"<title>Software Metrics · Software Metrics</title>",
		"TPI de ICS", "05/10/2026", "01/11/2026",
		"Planificado",
		"Sin sprint activo",
		"19 días", // del 13/10 al 01/11
		"Sin historias estimadas",
		`href="/proyectos/1/sprints"`, `href="/proyectos/1/integrantes"`, `href="/proyectos/1/backlog"`,
	)
}

// CA-04.1 · un sprint vencido sin cerrar sigue activo y se avisa que falta cerrarlo
func TestVistaDelProyecto_SprintVencidoSinCerrar(t *testing.T) {
	rec := get(conSprint(t, "2026-10-05", "2026-10-11"), "/proyectos/1")

	verificarContiene(t, rec.Body.String(),
		"En curso",
		"Sprint 1", "Primer MVP",
		"Terminó el 11/10/2026: falta cerrarlo",
	)
}

// CA-04.1 · el sprint activo con los días que le quedan
func TestVistaDelProyecto_SprintActivo(t *testing.T) {
	rec := get(conSprint(t, "2026-10-12", "2026-10-18"), "/proyectos/1")

	verificarContiene(t, rec.Body.String(), "En curso", "Sprint 1", "Le quedan 5 días")
	if strings.Contains(rec.Body.String(), "falta cerrarlo") {
		t.Error("un sprint en fecha no tiene que mostrar el aviso de cierre")
	}
}

// Sección 3 · sin sprint activo se muestra el próximo y cuándo empieza
func TestVistaDelProyecto_SinSprintActivo_MuestraElProximo(t *testing.T) {
	rec := get(conSprint(t, "2026-10-20", "2026-10-26"), "/proyectos/1")

	verificarContiene(t, rec.Body.String(), "Planificado", "Sin sprint activo", "Sprint 1", "empieza el 20/10/2026")
}

// RN4 · porcentaje de SP y aviso de historias sin estimar
func TestVistaDelProyecto_PorcentajeDeStoryPoints(t *testing.T) {
	h := conProyecto(t)
	sinEstimar := historiaValida()
	sinEstimar.Set("story_points", "")
	for _, campos := range []url.Values{historiaValida(), sinEstimar} {
		if rec := post(h, "/proyectos/1/backlog", campos); rec.Code != http.StatusSeeOther {
			t.Fatalf("no se pudo crear la historia: código %d", rec.Code)
		}
	}

	verificarContiene(t, get(h, "/proyectos/1").Body.String(), "0,0 %", "hay 1 historia sin estimar")
}

// Sección 7 · un proyecto que no existe (o un ID que no es un número) da 404
func TestVistaDelProyecto_Inexistente(t *testing.T) {
	h := conProyecto(t)
	for _, ruta := range []string{"/proyectos/999", "/proyectos/abc"} {
		if rec := get(h, ruta); rec.Code != http.StatusNotFound {
			t.Errorf("%s: código = %d, se esperaba %d", ruta, rec.Code, http.StatusNotFound)
		}
	}
}

// Sección 7 · si falla la base se muestra un mensaje general, sin detalles internos
func TestVistaDelProyecto_ErrorDeLaBase(t *testing.T) {
	proyectos := memory.NuevoProyectos()
	if _, err := proyectos.Guardar(context.Background(), project.Proyecto{Nombre: "Software Metrics"}); err != nil {
		t.Fatalf("no se pudo guardar el proyecto: %v", err)
	}
	h := server.NewRouter(server.Dependencias{
		Estado: service.NuevoEstadoProyectos(proyectos, sprintsCaidos{}, memory.NuevoHistorias(), relojDeLaVista),
	})

	rec := get(h, "/proyectos/1")

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("código = %d, se esperaba %d", rec.Code, http.StatusInternalServerError)
	}
	if strings.Contains(rec.Body.String(), "detalle interno") {
		t.Error("la página muestra el error interno")
	}
	verificarContiene(t, rec.Body.String(), "No se pudo cargar el proyecto")
}

// Sección 3 · en la lista, el nombre es el enlace a la vista del proyecto
func TestListaDeProyectos_ElNombreEnlazaALaVista(t *testing.T) {
	verificarContiene(t, get(conProyecto(t), "/proyectos").Body.String(), `href="/proyectos/1"`)
}

// RN1 · la lista muestra el estado calculado, no el guardado
func TestListaDeProyectos_MuestraElEstadoCalculado(t *testing.T) {
	verificarContiene(t, get(conSprint(t, "2026-10-05", "2026-10-11"), "/proyectos").Body.String(), "En curso")
}
