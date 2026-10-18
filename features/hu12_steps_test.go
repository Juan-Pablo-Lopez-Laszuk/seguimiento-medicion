package features

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// hu12 guarda el estado de un escenario de HU-12 (asignar historias a un sprint), con sus propios repositorios en
// memoria. Los sprints del Antecedentes se guardan recién antes del primer paso que los usa, así un escenario todavía
// puede cambiarlos (por ejemplo, "el proyecto no tiene sprints cerrados").
type hu12 struct {
	repoProyectos *memory.Proyectos
	repoSprints   *memory.Sprints
	repoHistorias *memory.Historias
	repoCriterios *memory.Criterios
	asignaciones  *service.Asignaciones
	proyectoID    int64
	pendientes    []sprint.Sprint
	asignacion    service.Asignacion
	err           error
}

// registrarPasosHU12 conecta cada frase de features/hu-12-asignar-historias.feature con una función de Go.
func registrarPasosHU12(sc *godog.ScenarioContext) {
	e := &hu12{
		repoProyectos: memory.NuevoProyectos(), repoSprints: memory.NuevoSprints(),
		repoHistorias: memory.NuevoHistorias(), repoCriterios: memory.NuevoCriterios(),
	}
	criterios := service.NuevoCriterios(e.repoProyectos, e.repoHistorias, e.repoCriterios)
	e.asignaciones = service.NuevoAsignaciones(e.repoProyectos, e.repoSprints, e.repoHistorias, criterios, func() time.Time {
		return time.Date(2026, 10, 15, 18, 0, 0, 0, time.UTC)
	})

	sc.Step(`^el proyecto "([^"]*)" para planificar sprints$`, e.elProyecto)
	sc.Step(`^que el Sprint (\d+) está cerrado con (\d+) SP completados$`, e.queElSprintEstaCerrado)
	sc.Step(`^que el Sprint (\d+) es el sprint activo$`, e.queElSprintEstaAbierto)
	sc.Step(`^que el Sprint (\d+) todavía no empezó$`, e.queElSprintEstaAbierto)
	sc.Step(`^el backlog para asignar tiene estas historias:$`, e.elBacklogTiene)
	sc.Step(`^que el proyecto no tiene sprints cerrados para comparar$`, e.sinSprintsCerrados)
	sc.Step(`^que el Sprint (\d+) ya tiene (\d+) SP comprometidos$`, e.queElSprintYaTiene)
	sc.Step(`^(?:que )?asign[oé] "([^"]*)" al Sprint (\d+)$`, e.asigno)
	sc.Step(`^quito "([^"]*)" del sprint$`, e.quito)
	sc.Step(`^"([^"]*)" queda en el Sprint (\d+) en estado "([^"]*)"$`, e.quedaEnElSprint)
	sc.Step(`^"([^"]*)" sigue sin sprint$`, e.sigueSinSprint)
	sc.Step(`^el Sprint (\d+) suma (\d+) SP comprometidos$`, e.elSprintSuma)
	sc.Step(`^la asignación da el error "([^"]*)"$`, e.laAsignacionDaElError)
	sc.Step(`^se avisa "([^"]*)"$`, e.seAvisa)
	sc.Step(`^no hay aviso de velocidad$`, func() error { return e.seAvisa("") })
}

func (e *hu12) elProyecto(nombre string) error {
	p, err := e.repoProyectos.Guardar(context.Background(), project.Proyecto{
		Nombre: nombre, FechaInicio: fechaDeEscenario("2026-10-05"), FechaFin: fechaDeEscenario("2026-11-01"),
	})
	e.proyectoID = p.ID
	return err
}

// nuevoSprint arma el Sprint N de una semana: el 1 empieza el 05/10, el 2 el 12/10 y así.
func (e *hu12) nuevoSprint(numero int) sprint.Sprint {
	inicio := fechaDeEscenario("2026-10-05").AddDate(0, 0, 7*(numero-1))
	return sprint.Sprint{
		ProyectoID: e.proyectoID, Numero: numero, Objetivo: fmt.Sprintf("Objetivo del Sprint %d", numero),
		FechaInicio: inicio, FechaFin: inicio.Add(6 * 24 * time.Hour), Estado: sprint.EstadoPlanificado,
	}
}

func (e *hu12) queElSprintEstaCerrado(numero, completados int) error {
	s := e.nuevoSprint(numero)
	s.Estado, s.SPCompletados = sprint.EstadoCerrado, &completados
	e.pendientes = append(e.pendientes, s)
	return nil
}

func (e *hu12) queElSprintEstaAbierto(numero int) error {
	e.pendientes = append(e.pendientes, e.nuevoSprint(numero))
	return nil
}

func (e *hu12) sinSprintsCerrados() error {
	for i := range e.pendientes {
		e.pendientes[i].Estado, e.pendientes[i].SPCompletados = sprint.EstadoPlanificado, nil
	}
	return nil
}

// guardarSprints guarda los sprints del Antecedentes, una sola vez.
func (e *hu12) guardarSprints() error {
	for _, s := range e.pendientes {
		if _, err := e.repoSprints.Guardar(context.Background(), s); err != nil {
			return fmt.Errorf("no se pudo guardar el Sprint %d: %w", s.Numero, err)
		}
	}
	e.pendientes = nil
	return nil
}

// elBacklogTiene guarda las historias de la tabla directo en los repositorios, con sus story points, sus criterios
// (ninguno cumplido) y su estado. La historia "HU-1 Login" queda con el número 1 y el título "Login".
func (e *hu12) elBacklogTiene(tabla *godog.Table) error {
	for _, fila := range tabla.Rows[1:] {
		ident, titulo, _ := strings.Cut(fila.Cells[0].Value, " ")
		numero, err := numeroDe(ident)
		if err != nil {
			return err
		}
		h := backlog.Historia{ProyectoID: e.proyectoID, Numero: numero, Titulo: titulo, Prioridad: backlog.PrioridadAlta,
			Estado: backlog.Estado(fila.Cells[3].Value)}
		if sp := fila.Cells[1].Value; sp != "sin estimar" {
			n, err := strconv.Atoi(sp)
			if err != nil {
				return fmt.Errorf("story points %q: %w", sp, err)
			}
			h.StoryPoints = &n
		}
		criterios, err := strconv.Atoi(fila.Cells[2].Value)
		if err != nil {
			return fmt.Errorf("criterios %q: %w", fila.Cells[2].Value, err)
		}
		if err := e.guardarHistoria(h, criterios); err != nil {
			return err
		}
	}
	return nil
}

func (e *hu12) guardarHistoria(h backlog.Historia, criterios int) error {
	guardada, err := e.repoHistorias.Guardar(context.Background(), h)
	if err != nil {
		return fmt.Errorf("no se pudo guardar %s: %w", h.Identificador(), err)
	}
	for i := 1; i <= criterios; i++ {
		c := backlog.Criterio{HistoriaID: guardada.ID, Descripcion: fmt.Sprintf("Criterio %d", i)}
		if _, err := e.repoCriterios.Guardar(context.Background(), c); err != nil {
			return fmt.Errorf("no se pudo guardar el criterio %d: %w", i, err)
		}
	}
	return nil
}

// queElSprintYaTiene asigna al sprint historias nuevas, listas, que suman esos SP: cada una con el valor más alto
// de la escala que entra en lo que falta (13 = una de 13; 40 = 21 + 13 + 5 + 1).
func (e *hu12) queElSprintYaTiene(numeroSprint, total int) error {
	if err := e.guardarSprints(); err != nil {
		return err
	}
	lista, err := e.repoHistorias.ListarPorProyecto(context.Background(), e.proyectoID)
	if err != nil {
		return err
	}
	numero := len(lista)
	for resto := total; resto > 0; {
		sp := mayorDeLaEscala(resto)
		numero++
		h := backlog.Historia{ProyectoID: e.proyectoID, Numero: numero, Titulo: "Ya comprometida", Prioridad: backlog.PrioridadMedia,
			Estado: backlog.EstadoPendiente, StoryPoints: &sp}
		if err := e.guardarHistoria(h, 1); err != nil {
			return err
		}
		if _, err := e.asignaciones.Asignar(context.Background(), e.proyectoID, numeroSprint, numero); err != nil {
			return fmt.Errorf("no se pudo comprometer la HU-%d: %w", numero, err)
		}
		resto -= sp
	}
	return nil
}

// mayorDeLaEscala devuelve el valor más alto de la escala de story points que no pasa de tope (tope > 0).
func mayorDeLaEscala(tope int) int {
	mayor := 1
	for _, v := range backlog.EscalaStoryPoints {
		if v <= tope {
			mayor = max(mayor, v)
		}
	}
	return mayor
}

func (e *hu12) asigno(ident string, numeroSprint int) error {
	if err := e.guardarSprints(); err != nil {
		return err
	}
	numero, err := numeroDe(ident)
	if err != nil {
		return err
	}
	e.asignacion, e.err = e.asignaciones.Asignar(context.Background(), e.proyectoID, numeroSprint, numero)
	return nil
}

// quito saca la historia del sprint en el que está.
func (e *hu12) quito(ident string) error {
	h, err := e.historia(ident)
	if err != nil {
		return err
	}
	if h.SprintID == nil {
		return fmt.Errorf("%s no está en ningún sprint", ident)
	}
	sprints, err := e.repoSprints.ListarPorProyecto(context.Background(), e.proyectoID)
	if err != nil {
		return err
	}
	for _, s := range sprints {
		if s.ID == *h.SprintID {
			_, e.err = e.asignaciones.Quitar(context.Background(), e.proyectoID, s.Numero, h.Numero)
			return nil
		}
	}
	return fmt.Errorf("no se encontró el sprint %d", *h.SprintID)
}

func (e *hu12) historia(ident string) (backlog.Historia, error) {
	numero, err := numeroDe(ident)
	if err != nil {
		return backlog.Historia{}, err
	}
	return e.repoHistorias.BuscarPorNumero(context.Background(), e.proyectoID, numero)
}

func (e *hu12) quedaEnElSprint(ident string, numeroSprint int, estado string) error {
	if e.err != nil {
		return fmt.Errorf("la asignación falló: %w", e.err)
	}
	h, err := e.historia(ident)
	if err != nil {
		return err
	}
	v, err := e.asignaciones.Ver(context.Background(), e.proyectoID, numeroSprint)
	if err != nil {
		return err
	}
	if h.SprintID == nil || *h.SprintID != v.Sprint.ID || string(h.Estado) != estado {
		return fmt.Errorf("%s quedó %q en el sprint %v, se esperaba %q en el Sprint %d", ident, h.Estado, h.SprintID, estado, numeroSprint)
	}
	return nil
}

func (e *hu12) sigueSinSprint(ident string) error {
	h, err := e.historia(ident)
	if err != nil {
		return err
	}
	if h.SprintID != nil {
		return fmt.Errorf("%s quedó en el sprint %d", ident, *h.SprintID)
	}
	return nil
}

func (e *hu12) elSprintSuma(numeroSprint, esperado int) error {
	v, err := e.asignaciones.Ver(context.Background(), e.proyectoID, numeroSprint)
	if err != nil {
		return err
	}
	if v.Comprometidos != esperado {
		return fmt.Errorf("el Sprint %d suma %d SP, se esperaban %d", numeroSprint, v.Comprometidos, esperado)
	}
	return nil
}

func (e *hu12) laAsignacionDaElError(mensaje string) error {
	if e.err == nil {
		return fmt.Errorf("se esperaba el error %q y la asignación valió", mensaje)
	}
	if e.err.Error() != mensaje {
		return fmt.Errorf("se esperaba el error %q y se obtuvo %q", mensaje, e.err.Error())
	}
	return nil
}

func (e *hu12) seAvisa(aviso string) error {
	if e.err != nil {
		return fmt.Errorf("la asignación falló: %w", e.err)
	}
	if e.asignacion.Aviso != aviso {
		return fmt.Errorf("aviso = %q, se esperaba %q", e.asignacion.Aviso, aviso)
	}
	return nil
}

// numeroDe convierte "HU-3" en 3.
func numeroDe(ident string) (int, error) {
	n, err := strconv.Atoi(strings.TrimPrefix(ident, "HU-"))
	if err != nil {
		return 0, fmt.Errorf("identificador de historia %q inválido", ident)
	}
	return n, nil
}
