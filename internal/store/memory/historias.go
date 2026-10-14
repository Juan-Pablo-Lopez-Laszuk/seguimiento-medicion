package memory

import (
	"context"
	"sync"

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

// copiarHistoria copia también los story points, que son un puntero, para que el repositorio y quien lo usa no
// compartan el mismo valor.
func copiarHistoria(h backlog.Historia) backlog.Historia {
	if h.StoryPoints != nil {
		puntos := *h.StoryPoints
		h.StoryPoints = &puntos
	}
	return h
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
