package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/account"
)

// Errores del servicio de cuentas, ya traducidos de los de Supabase (sección 7 de la spec). Ninguno lleva lo que Supabase
// contestó: pueden traer el email, la contraseña o el token (RN11).
var (
	ErrEmailYaRegistrado       = errors.New("ya hay una cuenta con ese email")
	ErrClaveDebil              = errors.New("la contraseña es demasiado débil")
	ErrEmailRechazado          = errors.New("el servicio de login no aceptó el email")
	ErrCredencialesIncorrectas = errors.New("el email o la contraseña no son correctos")
	ErrEmailSinConfirmar       = errors.New("el email todavía no se confirmó")
	ErrDemasiadosIntentos      = errors.New("demasiados intentos")
	// ErrServicioNoDisponible es cuando Supabase no responde a tiempo, falla o contesta algo inesperado (RN12).
	ErrServicioNoDisponible = errors.New("el servicio de inicio de sesión no está disponible")
)

// ResultadoRegistro dice si la cuenta quedó lista o Supabase pidió confirmar el email antes (RN5).
type ResultadoRegistro struct {
	RequiereConfirmacion bool
}

// Sesion es lo que devuelve un ingreso correcto: el token de acceso, cuándo vence y a quién pertenece.
type Sesion struct {
	Token   string
	Vence   time.Time
	Usuario Identidad
}

// ServicioDeCuentas crea cuentas y abre sesiones. Lo implementa Supabase; los tests usan uno falso.
type ServicioDeCuentas interface {
	Registrar(ctx context.Context, c account.Credenciales) (ResultadoRegistro, error)
	Ingresar(ctx context.Context, c account.Credenciales) (Sesion, error)
}

const (
	plazoDeSupabase = 5 * time.Second // RN12: más que esto el servicio no responde
	maxRespuesta    = 1 << 20         // no se lee más de 1 MiB de lo que contesta Supabase
)

// Supabase es el cliente de la API de Supabase Auth (GoTrue).
type Supabase struct {
	cfg   Config
	http  *http.Client
	ahora func() time.Time
}

// OpcionSupabase cambia el comportamiento por defecto de NuevoSupabase.
type OpcionSupabase func(*Supabase)

// ConClienteHTTP reemplaza el cliente HTTP, que por defecto espera 5 segundos como máximo.
func ConClienteHTTP(c *http.Client) OpcionSupabase { return func(s *Supabase) { s.http = c } }

// ConReloj reemplaza el reloj con el que se calcula cuándo vence la sesión.
func ConReloj(ahora func() time.Time) OpcionSupabase { return func(s *Supabase) { s.ahora = ahora } }

// NuevoSupabase arma el cliente de las cuentas del proyecto de Supabase indicado.
func NuevoSupabase(cfg Config, opciones ...OpcionSupabase) *Supabase {
	s := &Supabase{cfg: cfg, http: &http.Client{Timeout: plazoDeSupabase}, ahora: time.Now}
	for _, opcion := range opciones {
		opcion(s)
	}
	return s
}

// Registrar crea la cuenta (RN5). Con "Confirm email" activado en Supabase, la cuenta no puede ingresar hasta confirmar.
func (s *Supabase) Registrar(ctx context.Context, c account.Credenciales) (ResultadoRegistro, error) {
	cuerpo, err := s.pedir(ctx, "/auth/v1/signup", c)
	if err != nil {
		return ResultadoRegistro{}, err
	}
	var r struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(cuerpo, &r); err != nil {
		return ResultadoRegistro{}, inesperada("el registro")
	}
	// Con la confirmación de email activada Supabase devuelve el usuario sin sesión.
	return ResultadoRegistro{RequiereConfirmacion: r.AccessToken == ""}, nil
}

// Ingresar abre una sesión con el email y la contraseña (RN6).
func (s *Supabase) Ingresar(ctx context.Context, c account.Credenciales) (Sesion, error) {
	cuerpo, err := s.pedir(ctx, "/auth/v1/token?grant_type=password", c)
	if err != nil {
		return Sesion{}, err
	}
	var r struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		User        struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"user"`
	}
	if err := json.Unmarshal(cuerpo, &r); err != nil || r.AccessToken == "" || r.ExpiresIn <= 0 || !uuid.MatchString(r.User.ID) {
		return Sesion{}, inesperada("el ingreso")
	}
	return Sesion{
		Token:   r.AccessToken,
		Vence:   s.ahora().Add(time.Duration(r.ExpiresIn) * time.Second),
		Usuario: Identidad{ID: strings.ToLower(r.User.ID), Email: strings.ToLower(strings.TrimSpace(r.User.Email))},
	}, nil
}

// pedir manda las credenciales a Supabase y devuelve el cuerpo de una respuesta exitosa; si no lo fue, el error ya traducido.
func (s *Supabase) pedir(ctx context.Context, ruta string, c account.Credenciales) ([]byte, error) {
	enviado, err := json.Marshal(map[string]string{"email": c.Email, "password": c.Clave})
	if err != nil {
		return nil, fmt.Errorf("armar el pedido a Supabase: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.cfg.URL+ruta, bytes.NewReader(enviado))
	if err != nil {
		return nil, fmt.Errorf("armar el pedido a Supabase: %w", err)
	}
	req.Header.Set("apikey", s.cfg.AnonKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.http.Do(req)
	if err != nil {
		// El error de net/http nombra la URL (sin la clave ni los datos); no trae el cuerpo del pedido.
		log.Printf("auth: no se pudo consultar a Supabase: %v", err)
		return nil, fmt.Errorf("%w: sin respuesta", ErrServicioNoDisponible)
	}
	defer func() { _ = resp.Body.Close() }()

	cuerpo, err := io.ReadAll(io.LimitReader(resp.Body, maxRespuesta+1))
	if err != nil || len(cuerpo) > maxRespuesta {
		return nil, inesperada("la respuesta")
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, traducir(resp.StatusCode, cuerpo)
	}
	return cuerpo, nil
}

func inesperada(que string) error {
	log.Printf("auth: Supabase contestó algo inesperado en %s", que)
	return fmt.Errorf("%w: respuesta inesperada", ErrServicioNoDisponible)
}

// errorSupabase son los campos que usa la aplicación de un error de Supabase; las versiones nuevas informan error_code y
// las viejas solo el texto.
type errorSupabase struct {
	Codigo      string `json:"error_code"`
	Msg         string `json:"msg"`
	Message     string `json:"message"`
	Error       string `json:"error"`
	Descripcion string `json:"error_description"`
}

// traducir convierte el error de Supabase en uno de la aplicación. Lo que no reconoce es ErrServicioNoDisponible, y al
// log va solo el estado y el código del error, nunca el cuerpo.
func traducir(estado int, cuerpo []byte) error {
	var e errorSupabase
	_ = json.Unmarshal(cuerpo, &e) // si no es JSON queda vacío y se trata como un error desconocido
	texto := strings.ToLower(e.Msg + " " + e.Message + " " + e.Error + " " + e.Descripcion)

	switch {
	case estado == http.StatusTooManyRequests || (strings.HasPrefix(e.Codigo, "over_") && strings.HasSuffix(e.Codigo, "_rate_limit")):
		return ErrDemasiadosIntentos
	case estado >= 500:
		// sigue abajo: un error del servidor nunca es un error del usuario
	case e.Codigo == "user_already_exists" || e.Codigo == "email_exists" || strings.Contains(texto, "already registered"):
		return ErrEmailYaRegistrado
	case e.Codigo == "weak_password":
		return ErrClaveDebil
	case e.Codigo == "email_address_invalid":
		return ErrEmailRechazado
	case e.Codigo == "invalid_credentials" || strings.Contains(texto, "invalid login credentials"):
		return ErrCredencialesIncorrectas
	case e.Codigo == "email_not_confirmed" || strings.Contains(texto, "email not confirmed"):
		return ErrEmailSinConfirmar
	}
	log.Printf("auth: Supabase respondió %d (error_code %.64q)", estado, e.Codigo)
	return fmt.Errorf("%w: respondió %d", ErrServicioNoDisponible, estado)
}
