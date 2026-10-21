package service

import (
	"context"
	"fmt"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
)

// BuscadorHistorias es lo que los casos de uso de criterios necesitan del repositorio de historias.
type BuscadorHistorias interface {
	// BuscarPorNumero devuelve la historia HU-N del proyecto, o backlog.ErrHistoriaNoEncontrada si no existe.
	BuscarPorNumero(ctx context.Context, proyectoID int64, numero int) (backlog.Historia, error)
}

// RepositorioCriterios es lo que el caso de uso necesita de la base de datos. Lo implementan
// internal/store/memory (tests) e internal/store/postgres (Supabase).
type RepositorioCriterios interface {
	// ListarPorHistoria devuelve los criterios de la historia en el orden en que se cargaron.
	ListarPorHistoria(ctx context.Context, historiaID int64) ([]backlog.Criterio, error)
	// ListarPorHistorias devuelve los criterios de todas esas historias juntos, en el orden en que se cargaron.
	ListarPorHistorias(ctx context.Context, historiaIDs []int64) ([]backlog.Criterio, error)
	// Guardar guarda un criterio nuevo y lo devuelve con su ID.
	Guardar(ctx context.Context, c backlog.Criterio) (backlog.Criterio, error)
	// BuscarPorID devuelve el criterio dentro de su historia, o backlog.ErrCriterioNoEncontrado.
	BuscarPorID(ctx context.Context, historiaID, id int64) (backlog.Criterio, error)
	// Actualizar guarda los cambios de un criterio que ya existe, o devuelve backlog.ErrCriterioNoEncontrado.
	Actualizar(ctx context.Context, c backlog.Criterio) error
}

// ListadorEstimaciones es lo que la página de la historia necesita para mostrar su historial de estimaciones (HU-16b).
// Lo implementan internal/store/memory (tests) e internal/store/postgres (Supabase).
type ListadorEstimaciones interface {
	// ListarEstimaciones devuelve el historial de re-estimaciones de la historia, la más reciente primero.
	ListarEstimaciones(ctx context.Context, historiaID int64) ([]backlog.Reestimacion, error)
}

// Criterios agrupa los casos de uso de los criterios de aceptación de una historia (HU-07).
type Criterios struct {
	proyectos    BuscadorProyectos
	historias    BuscadorHistorias
	repo         RepositorioCriterios
	estimaciones ListadorEstimaciones // opcional (HU-16b): sin él, la vista no trae el historial
}

// NuevoCriterios arma el caso de uso.
func NuevoCriterios(proyectos BuscadorProyectos, historias BuscadorHistorias, repo RepositorioCriterios) *Criterios {
	return &Criterios{proyectos: proyectos, historias: historias, repo: repo}
}

// ConHistorial conecta el historial de estimaciones (HU-16b): desde entonces, la vista de la historia lo trae. Se hace
// aparte de NuevoCriterios para no cambiar a quienes arman este caso de uso sin historial.
func (s *Criterios) ConHistorial(estimaciones ListadorEstimaciones) *Criterios {
	s.estimaciones = estimaciones
	return s
}

// VistaHistoria es lo que muestra la página de una historia: el proyecto, la historia, sus criterios en orden, su
// progreso y su historial de estimaciones.
type VistaHistoria struct {
	Proyecto  project.Proyecto
	Historia  backlog.Historia
	Criterios []backlog.Criterio
	Progreso  backlog.Progreso
	Historial []backlog.Reestimacion // los cambios de la estimación, el más reciente primero (HU-16b)
}

// Ver devuelve la historia HU-N del proyecto con sus criterios y su progreso. Si el proyecto no existe, el error es
// project.ErrNoEncontrado; si la historia no existe en el proyecto, backlog.ErrHistoriaNoEncontrada.
func (s *Criterios) Ver(ctx context.Context, proyectoID int64, numero int) (VistaHistoria, error) {
	p, h, err := s.buscar(ctx, proyectoID, numero)
	if err != nil {
		return VistaHistoria{}, err
	}
	lista, err := s.repo.ListarPorHistoria(ctx, h.ID)
	if err != nil {
		return VistaHistoria{}, fmt.Errorf("listar los criterios de la historia %d: %w", h.ID, err)
	}
	v := VistaHistoria{Proyecto: p, Historia: h, Criterios: lista, Progreso: backlog.ProgresoDe(lista)}
	if s.estimaciones != nil {
		if v.Historial, err = s.estimaciones.ListarEstimaciones(ctx, h.ID); err != nil {
			return VistaHistoria{}, fmt.Errorf("listar el historial de estimaciones de la historia %d: %w", h.ID, err)
		}
	}
	return v, nil
}

// Agregar valida la descripción (HU-07) y guarda el criterio nuevo en la historia. Una descripción inválida se
// devuelve en domain.ErroresValidacion; en una historia Hecha, backlog.ErrHistoriaHecha; si el proyecto o la historia
// no existen, project.ErrNoEncontrado o backlog.ErrHistoriaNoEncontrada.
func (s *Criterios) Agregar(ctx context.Context, proyectoID int64, numero int, descripcion string) (backlog.Criterio, error) {
	_, h, err := s.buscar(ctx, proyectoID, numero)
	if err != nil {
		return backlog.Criterio{}, err
	}
	nuevo, err := backlog.NuevoCriterio(descripcion, h)
	if err != nil {
		return backlog.Criterio{}, err
	}
	guardado, err := s.repo.Guardar(ctx, nuevo)
	if err != nil {
		return backlog.Criterio{}, fmt.Errorf("guardar el criterio: %w", err)
	}
	return guardado, nil
}

// Marcar deja el criterio cumplido o no cumplido, según cumplido (RN3). Si el criterio no es de esa historia o no
// existe, el error es backlog.ErrCriterioNoEncontrado; en una historia Hecha no se desmarca (backlog.ErrHistoriaHecha).
func (s *Criterios) Marcar(ctx context.Context, proyectoID int64, numero int, criterioID int64, cumplido bool) (backlog.Criterio, error) {
	_, h, err := s.buscar(ctx, proyectoID, numero)
	if err != nil {
		return backlog.Criterio{}, err
	}
	c, err := s.repo.BuscarPorID(ctx, h.ID, criterioID)
	if err != nil {
		return backlog.Criterio{}, fmt.Errorf("buscar el criterio %d: %w", criterioID, err)
	}
	marcado, err := c.Marcar(cumplido, h)
	if err != nil {
		return backlog.Criterio{}, err
	}
	if err := s.repo.Actualizar(ctx, marcado); err != nil {
		return backlog.Criterio{}, fmt.Errorf("guardar el criterio %d: %w", criterioID, err)
	}
	return marcado, nil
}

// Progresos devuelve el progreso de cada historia, por su ID, con una sola consulta de criterios. Las historias sin
// criterios figuran en 0/0.
func (s *Criterios) Progresos(ctx context.Context, historias []backlog.Historia) (map[int64]backlog.Progreso, error) {
	ids := make([]int64, 0, len(historias))
	for _, h := range historias {
		ids = append(ids, h.ID)
	}
	lista, err := s.repo.ListarPorHistorias(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("listar los criterios de las historias: %w", err)
	}
	porHistoria := make(map[int64][]backlog.Criterio, len(ids))
	for _, c := range lista {
		porHistoria[c.HistoriaID] = append(porHistoria[c.HistoriaID], c)
	}
	progresos := make(map[int64]backlog.Progreso, len(ids))
	for _, id := range ids {
		progresos[id] = backlog.ProgresoDe(porHistoria[id])
	}
	return progresos, nil
}

// buscar trae el proyecto y su historia HU-N, o el error de "no encontrado" que corresponda.
func (s *Criterios) buscar(ctx context.Context, proyectoID int64, numero int) (project.Proyecto, backlog.Historia, error) {
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
