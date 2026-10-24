package service

import (
	"context"
	"fmt"
	"slices"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
)

// ReporteProyecto es todo lo que muestra el reporte del proyecto (HU-38): el resumen del panel y el reporte de cada
// sprint. HU-39 lo pasa a PDF.
type ReporteProyecto struct {
	Resumen VistaPanel      // los totales del proyecto, como el panel (RN2)
	Sprints []ReporteSprint // uno por sprint, ordenados por número, iguales a su reporte de HU-37 (RN1)
}

// Evolucion devuelve los reportes de los sprints cerrados, en orden: cada uno trae los SP de su cierre, su cumplimiento
// y la velocidad hasta ese sprint (CA-38.2, RN3).
func (r ReporteProyecto) Evolucion() []ReporteSprint {
	var cerrados []ReporteSprint
	for _, s := range r.Sprints {
		if s.Cerrado() {
			cerrados = append(cerrados, s)
		}
	}
	return cerrados
}

// SprintsTexto devuelve "2 cerrados de 3" (RN4).
func (r ReporteProyecto) SprintsTexto() string {
	if len(r.Sprints) == 0 {
		return sinDatos
	}
	return fmt.Sprintf("%s de %d", plural(len(r.Evolucion()), "cerrado"), len(r.Sprints))
}

// SPTexto devuelve los SP completados del proyecto (HU-30): "8 de 13 SP (61,5 %)".
func (r ReporteProyecto) SPTexto() string {
	sp := r.Resumen.StoryPoints
	if sp.Planificados == 0 {
		return sinDatos
	}
	return fmt.Sprintf("%g de %g SP (%s)", sp.Completados, sp.Planificados, r.Resumen.PorcentajeSP())
}

// VelocidadTexto devuelve la velocidad del equipo (HU-31): "4,0 SP por sprint (2 sprints cerrados)".
func (r ReporteProyecto) VelocidadTexto() string {
	if r.Resumen.Velocidad.SprintsUsados == 0 {
		return sinDatos
	}
	return velocidadTexto(r.Resumen.Velocidad)
}

func (r ReporteProyecto) sinHoras() bool {
	return r.Resumen.Horas.Estimadas == 0 && r.Resumen.Horas.Reales == 0
}

// HorasTexto devuelve las horas del proyecto (HU-33): "12 h estimadas · 11 h reales".
func (r ReporteProyecto) HorasTexto() string {
	if r.sinHoras() {
		return sinDatos
	}
	return fmt.Sprintf("%s estimadas · %s reales", horasTexto(r.Resumen.Horas.Estimadas), horasTexto(r.Resumen.Horas.Reales))
}

// DesviacionTexto devuelve la desviación de horas del proyecto (HU-33): "-1 h (-8,3 %)".
func (r ReporteProyecto) DesviacionTexto() string {
	if r.sinHoras() {
		return sinDatos
	}
	return desviacionTexto(r.Resumen.Horas)
}

// DefectosTexto devuelve los defectos del proyecto (HU-34): "3 detectados · 2 resueltos · 1 abierto".
func (r ReporteProyecto) DefectosTexto() string {
	d := r.Resumen.Defectos
	if d.Detectados == 0 {
		return sinDatos
	}
	return plural(d.Detectados, "detectado") + " · " + plural(d.Resueltos, "resuelto") + " · " + plural(d.Abiertos, "abierto")
}

// Proyecto devuelve el reporte del proyecto, o project.ErrNoEncontrado si no existe. Lee los listados del proyecto una
// sola vez y arma con ellos el reporte de cada sprint.
func (r *Reportes) Proyecto(ctx context.Context, proyectoID int64) (ReporteProyecto, error) {
	resumen, err := r.panel.Ver(ctx, proyectoID)
	if err != nil {
		return ReporteProyecto{}, err
	}
	sprints, err := r.sprints.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return ReporteProyecto{}, fmt.Errorf("listar los sprints del proyecto %d: %w", proyectoID, err)
	}
	slices.SortFunc(sprints, func(a, b sprint.Sprint) int { return a.Numero - b.Numero })
	d, err := r.datos(ctx, resumen.Proyecto, sprints)
	if err != nil {
		return ReporteProyecto{}, err
	}
	rep := ReporteProyecto{Resumen: resumen}
	for _, s := range sprints {
		bloque, err := r.armar(ctx, d, s)
		if err != nil {
			return ReporteProyecto{}, err
		}
		rep.Sprints = append(rep.Sprints, bloque)
	}
	return rep, nil
}
