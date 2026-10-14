package memory

import (
	"context"
	"slices"
	"sync"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
)

// Criterios guarda los criterios de aceptación de todas las historias en una lista en memoria.
type Criterios struct {
	mu       sync.Mutex
	lista    []backlog.Criterio
	ultimoID int64
}

// NuevoCriterios crea un repositorio vacío.
func NuevoCriterios() *Criterios {
	return &Criterios{}
}

// Guardar le asigna el próximo ID al criterio, lo guarda y lo devuelve con el ID.
func (r *Criterios) Guardar(_ context.Context, c backlog.Criterio) (backlog.Criterio, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ultimoID++
	c.ID = r.ultimoID
	r.lista = append(r.lista, c)
	return c, nil
}

// ListarPorHistoria devuelve los criterios de la historia en el orden en que se cargaron.
func (r *Criterios) ListarPorHistoria(ctx context.Context, historiaID int64) ([]backlog.Criterio, error) {
	return r.ListarPorHistorias(ctx, []int64{historiaID})
}

// ListarPorHistorias devuelve los criterios de todas esas historias, en el orden en que se cargaron.
func (r *Criterios) ListarPorHistorias(_ context.Context, historiaIDs []int64) ([]backlog.Criterio, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var lista []backlog.Criterio
	for _, c := range r.lista {
		if slices.Contains(historiaIDs, c.HistoriaID) {
			lista = append(lista, c)
		}
	}
	return lista, nil
}

// BuscarPorID devuelve el criterio dentro de su historia, o backlog.ErrCriterioNoEncontrado. Con el id de un
// criterio de otra historia también es "no encontrado" (RN8).
func (r *Criterios) BuscarPorID(_ context.Context, historiaID, id int64) (backlog.Criterio, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range r.lista {
		if c.ID == id && c.HistoriaID == historiaID {
			return c, nil
		}
	}
	return backlog.Criterio{}, backlog.ErrCriterioNoEncontrado
}

// Actualizar guarda los cambios de un criterio que ya existe, o devuelve backlog.ErrCriterioNoEncontrado.
func (r *Criterios) Actualizar(_ context.Context, c backlog.Criterio) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, existente := range r.lista {
		if existente.ID == c.ID && existente.HistoriaID == c.HistoriaID {
			r.lista[i] = c
			return nil
		}
	}
	return backlog.ErrCriterioNoEncontrado
}
