// Package auth conecta la aplicación con Supabase Auth (TEC-05): lee su configuración, verifica los tokens de acceso
// sin llamarlo en cada pedido y registra e ingresa cuentas. Spec: specs/TEC-05-autenticacion.md.
package auth

import (
	"context"
	"errors"
	"net/url"
	"strings"
)

// Nombres de las variables de entorno, los mismos de .env.example.
const (
	variableURL = "SUPABASE_URL"
	variableKey = "SUPABASE_ANON_KEY"
)

// Config es la configuración de Supabase Auth: la URL del proyecto (sin barra final) y su clave pública anon.
type Config struct {
	URL     string
	AnonKey string
}

// Emisor es el valor que tiene el campo "iss" de los tokens del proyecto (RN2).
func (c Config) Emisor() string { return c.URL + "/auth/v1" }

// URLJWKS es donde Supabase publica las claves públicas con las que firma los tokens (RN2).
func (c Config) URLJWKS() string { return c.Emisor() + "/.well-known/jwks.json" }

// ConfigDesdeEntorno lee SUPABASE_URL y SUPABASE_ANON_KEY (RN1). Sin ninguna de las dos devuelve activo en false: la
// aplicación funciona sin login. Si falta una sola, o la URL no es http(s), devuelve un error que corta el arranque:
// un login a medio configurar no puede quedar abierto sin avisar.
func ConfigDesdeEntorno(getenv func(string) string) (cfg Config, activo bool, err error) {
	direccion := strings.TrimSpace(getenv(variableURL))
	clave := strings.TrimSpace(getenv(variableKey))
	switch {
	case direccion == "" && clave == "":
		return Config{}, false, nil
	case direccion == "":
		return Config{}, false, errors.New("falta " + variableURL)
	case clave == "":
		return Config{}, false, errors.New("falta " + variableKey)
	}
	if u, err := url.Parse(direccion); err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return Config{}, false, errors.New(variableURL + " no es una URL http(s) válida")
	}
	return Config{URL: strings.TrimRight(direccion, "/"), AnonKey: clave}, true, nil
}

// Identidad es el usuario con sesión: su ID en Supabase Auth (un UUID) y su email.
type Identidad struct {
	ID    string
	Email string
}

type claveIdentidad struct{}

// ConIdentidad devuelve un contexto con el usuario de la sesión.
func ConIdentidad(ctx context.Context, i Identidad) context.Context {
	return context.WithValue(ctx, claveIdentidad{}, i)
}

// IdentidadDe devuelve el usuario con sesión del contexto, si hay uno.
func IdentidadDe(ctx context.Context) (Identidad, bool) {
	i, ok := ctx.Value(claveIdentidad{}).(Identidad)
	return i, ok
}
