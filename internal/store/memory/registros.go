package memory

import (
	"context"
	"slices"
	"sync"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
)

// Registros guarda los registros de esfuerzo en una lista en memoria (HU-24).
type Registros struct {
	mu       sync.Mutex
	lista    []effort.Registro
	ultimoID int64
}

// NuevoRegistros crea un repositorio vacío.
func NuevoRegistros() *Registros {
	return &Registros{}
}

// Guardar le asigna el próximo ID al registro, lo guarda y lo devuelve con el ID.
func (r *Registros) Guardar(_ context.Context, reg effort.Registro) (effort.Registro, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ultimoID++
	reg.ID = r.ultimoID
	r.lista = append(r.lista, reg)
	return reg, nil
}

// ListarPorHistoria devuelve los registros de la historia, la fecha más reciente primero.
func (r *Registros) ListarPorHistoria(_ context.Context, historiaID int64) ([]effort.Registro, error) {
	return r.filtrar(func(reg effort.Registro) bool { return reg.HistoriaID == historiaID }), nil
}

// ListarPorIntegrante devuelve los registros del integrante, la fecha más reciente primero.
func (r *Registros) ListarPorIntegrante(_ context.Context, integranteID int64) ([]effort.Registro, error) {
	return r.filtrar(func(reg effort.Registro) bool { return reg.IntegranteID == integranteID }), nil
}

func (r *Registros) filtrar(incluir func(effort.Registro) bool) []effort.Registro {
	r.mu.Lock()
	defer r.mu.Unlock()
	var lista []effort.Registro
	for _, reg := range r.lista {
		if incluir(reg) {
			lista = append(lista, reg)
		}
	}
	slices.SortStableFunc(lista, func(a, b effort.Registro) int {
		if c := b.Fecha.Compare(a.Fecha); c != 0 {
			return c
		}
		return int(b.ID - a.ID)
	})
	return lista
}

// ListarFiltrados devuelve los registros de esas historias que cumplen el filtro, la fecha más reciente primero
// (HU-26).
func (r *Registros) ListarFiltrados(_ context.Context, historiaIDs []int64, f effort.Filtro) ([]effort.Registro, error) {
	return r.filtrar(func(reg effort.Registro) bool { return slices.Contains(historiaIDs, reg.HistoriaID) && f.Cumple(reg) }), nil
}

// BuscarPorID devuelve el registro con ese ID, o effort.ErrRegistroNoEncontrado (HU-26).
func (r *Registros) BuscarPorID(_ context.Context, id int64) (effort.Registro, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, reg := range r.lista {
		if reg.ID == id {
			return reg, nil
		}
	}
	return effort.Registro{}, effort.ErrRegistroNoEncontrado
}

// Actualizar guarda el registro corregido en lugar del que tiene su ID, o devuelve effort.ErrRegistroNoEncontrado
// (HU-26).
func (r *Registros) Actualizar(_ context.Context, reg effort.Registro) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.lista {
		if r.lista[i].ID == reg.ID {
			r.lista[i] = reg
			return nil
		}
	}
	return effort.ErrRegistroNoEncontrado
}
