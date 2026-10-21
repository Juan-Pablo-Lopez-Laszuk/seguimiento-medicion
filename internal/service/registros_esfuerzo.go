package service

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
)

// RepositorioCorreccionRegistros es lo que la consulta y la corrección de registros necesitan de la base de datos
// (HU-26).
type RepositorioCorreccionRegistros interface {
	// ListarFiltrados devuelve los registros de esas historias que cumplen el filtro, la fecha más reciente primero.
	ListarFiltrados(ctx context.Context, historiaIDs []int64, f effort.Filtro) ([]effort.Registro, error)
	// ListarPorIntegrante devuelve los registros del integrante, para el total del día.
	ListarPorIntegrante(ctx context.Context, integranteID int64) ([]effort.Registro, error)
	// BuscarPorID devuelve el registro, o effort.ErrRegistroNoEncontrado.
	BuscarPorID(ctx context.Context, id int64) (effort.Registro, error)
	// Actualizar guarda el registro corregido, o devuelve effort.ErrRegistroNoEncontrado.
	Actualizar(ctx context.Context, r effort.Registro) error
}

// RegistrosEsfuerzo consulta y corrige los registros de esfuerzo de un proyecto (HU-26, spec:
// specs/HU-26-registros-esfuerzo.md).
type RegistrosEsfuerzo struct {
	proyectos   BuscadorProyectos
	sprints     ListadorSprints
	historias   ListadorHistorias
	integrantes ListadorIntegrantes
	tareas      ListadorTareas
	repo        RepositorioCorreccionRegistros
	horas       *Horas // las horas de cada historia, como la página de horas (HU-33)
	ahora       func() time.Time
}

// NuevoRegistrosEsfuerzo arma el caso de uso. ahora da la hora actual (time.Now en la aplicación, fija en los tests).
func NuevoRegistrosEsfuerzo(proyectos BuscadorProyectos, sprints ListadorSprints, historias ListadorHistorias, integrantes ListadorIntegrantes, tareas ListadorTareas, repo RepositorioCorreccionRegistros, horas *Horas, ahora func() time.Time) *RegistrosEsfuerzo {
	return &RegistrosEsfuerzo{proyectos: proyectos, sprints: sprints, historias: historias, integrantes: integrantes, tareas: tareas, repo: repo, horas: horas, ahora: ahora}
}

// FiltroRegistros son los filtros de la consulta como llegan de la URL: el ID del integrante y los números de la
// historia y del sprint. Un 0 no filtra.
type FiltroRegistros struct {
	IntegranteID int64
	Historia     int // número de HU-N
	Sprint       int // número de sprint
}

// FilaRegistro es un registro de la lista, con los nombres de su integrante, su historia y su tarea.
type FilaRegistro struct {
	effort.Registro
	Integrante string
	Historia   string // "HU-1 · Login"
	Tarea      string // "" si es de toda la historia
}

// VistaRegistros es lo que muestra la página de registros de horas del proyecto.
type VistaRegistros struct {
	Proyecto    project.Proyecto
	Filtro      FiltroRegistros // los filtros que se aplicaron: los que no existen quedan en 0 (RN3)
	Integrantes []member.Integrante
	Historias   []backlog.Historia
	Sprints     []sprint.Sprint
	Registros   []FilaRegistro // la fecha más reciente primero
	Total       float64        // horas de la lista
	PorHistoria []FilaHoras    // horas estimadas vs reales de cada historia de la lista (CA-26.3)
}

// Filtrado dice si se aplicó algún filtro.
func (v VistaRegistros) Filtrado() bool {
	return v.Filtro != FiltroRegistros{}
}

// Consultar devuelve los registros del proyecto que cumplen los filtros (CA-26.1) y las horas estimadas vs reales de
// sus historias (CA-26.3). Si el proyecto no existe, el error es project.ErrNoEncontrado.
func (s *RegistrosEsfuerzo) Consultar(ctx context.Context, proyectoID int64, pedido FiltroRegistros) (VistaRegistros, error) {
	p, err := s.proyectos.BuscarPorID(ctx, proyectoID)
	if err != nil {
		return VistaRegistros{}, fmt.Errorf("buscar el proyecto %d: %w", proyectoID, err)
	}
	v := VistaRegistros{Proyecto: p}
	if v.Integrantes, err = s.integrantes.ListarPorProyecto(ctx, proyectoID); err != nil {
		return VistaRegistros{}, fmt.Errorf("listar los integrantes del proyecto %d: %w", proyectoID, err)
	}
	if v.Historias, err = s.historias.ListarPorProyecto(ctx, proyectoID); err != nil {
		return VistaRegistros{}, fmt.Errorf("listar las historias del proyecto %d: %w", proyectoID, err)
	}
	if v.Sprints, err = s.sprints.ListarPorProyecto(ctx, proyectoID); err != nil {
		return VistaRegistros{}, fmt.Errorf("listar los sprints del proyecto %d: %w", proyectoID, err)
	}

	filtro := v.aplicar(pedido)
	ids := make([]int64, len(v.Historias))
	for i, h := range v.Historias {
		ids[i] = h.ID
	}
	registros, err := s.repo.ListarFiltrados(ctx, ids, filtro)
	if err != nil {
		return VistaRegistros{}, fmt.Errorf("listar los registros del proyecto %d: %w", proyectoID, err)
	}

	tareas := map[int64]string{}
	conRegistros := map[int64]bool{}
	for _, r := range registros {
		if !conRegistros[r.HistoriaID] {
			conRegistros[r.HistoriaID] = true
			lista, err := s.tareas.ListarPorHistoria(ctx, r.HistoriaID)
			if err != nil {
				return VistaRegistros{}, fmt.Errorf("listar las tareas de la historia %d: %w", r.HistoriaID, err)
			}
			for _, t := range lista {
				tareas[t.ID] = t.Titulo
			}
		}
		v.Registros = append(v.Registros, FilaRegistro{Registro: r, Integrante: nombreDeIntegrante(r.IntegranteID, v.Integrantes), Historia: v.nombreDeHistoria(r.HistoriaID), Tarea: tareas[r.TareaID]})
		v.Total += r.Horas
	}
	for _, h := range v.Historias {
		if !conRegistros[h.ID] {
			continue
		}
		d, err := s.horas.desviacionDe(ctx, h.ID)
		if err != nil {
			return VistaRegistros{}, err
		}
		v.PorHistoria = append(v.PorHistoria, FilaHoras{Nombre: h.Identificador() + " · " + h.Titulo, Desviacion: d})
	}
	return v, nil
}

// aplicar deja en v.Filtro solo los filtros que existen en el proyecto (RN3) y arma el filtro del dominio, con las
// fechas del sprint elegido (RN2).
func (v *VistaRegistros) aplicar(pedido FiltroRegistros) effort.Filtro {
	var f effort.Filtro
	if slices.ContainsFunc(v.Integrantes, func(i member.Integrante) bool { return i.ID == pedido.IntegranteID }) {
		v.Filtro.IntegranteID, f.IntegranteID = pedido.IntegranteID, pedido.IntegranteID
	}
	for _, h := range v.Historias {
		if h.Numero == pedido.Historia {
			v.Filtro.Historia, f.HistoriaID = pedido.Historia, h.ID
		}
	}
	for _, sp := range v.Sprints {
		if sp.Numero == pedido.Sprint {
			v.Filtro.Sprint, f.Desde, f.Hasta = pedido.Sprint, sp.FechaInicio, sp.FechaFin
		}
	}
	return f
}

func (v VistaRegistros) nombreDeHistoria(id int64) string {
	for _, h := range v.Historias {
		if h.ID == id {
			return h.Identificador() + " · " + h.Titulo
		}
	}
	return ""
}

func nombreDeIntegrante(id int64, integrantes []member.Integrante) string {
	for _, i := range integrantes {
		if i.ID == id {
			return i.Nombre
		}
	}
	return ""
}

// VistaCorreccion es lo que muestra la página para corregir un registro: el registro y, como en la página de
// esfuerzo de HU-24, su historia, los integrantes y las tareas de la historia.
type VistaCorreccion struct {
	VistaEsfuerzo
	Registro effort.Registro
}

// VerCorreccion devuelve el registro con su historia, los integrantes y las tareas. Si el registro no existe o es de
// otro proyecto, el error es effort.ErrRegistroNoEncontrado; si el proyecto no existe, project.ErrNoEncontrado.
func (s *RegistrosEsfuerzo) VerCorreccion(ctx context.Context, proyectoID, registroID int64) (VistaCorreccion, error) {
	p, err := s.proyectos.BuscarPorID(ctx, proyectoID)
	if err != nil {
		return VistaCorreccion{}, fmt.Errorf("buscar el proyecto %d: %w", proyectoID, err)
	}
	r, err := s.repo.BuscarPorID(ctx, registroID)
	if err != nil {
		return VistaCorreccion{}, fmt.Errorf("buscar el registro %d: %w", registroID, err)
	}
	historias, err := s.historias.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return VistaCorreccion{}, fmt.Errorf("listar las historias del proyecto %d: %w", proyectoID, err)
	}
	i := slices.IndexFunc(historias, func(h backlog.Historia) bool { return h.ID == r.HistoriaID })
	if i < 0 {
		return VistaCorreccion{}, fmt.Errorf("el registro %d no es del proyecto %d: %w", registroID, proyectoID, effort.ErrRegistroNoEncontrado)
	}
	v := VistaCorreccion{VistaEsfuerzo: VistaEsfuerzo{Proyecto: p, Historia: historias[i], Hoy: domain.HoyEnArgentina(s.ahora())}, Registro: r}
	if v.Integrantes, err = s.integrantes.ListarPorProyecto(ctx, proyectoID); err != nil {
		return VistaCorreccion{}, fmt.Errorf("listar los integrantes del proyecto %d: %w", proyectoID, err)
	}
	if v.Tareas, err = s.tareas.ListarPorHistoria(ctx, r.HistoriaID); err != nil {
		return VistaCorreccion{}, fmt.Errorf("listar las tareas de la historia %d: %w", r.HistoriaID, err)
	}
	return v, nil
}

// Corregir valida la corrección que hace el integrante quienID (CA-26.2) y guarda el registro corregido. Los datos
// inválidos y quién no puede corregir se devuelven en domain.ErroresValidacion.
func (s *RegistrosEsfuerzo) Corregir(ctx context.Context, proyectoID, registroID, quienID int64, d effort.DatosRegistro) (effort.Registro, error) {
	v, err := s.VerCorreccion(ctx, proyectoID, registroID)
	if err != nil {
		return effort.Registro{}, err
	}
	delIntegrante, err := s.repo.ListarPorIntegrante(ctx, v.Registro.IntegranteID)
	if err != nil {
		return effort.Registro{}, fmt.Errorf("listar los registros del integrante %d: %w", v.Registro.IntegranteID, err)
	}
	corregido, err := effort.Corregir(v.Registro, quienID, d, effort.ContextoRegistro{
		HistoriaID: v.Historia.ID, InicioProyecto: v.Proyecto.FechaInicio, FinProyecto: v.Proyecto.FechaFin,
		Hoy: v.Hoy, Integrantes: v.Integrantes, Tareas: v.Tareas, DelIntegrante: delIntegrante,
	})
	if err != nil {
		return effort.Registro{}, err
	}
	if err := s.repo.Actualizar(ctx, corregido); err != nil {
		return effort.Registro{}, fmt.Errorf("guardar el registro corregido: %w", err)
	}
	return corregido, nil
}
