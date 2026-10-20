package server_test

import (
	"net/http"
	"strings"
	"testing"
)

// Los tests de HU-15 cierran el Sprint 1 (12/10 al 18/10) con la HU-1 completada (5 SP) y la HU-3 sin completar
// (3 SP): queda con 5 de 8 SP. El Sprint 2 (19/10 al 25/10) sigue abierto.
func conSprintCerrado(t *testing.T) http.Handler {
	t.Helper()
	h := conHistoriaCompletada(t)
	if rec := post(h, "/proyectos/1/backlog/3/criterios", criterioValido("Cambia su foto")); rec.Code != http.StatusSeeOther {
		t.Fatalf("no se pudo agregar el criterio a la HU-3: código %d", rec.Code)
	}
	if rec := asignar(h, "3"); rec.Code != http.StatusSeeOther {
		t.Fatalf("no se pudo asignar la HU-3: código %d", rec.Code)
	}
	conSprint2(t, h)
	if rec := post(h, rutaDeCierre, nil); rec.Code != http.StatusSeeOther {
		t.Fatalf("no se pudo cerrar el Sprint 1: código %d", rec.Code)
	}
	return h
}

// CA-15.1 · la lista de sprints muestra los SP completados y el cumplimiento de los cerrados, y — en los abiertos
func TestListaDeSprints_MuestraElResultadoDeLosCerrados(t *testing.T) {
	pagina := get(conSprintCerrado(t), "/proyectos/1/sprints").Body.String()

	verificarContiene(t, pagina, "SP completados", "Cumplimiento", "5 de 8", "62,5 %")
	uno, dos := strings.Index(pagina, "Sprint 1</a>"), strings.Index(pagina, "Sprint 2</a>")
	if uno < 0 || dos < uno {
		t.Fatal("la lista tiene que mostrar el Sprint 1 antes que el Sprint 2")
	}
	if fila := pagina[dos:]; strings.Count(fila[:strings.Index(fila, "</tr>")], "—") != 2 {
		t.Error("el Sprint 2 está abierto: tiene que mostrar — en SP completados y en cumplimiento")
	}
}

// CA-15.2 · la página de un sprint cerrado muestra el resultado y la foto del cierre, con enlace a cada historia
func TestPaginaDeSprintCerrado_MuestraLaFoto(t *testing.T) {
	pagina := get(conSprintCerrado(t), "/proyectos/1/sprints/1").Body.String()

	verificarContiene(t, pagina,
		"5 de 8 SP completados · 62,5 % de cumplimiento · 1 de 2 historias",
		"Historias al cerrar",
		`href="/proyectos/1/backlog/1"`, "Iniciar sesión con Google", "Completada",
		`href="/proyectos/1/backlog/3"`, "Editar el perfil", "No completada")
}

// Sección 3 · la página de un sprint abierto no cambia: no tiene la foto
func TestPaginaDeSprintAbierto_NoTieneLaFoto(t *testing.T) {
	pagina := get(conSprintCerrado(t), "/proyectos/1/sprints/2").Body.String()

	if strings.Contains(pagina, "Historias al cerrar") || strings.Contains(pagina, "de cumplimiento") {
		t.Error("un sprint abierto no tiene que mostrar la foto del cierre")
	}
}

// RN3 · un sprint cerrado sin historias muestra 0 de 0, cumplimiento — y que se cerró sin historias
func TestSprintCerradoSinHistorias(t *testing.T) {
	h := conSprint(t, "2026-10-12", "2026-10-18")
	if rec := post(h, rutaDeCierre, nil); rec.Code != http.StatusSeeOther {
		t.Fatalf("no se pudo cerrar el Sprint 1: código %d", rec.Code)
	}

	verificarContiene(t, get(h, "/proyectos/1/sprints").Body.String(), "0 de 0")
	verificarContiene(t, get(h, "/proyectos/1/sprints/1").Body.String(), "El sprint se cerró sin historias")
}
