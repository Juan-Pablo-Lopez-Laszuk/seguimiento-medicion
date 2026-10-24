package memory

import (
	"context"
	"slices"
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

// TieneDeHistoria dice si hay algún defecto de esa historia (HU-06, RN10: con defectos no se elimina). Los defectos sin
// historia no son de ninguna.
func (r *Defectos) TieneDeHistoria(historiaID int64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return historiaID != 0 && slices.ContainsFunc(r.lista, func(d defect.Defecto) bool { return d.HistoriaID == historiaID })
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

// BuscarPorID devuelve el defecto dentro de su proyecto, o defect.ErrDefectoNoEncontrado.
func (r *Defectos) BuscarPorID(_ context.Context, proyectoID, id int64) (defect.Defecto, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, d := range r.lista {
		if d.ID == id && d.ProyectoID == proyectoID {
			return d, nil
		}
	}
	return defect.Defecto{}, defect.ErrDefectoNoEncontrado
}

// Actualizar guarda el estado y el sprint de resolución de un defecto que ya existe (HU-28), o devuelve
// defect.ErrDefectoNoEncontrado.
func (r *Defectos) Actualizar(_ context.Context, d defect.Defecto) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, existente := range r.lista {
		if existente.ID == d.ID && existente.ProyectoID == d.ProyectoID {
			r.lista[i].Estado, r.lista[i].SprintResolucionID = d.Estado, d.SprintResolucionID
			return nil
		}
	}
	return defect.ErrDefectoNoEncontrado
}

// ListarFiltrados devuelve los defectos del proyecto que cumplen el filtro, el más nuevo primero (HU-29).
func (r *Defectos) ListarFiltrados(ctx context.Context, proyectoID int64, f defect.Filtro) ([]defect.Defecto, error) {
	todos, err := r.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return nil, err
	}
	var lista []defect.Defecto
	for _, d := range todos {
		if f.Cumple(d) {
			lista = append(lista, d)
		}
	}
	return lista, nil
}
