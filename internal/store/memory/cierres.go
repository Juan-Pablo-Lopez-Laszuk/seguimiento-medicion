package memory

import (
	"context"
	"slices"
	"sync"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
)

// Cierres guarda el cierre de los sprints (HU-14): cambia el sprint y las historias en sus repositorios y guarda la foto
// de cada sprint cerrado. Es la versión en memoria de la transacción de PostgreSQL.
type Cierres struct {
	mu        sync.Mutex
	sprints   *Sprints
	historias *Historias
	fotos     map[int64][]sprint.FotoHistoria // por ID de sprint
}

// NuevoCierres crea el repositorio sobre los de sprints e historias, y le avisa al de historias que ya hay sprints
// cerrados, para que EnSprintCerrado (HU-06) los tenga en cuenta.
func NuevoCierres(sprints *Sprints, historias *Historias) *Cierres {
	c := &Cierres{sprints: sprints, historias: historias, fotos: map[int64][]sprint.FotoHistoria{}}
	historias.mu.Lock()
	historias.cierres = c
	historias.mu.Unlock()
	return c
}

// Cerrar guarda todo el cierre o nada (HU-14 RN6): devuelve sprint.ErrNoEncontrado si el sprint no existe,
// sprint.ErrYaCerrado si ya está cerrado y backlog.ErrHistoriaNoEncontrada si falta alguna de las que vuelven al backlog.
func (c *Cierres) Cerrar(_ context.Context, cierre sprint.Cierre) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sprints.mu.Lock()
	defer c.sprints.mu.Unlock()
	c.historias.mu.Lock()
	defer c.historias.mu.Unlock()

	// Primero se valida todo, después se cambia: así un error no deja el cierre a medias.
	i := slices.IndexFunc(c.sprints.lista, func(s sprint.Sprint) bool { return s.ID == cierre.Sprint.ID })
	if i < 0 {
		return sprint.ErrNoEncontrado
	}
	if c.sprints.lista[i].Estado == sprint.EstadoCerrado {
		return sprint.ErrYaCerrado
	}
	posiciones := make([]int, len(cierre.Devueltas))
	for k, devuelta := range cierre.Devueltas {
		posiciones[k] = slices.IndexFunc(c.historias.lista, func(h backlog.Historia) bool { return h.ID == devuelta.ID })
		if posiciones[k] < 0 {
			return backlog.ErrHistoriaNoEncontrada
		}
	}

	s := &c.sprints.lista[i]
	s.Estado, s.SPPlanificados, s.SPCompletados = sprint.EstadoCerrado, copiarEntero(cierre.Sprint.SPPlanificados), copiarEntero(cierre.Sprint.SPCompletados)
	for k, devuelta := range cierre.Devueltas {
		c.historias.lista[posiciones[k]].Estado, c.historias.lista[posiciones[k]].SprintID = devuelta.Estado, nil
	}
	c.fotos[cierre.Sprint.ID] = copiarFoto(cierre.Foto)
	return nil
}

// Foto devuelve una copia de la foto del cierre del sprint, ordenada por número de historia; vacía si no se cerró.
func (c *Cierres) Foto(_ context.Context, sprintID int64) ([]sprint.FotoHistoria, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return copiarFoto(c.fotos[sprintID]), nil
}

// esDeUnSprintCerrado dice si la historia está en un sprint cerrado o aparece en la foto de uno (HU-06 RN6). Se llama
// sin tener tomado el candado del repositorio de historias.
func (c *Cierres) esDeUnSprintCerrado(h backlog.Historia) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.sprints.mu.Lock()
	defer c.sprints.mu.Unlock()
	for _, s := range c.sprints.lista {
		if s.Estado != sprint.EstadoCerrado {
			continue
		}
		if h.SprintID != nil && *h.SprintID == s.ID {
			return true
		}
		if slices.ContainsFunc(c.fotos[s.ID], func(f sprint.FotoHistoria) bool { return f.HistoriaID == h.ID }) {
			return true
		}
	}
	return false
}

func copiarFoto(foto []sprint.FotoHistoria) []sprint.FotoHistoria {
	copia := make([]sprint.FotoHistoria, len(foto))
	for i, f := range foto {
		f.StoryPoints = copiarEntero(f.StoryPoints)
		copia[i] = f
	}
	return copia
}

func copiarEntero(n *int) *int {
	if n == nil {
		return nil
	}
	copia := *n
	return &copia
}
