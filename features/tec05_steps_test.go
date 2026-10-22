package features

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"net/http"
	"net/http/httptest"
	neturl "net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/cucumber/godog"
	"github.com/golang-jwt/jwt/v5"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/app"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/auth"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/account"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// uuidDeEscenario le da a cada email el mismo UUID de usuario en todos los pasos, como si fuera el ID que Supabase le asignó.
func uuidDeEscenario(email string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(email))
	return fmt.Sprintf("%08x-1111-4111-8111-111111111111", h.Sum32())
}

// supabaseDeEscenario es un Supabase Auth de mentira: publica el JWKS con su clave, crea cuentas en /signup y entrega tokens
// ES256 firmados de verdad en /token. Con noResponde contesta 503 a todo lo que no sea el JWKS.
type supabaseDeEscenario struct {
	server        *httptest.Server
	clave         *ecdsa.PrivateKey
	pideConfirmar bool
	noResponde    bool

	mu      sync.Mutex
	cuentas map[string]string // email → contraseña
}

func nuevoSupabaseDeEscenario() *supabaseDeEscenario {
	clave, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	s := &supabaseDeEscenario{clave: clave, cuentas: map[string]string{}}
	b64 := func(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }
	s.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var cuerpo struct{ Email, Password string }
		_ = json.NewDecoder(r.Body).Decode(&cuerpo)
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.noResponde && !strings.HasSuffix(r.URL.Path, "jwks.json") {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		switch r.URL.RequestURI() {
		case "/auth/v1/.well-known/jwks.json":
			_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{
				"kty": "EC", "crv": "P-256", "alg": "ES256", "use": "sig", "kid": "clave-del-escenario",
				"x": b64(clave.X.FillBytes(make([]byte, 32))), "y": b64(clave.Y.FillBytes(make([]byte, 32))),
			}}})
		case "/auth/v1/signup":
			if _, existe := s.cuentas[cuerpo.Email]; existe {
				w.WriteHeader(http.StatusUnprocessableEntity)
				_, _ = w.Write([]byte(`{"code":422,"error_code":"user_already_exists","msg":"User already registered"}`))
				return
			}
			s.cuentas[cuerpo.Email] = cuerpo.Password
			usuario := map[string]string{"id": uuidDeEscenario(cuerpo.Email), "email": cuerpo.Email}
			if !s.pideConfirmar {
				usuario["access_token"] = "sesion-abierta-al-registrarse"
			}
			_ = json.NewEncoder(w).Encode(usuario)
		case "/auth/v1/token?grant_type=password":
			if clave, existe := s.cuentas[cuerpo.Email]; !existe || clave != cuerpo.Password {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"code":400,"error_code":"invalid_credentials","msg":"Invalid login credentials"}`))
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token": s.token(cuerpo.Email, nil), "token_type": "bearer", "expires_in": 3600,
				"user": map[string]string{"id": uuidDeEscenario(cuerpo.Email), "email": cuerpo.Email},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	return s
}

// token firma un token de Ana como lo haría Supabase; cambiar modifica los claims antes de firmar.
func (s *supabaseDeEscenario) token(email string, cambiar func(jwt.MapClaims)) string {
	claims := jwt.MapClaims{
		"iss": s.server.URL + "/auth/v1", "aud": "authenticated", "sub": uuidDeEscenario(email), "email": email,
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	if cambiar != nil {
		cambiar(claims)
	}
	return s.firmar(jwt.SigningMethodES256, s.clave, claims)
}

func (s *supabaseDeEscenario) firmar(metodo jwt.SigningMethod, clave any, claims jwt.MapClaims) string {
	t := jwt.NewWithClaims(metodo, claims)
	t.Header["kid"] = "clave-del-escenario"
	var firmado string
	var err error
	if metodo == jwt.SigningMethodNone {
		firmado, err = t.SignedString(jwt.UnsafeAllowNoneSignatureType)
	} else {
		firmado, err = t.SignedString(clave)
	}
	if err != nil {
		panic(err)
	}
	return firmado
}

// tokenDe arma el token que describe el escenario ("un token vencido", "un token con otra audiencia"...). Una cadena vacía
// es "sin cookie".
func (s *supabaseDeEscenario) tokenDe(descripcion string) (string, error) {
	ana := "ana@mail.com"
	switch descripcion {
	case "ausente":
		return "", nil
	case "un texto que no es un token":
		return "esto-no-es-un-token", nil
	case "un token vencido":
		return s.token(ana, func(c jwt.MapClaims) { c["exp"] = time.Now().Add(-time.Second).Unix() }), nil
	case "un token firmado con otra clave":
		otra, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		return s.firmar(jwt.SigningMethodES256, otra, jwt.MapClaims{
			"iss": s.server.URL + "/auth/v1", "aud": "authenticated", "sub": uuidDeEscenario(ana), "exp": time.Now().Add(time.Hour).Unix(),
		}), nil
	case `un token sin firma ("alg: none")`:
		return s.firmar(jwt.SigningMethodNone, nil, jwt.MapClaims{
			"iss": s.server.URL + "/auth/v1", "aud": "authenticated", "sub": uuidDeEscenario(ana), "exp": time.Now().Add(time.Hour).Unix(),
		}), nil
	case "un token con algoritmo HS256":
		return s.firmar(jwt.SigningMethodHS256, []byte("un-secreto-compartido"), jwt.MapClaims{
			"iss": s.server.URL + "/auth/v1", "aud": "authenticated", "sub": uuidDeEscenario(ana), "exp": time.Now().Add(time.Hour).Unix(),
		}), nil
	case "un token de otro emisor":
		return s.token(ana, func(c jwt.MapClaims) { c["iss"] = "https://otro.supabase.co/auth/v1" }), nil
	case "un token con otra audiencia":
		return s.token(ana, func(c jwt.MapClaims) { c["aud"] = "anon" }), nil
	case `un token cuyo "sub" no es un UUID`:
		return s.token(ana, func(c jwt.MapClaims) { c["sub"] = "ana" }), nil
	}
	return "", fmt.Errorf("el escenario no conoce la cookie %q", descripcion)
}

// tec05 guarda el estado de un escenario de TEC-05 (spec specs/TEC-05-autenticacion.md): una aplicación con repositorios en
// memoria y, cuando el escenario lo pide, el login real (verificador y cliente de Supabase) contra un Supabase de mentira.
type tec05 struct {
	supabase  *supabaseDeEscenario
	cancelar  context.CancelFunc
	handler   http.Handler
	proyectos *service.Proyectos
	repoProy  *memory.Proyectos
	integr    *service.Integrantes
	repoInteg *memory.Integrantes
	cuentas   *service.Cuentas
	idProy    map[string]int64

	cookie       string // la cookie de sesión que mandan los pedidos
	rec          *httptest.ResponseRecorder
	usuario      auth.Identidad
	quien        member.Integrante
	errQuien     error
	variables    map[string]string
	errArranque  error
	proyectoPost string
}

func registrarPasosTEC05(sc *godog.ScenarioContext) {
	repoProy, repoInteg := memory.NuevoProyectos(), memory.NuevoIntegrantes()
	e := &tec05{
		repoProy: repoProy, repoInteg: repoInteg,
		proyectos: service.NuevoProyectos(repoProy, time.Now),
		integr:    service.NuevoIntegrantes(repoProy, repoInteg),
		idProy:    map[string]int64{},
		variables: map[string]string{},
	}
	sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		if e.cancelar != nil {
			e.cancelar()
		}
		if e.supabase != nil {
			e.supabase.server.Close()
		}
		return ctx, nil
	})

	// el servicio de login y la aplicación
	sc.Step(`^que el servicio de login está disponible$`, func() error { return e.levantar(false, false) })
	sc.Step(`^que el servicio de login pide confirmar el email$`, func() error { return e.levantar(true, false) })
	sc.Step(`^que el servicio de login no responde$`, func() error { return e.levantar(false, true) })
	sc.Step(`^que el login está activado$`, func() error { return e.levantar(false, false) })
	sc.Step(`^que el login no está configurado$`, e.sinLogin)
	sc.Step(`^que ya existe la cuenta "([^"]*)" con la contraseña "([^"]*)"$`, e.queExisteLaCuenta)

	// registro
	sc.Step(`^me registro con el email "([^"]*)" y la contraseña "([^"]*)"$`, func(email, clave string) error { return e.registrarse(email, clave, clave) })
	sc.Step(`^me registro con el email "([^"]*)", la contraseña "([^"]*)" y la repetición "([^"]*)"$`, e.registrarse)
	sc.Step(`^me registro con el email "([^"]*)" y una contraseña de (\d+) caracteres$`, func(email string, largo int) error {
		clave := strings.Repeat("a", largo)
		return e.registrarse(email, clave, clave)
	})
	sc.Step(`^la cuenta "([^"]*)" existe en el servicio de login$`, e.laCuentaExiste)
	sc.Step(`^no se crea ninguna cuenta$`, e.noSeCreaNingunaCuenta)
	sc.Step(`^la respuesta del registro es el mensaje "([^"]*)"$`, e.elRegistroMuestra)
	sc.Step(`^el registro falla en el campo "([^"]*)" con "([^"]*)"$`, e.elRegistroFalla)

	// ingreso y salida
	sc.Step(`^inicio sesión con el email "([^"]*)" y la contraseña "([^"]*)"$`, e.iniciarSesion)
	sc.Step(`^inicio sesión con la contraseña "([^"]*)" y siguiente "([^"]*)"$`, e.iniciarSesionConSiguiente)
	sc.Step(`^el ingreso es correcto y se guarda la cookie de sesión$`, e.elIngresoEsCorrecto)
	sc.Step(`^la cookie de sesión es "([^"]*)" y "([^"]*)"$`, e.laCookieEs)
	sc.Step(`^el ingreso falla con el código (\d+) y el mensaje "([^"]*)"$`, e.elIngresoFalla)
	sc.Step(`^no se guarda ninguna cookie de sesión$`, e.noSeGuardaCookie)
	sc.Step(`^el ingreso redirige a "([^"]*)"$`, e.elIngresoRedirige)
	sc.Step(`^cierro sesión$`, e.cerrarSesion)
	sc.Step(`^la cookie de sesión se borra$`, e.laCookieSeBorra)
	sc.Step(`^la respuesta lleva a "([^"]*)" con el mensaje "([^"]*)"$`, e.lleva)

	// el middleware
	sc.Step(`^que tengo una sesión válida de "([^"]*)"$`, e.tengoUnaSesionValida)
	sc.Step(`^que mi cookie de sesión es (.+)$`, e.miCookieEs)
	sc.Step(`^pido la página "([^"]*)"$`, func(ruta string) error { return e.pedirPagina(ruta, e.cookie) })
	sc.Step(`^pido la página "([^"]*)" sin haber iniciado sesión$`, func(ruta string) error { return e.pedirPagina(ruta, "") })
	sc.Step(`^creo el proyecto "([^"]*)" sin haber iniciado sesión$`, e.crearProyectoSinSesion)
	sc.Step(`^la respuesta tiene el código (\d+)$`, e.laRespuestaTieneElCodigo)
	sc.Step(`^la respuesta muestra el mensaje "([^"]*)"$`, e.laRespuestaMuestra)
	sc.Step(`^la respuesta no muestra la lista de proyectos$`, e.noMuestraLaLista)
	sc.Step(`^no se creó ningún proyecto$`, e.noSeCreoNingunProyecto)

	// la configuración
	sc.Step(`^que solo está definida la variable (\w+)$`, func(nombre string) error {
		e.variables = map[string]string{nombre: "valor"}
		return nil
	})
	sc.Step(`^arranca la aplicación$`, e.arranca)
	sc.Step(`^la aplicación no arranca y dice "([^"]*)"$`, e.noArrancaYDice)

	// el vínculo con los integrantes
	sc.Step(`^que hay un proyecto "([^"]*)" sin integrantes$`, func(nombre string) error { _, err := e.proyecto(nombre); return err })
	sc.Step(`^que "([^"]*)" con el email "([^"]*)" es integrante del proyecto "([^"]*)"$`, func(nombre, email, proyecto string) error {
		return e.integrante(nombre, email, proyecto)
	})
	sc.Step(`^que "([^"]*)" con el email "([^"]*)" es integrante de los proyectos "([^"]*)" y "([^"]*)"$`, func(nombre, email, uno, otro string) error {
		if err := e.integrante(nombre, email, uno); err != nil {
			return err
		}
		return e.integrante(nombre, email, otro)
	})
	sc.Step(`^que "([^"]*)" con el email "([^"]*)" es integrante del proyecto "([^"]*)" y ya está vinculada con otra cuenta$`, func(nombre, email, proyecto string) error {
		if err := e.integrante(nombre, email, proyecto); err != nil {
			return err
		}
		_, err := e.repoInteg.VincularPorEmail(context.Background(), email, uuidDeEscenario("otra@cuenta.com"))
		return err
	})
	sc.Step(`^que después "([^"]*)" con el email "([^"]*)" se registra como integrante del proyecto "([^"]*)"$`, func(nombre, email, proyecto string) error {
		return e.integrante(nombre, email, proyecto)
	})
	sc.Step(`^que inicié sesión como "([^"]*)"$`, func(email string) error {
		e.usuario = auth.Identidad{ID: uuidDeEscenario(email), Email: email}
		return nil
	})
	sc.Step(`^consulto quién soy en el proyecto "([^"]*)"$`, e.consultarQuienSoy)
	sc.Step(`^soy el integrante "([^"]*)"$`, e.soyElIntegrante)
	sc.Step(`^el resultado es que no soy integrante del proyecto$`, e.noSoyIntegrante)
	sc.Step(`^"([^"]*)" queda vinculada con la cuenta "([^"]*)" en los proyectos "([^"]*)" y "([^"]*)"$`, func(nombre, cuenta, uno, otro string) error {
		for _, p := range []string{uno, otro} {
			if err := e.estaVinculada(nombre, cuenta, p); err != nil {
				return err
			}
		}
		return nil
	})
	sc.Step(`^"([^"]*)" queda vinculada con la cuenta "([^"]*)" en el proyecto "([^"]*)"$`, e.estaVinculada)
	sc.Step(`^"([^"]*)" no queda vinculado con ninguna cuenta$`, e.noEstaVinculado)
	sc.Step(`^"([^"]*)" sigue vinculada con la otra cuenta$`, func(nombre string) error {
		return e.estaVinculada(nombre, "otra@cuenta.com", "Alfa")
	})
}

// sinServicioDeLogin es el servicio de cuentas de los escenarios que no levantan el login: no hay con quién hablar.
type sinServicioDeLogin struct{}

func (sinServicioDeLogin) Registrar(context.Context, account.Credenciales) (auth.ResultadoRegistro, error) {
	return auth.ResultadoRegistro{}, auth.ErrServicioNoDisponible
}

func (sinServicioDeLogin) Ingresar(context.Context, account.Credenciales) (auth.Sesion, error) {
	return auth.Sesion{}, auth.ErrServicioNoDisponible
}

// levantar arma el login real de la aplicación contra un Supabase de mentira (RN1: las dos variables están definidas).
func (e *tec05) levantar(pideConfirmar, noResponde bool) error {
	if e.supabase != nil {
		return nil
	}
	e.supabase = nuevoSupabaseDeEscenario()
	e.supabase.pideConfirmar, e.supabase.noResponde = pideConfirmar, noResponde
	cfg := auth.Config{URL: e.supabase.server.URL, AnonKey: "clave-anon"}
	ctx, cancelar := context.WithCancel(context.Background())
	e.cancelar = cancelar
	verificador, err := auth.NuevoVerificador(ctx, cfg)
	if err != nil {
		return fmt.Errorf("no se pudo armar el verificador: %w", err)
	}
	e.cuentas = service.NuevoCuentas(auth.NuevoSupabase(cfg), e.repoInteg)
	e.armarRouter(verificador)
	return nil
}

// sinLogin arma la aplicación como en los previews y los tests: sin variables de Supabase no se pide sesión.
func (e *tec05) sinLogin() error {
	e.cuentas = service.NuevoCuentas(sinServicioDeLogin{}, e.repoInteg)
	e.armarRouter(nil)
	return nil
}

func (e *tec05) armarRouter(verificador auth.Verificador) {
	sprints, historias := memory.NuevoSprints(), memory.NuevoHistorias()
	dep := server.Dependencias{
		Proyectos:   e.proyectos,
		Sprints:     service.NuevoSprints(e.repoProy, sprints),
		Integrantes: e.integr,
		Estado:      service.NuevoEstadoProyectos(e.repoProy, sprints, historias, time.Now),
	}
	if verificador != nil {
		dep.Verificador, dep.Cuentas = verificador, e.cuentas
	}
	e.handler = server.NewRouter(dep)
}

// asegurarCuentas prepara el caso de uso Cuentas para los escenarios que solo consultan el vínculo, sin pasar por el login.
func (e *tec05) asegurarCuentas() {
	if e.cuentas == nil {
		e.cuentas = service.NuevoCuentas(sinServicioDeLogin{}, e.repoInteg)
	}
}

func (e *tec05) pedir(metodo, ruta string, campos neturl.Values, cookie string) *httptest.ResponseRecorder {
	if e.handler == nil {
		panic("el escenario no armó la aplicación: falta un paso de \"el servicio de login está disponible\" o similar")
	}
	req := httptest.NewRequest(metodo, ruta, strings.NewReader(campos.Encode()))
	if metodo == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if cookie != "" {
		req.AddCookie(&http.Cookie{Name: "sesion", Value: cookie})
	}
	rec := httptest.NewRecorder()
	e.handler.ServeHTTP(rec, req)
	return rec
}

func (e *tec05) queExisteLaCuenta(email, clave string) error {
	if e.supabase == nil {
		return errors.New("falta el paso que levanta el servicio de login")
	}
	e.supabase.mu.Lock()
	defer e.supabase.mu.Unlock()
	e.supabase.cuentas[email] = clave
	return nil
}

// registro

func (e *tec05) registrarse(email, clave, repeticion string) error {
	e.rec = e.pedir(http.MethodPost, "/registro", neturl.Values{"email": {email}, "clave": {clave}, "repeticion": {repeticion}}, "")
	return nil
}

func (e *tec05) laCuentaExiste(email string) error {
	e.supabase.mu.Lock()
	defer e.supabase.mu.Unlock()
	if _, existe := e.supabase.cuentas[email]; !existe {
		return fmt.Errorf("la cuenta %q no se creó en el servicio de login", email)
	}
	return nil
}

func (e *tec05) noSeCreaNingunaCuenta() error {
	e.supabase.mu.Lock()
	defer e.supabase.mu.Unlock()
	if n := len(e.supabase.cuentas); n != 0 {
		return fmt.Errorf("no se tendría que haber creado ninguna cuenta y hay %d", n)
	}
	return nil
}

// elRegistroMuestra sigue la redirección del registro hasta el login, que es donde se ve el mensaje.
func (e *tec05) elRegistroMuestra(mensaje string) error {
	if e.rec.Code != http.StatusSeeOther {
		return fmt.Errorf("se esperaba que el registro redirija (303) y dio %d:\n%s", e.rec.Code, e.rec.Body.String())
	}
	pagina := e.pedir(http.MethodGet, e.rec.Header().Get("Location"), nil, "")
	if !strings.Contains(pagina.Body.String(), mensaje) {
		return fmt.Errorf("la página %s no muestra %q", e.rec.Header().Get("Location"), mensaje)
	}
	return nil
}

// elRegistroFalla comprueba el 422, el mensaje y que el error esté marcado en el campo.
func (e *tec05) elRegistroFalla(campo, mensaje string) error {
	if e.rec.Code != http.StatusUnprocessableEntity {
		return fmt.Errorf("se esperaba el código 422 y se obtuvo %d", e.rec.Code)
	}
	cuerpo := e.rec.Body.String()
	if !strings.Contains(cuerpo, mensaje) {
		return fmt.Errorf("la página no muestra %q", mensaje)
	}
	marcado := regexp.MustCompile(`class="form-control[^"]*is-invalid[^"]*"[^>]*id="` + regexp.QuoteMeta(campo) + `"`)
	if !marcado.MatchString(cuerpo) {
		return fmt.Errorf("el campo %q no está marcado con el error", campo)
	}
	return nil
}

// ingreso y salida

func (e *tec05) iniciarSesion(email, clave string) error {
	e.rec = e.pedir(http.MethodPost, "/login", neturl.Values{"email": {email}, "clave": {clave}}, "")
	return nil
}

func (e *tec05) iniciarSesionConSiguiente(clave, siguiente string) error {
	e.rec = e.pedir(http.MethodPost, "/login", neturl.Values{"email": {"ana@mail.com"}, "clave": {clave}, "siguiente": {siguiente}}, "")
	return nil
}

func (e *tec05) cookieDeLaRespuesta() *http.Cookie {
	for _, c := range e.rec.Result().Cookies() {
		if c.Name == "sesion" {
			return c
		}
	}
	return nil
}

func (e *tec05) elIngresoEsCorrecto() error {
	if e.rec.Code != http.StatusSeeOther {
		return fmt.Errorf("se esperaba el código 303 y se obtuvo %d:\n%s", e.rec.Code, e.rec.Body.String())
	}
	if c := e.cookieDeLaRespuesta(); c == nil || c.Value == "" {
		return errors.New("la respuesta no guardó la cookie sesion")
	}
	return nil
}

func (e *tec05) laCookieEs(primero, segundo string) error {
	c := e.cookieDeLaRespuesta()
	if c == nil {
		return errors.New("la respuesta no guardó la cookie sesion")
	}
	for _, atributo := range []string{primero, segundo} {
		var cumple bool
		switch atributo {
		case "HttpOnly":
			cumple = c.HttpOnly
		case "SameSite=Lax":
			cumple = c.SameSite == http.SameSiteLaxMode
		default:
			return fmt.Errorf("el escenario no conoce el atributo %q", atributo)
		}
		if !cumple {
			return fmt.Errorf("la cookie no es %s: %+v", atributo, c)
		}
	}
	return nil
}

func (e *tec05) elIngresoFalla(codigo int, mensaje string) error {
	if e.rec.Code != codigo {
		return fmt.Errorf("se esperaba el código %d y se obtuvo %d", codigo, e.rec.Code)
	}
	if !strings.Contains(e.rec.Body.String(), mensaje) {
		return fmt.Errorf("la página no muestra %q", mensaje)
	}
	return nil
}

func (e *tec05) noSeGuardaCookie() error {
	if c := e.cookieDeLaRespuesta(); c != nil {
		return fmt.Errorf("no se tendría que haber guardado la cookie y se guardó %+v", c)
	}
	return nil
}

func (e *tec05) elIngresoRedirige(destino string) error {
	if e.rec.Code != http.StatusSeeOther || e.rec.Header().Get("Location") != destino {
		return fmt.Errorf("se esperaba 303 a %q y se obtuvo %d a %q", destino, e.rec.Code, e.rec.Header().Get("Location"))
	}
	return nil
}

func (e *tec05) cerrarSesion() error {
	e.rec = e.pedir(http.MethodPost, "/salir", neturl.Values{}, e.cookie)
	return nil
}

func (e *tec05) laCookieSeBorra() error {
	if c := e.cookieDeLaRespuesta(); c == nil || c.Value != "" || c.MaxAge >= 0 {
		return fmt.Errorf("se esperaba que la respuesta borre la cookie y se obtuvo %+v", c)
	}
	return nil
}

func (e *tec05) lleva(ruta, mensaje string) error {
	destino := e.rec.Header().Get("Location")
	if e.rec.Code != http.StatusSeeOther || !strings.HasPrefix(destino, ruta) {
		return fmt.Errorf("se esperaba 303 a %s y se obtuvo %d a %q", ruta, e.rec.Code, destino)
	}
	if !strings.Contains(e.pedir(http.MethodGet, destino, nil, "").Body.String(), mensaje) {
		return fmt.Errorf("la página %s no muestra %q", destino, mensaje)
	}
	return nil
}

// el middleware

func (e *tec05) tengoUnaSesionValida(email string) error {
	if e.supabase == nil {
		return errors.New("falta el paso que activa el login")
	}
	e.cookie = e.supabase.token(email, nil)
	return nil
}

func (e *tec05) miCookieEs(descripcion string) error {
	if e.supabase == nil {
		return errors.New("falta el paso que activa el login")
	}
	token, err := e.supabase.tokenDe(descripcion)
	e.cookie = token
	return err
}

func (e *tec05) pedirPagina(ruta, cookie string) error {
	e.rec = e.pedir(http.MethodGet, ruta, nil, cookie)
	return nil
}

func (e *tec05) crearProyectoSinSesion(nombre string) error {
	e.proyectoPost = nombre
	e.rec = e.pedir(http.MethodPost, "/proyectos", neturl.Values{
		"nombre": {nombre}, "descripcion": {"Sin sesión"}, "fecha_inicio": {"2026-10-05"}, "fecha_fin": {"2026-11-01"},
	}, "")
	return nil
}

func (e *tec05) laRespuestaTieneElCodigo(codigo int) error {
	if e.rec.Code != codigo {
		return fmt.Errorf("se esperaba el código %d y se obtuvo %d", codigo, e.rec.Code)
	}
	return nil
}

func (e *tec05) laRespuestaMuestra(mensaje string) error {
	if !strings.Contains(e.rec.Body.String(), mensaje) {
		return fmt.Errorf("la respuesta no muestra %q:\n%s", mensaje, e.rec.Body.String())
	}
	return nil
}

// noMuestraLaLista usa el enlace a "Nuevo proyecto", que está en la lista de proyectos y no en la página de login.
func (e *tec05) noMuestraLaLista() error {
	if strings.Contains(e.rec.Body.String(), `href="/proyectos/nuevo"`) {
		return errors.New("la respuesta muestra la lista de proyectos")
	}
	return nil
}

func (e *tec05) noSeCreoNingunProyecto() error {
	lista, err := e.proyectos.Listar(context.Background())
	if err != nil {
		return err
	}
	for _, p := range lista {
		if p.Nombre == e.proyectoPost {
			return fmt.Errorf("se creó el proyecto %q", p.Nombre)
		}
	}
	return nil
}

// la configuración

func (e *tec05) arranca() error {
	_, e.errArranque = app.NewHandlerDesdeEntorno(func(nombre string) string { return e.variables[nombre] })
	return nil
}

func (e *tec05) noArrancaYDice(mensaje string) error {
	if e.errArranque == nil || !strings.Contains(e.errArranque.Error(), mensaje) {
		return fmt.Errorf("se esperaba que no arranque con %q y se obtuvo: %v", mensaje, e.errArranque)
	}
	return nil
}

// el vínculo con los integrantes

// proyecto devuelve el ID del proyecto con ese nombre, y lo crea la primera vez.
func (e *tec05) proyecto(nombre string) (int64, error) {
	if id, ok := e.idProy[nombre]; ok {
		return id, nil
	}
	p, err := e.proyectos.Crear(context.Background(), project.Datos{
		Nombre: nombre, FechaInicio: fechaDeEscenario("2026-10-05"), FechaFin: fechaDeEscenario("2026-11-01"),
	})
	if err != nil {
		return 0, fmt.Errorf("no se pudo crear el proyecto %q: %w", nombre, err)
	}
	e.idProy[nombre] = p.ID
	return p.ID, nil
}

func (e *tec05) integrante(nombre, email, proyecto string) error {
	id, err := e.proyecto(proyecto)
	if err != nil {
		return err
	}
	if _, err := e.integr.Registrar(context.Background(), id, member.Datos{Nombre: nombre, Email: email, Rol: member.RolProductBuilder}); err != nil {
		return fmt.Errorf("no se pudo registrar a %q en %q: %w", nombre, proyecto, err)
	}
	return nil
}

func (e *tec05) consultarQuienSoy(proyecto string) error {
	e.asegurarCuentas()
	id, err := e.proyecto(proyecto)
	if err != nil {
		return err
	}
	e.quien, e.errQuien = e.cuentas.Quien(context.Background(), e.usuario, id)
	return nil
}

func (e *tec05) soyElIntegrante(nombre string) error {
	if e.errQuien != nil {
		return fmt.Errorf("se esperaba ser %q y dio el error: %w", nombre, e.errQuien)
	}
	if e.quien.Nombre != nombre {
		return fmt.Errorf("se esperaba ser %q y se obtuvo %q", nombre, e.quien.Nombre)
	}
	return nil
}

func (e *tec05) noSoyIntegrante() error {
	if !errors.Is(e.errQuien, member.ErrNoEsIntegrante) {
		return fmt.Errorf("se esperaba member.ErrNoEsIntegrante y se obtuvo %v (integrante %+v)", e.errQuien, e.quien)
	}
	return nil
}

func (e *tec05) integranteDe(nombre, proyecto string) (member.Integrante, error) {
	id, err := e.proyecto(proyecto)
	if err != nil {
		return member.Integrante{}, err
	}
	lista, err := e.repoInteg.ListarPorProyecto(context.Background(), id)
	if err != nil {
		return member.Integrante{}, err
	}
	for _, i := range lista {
		if i.Nombre == nombre {
			return i, nil
		}
	}
	return member.Integrante{}, fmt.Errorf("%q no es integrante del proyecto %q", nombre, proyecto)
}

func (e *tec05) estaVinculada(nombre, cuenta, proyecto string) error {
	i, err := e.integranteDe(nombre, proyecto)
	if err != nil {
		return err
	}
	if esperado := uuidDeEscenario(cuenta); i.AuthUserID != esperado {
		return fmt.Errorf("%q en %q tendría que estar vinculada con %s y está con %q", nombre, proyecto, esperado, i.AuthUserID)
	}
	return nil
}

func (e *tec05) noEstaVinculado(nombre string) error {
	for proyecto := range e.idProy {
		i, err := e.integranteDe(nombre, proyecto)
		if err != nil {
			continue // no es integrante de ese proyecto
		}
		if i.AuthUserID != "" {
			return fmt.Errorf("%q en %q no tendría que estar vinculado y lo está con %q", nombre, proyecto, i.AuthUserID)
		}
	}
	return nil
}
