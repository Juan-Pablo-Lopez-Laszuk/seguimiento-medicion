package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

func historiaValida() url.Values {
	return url.Values{
		"titulo":       {"Iniciar sesión con Google"},
		"descripcion":  {"Para no tener que recordar otra contraseña."},
		"prioridad":    {"Alta"},
		"story_points": {"5"},
	}
}

// Hasta que exista la vista del proyecto (HU-04), al backlog se llega desde la lista de proyectos.
func TestListaDeProyectos_TieneEnlaceAlBacklogDeCadaProyecto(t *testing.T) {
	verificarContiene(t, get(conProyecto(t), "/proyectos").Body.String(), `href="/proyectos/1/backlog"`)
}

func TestListaDelBacklog_SinHistorias_MuestraElProyectoYEnlaceAlFormulario(t *testing.T) {
	rec := get(conProyecto(t), "/proyectos/1/backlog")

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	verificarContiene(t, rec.Body.String(),
		"<title>Backlog · Software Metrics</title>",
		"Software Metrics",
		"Todavía no hay historias",
		`href="/proyectos/1/backlog/nuevo"`,
	)
}

func TestFormularioNuevaHistoria_MuestraLosCampos(t *testing.T) {
	rec := get(conProyecto(t), "/proyectos/1/backlog/nuevo")

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	verificarContiene(t, rec.Body.String(),
		`action="/proyectos/1/backlog"`, `method="post"`,
		`name="titulo"`, `name="descripcion"`, `name="prioridad"`, `name="story_points"`,
		`value="Alta"`, `value="Media"`, `value="Baja"`,
		"Sin estimar", `value="0"`, `value="8"`, `value="21"`, // la escala acordada
	)
}

// CA-05.1 y CA-05.3 · después de crear se vuelve a la lista con el identificador, la prioridad, el estado y los
// story points de la historia
func TestCrearHistoria_DatosValidos_VuelveALaListaConLaHistoria(t *testing.T) {
	h := conProyecto(t)

	rec := post(h, "/proyectos/1/backlog", historiaValida())

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("código = %d, se esperaba %d (redirección)", rec.Code, http.StatusSeeOther)
	}
	destino := rec.Header().Get("Location")
	if destino != "/proyectos/1/backlog?creada=1" {
		t.Fatalf("redirige a %q, se esperaba /proyectos/1/backlog?creada=1", destino)
	}
	verificarContiene(t, get(h, destino).Body.String(),
		"Historia HU-1 creada", "HU-1", "Iniciar sesión con Google", "Alta", "Pendiente", "5",
	)
}

// Sin estimar y estimada en 0 se ven distinto en la lista
func TestListaDelBacklog_DiferenciaSinEstimarDeEstimadaEnCero(t *testing.T) {
	h := conProyecto(t)
	sinEstimar := historiaValida()
	sinEstimar.Set("titulo", "Historia sin estimar")
	sinEstimar.Set("story_points", "")
	enCero := historiaValida()
	enCero.Set("titulo", "Historia en cero")
	enCero.Set("story_points", "0")
	post(h, "/proyectos/1/backlog", sinEstimar)
	post(h, "/proyectos/1/backlog", enCero)

	html := get(h, "/proyectos/1/backlog").Body.String()

	verificarContiene(t, html, "Historia sin estimar", "Historia en cero", "Sin estimar", `<td class="text-nowrap">0</td>`)
	if strings.Count(html, "Sin estimar") != 1 {
		t.Errorf("solo una historia está sin estimar y la página lo dice %d veces", strings.Count(html, "Sin estimar"))
	}
}

// El título es texto del usuario: se escapa, no se interpreta como HTML
func TestListaDelBacklog_EscapaElTituloDelUsuario(t *testing.T) {
	h := conProyecto(t)
	campos := historiaValida()
	campos.Set("titulo", "<script>alert(1)</script>")
	post(h, "/proyectos/1/backlog", campos)

	html := get(h, "/proyectos/1/backlog").Body.String()

	if strings.Contains(html, "<script>alert(1)</script>") {
		t.Error("el título se mostró sin escapar")
	}
	verificarContiene(t, html, "&lt;script&gt;alert(1)&lt;/script&gt;")
}

// CA-05.2 y CA-05.3 · cada error aparece en su campo y se conservan los datos cargados
func TestCrearHistoria_DatosInvalidos_MuestraLosErroresYConservaLosDatos(t *testing.T) {
	h := conProyecto(t)
	campos := url.Values{
		"titulo":       {"   "},
		"descripcion":  {"Mi descripción"},
		"prioridad":    {"Urgente"},
		"story_points": {"4"},
	}

	rec := post(h, "/proyectos/1/backlog", campos)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusUnprocessableEntity)
	}
	verificarContiene(t, rec.Body.String(),
		"el título es obligatorio",
		"la prioridad tiene que ser Alta, Media o Baja",
		"los story points tienen que ser 0, 1, 2, 3, 5, 8, 13 o 21",
		"is-invalid",
		"Mi descripción",
	)
	verificarContiene(t, get(h, "/proyectos/1/backlog").Body.String(), "Todavía no hay historias")
}

// Lo que sí era válido (la prioridad y los story points elegidos) sigue elegido en el formulario
func TestCrearHistoria_ConUnErrorEnElTitulo_ConservaLaPrioridadYLosStoryPoints(t *testing.T) {
	campos := historiaValida()
	campos.Set("titulo", "")
	campos.Set("prioridad", "Baja")
	campos.Set("story_points", "8")

	rec := post(conProyecto(t), "/proyectos/1/backlog", campos)

	verificarContiene(t, rec.Body.String(), `<option value="Baja" selected>`, `<option value="8" selected>`)
}

// El proyecto que no existe (o un ID que no es un número) da 404 en las tres rutas.
func TestHistorias_ProyectoInexistente_Responde404(t *testing.T) {
	h := conProyecto(t)
	pedidos := map[string]*httptest.ResponseRecorder{
		"lista":                       get(h, "/proyectos/99/backlog"),
		"formulario":                  get(h, "/proyectos/99/backlog/nuevo"),
		"creación":                    post(h, "/proyectos/99/backlog", historiaValida()),
		"lista con ID no numérico":    get(h, "/proyectos/abc/backlog"),
		"creación con ID no numérico": post(h, "/proyectos/abc/backlog", historiaValida()),
	}
	for nombre, rec := range pedidos {
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s: código = %d, se esperaba %d", nombre, rec.Code, http.StatusNotFound)
		}
	}
}

// historiasCaidas simula la base de datos caída para las historias; el proyecto sí existe.
type historiasCaidas struct{}

func (historiasCaidas) BuscarPorID(context.Context, int64) (project.Proyecto, error) {
	return project.Proyecto{ID: 1, Nombre: "Software Metrics"}, nil
}
func (historiasCaidas) ListarPorProyecto(context.Context, int64) ([]backlog.Historia, error) {
	return nil, errBaseCaida
}
func (historiasCaidas) Guardar(_ context.Context, h backlog.Historia) (backlog.Historia, error) {
	return h, errBaseCaida
}

// Si falla la base se muestra un mensaje general, sin detalles internos.
func TestHistorias_ErrorDeLaBase_MuestraUnMensajeGeneral(t *testing.T) {
	h := server.NewRouter(server.Dependencias{
		Proyectos: service.NuevoProyectos(repoCaido{}, nil),
		Historias: service.NuevoHistorias(historiasCaidas{}, historiasCaidas{}),
	})
	pedidos := map[string]*httptest.ResponseRecorder{
		"lista":      get(h, "/proyectos/1/backlog"),
		"formulario": get(h, "/proyectos/1/backlog/nuevo"),
		"creación":   post(h, "/proyectos/1/backlog", historiaValida()),
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
