package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// escenarioAsignacion es el proyecto de la sección 6 de la spec de HU-12, guardado directo en los repositorios en
// memoria: el Sprint 1 cerrado con 16 SP completados, el 2 activo, el 3 futuro y el backlog de cinco historias:
// HU-1 Login (5 SP, 2 criterios), HU-2 Registro (sin estimar), HU-3 Perfil (3 SP, sin criterios), HU-4 Reportes
// (0 SP) y HU-5 Ayuda (Hecha).
type escenarioAsignacion struct {
	casos      *service.Asignaciones
	proyectos  *memory.Proyectos
	sprints    *memory.Sprints
	historias  *memory.Historias
	criterios  *memory.Criterios
	ids        map[int]int64 // ID de cada sprint, por número
	proyectoID int64
}

// nuevoEscenarioAsignacion arma el escenario. Con primeroCerrado en false el Sprint 1 está abierto: no hay velocidad.
func nuevoEscenarioAsignacion(t *testing.T, primeroCerrado bool) *escenarioAsignacion {
	t.Helper()
	ctx := context.Background()
	e := &escenarioAsignacion{
		proyectos: memory.NuevoProyectos(), sprints: memory.NuevoSprints(),
		historias: memory.NuevoHistorias(), criterios: memory.NuevoCriterios(), ids: map[int]int64{},
	}
	p, err := e.proyectos.Guardar(ctx, project.Proyecto{Nombre: "Software Metrics", FechaInicio: fechaISO(t, "2026-10-05"), FechaFin: fechaISO(t, "2026-11-01")})
	if err != nil {
		t.Fatalf("no se pudo guardar el proyecto: %v", err)
	}
	e.proyectoID = p.ID
	for numero := 1; numero <= 3; numero++ {
		inicio := fechaISO(t, "2026-10-05").AddDate(0, 0, 7*(numero-1))
		s := sprint.Sprint{ProyectoID: p.ID, Numero: numero, Objetivo: "Objetivo", FechaInicio: inicio, FechaFin: inicio.AddDate(0, 0, 6), Estado: sprint.EstadoPlanificado}
		if numero == 1 && primeroCerrado {
			s.Estado, s.SPCompletados = sprint.EstadoCerrado, puntos(16)
		}
		guardado, err := e.sprints.Guardar(ctx, s)
		if err != nil {
			t.Fatalf("no se pudo guardar el Sprint %d: %v", numero, err)
		}
		e.ids[numero] = guardado.ID
	}
	e.historia(t, "Login", puntos(5), 2, backlog.EstadoPendiente)
	e.historia(t, "Registro", nil, 1, backlog.EstadoPendiente)
	e.historia(t, "Perfil", puntos(3), 0, backlog.EstadoPendiente)
	e.historia(t, "Reportes", puntos(0), 1, backlog.EstadoPendiente)
	e.historia(t, "Ayuda", puntos(2), 1, backlog.EstadoHecho)
	e.casos = e.armar(e.historias)
	return e
}

// armar arma el caso de uso con los repositorios del escenario y el de historias que se le pasa.
func (e *escenarioAsignacion) armar(historias service.RepositorioAsignaciones) *service.Asignaciones {
	return service.NuevoAsignaciones(e.proyectos, e.sprints, historias, service.NuevoCriterios(e.proyectos, e.historias, e.criterios))
}

// historia guarda la siguiente historia del backlog con sus criterios (ninguno cumplido) y devuelve su número.
func (e *escenarioAsignacion) historia(t *testing.T, titulo string, sp *int, criterios int, estado backlog.Estado) int {
	t.Helper()
	ctx := context.Background()
	lista, _ := e.historias.ListarPorProyecto(ctx, e.proyectoID)
	h, err := e.historias.Guardar(ctx, backlog.Historia{ProyectoID: e.proyectoID, Numero: len(lista) + 1, Titulo: titulo, Prioridad: backlog.PrioridadAlta, Estado: estado, StoryPoints: sp})
	if err != nil {
		t.Fatalf("no se pudo guardar la historia %q: %v", titulo, err)
	}
	conCriterios(t, e.criterios, h.ID, criterios, 0)
	return h.Numero
}

// enSprint guarda una historia lista ya asignada al sprint, directo en el repositorio, y devuelve su número.
func (e *escenarioAsignacion) enSprint(t *testing.T, numeroSprint int, sp int, estado backlog.Estado) int {
	t.Helper()
	numero := e.historia(t, "Ya asignada", puntos(sp), 1, backlog.EstadoPendiente)
	h, _ := e.historias.BuscarPorNumero(context.Background(), e.proyectoID, numero)
	id := e.ids[numeroSprint]
	if err := e.historias.ActualizarSprint(context.Background(), h.ID, estado, &id); err != nil {
		t.Fatalf("no se pudo asignar la historia: %v", err)
	}
	return numero
}

func (e *escenarioAsignacion) guardada(t *testing.T, numero int) backlog.Historia {
	t.Helper()
	h, err := e.historias.BuscarPorNumero(context.Background(), e.proyectoID, numero)
	if err != nil {
		t.Fatalf("no se pudo buscar la HU-%d: %v", numero, err)
	}
	return h
}

// CA-12.1 · la historia lista queda guardada En sprint y en el sprint, con el total comprometido y sin aviso
func TestAsignar_HistoriaLista_QuedaGuardadaEnElSprint(t *testing.T) {
	e := nuevoEscenarioAsignacion(t, true)

	a, err := e.casos.Asignar(context.Background(), e.proyectoID, 2, 1)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if a.Historia.Estado != backlog.EstadoEnSprint || a.Sprint.Numero != 2 || a.Comprometidos != 5 || a.Aviso != "" {
		t.Errorf("se esperaba HU-1 En sprint en el Sprint 2 con 5 SP y sin aviso, y se obtuvo %+v", a)
	}
	h := e.guardada(t, 1)
	if h.Estado != backlog.EstadoEnSprint || h.SprintID == nil || *h.SprintID != e.ids[2] {
		t.Errorf("la historia guardada quedó %q en el sprint %v", h.Estado, h.SprintID)
	}
}

// CA-12.1 y CA-12.2 · lo que no se asigna no cambia nada en el repositorio
func TestAsignar_Rechazos_NoGuardanNada(t *testing.T) {
	casos := []struct {
		nombre         string
		sprint, numero int
		err            error
	}{
		{"sin estimar", 2, 2, backlog.ErrSinEstimar},
		{"sin criterios", 2, 3, backlog.ErrSinCriterios},
		{"hecha", 2, 5, backlog.ErrYaHecha},
		{"sprint cerrado", 1, 1, backlog.ErrSprintCerrado},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			e := nuevoEscenarioAsignacion(t, true)
			antes := e.guardada(t, c.numero)

			_, err := e.casos.Asignar(context.Background(), e.proyectoID, c.sprint, c.numero)

			if !errors.Is(err, c.err) {
				t.Fatalf("error = %v, se esperaba %v", err, c.err)
			}
			if despues := e.guardada(t, c.numero); despues.Estado != antes.Estado || despues.SprintID != nil {
				t.Errorf("la historia cambió: quedó %q en el sprint %v", despues.Estado, despues.SprintID)
			}
		})
	}
}

// CA-12.2 · una historia en un sprint abierto no se asigna a otro, y el error nombra el sprint en el que está
func TestAsignar_YaEnOtroSprint(t *testing.T) {
	e := nuevoEscenarioAsignacion(t, true)
	if _, err := e.casos.Asignar(context.Background(), e.proyectoID, 2, 1); err != nil {
		t.Fatalf("no se pudo asignar al Sprint 2: %v", err)
	}

	_, err := e.casos.Asignar(context.Background(), e.proyectoID, 3, 1)

	if !errors.Is(err, backlog.ErrEnOtroSprint) || err.Error() != "HU-1 ya está en el Sprint 2" {
		t.Errorf("error = %v, se esperaba \"HU-1 ya está en el Sprint 2\"", err)
	}
}

// CA-12.3 · si lo comprometido supera la velocidad se asigna igual, con el aviso; igualarla no avisa
func TestAsignar_AvisoPorVelocidad(t *testing.T) {
	e := nuevoEscenarioAsignacion(t, true)
	e.enSprint(t, 2, 8, backlog.EstadoEnSprint)
	e.enSprint(t, 2, 3, backlog.EstadoEnProgreso)

	a, err := e.casos.Asignar(context.Background(), e.proyectoID, 2, 4) // 11 + 0 = 11 SP
	if err != nil || a.Aviso != "" {
		t.Fatalf("con 11 SP no hay aviso: se obtuvo %q, %v", a.Aviso, err)
	}
	a, err = e.casos.Asignar(context.Background(), e.proyectoID, 2, 1) // 11 + 5 = 16 SP, igual a la velocidad
	if err != nil || a.Aviso != "" {
		t.Fatalf("con 16 SP (igual a la velocidad) no hay aviso: se obtuvo %q, %v", a.Aviso, err)
	}
	e.historia(t, "Una más", puntos(2), 1, backlog.EstadoPendiente)

	a, err = e.casos.Asignar(context.Background(), e.proyectoID, 2, 8) // 16 + 2 = 18 SP

	if err != nil {
		t.Fatalf("el aviso no bloquea: %v", err)
	}
	if a.Aviso != "Ojo: el Sprint 2 suma 18 SP y la velocidad del equipo es 16,0 SP" {
		t.Errorf("aviso = %q", a.Aviso)
	}
}

// CA-12.3 y decisión 8 · sin sprints cerrados no hay velocidad y no se avisa
func TestAsignar_SinSprintsCerrados_NoAvisa(t *testing.T) {
	e := nuevoEscenarioAsignacion(t, false)
	e.enSprint(t, 2, 21, backlog.EstadoEnSprint)
	e.enSprint(t, 2, 21, backlog.EstadoEnSprint)

	a, err := e.casos.Asignar(context.Background(), e.proyectoID, 2, 1)

	if err != nil || a.Comprometidos != 47 || a.Aviso != "" {
		t.Errorf("se esperaba 47 SP sin aviso y se obtuvo %+v, %v", a, err)
	}
}

// Sección 7 · proyecto, sprint o historia que no existen
func TestAsignar_NoEncontrados(t *testing.T) {
	e := nuevoEscenarioAsignacion(t, true)
	ctx := context.Background()

	if _, err := e.casos.Asignar(ctx, 99, 2, 1); !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("proyecto: error = %v, se esperaba project.ErrNoEncontrado", err)
	}
	if _, err := e.casos.Asignar(ctx, e.proyectoID, 9, 1); !errors.Is(err, sprint.ErrNoEncontrado) {
		t.Errorf("sprint: error = %v, se esperaba sprint.ErrNoEncontrado", err)
	}
	if _, err := e.casos.Asignar(ctx, e.proyectoID, 2, 99); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("historia: error = %v, se esperaba backlog.ErrHistoriaNoEncontrada", err)
	}
}

// historiasQueFallanAlAsignar es el repositorio en memoria con la base caída al guardar el sprint de la historia.
type historiasQueFallanAlAsignar struct {
	*memory.Historias
}

func (historiasQueFallanAlAsignar) ActualizarSprint(context.Context, int64, backlog.Estado, *int64) error {
	return errSprintsCaidos
}

// Si falla la base al guardar, el error se devuelve y no es uno de la regla
func TestAsignar_ErrorDeLaBase(t *testing.T) {
	e := nuevoEscenarioAsignacion(t, true)

	_, err := e.armar(historiasQueFallanAlAsignar{e.historias}).Asignar(context.Background(), e.proyectoID, 2, 1)

	if !errors.Is(err, errSprintsCaidos) {
		t.Errorf("error = %v, se esperaba el de la base", err)
	}
}

// RN4 · quitar una historia En sprint la deja Pendiente y sin sprint
func TestQuitar_HistoriaEnSprint(t *testing.T) {
	e := nuevoEscenarioAsignacion(t, true)
	numero := e.enSprint(t, 2, 5, backlog.EstadoEnSprint)

	if _, err := e.casos.Quitar(context.Background(), e.proyectoID, 2, numero); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if h := e.guardada(t, numero); h.Estado != backlog.EstadoPendiente || h.SprintID != nil {
		t.Errorf("la historia quedó %q en el sprint %v", h.Estado, h.SprintID)
	}
}

// RN4 · una historia En progreso no se quita; una que no está en ese sprint es "no encontrada" en él
func TestQuitar_Rechazos(t *testing.T) {
	e := nuevoEscenarioAsignacion(t, true)
	empezada := e.enSprint(t, 2, 5, backlog.EstadoEnProgreso)
	enOtro := e.enSprint(t, 3, 5, backlog.EstadoEnSprint)
	ctx := context.Background()

	if _, err := e.casos.Quitar(ctx, e.proyectoID, 2, empezada); !errors.Is(err, backlog.ErrYaEmpezada) {
		t.Errorf("en progreso: error = %v, se esperaba backlog.ErrYaEmpezada", err)
	}
	if _, err := e.casos.Quitar(ctx, e.proyectoID, 2, enOtro); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("en otro sprint: error = %v, se esperaba backlog.ErrHistoriaNoEncontrada", err)
	}
	if h := e.guardada(t, enOtro); h.SprintID == nil || *h.SprintID != e.ids[3] {
		t.Errorf("la historia del Sprint 3 no tenía que cambiar y quedó en %v", h.SprintID)
	}
}

// Sección 3 · la página del sprint: asignadas con el total comprometido, y candidatas con el motivo si no están listas
func TestVerSprint_AsignadasYCandidatas(t *testing.T) {
	e := nuevoEscenarioAsignacion(t, true)
	e.enSprint(t, 2, 13, backlog.EstadoEnSprint)
	e.enSprint(t, 3, 8, backlog.EstadoEnSprint) // de otro sprint: no aparece

	v, err := e.casos.Ver(context.Background(), e.proyectoID, 2)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if v.Proyecto.ID != e.proyectoID || v.Sprint.Numero != 2 || !v.Abierto {
		t.Errorf("se esperaba el Sprint 2 abierto del proyecto y se obtuvo %+v", v)
	}
	if len(v.Asignadas) != 1 || v.Asignadas[0].Numero != 6 || v.Comprometidos != 13 {
		t.Errorf("se esperaba solo la HU-6 con 13 SP comprometidos y se obtuvo %+v, %d", v.Asignadas, v.Comprometidos)
	}
	motivos := map[int]string{1: "", 2: "sin estimar", 3: "sin criterios de aceptación", 4: ""} // HU-5 está Hecha
	if len(v.Candidatas) != len(motivos) {
		t.Fatalf("se esperaban 4 candidatas y se obtuvo %+v", v.Candidatas)
	}
	for _, c := range v.Candidatas {
		if esperado, ok := motivos[c.Numero]; !ok || c.Motivo != esperado {
			t.Errorf("HU-%d: motivo = %q, se esperaba %q", c.Numero, c.Motivo, esperado)
		}
	}
	if v.Aviso != "" {
		t.Errorf("13 SP no superan la velocidad y se obtuvo el aviso %q", v.Aviso)
	}
}

// La página de un sprint cerrado lo dice; la de uno que supera la velocidad muestra el aviso
func TestVerSprint_CerradoYAviso(t *testing.T) {
	e := nuevoEscenarioAsignacion(t, true)
	e.enSprint(t, 2, 21, backlog.EstadoEnSprint)

	cerrado, err := e.casos.Ver(context.Background(), e.proyectoID, 1)
	if err != nil || cerrado.Abierto {
		t.Errorf("el Sprint 1 está cerrado: se obtuvo %+v, %v", cerrado, err)
	}
	excedido, err := e.casos.Ver(context.Background(), e.proyectoID, 2)
	if err != nil || excedido.Aviso != "Ojo: el Sprint 2 suma 21 SP y la velocidad del equipo es 16,0 SP" {
		t.Errorf("aviso = %q, %v", excedido.Aviso, err)
	}
	if _, err := e.casos.Ver(context.Background(), e.proyectoID, 9); !errors.Is(err, sprint.ErrNoEncontrado) {
		t.Errorf("sprint que no existe: error = %v, se esperaba sprint.ErrNoEncontrado", err)
	}
}
