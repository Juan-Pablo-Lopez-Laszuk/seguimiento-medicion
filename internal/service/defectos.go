package service

import (
	"context"
	"fmt"
	"slices"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/defect"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/metrics"
)

// Los sprints y las historias del proyecto se piden con ListadorSprints y ListadorHistorias (de HU-04).

// RepositorioDefectos es lo que el caso de uso necesita de la base de datos (HU-27).
type RepositorioDefectos interface {
	// Guardar guarda un defecto nuevo y lo devuelve con su ID.
	Guardar(ctx context.Context, d defect.Defecto) (defect.Defecto, error)
	// ListarPorProyecto devuelve los defectos del proyecto, el más nuevo primero.
	ListarPorProyecto(ctx context.Context, proyectoID int64) ([]defect.Defecto, error)
	// BuscarPorID devuelve el defecto dentro de su proyecto, o defect.ErrDefectoNoEncontrado (HU-28).
	BuscarPorID(ctx context.Context, proyectoID, id int64) (defect.Defecto, error)
	// Actualizar guarda el estado y el sprint de resolución, o devuelve defect.ErrDefectoNoEncontrado (HU-28).
	Actualizar(ctx context.Context, d defect.Defecto) error
}

// Defectos agrupa los casos de uso de los defectos de un proyecto (HU-27).
type Defectos struct {
	proyectos BuscadorProyectos
	sprints   ListadorSprints
	historias ListadorHistorias
	repo      RepositorioDefectos
}

// NuevoDefectos arma el caso de uso.
func NuevoDefectos(proyectos BuscadorProyectos, sprints ListadorSprints, historias ListadorHistorias, repo RepositorioDefectos) *Defectos {
	return &Defectos{proyectos: proyectos, sprints: sprints, historias: historias, repo: repo}
}

// VistaDefectos es lo que muestra la página de defectos: el proyecto, sus defectos y sus sprints e historias.
type VistaDefectos struct {
	Proyecto  project.Proyecto
	Defectos  []defect.Defecto
	Sprints   []sprint.Sprint
	Historias []backlog.Historia
}

// SprintDe devuelve el nombre del sprint con ese ID ("Sprint 1"), o "" si no es del proyecto.
func (v VistaDefectos) SprintDe(id int64) string {
	for _, s := range v.Sprints {
		if s.ID == id {
			return fmt.Sprintf("Sprint %d", s.Numero)
		}
	}
	return ""
}

// HistoriaDe devuelve "HU-N · título" de la historia con ese ID, o "" si el defecto no es de una historia.
func (v VistaDefectos) HistoriaDe(id int64) string {
	for _, h := range v.Historias {
		if h.ID == id {
			return h.Identificador() + " · " + h.Titulo
		}
	}
	return ""
}

// Siguientes devuelve los estados a los que puede pasar un defecto en el estado dado (HU-28), para los botones.
func (VistaDefectos) Siguientes(e defect.Estado) []defect.Estado {
	return defect.Siguientes(e)
}

// FilaDefectos es una fila del resumen de defectos (HU-34): un sprint con sus detectados y resueltos.
type FilaDefectos struct {
	Nombre string
	metrics.ConteoSprint
}

// ResumenDefectos es el resumen de defectos del proyecto (HU-34): una fila por sprint y el total.
type ResumenDefectos struct {
	Severidad  defect.Severidad // la del filtro; "" = todas
	Sprints    []FilaDefectos   // en el orden de los sprints
	Detectados int
	Resueltos  int
	Abiertos   int
}

// Resumen cuenta los defectos detectados y resueltos por sprint y en el proyecto (M8 y M9), solo los de esa severidad.
// Una severidad vacía o que no existe cuenta todos (CA-34.2).
func (v VistaDefectos) Resumen(severidad defect.Severidad) ResumenDefectos {
	if !slices.Contains(defect.Severidades, severidad) {
		severidad = ""
	}
	datos := make([]metrics.Defecto, len(v.Defectos))
	for i, d := range v.Defectos {
		datos[i] = metrics.Defecto{
			Severidad: string(d.Severidad), SprintDeteccion: d.SprintDeteccionID, SprintResolucion: d.SprintResolucionID,
			Resuelto: d.Estado == defect.EstadoResuelto || d.Estado == defect.EstadoCerrado,
		}
	}
	c := metrics.ContarDefectos(datos, string(severidad))
	r := ResumenDefectos{Severidad: severidad, Detectados: c.Detectados, Resueltos: c.Resueltos, Abiertos: c.Abiertos}
	for _, s := range v.Sprints {
		r.Sprints = append(r.Sprints, FilaDefectos{Nombre: fmt.Sprintf("Sprint %d", s.Numero), ConteoSprint: c.PorSprint[s.ID]})
	}
	return r
}

// Ver devuelve el proyecto con sus defectos, sprints e historias, o project.ErrNoEncontrado.
func (s *Defectos) Ver(ctx context.Context, proyectoID int64) (VistaDefectos, error) {
	v, err := s.contexto(ctx, proyectoID)
	if err != nil {
		return VistaDefectos{}, err
	}
	if v.Defectos, err = s.repo.ListarPorProyecto(ctx, proyectoID); err != nil {
		return VistaDefectos{}, fmt.Errorf("listar los defectos del proyecto %d: %w", proyectoID, err)
	}
	return v, nil
}

// Registrar valida los datos (HU-27) contra los sprints y las historias del proyecto y guarda el defecto Abierto. Los
// datos inválidos se devuelven en domain.ErroresValidacion; si el proyecto no existe, project.ErrNoEncontrado.
func (s *Defectos) Registrar(ctx context.Context, proyectoID int64, d defect.Datos) (defect.Defecto, error) {
	v, err := s.contexto(ctx, proyectoID)
	if err != nil {
		return defect.Defecto{}, err
	}
	nuevo, err := defect.Nuevo(d, proyectoID, v.Sprints, v.Historias)
	if err != nil {
		return defect.Defecto{}, err
	}
	guardado, err := s.repo.Guardar(ctx, nuevo)
	if err != nil {
		return defect.Defecto{}, fmt.Errorf("guardar el defecto: %w", err)
	}
	return guardado, nil
}

func (s *Defectos) contexto(ctx context.Context, proyectoID int64) (VistaDefectos, error) {
	p, err := s.proyectos.BuscarPorID(ctx, proyectoID)
	if err != nil {
		return VistaDefectos{}, fmt.Errorf("buscar el proyecto %d: %w", proyectoID, err)
	}
	sprints, err := s.sprints.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return VistaDefectos{}, fmt.Errorf("listar los sprints del proyecto %d: %w", proyectoID, err)
	}
	historias, err := s.historias.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return VistaDefectos{}, fmt.Errorf("listar las historias del proyecto %d: %w", proyectoID, err)
	}
	return VistaDefectos{Proyecto: p, Sprints: sprints, Historias: historias}, nil
}

// CambiarEstado pasa el defecto al estado nuevo (HU-28); al resolverlo, sprintResolucion es el sprint en que se
// resolvió. Los errores son los de defect.CambiarEstado, defect.ErrDefectoNoEncontrado o project.ErrNoEncontrado.
func (s *Defectos) CambiarEstado(ctx context.Context, proyectoID, defectoID int64, nuevo defect.Estado, sprintResolucion int64) (defect.Defecto, error) {
	v, err := s.contexto(ctx, proyectoID)
	if err != nil {
		return defect.Defecto{}, err
	}
	d, err := s.repo.BuscarPorID(ctx, proyectoID, defectoID)
	if err != nil {
		return defect.Defecto{}, fmt.Errorf("buscar el defecto %d: %w", defectoID, err)
	}
	cambiado, err := d.CambiarEstado(nuevo, sprintResolucion, v.Sprints)
	if err != nil {
		return defect.Defecto{}, err
	}
	if err := s.repo.Actualizar(ctx, cambiado); err != nil {
		return defect.Defecto{}, fmt.Errorf("guardar el defecto %d: %w", defectoID, err)
	}
	return cambiado, nil
}
