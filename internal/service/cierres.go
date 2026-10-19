package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
)

// RepositorioCierres guarda el cierre de un sprint todo junto (HU-14 RN6). Lo implementan internal/store/memory
// (tests) e internal/store/postgres (Supabase).
type RepositorioCierres interface {
	// Cerrar guarda el sprint Cerrado con sus SP, la foto y las historias que vuelven al backlog, o nada si falla.
	Cerrar(ctx context.Context, c sprint.Cierre) error
	// Foto devuelve la foto del cierre del sprint, ordenada por número de historia; vacía si no se cerró.
	Foto(ctx context.Context, sprintID int64) ([]sprint.FotoHistoria, error)
}

// Cierres es el caso de uso de cerrar el sprint activo (spec: specs/HU-14-cerrar-sprint.md). La regla está en
// sprint.Cerrar.
type Cierres struct {
	proyectos BuscadorProyectos
	sprints   ListadorSprints
	historias ListadorHistorias
	repo      RepositorioCierres
	ahora     func() time.Time
}

// NuevoCierres arma el caso de uso. ahora devuelve la fecha y hora actual (en los tests, una fija), para saber cuál es
// el sprint activo en la fecha de Argentina.
func NuevoCierres(proyectos BuscadorProyectos, sprints ListadorSprints, historias ListadorHistorias, repo RepositorioCierres, ahora func() time.Time) *Cierres {
	return &Cierres{proyectos: proyectos, sprints: sprints, historias: historias, repo: repo, ahora: ahora}
}

// VistaCierre es lo que muestra la pantalla de confirmación (sección 3 de la spec).
type VistaCierre struct {
	Proyecto     project.Proyecto
	Sprint       sprint.Sprint
	Planificados int                // SP de las historias del sprint
	Completados  int                // SP de las Hechas
	Historias    int                // cantidad de historias del sprint
	Hechas       int                // cantidad de Hechas
	Devueltas    []backlog.Historia // las que vuelven al backlog, con el estado que tienen ahora
	DiasAntes    int                // días que faltan para la fecha de fin; 0 el último día o si ya venció
}

// Preparar devuelve lo que va a pasar al cerrar el Sprint N, sin guardar nada. Si no se puede cerrar, devuelve la
// vista con el proyecto y el sprint (para mostrar el motivo) y el error de la regla (sprint.ErrYaCerrado o
// ErrNoEsElActivo); si el proyecto o el sprint no existen, project.ErrNoEncontrado o sprint.ErrNoEncontrado.
func (c *Cierres) Preparar(ctx context.Context, proyectoID int64, numero int) (VistaCierre, error) {
	v, _, err := c.armar(ctx, proyectoID, numero)
	return v, err
}

// Cerrar cierra el Sprint N (HU-14) y guarda todo junto. Devuelve los mismos errores que Preparar, y no guarda nada si
// no se puede cerrar.
func (c *Cierres) Cerrar(ctx context.Context, proyectoID int64, numero int) (sprint.Cierre, error) {
	_, cierre, err := c.armar(ctx, proyectoID, numero)
	if err != nil {
		return sprint.Cierre{}, err
	}
	if err := c.repo.Cerrar(ctx, cierre); err != nil {
		return sprint.Cierre{}, fmt.Errorf("cerrar el Sprint %d del proyecto %d: %w", numero, proyectoID, err)
	}
	return cierre, nil
}

// armar busca el sprint y sus historias y aplica la regla de cierre con la fecha de hoy en Argentina.
func (c *Cierres) armar(ctx context.Context, proyectoID int64, numero int) (VistaCierre, sprint.Cierre, error) {
	p, s, sprints, err := buscarSprint(ctx, c.proyectos, c.sprints, proyectoID, numero)
	if err != nil {
		return VistaCierre{}, sprint.Cierre{}, err
	}
	v := VistaCierre{Proyecto: p, Sprint: s}
	historias, err := c.historias.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return v, sprint.Cierre{}, fmt.Errorf("listar las historias del proyecto %d: %w", proyectoID, err)
	}
	hoy := domain.HoyEnArgentina(c.ahora())
	cierre, err := sprint.Cerrar(s, sprints, historias, hoy)
	if err != nil {
		return v, sprint.Cierre{}, err
	}

	v.Planificados, v.Completados = *cierre.Sprint.SPPlanificados, *cierre.Sprint.SPCompletados
	v.Historias = len(cierre.Foto)
	for _, f := range cierre.Foto {
		if f.Completada {
			v.Hechas++
		}
	}
	for _, h := range historias {
		if h.SprintID != nil && *h.SprintID == s.ID && h.Estado != backlog.EstadoHecho {
			v.Devueltas = append(v.Devueltas, h)
		}
	}
	v.DiasAntes = domain.DiasRestantes(hoy, s.FechaFin)
	return v, cierre, nil
}
