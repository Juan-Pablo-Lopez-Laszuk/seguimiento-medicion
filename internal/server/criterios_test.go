package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// conHistoria arma la aplicación con el proyecto de ejemplo (ID 1) y su HU-1 "Iniciar sesión con Google" ya creados.
func conHistoria(t *testing.T) http.Handler {
	t.Helper()
	h := conProyecto(t)
	if rec := post(h, "/proyectos/1/backlog", historiaValida()); rec.Code != http.StatusSeeOther {
		t.Fatalf("no se pudo crear la historia de ejemplo: código %d", rec.Code)
	}
	return h
}

func criterioValido(descripcion string) url.Values {
	return url.Values{"descripcion": {descripcion}}
}

func marca(valor string) url.Values {
	return url.Values{"cumplido": {valor}}
}

// CA-07.3 · la lista del backlog muestra el progreso de cada historia y enlaza a su página
func TestListaDelBacklog_MuestraElProgresoDeCadaHistoriaYEnlazaALaPagina(t *testing.T) {
	h := conHistoria(t)
	segunda := historiaValida()
	segunda.Set("titulo", "Ver el tablero")
	post(h, "/proyectos/1/backlog", segunda)
	post(h, "/proyectos/1/backlog/1/criterios", criterioValido("Entra con su cuenta"))
	post(h, "/proyectos/1/backlog/1/criterios", criterioValido("Muestra su nombre"))
	post(h, "/proyectos/1/backlog/1/criterios/1/cumplido", marca("1"))

	html := get(h, "/proyectos/1/backlog").Body.String()

	verificarContiene(t, html, `href="/proyectos/1/backlog/1"`, `href="/proyectos/1/backlog/2"`, "1/2", "Sin criterios")
}

// CA-07.1 y CA-07.3 · la página de una historia sin criterios lo dice y avisa que hace falta al menos uno
func TestPaginaDeLaHistoria_SinCriterios_AvisaQueNecesitaAlMenosUno(t *testing.T) {
	rec := get(conHistoria(t), "/proyectos/1/backlog/1")

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	verificarContiene(t, rec.Body.String(),
		"<title>HU-1 · Software Metrics</title>",
		"HU-1", "Iniciar sesión con Google", "Alta", "Pendiente",
		"Sin criterios todavía", "al menos uno para entrar a un sprint",
		`action="/proyectos/1/backlog/1/criterios"`, `method="post"`, `name="descripcion"`,
		`href="/proyectos/1/backlog"`, // para volver a la lista
	)
}

// CA-07.2 y CA-07.3 · después de agregar se vuelve a la página con el criterio, no cumplido, y el progreso
func TestAgregarCriterio_VuelveALaPaginaConElCriterioNoCumplido(t *testing.T) {
	h := conHistoria(t)

	rec := post(h, "/proyectos/1/backlog/1/criterios", criterioValido("  Entra con su cuenta de Google "))

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("código = %d, se esperaba %d (redirección)", rec.Code, http.StatusSeeOther)
	}
	destino := rec.Header().Get("Location")
	if destino != "/proyectos/1/backlog/1?agregado=1" {
		t.Fatalf("redirige a %q, se esperaba /proyectos/1/backlog/1?agregado=1", destino)
	}
	verificarContiene(t, get(h, destino).Body.String(),
		"Criterio agregado", "Entra con su cuenta de Google", "No cumplido", "0/1",
		`action="/proyectos/1/backlog/1/criterios/1/cumplido"`, `name="cumplido" value="1"`, "Marcar cumplido",
	)
}

// RN1 · cada error aparece debajo de su campo, se conserva lo que se escribió y no se guarda nada
func TestAgregarCriterio_DescripcionInvalida_MuestraElErrorYConservaLoEscrito(t *testing.T) {
	h := conHistoria(t)
	larga := strings.Repeat("a", 301)
	casos := map[string]struct{ enviada, error string }{
		"vacía": {"   ", "la descripción del criterio es obligatoria"},
		"larga": {larga, "la descripción del criterio no puede superar los 300 caracteres"},
	}
	for nombre, c := range casos {
		rec := post(h, "/proyectos/1/backlog/1/criterios", criterioValido(c.enviada))

		if rec.Code != http.StatusUnprocessableEntity {
			t.Fatalf("%s: código = %d, se esperaba %d", nombre, rec.Code, http.StatusUnprocessableEntity)
		}
		verificarContiene(t, rec.Body.String(), c.error, "is-invalid", "Iniciar sesión con Google")
		if nombre == "larga" {
			verificarContiene(t, rec.Body.String(), larga) // lo escrito sigue en el campo
		}
	}
	verificarContiene(t, get(h, "/proyectos/1/backlog/1").Body.String(), "Sin criterios todavía")
}

// RN3 y CA-07.2 · marcar deja el criterio cumplido o no cumplido, con su mensaje y el progreso al día
func TestMarcarCriterio_CumplidoYNoCumplido_VuelveALaPaginaConElProgreso(t *testing.T) {
	h := conHistoria(t)
	post(h, "/proyectos/1/backlog/1/criterios", criterioValido("Entra con su cuenta"))
	post(h, "/proyectos/1/backlog/1/criterios", criterioValido("Muestra su nombre"))

	rec := post(h, "/proyectos/1/backlog/1/criterios/1/cumplido", marca("1"))

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("código = %d, se esperaba %d (redirección)", rec.Code, http.StatusSeeOther)
	}
	if destino := rec.Header().Get("Location"); destino != "/proyectos/1/backlog/1?marcado=cumplido" {
		t.Fatalf("redirige a %q, se esperaba /proyectos/1/backlog/1?marcado=cumplido", destino)
	}
	cumplido := get(h, rec.Header().Get("Location")).Body.String()
	verificarContiene(t, cumplido, "Criterio marcado como cumplido", "1/2", "Cumplido", "Marcar no cumplido")

	rec = post(h, "/proyectos/1/backlog/1/criterios/1/cumplido", marca("0"))

	if destino := rec.Header().Get("Location"); destino != "/proyectos/1/backlog/1?marcado=no-cumplido" {
		t.Fatalf("redirige a %q, se esperaba /proyectos/1/backlog/1?marcado=no-cumplido", destino)
	}
	verificarContiene(t, get(h, rec.Header().Get("Location")).Body.String(), "Criterio marcado como no cumplido", "0/2")
}

// Un valor de marca que no es 1 ni 0 es un pedido inválido y no cambia nada
func TestMarcarCriterio_ValorInvalido_Responde400YNoCambiaNada(t *testing.T) {
	h := conHistoria(t)
	post(h, "/proyectos/1/backlog/1/criterios", criterioValido("Entra con su cuenta"))

	for _, valor := range []string{"", "2", "true", "si"} {
		if rec := post(h, "/proyectos/1/backlog/1/criterios/1/cumplido", marca(valor)); rec.Code != http.StatusBadRequest {
			t.Errorf("valor %q: código = %d, se esperaba %d", valor, rec.Code, http.StatusBadRequest)
		}
	}
	verificarContiene(t, get(h, "/proyectos/1/backlog/1").Body.String(), "0/1")
}

// Los parámetros del mensaje de la página solo muestran los textos conocidos, nunca lo que venga en la URL
func TestPaginaDeLaHistoria_ElMensajeSoloMuestraTextosConocidos(t *testing.T) {
	html := get(conHistoria(t), "/proyectos/1/backlog/1?agregado=%3Cscript%3E&marcado=cualquier-cosa").Body.String()

	if strings.Contains(html, "<script>") || strings.Contains(html, "cualquier-cosa") || strings.Contains(html, "Criterio agregado") {
		t.Error("la página mostró un mensaje que no corresponde a un valor conocido")
	}
}

// La descripción es texto del usuario: se escapa, no se interpreta como HTML
func TestPaginaDeLaHistoria_EscapaLaDescripcionDelCriterio(t *testing.T) {
	h := conHistoria(t)
	post(h, "/proyectos/1/backlog/1/criterios", criterioValido("<script>alert(1)</script>"))

	html := get(h, "/proyectos/1/backlog/1").Body.String()

	if strings.Contains(html, "<script>alert(1)</script>") {
		t.Error("la descripción se mostró sin escapar")
	}
	verificarContiene(t, html, "&lt;script&gt;alert(1)&lt;/script&gt;")
}

// RN8 · cada historia muestra y cuenta solo sus criterios
func TestPaginaDeLaHistoria_CadaHistoriaMuestraSusCriterios(t *testing.T) {
	h := conHistoria(t)
	segunda := historiaValida()
	segunda.Set("titulo", "Ver el tablero")
	post(h, "/proyectos/1/backlog", segunda)
	post(h, "/proyectos/1/backlog/1/criterios", criterioValido("Criterio de la primera"))
	post(h, "/proyectos/1/backlog/2/criterios", criterioValido("Criterio de la segunda"))

	primera := get(h, "/proyectos/1/backlog/1").Body.String()

	verificarContiene(t, primera, "Criterio de la primera")
	if strings.Contains(primera, "Criterio de la segunda") {
		t.Error("la página de la HU-1 muestra un criterio de la HU-2")
	}
	// y un criterio de otra historia en la URL es un "no encontrado"
	if rec := post(h, "/proyectos/1/backlog/1/criterios/2/cumplido", marca("1")); rec.Code != http.StatusNotFound {
		t.Errorf("criterio de otra historia: código = %d, se esperaba %d", rec.Code, http.StatusNotFound)
	}
}

// El proyecto, la historia o el criterio que no existen (o un ID que no es un número) dan 404
func TestCriterios_ProyectoHistoriaOCriterioInexistente_Responde404(t *testing.T) {
	h := conHistoria(t)
	post(h, "/proyectos/1/backlog/1/criterios", criterioValido("Entra con su cuenta"))
	pedidos := map[string]*httptest.ResponseRecorder{
		"página de un proyecto que no existe":      get(h, "/proyectos/99/backlog/1"),
		"página de una historia que no existe":     get(h, "/proyectos/1/backlog/99"),
		"página con número que no es un número":    get(h, "/proyectos/1/backlog/abc"),
		"agregar en un proyecto que no existe":     post(h, "/proyectos/99/backlog/1/criterios", criterioValido("Uno")),
		"agregar en una historia que no existe":    post(h, "/proyectos/1/backlog/99/criterios", criterioValido("Uno")),
		"agregar con número que no es un número":   post(h, "/proyectos/1/backlog/abc/criterios", criterioValido("Uno")),
		"marcar en una historia que no existe":     post(h, "/proyectos/1/backlog/99/criterios/1/cumplido", marca("1")),
		"marcar un criterio que no existe":         post(h, "/proyectos/1/backlog/1/criterios/99/cumplido", marca("1")),
		"marcar con criterio que no es un número":  post(h, "/proyectos/1/backlog/1/criterios/abc/cumplido", marca("1")),
		"marcar en un proyecto que no existe":      post(h, "/proyectos/99/backlog/1/criterios/1/cumplido", marca("1")),
		"marcar con proyecto que no es un número":  post(h, "/proyectos/abc/backlog/1/criterios/1/cumplido", marca("1")),
		"página con proyecto que no es un número":  get(h, "/proyectos/abc/backlog/1"),
		"agregar con proyecto que no es un número": post(h, "/proyectos/abc/backlog/1/criterios", criterioValido("Uno")),
	}
	for nombre, rec := range pedidos {
		if rec.Code != http.StatusNotFound {
			t.Errorf("%s: código = %d, se esperaba %d", nombre, rec.Code, http.StatusNotFound)
		}
	}
}

// conHistoriaHecha arma la aplicación con una HU-1 ya Hecha y con un criterio cumplido, guardados directo en los
// repositorios: pasar una historia a Hecho es de HU-08.
func conHistoriaHecha(t *testing.T) http.Handler {
	t.Helper()
	ctx := context.Background()
	proyectos, historias, criterios := memory.NuevoProyectos(), memory.NuevoHistorias(), memory.NuevoCriterios()
	servicioProyectos := service.NuevoProyectos(proyectos, time.Now)
	p, err := servicioProyectos.Crear(ctx, project.Datos{
		Nombre:      "Software Metrics",
		FechaInicio: time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC),
		FechaFin:    time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("no se pudo crear el proyecto: %v", err)
	}
	hecha, err := historias.Guardar(ctx, backlog.Historia{
		ProyectoID: p.ID, Numero: 1, Titulo: "Registrar horas",
		Prioridad: backlog.PrioridadAlta, Estado: backlog.EstadoHecho,
	})
	if err != nil {
		t.Fatalf("no se pudo guardar la historia Hecha: %v", err)
	}
	if _, err := criterios.Guardar(ctx, backlog.Criterio{HistoriaID: hecha.ID, Descripcion: "Se guardan las horas", Cumplido: true}); err != nil {
		t.Fatalf("no se pudo guardar el criterio: %v", err)
	}
	return server.NewRouter(server.Dependencias{
		Proyectos: servicioProyectos,
		Historias: service.NuevoHistorias(proyectos, historias),
		Criterios: service.NuevoCriterios(proyectos, historias, criterios),
	})
}

// RN7 · una historia Hecha no admite cambios en sus criterios: la página no ofrece cambiarlos y el servidor tampoco
func TestHistoriaHecha_NoAdmiteCambiosEnSusCriterios(t *testing.T) {
	h := conHistoriaHecha(t)

	pagina := get(h, "/proyectos/1/backlog/1")
	verificarContiene(t, pagina.Body.String(), "Registrar horas", "Se guardan las horas", "1/1", "Hecho",
		"la historia está Hecha: no se pueden cambiar sus criterios")
	if strings.Contains(pagina.Body.String(), `action="/proyectos/1/backlog/1/criterios"`) ||
		strings.Contains(pagina.Body.String(), "Marcar no cumplido") {
		t.Error("la página de una historia Hecha no tiene que ofrecer agregar ni desmarcar criterios")
	}

	agregar := post(h, "/proyectos/1/backlog/1/criterios", criterioValido("Uno más"))
	if agregar.Code != http.StatusConflict {
		t.Errorf("agregar: código = %d, se esperaba %d", agregar.Code, http.StatusConflict)
	}
	verificarContiene(t, agregar.Body.String(), "la historia está Hecha: no se pueden cambiar sus criterios")

	desmarcar := post(h, "/proyectos/1/backlog/1/criterios/1/cumplido", marca("0"))
	if desmarcar.Code != http.StatusConflict {
		t.Errorf("desmarcar: código = %d, se esperaba %d", desmarcar.Code, http.StatusConflict)
	}
	verificarContiene(t, get(h, "/proyectos/1/backlog/1").Body.String(), "1/1")
}

// historiaQueExiste hace que la historia HU-N exista siempre, para probar los criterios con la base caída.
type historiaQueExiste struct{}

func (historiaQueExiste) BuscarPorNumero(_ context.Context, proyectoID int64, numero int) (backlog.Historia, error) {
	return backlog.Historia{ID: 4, ProyectoID: proyectoID, Numero: numero, Titulo: "Iniciar sesión", Estado: backlog.EstadoPendiente}, nil
}

// criteriosCaidos simula la base de datos caída para los criterios; el proyecto y la historia sí existen.
type criteriosCaidos struct{}

func (criteriosCaidos) ListarPorHistoria(context.Context, int64) ([]backlog.Criterio, error) {
	return nil, errBaseCaida
}
func (criteriosCaidos) ListarPorHistorias(context.Context, []int64) ([]backlog.Criterio, error) {
	return nil, errBaseCaida
}
func (criteriosCaidos) Guardar(_ context.Context, c backlog.Criterio) (backlog.Criterio, error) {
	return c, errBaseCaida
}
func (criteriosCaidos) BuscarPorID(_ context.Context, historiaID, id int64) (backlog.Criterio, error) {
	return backlog.Criterio{ID: id, HistoriaID: historiaID}, errBaseCaida
}
func (criteriosCaidos) Actualizar(context.Context, backlog.Criterio) error { return errBaseCaida }

// Si falla la base se muestra un mensaje general, sin detalles internos.
func TestCriterios_ErrorDeLaBase_MuestraUnMensajeGeneral(t *testing.T) {
	h := server.NewRouter(server.Dependencias{
		Proyectos: service.NuevoProyectos(repoCaido{}, nil),
		Criterios: service.NuevoCriterios(sprintsCaidos{}, historiaQueExiste{}, criteriosCaidos{}),
	})
	pedidos := map[string]*httptest.ResponseRecorder{
		"página":  get(h, "/proyectos/1/backlog/1"),
		"agregar": post(h, "/proyectos/1/backlog/1/criterios", criterioValido("Uno")),
		"marcar":  post(h, "/proyectos/1/backlog/1/criterios/1/cumplido", marca("1")),
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
