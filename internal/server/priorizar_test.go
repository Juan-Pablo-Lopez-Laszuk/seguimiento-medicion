package server_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// backlogParaPriorizar arma la aplicación con el proyecto 1 y las historias de la feature de HU-09 (HU-1 Login Baja,
// HU-2 Registro Alta, HU-3 Perfil Media y HU-4 Reportes Alta), y devuelve también el repositorio de historias.
func backlogParaPriorizar(t *testing.T) (http.Handler, *memory.Historias) {
	t.Helper()
	ctx := context.Background()
	proyectos, historias := memory.NuevoProyectos(), memory.NuevoHistorias()
	if _, err := proyectos.Guardar(ctx, project.Proyecto{Nombre: "Software Metrics",
		FechaInicio: time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC), FechaFin: time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)}); err != nil {
		t.Fatalf("guardar el proyecto: %v", err)
	}
	casos := service.NuevoHistorias(proyectos, historias)
	for _, d := range []backlog.Datos{
		{Titulo: "Login", Prioridad: backlog.PrioridadBaja},
		{Titulo: "Registro", Prioridad: backlog.PrioridadAlta},
		{Titulo: "Perfil", Prioridad: backlog.PrioridadMedia},
		{Titulo: "Reportes", Prioridad: backlog.PrioridadAlta},
	} {
		if _, err := casos.Crear(ctx, 1, d); err != nil {
			t.Fatalf("crear %q: %v", d.Titulo, err)
		}
	}
	h := server.NewRouter(server.Dependencias{
		Proyectos: service.NuevoProyectos(proyectos, time.Now),
		Historias: casos,
		Criterios: service.NuevoCriterios(proyectos, historias, memory.NuevoCriterios()),
	})
	return h, historias
}

// completarHistoria deja la HU-numero Hecha y completada ese día, directo en el repositorio.
func completarHistoria(t *testing.T, repo *memory.Historias, numero int, fecha time.Time) {
	t.Helper()
	h, err := repo.BuscarPorNumero(context.Background(), 1, numero)
	if err != nil {
		t.Fatalf("buscar la HU-%d: %v", numero, err)
	}
	if err := repo.ActualizarCompletada(context.Background(), h.ID, backlog.EstadoHecho, &fecha); err != nil {
		t.Fatalf("completar la HU-%d: %v", numero, err)
	}
}

// verificarEnOrden comprueba que los textos aparezcan en la página en ese orden.
func verificarEnOrden(t *testing.T, html string, textos ...string) {
	t.Helper()
	desde := 0
	for _, texto := range textos {
		i := strings.Index(html[desde:], texto)
		if i < 0 {
			t.Errorf("no aparece %q después de %q (orden esperado: %v)", texto, html[max(desde-40, 0):desde], textos)
			return
		}
		desde += i + len(texto)
	}
}

// verificarQueNoContiene comprueba que la página no tenga ninguno de los textos.
func verificarQueNoContiene(t *testing.T, html string, textos ...string) {
	t.Helper()
	for _, texto := range textos {
		if strings.Contains(html, texto) {
			t.Errorf("la página no debería contener %q", texto)
		}
	}
}

// CA-09.1 · el backlog se muestra por prioridad; la primera de cada prioridad no tiene Subir y la última no tiene Bajar
func TestBacklog_OrdenadoPorPrioridadConSubirYBajar(t *testing.T) {
	h, _ := backlogParaPriorizar(t)

	rec := get(h, "/proyectos/1/backlog")

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	html := rec.Body.String()
	verificarEnOrden(t, html, ">Registro<", ">Reportes<", ">Perfil<", ">Login<")
	verificarContiene(t, html,
		`action="/proyectos/1/backlog/2/bajar"`, `action="/proyectos/1/backlog/4/subir"`, // las dos de Alta
	)
	verificarQueNoContiene(t, html,
		`action="/proyectos/1/backlog/2/subir"`, `action="/proyectos/1/backlog/4/bajar"`,
		`action="/proyectos/1/backlog/3/subir"`, `action="/proyectos/1/backlog/3/bajar"`, // única de Media
		`action="/proyectos/1/backlog/1/subir"`, `action="/proyectos/1/backlog/1/bajar"`, // única de Baja
		"Hechas", // sin Hechas no hay sección
	)
}

// CA-09.1 y CA-09.2 · subir vuelve al backlog con el aviso y el orden nuevo
func TestBacklog_SubirVuelveConElAvisoYElOrdenNuevo(t *testing.T) {
	h, _ := backlogParaPriorizar(t)

	rec := post(h, "/proyectos/1/backlog/4/subir", nil)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("código = %d, se esperaba %d (redirección)", rec.Code, http.StatusSeeOther)
	}
	destino := rec.Header().Get("Location")
	if destino != "/proyectos/1/backlog?subio=4" {
		t.Fatalf("redirige a %q, se esperaba /proyectos/1/backlog?subio=4", destino)
	}
	html := get(h, destino).Body.String()
	verificarContiene(t, html, "HU-4 subió un lugar.")
	verificarEnOrden(t, html, ">Reportes<", ">Registro<", ">Perfil<", ">Login<")
}

func TestBacklog_BajarVuelveConElAviso(t *testing.T) {
	h, _ := backlogParaPriorizar(t)

	rec := post(h, "/proyectos/1/backlog/2/bajar", nil)

	if destino := rec.Header().Get("Location"); rec.Code != http.StatusSeeOther || destino != "/proyectos/1/backlog?bajo=2" {
		t.Fatalf("código = %d y destino %q, se esperaba 303 a /proyectos/1/backlog?bajo=2", rec.Code, destino)
	}
	html := get(h, "/proyectos/1/backlog?bajo=2").Body.String()
	verificarContiene(t, html, "HU-2 bajó un lugar.")
	verificarEnOrden(t, html, ">Reportes<", ">Registro<")
}

// RN3 y RN4 · si no se puede mover, se muestra el backlog sin cambios con el motivo (409)
func TestBacklog_MovimientoQueNoSePuede(t *testing.T) {
	casos := []struct {
		nombre, ruta, motivo string
	}{
		{"la primera no sube", "/proyectos/1/backlog/2/subir", "HU-2 ya es la primera de su prioridad"},
		{"la última no baja", "/proyectos/1/backlog/4/bajar", "HU-4 ya es la última de su prioridad"},
		{"una Hecha no se reordena", "/proyectos/1/backlog/1/subir", "una historia Hecha no se reordena"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			h, repo := backlogParaPriorizar(t)
			completarHistoria(t, repo, 1, time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC))

			rec := post(h, c.ruta, nil)

			if rec.Code != http.StatusConflict {
				t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusConflict)
			}
			html := rec.Body.String()
			verificarContiene(t, html, c.motivo, "Backlog de Software Metrics")
			verificarEnOrden(t, html, ">Registro<", ">Reportes<", ">Perfil<")
		})
	}
}

// CA-09.3 · las Hechas van en su sección, la más reciente primero y sin botones
func TestBacklog_LasHechasVanAparte(t *testing.T) {
	h, repo := backlogParaPriorizar(t)
	completarHistoria(t, repo, 1, time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC))
	completarHistoria(t, repo, 3, time.Date(2026, 10, 14, 0, 0, 0, 0, time.UTC))

	html := get(h, "/proyectos/1/backlog").Body.String()

	verificarEnOrden(t, html, ">Registro<", ">Reportes<", "Hechas", ">Perfil<", "14/10/2026", ">Login<", "12/10/2026")
	verificarQueNoContiene(t, html, `action="/proyectos/1/backlog/1/`, `action="/proyectos/1/backlog/3/`)
}

// Todas Hechas: la sección de pendientes avisa que no queda nada
func TestBacklog_TodasHechas(t *testing.T) {
	h, repo := backlogParaPriorizar(t)
	for numero := 1; numero <= 4; numero++ {
		completarHistoria(t, repo, numero, time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC))
	}

	verificarContiene(t, get(h, "/proyectos/1/backlog").Body.String(), "No queda ninguna historia pendiente.", "Hechas")
}

// Sección 7 · historia o proyecto que no existen (404), y un movimiento que no es subir ni bajar no tiene ruta
func TestBacklog_MoverAlgoQueNoExiste(t *testing.T) {
	h, _ := backlogParaPriorizar(t)

	for _, ruta := range []string{"/proyectos/1/backlog/99/subir", "/proyectos/9/backlog/1/bajar", "/proyectos/1/backlog/2/arriba"} {
		if rec := post(h, ruta, nil); rec.Code != http.StatusNotFound && rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: código = %d, se esperaba 404", ruta, rec.Code)
		}
	}
	verificarContiene(t, post(h, "/proyectos/1/backlog/99/subir", nil).Body.String(), "no existe la historia HU-99")
}

// historiasSinOrden es el repositorio en memoria con una base que falla al guardar el orden.
type historiasSinOrden struct{ *memory.Historias }

func (historiasSinOrden) ActualizarOrden(context.Context, int64, map[int64]int) error {
	return errBaseCaida
}

// Sección 7 · si falla la base al guardar el orden se muestra un mensaje general (500)
func TestBacklog_FallaAlGuardarElOrden(t *testing.T) {
	_, repo := backlogParaPriorizar(t)
	h := server.NewRouter(server.Dependencias{Historias: service.NuevoHistorias(historiasCaidas{}, historiasSinOrden{repo})})

	rec := post(h, "/proyectos/1/backlog/4/subir", nil)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusInternalServerError)
	}
	verificarContiene(t, rec.Body.String(), "No se pudo guardar el orden. Probá de nuevo en unos minutos.")
}
