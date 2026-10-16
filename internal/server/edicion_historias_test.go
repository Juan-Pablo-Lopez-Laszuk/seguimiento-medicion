package server_test

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// aplicacionDeHistorias es la aplicación armada con sus repositorios en memoria a la vista, para dejar una historia
// en un estado que todavía no se puede alcanzar desde las pantallas y para ver qué quedó guardado.
type aplicacionDeHistorias struct {
	http.Handler
	historias *memory.Historias
	criterios *memory.Criterios
	tareas    *memory.Tareas
}

// conHistoriaParaEditar arma la aplicación con el proyecto de ejemplo (ID 1) y su HU-1 "Iniciar sesión con Google"
// (Alta, 5 story points) ya creados. envolver, si no es nil, cambia el repositorio que usa el caso de uso de editar y
// eliminar, para simular una base que rechaza el borrado.
func conHistoriaParaEditar(t *testing.T, envolver func(*memory.Historias) service.RepositorioEdicionHistorias) aplicacionDeHistorias {
	t.Helper()
	ctx := context.Background()
	proyectos := memory.NuevoProyectos()
	criterios, tareas := memory.NuevoCriterios(), memory.NuevoTareas()
	historias := memory.NuevoHistorias(criterios, tareas)
	servicioProyectos := service.NuevoProyectos(proyectos, time.Now)
	if _, err := servicioProyectos.Crear(ctx, project.Datos{
		Nombre:      "Software Metrics",
		FechaInicio: time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC),
		FechaFin:    time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("no se pudo crear el proyecto: %v", err)
	}
	repoEdicion := service.RepositorioEdicionHistorias(historias)
	if envolver != nil {
		repoEdicion = envolver(historias)
	}
	h := server.NewRouter(server.Dependencias{
		Proyectos:        servicioProyectos,
		Historias:        service.NuevoHistorias(proyectos, historias),
		Criterios:        service.NuevoCriterios(proyectos, historias, criterios),
		EdicionHistorias: service.NuevoEdicionHistorias(proyectos, repoEdicion, criterios, tareas),
	})
	if rec := post(h, "/proyectos/1/backlog", historiaValida()); rec.Code != http.StatusSeeOther {
		t.Fatalf("no se pudo crear la historia de ejemplo: código %d", rec.Code)
	}
	return aplicacionDeHistorias{Handler: h, historias: historias, criterios: criterios, tareas: tareas}
}

func edicionValida() url.Values {
	return url.Values{
		"titulo":      {"Iniciar sesión con Google o GitHub"},
		"descripcion": {"Con OAuth."},
		"prioridad":   {"Baja"},
	}
}

// CA-06.1 y CA-06.3 · la página de la historia ofrece editarla y eliminarla
func TestPaginaDeLaHistoria_OfreceEditarYEliminar(t *testing.T) {
	app := conHistoriaParaEditar(t, nil)

	verificarContiene(t, get(app, "/proyectos/1/backlog/1").Body.String(),
		`href="/proyectos/1/backlog/1/editar"`, `href="/proyectos/1/backlog/1/eliminar"`)
}

// CA-06.1 · el formulario de editar trae los datos actuales y no pide los story points ni el estado
func TestFormularioDeEditar_MuestraLosDatosActuales(t *testing.T) {
	app := conHistoriaParaEditar(t, nil)

	rec := get(app, "/proyectos/1/backlog/1/editar")

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	html := rec.Body.String()
	verificarContiene(t, html,
		"<title>Editar HU-1 · Software Metrics</title>",
		`action="/proyectos/1/backlog/1/editar"`, `method="post"`,
		`value="Iniciar sesión con Google"`, "Para no tener que recordar otra contraseña.",
		`<option value="Alta" selected>`,
		"HU-1 no cambia", `href="/proyectos/1/backlog/1"`, // el identificador y Cancelar
	)
	if strings.Contains(html, `name="story_points"`) || strings.Contains(html, `name="estado"`) {
		t.Error("el formulario de editar no tiene que pedir story points ni estado")
	}
}

// CA-06.1 · después de editar se vuelve a la página con los datos nuevos, el mismo identificador y lo demás igual
func TestEditarHistoria_Valido_VuelveALaPaginaConLosDatosNuevos(t *testing.T) {
	app := conHistoriaParaEditar(t, nil)

	rec := post(app, "/proyectos/1/backlog/1/editar", edicionValida())

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("código = %d, se esperaba %d (redirección)", rec.Code, http.StatusSeeOther)
	}
	if destino := rec.Header().Get("Location"); destino != "/proyectos/1/backlog/1?editada=1" {
		t.Fatalf("redirige a %q, se esperaba /proyectos/1/backlog/1?editada=1", destino)
	}
	html := get(app, "/proyectos/1/backlog/1?editada=1").Body.String()
	verificarContiene(t, html, "Historia actualizada.", "HU-1 · Iniciar sesión con Google o GitHub", "Con OAuth.", "Baja",
		"Story points: 5", "Pendiente")
}

// RN1 · con datos inválidos se vuelve al formulario con el error debajo de cada campo y lo que se escribió; no cambia nada
func TestEditarHistoria_DatosInvalidos_MuestraLosErroresYNoCambiaNada(t *testing.T) {
	app := conHistoriaParaEditar(t, nil)
	malos := url.Values{"titulo": {""}, "descripcion": {"Una descripción escrita"}, "prioridad": {"Urgente"}}

	rec := post(app, "/proyectos/1/backlog/1/editar", malos)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusUnprocessableEntity)
	}
	verificarContiene(t, rec.Body.String(),
		"el título es obligatorio", "la prioridad tiene que ser Alta, Media o Baja",
		"Una descripción escrita", `action="/proyectos/1/backlog/1/editar"`)
	pagina := get(app, "/proyectos/1/backlog/1").Body.String()
	verificarContiene(t, pagina, "Iniciar sesión con Google", "Para no tener que recordar otra contraseña.", "Alta")
}

// RN3 · se puede editar una historia Hecha
func TestEditarHistoria_HistoriaHecha_Vale(t *testing.T) {
	app := conHistoriaParaEditar(t, nil)
	if err := app.historias.ActualizarEstado(context.Background(), 1, backlog.EstadoHecho); err != nil {
		t.Fatalf("no se pudo dejar la historia Hecha: %v", err)
	}

	rec := post(app, "/proyectos/1/backlog/1/editar", edicionValida())

	if rec.Code != http.StatusSeeOther {
		t.Errorf("código = %d, se esperaba %d", rec.Code, http.StatusSeeOther)
	}
}

// Sección 7 · proyecto o historia que no existen
func TestEditarHistoria_ProyectoOHistoriaInexistente_Responde404(t *testing.T) {
	app := conHistoriaParaEditar(t, nil)

	for _, ruta := range []string{"/proyectos/99/backlog/1/editar", "/proyectos/1/backlog/99/editar", "/proyectos/1/backlog/abc/editar"} {
		if rec := get(app, ruta); rec.Code != http.StatusNotFound {
			t.Errorf("GET %s: código = %d, se esperaba %d", ruta, rec.Code, http.StatusNotFound)
		}
		if rec := post(app, ruta, edicionValida()); rec.Code != http.StatusNotFound {
			t.Errorf("POST %s: código = %d, se esperaba %d", ruta, rec.Code, http.StatusNotFound)
		}
	}
}
