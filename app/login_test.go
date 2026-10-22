package app_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	neturl "net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/app"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/postgres"
)

const usuarioAna = "11111111-1111-4111-8111-111111111111"

// supabaseFalso es un Supabase Auth de mentira para probar la aplicación completa (TEC-05): publica el JWKS, crea cuentas
// con /signup y, en /token, entrega un token ES256 firmado de verdad con el emisor, la audiencia y el sub de la cuenta.
type supabaseFalso struct {
	server *httptest.Server
	clave  *ecdsa.PrivateKey

	mu      sync.Mutex
	cuentas map[string]string // email → contraseña
}

func nuevoSupabaseFalso(t *testing.T) *supabaseFalso {
	t.Helper()
	clave, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	s := &supabaseFalso{clave: clave, cuentas: map[string]string{}}
	b64 := func(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var cuerpo struct{ Email, Password string }
		_ = json.NewDecoder(r.Body).Decode(&cuerpo)
		s.mu.Lock()
		defer s.mu.Unlock()
		switch r.URL.RequestURI() {
		case "/auth/v1/.well-known/jwks.json":
			_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{
				"kty": "EC", "crv": "P-256", "alg": "ES256", "use": "sig", "kid": "clave-falsa",
				"x": b64(clave.PublicKey.X.FillBytes(make([]byte, 32))), "y": b64(clave.PublicKey.Y.FillBytes(make([]byte, 32))),
			}}})
		case "/auth/v1/signup":
			s.cuentas[cuerpo.Email] = cuerpo.Password
			_ = json.NewEncoder(w).Encode(map[string]string{"id": usuarioAna, "email": cuerpo.Email}) // sin sesión: pide confirmar
		case "/auth/v1/token?grant_type=password":
			if clave, existe := s.cuentas[cuerpo.Email]; !existe || clave != cuerpo.Password {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"code":400,"error_code":"invalid_credentials","msg":"Invalid login credentials"}`))
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": s.token(time.Hour), "token_type": "bearer", "expires_in": 3600,
				"user": map[string]string{"id": usuarioAna, "email": cuerpo.Email},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(s.server.Close)
	return s
}

// token firma un token de Ana con la clave del JWKS; con un plazo negativo ya está vencido.
func (s *supabaseFalso) token(vigencia time.Duration) string {
	return s.tokenFirmadoCon(s.clave, vigencia)
}

func (s *supabaseFalso) tokenFirmadoCon(clave *ecdsa.PrivateKey, vigencia time.Duration) string {
	t := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"iss": s.server.URL + "/auth/v1", "aud": "authenticated", "sub": usuarioAna, "email": "ana@mail.com",
		"exp": time.Now().Add(vigencia).Unix(),
	})
	t.Header["kid"] = "clave-falsa"
	firmado, err := t.SignedString(clave)
	if err != nil {
		panic(err)
	}
	return firmado
}

func (s *supabaseFalso) entorno(base string) func(string) string {
	variables := map[string]string{"DATABASE_URL": base, "SUPABASE_URL": s.server.URL, "SUPABASE_ANON_KEY": "clave-anon"}
	return func(nombre string) string { return variables[nombre] }
}

func appConLogin(t *testing.T, s *supabaseFalso, base string) http.Handler {
	t.Helper()
	h, err := app.NewHandlerDesdeEntorno(s.entorno(base))
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	return h
}

func pedir(h http.Handler, metodo, ruta, cookie string, campos neturl.Values) *httptest.ResponseRecorder {
	req := httptest.NewRequest(metodo, ruta, strings.NewReader(campos.Encode()))
	if metodo == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: "sesion", Value: cookie})
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

// cookieDeSesion devuelve el valor de la cookie "sesion" que puso la respuesta.
func cookieDeSesion(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	for _, c := range rec.Result().Cookies() {
		if c.Name == "sesion" {
			return c.Value
		}
	}
	t.Fatalf("la respuesta (código %d) no puso la cookie sesion", rec.Code)
	return ""
}

func cuentaDeAna() neturl.Values {
	return neturl.Values{"email": {"ana@mail.com"}, "clave": {"clave-segura-1"}, "repeticion": {"clave-segura-1"}}
}

// RN1 · sin las variables de Supabase la aplicación arranca como siempre, sin pedir login
func TestNewHandlerDesdeEntorno_SinVariablesDeSupabase_NoPideLogin(t *testing.T) {
	h, err := app.NewHandlerDesdeEntorno(func(string) string { return "" })
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if rec := pedir(h, http.MethodGet, "/proyectos", "", nil); rec.Code != http.StatusOK {
		t.Errorf("/proyectos: código %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	if rec := pedir(h, http.MethodGet, "/login", "", nil); rec.Code != http.StatusNotFound {
		t.Errorf("/login: código %d, se esperaba %d", rec.Code, http.StatusNotFound)
	}
}

// RN1 · con una sola variable la aplicación no arranca y dice cuál falta
func TestNewHandlerDesdeEntorno_ConfiguracionIncompleta_NoArranca(t *testing.T) {
	casos := map[string]map[string]string{
		"falta SUPABASE_ANON_KEY": {"SUPABASE_URL": "https://abc.supabase.co"},
		"falta SUPABASE_URL":      {"SUPABASE_ANON_KEY": "la-clave-anon"},
	}
	for esperado, variables := range casos {
		_, err := app.NewHandlerDesdeEntorno(func(nombre string) string { return variables[nombre] })

		if err == nil || !strings.Contains(err.Error(), esperado) {
			t.Errorf("se esperaba el error %q y se obtuvo %v", esperado, err)
		}
	}
}

// Es el objetivo de TEC-05 de punta a punta, con el verificador y el cliente de Supabase de verdad: sin sesión no se entra,
// se registra, se ingresa, se entra y se sale; y un token falso o vencido no sirve
func TestNewHandlerDesdeEntorno_ConSupabase_RegistrarIngresarEntrarYSalir(t *testing.T) {
	s := nuevoSupabaseFalso(t)
	h := appConLogin(t, s, "")

	if rec := pedir(h, http.MethodGet, "/proyectos", "", nil); rec.Code != http.StatusUnauthorized {
		t.Fatalf("sin sesión: código %d, se esperaba %d", rec.Code, http.StatusUnauthorized)
	}
	if rec := pedir(h, http.MethodGet, "/health", "", nil); rec.Code != http.StatusOK {
		t.Errorf("/health: código %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	if rec := pedir(h, http.MethodPost, "/registro", "", cuentaDeAna()); rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != "/login?confirmar=1" {
		t.Fatalf("registro: se esperaba 303 a /login?confirmar=1 y se obtuvo %d a %q", rec.Code, rec.Header().Get("Location"))
	}

	ingreso := pedir(h, http.MethodPost, "/login", "", neturl.Values{"email": {"Ana@Mail.com"}, "clave": {"clave-segura-1"}})
	if ingreso.Code != http.StatusSeeOther {
		t.Fatalf("ingreso: código %d, se esperaba %d", ingreso.Code, http.StatusSeeOther)
	}
	cookie := cookieDeSesion(t, ingreso)

	conSesion := pedir(h, http.MethodGet, "/proyectos", cookie, nil)
	if conSesion.Code != http.StatusOK || !strings.Contains(conSesion.Body.String(), "ana@mail.com") {
		t.Errorf("con sesión: código %d, se esperaba %d con el email en la barra", conSesion.Code, http.StatusOK)
	}
	if rec := pedir(h, http.MethodPost, "/salir", cookie, neturl.Values{}); rec.Code != http.StatusSeeOther {
		t.Errorf("salir: código %d, se esperaba %d", rec.Code, http.StatusSeeOther)
	}

	otraClave, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	for nombre, token := range map[string]string{
		"firmado con otra clave": s.tokenFirmadoCon(otraClave, time.Hour),
		"vencido":                s.token(-time.Second),
	} {
		if rec := pedir(h, http.MethodGet, "/proyectos", token, nil); rec.Code != http.StatusUnauthorized {
			t.Errorf("token %s: código %d, se esperaba %d", nombre, rec.Code, http.StatusUnauthorized)
		}
	}
	if rec := pedir(h, http.MethodPost, "/login", "", neturl.Values{"email": {"ana@mail.com"}, "clave": {"otra-clave-9"}}); rec.Code != http.StatusUnauthorized {
		t.Errorf("contraseña incorrecta: código %d, se esperaba %d", rec.Code, http.StatusUnauthorized)
	}
}

// CA-05.3 · al ingresar, la cuenta queda vinculada en la base con los integrantes que tienen su email
func TestNewHandlerDesdeEntorno_ConPostgres_AlIngresarVinculaALosIntegrantes(t *testing.T) {
	url := urlConEsquemaMigrado(t)
	s := nuevoSupabaseFalso(t)
	h := appConLogin(t, s, url)
	pedir(h, http.MethodPost, "/registro", "", cuentaDeAna())
	ingresar := func() string {
		rec := pedir(h, http.MethodPost, "/login", "", neturl.Values{"email": {"ana@mail.com"}, "clave": {"clave-segura-1"}})
		return cookieDeSesion(t, rec)
	}
	cookie := ingresar()
	proyecto := neturl.Values{"nombre": {"Con login"}, "descripcion": {""}, "fecha_inicio": {"2026-10-05"}, "fecha_fin": {"2026-11-01"}}
	if rec := pedir(h, http.MethodPost, "/proyectos", cookie, proyecto); rec.Code != http.StatusSeeOther {
		t.Fatalf("crear proyecto con sesión: código %d", rec.Code)
	}
	integrante := neturl.Values{"nombre": {"Ana"}, "email": {"ana@mail.com"}, "rol": {"ProductBuilder"}}
	if rec := pedir(h, http.MethodPost, "/proyectos/1/integrantes", cookie, integrante); rec.Code != http.StatusSeeOther {
		t.Fatalf("registrar integrante con sesión: código %d", rec.Code)
	}
	pool, err := postgres.Conectar(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	vinculo := func() string {
		var id string
		if err := pool.QueryRow(context.Background(), `SELECT COALESCE(auth_user_id::text, '') FROM integrantes WHERE email = 'ana@mail.com'`).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	if id := vinculo(); id != "" {
		t.Fatalf("Ana se registró después de ingresar: todavía no tendría que estar vinculada y tiene %q", id)
	}

	ingresar()

	if id := vinculo(); id != usuarioAna {
		t.Errorf("después de ingresar Ana tendría que estar vinculada con %s y tiene %q", usuarioAna, id)
	}
}
