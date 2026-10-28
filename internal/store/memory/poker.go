package memory

import (
	"context"
	"sync"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
)

// Poker guarda las sesiones de Planning Poker y sus rondas en listas en memoria (épica E4). Cada historia de la épica
// le agrega sus métodos en su propio archivo (spec de HU-17, "Modelo común de Planning Poker").
type Poker struct {
	mu           sync.Mutex
	sesiones     []poker.Sesion
	rondas       []poker.Ronda
	ultimaSesion int64
	ultimaRonda  int64
}

// NuevoPoker crea un repositorio vacío.
func NuevoPoker() *Poker {
	return &Poker{}
}

// Abrir guarda una sesión Abierta de la historia con su Ronda 1 (HU-17, RN3). Si la historia ya tiene una sesión
// abierta devuelve poker.ErrSesionYaAbierta y no guarda nada (CA-17.2).
func (r *Poker) Abrir(_ context.Context, historiaID int64) (poker.Sesion, poker.Ronda, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.abierta(historiaID); ok {
		return poker.Sesion{}, poker.Ronda{}, poker.ErrSesionYaAbierta
	}
	r.ultimaSesion++
	r.ultimaRonda++
	s := poker.Sesion{ID: r.ultimaSesion, HistoriaID: historiaID, Estado: poker.EstadoAbierta}
	ronda := poker.Ronda{ID: r.ultimaRonda, SesionID: s.ID, Numero: 1}
	r.sesiones = append(r.sesiones, s)
	r.rondas = append(r.rondas, ronda)
	return s, ronda, nil
}

// SesionAbierta devuelve la sesión abierta de la historia y su ronda actual (la de número más alto), o
// poker.ErrSesionNoEncontrada.
func (r *Poker) SesionAbierta(_ context.Context, historiaID int64) (poker.Sesion, poker.Ronda, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s, ok := r.abierta(historiaID)
	if !ok {
		return poker.Sesion{}, poker.Ronda{}, poker.ErrSesionNoEncontrada
	}
	var actual poker.Ronda
	for _, ronda := range r.rondas {
		if ronda.SesionID == s.ID && ronda.Numero > actual.Numero {
			actual = ronda
		}
	}
	return s, actual, nil
}

// TieneDeHistoria dice si la historia tiene alguna sesión, abierta o cerrada: así no se elimina (HU-06, RN10), como
// hace la clave foránea en PostgreSQL.
func (r *Poker) TieneDeHistoria(historiaID int64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, s := range r.sesiones {
		if s.HistoriaID == historiaID {
			return true
		}
	}
	return false
}

// abierta busca la sesión abierta de la historia. Quien la llama tiene el candado.
func (r *Poker) abierta(historiaID int64) (poker.Sesion, bool) {
	for _, s := range r.sesiones {
		if s.HistoriaID == historiaID && s.Estado == poker.EstadoAbierta {
			return s, true
		}
	}
	return poker.Sesion{}, false
}
