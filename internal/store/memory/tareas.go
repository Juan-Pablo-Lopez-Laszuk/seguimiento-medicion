package memory

import (
	"context"
	"slices"
	"sync"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
)

// Tareas guarda las tareas de todas las historias en una lista en memoria (HU-25).
type Tareas struct {
	mu       sync.Mutex
	lista    []effort.Tarea
	ultimoID int64
}

// NuevoTareas crea un repositorio vacío.
func NuevoTareas() *Tareas {
	return &Tareas{}
}

// Guardar le asigna el próximo ID a la tarea, la guarda y la devuelve con el ID.
func (r *Tareas) Guardar(_ context.Context, t effort.Tarea) (effort.Tarea, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ultimoID++
	t.ID = r.ultimoID
	r.lista = append(r.lista, t)
	return t, nil
}

// ListarPorHistoria devuelve las tareas de la historia en el orden en que se cargaron (RN9).
func (r *Tareas) ListarPorHistoria(_ context.Context, historiaID int64) ([]effort.Tarea, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var lista []effort.Tarea
	for _, t := range r.lista {
		if t.HistoriaID == historiaID {
			lista = append(lista, t)
		}
	}
	return lista, nil
}

// BuscarPorID devuelve la tarea dentro de su historia, o effort.ErrTareaNoEncontrada. Con el id de una tarea de otra
// historia también es "no encontrado" (CA-25.2).
func (r *Tareas) BuscarPorID(_ context.Context, historiaID, id int64) (effort.Tarea, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.lista {
		if t.ID == id && t.HistoriaID == historiaID {
			return t, nil
		}
	}
	return effort.Tarea{}, effort.ErrTareaNoEncontrada
}

// Actualizar guarda los cambios de una tarea que ya existe en su historia, o devuelve effort.ErrTareaNoEncontrada.
func (r *Tareas) Actualizar(_ context.Context, t effort.Tarea) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, existente := range r.lista {
		if existente.ID == t.ID && existente.HistoriaID == t.HistoriaID {
			r.lista[i] = t
			return nil
		}
	}
	return effort.ErrTareaNoEncontrada
}

// EliminarDeHistoria borra todas las tareas de la historia (HU-06, RN9).
func (r *Tareas) EliminarDeHistoria(historiaID int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lista = slices.DeleteFunc(r.lista, func(t effort.Tarea) bool { return t.HistoriaID == historiaID })
}
