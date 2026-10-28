package auth_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/auth"
)

const (
	usuarioAna = "11111111-1111-4111-8111-111111111111"
	kidEC      = "clave-ec"
	kidRSA     = "clave-rsa"
)

// supabaseFalso hace de Supabase Auth para los tests del verificador: publica el JWKS con una clave ES256 y otra RS256, y
// firma tokens con ellas. Con caido en true el JWKS responde 500, como si Supabase no anduviera.
type supabaseFalso struct {
	t           *testing.T
	server      *httptest.Server
	ec          *ecdsa.PrivateKey
	rsa         *rsa.PrivateKey
	caido       atomic.Bool
	pedidosJWKS atomic.Int32
}

func nuevoSupabaseFalso(t *testing.T) *supabaseFalso {
	t.Helper()
	ec, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	s := &supabaseFalso{t: t, ec: ec, rsa: rsaKey}
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/v1/.well-known/jwks.json" {
			http.NotFound(w, r)
			return
		}
		s.pedidosJWKS.Add(1)
		if s.caido.Load() {
			http.Error(w, "caído", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{s.jwkEC(), s.jwkRSA()}})
	}))
	t.Cleanup(s.server.Close)
	return s
}

func (s *supabaseFalso) config() auth.Config {
	return auth.Config{URL: s.server.URL, AnonKey: "clave-anon"}
}

func b64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

func (s *supabaseFalso) jwkEC() map[string]string {
	// Forma sin comprimir: 0x04, X (32 bytes) e Y (32 bytes).
	p, err := s.ec.PublicKey.Bytes()
	if err != nil {
		s.t.Fatalf("clave pública EC: %v", err)
	}
	return map[string]string{"kty": "EC", "crv": "P-256", "alg": "ES256", "use": "sig", "kid": kidEC,
		"x": b64(p[1:33]), "y": b64(p[33:65])}
}

func (s *supabaseFalso) jwkRSA() map[string]string {
	p := s.rsa.PublicKey
	return map[string]string{"kty": "RSA", "alg": "RS256", "use": "sig", "kid": kidRSA,
		"n": b64(p.N.Bytes()), "e": b64(big.NewInt(int64(p.E)).Bytes())}
}

// claimsValidos son los de un token que Supabase le daría a Ana; cada test cambia lo que quiere probar.
func (s *supabaseFalso) claimsValidos() jwt.MapClaims {
	return jwt.MapClaims{
		"iss":   s.server.URL + "/auth/v1",
		"aud":   "authenticated",
		"sub":   usuarioAna,
		"email": "ana@mail.com",
		"role":  "authenticated",
		"exp":   time.Now().Add(time.Hour).Unix(),
	}
}

func (s *supabaseFalso) firmar(metodo jwt.SigningMethod, kid string, clave any, claims jwt.MapClaims) string {
	s.t.Helper()
	token := jwt.NewWithClaims(metodo, claims)
	token.Header["kid"] = kid
	firmado, err := token.SignedString(clave)
	if err != nil {
		s.t.Fatalf("firmar el token: %v", err)
	}
	return firmado
}

// tokenEC firma con la clave ES256 del JWKS, como Supabase.
func (s *supabaseFalso) tokenEC(claims jwt.MapClaims) string {
	return s.firmar(jwt.SigningMethodES256, kidEC, s.ec, claims)
}

func verificadorDe(t *testing.T, s *supabaseFalso, opciones ...auth.OpcionVerificador) auth.Verificador {
	t.Helper()
	v, err := auth.NuevoVerificador(t.Context(), s.config(), opciones...)
	if err != nil {
		t.Fatalf("NuevoVerificador: %v", err)
	}
	return v
}

// CA-05.2 · RN2: un token firmado con la clave del proyecto, vigente y con el emisor, la audiencia y el sub correctos, vale
func TestVerificar_TokenValido_DevuelveElUsuario(t *testing.T) {
	s := nuevoSupabaseFalso(t)
	v := verificadorDe(t, s)

	i, err := v.Verificar(t.Context(), s.tokenEC(s.claimsValidos()))

	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if i != (auth.Identidad{ID: usuarioAna, Email: "ana@mail.com"}) {
		t.Errorf("se esperaba a Ana y se obtuvo %+v", i)
	}
}

// RN2 · también se aceptan los tokens firmados con RS256, y el email se usa en minúsculas
func TestVerificar_TokenRS256_Vale(t *testing.T) {
	s := nuevoSupabaseFalso(t)
	v := verificadorDe(t, s)
	claims := s.claimsValidos()
	claims["email"] = " Ana@Mail.com "

	i, err := v.Verificar(t.Context(), s.firmar(jwt.SigningMethodRS256, kidRSA, s.rsa, claims))

	if err != nil || i.Email != "ana@mail.com" {
		t.Errorf("se esperaba a ana@mail.com sin error y se obtuvo %+v, %v", i, err)
	}
}

// RN2 · un token sin email igual identifica al usuario por su sub
func TestVerificar_TokenSinEmail_Vale(t *testing.T) {
	s := nuevoSupabaseFalso(t)
	v := verificadorDe(t, s)
	claims := s.claimsValidos()
	delete(claims, "email")

	i, err := v.Verificar(t.Context(), s.tokenEC(claims))

	if err != nil || i.ID != usuarioAna || i.Email != "" {
		t.Errorf("se esperaba el usuario sin email y se obtuvo %+v, %v", i, err)
	}
}

// CA-05.2 · RN3: un token vencido (hasta por un segundo) se distingue de uno inválido para pedir ingresar de nuevo
func TestVerificar_TokenVencido_DevuelveErrTokenVencido(t *testing.T) {
	s := nuevoSupabaseFalso(t)
	v := verificadorDe(t, s)
	claims := s.claimsValidos()
	claims["exp"] = time.Now().Add(-time.Second).Unix()

	_, err := v.Verificar(t.Context(), s.tokenEC(claims))

	if !errors.Is(err, auth.ErrTokenVencido) {
		t.Errorf("se esperaba auth.ErrTokenVencido y se obtuvo %v", err)
	}
}

// CA-05.2 · RN2 y RN3: todo lo demás que no vale es ErrTokenInvalido
func TestVerificar_TokensQueNoValen_DevuelvenErrTokenInvalido(t *testing.T) {
	s := nuevoSupabaseFalso(t)
	v := verificadorDe(t, s)
	otraEC, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	con := func(cambio func(jwt.MapClaims)) string {
		claims := s.claimsValidos()
		cambio(claims)
		return s.tokenEC(claims)
	}
	// HS256 con la clave pública como secreto: el ataque clásico de confusión de algoritmos
	hmacConLaPublica, err := jwt.NewWithClaims(jwt.SigningMethodHS256, s.claimsValidos()).SignedString([]byte(s.jwkEC()["x"]))
	if err != nil {
		t.Fatal(err)
	}
	sinFirma := jwt.NewWithClaims(jwt.SigningMethodNone, s.claimsValidos())
	sinFirma.Header["kid"] = kidEC
	tokenSinFirma, err := sinFirma.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatal(err)
	}

	casos := map[string]string{
		"vacío":                           "",
		"un texto que no es un token":     "esto-no-es-un-token",
		"firmado con otra clave":          s.firmar(jwt.SigningMethodES256, kidEC, otraEC, s.claimsValidos()),
		"con un kid que no está":          s.firmar(jwt.SigningMethodES256, "otro-kid", s.ec, s.claimsValidos()),
		"sin firma (alg none)":            tokenSinFirma,
		"HS256":                           hmacConLaPublica,
		"de otro emisor":                  con(func(c jwt.MapClaims) { c["iss"] = "https://otro.supabase.co/auth/v1" }),
		"sin emisor":                      con(func(c jwt.MapClaims) { delete(c, "iss") }),
		"con otra audiencia":              con(func(c jwt.MapClaims) { c["aud"] = "anon" }),
		"sin audiencia":                   con(func(c jwt.MapClaims) { delete(c, "aud") }),
		"con un sub que no es un UUID":    con(func(c jwt.MapClaims) { c["sub"] = "ana" }),
		"sin sub":                         con(func(c jwt.MapClaims) { delete(c, "sub") }),
		"sin vencimiento":                 con(func(c jwt.MapClaims) { delete(c, "exp") }),
		"que todavía no empezó (nbf)":     con(func(c jwt.MapClaims) { c["nbf"] = time.Now().Add(time.Hour).Unix() }),
		"con el payload alterado":         alterarPayload(t, s.tokenEC(s.claimsValidos())),
		"con una parte de más":            s.tokenEC(s.claimsValidos()) + ".x",
		"con espacios de más (no es JWT)": " " + s.tokenEC(s.claimsValidos()),
	}
	for nombre, token := range casos {
		_, err := v.Verificar(t.Context(), token)

		if !errors.Is(err, auth.ErrTokenInvalido) {
			t.Errorf("%s: se esperaba auth.ErrTokenInvalido y se obtuvo %v", nombre, err)
		}
		if errors.Is(err, auth.ErrTokenVencido) {
			t.Errorf("%s: no está vencido, es inválido", nombre)
		}
	}
}

// alterarPayload cambia el sub del token sin volver a firmarlo.
func alterarPayload(t *testing.T, token string) string {
	t.Helper()
	partes := strings.Split(token, ".")
	payload, err := base64.RawURLEncoding.DecodeString(partes[1])
	if err != nil {
		t.Fatal(err)
	}
	partes[1] = b64([]byte(strings.Replace(string(payload), usuarioAna, "99999999-9999-4999-8999-999999999999", 1)))
	return strings.Join(partes, ".")
}

// RN12 · si no se pueden bajar las claves, los tokens se rechazan (falla cerrada) pero la aplicación no se cae, y cuando
// Supabase vuelve los tokens vuelven a valer sin reiniciar
func TestVerificar_ConElJWKSCaido_RechazaYSeRecuperaSolo(t *testing.T) {
	s := nuevoSupabaseFalso(t)
	s.caido.Store(true)
	// por defecto las claves se vuelven a pedir como mucho una vez por minuto; el test no puede esperar tanto
	v, err := auth.NuevoVerificador(t.Context(), s.config(), auth.ConEsperaEntreRefrescos(10*time.Millisecond))
	if err != nil {
		t.Fatalf("con Supabase caído la aplicación igual tiene que poder arrancar: %v", err)
	}
	token := s.tokenEC(s.claimsValidos())

	if _, err := v.Verificar(t.Context(), token); !errors.Is(err, auth.ErrTokenInvalido) {
		t.Errorf("sin claves se esperaba auth.ErrTokenInvalido y se obtuvo %v", err)
	}

	s.caido.Store(false)
	// el verificador vuelve a pedir las claves cuando llega un kid que no conoce (como mucho una vez por minuto)
	if _, err := v.Verificar(t.Context(), token); err != nil {
		t.Errorf("con Supabase de vuelta se esperaba que el token valga y se obtuvo %v", err)
	}
}

// RN2 · las claves se bajan una vez y se reutilizan: verificar no vuelve a llamar a Supabase
func TestVerificar_ReutilizaLasClavesSinVolverAPedirlas(t *testing.T) {
	s := nuevoSupabaseFalso(t)
	v := verificadorDe(t, s)
	token := s.tokenEC(s.claimsValidos())

	for range 5 {
		if _, err := v.Verificar(t.Context(), token); err != nil {
			t.Fatalf("no se esperaba error: %v", err)
		}
	}

	if pedidos := s.pedidosJWKS.Load(); pedidos != 1 {
		t.Errorf("se esperaba 1 pedido del JWKS y hubo %d", pedidos)
	}
}

// RN2 · un token con un kid desconocido no abre la puerta a pedir las claves a cada rato: se limita a una vez por minuto
func TestVerificar_UnKidDesconocidoNoPideLasClavesEnCadaIntento(t *testing.T) {
	s := nuevoSupabaseFalso(t)
	v := verificadorDe(t, s)
	desconocido := s.firmar(jwt.SigningMethodES256, "otro-kid", s.ec, s.claimsValidos())

	for range 10 {
		_, _ = v.Verificar(t.Context(), desconocido)
	}

	if pedidos := s.pedidosJWKS.Load(); pedidos > 2 {
		t.Errorf("se esperaban a lo sumo 2 pedidos del JWKS (el del arranque y uno por el kid) y hubo %d", pedidos)
	}
}
