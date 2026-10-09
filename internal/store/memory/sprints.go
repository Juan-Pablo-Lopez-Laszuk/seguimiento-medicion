package memory

import (
	"context"
	"sync"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
)

// Sprints guarda los sprints de todos los proyectos en una lista en memoria.
type Sprints struct {
	mu       sync.Mutex
	lista    []sprint.Sprint
	ultimoID int64
}

// NuevoSprints crea un repositorio vacío.
func NuevoSprints() *Sprints {
	return &Sprints{}
}

// Guardar le asigna el próximo ID al sprint, lo guarda y lo devuelve con el ID.
func (r *Sprints) Guardar(_ context.Context, s sprint.Sprint) (sprint.Sprint, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ultimoID++
	s.ID = r.ultimoID
	r.lista = append(r.lista, s)
	return s, nil
}

// ListarPorProyecto devuelve una copia de los sprints del proyecto, en el orden en que se crearon.
func (r *Sprints) ListarPorProyecto(_ context.Context, proyectoID int64) ([]sprint.Sprint, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var lista []sprint.Sprint
	for _, s := range r.lista {
		if s.ProyectoID == proyectoID {
			lista = append(lista, s)
		}
	}
	return lista, nil
}
