package app_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/app"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/postgres"
)

func crearProyecto(h http.Handler, nombre string) *httptest.ResponseRecorder {
	campos := url.Values{
		"nombre":       {nombre},
		"descripcion":  {"TPI de ICS"},
		"fecha_inicio": {"2026-10-05"},
		"fecha_fin":    {"2026-11-01"},
	}
	req := httptest.NewRequest(http.MethodPost, "/proyectos", strings.NewReader(campos.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func listarProyectos(h http.Handler) string {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/proyectos", nil))
	return rec.Body.String()
}

// Sin DATABASE_URL la aplicación guarda en memoria: así se puede desarrollar y probar sin base de datos.
func TestNewHandlerConBase_SinURLGuardaEnMemoria(t *testing.T) {
	h, err := app.NewHandlerConBase("")
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if rec := crearProyecto(h, "En memoria"); rec.Code != http.StatusSeeOther {
		t.Fatalf("crear proyecto: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}
	if !strings.Contains(listarProyectos(h), "En memoria") {
		t.Error("el proyecto recién creado debería estar en la lista")
	}
}

func TestNewHandlerConBase_URLInvalidaDevuelveError(t *testing.T) {
	if _, err := app.NewHandlerConBase("esto no es una url de postgres %zz"); err == nil {
		t.Error("se esperaba un error por la URL inválida")
	}
}

// Es el objetivo del Sprint 1: lo que se carga queda guardado en la base y sobrevive a un reinicio.
func TestNewHandlerConBase_ConURLGuardaEnPostgresYSobreviveAUnReinicio(t *testing.T) {
	base := os.Getenv("TEST_DATABASE_URL")
	if base == "" {
		t.Skip("TEST_DATABASE_URL no está definida: se omiten los tests con base de datos")
	}
	ctx := context.Background()
	esquema := fmt.Sprintf("test_app_%d", time.Now().UnixNano())
	separador := "?"
	if strings.Contains(base, "?") {
		separador = "&"
	}

	admin, err := pgxpool.New(ctx, base)
	if err != nil {
		t.Fatalf("conectar a la base de pruebas: %v", err)
	}
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+esquema); err != nil {
		t.Fatalf("crear el esquema de prueba: %v", err)
	}
	t.Cleanup(func() {
		_, _ = admin.Exec(ctx, "DROP SCHEMA "+esquema+" CASCADE")
		admin.Close()
	})
	urlConEsquema := base + separador + "search_path=" + esquema

	pool, err := postgres.Conectar(ctx, urlConEsquema)
	if err != nil {
		t.Fatalf("conectar al esquema de prueba: %v", err)
	}
	defer pool.Close()
	if err := postgres.Subir(ctx, pool); err != nil {
		t.Fatalf("aplicar las migraciones: %v", err)
	}

	primera, err := app.NewHandlerConBase(urlConEsquema)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if rec := crearProyecto(primera, "Guardado en la base"); rec.Code != http.StatusSeeOther {
		t.Fatalf("crear proyecto: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}

	// "Reiniciar": una aplicación nueva, sin nada en memoria, con la misma base.
	reiniciada, err := app.NewHandlerConBase(urlConEsquema)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if !strings.Contains(listarProyectos(reiniciada), "Guardado en la base") {
		t.Error("el proyecto debería seguir estando después de reiniciar la aplicación")
	}
	// y la regla de nombre único sigue valiendo contra lo que hay en la base
	if rec := crearProyecto(reiniciada, "GUARDADO EN LA BASE"); rec.Code == http.StatusSeeOther {
		t.Error("un nombre repetido (con otras mayúsculas) no debería poder crearse")
	}
}
