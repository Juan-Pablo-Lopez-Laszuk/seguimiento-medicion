package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

func integranteValido() url.Values {
	return url.Values{
		"nombre": {"Juan Pablo"},
		"email":  {" JP@Mail.com "},
		"rol":    {string(member.RolAgileEnabler)},
	}
}

// Hasta que exista la vista del proyecto (HU-04), a los integrantes se llega desde la lista de proyectos.
func TestListaDeProyectos_TieneEnlaceALosIntegrantesDeCadaProyecto(t *testing.T) {
	verificarContiene(t, get(conProyecto(t), "/proyectos").Body.String(), `href="/proyectos/1/integrantes"`)
}

func TestListaDeIntegrantes_SinIntegrantes_MuestraElProyectoYEnlaceAlFormulario(t *testing.T) {
	rec := get(conProyecto(t), "/proyectos/1/integrantes")

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	verificarContiene(t, rec.Body.String(),
		"<title>Integrantes · Software Metrics</title>",
		"Software Metrics",
		"Todavía no hay integrantes",
		`href="/proyectos/1/integrantes/nuevo"`,
	)
}

func TestFormularioNuevoIntegrante_MuestraLosCamposYLosRoles(t *testing.T) {
	rec := get(conProyecto(t), "/proyectos/1/integrantes/nuevo")

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	verificarContiene(t, rec.Body.String(),
		`action="/proyectos/1/integrantes"`, `method="post"`,
		`name="nombre"`, `name="email"`, `type="email"`, `name="rol"`,
		`value="AgileEnabler"`, "Agile Enabler", `value="ProductBuilder"`, "Product Builder",
	)
}

// CA-03.1 y CA-03.2 · después de registrar se vuelve a la lista con el integrante activo y el email normalizado
func TestRegistrarIntegrante_DatosValidos_VuelveALaListaConElIntegrante(t *testing.T) {
	h := conProyecto(t)

	rec := post(h, "/proyectos/1/integrantes", integranteValido())

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("código = %d, se esperaba %d (redirección)", rec.Code, http.StatusSeeOther)
	}
	destino := rec.Header().Get("Location")
	if destino != "/proyectos/1/integrantes?registrado=1" {
		t.Fatalf("redirige a %q, se esperaba /proyectos/1/integrantes?registrado=1", destino)
	}
	verificarContiene(t, get(h, destino).Body.String(),
		"Integrante registrado", "Juan Pablo", "jp@mail.com", "Agile Enabler", "Activo",
		`action="/proyectos/1/integrantes/1/baja"`,
	)
}

// CA-03.1 y CA-03.2 · cada error aparece en su campo y se conservan los datos cargados, incluido el rol elegido
func TestRegistrarIntegrante_DatosInvalidos_MuestraLosErroresYConservaLosDatos(t *testing.T) {
	h := conProyecto(t)
	post(h, "/proyectos/1/integrantes", integranteValido())
	campos := url.Values{"nombre": {"<b>Ana</b>"}, "email": {"jp@mail.com"}, "rol": {string(member.RolAgileEnabler)}}

	rec := post(h, "/proyectos/1/integrantes", campos)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusUnprocessableEntity)
	}
	verificarContiene(t, rec.Body.String(),
		"ya hay un integrante con ese email en el proyecto",
		"el proyecto ya tiene un Agile Enabler activo",
		"is-invalid",
		`value="jp@mail.com"`,
		`value="AgileEnabler" selected`,
		"&lt;b&gt;Ana&lt;/b&gt;", // lo que escribió el usuario se muestra escapado
	)
}

// CA-03.3 · la baja deja al integrante en la lista como "Dado de baja" y sin el botón de baja
func TestDarDeBaja_VuelveALaListaConElIntegranteDadoDeBaja(t *testing.T) {
	h := conProyecto(t)
	post(h, "/proyectos/1/integrantes", integranteValido())

	rec := post(h, "/proyectos/1/integrantes/1/baja", nil)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("código = %d, se esperaba %d (redirección)", rec.Code, http.StatusSeeOther)
	}
	destino := rec.Header().Get("Location")
	if destino != "/proyectos/1/integrantes?baja=1" {
		t.Fatalf("redirige a %q, se esperaba /proyectos/1/integrantes?baja=1", destino)
	}
	html := get(h, destino).Body.String()
	verificarContiene(t, html, "Integrante dado de baja", "Juan Pablo", "Dado de baja")
	if strings.Contains(html, `action="/proyectos/1/integrantes/1/baja"`) {
		t.Error("un integrante dado de baja no debería tener el botón de baja")
	}
}

// RN8 · dar de baja dos veces muestra el aviso en la lista
func TestDarDeBaja_DosVeces_MuestraElAviso(t *testing.T) {
	h := conProyecto(t)
	post(h, "/proyectos/1/integrantes", integranteValido())
	post(h, "/proyectos/1/integrantes/1/baja", nil)

	rec := post(h, "/proyectos/1/integrantes/1/baja", nil)

	if rec.Code != http.StatusConflict {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusConflict)
	}
	verificarContiene(t, rec.Body.String(), "el integrante ya está dado de baja", "Juan Pablo")
}

// Un proyecto o integrante que no existe (o un ID que no es un número) da 404.
func TestIntegrantes_NoEncontrado_Responde404(t *testing.T) {
	h := conProyecto(t)
	post(h, "/proyectos/1/integrantes", integranteValido())
	pedidos := map[string]*httptest.ResponseRecorder{
		"lista":                        get(h, "/proyectos/99/integrantes"),
		"formulario":                   get(h, "/proyectos/99/integrantes/nuevo"),
		"registro":                     post(h, "/proyectos/99/integrantes", integranteValido()),
		"lista con ID no numérico":     get(h, "/proyectos/abc/integrantes"),
		"baja de integrante que falta": post(h, "/proyectos/1/integrantes/99/baja", nil),
		"baja en otro proyecto":        post(h, "/proyectos/99/integrantes/1/baja", nil),
		"baja con ID no numérico":      post(h, "/proyectos/1/integrantes/abc/baja", nil),
	}
	for nombre, rec := range pedidos {
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s: código = %d, se esperaba %d", nombre, rec.Code, http.StatusNotFound)
		}
	}
}

// proyectoQueExiste devuelve siempre el proyecto 1, para probar la base caída solo en los integrantes.
type proyectoQueExiste struct{}

func (proyectoQueExiste) BuscarPorID(context.Context, int64) (project.Proyecto, error) {
	return project.Proyecto{ID: 1, Nombre: "Software Metrics"}, nil
}

// integrantesCaidos simula la base de datos caída para los integrantes.
type integrantesCaidos struct{}

func (integrantesCaidos) ListarPorProyecto(context.Context, int64) ([]member.Integrante, error) {
	return nil, errBaseCaida
}
func (integrantesCaidos) Guardar(_ context.Context, i member.Integrante) (member.Integrante, error) {
	return i, errBaseCaida
}
func (integrantesCaidos) BuscarPorID(context.Context, int64, int64) (member.Integrante, error) {
	return member.Integrante{}, errBaseCaida
}
func (integrantesCaidos) Actualizar(context.Context, member.Integrante) error { return errBaseCaida }

// Si falla la base se muestra un mensaje general, sin detalles internos.
func TestIntegrantes_ErrorDeLaBase_MuestraUnMensajeGeneral(t *testing.T) {
	h := server.NewRouter(server.Dependencias{
		Proyectos:   service.NuevoProyectos(repoCaido{}, nil),
		Sprints:     service.NuevoSprints(sprintsCaidos{}, sprintsCaidos{}),
		Integrantes: service.NuevoIntegrantes(proyectoQueExiste{}, integrantesCaidos{}),
	})
	pedidos := map[string]*httptest.ResponseRecorder{
		"lista":      get(h, "/proyectos/1/integrantes"),
		"formulario": get(h, "/proyectos/1/integrantes/nuevo"),
		"registro":   post(h, "/proyectos/1/integrantes", integranteValido()),
		"baja":       post(h, "/proyectos/1/integrantes/1/baja", nil),
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
