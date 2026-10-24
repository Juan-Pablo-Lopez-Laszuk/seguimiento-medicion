package auth_test

import (
	"context"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/auth"
)

// entorno arma un getenv con las variables dadas, para no tocar las del proceso.
func entorno(variables map[string]string) func(string) string {
	return func(nombre string) string { return variables[nombre] }
}

// RN1 · sin SUPABASE_URL ni SUPABASE_ANON_KEY no hay login
func TestConfigDesdeEntorno_SinVariables_NoActivaElLogin(t *testing.T) {
	cfg, activo, err := auth.ConfigDesdeEntorno(entorno(nil))

	if err != nil || activo || cfg != (auth.Config{}) {
		t.Errorf("se esperaba el login sin activar y sin error y se obtuvo %+v, %v, %v", cfg, activo, err)
	}
}

// RN1 · con las dos variables el login se activa; la URL se usa sin espacios ni barra final (el emisor del token se arma con ella)
func TestConfigDesdeEntorno_ConLasDosVariables_ActivaElLogin(t *testing.T) {
	cfg, activo, err := auth.ConfigDesdeEntorno(entorno(map[string]string{
		"SUPABASE_URL":      " https://abc.supabase.co/ ",
		"SUPABASE_ANON_KEY": " la-clave-anon ",
	}))

	if err != nil || !activo {
		t.Fatalf("se esperaba el login activado sin error y se obtuvo %v, %v", activo, err)
	}
	if cfg.URL != "https://abc.supabase.co" || cfg.AnonKey != "la-clave-anon" {
		t.Errorf("se esperaba la URL y la clave limpias y se obtuvo %+v", cfg)
	}
	if cfg.Emisor() != "https://abc.supabase.co/auth/v1" {
		t.Errorf("Emisor() = %q", cfg.Emisor())
	}
	if cfg.URLJWKS() != "https://abc.supabase.co/auth/v1/.well-known/jwks.json" {
		t.Errorf("URLJWKS() = %q", cfg.URLJWKS())
	}
}

// RN1 · con una sola variable la aplicación no arranca y dice cuál falta; una en blanco cuenta como que no está
func TestConfigDesdeEntorno_ConfiguracionIncompleta_DiceCualFalta(t *testing.T) {
	casos := []struct {
		nombre    string
		variables map[string]string
		esperado  string
	}{
		{"solo la URL", map[string]string{"SUPABASE_URL": "https://abc.supabase.co"}, "falta SUPABASE_ANON_KEY"},
		{"solo la clave", map[string]string{"SUPABASE_ANON_KEY": "la-clave-anon"}, "falta SUPABASE_URL"},
		{"URL y clave en blanco", map[string]string{"SUPABASE_URL": "https://abc.supabase.co", "SUPABASE_ANON_KEY": "   "}, "falta SUPABASE_ANON_KEY"},
	}
	for _, c := range casos {
		_, activo, err := auth.ConfigDesdeEntorno(entorno(c.variables))

		if err == nil || err.Error() != c.esperado || activo {
			t.Errorf("%s: se esperaba el error %q sin activar y se obtuvo %v, %v", c.nombre, c.esperado, activo, err)
		}
	}
}

// RN1 · una URL que no es http(s) también corta el arranque, en lugar de fallar en cada ingreso
func TestConfigDesdeEntorno_URLInvalida_Error(t *testing.T) {
	for _, url := range []string{"esto no es una url", "abc.supabase.co", "ftp://abc.supabase.co", "https://"} {
		_, activo, err := auth.ConfigDesdeEntorno(entorno(map[string]string{"SUPABASE_URL": url, "SUPABASE_ANON_KEY": "la-clave-anon"}))

		if err == nil || activo {
			t.Errorf("URL %q: se esperaba un error y se obtuvo %v, %v", url, activo, err)
		}
	}
}

// RN8 · el usuario con sesión viaja en el contexto del pedido
func TestIdentidad_ViajaEnElContexto(t *testing.T) {
	ctx := context.Background()
	if _, ok := auth.IdentidadDe(ctx); ok {
		t.Fatal("un contexto sin sesión no tendría que tener identidad")
	}

	ana := auth.Identidad{ID: "11111111-1111-4111-8111-111111111111", Email: "ana@mail.com"}
	i, ok := auth.IdentidadDe(auth.ConIdentidad(ctx, ana))

	if !ok || i != ana {
		t.Errorf("se esperaba %+v y se obtuvo %+v, %v", ana, i, ok)
	}
}
