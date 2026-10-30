// Package ejemplo carga un proyecto de ejemplo con tres sprints de historia: dos cerrados y uno activo, con historias,
// criterios, tareas, horas y defectos. Sirve para la demo (y su plan B sin internet) y para probar la aplicación a mano.
//
// Todo se carga con los casos de uso de internal/service, los mismos que usan las pantallas: los datos pasan por las
// mismas reglas y quedan como si el equipo los hubiera cargado a mano. Las fechas se arman alrededor de hoy, así el
// ejemplo sirve cualquier día.
package ejemplo

import (
	"context"
	"fmt"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/defect"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
)

// NombrePorDefecto es el nombre del proyecto de ejemplo si no se pide otro.
const NombrePorDefecto = "Turnos de consultorio (ejemplo)"

// EmailsPorDefecto son los emails de los integrantes si no se piden otros: Juan Pablo, Mariano y Carolina.
var EmailsPorDefecto = [3]string{"juanpablo@example.com", "mariano@example.com", "carolina@example.com"}

// Datos cambian el nombre del proyecto y los emails de los integrantes. Con los emails de las cuentas del equipo, cada
// uno queda vinculado con su integrante al ingresar (TEC-05). Un valor vacío usa el de por defecto.
type Datos struct {
	Nombre string
	Emails [3]string // de Juan Pablo, Mariano y Carolina, en ese orden
}

// Cargar crea el proyecto de ejemplo con dep y lo devuelve. Si el nombre ya existe devuelve el error de validación de
// project (ErrNombreRepetido) sin cargar nada: así el ejemplo no se duplica. Si falla a mitad de camino, lo cargado
// hasta ahí queda guardado y el error dice en qué paso falló.
//
// Las fechas salen de hoy en Argentina porque los casos de uso también la usan: el sprint activo, la fecha de
// completado y el límite de las horas registradas.
func Cargar(ctx context.Context, dep server.Dependencias, d Datos) (project.Proyecto, error) {
	c := carga{ctx: ctx, dep: dep, datos: d, hoy: domain.HoyEnArgentina(time.Now())}
	pasos := []struct {
		nombre string
		hacer  func() error
	}{
		{"crear el proyecto", c.crearProyecto},
		{"registrar los integrantes", c.registrarIntegrantes},
		{"crear los sprints", c.crearSprints},
		{"cargar el backlog", c.cargarBacklog},
		{"trabajar los sprints", c.trabajarSprints},
		{"registrar los defectos", c.registrarDefectos},
	}
	for _, paso := range pasos {
		if err := paso.hacer(); err != nil {
			return c.proyecto, fmt.Errorf("%s: %w", paso.nombre, err)
		}
	}
	return c.proyecto, nil
}

// carga es lo que se va armando durante la carga: los IDs que devuelven los casos de uso para los pasos siguientes.
type carga struct {
	ctx         context.Context
	dep         server.Dependencias
	datos       Datos
	hoy         time.Time
	proyecto    project.Proyecto
	integrantes []member.Integrante // en el orden de integrantesDeEjemplo
	sprints     []sprint.Sprint     // en el orden de sprintsDeEjemplo
	historias   []backlog.Historia  // en el orden de historiasDeEjemplo: la i es la HU-(i+1)
	criterios   map[int][]int64     // los IDs de los criterios de cada historia, por número
}

// dia devuelve la fecha que está a n días de hoy (n negativo: en el pasado).
func (c *carga) dia(n int) time.Time { return c.hoy.AddDate(0, 0, n) }

func (c *carga) crearProyecto() error {
	nombre := c.datos.Nombre
	if nombre == "" {
		nombre = NombrePorDefecto
	}
	p, err := c.dep.Proyectos.Crear(c.ctx, project.Datos{
		Nombre:      nombre,
		Descripcion: "Aplicación para que un consultorio administre los turnos de sus pacientes. Proyecto de ejemplo cargado con cmd/ejemplo.",
		FechaInicio: c.dia(inicioProyecto),
		FechaFin:    c.dia(finProyecto),
	})
	c.proyecto = p
	return err
}

func (c *carga) registrarIntegrantes() error {
	for i, datos := range integrantesDeEjemplo {
		datos.Email = EmailsPorDefecto[i]
		if email := c.datos.Emails[i]; email != "" {
			datos.Email = email
		}
		integrante, err := c.dep.Integrantes.Registrar(c.ctx, c.proyecto.ID, datos)
		if err != nil {
			return fmt.Errorf("%s: %w", datos.Nombre, err)
		}
		c.integrantes = append(c.integrantes, integrante)
	}
	return nil
}

func (c *carga) crearSprints() error {
	for _, s := range sprintsDeEjemplo {
		creado, err := c.dep.Sprints.Crear(c.ctx, c.proyecto.ID, sprint.Datos{
			Objetivo: s.objetivo, FechaInicio: c.dia(s.desde), FechaFin: c.dia(s.hasta),
		})
		if err != nil {
			return err
		}
		c.sprints = append(c.sprints, creado)
	}
	return nil
}

// cargarBacklog crea cada historia con sus criterios, sus tareas y las horas trabajadas en cada tarea.
func (c *carga) cargarBacklog() error {
	c.criterios = map[int][]int64{}
	for _, h := range historiasDeEjemplo {
		historia, err := c.dep.Historias.Crear(c.ctx, c.proyecto.ID, backlog.Datos{
			Titulo: h.titulo, Descripcion: h.descripcion, Prioridad: h.prioridad, StoryPoints: h.storyPoints,
		})
		if err != nil {
			return fmt.Errorf("%q: %w", h.titulo, err)
		}
		c.historias = append(c.historias, historia)
		for _, descripcion := range h.criterios {
			criterio, err := c.dep.Criterios.Agregar(c.ctx, c.proyecto.ID, historia.Numero, descripcion)
			if err != nil {
				return fmt.Errorf("%s, criterio %q: %w", historia.Identificador(), descripcion, err)
			}
			c.criterios[historia.Numero] = append(c.criterios[historia.Numero], criterio.ID)
		}
		for _, t := range h.tareas {
			if err := c.cargarTarea(historia, t); err != nil {
				return fmt.Errorf("%s, tarea %q: %w", historia.Identificador(), t.titulo, err)
			}
		}
	}
	return nil
}

func (c *carga) cargarTarea(h backlog.Historia, t tareaDeEjemplo) error {
	responsable := c.integrantes[t.responsable]
	tarea, err := c.dep.Tareas.Agregar(c.ctx, c.proyecto.ID, h.Numero, effort.DatosTarea{
		Titulo: t.titulo, ResponsableID: responsable.ID, Horas: t.horas,
	})
	if err != nil {
		return err
	}
	for _, trabajo := range t.trabajo {
		_, err := c.dep.Esfuerzo.Registrar(c.ctx, c.proyecto.ID, h.Numero, effort.DatosRegistro{
			IntegranteID: responsable.ID,
			TareaID:      tarea.ID,
			Fecha:        c.dia(trabajo.dia).Format(time.DateOnly),
			Actividad:    t.titulo,
			Horas:        trabajo.horas,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// trabajarSprints recorre los sprints en orden como lo haría el equipo: asigna las historias, las empieza, cumple los
// criterios y completa las terminadas, y cierra el sprint si corresponde (las que no se terminaron vuelven al backlog).
func (c *carga) trabajarSprints() error {
	for _, s := range sprintsDeEjemplo {
		for _, numero := range s.asignadas {
			if _, err := c.dep.Asignaciones.Asignar(c.ctx, c.proyecto.ID, s.numero, numero); err != nil {
				return err
			}
		}
		for _, numero := range append(append([]int{}, s.completadas...), s.empezadas...) {
			if _, err := c.dep.Asignaciones.Empezar(c.ctx, c.proyecto.ID, s.numero, numero); err != nil {
				return err
			}
		}
		for _, numero := range s.completadas {
			if err := c.completar(s.numero, numero); err != nil {
				return err
			}
		}
		if s.cerrar {
			if _, err := c.dep.Cierres.Cerrar(c.ctx, c.proyecto.ID, s.numero); err != nil {
				return err
			}
		}
	}
	return nil
}

// completar cumple todos los criterios de la historia y la pasa a Hecho.
func (c *carga) completar(numeroSprint, numero int) error {
	for _, id := range c.criterios[numero] {
		if _, err := c.dep.Criterios.Marcar(c.ctx, c.proyecto.ID, numero, id, true); err != nil {
			return err
		}
	}
	_, err := c.dep.Asignaciones.Completar(c.ctx, c.proyecto.ID, numeroSprint, numero)
	return err
}

func (c *carga) registrarDefectos() error {
	for _, d := range defectosDeEjemplo {
		datos := defect.Datos{Descripcion: d.descripcion, Severidad: d.severidad, SprintID: c.sprints[d.sprint-1].ID}
		if d.historia != 0 {
			datos.HistoriaID = c.historias[d.historia-1].ID
		}
		defecto, err := c.dep.Defectos.Registrar(c.ctx, c.proyecto.ID, datos)
		if err != nil {
			return fmt.Errorf("%q: %w", d.descripcion, err)
		}
		for _, estado := range d.estados {
			var resolucion int64
			if estado == defect.EstadoResuelto {
				resolucion = c.sprints[d.resueltoEn-1].ID
			}
			if _, err := c.dep.Defectos.CambiarEstado(c.ctx, c.proyecto.ID, defecto.ID, estado, resolucion); err != nil {
				return fmt.Errorf("%q a %s: %w", d.descripcion, estado, err)
			}
		}
	}
	return nil
}
