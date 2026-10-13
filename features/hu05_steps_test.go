package features

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// hu05 guarda el estado de un escenario de HU-05 (crear ítem del Product Backlog), con sus propios repositorios
// en memoria. El primer proyecto que se crea es "el proyecto" de los pasos que no nombran ninguno.
type hu05 struct {
	proyectos  *service.Proyectos
	historias  *service.Historias
	idProyecto map[string]int64
	principal  int64
	creada     backlog.Historia
	err        error
}

// registrarPasosHU05 conecta cada frase de features/hu-05-crear-item-backlog.feature con una función de Go.
func registrarPasosHU05(sc *godog.ScenarioContext) {
	repoProyectos := memory.NuevoProyectos()
	e := &hu05{
		proyectos:  service.NuevoProyectos(repoProyectos, time.Now),
		historias:  service.NuevoHistorias(repoProyectos, memory.NuevoHistorias()),
		idProyecto: map[string]int64{},
	}

	sc.Step(`^el proyecto "([^"]*)" para el backlog$`, e.elProyectoParaElBacklog)
	sc.Step(`^que el backlog ya tiene la historia "([^"]*)" con prioridad "([^"]*)"$`, e.queElBacklogYaTieneLaHistoria)
	sc.Step(`^creo la historia "([^"]*)" con prioridad "([^"]*)" sin story points$`, e.creoLaHistoriaSinPuntos)
	sc.Step(`^creo la historia "([^"]*)" con prioridad "([^"]*)" y (\d+) story points$`, e.creoLaHistoriaConPuntos)
	sc.Step(`^creo la historia "([^"]*)" con prioridad "([^"]*)" y "([^"]*)" como story points$`, e.creoLaHistoriaConPuntosEscritos)
	sc.Step(`^creo en "([^"]*)" la historia "([^"]*)" con prioridad "([^"]*)" sin story points$`, e.creoEnLaHistoria)
	sc.Step(`^creo una historia con un título de (\d+) caracteres$`, e.creoUnaHistoriaConTituloDeLargo)
	sc.Step(`^se crea la historia HU-(\d+) "([^"]*)"$`, e.seCreaLaHistoria)
	sc.Step(`^la historia queda en estado "([^"]*)"$`, e.laHistoriaQuedaEnEstado)
	sc.Step(`^la historia queda sin estimar$`, e.laHistoriaQuedaSinEstimar)
	sc.Step(`^la historia queda con (\d+) story points$`, e.laHistoriaQuedaConPuntos)
	sc.Step(`^la historia da como resultado "([^"]*)"$`, e.laHistoriaDaComoResultado)
	sc.Step(`^el backlog no tiene historias$`, e.elBacklogNoTieneHistorias)
}

func (e *hu05) elProyectoParaElBacklog(nombre string) error {
	p, err := e.proyectos.Crear(context.Background(), project.Datos{
		Nombre:      nombre,
		FechaInicio: fechaDeEscenario("2026-10-05"),
		FechaFin:    fechaDeEscenario("2026-11-01"),
	})
	if err != nil {
		return fmt.Errorf("no se pudo crear el proyecto %q: %w", nombre, err)
	}
	e.idProyecto[nombre] = p.ID
	if e.principal == 0 {
		e.principal = p.ID
	}
	return nil
}

func (e *hu05) crear(proyectoID int64, d backlog.Datos) {
	e.creada, e.err = e.historias.Crear(context.Background(), proyectoID, d)
}

func (e *hu05) queElBacklogYaTieneLaHistoria(titulo, prioridad string) error {
	e.crear(e.principal, backlog.Datos{Titulo: titulo, Prioridad: backlog.Prioridad(prioridad)})
	return e.err
}

func (e *hu05) creoLaHistoriaSinPuntos(titulo, prioridad string) error {
	e.crear(e.principal, backlog.Datos{Titulo: titulo, Prioridad: backlog.Prioridad(prioridad)})
	return nil
}

func (e *hu05) creoLaHistoriaConPuntos(titulo, prioridad string, puntos int) error {
	return e.creoLaHistoriaConPuntosEscritos(titulo, prioridad, fmt.Sprint(puntos))
}

func (e *hu05) creoLaHistoriaConPuntosEscritos(titulo, prioridad, puntos string) error {
	e.crear(e.principal, backlog.Datos{Titulo: titulo, Prioridad: backlog.Prioridad(prioridad), StoryPoints: puntos})
	return nil
}

func (e *hu05) creoEnLaHistoria(proyecto, titulo, prioridad string) error {
	id, ok := e.idProyecto[proyecto]
	if !ok {
		return fmt.Errorf("el escenario no creó el proyecto %q", proyecto)
	}
	e.crear(id, backlog.Datos{Titulo: titulo, Prioridad: backlog.Prioridad(prioridad)})
	return nil
}

func (e *hu05) creoUnaHistoriaConTituloDeLargo(largo int) error {
	e.crear(e.principal, backlog.Datos{Titulo: strings.Repeat("a", largo), Prioridad: backlog.PrioridadMedia})
	return nil
}

func (e *hu05) seCreaLaHistoria(numero int, titulo string) error {
	if e.err != nil {
		return fmt.Errorf("no se creó la historia: %w", e.err)
	}
	if e.creada.Numero != numero || e.creada.Titulo != titulo {
		return fmt.Errorf("se esperaba la HU-%d %q y se obtuvo la %s %q",
			numero, titulo, e.creada.Identificador(), e.creada.Titulo)
	}
	return nil
}

func (e *hu05) laHistoriaQuedaEnEstado(estado string) error {
	if string(e.creada.Estado) != estado {
		return fmt.Errorf("estado: se esperaba %q y se obtuvo %q", estado, e.creada.Estado)
	}
	return nil
}

func (e *hu05) laHistoriaQuedaSinEstimar() error {
	if e.creada.StoryPoints != nil {
		return fmt.Errorf("se esperaba la historia sin estimar y tiene %d story points", *e.creada.StoryPoints)
	}
	return nil
}

func (e *hu05) laHistoriaQuedaConPuntos(puntos int) error {
	if e.creada.StoryPoints == nil || *e.creada.StoryPoints != puntos {
		return fmt.Errorf("se esperaban %d story points y se obtuvo %v", puntos, e.creada.StoryPoints)
	}
	return nil
}

// laHistoriaDaComoResultado acepta "creada" o "campo: mensaje de error".
func (e *hu05) laHistoriaDaComoResultado(resultado string) error {
	if resultado == "creada" {
		if e.err != nil {
			return fmt.Errorf("se esperaba que se cree y se obtuvo: %w", e.err)
		}
		return nil
	}
	campo, mensaje, ok := strings.Cut(resultado, ": ")
	if !ok {
		return fmt.Errorf("resultado mal escrito en el escenario: %q (se espera \"creada\" o \"campo: mensaje\")", resultado)
	}
	var errs domain.ErroresValidacion
	if !errors.As(e.err, &errs) {
		return fmt.Errorf("se esperaban errores de validación y se obtuvo: %v", e.err)
	}
	if errs[campo] == nil || errs[campo].Error() != mensaje {
		return fmt.Errorf("campo %q: se esperaba %q y se obtuvo %v", campo, mensaje, errs[campo])
	}
	return nil
}

func (e *hu05) elBacklogNoTieneHistorias() error {
	_, lista, err := e.historias.Listar(context.Background(), e.principal)
	if err != nil {
		return err
	}
	if len(lista) != 0 {
		return fmt.Errorf("no se esperaba ninguna historia y hay %d", len(lista))
	}
	return nil
}
