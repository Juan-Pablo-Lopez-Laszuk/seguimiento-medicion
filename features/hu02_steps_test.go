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

// hu02 guarda el estado de un escenario de HU-02 (modificar proyecto), con sus propios repositorios en memoria.
type hu02 struct {
	proyectos  *memory.Proyectos
	sprints    *memory.Sprints
	proyectoID int64
	hoy        time.Time
	err        error
}

// registrarPasosHU02 conecta cada frase de features/hu-02-modificar-proyecto.feature con una función de Go.
func registrarPasosHU02(sc *godog.ScenarioContext) {
	e := &hu02{proyectos: memory.NuevoProyectos(), sprints: memory.NuevoSprints()}

	sc.Step(`^el proyecto a modificar "([^"]*)" del "([^"]*)" al "([^"]*)"$`, e.elProyectoAModificar)
	sc.Step(`^que ese proyecto tiene el Sprint (\d+) del "([^"]*)" al "([^"]*)" y el Sprint (\d+) del "([^"]*)" al "([^"]*)"$`, e.queTieneDosSprints)
	sc.Step(`^que también existe el proyecto "([^"]*)"$`, e.queTambienExiste)
	sc.Step(`^que hoy es el "([^"]*)" para modificar$`, e.queHoyEsElParaModificar)
	sc.Step(`^modifico el proyecto con nombre "([^"]*)", descripción "([^"]*)" del "([^"]*)" al "([^"]*)"$`, e.modificoElProyecto)
	sc.Step(`^el proyecto queda con la descripción "([^"]*)"$`, e.elProyectoTieneLaDescripcion)
	sc.Step(`^el proyecto sigue con la descripción "([^"]*)"$`, e.elProyectoTieneLaDescripcion)
	sc.Step(`^el proyecto queda con el nombre "([^"]*)"$`, e.elProyectoTieneElNombre)
	sc.Step(`^el proyecto sigue con el nombre "([^"]*)"$`, e.elProyectoTieneElNombre)
	sc.Step(`^la modificación da como resultado "([^"]*)"$`, e.laModificacionDaComoResultado)
	sc.Step(`^la modificación da el error "([^"]*)"$`, e.laModificacionDaElError)
}

func (e *hu02) crearProyecto(nombre, inicio, fin string) (project.Proyecto, error) {
	return service.NuevoProyectos(e.proyectos, time.Now).Crear(context.Background(), project.Datos{
		Nombre:      nombre,
		FechaInicio: fechaDeEscenario(inicio),
		FechaFin:    fechaDeEscenario(fin),
	})
}

func (e *hu02) elProyectoAModificar(nombre, inicio, fin string) error {
	p, err := e.crearProyecto(nombre, inicio, fin)
	if err != nil {
		return fmt.Errorf("no se pudo crear el proyecto %q: %w", nombre, err)
	}
	e.proyectoID = p.ID
	return nil
}

func (e *hu02) queTieneDosSprints(n1 int, inicio1, fin1 string, n2 int, inicio2, fin2 string) error {
	casos := service.NuevoSprints(e.proyectos, e.sprints)
	for _, s := range []sprint.Datos{
		{Objetivo: fmt.Sprintf("Sprint %d", n1), FechaInicio: fechaDeEscenario(inicio1), FechaFin: fechaDeEscenario(fin1)},
		{Objetivo: fmt.Sprintf("Sprint %d", n2), FechaInicio: fechaDeEscenario(inicio2), FechaFin: fechaDeEscenario(fin2)},
	} {
		if _, err := casos.Crear(context.Background(), e.proyectoID, s); err != nil {
			return fmt.Errorf("no se pudo crear el sprint: %w", err)
		}
	}
	return nil
}

func (e *hu02) queTambienExiste(nombre string) error {
	_, err := e.crearProyecto(nombre, "2026-10-05", "2026-11-01")
	return err
}

func (e *hu02) queHoyEsElParaModificar(fecha string) error {
	e.hoy = fechaDeEscenario(fecha)
	return nil
}

func (e *hu02) modificoElProyecto(nombre, descripcion, inicio, fin string) error {
	ahora := func() time.Time { return e.hoy.Add(18 * time.Hour) } // 15:00 de Argentina
	_, e.err = service.NuevoEdicionProyectos(e.proyectos, e.sprints, ahora).Modificar(context.Background(), e.proyectoID, project.Datos{
		Nombre:      nombre,
		Descripcion: descripcion,
		FechaInicio: fechaDeEscenario(inicio),
		FechaFin:    fechaDeEscenario(fin),
	})
	return nil
}

func (e *hu02) guardado() (project.Proyecto, error) {
	return e.proyectos.BuscarPorID(context.Background(), e.proyectoID)
}

func (e *hu02) elProyectoTieneLaDescripcion(esperada string) error {
	p, err := e.guardado()
	if err != nil {
		return err
	}
	if p.Descripcion != esperada {
		return fmt.Errorf("se esperaba la descripción %q y se obtuvo %q (error de la modificación: %v)", esperada, p.Descripcion, e.err)
	}
	return nil
}

func (e *hu02) elProyectoTieneElNombre(esperado string) error {
	p, err := e.guardado()
	if err != nil {
		return err
	}
	if p.Nombre != esperado {
		return fmt.Errorf("se esperaba el nombre %q y se obtuvo %q (error de la modificación: %v)", esperado, p.Nombre, e.err)
	}
	return nil
}

// laModificacionDaComoResultado acepta "modificado" o "campo: mensaje" con el error esperado en ese campo.
func (e *hu02) laModificacionDaComoResultado(esperado string) error {
	if esperado == "modificado" {
		if e.err != nil {
			return fmt.Errorf("se esperaba que se modifique y dio error: %v", e.err)
		}
		return nil
	}
	campo, mensaje, _ := strings.Cut(esperado, ": ")
	var errs domain.ErroresValidacion
	if !errors.As(e.err, &errs) {
		return fmt.Errorf("se esperaba un error en %q y se obtuvo %v", campo, e.err)
	}
	if errs[campo] == nil || errs[campo].Error() != mensaje {
		return fmt.Errorf("campo %q: se esperaba %q y se obtuvo %v", campo, mensaje, errs[campo])
	}
	return nil
}

func (e *hu02) laModificacionDaElError(mensaje string) error {
	if e.err == nil || e.err.Error() != mensaje {
		return fmt.Errorf("se esperaba el error %q y se obtuvo %v", mensaje, e.err)
	}
	return nil
}
