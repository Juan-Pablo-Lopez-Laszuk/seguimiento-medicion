package app_test

import (
	"context"
	"encoding/json"
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
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
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

// HU-05: sin base, las historias del backlog también se guardan en memoria.
func TestNewHandlerConBase_SinURLLasHistoriasSeGuardanEnMemoria(t *testing.T) {
	h := nuevaApp(t, "")
	crearProyecto(h, "En memoria")

	campos := neturl.Values{"titulo": {"Historia en memoria"}, "prioridad": {"Media"}}
	if rec := post(h, "/proyectos/1/backlog", campos); rec.Code != http.StatusSeeOther {
		t.Fatalf("crear historia: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}

	if !strings.Contains(get(h, "/proyectos/1/backlog"), "Historia en memoria") {
		t.Error("la historia recién creada debería estar en el backlog")
	}
}

// HU-05: las historias quedan en la base, con sus story points, y la numeración sigue después de reiniciar.
func TestNewHandlerConBase_ConURLLasHistoriasSobrevivenAUnReinicio(t *testing.T) {
	url := urlConEsquemaMigrado(t)
	primera := nuevaApp(t, url)
	crearProyecto(primera, "Con backlog")

	campos := neturl.Values{"titulo": {"Historia guardada en la base"}, "prioridad": {"Alta"}, "story_points": {"13"}}
	if rec := post(primera, "/proyectos/1/backlog", campos); rec.Code != http.StatusSeeOther {
		t.Fatalf("crear historia: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}

	reiniciada := nuevaApp(t, url)
	pagina := get(reiniciada, "/proyectos/1/backlog")
	for _, esperado := range []string{"HU-1", "Historia guardada en la base", "Alta", "Pendiente", "13"} {
		if !strings.Contains(pagina, esperado) {
			t.Errorf("después de reiniciar el backlog debería mostrar %q:\n%s", esperado, pagina)
		}
	}

	// la numeración sigue de donde quedó, leyendo lo que hay en la base
	segunda := neturl.Values{"titulo": {"Segunda historia"}, "prioridad": {"Baja"}}
	if rec := post(reiniciada, "/proyectos/1/backlog", segunda); rec.Header().Get("Location") != "/proyectos/1/backlog?creada=2" {
		t.Errorf("la segunda historia tenía que ser la HU-2 y se redirigió a %q", rec.Header().Get("Location"))
	}
}

// saludDe pide /health a la app y devuelve el código y el estado de la base que informa (TEC-04).
func saludDe(t *testing.T, h http.Handler) (int, string) {
	t.Helper()
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	var cuerpo map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &cuerpo); err != nil {
		t.Fatalf("/health no respondió JSON: %v\n%s", err, rec.Body.String())
	}
	return rec.Code, cuerpo["base"]
}

// TEC-04: sin DATABASE_URL, /health avisa que la app guarda en memoria.
func TestNewHandlerConBase_SinURLHealthInformaMemoria(t *testing.T) {
	if codigo, base := saludDe(t, nuevaApp(t, "")); codigo != http.StatusOK || base != "memoria" {
		t.Errorf("/health = %d y base %q, se esperaba 200 y \"memoria\"", codigo, base)
	}
}

// TEC-04: con una DATABASE_URL que apunta a un puerto cerrado, /health avisa que no hay conexión. Es lo que
// pasaría en Vercel con la variable mal cargada.
func TestNewHandlerConBase_ConBaseInalcanzableHealthResponde503(t *testing.T) {
	h := nuevaApp(t, "postgres://usuario:clave@127.0.0.1:1/seguimiento?sslmode=disable&connect_timeout=1")

	if codigo, base := saludDe(t, h); codigo != http.StatusServiceUnavailable || base != "sin conexión" {
		t.Errorf("/health = %d y base %q, se esperaba 503 y \"sin conexión\"", codigo, base)
	}
}

// TEC-04: con la base de pruebas funcionando, /health informa que la base responde.
func TestNewHandlerConBase_ConBaseHealthInformaOK(t *testing.T) {
	h := nuevaApp(t, urlConEsquemaMigrado(t))

	if codigo, base := saludDe(t, h); codigo != http.StatusOK || base != "ok" {
		t.Errorf("/health = %d y base %q, se esperaba 200 y \"ok\"", codigo, base)
	}
}

// HU-07: sin base, los criterios de las historias también se guardan en memoria.
func TestNewHandlerConBase_SinURLLosCriteriosSeGuardanEnMemoria(t *testing.T) {
	h := nuevaApp(t, "")
	crearProyecto(h, "En memoria")
	post(h, "/proyectos/1/backlog", neturl.Values{"titulo": {"Historia en memoria"}, "prioridad": {"Media"}})

	campos := neturl.Values{"descripcion": {"Criterio en memoria"}}
	if rec := post(h, "/proyectos/1/backlog/1/criterios", campos); rec.Code != http.StatusSeeOther {
		t.Fatalf("agregar criterio: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}

	if !strings.Contains(get(h, "/proyectos/1/backlog/1"), "Criterio en memoria") {
		t.Error("el criterio recién agregado debería estar en la página de la historia")
	}
}

// HU-07: los criterios y si están cumplidos quedan en la base, y la lista del backlog sigue mostrando el progreso
// después de reiniciar la aplicación.
func TestNewHandlerConBase_ConURLLosCriteriosSobrevivenAUnReinicio(t *testing.T) {
	url := urlConEsquemaMigrado(t)
	primera := nuevaApp(t, url)
	crearProyecto(primera, "Con criterios")
	post(primera, "/proyectos/1/backlog", neturl.Values{"titulo": {"Historia con criterios"}, "prioridad": {"Alta"}})
	post(primera, "/proyectos/1/backlog/1/criterios", neturl.Values{"descripcion": {"Primer criterio guardado en la base"}})
	post(primera, "/proyectos/1/backlog/1/criterios", neturl.Values{"descripcion": {"Segundo criterio"}})
	if rec := post(primera, "/proyectos/1/backlog/1/criterios/1/cumplido", neturl.Values{"cumplido": {"1"}}); rec.Code != http.StatusSeeOther {
		t.Fatalf("marcar cumplido: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}

	reiniciada := nuevaApp(t, url)
	pagina := get(reiniciada, "/proyectos/1/backlog/1")
	for _, esperado := range []string{"Primer criterio guardado en la base", "Segundo criterio", "1/2", "Cumplido"} {
		if !strings.Contains(pagina, esperado) {
			t.Errorf("después de reiniciar la página de la historia debería mostrar %q:\n%s", esperado, pagina)
		}
	}
	if !strings.Contains(get(reiniciada, "/proyectos/1/backlog"), "1/2") {
		t.Error("después de reiniciar la lista del backlog debería mostrar el progreso 1/2")
	}

	// marcarlo de nuevo como no cumplido también se guarda
	post(reiniciada, "/proyectos/1/backlog/1/criterios/1/cumplido", neturl.Values{"cumplido": {"0"}})
	if !strings.Contains(get(nuevaApp(t, url), "/proyectos/1/backlog/1"), "0/2") {
		t.Error("desmarcar el criterio debería quedar guardado")
	}
}

// tareaDeLaApp registra una integrante y le agrega una tarea de 2,5 horas a la HU-1 del proyecto 1.
func tareaDeLaApp(t *testing.T, h http.Handler) {
	t.Helper()
	post(h, "/proyectos/1/integrantes", neturl.Values{"nombre": {"Carolina"}, "email": {"caro@mail.com"}, "rol": {"ProductBuilder"}})
	campos := neturl.Values{"titulo": {"Tarea guardada"}, "responsable": {"1"}, "horas": {"2,5"}}
	if rec := post(h, "/proyectos/1/backlog/1/tareas", campos); rec.Code != http.StatusSeeOther {
		t.Fatalf("agregar tarea: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}
}

func TestNewHandlerConBase_SinURLLasTareasSeGuardanEnMemoria(t *testing.T) {
	h := nuevaApp(t, "")
	crearProyecto(h, "En memoria")
	post(h, "/proyectos/1/backlog", neturl.Values{"titulo": {"Historia en memoria"}, "prioridad": {"Media"}})

	tareaDeLaApp(t, h)

	if !strings.Contains(get(h, "/proyectos/1/backlog/1/tareas"), "Tarea guardada") {
		t.Error("la tarea recién agregada debería estar en la página de tareas")
	}
}

// HU-25: las tareas, sus horas y si están completadas quedan en la base y sobreviven a un reinicio.
func TestNewHandlerConBase_ConURLLasTareasSobrevivenAUnReinicio(t *testing.T) {
	url := urlConEsquemaMigrado(t)
	primera := nuevaApp(t, url)
	crearProyecto(primera, "Con tareas")
	post(primera, "/proyectos/1/backlog", neturl.Values{"titulo": {"Historia con tareas"}, "prioridad": {"Alta"}})
	tareaDeLaApp(t, primera)
	if rec := post(primera, "/proyectos/1/backlog/1/tareas/1/horas", neturl.Values{"horas": {"4"}}); rec.Code != http.StatusSeeOther {
		t.Fatalf("cambiar horas: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}
	if rec := post(primera, "/proyectos/1/backlog/1/tareas/1/completada", neturl.Values{"completada": {"1"}}); rec.Code != http.StatusSeeOther {
		t.Fatalf("completar: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}

	pagina := get(nuevaApp(t, url), "/proyectos/1/backlog/1/tareas")
	for _, esperado := range []string{"Tarea guardada", "Carolina", "4 h", "Completada"} {
		if !strings.Contains(pagina, esperado) {
			t.Errorf("después de reiniciar la página de tareas debería mostrar %q:\n%s", esperado, pagina)
		}
	}
}

// HU-08: sin base, el cambio de estado también funciona en memoria. Hasta que exista HU-12 no hay forma de dejar una
// historia En sprint desde la aplicación, así que se comprueba la regla con una historia Pendiente.
func TestNewHandlerConBase_SinURLElCambioDeEstadoAplicaLaRegla(t *testing.T) {
	h := nuevaApp(t, "")
	crearProyecto(h, "En memoria")
	post(h, "/proyectos/1/backlog", neturl.Values{"titulo": {"Historia en memoria"}, "prioridad": {"Media"}})

	rec := post(h, "/proyectos/1/backlog/1/estado", neturl.Values{"estado": {"En progreso"}})

	if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), "no se puede pasar de Pendiente a En progreso") {
		t.Errorf("se esperaba el 409 con el motivo de la transición y se obtuvo %d:\n%s", rec.Code, rec.Body.String())
	}
}

// HU-08: el estado queda en la base y sobrevive a un reinicio de la aplicación.
func TestNewHandlerConBase_ConURLElEstadoDeLaHistoriaSobreviveAUnReinicio(t *testing.T) {
	url := urlConEsquemaMigrado(t)
	primera := nuevaApp(t, url)
	crearProyecto(primera, "Con estados")
	post(primera, "/proyectos/1/backlog", neturl.Values{"titulo": {"Historia con estado"}, "prioridad": {"Alta"}})
	// Pasar a En sprint es de HU-12: se deja la historia así directo en la base, como lo hará esa historia.
	pool, err := postgres.Conectar(context.Background(), url)
	if err != nil {
		t.Fatalf("conectar a la base de pruebas: %v", err)
	}
	defer pool.Close()
	if err := postgres.NuevoHistorias(pool).ActualizarEstado(context.Background(), 1, backlog.EstadoEnSprint); err != nil {
		t.Fatalf("no se pudo dejar la historia En sprint: %v", err)
	}

	if rec := post(primera, "/proyectos/1/backlog/1/estado", neturl.Values{"estado": {"En progreso"}}); rec.Code != http.StatusSeeOther {
		t.Fatalf("cambiar el estado: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}

	reiniciada := nuevaApp(t, url)
	if !strings.Contains(get(reiniciada, "/proyectos/1/backlog/1"), "En progreso") {
		t.Error("después de reiniciar la página de la historia tendría que mostrar En progreso")
	}
	if !strings.Contains(get(reiniciada, "/proyectos/1/backlog"), `text-bg-primary">En progreso`) {
		t.Error("después de reiniciar la lista del backlog tendría que mostrar la historia En progreso")
	}
}

// HU-16: sin base, las estimaciones también se guardan en memoria.
func TestNewHandlerConBase_SinURLLasEstimacionesSeGuardanEnMemoria(t *testing.T) {
	h := nuevaApp(t, "")
	crearProyecto(h, "En memoria")
	post(h, "/proyectos/1/backlog", neturl.Values{"titulo": {"Historia en memoria"}, "prioridad": {"Media"}})

	if rec := post(h, "/proyectos/1/backlog/1/estimacion", neturl.Values{"story_points": {"8"}}); rec.Code != http.StatusSeeOther {
		t.Fatalf("estimar: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}

	if !strings.Contains(get(h, "/proyectos/1/backlog/1"), "Story points: 8") {
		t.Error("la estimación recién guardada debería verse en la página de la historia")
	}
}

// HU-16: la estimación y su cambio quedan en la base y sobreviven a un reinicio de la aplicación.
func TestNewHandlerConBase_ConURLLaEstimacionSobreviveAUnReinicio(t *testing.T) {
	url := urlConEsquemaMigrado(t)
	primera := nuevaApp(t, url)
	crearProyecto(primera, "Con estimaciones")
	post(primera, "/proyectos/1/backlog", neturl.Values{"titulo": {"Historia sin estimar"}, "prioridad": {"Alta"}})

	for _, valor := range []string{"5", "13"} { // estimar y re-estimar
		if rec := post(primera, "/proyectos/1/backlog/1/estimacion", neturl.Values{"story_points": {valor}}); rec.Code != http.StatusSeeOther {
			t.Fatalf("estimar en %s: se esperaba %d y se obtuvo %d", valor, http.StatusSeeOther, rec.Code)
		}
	}

	reiniciada := nuevaApp(t, url)
	if !strings.Contains(get(reiniciada, "/proyectos/1/backlog/1"), "Story points: 13") {
		t.Error("después de reiniciar la página de la historia tendría que mostrar 13 story points")
	}
	if rec := post(reiniciada, "/proyectos/1/backlog/1/estimacion", neturl.Values{"story_points": {"4"}}); rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("un valor fuera de la escala: se esperaba %d y se obtuvo %d", http.StatusUnprocessableEntity, rec.Code)
	}
	if !strings.Contains(get(nuevaApp(t, url), "/proyectos/1/backlog/1"), "Story points: 13") {
		t.Error("un valor inválido no tenía que cambiar la estimación guardada")
	}
}

// HU-06: sin base, editar y eliminar una historia también funcionan en memoria, con sus criterios y tareas.
func TestNewHandlerConBase_SinURLSePuedeEditarYEliminarUnaHistoria(t *testing.T) {
	h := nuevaApp(t, "")
	crearProyecto(h, "En memoria")
	post(h, "/proyectos/1/backlog", neturl.Values{"titulo": {"Historia en memoria"}, "prioridad": {"Media"}})
	post(h, "/proyectos/1/backlog/1/criterios", neturl.Values{"descripcion": {"Un criterio"}})
	tareaDeLaApp(t, h)

	editar := post(h, "/proyectos/1/backlog/1/editar", neturl.Values{"titulo": {"Título corregido"}, "descripcion": {""}, "prioridad": {"Baja"}})
	if editar.Code != http.StatusSeeOther {
		t.Fatalf("editar: se esperaba %d y se obtuvo %d", http.StatusSeeOther, editar.Code)
	}
	if !strings.Contains(get(h, "/proyectos/1/backlog/1"), "Título corregido") {
		t.Error("la historia editada debería mostrar el título nuevo")
	}

	eliminar := post(h, "/proyectos/1/backlog/1/eliminar", nil)
	if eliminar.Code != http.StatusSeeOther {
		t.Fatalf("eliminar: se esperaba %d y se obtuvo %d", http.StatusSeeOther, eliminar.Code)
	}
	if !strings.Contains(get(h, "/proyectos/1/backlog/1"), "No encontramos esa historia") {
		t.Error("la historia eliminada ya no debería existir")
	}
}

// HU-06: lo editado y lo eliminado quedan en la base. Eliminar una historia con criterios y tareas funciona (la base no
// tiene ON DELETE CASCADE: los borra el repositorio en la misma transacción) y su número queda libre.
func TestNewHandlerConBase_ConURLLaEdicionYLaEliminacionSobrevivenAUnReinicio(t *testing.T) {
	url := urlConEsquemaMigrado(t)
	primera := nuevaApp(t, url)
	crearProyecto(primera, "Con ediciones")
	post(primera, "/proyectos/1/backlog", neturl.Values{"titulo": {"Historia que se edita"}, "prioridad": {"Alta"}})
	post(primera, "/proyectos/1/backlog", neturl.Values{"titulo": {"Historia que se elimina"}, "prioridad": {"Alta"}})
	post(primera, "/proyectos/1/backlog/2/criterios", neturl.Values{"descripcion": {"Un criterio"}})
	post(primera, "/proyectos/1/integrantes", neturl.Values{"nombre": {"Carolina"}, "email": {"caro@mail.com"}, "rol": {"ProductBuilder"}})
	post(primera, "/proyectos/1/backlog/2/tareas", neturl.Values{"titulo": {"Una tarea"}, "responsable": {"1"}, "horas": {"2"}})

	if rec := post(primera, "/proyectos/1/backlog/1/editar", neturl.Values{"titulo": {"Historia editada"}, "descripcion": {"Nueva"}, "prioridad": {"Baja"}}); rec.Code != http.StatusSeeOther {
		t.Fatalf("editar: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}
	if rec := post(primera, "/proyectos/1/backlog/2/eliminar", nil); rec.Code != http.StatusSeeOther {
		t.Fatalf("eliminar: se esperaba %d y se obtuvo %d:\n%s", http.StatusSeeOther, rec.Code, rec.Body.String())
	}

	reiniciada := nuevaApp(t, url)
	lista := get(reiniciada, "/proyectos/1/backlog")
	if !strings.Contains(lista, "Historia editada") || strings.Contains(lista, "Historia que se elimina") {
		t.Errorf("después de reiniciar el backlog tendría que tener la historia editada y no la eliminada:\n%s", lista)
	}
	if !strings.Contains(get(reiniciada, "/proyectos/1/backlog/1"), "Baja") {
		t.Error("después de reiniciar la prioridad editada tendría que seguir")
	}
	// el número de la última historia queda libre
	nueva := post(reiniciada, "/proyectos/1/backlog", neturl.Values{"titulo": {"Otra"}, "prioridad": {"Media"}})
	if destino := nueva.Header().Get("Location"); destino != "/proyectos/1/backlog?creada=2" {
		t.Errorf("la historia nueva redirige a %q, se esperaba ?creada=2 (HU-2)", destino)
	}
}
