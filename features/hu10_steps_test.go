package features

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// hu10 guarda el estado de un escenario de HU-10 (filtrar y buscar historias). Los filtros se prueban de punta a punta:
// cada paso pide la página del backlog a la aplicación armada con repositorios en memoria y lee el HTML.
type hu10 struct {
	repoProyectos *memory.Proyectos
	repoHistorias *memory.Historias
	historias     *service.Historias
	app           http.Handler
	proyectoID    int64
	pagina        string
}

// registrarPasosHU10 conecta cada frase de features/hu-10-filtrar-buscar-historias.feature con una función de Go.
func registrarPasosHU10(sc *godog.ScenarioContext) {
	e := &hu10{repoProyectos: memory.NuevoProyectos(), repoHistorias: memory.NuevoHistorias()}
	e.historias = service.NuevoHistorias(e.repoProyectos, e.repoHistorias)
	e.app = server.NewRouter(server.Dependencias{
		Proyectos: service.NuevoProyectos(e.repoProyectos, time.Now),
		Historias: e.historias,
		Criterios: service.NuevoCriterios(e.repoProyectos, e.repoHistorias, memory.NuevoCriterios()),
	})

	sc.Step(`^el proyecto "([^"]*)" con estas historias para filtrar:$`, e.elProyecto)
	sc.Step(`^filtro el backlog sin filtros$`, func() error { return e.filtrar(url.Values{}) })
	sc.Step(`^filtro el backlog por (estado|prioridad) "([^"]*)"$`, func(filtro, valor string) error {
		return e.filtrar(url.Values{filtro: {valor}})
	})
	sc.Step(`^busco "([^"]*)" en el backlog$`, func(texto string) error { return e.filtrar(url.Values{"texto": {texto}}) })
	sc.Step(`^filtro el backlog por estado "([^"]*)", prioridad "([^"]*)" y texto "([^"]*)"$`, func(estado, prioridad, texto string) error {
		return e.filtrar(url.Values{"estado": {estado}, "prioridad": {prioridad}, "texto": {texto}})
	})
	sc.Step(`^veo estas historias:$`, e.veoLaTabla)
	sc.Step(`^veo estas historias: "([^"]*)"$`, e.veoLaLista)
	sc.Step(`^no veo ninguna historia$`, func() error { return e.veo(nil) })
	sc.Step(`^veo (\d+) historias$`, e.veoLaCantidad)
	sc.Step(`^se informa que se ven (\d+) de (\d+) historias$`, func(n, total int) error {
		return e.seInforma(fmt.Sprintf("%d de %d historias", n, total))
	})
	sc.Step(`^se informa "([^"]*)"$`, e.seInforma)
	sc.Step(`^no se ofrece reordenar$`, e.noSeOfreceReordenar)
}

func (e *hu10) elProyecto(nombre string, tabla *godog.Table) error {
	ctx := context.Background()
	p, err := e.repoProyectos.Guardar(ctx, project.Proyecto{
		Nombre: nombre, FechaInicio: fechaDeEscenario("2026-10-05"), FechaFin: fechaDeEscenario("2026-11-01"),
	})
	if err != nil {
		return err
	}
	e.proyectoID = p.ID
	for _, fila := range tabla.Rows[1:] {
		titulo, prioridad, estado := fila.Cells[0].Value, fila.Cells[1].Value, backlog.Estado(fila.Cells[2].Value)
		h, err := e.historias.Crear(ctx, p.ID, backlog.Datos{Titulo: titulo, Prioridad: backlog.Prioridad(prioridad)})
		if err != nil {
			return fmt.Errorf("crear %q: %w", titulo, err)
		}
		switch estado {
		case backlog.EstadoPendiente:
		case backlog.EstadoHecho:
			dia := fechaDeEscenario("2026-10-14")
			err = e.repoHistorias.ActualizarCompletada(ctx, h.ID, estado, &dia)
		default:
			err = e.repoHistorias.ActualizarEstado(ctx, h.ID, estado)
		}
		if err != nil {
			return fmt.Errorf("dejar %q en %s: %w", titulo, estado, err)
		}
	}
	return nil
}

// filtrar pide la página del backlog con esos filtros en la URL, como el formulario.
func (e *hu10) filtrar(filtros url.Values) error {
	rec := httptest.NewRecorder()
	ruta := fmt.Sprintf("/proyectos/%d/backlog?%s", e.proyectoID, filtros.Encode())
	e.app.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, ruta, nil))
	if rec.Code != http.StatusOK {
		return fmt.Errorf("GET %s: código %d", ruta, rec.Code)
	}
	e.pagina = rec.Body.String()
	return nil
}

// tituloEnLaPagina es el enlace con el título de cada historia, en las dos secciones del backlog.
var tituloEnLaPagina = regexp.MustCompile(`class="text-reset text-decoration-none" href="/proyectos/\d+/backlog/\d+">([^<]+)</a>`)

// titulos devuelve los títulos de las historias que muestra la página, en orden.
func (e *hu10) titulos() []string {
	var titulos []string
	for _, m := range tituloEnLaPagina.FindAllStringSubmatch(e.pagina, -1) {
		titulos = append(titulos, html.UnescapeString(m[1]))
	}
	return titulos
}

func (e *hu10) veo(esperados []string) error {
	if got := e.titulos(); !slices.Equal(got, esperados) {
		return fmt.Errorf("se esperaba ver %q y se ven %q", esperados, got)
	}
	return nil
}

func (e *hu10) veoLaTabla(tabla *godog.Table) error {
	var esperados []string
	for _, fila := range tabla.Rows[1:] {
		esperados = append(esperados, fila.Cells[0].Value)
	}
	return e.veo(esperados)
}

func (e *hu10) veoLaLista(lista string) error {
	var esperados []string
	for _, titulo := range strings.Split(lista, ",") {
		esperados = append(esperados, strings.TrimSpace(titulo))
	}
	return e.veo(esperados)
}

func (e *hu10) veoLaCantidad(n int) error {
	if got := len(e.titulos()); got != n {
		return fmt.Errorf("se esperaban %d historias y se ven %d: %q", n, got, e.titulos())
	}
	return nil
}

func (e *hu10) seInforma(texto string) error {
	if !strings.Contains(e.pagina, html.EscapeString(texto)) && !strings.Contains(e.pagina, texto) {
		return fmt.Errorf("la página no dice %q", texto)
	}
	return nil
}

func (e *hu10) noSeOfreceReordenar() error {
	if strings.Contains(e.pagina, `/subir"`) || strings.Contains(e.pagina, `/bajar"`) {
		return fmt.Errorf("la página ofrece Subir o Bajar con filtros puestos")
	}
	return nil
}
