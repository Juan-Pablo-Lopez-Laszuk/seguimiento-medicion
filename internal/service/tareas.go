package service

import (
	"context"
	"fmt"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
)

// ListadorIntegrantes es lo que los casos de uso de tareas necesitan del repositorio de integrantes: los del proyecto,
// para validar y mostrar el responsable de cada tarea.
type ListadorIntegrantes interface {
	// ListarPorProyecto devuelve los integrantes del proyecto (activos y dados de baja) en el orden en que se
	// registraron.
	ListarPorProyecto(ctx context.Context, proyectoID int64) ([]member.Integrante, error)
}

// RepositorioTareas es lo que el caso de uso necesita de la base de datos. Lo implementan internal/store/memory
// (tests) e internal/store/postgres (Supabase).
type RepositorioTareas interface {
	// ListarPorHistoria devuelve las tareas de la historia en el orden en que se cargaron.
	ListarPorHistoria(ctx context.Context, historiaID int64) ([]effort.Tarea, error)
	// Guardar guarda una tarea nueva y la devuelve con su ID.
	Guardar(ctx context.Context, t effort.Tarea) (effort.Tarea, error)
	// BuscarPorID devuelve la tarea dentro de su historia, o effort.ErrTareaNoEncontrada.
	BuscarPorID(ctx context.Context, historiaID, id int64) (effort.Tarea, error)
	// Actualizar guarda las horas y si está completada, o devuelve effort.ErrTareaNoEncontrada.
	Actualizar(ctx context.Context, t effort.Tarea) error
}

// Tareas agrupa los casos de uso de las tareas de una historia (HU-25).
type Tareas struct {
	proyectos   BuscadorProyectos
	historias   BuscadorHistorias
	integrantes ListadorIntegrantes
	repo        RepositorioTareas
}

// NuevoTareas arma el caso de uso.
func NuevoTareas(proyectos BuscadorProyectos, historias BuscadorHistorias, integrantes ListadorIntegrantes, repo RepositorioTareas) *Tareas {
	return &Tareas{proyectos: proyectos, historias: historias, integrantes: integrantes, repo: repo}
}

// VistaTareas es lo que muestra la página de tareas: el proyecto, la historia, sus tareas en orden, las horas
// estimadas de la historia (la suma de sus tareas, CA-23.2) y los integrantes del proyecto.
type VistaTareas struct {
	Proyecto       project.Proyecto
	Historia       backlog.Historia
	Tareas         []effort.Tarea
	HorasEstimadas float64
	// Integrantes son todos los del proyecto, también los dados de baja, para mostrar el responsable de cada tarea.
	Integrantes []member.Integrante
}

// Activos devuelve los integrantes que pueden ser responsables de una tarea nueva (RN2), en el orden en que se
// registraron.
func (v VistaTareas) Activos() []member.Integrante {
	var activos []member.Integrante
	for _, i := range v.Integrantes {
		if i.Activo {
			activos = append(activos, i)
		}
	}
	return activos
}

// NombreDe devuelve el nombre del integrante con ese ID, aunque esté dado de baja, o "" si no es del proyecto.
func (v VistaTareas) NombreDe(id int64) string {
	for _, i := range v.Integrantes {
		if i.ID == id {
			return i.Nombre
		}
	}
	return ""
}

// Ver devuelve la historia HU-N del proyecto con sus tareas, la suma de sus horas y los integrantes. Si el proyecto
// no existe, el error es project.ErrNoEncontrado; si la historia no existe en el proyecto, backlog.ErrHistoriaNoEncontrada.
func (s *Tareas) Ver(ctx context.Context, proyectoID int64, numero int) (VistaTareas, error) {
	p, h, err := s.buscar(ctx, proyectoID, numero)
	if err != nil {
		return VistaTareas{}, err
	}
	tareas, err := s.repo.ListarPorHistoria(ctx, h.ID)
	if err != nil {
		return VistaTareas{}, fmt.Errorf("listar las tareas de la historia %d: %w", h.ID, err)
	}
	horas, err := effort.HorasEstimadasDeHistoria(0, tareas)
	if err != nil {
		return VistaTareas{}, fmt.Errorf("sumar las horas de la historia %d: %w", h.ID, err)
	}
	integrantes, err := s.integrantes.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return VistaTareas{}, fmt.Errorf("listar los integrantes del proyecto %d: %w", proyectoID, err)
	}
	return VistaTareas{Proyecto: p, Historia: h, Tareas: tareas, HorasEstimadas: horas, Integrantes: integrantes}, nil
}

// Agregar valida los datos (HU-25) contra los integrantes del proyecto y guarda la tarea nueva en la historia HU-N.
// Los datos inválidos se devuelven en domain.ErroresValidacion; si el proyecto o la historia no existen,
// project.ErrNoEncontrado o backlog.ErrHistoriaNoEncontrada.
func (s *Tareas) Agregar(ctx context.Context, proyectoID int64, numero int, d effort.DatosTarea) (effort.Tarea, error) {
	_, h, err := s.buscar(ctx, proyectoID, numero)
	if err != nil {
		return effort.Tarea{}, err
	}
	integrantes, err := s.integrantes.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return effort.Tarea{}, fmt.Errorf("listar los integrantes del proyecto %d: %w", proyectoID, err)
	}
	nueva, err := effort.NuevaTarea(d, h.ID, integrantes)
	if err != nil {
		return effort.Tarea{}, err
	}
	guardada, err := s.repo.Guardar(ctx, nueva)
	if err != nil {
		return effort.Tarea{}, fmt.Errorf("guardar la tarea: %w", err)
	}
	return guardada, nil
}

// CambiarEstimacion cambia las horas estimadas de la tarea (CA-25.3). En una tarea completada el error es
// effort.ErrTareaCompletada; con horas inválidas, effort.ErrHorasInvalidas; si la tarea no es de esa historia o no
// existe, effort.ErrTareaNoEncontrada.
func (s *Tareas) CambiarEstimacion(ctx context.Context, proyectoID int64, numero int, tareaID int64, horas string) (effort.Tarea, error) {
	return s.cambiar(ctx, proyectoID, numero, tareaID, func(t effort.Tarea) (effort.Tarea, error) {
		return t.CambiarEstimacion(horas)
	})
}

// Marcar deja la tarea completada o pendiente, según completada (RN7). Si la tarea no es de esa historia o no existe,
// el error es effort.ErrTareaNoEncontrada.
func (s *Tareas) Marcar(ctx context.Context, proyectoID int64, numero int, tareaID int64, completada bool) (effort.Tarea, error) {
	return s.cambiar(ctx, proyectoID, numero, tareaID, func(t effort.Tarea) (effort.Tarea, error) {
		return t.Marcar(completada), nil
	})
}

// cambiar busca la tarea dentro de la historia HU-N, le aplica el cambio del dominio y la guarda.
func (s *Tareas) cambiar(ctx context.Context, proyectoID int64, numero int, tareaID int64, cambio func(effort.Tarea) (effort.Tarea, error)) (effort.Tarea, error) {
	_, h, err := s.buscar(ctx, proyectoID, numero)
	if err != nil {
		return effort.Tarea{}, err
	}
	t, err := s.repo.BuscarPorID(ctx, h.ID, tareaID)
	if err != nil {
		return effort.Tarea{}, fmt.Errorf("buscar la tarea %d: %w", tareaID, err)
	}
	cambiada, err := cambio(t)
	if err != nil {
		return effort.Tarea{}, err
	}
	if err := s.repo.Actualizar(ctx, cambiada); err != nil {
		return effort.Tarea{}, fmt.Errorf("guardar la tarea %d: %w", tareaID, err)
	}
	return cambiada, nil
}

// buscar trae el proyecto y su historia HU-N, o el error de "no encontrado" que corresponda.
func (s *Tareas) buscar(ctx context.Context, proyectoID int64, numero int) (project.Proyecto, backlog.Historia, error) {
	p, err := s.proyectos.BuscarPorID(ctx, proyectoID)
	if err != nil {
		return project.Proyecto{}, backlog.Historia{}, fmt.Errorf("buscar el proyecto %d: %w", proyectoID, err)
	}
	h, err := s.historias.BuscarPorNumero(ctx, proyectoID, numero)
	if err != nil {
		return project.Proyecto{}, backlog.Historia{}, fmt.Errorf("buscar la historia HU-%d del proyecto %d: %w", numero, proyectoID, err)
	}
	return p, h, nil
}
