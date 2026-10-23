package server_test

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

// CA-10.1 · el backlog tiene el formulario de filtros, que se manda por GET
func TestBacklog_TieneElFormularioDeFiltros(t *testing.T) {
	h, _ := backlogParaPriorizar(t)

	html := get(h, "/proyectos/1/backlog").Body.String()

	verificarContiene(t, html,
		`action="/proyectos/1/backlog" method="get"`,
		`name="estado"`, `name="prioridad"`, `name="texto"`,
		`<option value="">Todos</option>`, `<option value="">Todas</option>`, `value="En progreso"`,
	)
	// Sin filtros no hay cantidad, ni Quitar filtros, ni aviso, y se puede reordenar.
	verificarQueNoContiene(t, html, "de 4 historias", "Quitar filtros", "Para cambiar el orden")
	verificarContiene(t, html, `action="/proyectos/1/backlog/4/subir"`)
}

// CA-10.1 y RN5 · con un filtro se ven las que cumplen, la cantidad, el filtro elegido y sin Subir ni Bajar
func TestBacklog_FiltrarPorPrioridad(t *testing.T) {
	h, _ := backlogParaPriorizar(t)

	rec := get(h, "/proyectos/1/backlog?prioridad=Alta")

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	html := rec.Body.String()
	verificarEnOrden(t, html, ">Registro<", ">Reportes<")
	verificarQueNoContiene(t, html, ">Perfil<", ">Login<", `/subir"`, `/bajar"`)
	verificarContiene(t, html,
		"2 de 4 historias", `<option value="Alta" selected>`,
		"Para cambiar el orden, quitá los filtros.", `href="/proyectos/1/backlog">Quitar filtros`,
	)
}

// CA-10.1 y RN2 · la búsqueda por texto, que vuelve a aparecer en el campo
func TestBacklog_BuscarPorTexto(t *testing.T) {
	h, _ := backlogParaPriorizar(t)

	html := get(h, "/proyectos/1/backlog?texto=PERF").Body.String()

	verificarContiene(t, html, ">Perfil<", `value="PERF"`, "1 de 4 historias")
	verificarQueNoContiene(t, html, ">Registro<", ">Reportes<", ">Login<")
}

// RN3 · filtrar por Hecho deja solo la sección Hechas, sin el aviso de que no queda nada pendiente
func TestBacklog_FiltrarLasHechas(t *testing.T) {
	h, repo := backlogParaPriorizar(t)
	completarHistoria(t, repo, 1, time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC))

	html := get(h, "/proyectos/1/backlog?estado=Hecho").Body.String()

	verificarContiene(t, html, "Hechas", ">Login<", "1 de 4 historias")
	verificarQueNoContiene(t, html, ">Registro<", "No queda ninguna historia pendiente.", "Ninguna historia coincide")
}

// CA-10.2 · sin resultados, un mensaje claro con el enlace para quitar los filtros
func TestBacklog_FiltrosSinResultados(t *testing.T) {
	h, _ := backlogParaPriorizar(t)

	html := get(h, "/proyectos/1/backlog?estado=Hecho&prioridad=Alta").Body.String()

	verificarContiene(t, html, "Ninguna historia coincide con los filtros.", `href="/proyectos/1/backlog">Quitar filtros`)
	verificarQueNoContiene(t, html, "Todavía no hay historias", "No queda ninguna historia pendiente.", "Hechas</h2>")
}

// Un backlog vacío dice lo de siempre, aunque haya filtros
func TestBacklog_VacioConFiltros(t *testing.T) {
	html := get(conProyecto(t), "/proyectos/1/backlog?prioridad=Alta").Body.String()

	verificarContiene(t, html, "Todavía no hay historias")
	verificarQueNoContiene(t, html, "Ninguna historia coincide")
}

// RN4 · un valor que no existe se ignora: se ve todo y se puede reordenar
func TestBacklog_FiltrosQueNoExistenSeIgnoran(t *testing.T) {
	h, _ := backlogParaPriorizar(t)

	html := get(h, "/proyectos/1/backlog?estado=Archivada&prioridad=Urgente").Body.String()

	verificarEnOrden(t, html, ">Registro<", ">Reportes<", ">Perfil<", ">Login<")
	verificarContiene(t, html, `action="/proyectos/1/backlog/4/subir"`)
	verificarQueNoContiene(t, html, "de 4 historias", "Quitar filtros")
}

// El texto buscado se muestra escapado
func TestBacklog_ElTextoBuscadoSeMuestraEscapado(t *testing.T) {
	h, _ := backlogParaPriorizar(t)

	html := get(h, "/proyectos/1/backlog?texto=%3Cscript%3Ealert(1)%3C%2Fscript%3E").Body.String()

	if strings.Contains(html, "<script>alert(1)") {
		t.Error("la página mostró el texto buscado sin escapar")
	}
	verificarContiene(t, html, "Ninguna historia coincide con los filtros.")
}
