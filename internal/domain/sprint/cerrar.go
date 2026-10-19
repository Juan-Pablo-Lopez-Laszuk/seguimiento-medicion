package sprint

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/metrics"
)

// Errores del cierre (sección 7 de specs/HU-14-cerrar-sprint.md). Se devuelven con un mensaje que nombra el sprint.
var (
	// ErrYaCerrado se devuelve al cerrar un sprint que ya está cerrado (CA-14.4).
	ErrYaCerrado = errors.New("el sprint ya está cerrado")
	// ErrNoEsElActivo se devuelve al cerrar un sprint que no es el activo (CA-14.1).
	ErrNoEsElActivo = errors.New("solo se cierra el sprint activo")
)

// errorDeCierre es un error de la regla con un mensaje para el usuario; sigue siendo su error base para errors.Is.
type errorDeCierre struct {
	base    error
	mensaje string
}

func (e errorDeCierre) Error() string { return e.mensaje }
func (e errorDeCierre) Unwrap() error { return e.base }

// FotoHistoria es una fila de la foto del cierre (sprint_historias): una historia que estaba en el sprint, con sus
// story points en ese momento (nil si estaba sin estimar) y si estaba Hecha (RN3).
type FotoHistoria struct {
	HistoriaID  int64
	Numero      int
	StoryPoints *int
	Completada  bool
}

// Cierre es lo que hay que guardar al cerrar un sprint, todo junto (RN6): el sprint Cerrado con sus SP, la foto de sus
// historias y las historias que vuelven al backlog Pendientes y sin sprint.
type Cierre struct {
	Sprint    Sprint
	Foto      []FotoHistoria
	Devueltas []backlog.Historia
}

// Cerrar cierra el sprint s (spec: specs/HU-14-cerrar-sprint.md). sprints son los del proyecto, para saber cuál es el
// activo (RN1), e historias pueden ser todas las del proyecto: se toman las que están en s. No modifica lo que recibe.
func Cerrar(s Sprint, sprints []Sprint, historias []backlog.Historia, hoy time.Time) (Cierre, error) {
	if err := puedeCerrarse(s, sprints, hoy); err != nil {
		return Cierre{}, err
	}
	delSprint := historiasDe(s, historias)
	sp, err := metrics.CalcularStoryPoints(paraMetricas(delSprint))
	if err != nil {
		return Cierre{}, err
	}
	planificados, completados := int(sp.Planificados), int(sp.Completados)
	s.Estado, s.SPPlanificados, s.SPCompletados = EstadoCerrado, &planificados, &completados

	cierre := Cierre{Sprint: s}
	for _, h := range delSprint {
		hecha := h.Estado == backlog.EstadoHecho
		cierre.Foto = append(cierre.Foto, FotoHistoria{HistoriaID: h.ID, Numero: h.Numero, StoryPoints: copiarPuntos(h.StoryPoints), Completada: hecha})
		if hecha {
			continue
		}
		// La vuelta a Pendiente es la de HU-08 ("al cerrar el sprint"); los criterios no importan para ir a Pendiente.
		devuelta, err := h.CambiarEstado(backlog.EstadoPendiente, backlog.Progreso{})
		if err != nil {
			return Cierre{}, fmt.Errorf("devolver %s al backlog: %w", h.Identificador(), err)
		}
		devuelta.SprintID, devuelta.StoryPoints = nil, copiarPuntos(h.StoryPoints)
		cierre.Devueltas = append(cierre.Devueltas, devuelta)
	}
	return cierre, nil
}

// puedeCerrarse devuelve el error de la regla si s no se puede cerrar: ya está cerrado (CA-14.4) o no es el activo
// (CA-14.1, HU-04 RN2).
func puedeCerrarse(s Sprint, sprints []Sprint, hoy time.Time) error {
	if s.Estado == EstadoCerrado {
		return errorDeCierre{base: ErrYaCerrado, mensaje: fmt.Sprintf("el Sprint %d ya está cerrado", s.Numero)}
	}
	activo, hay := Activo(sprints, hoy)
	switch {
	case !hay:
		return errorDeCierre{base: ErrNoEsElActivo, mensaje: ErrNoEsElActivo.Error() + ": hoy no hay ninguno"}
	case activo.ID != s.ID:
		return errorDeCierre{base: ErrNoEsElActivo, mensaje: fmt.Sprintf("%s (el Sprint %d)", ErrNoEsElActivo, activo.Numero)}
	}
	return nil
}

// historiasDe devuelve las historias que están en el sprint, ordenadas por número.
func historiasDe(s Sprint, historias []backlog.Historia) []backlog.Historia {
	var delSprint []backlog.Historia
	for _, h := range historias {
		if h.SprintID != nil && *h.SprintID == s.ID {
			delSprint = append(delSprint, h)
		}
	}
	slices.SortFunc(delSprint, func(a, b backlog.Historia) int { return a.Numero - b.Numero })
	return delSprint
}

// paraMetricas convierte las historias a los datos que usa el cálculo de HU-30.
func paraMetricas(historias []backlog.Historia) []metrics.Historia {
	datos := make([]metrics.Historia, len(historias))
	for i, h := range historias {
		datos[i] = metrics.Historia{Hecha: h.Estado == backlog.EstadoHecho, SinEstimar: h.StoryPoints == nil}
		if h.StoryPoints != nil {
			datos[i].SP = float64(*h.StoryPoints)
		}
	}
	return datos
}

// copiarPuntos devuelve una copia de los story points, así la foto no comparte el valor con la historia.
func copiarPuntos(p *int) *int {
	if p == nil {
		return nil
	}
	copia := *p
	return &copia
}
