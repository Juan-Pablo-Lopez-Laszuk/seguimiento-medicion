package server_test

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// conHistoriaEIntegrantes arma la aplicación con el proyecto de ejemplo, su HU-1 y dos integrantes activos: Juan
// Pablo (ID 1) y Carolina (ID 2).
func conHistoriaEIntegrantes(t *testing.T) http.Handler {
	t.Helper()
	h := conHistoria(t)
	carolina := url.Values{"nombre": {"Carolina"}, "email": {"caro@mail.com"}, "rol": {string(member.RolProductBuilder)}}
	for _, datos := range []url.Values{integranteValido(), carolina} {
		if rec := post(h, "/proyectos/1/integrantes", datos); rec.Code != http.StatusSeeOther {
			t.Fatalf("no se pudo registrar a %s: código %d", datos.Get("nombre"), rec.Code)
		}
	}
	return h
}

func tareaValida() url.Values {
	return url.Values{"titulo": {"Pantalla de login"}, "responsable": {"2"}, "horas": {"2,5"}}
}

// agregarTarea agrega la tarea de ejemplo (ID 1) a la HU-1.
func agregarTarea(t *testing.T, h http.Handler) {
	t.Helper()
	if rec := post(h, "/proyectos/1/backlog/1/tareas", tareaValida()); rec.Code != http.StatusSeeOther {
		t.Fatalf("no se pudo agregar la tarea de ejemplo: código %d", rec.Code)
	}
}

func horas(valor string) url.Values {
	return url.Values{"horas": {valor}}
}

func completada(valor string) url.Values {
	return url.Values{"completada": {valor}}
}

// A las tareas se llega desde la página de la historia (HU-07).
func TestPaginaDeLaHistoria_EnlazaASusTareas(t *testing.T) {
	verificarContiene(t, get(conHistoria(t), "/proyectos/1/backlog/1").Body.String(), `href="/proyectos/1/backlog/1/tareas"`)
}

// CA-25.1 · la página muestra la historia, que no tiene tareas, y el formulario con los integrantes activos
func TestPaginaDeTareas_SinTareas_MuestraElFormularioConLosIntegrantes(t *testing.T) {
	rec := get(conHistoriaEIntegrantes(t), "/proyectos/1/backlog/1/tareas")

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	verificarContiene(t, rec.Body.String(),
		"<title>Tareas de HU-1 · Software Metrics</title>",
		"HU-1", "Iniciar sesión con Google", "Sin tareas todavía",
		`action="/proyectos/1/backlog/1/tareas"`, `method="post"`,
		`name="titulo"`, `name="responsable"`, `name="horas"`,
		`<option value="1">Juan Pablo</option>`, `<option value="2">Carolina</option>`,
		`href="/proyectos/1/backlog/1"`, // para volver a la historia
	)
}

// RN2 · sin integrantes activos no se puede elegir un responsable: la página avisa y enlaza a registrar uno
func TestPaginaDeTareas_SinIntegrantesActivos_AvisaQueHayQueRegistrarUno(t *testing.T) {
	html := get(conHistoria(t), "/proyectos/1/backlog/1/tareas").Body.String()

	verificarContiene(t, html, "Registrá un integrante", `href="/proyectos/1/integrantes/nuevo"`)
	if strings.Contains(html, `name="responsable"`) {
		t.Error("sin integrantes activos no debería mostrarse el formulario")
	}
}

// CA-25.1 y CA-23.2 · después de agregar se vuelve a la página con la tarea pendiente y la suma de horas
func TestAgregarTarea_VuelveALaPaginaConLaTareaYLaSuma(t *testing.T) {
	h := conHistoriaEIntegrantes(t)

	rec := post(h, "/proyectos/1/backlog/1/tareas", tareaValida())

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("código = %d, se esperaba %d (redirección)", rec.Code, http.StatusSeeOther)
	}
	destino := rec.Header().Get("Location")
	if destino != "/proyectos/1/backlog/1/tareas?agregada=1" {
		t.Fatalf("redirige a %q, se esperaba /proyectos/1/backlog/1/tareas?agregada=1", destino)
	}
	post(h, "/proyectos/1/backlog/1/tareas", url.Values{"titulo": {"Guardar la sesión"}, "responsable": {"1"}, "horas": {"3"}})
	verificarContiene(t, get(h, destino).Body.String(),
		"Tarea agregada.",
		"Pantalla de login", "Carolina", "2,5 h", "Pendiente",
		"Guardar la sesión", "Juan Pablo", "3 h",
		"Horas estimadas de la historia", "5,5 h",
		`action="/proyectos/1/backlog/1/tareas/1/horas"`, `action="/proyectos/1/backlog/1/tareas/1/completada"`,
		`name="completada" value="1"`, "Completar",
	)
}

// RN1 a RN3 · cada error aparece debajo de su campo, se conserva lo que se escribió y no se guarda nada
func TestAgregarTarea_DatosInvalidos_MuestraLosErroresYConservaLoEscrito(t *testing.T) {
	h := conHistoriaEIntegrantes(t)

	rec := post(h, "/proyectos/1/backlog/1/tareas", url.Values{"titulo": {"   "}, "responsable": {"2"}, "horas": {"cero"}})

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusUnprocessableEntity)
	}
	verificarContiene(t, rec.Body.String(),
		"el título de la tarea es obligatorio", "las horas estimadas deben ser mayores a 0",
		"is-invalid", `value="cero"`, `<option value="2" selected>Carolina</option>`,
	)
	if html := get(h, "/proyectos/1/backlog/1/tareas").Body.String(); !strings.Contains(html, "Sin tareas todavía") {
		t.Error("no se debería haber guardado ninguna tarea")
	}
}

// RN2 · un integrante dado de baja no se ofrece como responsable, y si igual llega, se rechaza
func TestAgregarTarea_ResponsableDadoDeBaja_NoSeOfreceYSeRechaza(t *testing.T) {
	h := conHistoriaEIntegrantes(t)
	post(h, "/proyectos/1/integrantes/1/baja", nil)

	if html := get(h, "/proyectos/1/backlog/1/tareas").Body.String(); strings.Contains(html, `<option value="1">`) {
		t.Error("un integrante dado de baja no debería ofrecerse como responsable")
	}
	rec := post(h, "/proyectos/1/backlog/1/tareas", url.Values{"titulo": {"Guardar la sesión"}, "responsable": {"1"}, "horas": {"3"}})

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusUnprocessableEntity)
	}
	verificarContiene(t, rec.Body.String(), "el responsable tiene que ser un integrante activo del proyecto")
}

// CA-25.3 · se cambian las horas de una tarea pendiente y la suma de la historia se actualiza
func TestCambiarHoras_TareaPendiente_VuelveALaPaginaConLasHorasNuevas(t *testing.T) {
	h := conHistoriaEIntegrantes(t)
	agregarTarea(t, h)

	rec := post(h, "/proyectos/1/backlog/1/tareas/1/horas", horas("4"))

	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/proyectos/1/backlog/1/tareas?horas=1" {
		t.Fatalf("código = %d y destino %q, se esperaba 303 a /proyectos/1/backlog/1/tareas?horas=1", rec.Code, rec.Header().Get("Location"))
	}
	html := get(h, rec.Header().Get("Location")).Body.String()
	verificarContiene(t, html, "Horas estimadas actualizadas.", "4 h")
	if strings.Contains(html, "2,5 h") {
		t.Error("ya no deberían verse las horas viejas")
	}
}

// RN3 · horas inválidas: el error se muestra arriba de la lista y la tarea queda como estaba
func TestCambiarHoras_HorasInvalidas_MuestraElErrorYNoCambiaNada(t *testing.T) {
	h := conHistoriaEIntegrantes(t)
	agregarTarea(t, h)

	rec := post(h, "/proyectos/1/backlog/1/tareas/1/horas", horas("0"))

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusUnprocessableEntity)
	}
	verificarContiene(t, rec.Body.String(), "alert-danger", "Pantalla de login: las horas estimadas deben ser mayores a 0", "2,5 h")
}

// RN7 y CA-25.3 · completar y volver a pendiente; una tarea completada no muestra el formulario de horas
func TestMarcarTarea_CompletadaYPendiente_VuelveALaPagina(t *testing.T) {
	h := conHistoriaEIntegrantes(t)
	agregarTarea(t, h)

	rec := post(h, "/proyectos/1/backlog/1/tareas/1/completada", completada("1"))

	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/proyectos/1/backlog/1/tareas?marcada=completada" {
		t.Fatalf("código = %d y destino %q, se esperaba 303 a ?marcada=completada", rec.Code, rec.Header().Get("Location"))
	}
	html := get(h, rec.Header().Get("Location")).Body.String()
	verificarContiene(t, html, "Tarea marcada como completada.", "Completada", `name="completada" value="0"`, "Volver a pendiente", "2,5 h")
	if strings.Contains(html, `action="/proyectos/1/backlog/1/tareas/1/horas"`) {
		t.Error("una tarea completada no debería tener el formulario para cambiar sus horas")
	}

	rec = post(h, "/proyectos/1/backlog/1/tareas/1/completada", completada("0"))
	verificarContiene(t, get(h, rec.Header().Get("Location")).Body.String(), "Tarea marcada como pendiente.", "Pendiente")
}

// CA-25.3 · cambiar las horas de una tarea completada se rechaza con un aviso (409) y no cambia nada
func TestCambiarHoras_TareaCompletada_Responde409ConAviso(t *testing.T) {
	h := conHistoriaEIntegrantes(t)
	agregarTarea(t, h)
	post(h, "/proyectos/1/backlog/1/tareas/1/completada", completada("1"))

	rec := post(h, "/proyectos/1/backlog/1/tareas/1/horas", horas("5"))

	if rec.Code != http.StatusConflict {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusConflict)
	}
	verificarContiene(t, rec.Body.String(), "alert-warning", "la tarea está completada: no se puede cambiar su estimación", "2,5 h")
}

func TestMarcarTarea_ValorInvalido_Responde400(t *testing.T) {
	h := conHistoriaEIntegrantes(t)
	agregarTarea(t, h)

	if rec := post(h, "/proyectos/1/backlog/1/tareas/1/completada", completada("si")); rec.Code != http.StatusBadRequest {
		t.Errorf("código = %d, se esperaba %d", rec.Code, http.StatusBadRequest)
	}
	verificarContiene(t, get(h, "/proyectos/1/backlog/1/tareas").Body.String(), "Pendiente")
}

// La página solo muestra los mensajes que pone el servidor, y escapa lo que escribió el usuario.
func TestPaginaDeTareas_SoloMensajesConocidosYEscapaElTitulo(t *testing.T) {
	h := conHistoriaEIntegrantes(t)
	post(h, "/proyectos/1/backlog/1/tareas", url.Values{"titulo": {"<script>alert(1)</script>"}, "responsable": {"2"}, "horas": {"1"}})

	html := get(h, "/proyectos/1/backlog/1/tareas?agregada=<b>hola</b>&marcada=otra").Body.String()

	if strings.Contains(html, "<script>alert(1)</script>") || strings.Contains(html, "<b>hola</b>") || strings.Contains(html, "Tarea agregada.") {
		t.Error("la página no debería mostrar HTML sin escapar ni mensajes que no puso el servidor")
	}
	verificarContiene(t, html, "&lt;script&gt;alert(1)&lt;/script&gt;")
}

// CA-25.2 y sección 7 · proyecto, historia o tarea que no existen, o una tarea de otra historia, son un 404
func TestTareas_ProyectoHistoriaOTareaInexistente_Responde404(t *testing.T) {
	h := conHistoriaEIntegrantes(t)
	segunda := historiaValida()
	segunda.Set("titulo", "Ver el tablero")
	post(h, "/proyectos/1/backlog", segunda)
	agregarTarea(t, h) // tarea 1, de la HU-1

	pedidos := map[string]int{
		"proyecto inexistente":       get(h, "/proyectos/99/backlog/1/tareas").Code,
		"historia inexistente":       get(h, "/proyectos/1/backlog/99/tareas").Code,
		"historia que no es número":  get(h, "/proyectos/1/backlog/abc/tareas").Code,
		"agregar a otra inexistente": post(h, "/proyectos/1/backlog/99/tareas", tareaValida()).Code,
		"tarea inexistente":          post(h, "/proyectos/1/backlog/1/tareas/99/horas", horas("4")).Code,
		"tarea de otra historia":     post(h, "/proyectos/1/backlog/2/tareas/1/horas", horas("4")).Code,
		"marcar de otra historia":    post(h, "/proyectos/1/backlog/2/tareas/1/completada", completada("1")).Code,
		"tarea que no es número":     post(h, "/proyectos/1/backlog/1/tareas/abc/completada", completada("1")).Code,
	}
	for nombre, codigo := range pedidos {
		if codigo != http.StatusNotFound {
			t.Errorf("%s: código = %d, se esperaba %d", nombre, codigo, http.StatusNotFound)
		}
	}
}

// tareasCaidas simula una base que no responde.
type tareasCaidas struct{}

func (tareasCaidas) ListarPorHistoria(context.Context, int64) ([]effort.Tarea, error) {
	return nil, errBaseCaida
}
func (tareasCaidas) Guardar(context.Context, effort.Tarea) (effort.Tarea, error) {
	return effort.Tarea{}, errBaseCaida
}
func (tareasCaidas) BuscarPorID(context.Context, int64, int64) (effort.Tarea, error) {
	return effort.Tarea{}, errBaseCaida
}
func (tareasCaidas) Actualizar(context.Context, effort.Tarea) error { return errBaseCaida }

// Si la base falla, la página muestra un mensaje general (500) y no el error interno.
func TestTareas_ErrorDeLaBase_MuestraUnMensajeGeneral(t *testing.T) {
	proyectos, historias, integrantes := memory.NuevoProyectos(), memory.NuevoHistorias(), memory.NuevoIntegrantes()
	h := server.NewRouter(server.Dependencias{
		Proyectos: service.NuevoProyectos(proyectos, relojDeLaVista),
		Historias: service.NuevoHistorias(proyectos, historias),
		Criterios: service.NuevoCriterios(proyectos, historias, memory.NuevoCriterios()),
		Tareas:    service.NuevoTareas(proyectos, historias, integrantes, tareasCaidas{}),
	})
	post(h, "/proyectos", formularioValido())
	post(h, "/proyectos/1/backlog", historiaValida())

	rec := get(h, "/proyectos/1/backlog/1/tareas")

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusInternalServerError)
	}
	if strings.Contains(rec.Body.String(), errBaseCaida.Error()) {
		t.Error("no se debería mostrar el error interno")
	}
}
