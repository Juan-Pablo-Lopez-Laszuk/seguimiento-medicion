package auth_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/auth"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/account"
)

var _ auth.ServicioDeCuentas = (*auth.Supabase)(nil)

var (
	credencialesAna = account.Credenciales{Email: "ana@mail.com", Clave: "clave-segura-1"}
	ahoraFijo       = time.Date(2026, time.October, 22, 10, 0, 0, 0, time.UTC)
)

// pedidoRecibido es lo que le llegó al Supabase falso.
type pedidoRecibido struct {
	metodo string
	ruta   string // con la query
	apikey string
	tipo   string
	cuerpo map[string]string
}

// supabaseQueResponde levanta un Supabase falso que contesta siempre lo mismo y anota el último pedido.
func supabaseQueResponde(t *testing.T, estado int, cuerpo string) (*httptest.Server, *pedidoRecibido) {
	t.Helper()
	pedido := &pedidoRecibido{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bytes, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(bytes, &pedido.cuerpo)
		pedido.metodo, pedido.ruta = r.Method, r.URL.RequestURI()
		pedido.apikey, pedido.tipo = r.Header.Get("apikey"), r.Header.Get("Content-Type")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(estado)
		_, _ = io.WriteString(w, cuerpo)
	}))
	t.Cleanup(server.Close)
	return server, pedido
}

func clienteDe(server *httptest.Server, opciones ...auth.OpcionSupabase) *auth.Supabase {
	opciones = append([]auth.OpcionSupabase{auth.ConReloj(func() time.Time { return ahoraFijo })}, opciones...)
	return auth.NuevoSupabase(auth.Config{URL: server.URL, AnonKey: "clave-anon"}, opciones...)
}

const sesionDeAna = `{"access_token":"token-de-ana","token_type":"bearer","expires_in":28800,"refresh_token":"r",
 "user":{"id":"11111111-1111-4111-8111-111111111111","email":"Ana@Mail.com","aud":"authenticated"}}`

func errorDeSupabase(estado int, codigo, mensaje string) string {
	b, _ := json.Marshal(map[string]any{"code": estado, "error_code": codigo, "msg": mensaje})
	return string(b)
}

// CA-05.1 · RN5: registrar manda el email y la contraseña a /auth/v1/signup con la clave pública
func TestSupabaseRegistrar_ConSesion_NoRequiereConfirmacion(t *testing.T) {
	server, pedido := supabaseQueResponde(t, http.StatusOK, sesionDeAna)

	res, err := clienteDe(server).Registrar(t.Context(), credencialesAna)

	if err != nil || res.RequiereConfirmacion {
		t.Fatalf("se esperaba la cuenta creada sin confirmación y se obtuvo %+v, %v", res, err)
	}
	if pedido.metodo != http.MethodPost || pedido.ruta != "/auth/v1/signup" {
		t.Errorf("se esperaba POST /auth/v1/signup y se obtuvo %s %s", pedido.metodo, pedido.ruta)
	}
	if pedido.apikey != "clave-anon" || pedido.tipo != "application/json" {
		t.Errorf("se esperaba la clave anon y JSON y se obtuvo %q y %q", pedido.apikey, pedido.tipo)
	}
	if pedido.cuerpo["email"] != "ana@mail.com" || pedido.cuerpo["password"] != "clave-segura-1" {
		t.Errorf("el cuerpo del pedido es incorrecto: %v", pedido.cuerpo)
	}
}

// RN5 · si Supabase pide confirmar el email, responde el usuario sin sesión
func TestSupabaseRegistrar_ConConfirmacionDeEmail_LoAvisa(t *testing.T) {
	server, _ := supabaseQueResponde(t, http.StatusOK,
		`{"id":"11111111-1111-4111-8111-111111111111","email":"ana@mail.com","confirmation_sent_at":"2026-10-22T10:00:00Z"}`)

	res, err := clienteDe(server).Registrar(t.Context(), credencialesAna)

	if err != nil || !res.RequiereConfirmacion {
		t.Errorf("se esperaba que pida confirmar el email y se obtuvo %+v, %v", res, err)
	}
}

// Los errores de Supabase se traducen a los de la aplicación (sección 7 de la spec)
func TestSupabaseRegistrar_TraduceLosErroresDeSupabase(t *testing.T) {
	casos := []struct {
		nombre   string
		estado   int
		cuerpo   string
		esperado error
	}{
		{"email ya registrado", 422, errorDeSupabase(422, "user_already_exists", "User already registered"), auth.ErrEmailYaRegistrado},
		{"email ya registrado (otro código)", 422, errorDeSupabase(422, "email_exists", "A user with this email address has already been registered"), auth.ErrEmailYaRegistrado},
		{"email ya registrado (versión vieja)", 400, `{"msg":"User already registered"}`, auth.ErrEmailYaRegistrado},
		{"contraseña débil", 422, errorDeSupabase(422, "weak_password", "Password should be at least 6 characters."), auth.ErrClaveDebil},
		{"email rechazado", 422, errorDeSupabase(422, "email_address_invalid", "Email address is invalid"), auth.ErrEmailRechazado},
		{"demasiados intentos", 429, errorDeSupabase(429, "over_request_rate_limit", "Request rate limit reached"), auth.ErrDemasiadosIntentos},
		{"demasiados emails", 429, errorDeSupabase(429, "over_email_send_rate_limit", "email rate limit exceeded"), auth.ErrDemasiadosIntentos},
		{"Supabase con error", 500, `{"msg":"error interno"}`, auth.ErrServicioNoDisponible},
		{"puerta de enlace caída", 502, "<html>Bad gateway</html>", auth.ErrServicioNoDisponible},
		{"clave anon inválida", 401, `{"message":"Invalid API key"}`, auth.ErrServicioNoDisponible},
		{"registro deshabilitado", 422, errorDeSupabase(422, "signup_disabled", "Signups not allowed for this instance"), auth.ErrServicioNoDisponible},
		{"respuesta que no es JSON", 200, "no soy JSON", auth.ErrServicioNoDisponible},
	}
	for _, c := range casos {
		server, _ := supabaseQueResponde(t, c.estado, c.cuerpo)

		_, err := clienteDe(server).Registrar(t.Context(), credencialesAna)

		if !errors.Is(err, c.esperado) {
			t.Errorf("%s: se esperaba %v y se obtuvo %v", c.nombre, c.esperado, err)
		}
	}
}

// CA-05.1 · RN6: ingresar pide el token a /auth/v1/token con el grant_type password y devuelve la sesión y el usuario
func TestSupabaseIngresar_CredencialesCorrectas_DevuelveLaSesion(t *testing.T) {
	server, pedido := supabaseQueResponde(t, http.StatusOK, sesionDeAna)

	sesion, err := clienteDe(server).Ingresar(t.Context(), credencialesAna)

	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if sesion.Token != "token-de-ana" {
		t.Errorf("token = %q", sesion.Token)
	}
	if !sesion.Vence.Equal(ahoraFijo.Add(28800 * time.Second)) {
		t.Errorf("la sesión vence %v, se esperaba %v", sesion.Vence, ahoraFijo.Add(28800*time.Second))
	}
	if sesion.Usuario != (auth.Identidad{ID: "11111111-1111-4111-8111-111111111111", Email: "ana@mail.com"}) {
		t.Errorf("usuario = %+v", sesion.Usuario)
	}
	if pedido.metodo != http.MethodPost || pedido.ruta != "/auth/v1/token?grant_type=password" {
		t.Errorf("se esperaba POST /auth/v1/token?grant_type=password y se obtuvo %s %s", pedido.metodo, pedido.ruta)
	}
	if pedido.apikey != "clave-anon" || pedido.cuerpo["email"] != "ana@mail.com" || pedido.cuerpo["password"] != "clave-segura-1" {
		t.Errorf("el pedido es incorrecto: apikey %q, cuerpo %v", pedido.apikey, pedido.cuerpo)
	}
}

func TestSupabaseIngresar_TraduceLosErroresDeSupabase(t *testing.T) {
	casos := []struct {
		nombre   string
		estado   int
		cuerpo   string
		esperado error
	}{
		{"credenciales incorrectas", 400, errorDeSupabase(400, "invalid_credentials", "Invalid login credentials"), auth.ErrCredencialesIncorrectas},
		{"credenciales incorrectas (versión vieja)", 400, `{"error":"invalid_grant","error_description":"Invalid login credentials"}`, auth.ErrCredencialesIncorrectas},
		{"email sin confirmar", 400, errorDeSupabase(400, "email_not_confirmed", "Email not confirmed"), auth.ErrEmailSinConfirmar},
		{"email sin confirmar (versión vieja)", 400, `{"error":"invalid_grant","error_description":"Email not confirmed"}`, auth.ErrEmailSinConfirmar},
		{"demasiados intentos", 429, errorDeSupabase(429, "over_request_rate_limit", "Request rate limit reached"), auth.ErrDemasiadosIntentos},
		{"Supabase con error", 500, `{"msg":"error interno"}`, auth.ErrServicioNoDisponible},
		{"clave anon inválida", 401, `{"message":"Invalid API key"}`, auth.ErrServicioNoDisponible},
		{"sesión sin token", 200, `{"user":{"id":"11111111-1111-4111-8111-111111111111"}}`, auth.ErrServicioNoDisponible},
		{"usuario sin UUID", 200, `{"access_token":"t","expires_in":60,"user":{"id":"ana","email":"ana@mail.com"}}`, auth.ErrServicioNoDisponible},
		{"respuesta que no es JSON", 200, "no soy JSON", auth.ErrServicioNoDisponible},
		{"respuesta enorme", 200, strings.Repeat("x", 2<<20), auth.ErrServicioNoDisponible},
	}
	for _, c := range casos {
		server, _ := supabaseQueResponde(t, c.estado, c.cuerpo)

		_, err := clienteDe(server).Ingresar(t.Context(), credencialesAna)

		if !errors.Is(err, c.esperado) {
			t.Errorf("%s: se esperaba %v y se obtuvo %v", c.nombre, c.esperado, err)
		}
	}
}

// RN12 · sin conexión o sin respuesta a tiempo el servicio no está disponible
func TestSupabase_SinConexionOSinRespuestaATiempo_NoDisponible(t *testing.T) {
	cerrado, _ := supabaseQueResponde(t, http.StatusOK, sesionDeAna)
	cerrado.Close()
	lento := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(5 * time.Second):
		}
	}))
	t.Cleanup(lento.Close)
	conPlazoCorto := auth.ConClienteHTTP(&http.Client{Timeout: 50 * time.Millisecond})

	for nombre, server := range map[string]*httptest.Server{"sin conexión": cerrado, "sin respuesta a tiempo": lento} {
		if _, err := clienteDe(server, conPlazoCorto).Registrar(t.Context(), credencialesAna); !errors.Is(err, auth.ErrServicioNoDisponible) {
			t.Errorf("%s al registrar: se esperaba auth.ErrServicioNoDisponible y se obtuvo %v", nombre, err)
		}
		if _, err := clienteDe(server, conPlazoCorto).Ingresar(t.Context(), credencialesAna); !errors.Is(err, auth.ErrServicioNoDisponible) {
			t.Errorf("%s al ingresar: se esperaba auth.ErrServicioNoDisponible y se obtuvo %v", nombre, err)
		}
	}
}

// RN11 · los errores no llevan la contraseña, el token ni lo que Supabase contestó
func TestSupabase_LosErroresNoFiltranDatosSensibles(t *testing.T) {
	server, _ := supabaseQueResponde(t, http.StatusInternalServerError,
		`{"msg":"falló con la clave clave-segura-1 y el token token-de-ana","email":"ana@mail.com"}`)

	_, errRegistro := clienteDe(server).Registrar(t.Context(), credencialesAna)
	_, errIngreso := clienteDe(server).Ingresar(t.Context(), credencialesAna)

	for _, err := range []error{errRegistro, errIngreso} {
		for _, secreto := range []string{"clave-segura-1", "token-de-ana", "clave-anon"} {
			if err == nil || strings.Contains(err.Error(), secreto) {
				t.Errorf("el error %q no tendría que contener %q", err, secreto)
			}
		}
	}
}
