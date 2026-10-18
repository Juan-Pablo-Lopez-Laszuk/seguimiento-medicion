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

// hu13 guarda el estado de un escenario de HU-13 (registrar historias completadas), con sus propios repositorios en
// memoria y un reloj que fija cada escenario.
type hu13 struct {
	repoProyectos *memory.Proyectos
	repoSprints   *memory.Sprints
	repoHistorias *memory.Historias
	repoCriterios *memory.Criterios
	asignaciones  *service.Asignaciones
	proyectoID    int64
	sprintIDs     map[int]int64 // ID de cada sprint, por número
	ahora         time.Time
	err           error
}

// zonaArgentina es la hora de Argentina (UTC−3), para armar "ahora" en los escenarios.
var zonaArgentina = time.FixedZone("ART", -3*60*60)

// registrarPasosHU13 conecta cada frase de features/hu-13-completar-historias.feature con una función de Go.
func registrarPasosHU13(sc *godog.ScenarioContext) {
	e := &hu13{
		repoProyectos: memory.NuevoProyectos(), repoSprints: memory.NuevoSprints(),
		repoHistorias: memory.NuevoHistorias(), repoCriterios: memory.NuevoCriterios(), sprintIDs: map[int]int64{},
	}
	criterios := service.NuevoCriterios(e.repoProyectos, e.repoHistorias, e.repoCriterios)
	e.asignaciones = service.NuevoAsignaciones(e.repoProyectos, e.repoSprints, e.repoHistorias, criterios, func() time.Time { return e.ahora })

	sc.Step(`^el proyecto "([^"]*)" con el Sprint (\d+) activo del "([^"]*)" al "([^"]*)"$`, e.elProyectoConElSprintActivo)
	sc.Step(`^que el Sprint (\d+) empieza el "([^"]*)"$`, e.queElSprintEmpieza)
	sc.Step(`^el Sprint (\d+) tiene estas historias para completar:$`, e.elSprintTieneEstasHistorias)
	sc.Step(`^que hoy es el "([^"]*)" para completar$`, e.queHoyEs)
	sc.Step(`^que ahora son las "([^"]*)" del "([^"]*)" en Argentina$`, e.queAhoraSonLas)
	sc.Step(`^que "([^"]*)" está En progreso en el Sprint (\d+) con sus criterios cumplidos$`, e.queEstaEnProgresoEnElSprint)
	sc.Step(`^que el Sprint (\d+) todavía no se cerró$`, e.queElSprintNoSeCerro)
	sc.Step(`^completo "([^"]*)"$`, e.completo)
	sc.Step(`^"([^"]*)" queda Hecha con fecha de completado "([^"]*)"$`, e.quedaHecha)
	sc.Step(`^el Sprint (\d+) lleva (\d+) de (\d+) SP completados$`, e.elSprintLleva)
	sc.Step(`^completar da el error "([^"]*)"$`, e.completarDaElError)
	sc.Step(`^"([^"]*)" sigue "([^"]*)"$`, e.sigue)
}

// elProyectoConElSprintActivo guarda el proyecto, el Sprint 1 ya cerrado (la semana anterior) y el sprint activo.
func (e *hu13) elProyectoConElSprintActivo(nombre string, numero int, inicio, fin string) error {
	ctx := context.Background()
	p, err := e.repoProyectos.Guardar(ctx, project.Proyecto{
		Nombre: nombre, FechaInicio: fechaDeEscenario("2026-10-05"), FechaFin: fechaDeEscenario("2026-11-01"),
	})
	if err != nil {
		return err
	}
	e.proyectoID = p.ID
	desde := fechaDeEscenario(inicio)
	for n := 1; n < numero; n++ {
		anterior := desde.AddDate(0, 0, -7*(numero-n))
		if err := e.guardarSprint(n, anterior, anterior.AddDate(0, 0, 6), sprint.EstadoCerrado); err != nil {
			return err
		}
	}
	return e.guardarSprint(numero, desde, fechaDeEscenario(fin), sprint.EstadoPlanificado)
}

func (e *hu13) queElSprintEmpieza(numero int, inicio string) error {
	desde := fechaDeEscenario(inicio)
	return e.guardarSprint(numero, desde, desde.AddDate(0, 0, 6), sprint.EstadoPlanificado)
}

func (e *hu13) guardarSprint(numero int, inicio, fin time.Time, estado sprint.Estado) error {
	s, err := e.repoSprints.Guardar(context.Background(), sprint.Sprint{
		ProyectoID: e.proyectoID, Numero: numero, Objetivo: fmt.Sprintf("Objetivo del Sprint %d", numero),
		FechaInicio: inicio, FechaFin: fin, Estado: estado,
	})
	if err != nil {
		return fmt.Errorf("no se pudo guardar el Sprint %d: %w", numero, err)
	}
	e.sprintIDs[numero] = s.ID
	return nil
}

// elSprintTieneEstasHistorias guarda las historias de la tabla ya asignadas al sprint, con sus story points, su
// estado y sus criterios ("1 de 2" son dos criterios, el primero cumplido).
func (e *hu13) elSprintTieneEstasHistorias(numeroSprint int, tabla *godog.Table) error {
	for _, fila := range tabla.Rows[1:] {
		sp, err := strconv.Atoi(fila.Cells[1].Value)
		if err != nil {
			return fmt.Errorf("story points %q: %w", fila.Cells[1].Value, err)
		}
		var cumplidos, total int
		if _, err := fmt.Sscanf(fila.Cells[2].Value, "%d de %d", &cumplidos, &total); err != nil {
			return fmt.Errorf("criterios %q: %w", fila.Cells[2].Value, err)
		}
		if err := e.guardarHistoria(fila.Cells[0].Value, sp, numeroSprint, backlog.Estado(fila.Cells[3].Value), total, cumplidos); err != nil {
			return err
		}
	}
	return nil
}

// guardarHistoria guarda "HU-N Título" en el sprint, en ese estado y con esos criterios, directo en los repositorios.
func (e *hu13) guardarHistoria(identYTitulo string, sp, numeroSprint int, estado backlog.Estado, total, cumplidos int) error {
	ctx := context.Background()
	ident, titulo, _ := strings.Cut(identYTitulo, " ")
	numero, err := numeroDe(ident)
	if err != nil {
		return err
	}
	sprintID := e.sprintIDs[numeroSprint]
	h, err := e.repoHistorias.Guardar(ctx, backlog.Historia{ProyectoID: e.proyectoID, Numero: numero, Titulo: titulo,
		Prioridad: backlog.PrioridadAlta, Estado: estado, StoryPoints: &sp, SprintID: &sprintID})
	if err != nil {
		return fmt.Errorf("no se pudo guardar %s: %w", ident, err)
	}
	for i := 1; i <= total; i++ {
		c := backlog.Criterio{HistoriaID: h.ID, Descripcion: fmt.Sprintf("Criterio %d", i), Cumplido: i <= cumplidos}
		if _, err := e.repoCriterios.Guardar(ctx, c); err != nil {
			return fmt.Errorf("no se pudo guardar el criterio %d: %w", i, err)
		}
	}
	return nil
}

// queHoyEs fija "ahora" al mediodía de Argentina de ese día.
func (e *hu13) queHoyEs(fecha string) error {
	return e.queAhoraSonLas("12:00", fecha)
}

func (e *hu13) queAhoraSonLas(hora, fecha string) error {
	ahora, err := time.ParseInLocation("2006-01-02 15:04", fecha+" "+hora, zonaArgentina)
	if err != nil {
		return fmt.Errorf("fecha y hora %q %q: %w", fecha, hora, err)
	}
	e.ahora = ahora.UTC() // el servidor corre en UTC (Vercel)
	return nil
}

func (e *hu13) queEstaEnProgresoEnElSprint(identYTitulo string, numeroSprint int) error {
	return e.guardarHistoria(identYTitulo, 3, numeroSprint, backlog.EstadoEnProgreso, 1, 1)
}

func (e *hu13) queElSprintNoSeCerro(numero int) error {
	sprints, err := e.repoSprints.ListarPorProyecto(context.Background(), e.proyectoID)
	if err != nil {
		return err
	}
	for _, s := range sprints {
		if s.Numero == numero && s.Estado == sprint.EstadoCerrado {
			return fmt.Errorf("el Sprint %d está cerrado", numero)
		}
	}
	return nil
}

// completo completa la historia desde la página del sprint en el que está.
func (e *hu13) completo(ident string) error {
	h, err := e.historia(ident)
	if err != nil {
		return err
	}
	for numero, id := range e.sprintIDs {
		if h.SprintID != nil && *h.SprintID == id {
			_, e.err = e.asignaciones.Completar(context.Background(), e.proyectoID, numero, h.Numero)
			return nil
		}
	}
	return fmt.Errorf("%s no está en ningún sprint", ident)
}

func (e *hu13) historia(ident string) (backlog.Historia, error) {
	numero, err := numeroDe(ident)
	if err != nil {
		return backlog.Historia{}, err
	}
	return e.repoHistorias.BuscarPorNumero(context.Background(), e.proyectoID, numero)
}

func (e *hu13) quedaHecha(ident, fecha string) error {
	if e.err != nil {
		return fmt.Errorf("completar falló: %w", e.err)
	}
	h, err := e.historia(ident)
	if err != nil {
		return err
	}
	if h.Estado != backlog.EstadoHecho || h.CompletadaEn == nil || h.CompletadaEn.Format(time.DateOnly) != fecha {
		return fmt.Errorf("%s quedó %q con fecha %v, se esperaba Hecho el %s", ident, h.Estado, h.CompletadaEn, fecha)
	}
	return nil
}

func (e *hu13) elSprintLleva(numeroSprint, completados, comprometidos int) error {
	v, err := e.asignaciones.Ver(context.Background(), e.proyectoID, numeroSprint)
	if err != nil {
		return err
	}
	if v.Completados != completados || v.Comprometidos != comprometidos {
		return fmt.Errorf("el Sprint %d lleva %d de %d SP, se esperaba %d de %d", numeroSprint, v.Completados, v.Comprometidos, completados, comprometidos)
	}
	return nil
}

func (e *hu13) completarDaElError(mensaje string) error {
	if e.err == nil {
		return fmt.Errorf("se esperaba el error %q y se completó", mensaje)
	}
	if e.err.Error() != mensaje {
		return fmt.Errorf("se esperaba el error %q y se obtuvo %q", mensaje, e.err.Error())
	}
	return nil
}

func (e *hu13) sigue(ident, estado string) error {
	h, err := e.historia(ident)
	if err != nil {
		return err
	}
	if string(h.Estado) != estado || h.CompletadaEn != nil {
		return fmt.Errorf("%s quedó %q, se esperaba que siga %q", ident, h.Estado, estado)
	}
	return nil
}
