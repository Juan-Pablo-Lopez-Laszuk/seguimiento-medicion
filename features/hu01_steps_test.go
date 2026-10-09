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
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// hu01 guarda el estado de un escenario de HU-01 (crear proyecto). Se crea uno nuevo por escenario,
// con su propio repositorio en memoria, así un escenario no le deja proyectos al siguiente.
type hu01 struct {
	repo     *memory.Proyectos
	servicio *service.Proyectos
	creado   project.Proyecto
	err      error
}

// registrarPasosHU01 conecta cada frase de features/hu-01-crear-proyecto.feature con una función de Go.
func registrarPasosHU01(sc *godog.ScenarioContext) {
	e := &hu01{repo: memory.NuevoProyectos()}
	e.servicio = service.NuevoProyectos(e.repo, time.Now)

	sc.Step(`^que existe el proyecto "([^"]*)"$`, e.queExisteElProyecto)
	sc.Step(`^creo el proyecto "([^"]*)" con descripción "([^"]*)" del "([^"]*)" al "([^"]*)"$`, e.creoElProyecto)
	sc.Step(`^creo el proyecto "([^"]*)" sin descripción del "([^"]*)" al "([^"]*)"$`, e.creoElProyectoSinDescripcion)
	sc.Step(`^creo un proyecto con un nombre de (\d+) caracteres del "([^"]*)" al "([^"]*)"$`, e.creoUnProyectoConNombreDeLargo)
	sc.Step(`^el proyecto "([^"]*)" queda creado$`, e.elProyectoQuedaCreado)
	sc.Step(`^su estado es "([^"]*)"$`, e.suEstadoEs)
	sc.Step(`^el resultado es "([^"]*)"$`, e.elResultadoEs)
	sc.Step(`^se informa el error "([^"]*)" en el campo "([^"]*)"$`, e.seInformaElErrorEnElCampo)
	sc.Step(`^hay (\d+) proyecto con el nombre "([^"]*)"$`, e.hayProyectosConElNombre)
	sc.Step(`^no se crea ningún proyecto$`, e.noSeCreaNingunProyecto)
}

// fechaDeEscenario convierte "AAAA-MM-DD" en fecha; vacío o inválido queda como fecha cero, igual que
// un campo vacío del formulario.
func fechaDeEscenario(texto string) time.Time {
	f, err := time.Parse(time.DateOnly, texto)
	if err != nil {
		return time.Time{}
	}
	return f
}

func (e *hu01) crear(nombre, descripcion, inicio, fin string) {
	e.creado, e.err = e.servicio.Crear(context.Background(), project.Datos{
		Nombre:      nombre,
		Descripcion: descripcion,
		FechaInicio: fechaDeEscenario(inicio),
		FechaFin:    fechaDeEscenario(fin),
	})
}

func (e *hu01) queExisteElProyecto(nombre string) error {
	e.crear(nombre, "", "2026-01-01", "2026-12-31")
	return e.err
}

func (e *hu01) creoElProyecto(nombre, descripcion, inicio, fin string) error {
	e.crear(nombre, descripcion, inicio, fin)
	return nil
}

func (e *hu01) creoElProyectoSinDescripcion(nombre, inicio, fin string) error {
	e.crear(nombre, "", inicio, fin)
	return nil
}

func (e *hu01) creoUnProyectoConNombreDeLargo(largo int, inicio, fin string) error {
	e.crear(strings.Repeat("a", largo), "", inicio, fin)
	return nil
}

func (e *hu01) elProyectoQuedaCreado(nombre string) error {
	if e.err != nil {
		return fmt.Errorf("no se creó el proyecto: %w", e.err)
	}
	if e.creado.Nombre != nombre {
		return fmt.Errorf("nombre: se esperaba %q y se obtuvo %q", nombre, e.creado.Nombre)
	}
	return e.hayProyectosConElNombre(1, nombre)
}

func (e *hu01) suEstadoEs(estado string) error {
	if string(e.creado.Estado) != estado {
		return fmt.Errorf("estado: se esperaba %q y se obtuvo %q", estado, e.creado.Estado)
	}
	return nil
}

// elResultadoEs acepta "creado" o "campo: mensaje de error".
func (e *hu01) elResultadoEs(resultado string) error {
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
	return e.seInformaElErrorEnElCampo(mensaje, campo)
}

func (e *hu01) seInformaElErrorEnElCampo(mensaje, campo string) error {
	var errs domain.ErroresValidacion
	if !errors.As(e.err, &errs) {
		return fmt.Errorf("se esperaban errores de validación y se obtuvo: %v", e.err)
	}
	if errs[campo] == nil || errs[campo].Error() != mensaje {
		return fmt.Errorf("campo %q: se esperaba %q y se obtuvo %v", campo, mensaje, errs[campo])
	}
	return nil
}

func (e *hu01) hayProyectosConElNombre(cantidad int, nombre string) error {
	lista, err := e.repo.Listar(context.Background())
	if err != nil {
		return err
	}
	n := 0
	for _, p := range lista {
		if strings.EqualFold(p.Nombre, nombre) {
			n++
		}
	}
	if n != cantidad {
		return fmt.Errorf("se esperaban %d proyectos llamados %q y hay %d", cantidad, nombre, n)
	}
	return nil
}

func (e *hu01) noSeCreaNingunProyecto() error {
	lista, err := e.repo.Listar(context.Background())
	if err != nil {
		return err
	}
	if len(lista) != 0 {
		return fmt.Errorf("no se esperaba ningún proyecto y hay %d", len(lista))
	}
	return nil
}
