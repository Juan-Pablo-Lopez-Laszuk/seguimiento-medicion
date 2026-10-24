package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/time/rate"
)

// Errores de la verificación de un token (RN2 y RN3). El texto de ErrTokenInvalido puede traer el motivo técnico: sirve
// para el log, nunca para mostrárselo al usuario.
var (
	ErrTokenInvalido = errors.New("el token no es válido")
	ErrTokenVencido  = errors.New("el token venció")
)

// Verificador comprueba el token de acceso de una sesión y dice a qué usuario pertenece.
type Verificador interface {
	// Verificar devuelve el usuario del token, ErrTokenVencido si venció o ErrTokenInvalido si no vale por cualquier otro
	// motivo (formato, firma, algoritmo, emisor, audiencia o sub).
	Verificar(ctx context.Context, token string) (Identidad, error)
}

const (
	audiencia           = "authenticated" // el "aud" que Supabase pone a los tokens de usuarios con sesión
	plazoDelJWKS        = 5 * time.Second // lo máximo que se espera a Supabase al bajar las claves (RN12)
	esperaMaximaRefresh = time.Second     // lo máximo que un pedido espera para provocar otra bajada de claves
	refrescoPeriodico   = time.Hour       // cada cuánto se vuelven a bajar las claves, haya o no cambios
	esperaPorDefecto    = time.Minute     // lo mínimo entre dos bajadas provocadas por un kid desconocido
)

// uuid es el formato del "sub" de un usuario de Supabase Auth (RN2).
var uuid = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// OpcionVerificador cambia el comportamiento por defecto de NuevoVerificador.
type OpcionVerificador func(*opcionesVerificador)

type opcionesVerificador struct {
	esperaEntreRefrescos time.Duration
}

// ConEsperaEntreRefrescos fija cuánto tiene que pasar, como mínimo, entre dos bajadas de las claves provocadas por un
// token con un kid desconocido. Por defecto es un minuto: así un token inventado no hace que se le pida el JWKS a
// Supabase a cada rato, y si estaba caído se recupera solo en ese tiempo. Los tests la acortan.
func ConEsperaEntreRefrescos(d time.Duration) OpcionVerificador {
	return func(o *opcionesVerificador) { o.esperaEntreRefrescos = d }
}

type verificadorJWKS struct {
	claves keyfunc.Keyfunc
	parser *jwt.Parser
}

// claimsSupabase son los datos del token que usa la aplicación.
type claimsSupabase struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
}

// NuevoVerificador arma el verificador de los tokens de Supabase (RN2). Las claves públicas se bajan del JWKS del
// proyecto y se reutilizan; se vuelven a bajar cada hora y cuando llega un token con un kid que no se conoce. Si Supabase
// no responde al arrancar, no es un error: los tokens se rechazan hasta que las claves se puedan bajar (RN12). El contexto
// termina la actualización de las claves.
func NuevoVerificador(ctx context.Context, cfg Config, opciones ...OpcionVerificador) (Verificador, error) {
	o := opcionesVerificador{esperaEntreRefrescos: esperaPorDefecto}
	for _, opcion := range opciones {
		opcion(&o)
	}

	direccion := cfg.URLJWKS()
	almacen, err := jwkset.NewStorageFromHTTP(direccion, jwkset.HTTPClientStorageOptions{
		Ctx:                       ctx,
		HTTPTimeout:               plazoDelJWKS,
		NoErrorReturnFirstHTTPReq: true,
		RefreshInterval:           refrescoPeriodico,
		RefreshErrorHandler: func(_ context.Context, err error) {
			log.Printf("auth: no se pudieron bajar las claves de %s: %v", direccion, err)
		},
	})
	if err != nil {
		return nil, fmt.Errorf("preparar la descarga de las claves de Supabase: %w", err)
	}
	cliente, err := jwkset.NewHTTPClient(jwkset.HTTPClientOptions{
		HTTPURLs:          map[string]jwkset.Storage{direccion: almacen},
		RateLimitWaitMax:  esperaMaximaRefresh,
		RefreshUnknownKID: rate.NewLimiter(rate.Every(o.esperaEntreRefrescos), 1),
	})
	if err != nil {
		return nil, fmt.Errorf("preparar el cliente de las claves de Supabase: %w", err)
	}
	claves, err := keyfunc.New(keyfunc.Options{Ctx: ctx, Storage: cliente})
	if err != nil {
		return nil, fmt.Errorf("preparar la búsqueda de claves de Supabase: %w", err)
	}

	return verificadorJWKS{
		claves: claves,
		parser: jwt.NewParser(
			jwt.WithValidMethods([]string{"ES256", "RS256"}),
			jwt.WithIssuer(cfg.Emisor()),
			jwt.WithAudience(audiencia),
			jwt.WithExpirationRequired(),
		),
	}, nil
}

func (v verificadorJWKS) Verificar(ctx context.Context, token string) (Identidad, error) {
	var claims claimsSupabase
	if _, err := v.parser.ParseWithClaims(token, &claims, v.claves.KeyfuncCtx(ctx)); err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return Identidad{}, ErrTokenVencido
		}
		return Identidad{}, fmt.Errorf("%w: %v", ErrTokenInvalido, err)
	}
	if !uuid.MatchString(claims.Subject) {
		return Identidad{}, fmt.Errorf("%w: el sub no es un UUID", ErrTokenInvalido)
	}
	return Identidad{ID: strings.ToLower(claims.Subject), Email: strings.ToLower(strings.TrimSpace(claims.Email))}, nil
}
