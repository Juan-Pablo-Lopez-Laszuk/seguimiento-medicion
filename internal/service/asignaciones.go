package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/metrics"
)

// RepositorioAsignaciones es lo que el caso de uso necesita del repositorio de historias. Lo implementan
// internal/store/memory (tests) e internal/store/postgres (Supabase).
type RepositorioAsignaciones interface {
	ListadorHistorias
	BuscadorHistorias
	// ActualizarSprint guarda juntos el estado y el sprint de la historia (nil = sin sprint), o devuelve
	// backlog.ErrHistoriaNoEncontrada si no existe.
	ActualizarSprint(ctx context.Context, historiaID int64, estado backlog.Estado, sprintID *int64) error
}

// Asignaciones agrupa los casos de uso del Sprint Backlog (HU-12): ver un sprint con sus historias, asignarle una
// historia y quitársela. Las reglas están en backlog.Historia.AsignarA y QuitarDelSprint.
type Asignaciones struct {
	proyectos BuscadorProyectos
	sprints   ListadorSprints
	historias RepositorioAsignaciones
	criterios *Criterios
}

// NuevoAsignaciones arma el caso de uso. criterios trae el progreso de las historias (HU-07) con una sola consulta.
func NuevoAsignaciones(proyectos BuscadorProyectos, sprints ListadorSprints, historias RepositorioAsignaciones, criterios *Criterios) *Asignaciones {
	return &Asignaciones{proyectos: proyectos, sprints: sprints, historias: historias, criterios: criterios}
}

// VistaSprint es lo que muestra la página de un sprint (sección 3 de la spec de HU-12).
type VistaSprint struct {
	Proyecto      project.Proyecto
	Sprint        sprint.Sprint
	Abierto       bool               // false si el sprint está cerrado: no se le asignan historias (RN2)
	Asignadas     []backlog.Historia // las del sprint, en el orden del backlog
	Comprometidos int                // SP de las asignadas
	Candidatas    []Candidata        // las del backlog en Pendiente y sin sprint
	Aviso         string             // aviso por velocidad (RN5), vacío si no supera
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
	v := VistaSprint{Proyecto: p, Sprint: destino, Abierto: destino.Estado != sprint.EstadoCerrado}
	var candidatas []backlog.Historia
	for _, h := range historias {
		switch {
		case h.SprintID != nil && *h.SprintID == destino.ID:
			v.Asignadas = append(v.Asignadas, h)
		case h.SprintID == nil && h.Estado == backlog.EstadoPendiente:
			candidatas = append(candidatas, h)
		}
	}
	progresos, err := s.criterios.Progresos(ctx, candidatas)
	if err != nil {
		return VistaSprint{}, err
	}
	abierto := datosDe(destino)
	abierto.Cerrado = false // el motivo es el de la historia, aunque el sprint esté cerrado
	for _, h := range candidatas {
		_, errAsignar := h.AsignarA(abierto, progresos[h.ID], nil)
		v.Candidatas = append(v.Candidatas, Candidata{Historia: h, Motivo: motivo(errAsignar)})
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

// buscarSprint trae el proyecto, el Sprint N y todos los sprints del proyecto, o el error de "no encontrado".
func (s *Asignaciones) buscarSprint(ctx context.Context, proyectoID int64, numero int) (project.Proyecto, sprint.Sprint, []sprint.Sprint, error) {
	p, err := s.proyectos.BuscarPorID(ctx, proyectoID)
	if err != nil {
		return project.Proyecto{}, sprint.Sprint{}, nil, fmt.Errorf("buscar el proyecto %d: %w", proyectoID, err)
	}
	sprints, err := s.sprints.ListarPorProyecto(ctx, proyectoID)
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
	historial := make([]metrics.Sprint, 0, len(sprints))
	for _, s := range sprints {
		m := metrics.Sprint{Numero: s.Numero, Estado: metrics.SprintPlanificado}
		if s.Estado == sprint.EstadoCerrado {
			m.Estado = metrics.SprintCerrado
		}
		if s.SPCompletados != nil {
			m.SPCompletados = float64(*s.SPCompletados)
		}
		historial = append(historial, m)
	}
	v, err := metrics.CalcularVelocidad(historial, metrics.VentanaPorDefecto)
	if err != nil {
		return "", fmt.Errorf("calcular la velocidad: %w", err)
	}
	if v.SprintsUsados == 0 || float64(comprometidos) <= v.Valor {
		return "", nil
	}
	velocidad := strings.Replace(fmt.Sprintf("%.1f", v.Valor), ".", ",", 1)
	return fmt.Sprintf("Ojo: el Sprint %d suma %d SP y la velocidad del equipo es %s SP", destino.Numero, comprometidos, velocidad), nil
}
