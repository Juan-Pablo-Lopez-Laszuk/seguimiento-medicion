package memory

import (
	"context"
	"errors"
	"strings"
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
	i.AuthUserID = r.lista[pos].AuthUserID // el vínculo con el usuario no se cambia desde acá
	r.lista[pos] = i
	return nil
}

// VincularPorEmail vincula al usuario con todos los integrantes de cualquier proyecto que tienen ese email y todavía no
// tienen un usuario. Devuelve cuántos vinculó (TEC-05, RN7).
func (r *Integrantes) VincularPorEmail(_ context.Context, email, usuarioID string) (int, error) {
	if usuarioID == "" {
		return 0, errors.New("vincular integrantes: falta el usuario")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	vinculados := 0
	for pos := range r.lista {
		if i := &r.lista[pos]; i.AuthUserID == "" && strings.EqualFold(i.Email, email) {
			i.AuthUserID = usuarioID
			vinculados++
		}
	}
	return vinculados, nil
}

// BuscarPorUsuario devuelve el integrante del proyecto vinculado con el usuario, o member.ErrNoEncontrado.
func (r *Integrantes) BuscarPorUsuario(_ context.Context, proyectoID int64, usuarioID string) (member.Integrante, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if usuarioID != "" {
		for _, i := range r.lista {
			if i.ProyectoID == proyectoID && i.AuthUserID == usuarioID {
				return i, nil
			}
		}
	}
	return member.Integrante{}, member.ErrNoEncontrado
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
