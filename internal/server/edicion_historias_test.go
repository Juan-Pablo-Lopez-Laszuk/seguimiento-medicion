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

// conCriteriosYTareas deja la HU-1 con n criterios (por la pantalla) y m tareas (directo en el repositorio).
func (a aplicacionDeHistorias) conCriteriosYTareas(t *testing.T, criterios, tareas int) {
	t.Helper()
	for i := 0; i < criterios; i++ {
		if rec := post(a, "/proyectos/1/backlog/1/criterios", criterioValido("Un criterio")); rec.Code != http.StatusSeeOther {
			t.Fatalf("no se pudo agregar un criterio: código %d", rec.Code)
		}
	}
	for i := 0; i < tareas; i++ {
		if _, err := a.tareas.Guardar(context.Background(), effort.Tarea{HistoriaID: 1, Titulo: "Una tarea"}); err != nil {
			t.Fatalf("no se pudo guardar una tarea: %v", err)
		}
	}
}

// CA-06.3 y RN8 · la primera pantalla pide confirmar, dice qué se pierde y no elimina nada
func TestPantallaDeEliminar_PideConfirmacionYDiceQueSePierde(t *testing.T) {
	app := conHistoriaParaEditar(t, nil)
	app.conCriteriosYTareas(t, 2, 1)

	rec := get(app, "/proyectos/1/backlog/1/eliminar")

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	verificarContiene(t, rec.Body.String(),
		"<title>Eliminar HU-1 · Software Metrics</title>",
		"¿Eliminar HU-1", "Iniciar sesión con Google",
		"2 criterios de aceptación", "1 tarea", "no se puede deshacer",
		`action="/proyectos/1/backlog/1/eliminar"`, `method="post"`, "Eliminar historia",
		`href="/proyectos/1/backlog/1"`, // Cancelar vuelve a la historia
	)
	if rec := get(app, "/proyectos/1/backlog/1"); rec.Code != http.StatusOK {
		t.Errorf("mostrar la pantalla no tenía que eliminar la historia: código %d", rec.Code)
	}
	if cs, _ := app.criterios.ListarPorHistoria(context.Background(), 1); len(cs) != 2 {
		t.Errorf("los criterios tenían que seguir (2) y hay %d", len(cs))
	}
}

// la cantidad se dice en singular o en plural según corresponda
func TestPantallaDeEliminar_DiceLaCantidadEnSingularOEnPlural(t *testing.T) {
	uno := conHistoriaParaEditar(t, nil)
	uno.conCriteriosYTareas(t, 1, 1)
	ninguno := conHistoriaParaEditar(t, nil)

	htmlUno := get(uno, "/proyectos/1/backlog/1/eliminar").Body.String()
	htmlNinguno := get(ninguno, "/proyectos/1/backlog/1/eliminar").Body.String()

	verificarContiene(t, htmlUno, "1 criterio de aceptación", "1 tarea")
	verificarContiene(t, htmlNinguno, "0 criterios de aceptación", "0 tareas")
	if strings.Contains(htmlUno, "1 criterios") || strings.Contains(htmlUno, "1 tareas") {
		t.Error("con uno tiene que decir criterio y tarea, en singular")
	}
}

// CA-06.2 · si no se puede eliminar, la pantalla explica por qué y no ofrece eliminar
func TestPantallaDeEliminar_HistoriaHecha_ExplicaPorQueNoSePuede(t *testing.T) {
	app := conHistoriaParaEditar(t, nil)
	if err := app.historias.ActualizarEstado(context.Background(), 1, backlog.EstadoHecho); err != nil {
		t.Fatalf("no se pudo dejar la historia Hecha: %v", err)
	}

	rec := get(app, "/proyectos/1/backlog/1/eliminar")

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	html := rec.Body.String()
	verificarContiene(t, html, "la historia está Hecha: no se puede eliminar", `href="/proyectos/1/backlog/1"`)
	if strings.Contains(html, `action="/proyectos/1/backlog/1/eliminar"`) || strings.Contains(html, "Eliminar historia") {
		t.Error("la pantalla no tiene que ofrecer eliminar una historia Hecha")
	}
}

// CA-06.3 · al confirmar se elimina la historia con sus criterios y sus tareas, y se vuelve al backlog con el aviso
func TestEliminarHistoria_Confirmado_EliminaYVuelveAlBacklog(t *testing.T) {
	app := conHistoriaParaEditar(t, nil)
	app.conCriteriosYTareas(t, 2, 1)

	rec := post(app, "/proyectos/1/backlog/1/eliminar", nil)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("código = %d, se esperaba %d (redirección)", rec.Code, http.StatusSeeOther)
	}
	if destino := rec.Header().Get("Location"); destino != "/proyectos/1/backlog?eliminada=1" {
		t.Fatalf("redirige a %q, se esperaba /proyectos/1/backlog?eliminada=1", destino)
	}
	lista := get(app, "/proyectos/1/backlog?eliminada=1").Body.String()
	verificarContiene(t, lista, "Historia HU-1 eliminada.")
	if strings.Contains(lista, "Iniciar sesión con Google") {
		t.Error("la historia eliminada no tiene que aparecer en el backlog")
	}
	if rec := get(app, "/proyectos/1/backlog/1"); rec.Code != http.StatusNotFound {
		t.Errorf("la página de la historia eliminada: código = %d, se esperaba %d", rec.Code, http.StatusNotFound)
	}
	if cs, _ := app.criterios.ListarPorHistoria(context.Background(), 1); len(cs) != 0 {
		t.Errorf("quedaron %d criterios", len(cs))
	}
	if ts, _ := app.tareas.ListarPorHistoria(context.Background(), 1); len(ts) != 0 {
		t.Errorf("quedaron %d tareas", len(ts))
	}
}

// RN11 · las demás historias conservan su número, y la próxima toma el más alto más uno
func TestEliminarHistoria_LasDemasConservanSuNumero(t *testing.T) {
	app := conHistoriaParaEditar(t, nil)
	for _, titulo := range []string{"Registro con email", "Recuperar contraseña"} {
		campos := historiaValida()
		campos.Set("titulo", titulo)
		post(app, "/proyectos/1/backlog", campos)
	}

	post(app, "/proyectos/1/backlog/2/eliminar", nil)

	lista := get(app, "/proyectos/1/backlog").Body.String()
	verificarContiene(t, lista, "HU-1", "HU-3", "Recuperar contraseña")
	if strings.Contains(lista, "Registro con email") {
		t.Error("la HU-2 eliminada no tiene que aparecer")
	}
	nueva := post(app, "/proyectos/1/backlog", historiaValida())
	if destino := nueva.Header().Get("Location"); destino != "/proyectos/1/backlog?creada=4" {
		t.Errorf("la historia nueva redirige a %q, se esperaba ?creada=4 (HU-4)", destino)
	}
}

// RN11 · si se elimina la última, su número queda libre para la próxima
func TestEliminarHistoria_LaUltima_LiberaSuNumero(t *testing.T) {
	app := conHistoriaParaEditar(t, nil)

	post(app, "/proyectos/1/backlog/1/eliminar", nil)
	nueva := post(app, "/proyectos/1/backlog", historiaValida())

	if destino := nueva.Header().Get("Location"); destino != "/proyectos/1/backlog?creada=1" {
		t.Errorf("la historia nueva redirige a %q, se esperaba ?creada=1 (HU-1)", destino)
	}
}

// CA-06.2 y RN5 · una historia Hecha no se elimina: 409 con el motivo, y no se borra nada
func TestEliminarHistoria_HistoriaHecha_Responde409YNoBorraNada(t *testing.T) {
	app := conHistoriaParaEditar(t, nil)
	app.conCriteriosYTareas(t, 2, 1)
	if err := app.historias.ActualizarEstado(context.Background(), 1, backlog.EstadoHecho); err != nil {
		t.Fatalf("no se pudo dejar la historia Hecha: %v", err)
	}

	rec := post(app, "/proyectos/1/backlog/1/eliminar", nil)

	if rec.Code != http.StatusConflict {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusConflict)
	}
	verificarContiene(t, rec.Body.String(), "la historia está Hecha: no se puede eliminar")
	if rec := get(app, "/proyectos/1/backlog/1"); rec.Code != http.StatusOK {
		t.Errorf("la historia tenía que seguir: código %d", rec.Code)
	}
	if cs, _ := app.criterios.ListarPorHistoria(context.Background(), 1); len(cs) != 2 {
		t.Errorf("los criterios tenían que seguir (2) y hay %d", len(cs))
	}
}

// historiasQueNoSeBorran simula que la base rechaza el borrado por otros registros asociados (RN10).
type historiasQueNoSeBorran struct{ *memory.Historias }

func (historiasQueNoSeBorran) Eliminar(context.Context, int64) error {
	return backlog.ErrHistoriaConRegistros
}

// RN10 · si la historia tiene otros registros asociados, 409 con el motivo y la historia sigue
func TestEliminarHistoria_ConRegistrosAsociados_Responde409(t *testing.T) {
	app := conHistoriaParaEditar(t, func(h *memory.Historias) service.RepositorioEdicionHistorias {
		return historiasQueNoSeBorran{h}
	})

	rec := post(app, "/proyectos/1/backlog/1/eliminar", nil)

	if rec.Code != http.StatusConflict {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusConflict)
	}
	verificarContiene(t, rec.Body.String(), "la historia tiene registros asociados")
	if rec := get(app, "/proyectos/1/backlog/1"); rec.Code != http.StatusOK {
		t.Errorf("la historia tenía que seguir: código %d", rec.Code)
	}
}

// Sección 7 · proyecto o historia que no existen (también una historia que ya se eliminó)
func TestEliminarHistoria_ProyectoOHistoriaInexistente_Responde404(t *testing.T) {
	app := conHistoriaParaEditar(t, nil)

	for _, ruta := range []string{"/proyectos/99/backlog/1/eliminar", "/proyectos/1/backlog/99/eliminar", "/proyectos/1/backlog/abc/eliminar"} {
		if rec := get(app, ruta); rec.Code != http.StatusNotFound {
			t.Errorf("GET %s: código = %d, se esperaba %d", ruta, rec.Code, http.StatusNotFound)
		}
		if rec := post(app, ruta, nil); rec.Code != http.StatusNotFound {
			t.Errorf("POST %s: código = %d, se esperaba %d", ruta, rec.Code, http.StatusNotFound)
		}
	}
	post(app, "/proyectos/1/backlog/1/eliminar", nil)
	if rec := post(app, "/proyectos/1/backlog/1/eliminar", nil); rec.Code != http.StatusNotFound {
		t.Errorf("eliminar dos veces: código = %d, se esperaba %d", rec.Code, http.StatusNotFound)
	}
}

// El aviso del backlog solo muestra el número que el servidor pone: cualquier otro valor se ignora
func TestListaDelBacklog_Eliminada_SoloMuestraUnNumeroValido(t *testing.T) {
	app := conHistoriaParaEditar(t, nil)

	verificarContiene(t, get(app, "/proyectos/1/backlog?eliminada=7").Body.String(), "Historia HU-7 eliminada.")
	for _, consulta := range []string{"eliminada=abc", "eliminada=-3", "eliminada=0", "eliminada=%3Cb%3E"} {
		if strings.Contains(get(app, "/proyectos/1/backlog?"+consulta).Body.String(), "eliminada.") {
			t.Errorf("?%s: no tiene que mostrar el aviso", consulta)
		}
	}
}
