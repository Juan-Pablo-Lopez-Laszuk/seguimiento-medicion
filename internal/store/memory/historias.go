package memory

import (
	"context"
	"slices"
	"sync"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
)

// Dependiente es un repositorio con datos que cuelgan de una historia (sus criterios, sus tareas). Al eliminar la
// historia se le pide que borre lo suyo (HU-06, RN9): en PostgreSQL lo hace la misma transacción.
type Dependiente interface {
	// EliminarDeHistoria borra todo lo que el repositorio tiene de esa historia.
	EliminarDeHistoria(historiaID int64)
}

// Historias guarda las historias de todos los proyectos en una lista en memoria.
type Historias struct {
	mu           sync.Mutex
	lista        []backlog.Historia
	ultimoID     int64
	dependientes []Dependiente
}

// NuevoHistorias crea un repositorio vacío. Los dependientes, si hay, son los repositorios de lo que cuelga de una
// historia: eliminarla borra también lo suyo en cada uno.
func NuevoHistorias(dependientes ...Dependiente) *Historias {
	return &Historias{dependientes: dependientes}
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

// ActualizarDatos guarda el título, la descripción y la prioridad de la historia (HU-06) sin tocar el número, el estado
// ni los story points, o devuelve backlog.ErrHistoriaNoEncontrada si no existe.
func (r *Historias) ActualizarDatos(_ context.Context, h backlog.Historia) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, existente := range r.lista {
		if existente.ID == h.ID {
			r.lista[i].Titulo, r.lista[i].Descripcion, r.lista[i].Prioridad = h.Titulo, h.Descripcion, h.Prioridad
			return nil
		}
	}
	return backlog.ErrHistoriaNoEncontrada
}

// Eliminar borra la historia (HU-06) y lo que cuelga de ella en los dependientes, o devuelve
// backlog.ErrHistoriaNoEncontrada si no existe. Las demás historias conservan su número.
func (r *Historias) Eliminar(_ context.Context, historiaID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, h := range r.lista {
		if h.ID == historiaID {
			r.lista = slices.Delete(r.lista, i, i+1)
			for _, d := range r.dependientes {
				d.EliminarDeHistoria(historiaID)
			}
			return nil
		}
	}
	return backlog.ErrHistoriaNoEncontrada
}

// EnSprintCerrado dice si la historia es de un sprint cerrado (HU-06, RN6). En memoria todavía no se puede asignar una
// historia a un sprint (HU-12), así que nunca lo es.
func (r *Historias) EnSprintCerrado(_ context.Context, _ int64) (bool, error) {
	return false, nil
}
