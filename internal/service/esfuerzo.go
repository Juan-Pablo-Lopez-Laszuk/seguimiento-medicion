package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
)

// ListadorTareas es lo que el registro de esfuerzo necesita del repositorio de tareas (HU-25).
type ListadorTareas interface {
	ListarPorHistoria(ctx context.Context, historiaID int64) ([]effort.Tarea, error)
}

// RepositorioRegistros es lo que los casos de uso de esfuerzo necesitan de la base de datos: registrar (HU-24) y
// consultar y corregir (HU-26).
type RepositorioRegistros interface {
	RepositorioCorreccionRegistros
	// Guardar guarda un registro nuevo y lo devuelve con su ID.
	Guardar(ctx context.Context, r effort.Registro) (effort.Registro, error)
	// ListarPorHistoria devuelve los registros de la historia, la fecha más reciente primero.
	ListarPorHistoria(ctx context.Context, historiaID int64) ([]effort.Registro, error)
	// ListarPorIntegrante devuelve los registros del integrante, para el total del día.
	ListarPorIntegrante(ctx context.Context, integranteID int64) ([]effort.Registro, error)
}

// Esfuerzo agrupa los casos de uso del registro de esfuerzo de una historia (HU-24).
type Esfuerzo struct {
	proyectos   BuscadorProyectos
	historias   BuscadorHistorias
	integrantes ListadorIntegrantes
	tareas      ListadorTareas
	repo        RepositorioRegistros
	ahora       func() time.Time
}

// NuevoEsfuerzo arma el caso de uso. ahora da la hora actual (time.Now en la aplicación, fija en los tests).
func NuevoEsfuerzo(proyectos BuscadorProyectos, historias BuscadorHistorias, integrantes ListadorIntegrantes, tareas ListadorTareas, repo RepositorioRegistros, ahora func() time.Time) *Esfuerzo {
	return &Esfuerzo{proyectos: proyectos, historias: historias, integrantes: integrantes, tareas: tareas, repo: repo, ahora: ahora}
}

// VistaEsfuerzo es lo que muestra la página de esfuerzo de una historia.
type VistaEsfuerzo struct {
	Proyecto    project.Proyecto
	Historia    backlog.Historia
	Registros   []effort.Registro
	HorasReales float64
	Integrantes []member.Integrante // todos, también los dados de baja, para mostrar quién cargó cada registro
	Tareas      []effort.Tarea
	Hoy         time.Time // en Argentina; es el máximo del campo fecha
}

// Activos devuelve los integrantes que pueden registrar horas (CA-24.3).
func (v VistaEsfuerzo) Activos() []member.Integrante {
	var activos []member.Integrante
	for _, i := range v.Integrantes {
		if i.Activo {
			activos = append(activos, i)
		}
	}
	return activos
}

// NombreDe devuelve el nombre del integrante con ese ID, o "" si no es del proyecto.
func (v VistaEsfuerzo) NombreDe(id int64) string {
	for _, i := range v.Integrantes {
		if i.ID == id {
			return i.Nombre
		}
	}
	return ""
}

// TituloDeTarea devuelve el título de la tarea con ese ID, o "" si el registro es de toda la historia.
func (v VistaEsfuerzo) TituloDeTarea(id int64) string {
	for _, t := range v.Tareas {
		if t.ID == id {
			return t.Titulo
		}
	}
	return ""
}

// Ver devuelve la historia HU-N con sus registros, sus horas reales (M5), los integrantes y sus tareas.
func (s *Esfuerzo) Ver(ctx context.Context, proyectoID int64, numero int) (VistaEsfuerzo, error) {
	v, err := s.contexto(ctx, proyectoID, numero)
	if err != nil {
		return VistaEsfuerzo{}, err
	}
	if v.Registros, err = s.repo.ListarPorHistoria(ctx, v.Historia.ID); err != nil {
		return VistaEsfuerzo{}, fmt.Errorf("listar los registros de la historia %d: %w", v.Historia.ID, err)
	}
	v.HorasReales = effort.HorasReales(v.Registros)
	return v, nil
}

// Registrar valida los datos (HU-24) y guarda el registro en la historia HU-N. Los datos inválidos se devuelven en
// domain.ErroresValidacion; si el proyecto o la historia no existen, project.ErrNoEncontrado o
// backlog.ErrHistoriaNoEncontrada.
func (s *Esfuerzo) Registrar(ctx context.Context, proyectoID int64, numero int, d effort.DatosRegistro) (effort.Registro, error) {
	v, err := s.contexto(ctx, proyectoID, numero)
	if err != nil {
		return effort.Registro{}, err
	}
	delIntegrante, err := s.repo.ListarPorIntegrante(ctx, d.IntegranteID)
	if err != nil {
		return effort.Registro{}, fmt.Errorf("listar los registros del integrante %d: %w", d.IntegranteID, err)
	}
	nuevo, err := effort.NuevoRegistro(d, effort.ContextoRegistro{
		HistoriaID: v.Historia.ID, InicioProyecto: v.Proyecto.FechaInicio, FinProyecto: v.Proyecto.FechaFin,
		Hoy: v.Hoy, Integrantes: v.Integrantes, Tareas: v.Tareas, DelIntegrante: delIntegrante,
	})
	if err != nil {
		return effort.Registro{}, err
	}
	guardado, err := s.repo.Guardar(ctx, nuevo)
	if err != nil {
		return effort.Registro{}, fmt.Errorf("guardar el registro: %w", err)
	}
	return guardado, nil
}

// contexto trae el proyecto, la historia, los integrantes, las tareas y la fecha de hoy.
func (s *Esfuerzo) contexto(ctx context.Context, proyectoID int64, numero int) (VistaEsfuerzo, error) {
	p, err := s.proyectos.BuscarPorID(ctx, proyectoID)
	if err != nil {
		return VistaEsfuerzo{}, fmt.Errorf("buscar el proyecto %d: %w", proyectoID, err)
	}
	h, err := s.historias.BuscarPorNumero(ctx, proyectoID, numero)
	if err != nil {
		return VistaEsfuerzo{}, fmt.Errorf("buscar la historia HU-%d del proyecto %d: %w", numero, proyectoID, err)
	}
	integrantes, err := s.integrantes.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return VistaEsfuerzo{}, fmt.Errorf("listar los integrantes del proyecto %d: %w", proyectoID, err)
	}
	tareas, err := s.tareas.ListarPorHistoria(ctx, h.ID)
	if err != nil {
		return VistaEsfuerzo{}, fmt.Errorf("listar las tareas de la historia %d: %w", h.ID, err)
	}
	return VistaEsfuerzo{Proyecto: p, Historia: h, Integrantes: integrantes, Tareas: tareas, Hoy: domain.HoyEnArgentina(s.ahora())}, nil
}
