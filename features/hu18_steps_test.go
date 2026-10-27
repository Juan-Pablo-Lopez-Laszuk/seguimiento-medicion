package features

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strings"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// hu18 guarda el estado de un escenario de HU-18 (votar de forma oculta), con sus propios repositorios en memoria. La
// sesión se abre, se vota y los integrantes se dan de baja con los casos de uso, como en la aplicación; quién votó se
// pide también por HTTP, como lo pide HTMX.
type hu18 struct {
	repoProyectos   *memory.Proyectos
	repoHistorias   *memory.Historias
	repoIntegrantes *memory.Integrantes
	integrantes     *service.Integrantes
	poker           *service.Poker
	router          http.Handler
	proyectoID      int64
	porNombre       map[string]member.Integrante
	porTitulo       map[string]backlog.Historia
	// valores son los valores que se votaron, para buscarlos donde no tienen que estar (RN5).
	valores []string
	err     error
}

// registrarPasosHU18 conecta cada frase de features/hu-18-votar-oculto.feature con una función de Go.
func registrarPasosHU18(sc *godog.ScenarioContext) {
	repoPoker := memory.NuevoPoker()
	e := &hu18{
		repoProyectos: memory.NuevoProyectos(), repoHistorias: memory.NuevoHistorias(), repoIntegrantes: memory.NuevoIntegrantes(),
		porNombre: map[string]member.Integrante{}, porTitulo: map[string]backlog.Historia{},
	}
	e.integrantes = service.NuevoIntegrantes(e.repoProyectos, e.repoIntegrantes)
	e.poker = service.NuevoPoker(e.repoProyectos, e.repoHistorias, e.repoIntegrantes, repoPoker).ConVotos(repoPoker)
	e.router = server.NewRouter(server.Dependencias{Poker: e.poker})

	sc.Step(`^el proyecto "([^"]*)" con una sesión de Planning Poker abierta para "([^"]*)" y estos integrantes:$`, e.elProyectoConUnaSesionAbierta)
	sc.Step(`^"([^"]*)" vota "([^"]*)"$`, e.vota)
	sc.Step(`^que "([^"]*)" votó "([^"]*)"$`, e.voto)
	sc.Step(`^"([^"]*)" vota "([^"]*)" en la historia "([^"]*)", que no tiene sesión$`, e.votaEnLaHistoria)
	sc.Step(`^en la ronda votaron (\d+) de (\d+)$`, e.enLaRondaVotaron)
	sc.Step(`^votaron estos integrantes:$`, e.votaronEstosIntegrantes)
	sc.Step(`^la lista de quién votó no muestra ningún valor$`, e.laListaNoMuestraValores)
	sc.Step(`^la vista de la sesión no tiene ningún valor de voto$`, e.laVistaNoTieneValores)
	sc.Step(`^votar da el error "([^"]*)"$`, e.votarDaElError)
	sc.Step(`^"([^"]*)" queda dado de baja$`, e.quedaDadoDeBaja)
}

// elProyectoConUnaSesionAbierta crea el proyecto con los integrantes, la historia de la sesión (Pendiente) y "Perfil"
// (En progreso, sin sesión), y abre la sesión con el Agile Enabler.
func (e *hu18) elProyectoConUnaSesionAbierta(nombre, titulo string, tabla *godog.Table) error {
	ctx := context.Background()
	p, err := e.repoProyectos.Guardar(ctx, project.Proyecto{
		Nombre: nombre, FechaInicio: fechaDeEscenario("2026-10-05"), FechaFin: fechaDeEscenario("2026-11-01"),
	})
	if err != nil {
		return err
	}
	e.proyectoID = p.ID
	var agileEnabler int64
	for _, fila := range tabla.Rows[1:] {
		nombre, rol, activo := fila.Cells[0].Value, fila.Cells[1].Value, fila.Cells[2].Value
		i, err := e.repoIntegrantes.Guardar(ctx, member.Integrante{
			ProyectoID: p.ID, Nombre: nombre, Email: fmt.Sprintf("integrante%d@mail.com", len(e.porNombre)+1),
			Rol: member.Rol(rol), Activo: activo == "sí",
		})
		if err != nil {
			return err
		}
		e.porNombre[nombre] = i
		if i.Rol == member.RolAgileEnabler && i.Activo {
			agileEnabler = i.ID
		}
	}
	for n, h := range []backlog.Historia{{Titulo: titulo, Estado: backlog.EstadoPendiente}, {Titulo: "Perfil", Estado: backlog.EstadoEnProgreso}} {
		h.ProyectoID, h.Numero, h.Prioridad = p.ID, n+1, backlog.PrioridadAlta
		guardada, err := e.repoHistorias.Guardar(ctx, h)
		if err != nil {
			return err
		}
		e.porTitulo[guardada.Titulo] = guardada
	}
	_, err = e.poker.Abrir(ctx, p.ID, 1, agileEnabler)
	return err
}

// vota guarda el error para el paso que lo revisa. Vota en la historia de la sesión (HU-1).
func (e *hu18) vota(quien, valor string) error {
	return e.votaEnLaHistoria(quien, valor, e.tituloDeLaSesion())
}

// voto es el Dado: el voto tiene que guardarse.
func (e *hu18) voto(quien, valor string) error {
	if err := e.vota(quien, valor); err != nil {
		return err
	}
	return e.err
}

func (e *hu18) votaEnLaHistoria(quien, valor, titulo string) error {
	h, ok := e.porTitulo[titulo]
	if !ok {
		return fmt.Errorf("no hay una historia %q en el escenario", titulo)
	}
	_, e.err = e.poker.Votar(context.Background(), e.proyectoID, h.Numero, e.porNombre[quien].ID, valor)
	if e.err == nil {
		e.valores = append(e.valores, valor)
	}
	return nil
}

func (e *hu18) tituloDeLaSesion() string {
	for titulo, h := range e.porTitulo {
		if h.Numero == 1 {
			return titulo
		}
	}
	return ""
}

func (e *hu18) vista() (service.VistaSesion, error) {
	return e.poker.Ver(context.Background(), e.proyectoID, 1)
}

// enLaRondaVotaron cuenta los votos de la vista. Después de un error, comprueba que no se guardó nada.
func (e *hu18) enLaRondaVotaron(votaron, participantes int) error {
	v, err := e.vista()
	if err != nil {
		return err
	}
	if len(v.Votaron) != votaron || len(v.Participantes) != participantes {
		return fmt.Errorf("votaron %d de %d, se esperaba %d de %d", len(v.Votaron), len(v.Participantes), votaron, participantes)
	}
	return nil
}

func (e *hu18) votaronEstosIntegrantes(tabla *godog.Table) error {
	v, err := e.vista()
	if err != nil {
		return err
	}
	var quiero, obtenidos []string
	for _, fila := range tabla.Rows[1:] {
		quiero = append(quiero, fila.Cells[0].Value)
	}
	for _, p := range v.Participantes {
		if v.Votaron[p.ID] {
			obtenidos = append(obtenidos, p.Nombre)
		}
	}
	if !slices.Equal(obtenidos, quiero) {
		return fmt.Errorf("votaron %v, se esperaba %v", obtenidos, quiero)
	}
	return nil
}

// laListaNoMuestraValores pide quién votó como lo pide HTMX (GET …/poker/votos) y busca los valores votados (CA-18.3).
func (e *hu18) laListaNoMuestraValores() error {
	ruta := fmt.Sprintf("/proyectos/%d/backlog/1/poker/votos", e.proyectoID)
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, ruta, nil))
	if rec.Code != http.StatusOK {
		return fmt.Errorf("GET %s: código %d", ruta, rec.Code)
	}
	cuerpo := strings.ReplaceAll(rec.Body.String(), ruta, "")
	if !strings.Contains(cuerpo, "votaron") {
		return fmt.Errorf("la lista de quién votó no está en la respuesta:\n%s", cuerpo)
	}
	for _, valor := range e.valores {
		if strings.Contains(cuerpo, valor) {
			return fmt.Errorf("la lista de quién votó muestra el valor %s:\n%s", valor, cuerpo)
		}
	}
	return nil
}

// laVistaNoTieneValores revisa que la vista del caso de uso no tenga los valores (CA-18.2, RN5): Votaron solo dice
// qué participantes votaron, y ningún campo de la vista es un poker.Voto.
func (e *hu18) laVistaNoTieneValores() error {
	v, err := e.vista()
	if err != nil {
		return err
	}
	if len(v.Votaron) == 0 {
		return errors.New("se esperaba que alguien haya votado")
	}
	for id, voto := range v.Votaron {
		if !voto || !slices.ContainsFunc(v.Participantes, func(p member.Integrante) bool { return p.ID == id }) {
			return fmt.Errorf("Votaron tiene algo que no es un participante que votó: %v", v.Votaron)
		}
	}
	return tieneVotos(reflect.TypeOf(v), "VistaSesion")
}

// tieneVotos devuelve un error si el tipo tiene, en algún nivel, un poker.Voto.
func tieneVotos(tipo reflect.Type, camino string) error {
	switch tipo.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Array:
		return tieneVotos(tipo.Elem(), camino+"[]")
	case reflect.Map:
		if err := tieneVotos(tipo.Key(), camino+"[clave]"); err != nil {
			return err
		}
		return tieneVotos(tipo.Elem(), camino+"[]")
	case reflect.Struct:
		if tipo == reflect.TypeFor[poker.Voto]() {
			return fmt.Errorf("%s es un poker.Voto", camino)
		}
		for i := range tipo.NumField() {
			if err := tieneVotos(tipo.Field(i).Type, camino+"."+tipo.Field(i).Name); err != nil {
				return err
			}
		}
	default:
	}
	return nil
}

func (e *hu18) votarDaElError(mensaje string) error {
	if e.err == nil {
		return errors.New("se esperaba un error al votar y no lo hubo")
	}
	if e.err.Error() != mensaje {
		return fmt.Errorf("error = %q, se esperaba %q", e.err.Error(), mensaje)
	}
	return nil
}

func (e *hu18) quedaDadoDeBaja(nombre string) error {
	_, err := e.integrantes.DarDeBaja(context.Background(), e.proyectoID, e.porNombre[nombre].ID)
	return err
}
