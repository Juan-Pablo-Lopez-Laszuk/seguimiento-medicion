package server_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// get y post simulan pedidos al mismo router, así lo que se crea en un pedido se ve en el siguiente.
func get(h http.Handler, ruta string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, ruta, nil))
	return rec
}

func post(h http.Handler, ruta string, campos url.Values) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, ruta, strings.NewReader(campos.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func formularioValido() url.Values {
	return url.Values{
		"nombre":       {"Software Metrics"},
		"descripcion":  {"TPI de ICS"},
		"fecha_inicio": {"2026-10-05"},
		"fecha_fin":    {"2026-11-01"},
	}
}

func verificarContiene(t *testing.T, html string, esperados ...string) {
	t.Helper()
	for _, e := range esperados {
		if !strings.Contains(html, e) {
			t.Errorf("la página no contiene %q", e)
		}
	}
}

func TestListaDeProyectos_SinProyectos_MuestraAvisoYEnlaceAlFormulario(t *testing.T) {
	rec := get(nuevoRouter(), "/proyectos")

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	verificarContiene(t, rec.Body.String(),
		"<title>Proyectos · Software Metrics</title>",
		"Todavía no hay proyectos",
		`href="/proyectos/nuevo"`,
		`href="/proyectos"`, // enlace en la barra de navegación
	)
}

func TestFormularioNuevoProyecto_MuestraLosCampos(t *testing.T) {
	rec := get(nuevoRouter(), "/proyectos/nuevo")

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	verificarContiene(t, rec.Body.String(),
		`action="/proyectos"`, `method="post"`,
		`name="nombre"`, `name="descripcion"`,
		`name="fecha_inicio"`, `name="fecha_fin"`, `type="date"`,
	)
}

// CA-01.3 · después de crear se vuelve a la lista con el mensaje y el proyecto nuevo visible
func TestCrearProyecto_DatosValidos_VuelveALaListaConElProyecto(t *testing.T) {
	h := nuevoRouter()

	rec := post(h, "/proyectos", formularioValido())

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("código = %d, se esperaba %d (redirección)", rec.Code, http.StatusSeeOther)
	}
	destino := rec.Header().Get("Location")
	if destino != "/proyectos?creado=1" {
		t.Fatalf("redirige a %q, se esperaba /proyectos?creado=1", destino)
	}
	verificarContiene(t, get(h, destino).Body.String(),
		"Proyecto creado", "Software Metrics", "TPI de ICS", "Planificado", "05/10/2026", "01/11/2026",
	)
}

// CA-01.4 · cada error aparece en su campo y se conservan los datos cargados
func TestCrearProyecto_DatosInvalidos_MuestraLosErroresYConservaLosDatos(t *testing.T) {
	h := nuevoRouter()
	campos := url.Values{
		"nombre":       {"ab"},
		"descripcion":  {"<b>hola</b>"},
		"fecha_inicio": {""},
		"fecha_fin":    {"2026-11-01"},
	}

	rec := post(h, "/proyectos", campos)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusUnprocessableEntity)
	}
	verificarContiene(t, rec.Body.String(),
		"el nombre debe tener entre 3 y 100 caracteres",
		"la fecha de inicio es obligatoria y debe ser válida",
		"is-invalid",
		`value="ab"`, `value="2026-11-01"`,
		"&lt;b&gt;hola&lt;/b&gt;", // lo que escribió el usuario se muestra escapado, nunca como HTML
	)
	verificarContiene(t, get(h, "/proyectos").Body.String(), "Todavía no hay proyectos")
}

// CA-01.1 · el nombre repetido se informa en el formulario
func TestCrearProyecto_NombreRepetido_MuestraElError(t *testing.T) {
	h := nuevoRouter()
	post(h, "/proyectos", formularioValido())

	rec := post(h, "/proyectos", formularioValido())

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusUnprocessableEntity)
	}
	verificarContiene(t, rec.Body.String(), "ya existe un proyecto con ese nombre")
}

// repoCaido simula la base de datos caída.
type repoCaido struct{}

var errBaseCaida = errors.New("base caída: detalle interno")

func (repoCaido) ExisteNombre(context.Context, string) (bool, error) { return false, errBaseCaida }
func (repoCaido) Guardar(_ context.Context, p project.Proyecto) (project.Proyecto, error) {
	return p, errBaseCaida
}
func (repoCaido) Listar(context.Context) ([]project.Proyecto, error) { return nil, errBaseCaida }

// Si falla la base se muestra un mensaje general, sin detalles internos, y se conservan los datos del formulario.
func TestProyectos_ErrorDeLaBase_MuestraUnMensajeGeneral(t *testing.T) {
	h := server.NewRouter(server.Dependencias{Proyectos: service.NuevoProyectos(repoCaido{}, time.Now)})

	lista := get(h, "/proyectos")
	creacion := post(h, "/proyectos", formularioValido())

	for nombre, rec := range map[string]*httptest.ResponseRecorder{"lista": lista, "creación": creacion} {
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("%s: código = %d, se esperaba %d", nombre, rec.Code, http.StatusInternalServerError)
		}
		if strings.Contains(rec.Body.String(), "detalle interno") {
			t.Errorf("%s: la página muestra el error interno", nombre)
		}
	}
	verificarContiene(t, creacion.Body.String(), "No se pudo guardar el proyecto", `value="Software Metrics"`)
}
