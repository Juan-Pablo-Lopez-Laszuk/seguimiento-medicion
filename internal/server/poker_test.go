package server_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// Mensajes de la spec de HU-17 (secciones 3 y 7).
const (
	mensajeSesionAbierta = "Sesión de Planning Poker abierta."
	mensajeNoIntegrante  = "No sos integrante de este proyecto."
)

// appPoker es la aplicación en memoria con el proyecto 1 "Software Metrics", sus integrantes (Juan Agile Enabler,
// Carolina Product Builder y Lucía dada de baja; Ana, con la cuenta de los tests de TEC-05, si se pide; si es la Agile Enabler, Juan pasa a Product Builder) y las
// historias HU-1 Login (Pendiente) y HU-2 Reportes (Hecha). Si el repositorio de sesiones también guarda votos, se
// conectan (HU-18).
type appPoker struct {
	dep       server.Dependencias
	historias *memory.Historias
	porNombre map[string]member.Integrante
}

// nuevaAppPoker arma la aplicación. anaRol es el rol de Ana en el proyecto, o "" si Ana no es integrante.
func nuevaAppPoker(t *testing.T, anaRol member.Rol, sesiones service.RepositorioSesionesPoker) *appPoker {
	t.Helper()
	ctx := context.Background()
	proyectos, historias, integrantes, criterios := memory.NuevoProyectos(), memory.NuevoHistorias(), memory.NuevoIntegrantes(), memory.NuevoCriterios()
	if sesiones == nil {
		sesiones = memory.NuevoPoker()
	}
	p, err := proyectos.Guardar(ctx, project.Proyecto{Nombre: "Software Metrics", FechaInicio: time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC), FechaFin: time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("guardar el proyecto: %v", err)
	}
	a := &appPoker{historias: historias, porNombre: map[string]member.Integrante{}}
	lista := []member.Integrante{
		{Nombre: "Juan", Email: "juan@mail.com", Rol: member.RolAgileEnabler, Activo: true},
		{Nombre: "Carolina", Email: "carolina@mail.com", Rol: member.RolProductBuilder, Activo: true},
		{Nombre: "Lucía", Email: "lucia@mail.com", Rol: member.RolProductBuilder},
	}
	if anaRol == member.RolAgileEnabler {
		// Ana ocupa el lugar de Juan: el proyecto tiene un solo Agile Enabler activo.
		lista[0].Rol = member.RolProductBuilder
	}
	if anaRol != "" {
		lista = append(lista, member.Integrante{Nombre: "Ana", Email: "ana@mail.com", Rol: anaRol, Activo: true})
	}
	for _, i := range lista {
		i.ProyectoID = p.ID
		guardado, err := integrantes.Guardar(ctx, i)
		if err != nil {
			t.Fatalf("guardar a %s: %v", i.Nombre, err)
		}
		a.porNombre[i.Nombre] = guardado
	}
	for n, h := range []backlog.Historia{{Titulo: "Login", Estado: backlog.EstadoPendiente}, {Titulo: "Reportes", Estado: backlog.EstadoHecho}} {
		h.ProyectoID, h.Numero, h.Prioridad = p.ID, n+1, backlog.PrioridadAlta
		if _, err := historias.Guardar(ctx, h); err != nil {
			t.Fatalf("guardar %q: %v", h.Titulo, err)
		}
	}
	casos := service.NuevoPoker(proyectos, historias, integrantes, sesiones)
	if votos, ok := sesiones.(service.RepositorioVotosPoker); ok {
		casos.ConVotos(votos)
	}
	a.dep = server.Dependencias{
		Proyectos:   service.NuevoProyectos(proyectos, time.Now),
		Integrantes: service.NuevoIntegrantes(proyectos, integrantes),
		Historias:   service.NuevoHistorias(proyectos, historias),
		Criterios:   service.NuevoCriterios(proyectos, historias, criterios),
		Poker:       casos,
		Cuentas:     service.NuevoCuentas(&servicioDeLoginSimulado{}, integrantes),
	}
	return a
}

func (a *appPoker) sinLogin() http.Handler { return server.NewRouter(a.dep) }

func (a *appPoker) conLogin() http.Handler {
	dep := a.dep
	dep.Verificador = verificadorDeTokens{}
	return server.NewRouter(dep)
}

func (a *appPoker) abre(nombre string) url.Values {
	return url.Values{"quien": {fmt.Sprint(a.porNombre[nombre].ID)}}
}

const rutaPoker = "/proyectos/1/backlog/1/poker"

func verificarNoContiene(t *testing.T, cuerpo string, textos ...string) {
	t.Helper()
	for _, texto := range textos {
		if strings.Contains(cuerpo, texto) {
			t.Errorf("la página no tenía que mostrar %q", texto)
		}
	}
}

// T5 · la página de la historia tiene el enlace a su sesión de Planning Poker
func TestHistoria_EnlazaAPlanningPoker(t *testing.T) {
	rec := pedido(nuevaAppPoker(t, "", nil).sinLogin(), "/proyectos/1/backlog/1", nil)

	verificarContiene(t, rec.Body.String(), `href="`+rutaPoker+`"`, "Planning Poker")
}

// Sin sesión abierta la página ofrece abrirla; sin login, con el desplegable Quién abre con los integrantes activos
func TestPoker_SinSesion_OfreceAbrirla(t *testing.T) {
	rec := pedido(nuevaAppPoker(t, "", nil).sinLogin(), rutaPoker, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba 200", rec.Code)
	}
	cuerpo := rec.Body.String()
	verificarContiene(t, cuerpo, "HU-1 · Login", `action="`+rutaPoker+`"`, `name="quien"`, "Quién abre", "Abrir sesión", "Participantes", "Juan", "Agile Enabler", "Carolina")
	verificarNoContiene(t, cuerpo, "Lucía", "Ronda 1")
}

// CA-17.1, RN3 y CA-17.3 · el Agile Enabler abre la sesión: 303 a la página, que muestra la Ronda 1 y los participantes
func TestPoker_AbrirSinLogin(t *testing.T) {
	a := nuevaAppPoker(t, "", nil)
	h := a.sinLogin()

	rec := pedido(h, rutaPoker, a.abre("Juan"))

	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != rutaPoker+"?abierta=1" {
		t.Fatalf("código = %d, Location = %q; se esperaba 303 a %s?abierta=1", rec.Code, rec.Header().Get("Location"), rutaPoker)
	}
	cuerpo := pedido(h, rutaPoker+"?abierta=1", nil).Body.String()
	verificarContiene(t, cuerpo, mensajeSesionAbierta, "Sesión abierta", "Ronda 1", "Juan", "Carolina")
	verificarNoContiene(t, cuerpo, "Abrir sesión", "Lucía")
}

// Errores al abrir, sin login: el código de la spec y la página con el aviso
func TestPoker_AbrirSinLogin_Errores(t *testing.T) {
	casos := []struct {
		nombre string
		ruta   string
		quien  string
		codigo int
		texto  string
	}{
		{"historia Hecha (CA-17.1)", "/proyectos/1/backlog/2/poker", "Juan", http.StatusConflict, poker.ErrHistoriaHecha.Error()},
		{"la abre un Product Builder (RN4)", rutaPoker, "Carolina", http.StatusForbidden, poker.ErrSoloAgileEnabler.Error()},
		{"no se eligió quién la abre (RN4)", rutaPoker, "", http.StatusUnprocessableEntity, poker.ErrSoloAgileEnabler.Error()},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			a := nuevaAppPoker(t, "", nil)
			campos := url.Values{"quien": {""}}
			if c.quien != "" {
				campos = a.abre(c.quien)
			}

			rec := pedido(a.sinLogin(), c.ruta, campos)

			if rec.Code != c.codigo {
				t.Errorf("código = %d, se esperaba %d", rec.Code, c.codigo)
			}
			verificarContiene(t, rec.Body.String(), c.texto)
		})
	}
}

// Sin elegir quién abre, el error va debajo del desplegable
func TestPoker_AbrirSinElegirQuien_ElErrorVaEnElCampo(t *testing.T) {
	rec := pedido(nuevaAppPoker(t, "", nil).sinLogin(), rutaPoker, url.Values{"quien": {""}})

	verificarContiene(t, rec.Body.String(), "is-invalid", `<div class="invalid-feedback">`+poker.ErrSoloAgileEnabler.Error())
}

// CA-17.2 · abrir otra vez da 409 y la página sigue mostrando la sesión abierta
func TestPoker_AbrirDosVeces(t *testing.T) {
	a := nuevaAppPoker(t, "", nil)
	h := a.sinLogin()
	pedido(h, rutaPoker, a.abre("Juan"))

	rec := pedido(h, rutaPoker, a.abre("Juan"))

	if rec.Code != http.StatusConflict {
		t.Errorf("código = %d, se esperaba 409", rec.Code)
	}
	verificarContiene(t, rec.Body.String(), poker.ErrSesionYaAbierta.Error(), "Ronda 1")
}

// CA-17.1 · la página de una historia Hecha no ofrece abrir la sesión
func TestPoker_HistoriaHecha_NoOfreceAbrir(t *testing.T) {
	rec := pedido(nuevaAppPoker(t, "", nil).sinLogin(), "/proyectos/1/backlog/2/poker", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba 200", rec.Code)
	}
	verificarContiene(t, rec.Body.String(), "La historia está Hecha: ya no se estima.")
	verificarNoContiene(t, rec.Body.String(), "Abrir sesión")
}

func TestPoker_NoEncontrado(t *testing.T) {
	a := nuevaAppPoker(t, "", nil)
	for _, ruta := range []string{"/proyectos/9/backlog/1/poker", "/proyectos/1/backlog/9/poker", "/proyectos/1/backlog/x/poker", "/proyectos/x/backlog/1/poker"} {
		if rec := pedido(a.sinLogin(), ruta, nil); rec.Code != http.StatusNotFound {
			t.Errorf("GET %s: código = %d, se esperaba 404", ruta, rec.Code)
		}
		if rec := pedido(a.sinLogin(), ruta, a.abre("Juan")); rec.Code != http.StatusNotFound {
			t.Errorf("POST %s: código = %d, se esperaba 404", ruta, rec.Code)
		}
	}
}

// pokerQueFalla es el repositorio de sesiones con la base caída.
type pokerQueFalla struct{}

var errPokerCaido = errors.New("la base no responde")

func (pokerQueFalla) Abrir(context.Context, int64) (poker.Sesion, poker.Ronda, error) {
	return poker.Sesion{}, poker.Ronda{}, errPokerCaido
}

func (pokerQueFalla) SesionAbierta(context.Context, int64) (poker.Sesion, poker.Ronda, error) {
	return poker.Sesion{}, poker.Ronda{}, errPokerCaido
}

func TestPoker_FallaLaBase(t *testing.T) {
	a := nuevaAppPoker(t, "", pokerQueFalla{})

	if rec := pedido(a.sinLogin(), rutaPoker, nil); rec.Code != http.StatusInternalServerError {
		t.Errorf("GET: código = %d, se esperaba 500", rec.Code)
	}
	rec := pedido(a.sinLogin(), rutaPoker, a.abre("Juan"))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("POST: código = %d, se esperaba 500", rec.Code)
	}
	verificarContiene(t, rec.Body.String(), "No se pudo abrir la sesión. Probá de nuevo en unos minutos.")
}

// RN6 · con login abre el integrante vinculado a la cuenta: no se elige quién abre (y si viene, no se usa)
func TestPoker_ConLogin_AbreElIntegranteDeLaCuenta(t *testing.T) {
	a := nuevaAppPoker(t, member.RolAgileEnabler, nil)
	h := a.conLogin()

	pagina := pedido(h, rutaPoker, nil, conSesion("token-de-ana"))
	verificarNoContiene(t, pagina.Body.String(), `name="quien"`)
	verificarContiene(t, pagina.Body.String(), "Abrir sesión")

	rec := pedido(h, rutaPoker, a.abre("Carolina"), conSesion("token-de-ana"))
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("código = %d, se esperaba 303: %s", rec.Code, rec.Body.String())
	}
	verificarContiene(t, pedido(h, rutaPoker, nil, conSesion("token-de-ana")).Body.String(), "Ronda 1")
}

// RN4 y RN6 · con login, la cuenta de una Product Builder no abre la sesión, aunque elija a otro integrante
func TestPoker_ConLogin_UnaProductBuilderNoAbre(t *testing.T) {
	a := nuevaAppPoker(t, member.RolProductBuilder, nil)

	rec := pedido(a.conLogin(), rutaPoker, a.abre("Juan"), conSesion("token-de-ana"))

	if rec.Code != http.StatusForbidden {
		t.Errorf("código = %d, se esperaba 403", rec.Code)
	}
	verificarContiene(t, rec.Body.String(), poker.ErrSoloAgileEnabler.Error())
}

// Con login, una cuenta que no es integrante del proyecto ve la sesión pero no la abre
func TestPoker_ConLogin_NoEsIntegrante(t *testing.T) {
	a := nuevaAppPoker(t, "", nil)
	h := a.conLogin()

	if rec := pedido(h, rutaPoker, nil, conSesion("token-de-ana")); rec.Code != http.StatusOK {
		t.Errorf("ver: código = %d, se esperaba 200", rec.Code)
	}
	rec := pedido(h, rutaPoker, url.Values{}, conSesion("token-de-ana"))
	if rec.Code != http.StatusForbidden {
		t.Errorf("abrir: código = %d, se esperaba 403", rec.Code)
	}
	verificarContiene(t, rec.Body.String(), mensajeNoIntegrante)
}

// TEC-05 · con el login activado, ver y abrir la sesión piden sesión
func TestPoker_ConLogin_PideSesion(t *testing.T) {
	h := nuevaAppPoker(t, "", nil).conLogin()
	if rec := pedido(h, rutaPoker, url.Values{}); rec.Code != http.StatusUnauthorized {
		t.Errorf("abrir sin sesión: código = %d, se esperaba 401", rec.Code)
	}
	if rec := pedido(h, rutaPoker, nil); rec.Code == http.StatusOK {
		t.Errorf("ver sin sesión: código = %d, no se esperaba 200", rec.Code)
	}
}
