package features

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// hu26 guarda el estado de un escenario de HU-26 (consultar y corregir registros de esfuerzo), con sus propios
// repositorios en memoria.
type hu26 struct {
	proyectos    *memory.Proyectos
	sprints      *memory.Sprints
	historias    *memory.Historias
	integrantes  *memory.Integrantes
	tareas       *memory.Tareas
	registros    *memory.Registros
	casos        *service.RegistrosEsfuerzo
	hoy          time.Time
	proyecto     project.Proyecto
	porNombre    map[string]member.Integrante
	historiaDe   map[string]backlog.Historia // por "HU-N"
	porActividad map[string]effort.Registro
	vista        service.VistaRegistros
	err          error
}

// registrarPasosHU26 conecta cada frase de features/hu-26-registros-esfuerzo.feature con una función de Go.
func registrarPasosHU26(sc *godog.ScenarioContext) {
	e := &hu26{
		proyectos: memory.NuevoProyectos(), sprints: memory.NuevoSprints(), historias: memory.NuevoHistorias(),
		integrantes: memory.NuevoIntegrantes(), tareas: memory.NuevoTareas(), registros: memory.NuevoRegistros(),
		porNombre: map[string]member.Integrante{}, historiaDe: map[string]backlog.Historia{}, porActividad: map[string]effort.Registro{},
	}
	ahora := func() time.Time { return e.hoy.Add(15 * time.Hour) }
	e.casos = service.NuevoRegistrosEsfuerzo(e.proyectos, e.sprints, e.historias, e.integrantes, e.tareas, e.registros,
		service.NuevoHoras(e.proyectos, e.sprints, e.historias, e.tareas, e.registros), ahora)

	sc.Step(`^el proyecto con el Sprint 1 del (\d\d/\d\d/\d{4}) al (\d\d/\d\d/\d{4}), el Sprint 2 del (\d\d/\d\d/\d{4}) al (\d\d/\d\d/\d{4}) y hoy (\d\d/\d\d/\d{4})$`, e.elProyecto)
	sc.Step(`^el equipo con "([^"]*)" de Agile Enabler y "([^"]*)" y "([^"]*)" de Product Builder$`, e.elEquipo)
	sc.Step(`^las historias con tareas estimadas:$`, e.lasHistorias)
	sc.Step(`^estos registros de horas:$`, e.losRegistros)
	sc.Step(`^consulto los registros de "([^"]*)" en el "([^"]*)"$`, e.consultarDeEn)
	sc.Step(`^consulto los registros de la historia "([^"]*)"$`, e.consultarHistoria)
	sc.Step(`^veo solo los registros "([^"]*)"$`, e.veoSolo)
	sc.Step(`^veo los (\d+) registros$`, e.veoLos)
	sc.Step(`^el total de horas de la lista es "([^"]*)"$`, e.elTotal)
	sc.Step(`^la historia "([^"]*)" tiene "([^"]*)" horas estimadas y "([^"]*)" horas reales$`, e.horasDeLaHistoria)
	sc.Step(`^"([^"]*)" corrige el registro "([^"]*)" a "([^"]*)" horas$`, e.corregir)
	sc.Step(`^el registro queda corregido$`, e.quedaCorregido)
	sc.Step(`^la corrección da el error "([^"]*)"$`, e.daElError)
}

func (e *hu26) elProyecto(inicio1, fin1, inicio2, fin2, hoy string) error {
	ctx := context.Background()
	e.hoy = fechaDDMM(hoy)
	var err error
	if e.proyecto, err = service.NuevoProyectos(e.proyectos, func() time.Time { return fechaDDMM(inicio1) }).Crear(ctx,
		project.Datos{Nombre: "Software Metrics", FechaInicio: fechaDDMM(inicio1), FechaFin: fechaDDMM(fin2)}); err != nil {
		return fmt.Errorf("no se pudo crear el proyecto: %w", err)
	}
	sprints := service.NuevoSprints(e.proyectos, e.sprints)
	for _, d := range []sprint.Datos{{Objetivo: "Uno", FechaInicio: fechaDDMM(inicio1), FechaFin: fechaDDMM(fin1)}, {Objetivo: "Dos", FechaInicio: fechaDDMM(inicio2), FechaFin: fechaDDMM(fin2)}} {
		if _, err := sprints.Crear(ctx, e.proyecto.ID, d); err != nil {
			return fmt.Errorf("no se pudo crear el sprint %q: %w", d.Objetivo, err)
		}
	}
	return nil
}

func (e *hu26) elEquipo(agileEnabler, uno, otro string) error {
	casos := service.NuevoIntegrantes(e.proyectos, e.integrantes)
	for _, d := range []member.Datos{{Nombre: agileEnabler, Rol: member.RolAgileEnabler}, {Nombre: uno, Rol: member.RolProductBuilder}, {Nombre: otro, Rol: member.RolProductBuilder}} {
		d.Email = strings.ReplaceAll(strings.ToLower(d.Nombre), " ", "") + "@mail.com"
		i, err := casos.Registrar(context.Background(), e.proyecto.ID, d)
		if err != nil {
			return fmt.Errorf("no se pudo registrar a %s: %w", d.Nombre, err)
		}
		e.porNombre[d.Nombre] = i
	}
	return nil
}

func (e *hu26) lasHistorias(tabla *godog.Table) error {
	ctx := context.Background()
	historias := service.NuevoHistorias(e.proyectos, e.historias)
	tareas := service.NuevoTareas(e.proyectos, e.historias, e.integrantes, e.tareas)
	var responsable int64
	for _, i := range e.porNombre {
		responsable = i.ID
	}
	for _, fila := range tabla.Rows[1:] {
		h, err := historias.Crear(ctx, e.proyecto.ID, backlog.Datos{Titulo: fila.Cells[1].Value, Prioridad: backlog.PrioridadAlta})
		if err != nil {
			return fmt.Errorf("no se pudo crear la historia %s: %w", fila.Cells[0].Value, err)
		}
		if _, err := tareas.Agregar(ctx, e.proyecto.ID, h.Numero, effort.DatosTarea{Titulo: "Tarea de " + h.Titulo, ResponsableID: responsable, Horas: fila.Cells[2].Value}); err != nil {
			return fmt.Errorf("no se pudo agregar la tarea de %s: %w", fila.Cells[0].Value, err)
		}
		e.historiaDe[fila.Cells[0].Value] = h
	}
	return nil
}

func (e *hu26) losRegistros(tabla *godog.Table) error {
	esfuerzo := service.NuevoEsfuerzo(e.proyectos, e.historias, e.integrantes, e.tareas, e.registros, func() time.Time { return e.hoy.Add(15 * time.Hour) })
	for _, fila := range tabla.Rows[1:] {
		c := fila.Cells
		r, err := esfuerzo.Registrar(context.Background(), e.proyecto.ID, e.historiaDe[c[1].Value].Numero,
			effort.DatosRegistro{IntegranteID: e.porNombre[c[0].Value].ID, Fecha: c[2].Value, Actividad: c[3].Value, Horas: c[4].Value})
		if err != nil {
			return fmt.Errorf("no se pudo registrar %q: %w", c[3].Value, err)
		}
		e.porActividad[c[3].Value] = r
	}
	return nil
}

func (e *hu26) consultar(f service.FiltroRegistros) error {
	var err error
	e.vista, err = e.casos.Consultar(context.Background(), e.proyecto.ID, f)
	return err
}

// consultarDeEn filtra por integrante y sprint; un nombre o un sprint que no existen se mandan igual (RN3).
func (e *hu26) consultarDeEn(nombre, sprint string) error {
	f := service.FiltroRegistros{IntegranteID: 999}
	if i, ok := e.porNombre[nombre]; ok {
		f.IntegranteID = i.ID
	}
	f.Sprint, _ = strconv.Atoi(strings.TrimPrefix(sprint, "Sprint "))
	return e.consultar(f)
}

func (e *hu26) consultarHistoria(historia string) error {
	return e.consultar(service.FiltroRegistros{Historia: e.historiaDe[historia].Numero})
}

func (e *hu26) veoSolo(esperadas string) error {
	nombres := make([]string, len(e.vista.Registros))
	for i, r := range e.vista.Registros {
		nombres[i] = r.Actividad
	}
	if got := strings.Join(nombres, ", "); got != esperadas {
		return fmt.Errorf("se ven los registros %q, se esperaba %q", got, esperadas)
	}
	return nil
}

func (e *hu26) veoLos(cantidad int) error {
	if len(e.vista.Registros) != cantidad {
		return fmt.Errorf("se ven %d registros, se esperaban %d", len(e.vista.Registros), cantidad)
	}
	return nil
}

func (e *hu26) elTotal(esperado string) error {
	if got := effort.FormatearHoras(e.vista.Total); got != esperado {
		return fmt.Errorf("el total es %q, se esperaba %q", got, esperado)
	}
	return nil
}

func (e *hu26) horasDeLaHistoria(historia, estimadas, reales string) error {
	if err := e.consultarHistoria(historia); err != nil {
		return err
	}
	if len(e.vista.PorHistoria) != 1 {
		return fmt.Errorf("se esperaba una fila de horas para %s y hay %d", historia, len(e.vista.PorHistoria))
	}
	d := e.vista.PorHistoria[0]
	if effort.FormatearHoras(d.Estimadas) != estimadas || effort.FormatearHoras(d.Reales) != reales {
		return fmt.Errorf("%s tiene %s h estimadas y %s h reales, se esperaba %s y %s", historia,
			effort.FormatearHoras(d.Estimadas), effort.FormatearHoras(d.Reales), estimadas, reales)
	}
	return nil
}

func (e *hu26) corregir(nombre, actividad, horas string) error {
	r := e.porActividad[actividad]
	_, e.err = e.casos.Corregir(context.Background(), e.proyecto.ID, r.ID, e.porNombre[nombre].ID,
		effort.DatosRegistro{TareaID: r.TareaID, Fecha: r.Fecha.Format(time.DateOnly), Actividad: r.Actividad, Horas: horas})
	return nil
}

func (e *hu26) quedaCorregido() error {
	if e.err != nil {
		return fmt.Errorf("se esperaba que se corrija y hubo error: %w", e.err)
	}
	return nil
}

func (e *hu26) daElError(esperado string) error {
	if e.err == nil {
		return fmt.Errorf("se esperaba el error %q y se corrigió", esperado)
	}
	if got := mensajeDelError(e.err); got != esperado {
		return fmt.Errorf("error = %q, se esperaba %q", got, esperado)
	}
	return nil
}
