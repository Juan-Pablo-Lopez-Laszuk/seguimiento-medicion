package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
)

// RepositorioAsignaciones es lo que el caso de uso necesita del repositorio de historias. Lo implementan
// internal/store/memory (tests) e internal/store/postgres (Supabase).
type RepositorioAsignaciones interface {
	ListadorHistorias
	BuscadorHistorias
	// ActualizarSprint guarda juntos el estado y el sprint de la historia (nil = sin sprint), o devuelve
	// backlog.ErrHistoriaNoEncontrada si no existe.
	ActualizarSprint(ctx context.Context, historiaID int64, estado backlog.Estado, sprintID *int64) error
	// ActualizarCompletada guarda juntos el estado y la fecha de completado de la historia (HU-13), o devuelve
	// backlog.ErrHistoriaNoEncontrada si no existe.
	ActualizarCompletada(ctx context.Context, historiaID int64, estado backlog.Estado, completadaEn *time.Time) error
}

// Asignaciones agrupa los casos de uso de la página de un sprint: ver el sprint con sus historias, asignarle una
// historia y quitársela (HU-12), y empezar y completar sus historias (HU-13). Las reglas están en backlog.Historia
// (AsignarA, QuitarDelSprint, CambiarEstado y Completar).
type Asignaciones struct {
	proyectos BuscadorProyectos
	sprints   ListadorSprints
	historias RepositorioAsignaciones
	criterios *Criterios
	ahora     func() time.Time
}

// NuevoAsignaciones arma el caso de uso. criterios trae el progreso de las historias (HU-07) con una sola consulta;
// ahora devuelve la fecha y hora actual (en los tests, una fija), para el sprint activo y la fecha de completado.
func NuevoAsignaciones(proyectos BuscadorProyectos, sprints ListadorSprints, historias RepositorioAsignaciones, criterios *Criterios, ahora func() time.Time) *Asignaciones {
	return &Asignaciones{proyectos: proyectos, sprints: sprints, historias: historias, criterios: criterios, ahora: ahora}
}

// VistaSprint es lo que muestra la página de un sprint (sección 3 de la spec de HU-12).
type VistaSprint struct {
	Proyecto      project.Proyecto
	Sprint        sprint.Sprint
	Abierto       bool        // false si el sprint está cerrado: no se le asignan historias (HU-12 RN2)
	Activo        bool        // true si es el sprint activo: solo ahí se completan historias (HU-13 RN1)
	Asignadas     []Asignada  // las del sprint, en el orden del backlog
	Comprometidos int         // SP de las asignadas
	Completados   int         // SP de las asignadas Hechas (HU-13)
	Candidatas    []Candidata // las del backlog en Pendiente y sin sprint, en el orden del backlog (HU-09 RN9)
	Aviso         string      // aviso por velocidad (HU-12 RN5), vacío si no supera
}

// Asignada es una historia del sprint con el progreso de sus criterios, que dice si ya se puede completar (HU-13).
type Asignada struct {
	backlog.Historia
	Progreso backlog.Progreso
}

// Faltan es la cantidad de criterios que le faltan cumplir.
func (a Asignada) Faltan() int {
	return a.Progreso.Total - a.Progreso.Cumplidos
}

// Candidata es una historia del backlog que se podría asignar. Motivo dice por qué todavía no se puede, por ejemplo
// "sin estimar"; vacío si está lista.
type Candidata struct {
	backlog.Historia
	Motivo string
}

// Asignacion es el resultado de asignar: la historia como quedó, el sprint, el total comprometido con ella y el aviso
// por velocidad (vacío si no supera).
type Asignacion struct {
	Historia      backlog.Historia
	Sprint        sprint.Sprint
	Comprometidos int
	Aviso         string
}

// Ver devuelve el Sprint N del proyecto con sus historias asignadas y las candidatas. Si el proyecto no existe, el
// error es project.ErrNoEncontrado; si el sprint no existe en el proyecto, sprint.ErrNoEncontrado.
func (s *Asignaciones) Ver(ctx context.Context, proyectoID int64, numeroSprint int) (VistaSprint, error) {
	p, destino, sprints, err := s.buscarSprint(ctx, proyectoID, numeroSprint)
	if err != nil {
		return VistaSprint{}, err
	}
	historias, err := s.historias.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return VistaSprint{}, fmt.Errorf("listar las historias del proyecto %d: %w", proyectoID, err)
	}
	v := VistaSprint{Proyecto: p, Sprint: destino, Abierto: destino.Estado != sprint.EstadoCerrado, Activo: s.esElActivo(destino, sprints)}
	var asignadas, candidatas []backlog.Historia
	for _, h := range historias {
		switch {
		case h.SprintID != nil && *h.SprintID == destino.ID:
			asignadas = append(asignadas, h)
		case h.SprintID == nil && h.Estado == backlog.EstadoPendiente:
			candidatas = append(candidatas, h)
		}
	}
	progresos, err := s.criterios.Progresos(ctx, append(append([]backlog.Historia{}, asignadas...), candidatas...))
	if err != nil {
		return VistaSprint{}, err
	}
	for _, h := range asignadas {
		v.Asignadas = append(v.Asignadas, Asignada{Historia: h, Progreso: progresos[h.ID]})
		if h.Estado == backlog.EstadoHecho && h.StoryPoints != nil {
			v.Completados += *h.StoryPoints
		}
	}
	candidatas, _ = backlog.Ordenar(candidatas) // todas son Pendiente: el orden del backlog (HU-09 RN9)
	abierto := datosDe(destino)
	abierto.Cerrado = false // el motivo es el de la historia, aunque el sprint esté cerrado
	for _, h := range candidatas {
		_, errAsignar := h.AsignarA(abierto, progresos[h.ID], nil)
		v.Candidatas = append(v.Candidatas, Candidata{Historia: h, Motivo: motivo(errAsignar)})
	}
	if !v.Abierto && destino.SPPlanificados != nil && destino.SPCompletados != nil {
		// Un sprint cerrado muestra los SP de la foto del cierre (HU-14): las no completadas ya volvieron al backlog.
		v.Comprometidos, v.Completados = *destino.SPPlanificados, *destino.SPCompletados
		return v, nil
	}
	v.Comprometidos = comprometidos(historias, destino.ID)
	v.Aviso, err = avisoDeVelocidad(destino, v.Comprometidos, sprints)
	return v, err
}

// Asignar asigna la historia HU-N al Sprint M (HU-12) y la guarda En sprint y en ese sprint. Si no se puede, devuelve
// el error de la regla (backlog.ErrSinEstimar, ErrSinCriterios, ErrYaHecha, ErrEnOtroSprint, ErrSprintCerrado o los
// de HU-08) y no guarda nada; si el proyecto, el sprint o la historia no existen, project.ErrNoEncontrado,
// sprint.ErrNoEncontrado o backlog.ErrHistoriaNoEncontrada. Si los SP comprometidos superan la velocidad, se asigna
// igual y Aviso lo dice (RN5).
func (s *Asignaciones) Asignar(ctx context.Context, proyectoID int64, numeroSprint, numero int) (Asignacion, error) {
	_, destino, sprints, err := s.buscarSprint(ctx, proyectoID, numeroSprint)
	if err != nil {
		return Asignacion{}, err
	}
	h, err := s.buscarHistoria(ctx, proyectoID, numero)
	if err != nil {
		return Asignacion{}, err
	}
	progresos, err := s.criterios.Progresos(ctx, []backlog.Historia{h})
	if err != nil {
		return Asignacion{}, err
	}
	todos := make([]backlog.DatosSprint, len(sprints))
	for i, otro := range sprints {
		todos[i] = datosDe(otro)
	}
	asignada, err := h.AsignarA(datosDe(destino), progresos[h.ID], todos)
	if err != nil {
		return Asignacion{}, err
	}
	if err := s.historias.ActualizarSprint(ctx, h.ID, asignada.Estado, asignada.SprintID); err != nil {
		return Asignacion{}, fmt.Errorf("guardar la historia %d en el sprint %d: %w", h.ID, destino.ID, err)
	}
	historias, err := s.historias.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return Asignacion{}, fmt.Errorf("listar las historias del proyecto %d: %w", proyectoID, err)
	}
	a := Asignacion{Historia: asignada, Sprint: destino, Comprometidos: comprometidos(historias, destino.ID)}
	a.Aviso, err = avisoDeVelocidad(destino, a.Comprometidos, sprints)
	return a, err
}

// Quitar saca la historia HU-N del Sprint M y la guarda Pendiente y sin sprint (RN4). Si ya empezó, el error es
// backlog.ErrYaEmpezada; si no está en ese sprint, backlog.ErrHistoriaNoEncontrada, igual que si no existe.
func (s *Asignaciones) Quitar(ctx context.Context, proyectoID int64, numeroSprint, numero int) (backlog.Historia, error) {
	_, destino, _, err := s.buscarSprint(ctx, proyectoID, numeroSprint)
	if err != nil {
		return backlog.Historia{}, err
	}
	h, err := s.buscarHistoria(ctx, proyectoID, numero)
	if err != nil {
		return backlog.Historia{}, err
	}
	if h.SprintID == nil || *h.SprintID != destino.ID {
		return backlog.Historia{}, fmt.Errorf("la historia HU-%d no está en el Sprint %d: %w", numero, numeroSprint, backlog.ErrHistoriaNoEncontrada)
	}
	quitada, err := h.QuitarDelSprint()
	if err != nil {
		return backlog.Historia{}, err
	}
	if err := s.historias.ActualizarSprint(ctx, h.ID, quitada.Estado, nil); err != nil {
		return backlog.Historia{}, fmt.Errorf("quitar la historia %d del sprint %d: %w", h.ID, destino.ID, err)
	}
	return quitada, nil
}

// Empezar pasa la historia HU-N del Sprint M a En progreso, con la transición de HU-08, y sigue en su sprint
// (HU-13 RN2). Si no está en ese sprint, el error es backlog.ErrHistoriaNoEncontrada.
func (s *Asignaciones) Empezar(ctx context.Context, proyectoID int64, numeroSprint, numero int) (backlog.Historia, error) {
	_, h, _, progreso, err := s.historiaDelSprint(ctx, proyectoID, numeroSprint, numero)
	if err != nil {
		return backlog.Historia{}, err
	}
	empezada, err := h.CambiarEstado(backlog.EstadoEnProgreso, progreso)
	if err != nil {
		return backlog.Historia{}, err
	}
	if err := s.historias.ActualizarSprint(ctx, h.ID, empezada.Estado, empezada.SprintID); err != nil {
		return backlog.Historia{}, fmt.Errorf("guardar el estado de la historia %d: %w", h.ID, err)
	}
	return empezada, nil
}

// Completar pasa la historia HU-N del Sprint M a Hecho con la fecha de hoy en Argentina (HU-13). Si no se puede,
// devuelve el error de la regla (backlog.ErrNoEsElSprintActivo o los de HU-08) y no guarda nada; si la historia no
// está en ese sprint, backlog.ErrHistoriaNoEncontrada.
func (s *Asignaciones) Completar(ctx context.Context, proyectoID int64, numeroSprint, numero int) (backlog.Historia, error) {
	destino, h, sprints, progreso, err := s.historiaDelSprint(ctx, proyectoID, numeroSprint, numero)
	if err != nil {
		return backlog.Historia{}, err
	}
	hecha, err := h.Completar(s.esElActivo(destino, sprints), progreso, domain.HoyEnArgentina(s.ahora()))
	if err != nil {
		return backlog.Historia{}, err
	}
	if err := s.historias.ActualizarCompletada(ctx, h.ID, hecha.Estado, hecha.CompletadaEn); err != nil {
		return backlog.Historia{}, fmt.Errorf("guardar la historia %d completada: %w", h.ID, err)
	}
	return hecha, nil
}

// historiaDelSprint trae el Sprint M, la historia HU-N (que tiene que estar en ese sprint), los sprints del proyecto
// y el progreso de los criterios de la historia.
func (s *Asignaciones) historiaDelSprint(ctx context.Context, proyectoID int64, numeroSprint, numero int) (sprint.Sprint, backlog.Historia, []sprint.Sprint, backlog.Progreso, error) {
	_, destino, sprints, err := s.buscarSprint(ctx, proyectoID, numeroSprint)
	if err != nil {
		return sprint.Sprint{}, backlog.Historia{}, nil, backlog.Progreso{}, err
	}
	h, err := s.buscarHistoria(ctx, proyectoID, numero)
	if err != nil {
		return sprint.Sprint{}, backlog.Historia{}, nil, backlog.Progreso{}, err
	}
	if h.SprintID == nil || *h.SprintID != destino.ID {
		return sprint.Sprint{}, backlog.Historia{}, nil, backlog.Progreso{},
			fmt.Errorf("la historia HU-%d no está en el Sprint %d: %w", numero, numeroSprint, backlog.ErrHistoriaNoEncontrada)
	}
	progresos, err := s.criterios.Progresos(ctx, []backlog.Historia{h})
	if err != nil {
		return sprint.Sprint{}, backlog.Historia{}, nil, backlog.Progreso{}, err
	}
	return destino, h, sprints, progresos[h.ID], nil
}

// esElActivo dice si el sprint es el activo hoy, en la fecha de Argentina (HU-04 RN2 y RN6).
func (s *Asignaciones) esElActivo(destino sprint.Sprint, sprints []sprint.Sprint) bool {
	activo, ok := sprint.Activo(sprints, domain.HoyEnArgentina(s.ahora()))
	return ok && activo.ID == destino.ID
}

func (s *Asignaciones) buscarSprint(ctx context.Context, proyectoID int64, numero int) (project.Proyecto, sprint.Sprint, []sprint.Sprint, error) {
	return buscarSprint(ctx, s.proyectos, s.sprints, proyectoID, numero)
}

// buscarSprint trae el proyecto, el Sprint N y todos los sprints del proyecto, o el error de "no encontrado". La usan
// la página del sprint (HU-12 y HU-13) y el cierre (HU-14).
func buscarSprint(ctx context.Context, proyectos BuscadorProyectos, repo ListadorSprints, proyectoID int64, numero int) (project.Proyecto, sprint.Sprint, []sprint.Sprint, error) {
	p, err := proyectos.BuscarPorID(ctx, proyectoID)
	if err != nil {
		return project.Proyecto{}, sprint.Sprint{}, nil, fmt.Errorf("buscar el proyecto %d: %w", proyectoID, err)
	}
	sprints, err := repo.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return project.Proyecto{}, sprint.Sprint{}, nil, fmt.Errorf("listar los sprints del proyecto %d: %w", proyectoID, err)
	}
	for _, sp := range sprints {
		if sp.Numero == numero {
			return p, sp, sprints, nil
		}
	}
	return project.Proyecto{}, sprint.Sprint{}, nil, fmt.Errorf("buscar el Sprint %d del proyecto %d: %w", numero, proyectoID, sprint.ErrNoEncontrado)
}

func (s *Asignaciones) buscarHistoria(ctx context.Context, proyectoID int64, numero int) (backlog.Historia, error) {
	h, err := s.historias.BuscarPorNumero(ctx, proyectoID, numero)
	if err != nil {
		return backlog.Historia{}, fmt.Errorf("buscar la historia HU-%d del proyecto %d: %w", numero, proyectoID, err)
	}
	return h, nil
}

// datosDe es lo que la regla de backlog necesita saber del sprint.
func datosDe(s sprint.Sprint) backlog.DatosSprint {
	return backlog.DatosSprint{ID: s.ID, Numero: s.Numero, Cerrado: s.Estado == sprint.EstadoCerrado}
}

// motivo traduce por qué una candidata todavía no se puede asignar a lo que muestra la lista; vacío si está lista.
func motivo(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, backlog.ErrSinEstimar):
		return "sin estimar"
	case errors.Is(err, backlog.ErrSinCriterios):
		return "sin criterios de aceptación"
	}
	return err.Error()
}

// comprometidos suma los SP de las historias del sprint. Una sin estimar no puede estar asignada, y cuenta como 0.
func comprometidos(historias []backlog.Historia, sprintID int64) int {
	total := 0
	for _, h := range historias {
		if h.SprintID != nil && *h.SprintID == sprintID && h.StoryPoints != nil {
			total += *h.StoryPoints
		}
	}
	return total
}

// avisoDeVelocidad devuelve el aviso si los SP comprometidos del sprint superan la velocidad del equipo (RN5): la de
// HU-31, con los SP completados de los últimos sprints cerrados. Sin sprints cerrados no hay velocidad ni aviso.
func avisoDeVelocidad(destino sprint.Sprint, comprometidos int, sprints []sprint.Sprint) (string, error) {
	v, err := velocidadDe(sprints)
	if err != nil {
		return "", err
	}
	if v.SprintsUsados == 0 || float64(comprometidos) <= v.Valor {
		return "", nil
	}
	velocidad := strings.Replace(fmt.Sprintf("%.1f", v.Valor), ".", ",", 1)
	return fmt.Sprintf("Ojo: el Sprint %d suma %d SP y la velocidad del equipo es %s SP", destino.Numero, comprometidos, velocidad), nil
}
