package server

import (
	"cmp"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/auth"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/account"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// nombreCookie es la cookie con el token de acceso de Supabase (TEC-05, RN4).
const nombreCookie = "sesion"

// Mensajes para el usuario (secciones 3 y 7 de specs/TEC-05-autenticacion.md).
const (
	mensajeSinSesion          = "Iniciá sesión para continuar."
	mensajeSesionVencida      = "Tu sesión venció, iniciá sesión de nuevo."
	mensajeCuentaCreada       = "Cuenta creada. Ya podés iniciar sesión."
	mensajeConfirmarEmail     = "Te enviamos un email para confirmar la cuenta. Confirmala y después iniciá sesión."
	mensajeSalio              = "Cerraste sesión."
	mensajeCredenciales       = "El email o la contraseña no son correctos."
	mensajeEmailSinConfirmar  = "Confirmá tu email antes de iniciar sesión."
	mensajeServicioCaido      = "No pudimos conectar con el servicio de inicio de sesión. Probá de nuevo en unos minutos."
	mensajeDemasiadosIntentos = "Demasiados intentos. Esperá un momento y volvé a probar."
	mensajeFalloInesperado    = "No se pudo completar la operación. Probá de nuevo en unos minutos."
)

// cuentas agrupa los handlers de registro, ingreso y salida (TEC-05).
type cuentas struct {
	casos *service.Cuentas
}

// formLogin son los datos de la página login.html.
type formLogin struct {
	Email     string
	Siguiente string // ruta interna a la que volver después de ingresar (RN9); vacío si no hay
	Exito     string // aviso verde: "Cuenta creada…", "Cerraste sesión."
	Aviso     string // aviso amarillo: "Iniciá sesión para continuar."
	Error     string // aviso rojo: el ingreso falló
}

// formRegistro son los datos de la página registro.html. Nunca lleva las contraseñas.
type formRegistro struct {
	Email        string
	Errores      map[string]string
	ErrorGeneral string
}

// formularioDeLogin muestra el formulario de ingreso, con el aviso que corresponda si se llegó desde el registro o al salir.
func (h cuentas) formularioDeLogin(w http.ResponseWriter, r *http.Request) {
	form := formLogin{Siguiente: rutaInterna(r.URL.Query().Get("siguiente"))}
	switch {
	case r.URL.Query().Get("registrado") == "1":
		form.Exito = mensajeCuentaCreada
	case r.URL.Query().Get("confirmar") == "1":
		form.Exito = mensajeConfirmarEmail
	case r.URL.Query().Get("salio") == "1":
		form.Exito = mensajeSalio
	}
	renderLogin(w, http.StatusOK, form)
}

// ingresar recibe el formulario de ingreso. Si las credenciales son correctas guarda la cookie y vuelve (303) a donde se
// estaba; si no, muestra el formulario con el motivo y sin cookie (RN6 y RN12).
func (h cuentas) ingresar(w http.ResponseWriter, r *http.Request) {
	form := formLogin{Email: r.PostFormValue("email"), Siguiente: rutaInterna(r.PostFormValue("siguiente"))}
	sesion, err := h.casos.Ingresar(r.Context(), form.Email, r.PostFormValue("clave"))

	switch {
	case err == nil:
		http.SetCookie(w, cookieDeSesion(r, sesion))
		http.Redirect(w, r, cmp.Or(form.Siguiente, "/"), http.StatusSeeOther)
	case errors.Is(err, auth.ErrCredencialesIncorrectas):
		form.Error = mensajeCredenciales
		renderLogin(w, http.StatusUnauthorized, form)
	case errors.Is(err, auth.ErrEmailSinConfirmar):
		form.Error = mensajeEmailSinConfirmar
		renderLogin(w, http.StatusUnauthorized, form)
	case errors.Is(err, auth.ErrDemasiadosIntentos):
		form.Error = mensajeDemasiadosIntentos
		renderLogin(w, http.StatusTooManyRequests, form)
	case errors.Is(err, auth.ErrServicioNoDisponible):
		form.Error = mensajeServicioCaido
		renderLogin(w, http.StatusServiceUnavailable, form)
	default:
		log.Printf("ingresar: %v", err)
		form.Error = mensajeFalloInesperado
		renderLogin(w, http.StatusInternalServerError, form)
	}
}

// formularioDeRegistro muestra el formulario vacío para crear una cuenta.
func (h cuentas) formularioDeRegistro(w http.ResponseWriter, _ *http.Request) {
	renderRegistro(w, http.StatusOK, formRegistro{})
}

// registrar recibe el formulario de registro. Si sale bien vuelve (303) al login con el aviso; si no, muestra el formulario
// con los errores de cada campo y el email escrito (RN5).
func (h cuentas) registrar(w http.ResponseWriter, r *http.Request) {
	form := formRegistro{Email: r.PostFormValue("email")}
	resultado, err := h.casos.Registrar(r.Context(), account.DatosRegistro{
		Email:      form.Email,
		Clave:      r.PostFormValue("clave"),
		Repeticion: r.PostFormValue("repeticion"),
	})

	var errs domain.ErroresValidacion
	switch {
	case err == nil && resultado.RequiereConfirmacion:
		http.Redirect(w, r, "/login?confirmar=1", http.StatusSeeOther)
	case err == nil:
		http.Redirect(w, r, "/login?registrado=1", http.StatusSeeOther)
	case errors.As(err, &errs):
		form.Errores = make(map[string]string, len(errs))
		for campo, e := range errs {
			form.Errores[campo] = e.Error()
		}
		renderRegistro(w, http.StatusUnprocessableEntity, form)
	case errors.Is(err, auth.ErrDemasiadosIntentos):
		form.ErrorGeneral = mensajeDemasiadosIntentos
		renderRegistro(w, http.StatusTooManyRequests, form)
	case errors.Is(err, auth.ErrServicioNoDisponible):
		form.ErrorGeneral = mensajeServicioCaido
		renderRegistro(w, http.StatusServiceUnavailable, form)
	default:
		log.Printf("registrar cuenta: %v", err)
		form.ErrorGeneral = mensajeFalloInesperado
		renderRegistro(w, http.StatusInternalServerError, form)
	}
}

// salir borra la cookie y vuelve al login (RN10). No revoca el token en Supabase: se verifica localmente.
func (h cuentas) salir(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, cookieBorrada(r))
	http.Redirect(w, r, "/login?salio=1", http.StatusSeeOther)
}

// requiereSesion rechaza con 401 y la página de login todo pedido sin una sesión válida, antes de que llegue a un handler
// (RN3). Una cookie inválida o vencida se borra en la misma respuesta. Con sesión válida el usuario queda en el contexto
// del pedido (auth.IdentidadDe) y en la barra de navegación de las páginas.
func requiereSesion(v auth.Verificador) func(http.Handler) http.Handler {
	return func(siguiente http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(nombreCookie)
			if err != nil {
				rechazar(w, r, mensajeSinSesion)
				return
			}
			identidad, err := v.Verificar(r.Context(), cookie.Value)
			if err != nil {
				http.SetCookie(w, cookieBorrada(r))
				if errors.Is(err, auth.ErrTokenVencido) {
					rechazar(w, r, mensajeSesionVencida)
				} else {
					rechazar(w, r, mensajeSinSesion)
				}
				return
			}
			siguiente.ServeHTTP(respuestaConUsuario{ResponseWriter: w, email: identidad.Email}, r.WithContext(auth.ConIdentidad(r.Context(), identidad)))
		})
	}
}

// rechazar responde 401 con la página de login. Si el pedido era un GET, ofrece volver a esa ruta después de ingresar;
// un POST no se puede repetir, así que no se recuerda.
func rechazar(w http.ResponseWriter, r *http.Request, motivo string) {
	form := formLogin{Aviso: motivo}
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		form.Siguiente = rutaInterna(r.URL.RequestURI())
	}
	renderLogin(w, http.StatusUnauthorized, form)
}

// respuestaConUsuario es la respuesta de un pedido con sesión válida: render le pone el email del usuario a la barra de
// navegación sin que cada handler tenga que pasarlo.
type respuestaConUsuario struct {
	http.ResponseWriter
	email string
}

// Unwrap deja que http.ResponseController llegue a la respuesta original.
func (r respuestaConUsuario) Unwrap() http.ResponseWriter { return r.ResponseWriter }

func renderLogin(w http.ResponseWriter, codigo int, form formLogin) {
	render(w, codigo, "login.html", Pagina{Titulo: "Iniciar sesión", Datos: form})
}

func renderRegistro(w http.ResponseWriter, codigo int, form formRegistro) {
	render(w, codigo, "registro.html", Pagina{Titulo: "Crear cuenta", Datos: form})
}

// cookieDeSesion arma la cookie con el token (RN4): HttpOnly, SameSite=Lax y de Path=/; Secure si el pedido llegó por HTTPS;
// vence cuando vence el token.
func cookieDeSesion(r *http.Request, sesion auth.Sesion) *http.Cookie {
	return &http.Cookie{
		Name:     nombreCookie,
		Value:    sesion.Token,
		Path:     "/",
		Expires:  sesion.Vence,
		HttpOnly: true,
		Secure:   porHTTPS(r),
		SameSite: http.SameSiteLaxMode,
	}
}

// cookieBorrada es la cookie de sesión vencida, que le indica al navegador que la borre. Lleva los mismos atributos que la
// original para que la reemplace.
func cookieBorrada(r *http.Request) *http.Cookie {
	return &http.Cookie{
		Name:     nombreCookie,
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   porHTTPS(r),
		SameSite: http.SameSiteLaxMode,
	}
}

// porHTTPS dice si el pedido llegó por HTTPS, también detrás de un proxy como el de Vercel.
func porHTTPS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	primero, _, _ := strings.Cut(r.Header.Get("X-Forwarded-Proto"), ",")
	return strings.EqualFold(strings.TrimSpace(primero), "https")
}

// rutaInterna devuelve la ruta si es una dirección interna de la aplicación y "" si no (RN9): tiene que empezar con "/" y
// no con "//" ni con "/\", y no puede tener caracteres de control (los navegadores ignoran el tabulador y el salto de línea,
// así que "/\t/sitio.com" terminaría siendo "//sitio.com"). Evita que el login redirija a otro sitio.
func rutaInterna(ruta string) string {
	if !strings.HasPrefix(ruta, "/") || strings.HasPrefix(ruta, "//") || strings.HasPrefix(ruta, `/\`) {
		return ""
	}
	for _, c := range ruta {
		if c < ' ' || c == 0x7f {
			return ""
		}
	}
	return ruta
}
