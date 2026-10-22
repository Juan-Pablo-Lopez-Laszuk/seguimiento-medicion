package server_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/auth"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/account"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

const usuarioDeAna = "11111111-1111-4111-8111-111111111111"

// Mensajes de la spec (secciones 3 y 7).
const (
	mensajeSinSesion          = "Iniciá sesión para continuar."
	mensajeSesionVencida      = "Tu sesión venció, iniciá sesión de nuevo."
	mensajeCuentaCreada       = "Cuenta creada. Ya podés iniciar sesión."
	mensajeConfirmarEmail     = "Te enviamos un email para confirmar la cuenta. Confirmala y después iniciá sesión."
	mensajeSalio              = "Cerraste sesión."
	mensajeCredenciales       = "El email o la contraseña no son correctos."
	mensajeSinConfirmar       = "Confirmá tu email antes de iniciar sesión."
	mensajeSinServicio        = "No pudimos conectar con el servicio de inicio de sesión. Probá de nuevo en unos minutos."
	mensajeDemasiadosIntentos = "Demasiados intentos. Esperá un momento y volvé a probar."
)

var venceLaSesion = time.Date(2026, time.October, 22, 18, 0, 0, 0, time.UTC)

// verificadorDeTokens reconoce tres tokens: el de Ana, uno vencido y cualquier otro, que no vale.
type verificadorDeTokens struct{}

func (verificadorDeTokens) Verificar(_ context.Context, token string) (auth.Identidad, error) {
	switch token {
	case "token-de-ana":
		return auth.Identidad{ID: usuarioDeAna, Email: "ana@mail.com"}, nil
	case "token-vencido":
		return auth.Identidad{}, auth.ErrTokenVencido
	}
	return auth.Identidad{}, fmt.Errorf("%w: no lo conozco", auth.ErrTokenInvalido)
}

// servicioDeLoginSimulado hace de Supabase Auth: la cuenta de Ana existe y confirmada, salvo que se indique otra cosa.
type servicioDeLoginSimulado struct {
	pideConfirmar bool
	sinConfirmar  bool
	falla         error // si no es nil, todo lo que se le pide falla con este error
	registradas   []account.Credenciales
}

func (l *servicioDeLoginSimulado) Registrar(_ context.Context, c account.Credenciales) (auth.ResultadoRegistro, error) {
	if l.falla != nil {
		return auth.ResultadoRegistro{}, l.falla
	}
	if c.Email == "ana@mail.com" {
		return auth.ResultadoRegistro{}, auth.ErrEmailYaRegistrado
	}
	l.registradas = append(l.registradas, c)
	return auth.ResultadoRegistro{RequiereConfirmacion: l.pideConfirmar}, nil
}

func (l *servicioDeLoginSimulado) Ingresar(_ context.Context, c account.Credenciales) (auth.Sesion, error) {
	switch {
	case l.falla != nil:
		return auth.Sesion{}, l.falla
	case c.Email != "ana@mail.com" || c.Clave != "clave-segura-1":
		return auth.Sesion{}, auth.ErrCredencialesIncorrectas
	case l.sinConfirmar:
		return auth.Sesion{}, auth.ErrEmailSinConfirmar
	}
	return auth.Sesion{Token: "token-de-ana", Vence: venceLaSesion, Usuario: auth.Identidad{ID: usuarioDeAna, Email: "ana@mail.com"}}, nil
}

// routerConLogin arma la aplicación con el login activado, el servicio de login simulado y repositorios en memoria vacíos.
func routerConLogin(login *servicioDeLoginSimulado) http.Handler {
	dep := dependenciasEnMemoria(relojDeLaVista)
	dep.Cuentas = service.NuevoCuentas(login, memory.NuevoIntegrantes())
	dep.Verificador = verificadorDeTokens{}
	return server.NewRouter(dep)
}

// pedido simula un pedido a la aplicación; campos == nil es un GET y, si no, un POST de formulario.
func pedido(h http.Handler, ruta string, campos url.Values, opciones ...func(*http.Request)) *httptest.ResponseRecorder {
	metodo, cuerpo := http.MethodGet, ""
	if campos != nil {
		metodo, cuerpo = http.MethodPost, campos.Encode()
	}
	req := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	if campos != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for _, opcion := range opciones {
		opcion(req)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func conSesion(token string) func(*http.Request) {
	return func(r *http.Request) { r.AddCookie(&http.Cookie{Name: "sesion", Value: token}) }
}

func porHTTPS(r *http.Request) { r.Header.Set("X-Forwarded-Proto", "https") }

// cookieDeSesion devuelve la cookie "sesion" que puso la respuesta, o nil.
func cookieDeSesion(rec *httptest.ResponseRecorder) *http.Cookie {
	for _, c := range rec.Result().Cookies() {
		if c.Name == "sesion" {
			return c
		}
	}
	return nil
}

func verificarCookieBorrada(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()
	c := cookieDeSesion(rec)
	if c == nil || c.Value != "" || c.MaxAge >= 0 {
		t.Errorf("se esperaba que la respuesta borre la cookie sesion y se obtuvo %+v", c)
	}
}

func credencialesDeAna() url.Values {
	return url.Values{"email": {"Ana@Mail.com"}, "clave": {"clave-segura-1"}}
}

// RN1 · sin login configurado la aplicación no pide sesión y no tiene las pantallas de cuentas
func TestSinLogin_NoPideSesionNiTieneLasPantallasDeCuentas(t *testing.T) {
	h := nuevoRouter()

	if rec := get(h, "/proyectos"); rec.Code != http.StatusOK {
		t.Errorf("/proyectos sin login configurado: código %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	for _, ruta := range []string{"/login", "/registro"} {
		if rec := get(h, ruta); rec.Code != http.StatusNotFound {
			t.Errorf("%s sin login configurado: código %d, se esperaba %d", ruta, rec.Code, http.StatusNotFound)
		}
	}
	if strings.Contains(get(h, "/proyectos").Body.String(), `action="/salir"`) {
		t.Error("sin login configurado la barra de navegación no tendría que tener el botón Salir")
	}
}

// CA-05.2 · con una sesión válida se accede, y la barra de navegación muestra el email y el botón Salir
func TestConLogin_SesionValida_AccedeYMuestraElUsuario(t *testing.T) {
	rec := pedido(routerConLogin(&servicioDeLoginSimulado{}), "/proyectos", nil, conSesion("token-de-ana"))

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	verificarContiene(t, rec.Body.String(), "ana@mail.com", `action="/salir"`, "Salir")
}

// CA-05.2 · RN3: sin sesión válida el pedido se rechaza con 401 y la página de login con el motivo, y no llega al handler
func TestConLogin_SinSesionValida_Responde401ConLaPaginaDeLogin(t *testing.T) {
	h := routerConLogin(&servicioDeLoginSimulado{})
	reservado := formularioValido()
	reservado.Set("nombre", "Proyecto Reservado")
	if rec := pedido(h, "/proyectos", reservado, conSesion("token-de-ana")); rec.Code != http.StatusSeeOther {
		t.Fatalf("no se pudo crear el proyecto de ejemplo con sesión: código %d", rec.Code)
	}
	casos := []struct {
		nombre        string
		opcion        func(*http.Request)
		mensaje       string
		borraLaCookie bool
	}{
		{"sin cookie", func(*http.Request) {}, mensajeSinSesion, false},
		{"cookie con basura", conSesion("esto-no-es-un-token"), mensajeSinSesion, true},
		{"cookie vacía", conSesion(""), mensajeSinSesion, true},
		{"token vencido", conSesion("token-vencido"), mensajeSesionVencida, true},
	}
	for _, c := range casos {
		rec := pedido(h, "/proyectos", nil, c.opcion)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s: código = %d, se esperaba %d", c.nombre, rec.Code, http.StatusUnauthorized)
		}
		verificarContiene(t, rec.Body.String(), "<title>Iniciar sesión · Software Metrics</title>", c.mensaje, `name="email"`, `name="clave"`)
		if strings.Contains(rec.Body.String(), "Proyecto Reservado") {
			t.Errorf("%s: la respuesta no tendría que mostrar la lista de proyectos", c.nombre)
		}
		if c.borraLaCookie {
			verificarCookieBorrada(t, rec)
		} else if cookieDeSesion(rec) != nil {
			t.Errorf("%s: no había cookie y la respuesta no tendría que poner ninguna", c.nombre)
		}
	}
}

// RN3 · un pedido que modifica datos sin sesión también se rechaza, y no se ejecuta
func TestConLogin_PostSinSesion_NoEjecutaNada(t *testing.T) {
	h := routerConLogin(&servicioDeLoginSimulado{})

	rec := pedido(h, "/proyectos", formularioValido())

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusUnauthorized)
	}
	if strings.Contains(rec.Body.String(), `name="siguiente"`) {
		t.Error("después de ingresar no se puede repetir un POST: la página no tendría que ofrecer volver a esa ruta")
	}
	verificarContiene(t, pedido(h, "/proyectos", nil, conSesion("token-de-ana")).Body.String(), "Todavía no hay proyectos")
}

// RN9 · la página de 401 de un GET ofrece volver a esa ruta después de ingresar
func TestConLogin_SinSesion_LaPaginaDeLoginRecuerdaAdondeIba(t *testing.T) {
	h := routerConLogin(&servicioDeLoginSimulado{})

	rec := pedido(h, "/proyectos/1/backlog", nil)

	verificarContiene(t, rec.Body.String(), `name="siguiente" value="/proyectos/1/backlog"`)
}

// RN1 · /health, /login y /registro quedan abiertos con el login activado
func TestConLogin_RutasAbiertas(t *testing.T) {
	h := routerConLogin(&servicioDeLoginSimulado{})
	for _, ruta := range []string{"/health", "/login", "/registro"} {
		if rec := pedido(h, ruta, nil); rec.Code != http.StatusOK {
			t.Errorf("%s sin sesión: código %d, se esperaba %d", ruta, rec.Code, http.StatusOK)
		}
	}
}

// CA-05.1 · el formulario de ingreso
func TestFormularioDeLogin_MuestraLosCamposYLosAvisos(t *testing.T) {
	h := routerConLogin(&servicioDeLoginSimulado{})

	rec := pedido(h, "/login", nil)

	verificarContiene(t, rec.Body.String(), "<title>Iniciar sesión · Software Metrics</title>",
		`action="/login"`, `method="post"`, `name="email"`, `type="email"`, `name="clave"`, `type="password"`, `href="/registro"`)
	for consulta, mensaje := range map[string]string{"registrado=1": mensajeCuentaCreada, "confirmar=1": mensajeConfirmarEmail, "salio=1": mensajeSalio} {
		verificarContiene(t, pedido(h, "/login?"+consulta, nil).Body.String(), mensaje)
	}
}

// CA-05.1 · RN4 y RN6: ingreso correcto guarda la cookie de sesión y va a la página de inicio
func TestIngresar_CredencialesCorrectas_GuardaLaCookieYVaAlInicio(t *testing.T) {
	rec := pedido(routerConLogin(&servicioDeLoginSimulado{}), "/login", credencialesDeAna())

	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/" {
		t.Fatalf("se esperaba 303 a / y se obtuvo %d a %q", rec.Code, rec.Header().Get("Location"))
	}
	c := cookieDeSesion(rec)
	if c == nil || c.Value != "token-de-ana" {
		t.Fatalf("se esperaba la cookie sesion con el token y se obtuvo %+v", c)
	}
	if !c.HttpOnly || c.SameSite != http.SameSiteLaxMode || c.Path != "/" || c.Secure {
		t.Errorf("la cookie tendría que ser HttpOnly, SameSite=Lax, de Path=/ y no Secure por HTTP: %+v", c)
	}
	if !c.Expires.Equal(venceLaSesion) {
		t.Errorf("la cookie vence %v, se esperaba cuando vence el token: %v", c.Expires, venceLaSesion)
	}
}

// RN4 · la cookie es Secure cuando el pedido llegó por HTTPS, también detrás de un proxy como el de Vercel
func TestIngresar_PorHTTPS_LaCookieEsSecure(t *testing.T) {
	h := routerConLogin(&servicioDeLoginSimulado{})

	if c := cookieDeSesion(pedido(h, "/login", credencialesDeAna(), porHTTPS)); c == nil || !c.Secure {
		t.Errorf("con X-Forwarded-Proto https la cookie tendría que ser Secure: %+v", c)
	}
	if c := cookieDeSesion(pedido(h, "/salir", url.Values{}, porHTTPS)); c == nil || !c.Secure {
		t.Errorf("la cookie que borra Salir también tendría que ser Secure: %+v", c)
	}
}

// RN9 · después de ingresar se vuelve a donde se estaba, pero solo a rutas internas
func TestIngresar_Siguiente_SoloAceptaRutasInternas(t *testing.T) {
	h := routerConLogin(&servicioDeLoginSimulado{})
	casos := []struct{ siguiente, destino string }{
		{"/proyectos/1/backlog", "/proyectos/1/backlog"},
		{"/proyectos?pagina=2", "/proyectos?pagina=2"},
		{"", "/"},
		{"https://malo.com", "/"},
		{"//malo.com", "/"},
		{`/\malo.com`, "/"},
		{"/\t/malo.com", "/"}, // el navegador ignora el tabulador y queda //malo.com
		{"/\n/malo.com", "/"},
		{"malo.com", "/"},
	}
	for _, c := range casos {
		campos := credencialesDeAna()
		campos.Set("siguiente", c.siguiente)

		rec := pedido(h, "/login", campos)

		if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != c.destino {
			t.Errorf("siguiente %q: se esperaba 303 a %q y se obtuvo %d a %q", c.siguiente, c.destino, rec.Code, rec.Header().Get("Location"))
		}
	}
	// el formulario tampoco devuelve un siguiente que no sea interno
	if strings.Contains(pedido(h, "/login?siguiente=https://malo.com", nil).Body.String(), "malo.com") {
		t.Error("el formulario no tendría que ofrecer un siguiente externo")
	}
	verificarContiene(t, pedido(h, "/login?siguiente=/proyectos", nil).Body.String(), `name="siguiente" value="/proyectos"`)
}

// CA-05.1 · RN6 y sección 7: los errores del ingreso, con su código y mensaje, sin cookie y sin decir cuál dato falló
func TestIngresar_Errores(t *testing.T) {
	casos := []struct {
		nombre  string
		login   *servicioDeLoginSimulado
		campos  url.Values
		codigo  int
		mensaje string
	}{
		{"contraseña incorrecta", &servicioDeLoginSimulado{}, url.Values{"email": {"ana@mail.com"}, "clave": {"otra-clave-9"}}, http.StatusUnauthorized, mensajeCredenciales},
		{"email que no existe", &servicioDeLoginSimulado{}, url.Values{"email": {"nadie@mail.com"}, "clave": {"clave-segura-1"}}, http.StatusUnauthorized, mensajeCredenciales},
		{"sin completar", &servicioDeLoginSimulado{}, url.Values{"email": {""}, "clave": {""}}, http.StatusUnauthorized, mensajeCredenciales},
		{"email sin confirmar", &servicioDeLoginSimulado{sinConfirmar: true}, credencialesDeAna(), http.StatusUnauthorized, mensajeSinConfirmar},
		{"servicio caído", &servicioDeLoginSimulado{falla: auth.ErrServicioNoDisponible}, credencialesDeAna(), http.StatusServiceUnavailable, mensajeSinServicio},
		{"demasiados intentos", &servicioDeLoginSimulado{falla: auth.ErrDemasiadosIntentos}, credencialesDeAna(), http.StatusTooManyRequests, mensajeDemasiadosIntentos},
	}
	for _, c := range casos {
		rec := pedido(routerConLogin(c.login), "/login", c.campos)

		if rec.Code != c.codigo {
			t.Errorf("%s: código = %d, se esperaba %d", c.nombre, rec.Code, c.codigo)
		}
		verificarContiene(t, rec.Body.String(), c.mensaje)
		if cookieDeSesion(rec) != nil {
			t.Errorf("%s: no se tendría que guardar ninguna cookie", c.nombre)
		}
		if strings.Contains(rec.Body.String(), "clave-segura-1") || strings.Contains(rec.Body.String(), "otra-clave-9") {
			t.Errorf("%s: la página no tendría que devolver la contraseña", c.nombre)
		}
	}
}

// El email escrito se conserva cuando el ingreso falla
func TestIngresar_Error_ConservaElEmail(t *testing.T) {
	rec := pedido(routerConLogin(&servicioDeLoginSimulado{}), "/login", url.Values{"email": {"ana@mail.com"}, "clave": {"mal"}})

	verificarContiene(t, rec.Body.String(), `value="ana@mail.com"`)
}

// CA-05.1 · la pantalla de registro
func TestFormularioDeRegistro_MuestraLosCampos(t *testing.T) {
	rec := pedido(routerConLogin(&servicioDeLoginSimulado{}), "/registro", nil)

	verificarContiene(t, rec.Body.String(), "<title>Crear cuenta · Software Metrics</title>",
		`action="/registro"`, `method="post"`, `name="email"`, `name="clave"`, `name="repeticion"`, `type="password"`, `href="/login"`)
}

// CA-05.1 · RN5: registrarse bien vuelve al login con el mensaje, según pida o no confirmar el email
func TestRegistrar_DatosValidos_VuelveAlLogin(t *testing.T) {
	campos := url.Values{"email": {" Beto@Mail.com "}, "clave": {"clave-segura-1"}, "repeticion": {"clave-segura-1"}}
	casos := []struct {
		nombre  string
		login   *servicioDeLoginSimulado
		destino string
	}{
		{"cuenta lista", &servicioDeLoginSimulado{}, "/login?registrado=1"},
		{"pide confirmar el email", &servicioDeLoginSimulado{pideConfirmar: true}, "/login?confirmar=1"},
	}
	for _, c := range casos {
		rec := pedido(routerConLogin(c.login), "/registro", campos)

		if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != c.destino {
			t.Errorf("%s: se esperaba 303 a %q y se obtuvo %d a %q", c.nombre, c.destino, rec.Code, rec.Header().Get("Location"))
		}
		if len(c.login.registradas) != 1 || c.login.registradas[0].Email != "beto@mail.com" {
			t.Errorf("%s: se esperaba registrar a beto@mail.com y se obtuvo %+v", c.nombre, c.login.registradas)
		}
		if cookieDeSesion(rec) != nil {
			t.Errorf("%s: registrarse no inicia sesión", c.nombre)
		}
	}
}

// CA-05.1 · RN5: los errores del registro aparecen todos juntos, cada uno en su campo, con el email escrito y sin las contraseñas
func TestRegistrar_DatosInvalidos_MuestraLosErroresYConservaElEmail(t *testing.T) {
	login := &servicioDeLoginSimulado{}

	rec := pedido(routerConLogin(login), "/registro", url.Values{"email": {"ana@mail"}, "clave": {"corta12"}, "repeticion": {"otra-distinta"}})

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusUnprocessableEntity)
	}
	verificarContiene(t, rec.Body.String(), "el email no tiene un formato válido", "la contraseña tiene que tener entre 8 y 72 caracteres",
		"las contraseñas no coinciden", "is-invalid", `value="ana@mail"`)
	for _, secreto := range []string{`value="corta12"`, `value="otra-distinta"`} {
		if strings.Contains(rec.Body.String(), secreto) {
			t.Errorf("la página no tendría que devolver la contraseña: contiene %s", secreto)
		}
	}
	if len(login.registradas) != 0 {
		t.Error("con datos inválidos no se le tendría que haber pedido nada al servicio")
	}
}

// RN5 · un email que ya tiene cuenta se informa debajo del campo
func TestRegistrar_EmailYaRegistrado_LoMuestraEnElCampo(t *testing.T) {
	rec := pedido(routerConLogin(&servicioDeLoginSimulado{}), "/registro",
		url.Values{"email": {"ana@mail.com"}, "clave": {"otra-clave-9"}, "repeticion": {"otra-clave-9"}})

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusUnprocessableEntity)
	}
	verificarContiene(t, rec.Body.String(), "ya hay una cuenta con ese email")
}

// RN12 · si el servicio no responde o limita los intentos, el registro lo informa
func TestRegistrar_ServicioCaidoOLimitado(t *testing.T) {
	campos := url.Values{"email": {"beto@mail.com"}, "clave": {"clave-segura-1"}, "repeticion": {"clave-segura-1"}}
	casos := []struct {
		falla   error
		codigo  int
		mensaje string
	}{
		{auth.ErrServicioNoDisponible, http.StatusServiceUnavailable, mensajeSinServicio},
		{auth.ErrDemasiadosIntentos, http.StatusTooManyRequests, mensajeDemasiadosIntentos},
	}
	for _, c := range casos {
		rec := pedido(routerConLogin(&servicioDeLoginSimulado{falla: c.falla}), "/registro", campos)

		if rec.Code != c.codigo {
			t.Errorf("%v: código = %d, se esperaba %d", c.falla, rec.Code, c.codigo)
		}
		verificarContiene(t, rec.Body.String(), c.mensaje, `value="beto@mail.com"`)
	}
}

// CA-05.1 · RN10: salir borra la cookie y vuelve al login con el aviso; funciona aunque la sesión ya no valga
func TestSalir_BorraLaCookieYVuelveAlLogin(t *testing.T) {
	h := routerConLogin(&servicioDeLoginSimulado{})
	for nombre, opcion := range map[string]func(*http.Request){"con sesión": conSesion("token-de-ana"), "con sesión vencida": conSesion("token-vencido")} {
		rec := pedido(h, "/salir", url.Values{}, opcion)

		if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/login?salio=1" {
			t.Errorf("%s: se esperaba 303 a /login?salio=1 y se obtuvo %d a %q", nombre, rec.Code, rec.Header().Get("Location"))
		}
		verificarCookieBorrada(t, rec)
	}
	if rec := get(h, "/salir"); rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("GET /salir: código %d, se esperaba %d (solo se sale con POST)", rec.Code, http.StatusMethodNotAllowed)
	}
}
