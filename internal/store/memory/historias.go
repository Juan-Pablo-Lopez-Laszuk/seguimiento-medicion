package memory

import (
	"context"
	"sync"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
)

// Historias guarda las historias de todos los proyectos en una lista en memoria.
type Historias struct {
	mu       sync.Mutex
	lista    []backlog.Historia
	ultimoID int64
}

// NuevoHistorias crea un repositorio vacío.
func NuevoHistorias() *Historias {
	return &Historias{}
}

// Guardar le asigna el próximo ID a la historia, la guarda y la devuelve con el ID.
func (r *Historias) Guardar(_ context.Context, h backlog.Historia) (backlog.Historia, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ultimoID++
	h.ID = r.ultimoID
	r.lista = append(r.lista, copiarHistoria(h))
	return copiarHistoria(h), nil
}

// ListarPorProyecto devuelve una copia de las historias del proyecto, en el orden en que se crearon.
func (r *Historias) ListarPorProyecto(_ context.Context, proyectoID int64) ([]backlog.Historia, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var lista []backlog.Historia
	for _, h := range r.lista {
		if h.ProyectoID == proyectoID {
			lista = append(lista, copiarHistoria(h))
		}
	}
	return lista, nil
}

// copiarHistoria copia también los story points, el sprint y la fecha de completado, que son punteros, para que el repositorio y quien lo usa
// no compartan el mismo valor.
func copiarHistoria(h backlog.Historia) backlog.Historia {
	if h.StoryPoints != nil {
		puntos := *h.StoryPoints
		h.StoryPoints = &puntos
	}
	h.SprintID = copiarID(h.SprintID)
	h.CompletadaEn = copiarFecha(h.CompletadaEn)
	return h
}

func copiarFecha(f *time.Time) *time.Time {
	if f == nil {
		return nil
	}
	copia := *f
	return &copia
}

func copiarID(id *int64) *int64 {
	if id == nil {
		return nil
	}
	copia := *id
	return &copia
}

// BuscarPorNumero devuelve la historia con ese número (HU-N) dentro del proyecto, o backlog.ErrHistoriaNoEncontrada.
func (r *Historias) BuscarPorNumero(_ context.Context, proyectoID int64, numero int) (backlog.Historia, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, h := range r.lista {
		if h.ProyectoID == proyectoID && h.Numero == numero {
			return copiarHistoria(h), nil
		}
	}
	return backlog.Historia{}, backlog.ErrHistoriaNoEncontrada
}

// ActualizarEstado guarda el estado nuevo de la historia sin tocar los demás datos, o devuelve
// backlog.ErrHistoriaNoEncontrada si no existe.
func (r *Historias) ActualizarEstado(_ context.Context, historiaID int64, estado backlog.Estado) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, h := range r.lista {
		if h.ID == historiaID {
			r.lista[i].Estado = estado
			return nil
		}
	}
	return backlog.ErrHistoriaNoEncontrada
}

// ActualizarSprint guarda juntos el estado y el sprint de la historia (nil = sin sprint) sin tocar los demás datos, o
// devuelve backlog.ErrHistoriaNoEncontrada si no existe. Lo usa HU-12 al asignar y al quitar.
func (r *Historias) ActualizarSprint(_ context.Context, historiaID int64, estado backlog.Estado, sprintID *int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, h := range r.lista {
		if h.ID == historiaID {
			r.lista[i].Estado, r.lista[i].SprintID = estado, copiarID(sprintID)
			return nil
		}
	}
	return backlog.ErrHistoriaNoEncontrada
}

// ActualizarCompletada guarda juntos el estado y la fecha de completado de la historia sin tocar los demás datos, o
// devuelve backlog.ErrHistoriaNoEncontrada si no existe. Lo usa HU-13 al completar.
func (r *Historias) ActualizarCompletada(_ context.Context, historiaID int64, estado backlog.Estado, completadaEn *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, h := range r.lista {
		if h.ID == historiaID {
			r.lista[i].Estado, r.lista[i].CompletadaEn = estado, copiarFecha(completadaEn)
			return nil
		}
	}
	return backlog.ErrHistoriaNoEncontrada
}
