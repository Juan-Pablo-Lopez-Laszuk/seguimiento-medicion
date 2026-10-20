package service

import (
	"context"
	"fmt"
	"slices"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/metrics"
)

// sinResultado es lo que muestra la lista para un sprint sin resultado (HU-15 RN1).
const sinResultado = "—"

// ResumenSprint es una fila de la lista de sprints (HU-11) con el resultado de los cerrados (HU-15).
type ResumenSprint struct {
	sprint.Sprint
	ConResultado bool                // true si está cerrado y tiene los SP de la foto del cierre
	SP           metrics.StoryPoints // SP planificados y completados de la foto; valen solo si ConResultado
}

// SPCompletadosTexto devuelve "5 de 13", o "—" si el sprint no tiene resultado.
func (r ResumenSprint) SPCompletadosTexto() string {
	if !r.ConResultado {
		return sinResultado
	}
	return fmt.Sprintf("%g de %g", r.SP.Completados, r.SP.Planificados)
}

// CumplimientoTexto devuelve el cumplimiento ("38,5 %"), o "—" si el sprint no tiene resultado o no tenía SP
// planificados (RN3).
func (r ResumenSprint) CumplimientoTexto() string {
	if !r.ConResultado {
		return sinResultado
	}
	return r.SP.CumplimientoTexto()
}

// Resumir devuelve el proyecto y sus sprints ordenados por número (RN4), con el resultado de los cerrados (RN1). Los
// errores son los de Listar.
func (s *Sprints) Resumir(ctx context.Context, proyectoID int64) (project.Proyecto, []ResumenSprint, error) {
	p, lista, err := s.Listar(ctx, proyectoID)
	if err != nil {
		return project.Proyecto{}, nil, err
	}
	slices.SortFunc(lista, func(a, b sprint.Sprint) int { return a.Numero - b.Numero })
	resumen := make([]ResumenSprint, len(lista))
	for i, sp := range lista {
		resumen[i] = ResumenSprint{Sprint: sp}
		if sp.Estado == sprint.EstadoCerrado && sp.SPPlanificados != nil && sp.SPCompletados != nil {
			resumen[i].ConResultado = true
			resumen[i].SP = metrics.StoryPoints{Planificados: float64(*sp.SPPlanificados), Completados: float64(*sp.SPCompletados)}
		}
	}
	return p, resumen, nil
}

// DetalleCierre es el resultado de un sprint cerrado con la foto del cierre, que muestra su página (HU-15 CA-15.2).
type DetalleCierre struct {
	Cerrado   bool                // false si el sprint sigue abierto: no hay detalle
	SP        metrics.StoryPoints // SP planificados y completados de la foto
	Historias []FilaFoto          // la foto, ordenada por número de historia
}

// FilaFoto es una historia de la foto del cierre: los SP y el resultado son los del cierre, el título es el de hoy
// (RN6).
type FilaFoto struct {
	ProyectoID  int64
	Numero      int
	Titulo      string
	StoryPoints *int
	Completada  bool
}

// Identificador devuelve "HU-N", como backlog.Historia.
func (f FilaFoto) Identificador() string {
	return fmt.Sprintf("HU-%d", f.Numero)
}

// Resultado devuelve "Completada" o "No completada".
func (f FilaFoto) Resultado() string {
	if f.Completada {
		return "Completada"
	}
	return "No completada"
}

// Completadas devuelve cuántas historias de la foto se completaron (la cuenta de HU-32).
func (d DetalleCierre) Completadas() int {
	n := 0
	for _, f := range d.Historias {
		if f.Completada {
			n++
		}
	}
	return n
}

// Resumen devuelve el resultado como se muestra arriba de la página: "5 de 13 SP completados · 38,5 % de
// cumplimiento · 1 de 3 historias", o "El sprint se cerró sin historias" si la foto está vacía (RN3).
func (d DetalleCierre) Resumen() string {
	if len(d.Historias) == 0 {
		return "El sprint se cerró sin historias"
	}
	return fmt.Sprintf("%g de %g SP completados · %s de cumplimiento · %d de %d historias",
		d.SP.Completados, d.SP.Planificados, d.SP.CumplimientoTexto(), d.Completadas(), len(d.Historias))
}

// Detalle devuelve el resultado del Sprint N con la foto del cierre (RN5 y RN6). Si el sprint sigue abierto, el
// detalle viene con Cerrado en false y sin error. Si el proyecto o el sprint no existen, el error es
// project.ErrNoEncontrado o sprint.ErrNoEncontrado.
func (c *Cierres) Detalle(ctx context.Context, proyectoID int64, numero int) (DetalleCierre, error) {
	_, s, _, err := buscarSprint(ctx, c.proyectos, c.sprints, proyectoID, numero)
	if err != nil || s.Estado != sprint.EstadoCerrado {
		return DetalleCierre{}, err
	}
	d := DetalleCierre{Cerrado: true}
	if s.SPPlanificados != nil && s.SPCompletados != nil {
		d.SP = metrics.StoryPoints{Planificados: float64(*s.SPPlanificados), Completados: float64(*s.SPCompletados)}
	}
	foto, err := c.repo.Foto(ctx, s.ID)
	if err != nil {
		return DetalleCierre{}, fmt.Errorf("leer la foto del Sprint %d del proyecto %d: %w", numero, proyectoID, err)
	}
	historias, err := c.historias.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return DetalleCierre{}, fmt.Errorf("listar las historias del proyecto %d: %w", proyectoID, err)
	}
	titulos := make(map[int64]string, len(historias))
	for _, h := range historias {
		titulos[h.ID] = h.Titulo
	}
	for _, f := range foto {
		d.Historias = append(d.Historias, FilaFoto{ProyectoID: proyectoID, Numero: f.Numero, Titulo: titulos[f.HistoriaID],
			StoryPoints: f.StoryPoints, Completada: f.Completada})
	}
	return d, nil
}
