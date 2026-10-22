package memory

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
)

// Dependiente es un repositorio con datos que cuelgan de una historia (sus criterios, sus tareas). Al eliminar la
// historia se le pide que borre lo suyo (HU-06, RN9): en PostgreSQL lo hace la misma transacción.
type Dependiente interface {
	// EliminarDeHistoria borra todo lo que el repositorio tiene de esa historia.
	EliminarDeHistoria(historiaID int64)
}

// ConRegistros es un repositorio con registros que apuntan a una historia (los de esfuerzo, los defectos). Mientras tenga
// alguno, la historia no se elimina (HU-06, RN10): en PostgreSQL lo impide la clave foránea.
type ConRegistros interface {
	// TieneDeHistoria dice si el repositorio tiene algún registro de esa historia.
	TieneDeHistoria(historiaID int64) bool
}

// Historias guarda las historias de todos los proyectos en una lista en memoria.
type Historias struct {
	mu           sync.Mutex
	lista        []backlog.Historia
	ultimoID     int64
	dependientes []Dependiente
	asociados    []ConRegistros
	cierres      *Cierres // lo conecta NuevoCierres (HU-14), para EnSprintCerrado
	// estimaciones es el historial de re-estimaciones de cada historia (HU-16b), de la más antigua a la más reciente.
	estimaciones map[int64][]backlog.Reestimacion
	ahora        func() time.Time
}

// NuevoHistorias crea un repositorio vacío. Los dependientes, si hay, son los repositorios de lo que cuelga de una
// historia: eliminarla borra también lo suyo en cada uno.
func NuevoHistorias(dependientes ...Dependiente) *Historias {
	return &Historias{dependientes: dependientes, estimaciones: map[int64][]backlog.Reestimacion{}, ahora: time.Now}
}

// ConRegistrosAsociados le dice al repositorio qué otros repositorios tienen registros de una historia (esfuerzo,
// defectos): una historia con registros ahí no se elimina y Eliminar devuelve backlog.ErrHistoriaConRegistros, como hace
// PostgreSQL con sus claves foráneas (HU-06, RN10). Sin llamarlo, no hay nada que lo impida.
func (r *Historias) ConRegistrosAsociados(asociados ...ConRegistros) *Historias {
	r.asociados = asociados
	return r
}

// ConReloj cambia el reloj con el que se anota la hora de cada re-estimación (HU-16b): en los tests, uno fijo. Sin
// llamarlo se usa la hora actual.
func (r *Historias) ConReloj(ahora func() time.Time) *Historias {
	r.ahora = ahora
	return r
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

// copiarHistoria copia también los story points, el sprint y la fecha de completado, que son punteros, para que el repositorio y quien lo usa
// no compartan el mismo valor.
func copiarHistoria(h backlog.Historia) backlog.Historia {
	if h.StoryPoints != nil {
		puntos := *h.StoryPoints
		h.StoryPoints = &puntos
	}
	h.SprintID = copiarID(h.SprintID)
	h.CompletadaEn = copiarFecha(h.CompletadaEn)
	return h
}

func copiarFecha(f *time.Time) *time.Time {
	if f == nil {
		return nil
	}
	copia := *f
	return &copia
}

func copiarID(id *int64) *int64 {
	if id == nil {
		return nil
	}
	copia := *id
	return &copia
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

// ActualizarSprint guarda juntos el estado y el sprint de la historia (nil = sin sprint) sin tocar los demás datos, o
// devuelve backlog.ErrHistoriaNoEncontrada si no existe. Lo usa HU-12 al asignar y al quitar.
func (r *Historias) ActualizarSprint(_ context.Context, historiaID int64, estado backlog.Estado, sprintID *int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, h := range r.lista {
		if h.ID == historiaID {
			r.lista[i].Estado, r.lista[i].SprintID = estado, copiarID(sprintID)
			return nil
		}
	}
	return backlog.ErrHistoriaNoEncontrada
}

// ActualizarCompletada guarda juntos el estado y la fecha de completado de la historia sin tocar los demás datos, o
// devuelve backlog.ErrHistoriaNoEncontrada si no existe. Lo usa HU-13 al completar.
func (r *Historias) ActualizarCompletada(_ context.Context, historiaID int64, estado backlog.Estado, completadaEn *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, h := range r.lista {
		if h.ID == historiaID {
			r.lista[i].Estado, r.lista[i].CompletadaEn = estado, copiarFecha(completadaEn)
			return nil
		}
	}
	return backlog.ErrHistoriaNoEncontrada
}

// ActualizarStoryPoints guarda los story points de la historia sin tocar los demás datos, o devuelve
// backlog.ErrHistoriaNoEncontrada si no existe. Es el único lugar donde se escribe la estimación (HU-16): si la historia
// ya tenía una y el valor es distinto, anota el cambio en el historial (HU-16b, CA-16.3).
func (r *Historias) ActualizarStoryPoints(_ context.Context, historiaID int64, puntos int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, h := range r.lista {
		if h.ID == historiaID {
			if cambio, anota := backlog.NuevaReestimacion(h.StoryPoints, puntos, r.ahora()); anota {
				r.estimaciones[historiaID] = append(r.estimaciones[historiaID], cambio)
			}
			r.lista[i].StoryPoints = &puntos
			return nil
		}
	}
	return backlog.ErrHistoriaNoEncontrada
}

// ListarEstimaciones devuelve una copia del historial de re-estimaciones de la historia, la más reciente primero
// (HU-16b). Una historia que nunca se re-estimó, o que no existe, tiene el historial vacío.
func (r *Historias) ListarEstimaciones(_ context.Context, historiaID int64) ([]backlog.Reestimacion, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	guardado := r.estimaciones[historiaID]
	lista := make([]backlog.Reestimacion, len(guardado))
	for i, e := range guardado {
		lista[len(guardado)-1-i] = e
	}
	return lista, nil
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
// backlog.ErrHistoriaNoEncontrada si no existe y backlog.ErrHistoriaConRegistros, sin borrar nada, si tiene registros de
// esfuerzo o defectos (RN10). Las demás historias conservan su número.
func (r *Historias) Eliminar(_ context.Context, historiaID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, h := range r.lista {
		if h.ID == historiaID {
			for _, asociado := range r.asociados {
				if asociado.TieneDeHistoria(historiaID) {
					return backlog.ErrHistoriaConRegistros
				}
			}
			r.lista = slices.Delete(r.lista, i, i+1)
			delete(r.estimaciones, historiaID)
			for _, d := range r.dependientes {
				d.EliminarDeHistoria(historiaID)
			}
			return nil
		}
	}
	return backlog.ErrHistoriaNoEncontrada
}

// EnSprintCerrado dice si la historia es de un sprint cerrado (HU-06, RN6): si su sprint está cerrado o si aparece en
// la foto del cierre de uno (HU-14), igual que en PostgreSQL. Sin el repositorio de cierres, ningún sprint se cerró.
func (r *Historias) EnSprintCerrado(_ context.Context, historiaID int64) (bool, error) {
	r.mu.Lock()
	i := slices.IndexFunc(r.lista, func(h backlog.Historia) bool { return h.ID == historiaID })
	var h backlog.Historia
	if i >= 0 {
		h = copiarHistoria(r.lista[i])
	}
	cierres := r.cierres
	r.mu.Unlock() // el cierre toma los candados en otro orden: se consulta sin tener tomado este

	if i < 0 || cierres == nil {
		return false, nil
	}
	return cierres.esDeUnSprintCerrado(h), nil
}
