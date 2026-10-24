package service

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/defect"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/metrics"
)

// ListadorDefectos es lo que el reporte necesita del repositorio de defectos (HU-37).
type ListadorDefectos interface {
	ListarPorProyecto(ctx context.Context, proyectoID int64) ([]defect.Defecto, error)
}

// Reportes arma el reporte de un sprint para la Sprint Review (spec: specs/HU-37-reporte-sprint.md) y el del proyecto
// para el cliente (specs/HU-38-reporte-proyecto.md). No calcula nada propio: junta los datos y usa las funciones de
// internal/metrics (CA-37.2). Solo lee: no cambia nada (HU-37 RN7, HU-38 RN6).
type Reportes struct {
	proyectos BuscadorProyectos
	sprints   ListadorSprints
	historias ListadorHistorias
	cierres   *Cierres // la foto del cierre (HU-15)
	horas     *Horas   // las horas de cada historia, como la página de horas (HU-33)
	defectos  ListadorDefectos
	panel     *Panel // el resumen del reporte del proyecto (HU-38 RN2)
}

// NuevoReportes arma el caso de uso.
func NuevoReportes(proyectos BuscadorProyectos, sprints ListadorSprints, historias ListadorHistorias, cierres *Cierres, horas *Horas, defectos ListadorDefectos, panel *Panel) *Reportes {
	return &Reportes{proyectos: proyectos, sprints: sprints, historias: historias, cierres: cierres, horas: horas, defectos: defectos, panel: panel}
}

// ReporteSprint es todo lo que muestra el reporte de un sprint. HU-38 arma el del proyecto con uno por sprint y HU-39
// lo pasa a PDF.
type ReporteSprint struct {
	Proyecto    project.Proyecto
	Sprint      sprint.Sprint
	SP          metrics.StoryPoints  // M1 y M2: los de la foto si está cerrado (RN1), los de hoy si no (RN2)
	Completadas metrics.Porcentaje   // M7 (RN3)
	Velocidad   metrics.Velocidad    // M3 con los sprints cerrados hasta este (RN4)
	Horas       metrics.Desviacion   // M4, M5 y M6 del sprint: la suma de sus historias (RN5)
	Historias   []FilaReporte        // ordenadas por número
	Conteo      metrics.ConteoSprint // M8 y M9 del sprint (RN6)
	Defectos    []FilaDefecto        // los detectados o resueltos en el sprint, ordenados por ID
}

// FilaDefecto es un defecto de la tabla del reporte, con los nombres de su historia y de sus sprints.
type FilaDefecto struct {
	defect.Defecto
	Historia    string // "HU-1 · Login", o "" si no es de una historia
	DetectadoEn string // "Sprint 1"
	ResueltoEn  string // "Sprint 1", o "—" si no está resuelto
}

// FilaReporte es una historia del reporte: la fila de la foto (o la historia de hoy) con sus horas.
type FilaReporte struct {
	FilaFoto
	Horas metrics.Desviacion
}

// Cerrado dice si el sprint está cerrado; si no, el reporte es parcial (RN2).
func (r ReporteSprint) Cerrado() bool {
	return r.Sprint.Estado == sprint.EstadoCerrado
}

// SPTexto devuelve "5 de 13 SP".
func (r ReporteSprint) SPTexto() string {
	return fmt.Sprintf("%g de %g SP", r.SP.Completados, r.SP.Planificados)
}

// CumplimientoTexto devuelve el cumplimiento de HU-15 ("38,5 %"), o "—" sin SP planificados.
func (r ReporteSprint) CumplimientoTexto() string {
	return r.SP.CumplimientoTexto()
}

// CompletadasTexto devuelve "1 de 3 (33,3 %)".
func (r ReporteSprint) CompletadasTexto() string {
	return fmt.Sprintf("%d de %d (%s)", r.Completadas.Hechas, r.Completadas.Planificadas, r.Completadas.Texto())
}

// VelocidadTexto devuelve "5,0 SP por sprint (1 sprint cerrado)", o el aviso de HU-31 si no hay sprints cerrados.
func (r ReporteSprint) VelocidadTexto() string {
	if aviso := r.Velocidad.Mensaje(); aviso != "" {
		return aviso
	}
	return velocidadTexto(r.Velocidad)
}

// VelocidadPorSprint devuelve solo el valor de la velocidad ("5,0 SP por sprint"), para la evolución de HU-38.
func (r ReporteSprint) VelocidadPorSprint() string {
	return unDecimal(r.Velocidad.Valor) + " SP por sprint"
}

func velocidadTexto(v metrics.Velocidad) string {
	sprints := "1 sprint cerrado"
	if v.SprintsUsados != 1 {
		sprints = fmt.Sprintf("%d sprints cerrados", v.SprintsUsados)
	}
	return fmt.Sprintf("%s SP por sprint (%s)", unDecimal(v.Valor), sprints)
}

// HorasTexto devuelve "10 h estimadas · 11 h reales".
func (r ReporteSprint) HorasTexto() string {
	return fmt.Sprintf("%s estimadas · %s reales", horasTexto(r.Horas.Estimadas), horasTexto(r.Horas.Reales))
}

// DesviacionTexto devuelve "+1 h (+10,0 %)", o "0 h (no aplicable)" sin horas estimadas.
func (r ReporteSprint) DesviacionTexto() string {
	return desviacionTexto(r.Horas)
}

// DefectosTexto devuelve "2 detectados · 1 resuelto".
func (r ReporteSprint) DefectosTexto() string {
	return plural(r.Conteo.Detectados, "detectado") + " · " + plural(r.Conteo.Resueltos, "resuelto")
}

// plural devuelve "1 resuelto" o "2 resueltos".
func plural(n int, palabra string) string {
	if n == 1 {
		return "1 " + palabra
	}
	return fmt.Sprintf("%d %ss", n, palabra)
}

// Advertencia devuelve el aviso de HU-30 si el sprint abierto tiene historias sin estimar, o "".
func (r ReporteSprint) Advertencia() string {
	return r.SP.Advertencia()
}

// EstimadasTexto devuelve las horas estimadas de la historia ("6 h").
func (f FilaReporte) EstimadasTexto() string {
	return horasTexto(f.Horas.Estimadas)
}

// RealesTexto devuelve las horas reales de la historia ("8 h").
func (f FilaReporte) RealesTexto() string {
	return horasTexto(f.Horas.Reales)
}

// DesviacionTexto devuelve la desviación de la historia ("+2 h (+33,3 %)").
func (f FilaReporte) DesviacionTexto() string {
	return desviacionTexto(f.Horas)
}

func horasTexto(horas float64) string {
	return effort.FormatearHoras(horas) + " h"
}

func desviacionTexto(d metrics.Desviacion) string {
	return fmt.Sprintf("%s (%s)", d.HorasTexto(), d.PorcentajeTexto())
}

// unDecimal devuelve el número con 1 decimal y coma decimal ("16,5"), como se muestra en pantalla (RG-2).
func unDecimal(valor float64) string {
	return strings.Replace(fmt.Sprintf("%.1f", valor), ".", ",", 1)
}

// Sprint devuelve el reporte del Sprint N del proyecto. Si el proyecto o el sprint no existen, el error es
// project.ErrNoEncontrado o sprint.ErrNoEncontrado (CA-37.3).
func (r *Reportes) Sprint(ctx context.Context, proyectoID int64, numero int) (ReporteSprint, error) {
	p, s, sprints, err := buscarSprint(ctx, r.proyectos, r.sprints, proyectoID, numero)
	if err != nil {
		return ReporteSprint{}, err
	}
	d, err := r.datos(ctx, p, sprints)
	if err != nil {
		return ReporteSprint{}, err
	}
	return r.armar(ctx, d, s)
}

// datosReporte son los listados del proyecto que usa el reporte de cualquiera de sus sprints. El reporte del proyecto
// los lee una sola vez para todos los sprints.
type datosReporte struct {
	proyecto  project.Proyecto
	sprints   []sprint.Sprint
	historias []backlog.Historia
	defectos  []defect.Defecto
}

func (r *Reportes) datos(ctx context.Context, p project.Proyecto, sprints []sprint.Sprint) (datosReporte, error) {
	historias, err := r.historias.ListarPorProyecto(ctx, p.ID)
	if err != nil {
		return datosReporte{}, fmt.Errorf("listar las historias del proyecto %d: %w", p.ID, err)
	}
	defectos, err := r.defectos.ListarPorProyecto(ctx, p.ID)
	if err != nil {
		return datosReporte{}, fmt.Errorf("listar los defectos del proyecto %d: %w", p.ID, err)
	}
	return datosReporte{proyecto: p, sprints: sprints, historias: historias, defectos: defectos}, nil
}

// armar arma el reporte del sprint s con los listados del proyecto.
func (r *Reportes) armar(ctx context.Context, d datosReporte, s sprint.Sprint) (ReporteSprint, error) {
	historias, numero := d.historias, s.Numero
	rep := ReporteSprint{Proyecto: d.proyecto, Sprint: s}
	filas, err := r.filas(ctx, &rep, historias)
	if err != nil {
		return ReporteSprint{}, err
	}

	ids := make(map[int]int64, len(historias))
	for _, h := range historias {
		ids[h.Numero] = h.ID
	}
	paraMetricas := make([]metrics.Historia, 0, len(filas))
	desviaciones := make([]metrics.Desviacion, 0, len(filas))
	for _, f := range filas {
		d, err := r.horas.desviacionDe(ctx, ids[f.Numero])
		if err != nil {
			return ReporteSprint{}, err
		}
		rep.Historias = append(rep.Historias, FilaReporte{FilaFoto: f, Horas: d})
		desviaciones = append(desviaciones, d)
		paraMetricas = append(paraMetricas, historiaParaMetricas(f))
	}
	rep.Horas = metrics.SumarDesviaciones(desviaciones)
	rep.Completadas = metrics.CalcularPorcentajeCompletadas(paraMetricas)
	if !rep.Cerrado() {
		if rep.SP, err = metrics.CalcularStoryPoints(paraMetricas); err != nil {
			return ReporteSprint{}, fmt.Errorf("calcular los SP del Sprint %d: %w", numero, err)
		}
	}
	conDefectos(&rep, d)
	hastaEste := slices.DeleteFunc(slices.Clone(d.sprints), func(otro sprint.Sprint) bool { return otro.Numero > numero })
	if rep.Velocidad, err = velocidadDe(hastaEste); err != nil {
		return ReporteSprint{}, err
	}
	return rep, nil
}

// filas devuelve las historias del reporte: la foto del cierre si el sprint está cerrado, con sus SP guardados (RN1),
// o las historias de hoy del sprint, ordenadas por número (RN2).
func (r *Reportes) filas(ctx context.Context, rep *ReporteSprint, historias []backlog.Historia) ([]FilaFoto, error) {
	if rep.Cerrado() {
		d, err := r.cierres.Detalle(ctx, rep.Proyecto.ID, rep.Sprint.Numero)
		if err != nil {
			return nil, err
		}
		rep.SP = d.SP
		return d.Historias, nil
	}
	var filas []FilaFoto
	for _, h := range historias {
		if h.SprintID != nil && *h.SprintID == rep.Sprint.ID {
			filas = append(filas, FilaFoto{ProyectoID: h.ProyectoID, Numero: h.Numero, Titulo: h.Titulo,
				StoryPoints: h.StoryPoints, Completada: h.Estado == backlog.EstadoHecho})
		}
	}
	slices.SortFunc(filas, func(a, b FilaFoto) int { return a.Numero - b.Numero })
	return filas, nil
}

// conDefectos cuenta los defectos detectados y resueltos en el sprint con el resumen de HU-34 (M8 y M9) y arma la
// tabla con los que se detectaron o se resolvieron en él (RN6).
func conDefectos(rep *ReporteSprint, d datosReporte) {
	vista := VistaDefectos{Proyecto: rep.Proyecto, Defectos: d.defectos, Sprints: d.sprints, Historias: d.historias}
	resumen := vista.Resumen("")
	for i, s := range d.sprints {
		if s.ID == rep.Sprint.ID {
			rep.Conteo = resumen.Sprints[i].ConteoSprint
		}
	}
	for _, def := range d.defectos {
		if def.SprintDeteccionID != rep.Sprint.ID && def.SprintResolucionID != rep.Sprint.ID {
			continue
		}
		f := FilaDefecto{Defecto: def, Historia: vista.HistoriaDe(def.HistoriaID), DetectadoEn: vista.SprintDe(def.SprintDeteccionID), ResueltoEn: sinResultado}
		if def.SprintResolucionID != 0 {
			f.ResueltoEn = vista.SprintDe(def.SprintResolucionID)
		}
		rep.Defectos = append(rep.Defectos, f)
	}
	slices.SortFunc(rep.Defectos, func(a, b FilaDefecto) int { return int(a.ID - b.ID) })
}

// historiaParaMetricas pasa una fila del reporte a lo que necesitan M1, M2 y M7.
func historiaParaMetricas(f FilaFoto) metrics.Historia {
	if f.StoryPoints == nil {
		return metrics.Historia{Hecha: f.Completada, SinEstimar: true}
	}
	return metrics.Historia{SP: float64(*f.StoryPoints), Hecha: f.Completada}
}
