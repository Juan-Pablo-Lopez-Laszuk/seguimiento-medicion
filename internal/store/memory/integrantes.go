package memory

import (
	"context"
	"sync"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
)

// Integrantes guarda los integrantes de todos los proyectos en una lista en memoria.
type Integrantes struct {
	mu       sync.Mutex
	lista    []member.Integrante
	ultimoID int64
}

// NuevoIntegrantes crea un repositorio vacío.
func NuevoIntegrantes() *Integrantes {
	return &Integrantes{}
}

// Guardar le asigna el próximo ID al integrante, lo guarda y lo devuelve con el ID.
func (r *Integrantes) Guardar(_ context.Context, i member.Integrante) (member.Integrante, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ultimoID++
	i.ID = r.ultimoID
	r.lista = append(r.lista, i)
	return i, nil
}

// ListarPorProyecto devuelve una copia de los integrantes del proyecto, en el orden en que se registraron.
func (r *Integrantes) ListarPorProyecto(_ context.Context, proyectoID int64) ([]member.Integrante, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var lista []member.Integrante
	for _, i := range r.lista {
		if i.ProyectoID == proyectoID {
			lista = append(lista, i)
		}
	}
	return lista, nil
}

// BuscarPorID devuelve el integrante con ese ID dentro del proyecto, o member.ErrNoEncontrado.
func (r *Integrantes) BuscarPorID(_ context.Context, proyectoID, id int64) (member.Integrante, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if pos := r.posicion(proyectoID, id); pos >= 0 {
		return r.lista[pos], nil
	}
	return member.Integrante{}, member.ErrNoEncontrado
}

// Actualizar reemplaza el integrante guardado con el mismo ID y proyecto, o devuelve member.ErrNoEncontrado.
func (r *Integrantes) Actualizar(_ context.Context, i member.Integrante) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	pos := r.posicion(i.ProyectoID, i.ID)
	if pos < 0 {
		return member.ErrNoEncontrado
	}
	r.lista[pos] = i
	return nil
}

// posicion devuelve dónde está el integrante en la lista, o -1. Se llama con el mutex tomado.
func (r *Integrantes) posicion(proyectoID, id int64) int {
	for pos, i := range r.lista {
		if i.ID == id && i.ProyectoID == proyectoID {
			return pos
		}
	}
	return -1
}
