package memory

import (
	"context"
	"strings"
	"sync"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
)

// Proyectos guarda los proyectos en una lista en memoria. El mutex evita problemas si llegan
// varios pedidos a la vez.
type Proyectos struct {
	mu       sync.Mutex
	lista    []project.Proyecto
	ultimoID int64
}

// NuevoProyectos crea un repositorio vacío.
func NuevoProyectos() *Proyectos {
	return &Proyectos{}
}

// Guardar le asigna el próximo ID al proyecto, lo guarda y lo devuelve con el ID.
func (r *Proyectos) Guardar(_ context.Context, p project.Proyecto) (project.Proyecto, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ultimoID++
	p.ID = r.ultimoID
	r.lista = append(r.lista, p)
	return p, nil
}

// Listar devuelve una copia de los proyectos en el orden en que se crearon.
func (r *Proyectos) Listar(_ context.Context) ([]project.Proyecto, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]project.Proyecto(nil), r.lista...), nil
}

// ExisteNombre informa si ya hay un proyecto con ese nombre, sin distinguir mayúsculas (RN3).
func (r *Proyectos) ExisteNombre(_ context.Context, nombre string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range r.lista {
		if strings.EqualFold(p.Nombre, nombre) {
			return true, nil
		}
	}
	return false, nil
}

// ExisteOtroConNombre informa si otro proyecto (no el de exceptoID) ya tiene ese nombre, sin distinguir
// mayúsculas (HU-02 RN2).
func (r *Proyectos) ExisteOtroConNombre(_ context.Context, nombre string, exceptoID int64) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range r.lista {
		if p.ID != exceptoID && strings.EqualFold(p.Nombre, nombre) {
			return true, nil
		}
	}
	return false, nil
}

// Actualizar reemplaza el proyecto que tiene el mismo ID, o devuelve project.ErrNoEncontrado si no existe.
func (r *Proyectos) Actualizar(_ context.Context, p project.Proyecto) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.lista {
		if r.lista[i].ID == p.ID {
			r.lista[i] = p
			return nil
		}
	}
	return project.ErrNoEncontrado
}

// BuscarPorID devuelve el proyecto con ese ID, o project.ErrNoEncontrado si no existe.
func (r *Proyectos) BuscarPorID(_ context.Context, id int64) (project.Proyecto, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range r.lista {
		if p.ID == id {
			return p, nil
		}
	}
	return project.Proyecto{}, project.ErrNoEncontrado
}
