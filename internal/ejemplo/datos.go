package ejemplo

import (
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/defect"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
)

// Las fechas del ejemplo, en días desde hoy. Los sprints duran dos semanas: el 1 y el 2 ya terminaron, el 3 es el
// activo (hoy es su octavo día) y el 4 todavía no empezó.
const (
	inicioProyecto = -35
	finProyecto    = 27
)

// Quiénes trabajan en cada tarea: el índice en integrantesDeEjemplo.
const (
	juanPablo = iota
	mariano
	carolina
)

// integrantesDeEjemplo son el equipo, sin el email: lo pone Cargar.
var integrantesDeEjemplo = []member.Datos{
	{Nombre: "Juan Pablo López", Rol: member.RolAgileEnabler},
	{Nombre: "Mariano Piastrellini", Rol: member.RolProductBuilder},
	{Nombre: "Carolina Bravo", Rol: member.RolProductBuilder},
}

type sprintDeEjemplo struct {
	numero       int
	objetivo     string
	desde, hasta int   // días desde hoy
	asignadas    []int // números de las historias que se le asignan
	completadas  []int // las que se terminan
	empezadas    []int // las que quedan En progreso
	cerrar       bool
}

// sprintsDeEjemplo dan 8 de 16 SP en el Sprint 1 y 11 de 13 en el 2: velocidad 9,5. El Sprint 3 compromete 12 SP, más
// que la velocidad, así la página del sprint muestra el aviso (HU-12).
var sprintsDeEjemplo = []sprintDeEjemplo{
	{
		numero: 1, objetivo: "Registrar pacientes y ver la agenda del día", desde: -35, hasta: -22,
		asignadas: []int{1, 2, 3}, completadas: []int{1, 2}, empezadas: []int{3}, cerrar: true,
	},
	{
		numero: 2, objetivo: "Reservar y cancelar turnos sin superposiciones", desde: -21, hasta: -8,
		asignadas: []int{3, 4, 6}, completadas: []int{3, 4}, empezadas: []int{6}, cerrar: true,
	},
	{
		numero: 3, objetivo: "Avisar a los pacientes y ordenar la agenda", desde: -7, hasta: 6,
		asignadas: []int{5, 6, 7}, completadas: []int{5}, empezadas: []int{6},
	},
	{numero: 4, objetivo: "Reportes para la dirección del consultorio", desde: 7, hasta: 20},
}

type historiaDeEjemplo struct {
	titulo      string
	descripcion string
	prioridad   backlog.Prioridad
	storyPoints string // "" = sin estimar
	criterios   []string
	tareas      []tareaDeEjemplo
}

type tareaDeEjemplo struct {
	titulo      string
	responsable int    // juanPablo, mariano o carolina
	horas       string // estimadas
	trabajo     []horasDeUnDia
}

// horasDeUnDia son las horas que el responsable trabajó en la tarea un día.
type horasDeUnDia struct {
	dia   int // días desde hoy
	horas string
}

// historiasDeEjemplo son el backlog, en orden: la primera es la HU-1. Suman 54 h estimadas y 55 reales. La HU-8 y la
// HU-10 quedan listas para asignar en la demo, y la HU-9 sin estimar ni criterios.
var historiasDeEjemplo = []historiaDeEjemplo{
	{
		titulo: "Registrar pacientes", prioridad: backlog.PrioridadAlta, storyPoints: "3",
		descripcion: "Como recepcionista quiero registrar a los pacientes para poder darles turnos.",
		criterios:   []string{"Pide nombre, DNI y teléfono", "No acepta un DNI que ya está registrado"},
		tareas: []tareaDeEjemplo{
			{"Formulario de alta", mariano, "4", []horasDeUnDia{{-33, "5"}}},
			{"Validaciones del alta", carolina, "3", []horasDeUnDia{{-32, "3"}}},
		},
	},
	{
		titulo: "Ver la agenda del día", prioridad: backlog.PrioridadAlta, storyPoints: "5",
		descripcion: "Como médico quiero ver los turnos del día para organizar la atención.",
		criterios:   []string{"Muestra los turnos ordenados por horario", "Marca los sobreturnos"},
		tareas: []tareaDeEjemplo{
			{"Vista de la agenda", juanPablo, "6", []horasDeUnDia{{-30, "4"}, {-29, "3"}}},
			{"Consulta de turnos", mariano, "4", []horasDeUnDia{{-28, "4"}}},
		},
	},
	{
		titulo: "Reservar un turno", prioridad: backlog.PrioridadAlta, storyPoints: "8",
		descripcion: "Como recepcionista quiero reservar un turno para un paciente en un horario libre.",
		criterios: []string{
			"No deja reservar un horario ocupado", "No deja reservar en el pasado", "Confirma la reserva en pantalla",
		},
		tareas: []tareaDeEjemplo{
			{"Reglas de la reserva", carolina, "8", []horasDeUnDia{{-26, "4"}, {-18, "5"}}},
			{"Pantalla de reserva", mariano, "6", []horasDeUnDia{{-25, "3"}, {-17, "4"}}},
			{"Pruebas de la reserva", juanPablo, "3", []horasDeUnDia{{-15, "4"}}},
		},
	},
	{
		titulo: "Cancelar un turno", prioridad: backlog.PrioridadMedia, storyPoints: "3",
		descripcion: "Como recepcionista quiero cancelar un turno para liberar el horario.",
		criterios:   []string{"Pide confirmación antes de cancelar", "Libera el horario en la agenda"},
		tareas:      []tareaDeEjemplo{{"Cancelación", juanPablo, "4", []horasDeUnDia{{-14, "3"}}}},
	},
	{
		titulo: "Recordatorio por email", prioridad: backlog.PrioridadMedia, storyPoints: "5",
		descripcion: "Como paciente quiero recibir un recordatorio para no olvidarme del turno.",
		criterios:   []string{"Sale 24 h antes del turno", "Incluye la fecha, la hora y la dirección"},
		tareas: []tareaDeEjemplo{
			{"Envío de emails", mariano, "5", []horasDeUnDia{{-6, "6"}}},
			{"Plantilla del email", carolina, "2", []horasDeUnDia{{-5, "2"}}},
		},
	},
	{
		titulo: "Bloquear horarios", prioridad: backlog.PrioridadMedia, storyPoints: "2",
		descripcion: "Como médico quiero bloquear horarios para mis vacaciones o congresos.",
		criterios:   []string{"Un horario bloqueado no se puede reservar"},
		tareas:      []tareaDeEjemplo{{"Bloqueos de la agenda", carolina, "3", []horasDeUnDia{{-12, "1"}, {-3, "2"}}}},
	},
	{
		titulo: "Lista de espera", prioridad: backlog.PrioridadBaja, storyPoints: "5",
		descripcion: "Como paciente quiero anotarme en una lista de espera para conseguir un turno antes.",
		criterios:   []string{"Avisa al primero de la lista cuando se libera un turno"},
		tareas:      []tareaDeEjemplo{{"Lista de espera", juanPablo, "6", []horasDeUnDia{{-2, "2"}}}},
	},
	{
		titulo: "Reporte mensual de turnos", prioridad: backlog.PrioridadBaja, storyPoints: "3",
		descripcion: "Como director quiero un reporte mensual para ver cuántos turnos se atendieron.",
		criterios:   []string{"Cuenta los turnos atendidos y los cancelados"},
	},
	{
		titulo: "Historial del paciente", prioridad: backlog.PrioridadMedia,
		descripcion: "Como médico quiero ver las consultas anteriores de un paciente.",
	},
	{
		titulo: "Exportar la agenda", prioridad: backlog.PrioridadBaja, storyPoints: "2",
		descripcion: "Como recepcionista quiero exportar la agenda del día para imprimirla.",
		criterios:   []string{"Descarga la agenda del día en PDF"},
	},
}

type defectoDeEjemplo struct {
	descripcion string
	severidad   defect.Severidad
	sprint      int // número del sprint en que se detectó
	historia    int // número de la historia; 0 si no es de una
	estados     []defect.Estado
	resueltoEn  int // número del sprint en que se resolvió, si pasa por Resuelto
}

// defectosDeEjemplo son 5 detectados (2 en el Sprint 1, 1 en el 2 y 2 en el 3) y 3 resueltos (1 en el Sprint 1 y 2 en
// el 2).
var defectosDeEjemplo = []defectoDeEjemplo{
	{
		descripcion: "El alta acepta un DNI que ya está registrado", severidad: defect.SeveridadAlta, sprint: 1, historia: 1,
		estados: []defect.Estado{defect.EstadoEnProgreso, defect.EstadoResuelto, defect.EstadoCerrado}, resueltoEn: 1,
	},
	{
		descripcion: "La agenda no muestra los sobreturnos", severidad: defect.SeveridadMedia, sprint: 1, historia: 2,
		estados: []defect.Estado{defect.EstadoEnProgreso, defect.EstadoResuelto}, resueltoEn: 2,
	},
	{
		descripcion: "Se puede reservar un turno en un horario que ya pasó", severidad: defect.SeveridadCritica, sprint: 2, historia: 3,
		estados: []defect.Estado{defect.EstadoEnProgreso, defect.EstadoResuelto, defect.EstadoCerrado}, resueltoEn: 2,
	},
	{
		descripcion: "El recordatorio sale sin la dirección del consultorio", severidad: defect.SeveridadBaja, sprint: 3, historia: 5,
	},
	{
		descripcion: "Cancelar un turno no libera el horario en la agenda", severidad: defect.SeveridadAlta, sprint: 3, historia: 4,
		estados: []defect.Estado{defect.EstadoEnProgreso},
	},
}
