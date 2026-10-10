package app_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	neturl "net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/app"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/postgres"
)

func crearProyecto(h http.Handler, nombre string) *httptest.ResponseRecorder {
	campos := neturl.Values{
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

// urlConEsquemaMigrado crea un esquema exclusivo de este test en la base TEST_DATABASE_URL, le aplica las
// migraciones y devuelve la URL que apunta a él. El esquema se borra al terminar. Sin la variable, omite el test.
func urlConEsquemaMigrado(t *testing.T) string {
	t.Helper()
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
	return urlConEsquema
}

func nuevaApp(t *testing.T, url string) http.Handler {
	t.Helper()
	h, err := app.NewHandlerConBase(url)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	return h
}

func post(h http.Handler, ruta string, campos neturl.Values) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, ruta, strings.NewReader(campos.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func get(h http.Handler, ruta string) string {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, ruta, nil))
	return rec.Body.String()
}

// Es el objetivo del Sprint 1: lo que se carga queda guardado en la base y sobrevive a un reinicio.
func TestNewHandlerConBase_ConURLGuardaEnPostgresYSobreviveAUnReinicio(t *testing.T) {
	url := urlConEsquemaMigrado(t)

	primera := nuevaApp(t, url)
	if rec := crearProyecto(primera, "Guardado en la base"); rec.Code != http.StatusSeeOther {
		t.Fatalf("crear proyecto: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}

	// "Reiniciar": una aplicación nueva, sin nada en memoria, con la misma base.
	reiniciada := nuevaApp(t, url)
	if !strings.Contains(listarProyectos(reiniciada), "Guardado en la base") {
		t.Error("el proyecto debería seguir estando después de reiniciar la aplicación")
	}
	// y la regla de nombre único sigue valiendo contra lo que hay en la base
	if rec := crearProyecto(reiniciada, "GUARDADO EN LA BASE"); rec.Code == http.StatusSeeOther {
		t.Error("un nombre repetido (con otras mayúsculas) no debería poder crearse")
	}
}

func TestNewHandlerConBase_ConURLLosSprintsSobrevivenAUnReinicio(t *testing.T) {
	url := urlConEsquemaMigrado(t)
	primera := nuevaApp(t, url)
	crearProyecto(primera, "Con sprints")

	campos := neturl.Values{"objetivo": {"Primer MVP guardado en la base"}, "fecha_inicio": {"2026-10-05"}, "fecha_fin": {"2026-10-11"}}
	if rec := post(primera, "/proyectos/1/sprints", campos); rec.Code != http.StatusSeeOther {
		t.Fatalf("crear sprint: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}

	if !strings.Contains(get(nuevaApp(t, url), "/proyectos/1/sprints"), "Primer MVP guardado en la base") {
		t.Error("el sprint debería seguir estando después de reiniciar la aplicación")
	}
}

// HU-03: los integrantes y su baja también quedan en la base, no en la memoria de la aplicación.
func TestNewHandlerConBase_ConURLLosIntegrantesSobrevivenAUnReinicio(t *testing.T) {
	url := urlConEsquemaMigrado(t)
	primera := nuevaApp(t, url)
	crearProyecto(primera, "Con integrantes")

	campos := neturl.Values{"nombre": {"Juan Pablo"}, "email": {"JuanPablo@Mail.com"}, "rol": {"AgileEnabler"}}
	if rec := post(primera, "/proyectos/1/integrantes", campos); rec.Code != http.StatusSeeOther {
		t.Fatalf("registrar integrante: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}

	reiniciada := nuevaApp(t, url)
	pagina := get(reiniciada, "/proyectos/1/integrantes")
	if !strings.Contains(pagina, "Juan Pablo") || !strings.Contains(pagina, "juanpablo@mail.com") {
		t.Errorf("el integrante debería seguir estando después de reiniciar la aplicación:\n%s", pagina)
	}

	// la baja también se guarda: el integrante sigue en la lista, como dado de baja
	if rec := post(reiniciada, "/proyectos/1/integrantes/1/baja", neturl.Values{}); rec.Code != http.StatusSeeOther {
		t.Fatalf("dar de baja: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}
	if !strings.Contains(get(nuevaApp(t, url), "/proyectos/1/integrantes"), "Dado de baja") {
		t.Error("la baja debería seguir registrada después de reiniciar la aplicación")
	}
}
