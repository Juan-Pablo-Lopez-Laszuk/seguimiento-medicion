package features

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// hu24 guarda el estado de un escenario de HU-24 (registrar esfuerzo), con sus propios repositorios en memoria.
type hu24 struct {
	repoProyectos   *memory.Proyectos
	repoHistorias   *memory.Historias
	repoIntegrantes *memory.Integrantes
	integrantes     *service.Integrantes
	esfuerzo        *service.Esfuerzo
	hoy             time.Time
	proyectoID      int64
	historia        backlog.Historia
	porNombre       map[string]member.Integrante
	err             error
}

// registrarPasosHU24 conecta cada frase de features/hu-24-registro-esfuerzo.feature con una función de Go.
func registrarPasosHU24(sc *godog.ScenarioContext) {
	e := &hu24{
		repoProyectos: memory.NuevoProyectos(), repoHistorias: memory.NuevoHistorias(),
		repoIntegrantes: memory.NuevoIntegrantes(), porNombre: map[string]member.Integrante{},
	}
	e.integrantes = service.NuevoIntegrantes(e.repoProyectos, e.repoIntegrantes)
	e.esfuerzo = service.NuevoEsfuerzo(e.repoProyectos, e.repoHistorias, e.repoIntegrantes, memory.NuevoTareas(),
		memory.NuevoRegistros(), func() time.Time { return e.hoy.Add(15 * time.Hour) })

	sc.Step(`^el proyecto del (\d\d/\d\d/\d{4}) al (\d\d/\d\d/\d{4}) con la historia "([^"]*)", la integrante "([^"]*)" y hoy (\d\d/\d\d/\d{4})$`, e.elProyecto)
	sc.Step(`^que "([^"]*)" ya registró "([^"]*)" horas el "([^"]*)"$`, func(nombre, horas, fecha string) error {
		e.registrar(nombre, horas, fecha, "Trabajo previo")
		if e.err != nil {
			return fmt.Errorf("no se pudo cargar el registro previo: %w", e.err)
		}
		return nil
	})
	sc.Step(`^que "([^"]*)" está dada de baja$`, e.estaDadaDeBaja)
	sc.Step(`^"([^"]*)" registra "([^"]*)" horas el "([^"]*)" en "([^"]*)"$`, func(nombre, horas, fecha, actividad string) error {
		e.registrar(nombre, horas, fecha, actividad)
		return nil
	})
	sc.Step(`^se registra el esfuerzo$`, func() error { return e.resultado("registrado") })
	sc.Step(`^el esfuerzo da como resultado "([^"]*)"$`, e.resultado)
	sc.Step(`^la historia suma "([^"]*)" horas reales$`, e.sumaHorasReales)
}

func fechaDDMM(texto string) time.Time {
	t, _ := time.Parse("02/01/2006", texto)
	return t
}

func (e *hu24) elProyecto(inicio, fin, titulo, nombre, hoy string) error {
	e.hoy = fechaDDMM(hoy)
	p, err := service.NuevoProyectos(e.repoProyectos, func() time.Time { return fechaDDMM(inicio) }).Crear(context.Background(),
		project.Datos{Nombre: "Software Metrics", FechaInicio: fechaDDMM(inicio), FechaFin: fechaDDMM(fin)})
	if err != nil {
		return fmt.Errorf("no se pudo crear el proyecto: %w", err)
	}
	e.proyectoID = p.ID
	if e.historia, err = service.NuevoHistorias(e.repoProyectos, e.repoHistorias).Crear(context.Background(), p.ID,
		backlog.Datos{Titulo: titulo, Prioridad: backlog.PrioridadAlta}); err != nil {
		return fmt.Errorf("no se pudo crear la historia: %w", err)
	}
	i, err := e.integrantes.Registrar(context.Background(), p.ID,
		member.Datos{Nombre: nombre, Email: strings.ToLower(nombre) + "@mail.com", Rol: member.RolProductBuilder})
	if err != nil {
		return fmt.Errorf("no se pudo registrar a %s: %w", nombre, err)
	}
	e.porNombre[nombre] = i
	return nil
}

func (e *hu24) estaDadaDeBaja(nombre string) error {
	_, err := e.integrantes.DarDeBaja(context.Background(), e.proyectoID, e.porNombre[nombre].ID)
	return err
}

func (e *hu24) registrar(nombre, horas, fecha, actividad string) {
	_, e.err = e.esfuerzo.Registrar(context.Background(), e.proyectoID, e.historia.Numero,
		effort.DatosRegistro{IntegranteID: e.porNombre[nombre].ID, Fecha: fecha, Actividad: actividad, Horas: horas})
}

func (e *hu24) resultado(esperado string) error {
	if esperado == "registrado" {
		if e.err != nil {
			return fmt.Errorf("se esperaba que se registre y hubo error: %w", e.err)
		}
		return nil
	}
	if e.err == nil {
		return fmt.Errorf("se esperaba el error %q y se registró", esperado)
	}
	if got := mensajeDelError(e.err); got != esperado {
		return fmt.Errorf("error = %q, se esperaba %q", got, esperado)
	}
	return nil
}

func (e *hu24) sumaHorasReales(esperado string) error {
	v, err := e.esfuerzo.Ver(context.Background(), e.proyectoID, e.historia.Numero)
	if err != nil {
		return err
	}
	if got := effort.FormatearHoras(v.HorasReales); got != esperado {
		return fmt.Errorf("la historia suma %q horas reales, se esperaba %q", got, esperado)
	}
	return nil
}
