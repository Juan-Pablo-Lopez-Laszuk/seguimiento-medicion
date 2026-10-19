package memory

import (
	"context"
	"sync"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/defect"
)

// Defectos guarda los defectos de todos los proyectos en una lista en memoria (HU-27).
type Defectos struct {
	mu       sync.Mutex
	lista    []defect.Defecto
	ultimoID int64
}

// NuevoDefectos crea un repositorio vacío.
func NuevoDefectos() *Defectos {
	return &Defectos{}
}

// Guardar le asigna el próximo ID al defecto, lo guarda y lo devuelve con el ID.
func (r *Defectos) Guardar(_ context.Context, d defect.Defecto) (defect.Defecto, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ultimoID++
	d.ID = r.ultimoID
	r.lista = append(r.lista, d)
	return d, nil
}

// ListarPorProyecto devuelve los defectos del proyecto, el más nuevo primero.
func (r *Defectos) ListarPorProyecto(_ context.Context, proyectoID int64) ([]defect.Defecto, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var lista []defect.Defecto
	for i := len(r.lista) - 1; i >= 0; i-- {
		if r.lista[i].ProyectoID == proyectoID {
			lista = append(lista, r.lista[i])
		}
	}
	return lista, nil
}
