package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// conProyecto arma la aplicación con el proyecto de ejemplo ya creado (ID 1, del 05/10/2026 al 01/11/2026).
func conProyecto(t *testing.T) http.Handler {
	t.Helper()
	h := nuevoRouter()
	if rec := post(h, "/proyectos", formularioValido()); rec.Code != http.StatusSeeOther {
		t.Fatalf("no se pudo crear el proyecto de ejemplo: código %d", rec.Code)
	}
	return h
}

func sprintValido() url.Values {
	return url.Values{
		"objetivo":     {"Primer MVP"},
		"fecha_inicio": {"2026-10-05"},
		"fecha_fin":    {"2026-10-11"},
	}
}

// Hasta que exista la vista del proyecto (HU-04), a los sprints se llega desde la lista de proyectos.
func TestListaDeProyectos_TieneEnlaceALosSprintsDeCadaProyecto(t *testing.T) {
	verificarContiene(t, get(conProyecto(t), "/proyectos").Body.String(), `href="/proyectos/1/sprints"`)
}

func TestListaDeSprints_SinSprints_MuestraElProyectoYEnlaceAlFormulario(t *testing.T) {
	rec := get(conProyecto(t), "/proyectos/1/sprints")

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	verificarContiene(t, rec.Body.String(),
		"<title>Sprints · Software Metrics</title>",
		"Software Metrics", "05/10/2026", "01/11/2026", // el proyecto y su rango
		"Todavía no hay sprints",
		`href="/proyectos/1/sprints/nuevo"`,
	)
}

// El formulario limita las fechas al rango del proyecto (RN3); igual lo valida el servidor.
func TestFormularioNuevoSprint_MuestraLosCamposYElRangoDelProyecto(t *testing.T) {
	rec := get(conProyecto(t), "/proyectos/1/sprints/nuevo")

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	verificarContiene(t, rec.Body.String(),
		`action="/proyectos/1/sprints"`, `method="post"`,
		`name="objetivo"`, `name="fecha_inicio"`, `name="fecha_fin"`,
		`min="2026-10-05"`, `max="2026-11-01"`,
	)
}

// CA-11.3 · después de crear se vuelve a la lista con el número del sprint, su goal y estado Planificado
func TestCrearSprint_DatosValidos_VuelveALaListaConElSprint(t *testing.T) {
	h := conProyecto(t)

	rec := post(h, "/proyectos/1/sprints", sprintValido())

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("código = %d, se esperaba %d (redirección)", rec.Code, http.StatusSeeOther)
	}
	destino := rec.Header().Get("Location")
	if destino != "/proyectos/1/sprints?creado=1" {
		t.Fatalf("redirige a %q, se esperaba /proyectos/1/sprints?creado=1", destino)
	}
	verificarContiene(t, get(h, destino).Body.String(),
		"Sprint 1 creado", "Primer MVP", "Planificado", "05/10/2026", "11/10/2026",
	)
}

// CA-11.1 y CA-11.2 · cada error aparece en su campo y se conservan los datos cargados
func TestCrearSprint_DatosInvalidos_MuestraLosErroresYConservaLosDatos(t *testing.T) {
	h := conProyecto(t)
	campos := url.Values{
		"objetivo":     {"   "},
		"fecha_inicio": {"2026-10-04"},
		"fecha_fin":    {"2026-10-11"},
	}

	rec := post(h, "/proyectos/1/sprints", campos)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusUnprocessableEntity)
	}
	verificarContiene(t, rec.Body.String(),
		"el Sprint Goal es obligatorio",
		"la fecha tiene que estar dentro del proyecto (del 05/10/2026 al 01/11/2026)",
		"is-invalid",
		`value="2026-10-04"`, `value="2026-10-11"`,
	)
	verificarContiene(t, get(h, "/proyectos/1/sprints").Body.String(), "Todavía no hay sprints")
}

// Un proyecto que no existe (o un ID que no es un número) da 404 en las tres rutas.
func TestSprints_ProyectoInexistente_Responde404(t *testing.T) {
	h := conProyecto(t)
	pedidos := map[string]*httptest.ResponseRecorder{
		"lista":                       get(h, "/proyectos/99/sprints"),
		"formulario":                  get(h, "/proyectos/99/sprints/nuevo"),
		"creación":                    post(h, "/proyectos/99/sprints", sprintValido()),
		"lista con ID no numérico":    get(h, "/proyectos/abc/sprints"),
		"creación con ID no numérico": post(h, "/proyectos/abc/sprints", sprintValido()),
	}
	for nombre, rec := range pedidos {
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s: código = %d, se esperaba %d", nombre, rec.Code, http.StatusNotFound)
		}
	}
}

// sprintsCaidos simula la base de datos caída para los sprints; el proyecto sí existe.
type sprintsCaidos struct{}

func (sprintsCaidos) BuscarPorID(context.Context, int64) (project.Proyecto, error) {
	return project.Proyecto{ID: 1, Nombre: "Software Metrics"}, nil
}
func (sprintsCaidos) ListarPorProyecto(context.Context, int64) ([]sprint.Sprint, error) {
	return nil, errBaseCaida
}
func (sprintsCaidos) Guardar(_ context.Context, s sprint.Sprint) (sprint.Sprint, error) {
	return s, errBaseCaida
}

// Si falla la base se muestra un mensaje general, sin detalles internos.
func TestSprints_ErrorDeLaBase_MuestraUnMensajeGeneral(t *testing.T) {
	h := server.NewRouter(server.Dependencias{
		Proyectos: service.NuevoProyectos(repoCaido{}, nil),
		Sprints:   service.NuevoSprints(sprintsCaidos{}, sprintsCaidos{}),
	})
	pedidos := map[string]*httptest.ResponseRecorder{
		"lista":      get(h, "/proyectos/1/sprints"),
		"formulario": get(h, "/proyectos/1/sprints/nuevo"),
		"creación":   post(h, "/proyectos/1/sprints", sprintValido()),
	}
	for nombre, rec := range pedidos {
		if rec.Code != http.StatusInternalServerError {
			t.Errorf("%s: código = %d, se esperaba %d", nombre, rec.Code, http.StatusInternalServerError)
		}
		if strings.Contains(rec.Body.String(), "detalle interno") {
			t.Errorf("%s: la página muestra el error interno", nombre)
		}
	}
}
