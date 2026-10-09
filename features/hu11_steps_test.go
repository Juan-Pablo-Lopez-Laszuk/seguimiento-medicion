package features

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// hu11 guarda el estado de un escenario de HU-11 (crear sprint), con sus propios repositorios en memoria.
// El primer proyecto que se crea en el escenario es "el proyecto" de los pasos que no nombran ninguno.
type hu11 struct {
	proyectos  *service.Proyectos
	sprints    *service.Sprints
	idProyecto map[string]int64
	principal  int64
	creado     sprint.Sprint
	err        error
}

// registrarPasosHU11 conecta cada frase de features/hu-11-crear-sprint.feature con una función de Go.
func registrarPasosHU11(sc *godog.ScenarioContext) {
	repoProyectos := memory.NuevoProyectos()
	e := &hu11{
		proyectos:  service.NuevoProyectos(repoProyectos, time.Now),
		sprints:    service.NuevoSprints(repoProyectos, memory.NuevoSprints()),
		idProyecto: map[string]int64{},
	}

	sc.Step(`^el proyecto "([^"]*)" que va del "([^"]*)" al "([^"]*)"$`, e.elProyectoQueVa)
	sc.Step(`^que el proyecto ya tiene un sprint del "([^"]*)" al "([^"]*)"$`, e.queElProyectoYaTieneUnSprint)
	sc.Step(`^creo un sprint con el goal "([^"]*)" del "([^"]*)" al "([^"]*)"$`, e.creoUnSprint)
	sc.Step(`^creo en "([^"]*)" un sprint con el goal "([^"]*)" del "([^"]*)" al "([^"]*)"$`, e.creoUnSprintEn)
	sc.Step(`^se crea el Sprint (\d+) con el goal "([^"]*)"$`, e.seCreaElSprint)
	sc.Step(`^el sprint queda en estado "([^"]*)"$`, e.elSprintQuedaEnEstado)
	sc.Step(`^el sprint da como resultado "([^"]*)"$`, e.elSprintDaComoResultado)
	sc.Step(`^el proyecto no tiene sprints$`, e.elProyectoNoTieneSprints)
}

func (e *hu11) elProyectoQueVa(nombre, inicio, fin string) error {
	p, err := e.proyectos.Crear(context.Background(), project.Datos{
		Nombre:      nombre,
		FechaInicio: fechaDeEscenario(inicio),
		FechaFin:    fechaDeEscenario(fin),
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

func (e *hu11) crear(proyectoID int64, goal, inicio, fin string) {
	e.creado, e.err = e.sprints.Crear(context.Background(), proyectoID, sprint.Datos{
		Objetivo:    goal,
		FechaInicio: fechaDeEscenario(inicio),
		FechaFin:    fechaDeEscenario(fin),
	})
}

func (e *hu11) queElProyectoYaTieneUnSprint(inicio, fin string) error {
	e.crear(e.principal, "Sprint anterior", inicio, fin)
	return e.err
}

func (e *hu11) creoUnSprint(goal, inicio, fin string) error {
	e.crear(e.principal, goal, inicio, fin)
	return nil
}

func (e *hu11) creoUnSprintEn(proyecto, goal, inicio, fin string) error {
	id, ok := e.idProyecto[proyecto]
	if !ok {
		return fmt.Errorf("el escenario no creó el proyecto %q", proyecto)
	}
	e.crear(id, goal, inicio, fin)
	return nil
}

func (e *hu11) seCreaElSprint(numero int, goal string) error {
	if e.err != nil {
		return fmt.Errorf("no se creó el sprint: %w", e.err)
	}
	if e.creado.Numero != numero || e.creado.Objetivo != goal {
		return fmt.Errorf("se esperaba el Sprint %d %q y se obtuvo el Sprint %d %q",
			numero, goal, e.creado.Numero, e.creado.Objetivo)
	}
	return nil
}

func (e *hu11) elSprintQuedaEnEstado(estado string) error {
	if string(e.creado.Estado) != estado {
		return fmt.Errorf("estado: se esperaba %q y se obtuvo %q", estado, e.creado.Estado)
	}
	return nil
}

// elSprintDaComoResultado acepta "creado" o "campo: mensaje de error".
func (e *hu11) elSprintDaComoResultado(resultado string) error {
	if resultado == "creado" {
		if e.err != nil {
			return fmt.Errorf("se esperaba que se cree y se obtuvo: %w", e.err)
		}
		return nil
	}
	campo, mensaje, ok := strings.Cut(resultado, ": ")
	if !ok {
		return fmt.Errorf("resultado mal escrito en el escenario: %q (se espera \"creado\" o \"campo: mensaje\")", resultado)
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

func (e *hu11) elProyectoNoTieneSprints() error {
	_, lista, err := e.sprints.Listar(context.Background(), e.principal)
	if err != nil {
		return err
	}
	if len(lista) != 0 {
		return fmt.Errorf("no se esperaba ningún sprint y hay %d", len(lista))
	}
	return nil
}
