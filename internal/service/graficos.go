package service

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/defect"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/metrics"
)

// Graficos arma los datos de los gráficos de métricas de un proyecto (HU-36, spec: specs/HU-36-graficos.md). No
// calcula nada nuevo salvo el burndown: usa el reporte de cada sprint de HU-37 (RN1). Solo lee: no cambia nada (RN6).
type Graficos struct {
	reportes  *Reportes
	defectos  *Defectos
	historias ListadorHistorias
	ahora     func() time.Time
}

// NuevoGraficos arma el caso de uso. ahora da la hora actual (time.Now en la aplicación, fija en los tests).
func NuevoGraficos(reportes *Reportes, defectos *Defectos, historias ListadorHistorias, ahora func() time.Time) *Graficos {
	return &Graficos{reportes: reportes, defectos: defectos, historias: historias, ahora: ahora}
}

// GraficoVelocidad son los SP completados de cada sprint cerrado y la velocidad hasta ese sprint (M2 y M3).
type GraficoVelocidad struct {
	Sprints     []string  `json:"sprints"`
	Completados []float64 `json:"completados"`
	Velocidad   []float64 `json:"velocidad"`
}

// GraficoStoryPoints son los SP planificados y completados de cada sprint (M1 y M2).
type GraficoStoryPoints struct {
	Sprints      []string  `json:"sprints"`
	Planificados []float64 `json:"planificados"`
	Completados  []float64 `json:"completados"`
}

// GraficoHoras son las horas estimadas y reales de cada sprint (M4 y M5).
type GraficoHoras struct {
	Sprints   []string  `json:"sprints"`
	Estimadas []float64 `json:"estimadas"`
	Reales    []float64 `json:"reales"`
}

// GraficoDefectos son los defectos detectados (M8) y los abiertos de cada severidad.
type GraficoDefectos struct {
	Severidades []string `json:"severidades"`
	Detectados  []int    `json:"detectados"`
	Abiertos    []int    `json:"abiertos"`
}

// GraficoBurndown es el burndown de un sprint: por día, la línea ideal y los SP pendientes (nil si el día no pasó).
type GraficoBurndown struct {
	Sprint     string     `json:"sprint"`
	Numero     int        `json:"-"`
	Dias       []string   `json:"dias"`
	Ideal      []float64  `json:"ideal"`
	Pendientes []*float64 `json:"pendientes"`
}

// DatosGraficos son los datos de los cinco gráficos (CA-36.1), como los devuelve el endpoint JSON (CA-36.2).
type DatosGraficos struct {
	Proyecto    project.Proyecto   `json:"-"`
	Sprints     []int              `json:"-"` // los números de los sprints, para elegir el del burndown
	Velocidad   GraficoVelocidad   `json:"velocidad"`
	StoryPoints GraficoStoryPoints `json:"storyPoints"`
	Horas       GraficoHoras       `json:"horas"`
	Defectos    GraficoDefectos    `json:"defectos"`
	Burndown    *GraficoBurndown   `json:"burndown"` // nil si el proyecto no tiene sprints (RN5)
}

// Datos devuelve los datos de los gráficos del proyecto, con el burndown del Sprint numeroSprint (o, si no existe, el
// del sprint activo o el último, RN5). Si el proyecto no existe, el error es project.ErrNoEncontrado.
func (g *Graficos) Datos(ctx context.Context, proyectoID int64, numeroSprint int) (DatosGraficos, error) {
	rep, err := g.reportes.Proyecto(ctx, proyectoID)
	if err != nil {
		return DatosGraficos{}, err
	}
	d := DatosGraficos{
		Proyecto:    rep.Resumen.Proyecto,
		Sprints:     []int{},
		Velocidad:   GraficoVelocidad{Sprints: []string{}, Completados: []float64{}, Velocidad: []float64{}},
		StoryPoints: GraficoStoryPoints{Sprints: []string{}, Planificados: []float64{}, Completados: []float64{}},
		Horas:       GraficoHoras{Sprints: []string{}, Estimadas: []float64{}, Reales: []float64{}},
	}
	for _, s := range rep.Sprints {
		nombre := fmt.Sprintf("Sprint %d", s.Sprint.Numero)
		d.Sprints = append(d.Sprints, s.Sprint.Numero)
		d.StoryPoints.Sprints = append(d.StoryPoints.Sprints, nombre)
		d.StoryPoints.Planificados = append(d.StoryPoints.Planificados, s.SP.Planificados)
		d.StoryPoints.Completados = append(d.StoryPoints.Completados, s.SP.Completados)
		d.Horas.Sprints = append(d.Horas.Sprints, nombre)
		d.Horas.Estimadas = append(d.Horas.Estimadas, s.Horas.Estimadas)
		d.Horas.Reales = append(d.Horas.Reales, s.Horas.Reales)
	}
	for _, s := range rep.Evolucion() {
		d.Velocidad.Sprints = append(d.Velocidad.Sprints, fmt.Sprintf("Sprint %d", s.Sprint.Numero))
		d.Velocidad.Completados = append(d.Velocidad.Completados, s.SP.Completados)
		d.Velocidad.Velocidad = append(d.Velocidad.Velocidad, unDecimalNumero(s.Velocidad.Valor))
	}
	if d.Defectos, err = g.porSeveridad(ctx, proyectoID); err != nil {
		return DatosGraficos{}, err
	}
	if d.Burndown, err = g.burndown(ctx, rep, numeroSprint); err != nil {
		return DatosGraficos{}, err
	}
	return d, nil
}

// porSeveridad cuenta los defectos detectados y los abiertos (ni Resueltos ni Cerrados) de cada severidad.
func (g *Graficos) porSeveridad(ctx context.Context, proyectoID int64) (GraficoDefectos, error) {
	v, err := g.defectos.Ver(ctx, proyectoID)
	if err != nil {
		return GraficoDefectos{}, err
	}
	gd := GraficoDefectos{Severidades: make([]string, len(defect.Severidades)), Detectados: make([]int, len(defect.Severidades)), Abiertos: make([]int, len(defect.Severidades))}
	for i, sev := range defect.Severidades {
		gd.Severidades[i] = string(sev)
		for _, d := range v.Defectos {
			if d.Severidad != sev {
				continue
			}
			gd.Detectados[i]++
			if d.Estado != defect.EstadoResuelto && d.Estado != defect.EstadoCerrado {
				gd.Abiertos[i]++
			}
		}
	}
	return gd, nil
}

// burndown arma el burndown del sprint elegido (RN5) con las fechas en que se completó cada historia (HU-13).
func (g *Graficos) burndown(ctx context.Context, rep ReporteProyecto, numeroSprint int) (*GraficoBurndown, error) {
	if len(rep.Sprints) == 0 {
		return nil, nil
	}
	hoy := domain.HoyEnArgentina(g.ahora())
	elegido := rep.Sprints[len(rep.Sprints)-1]
	sprints := make([]sprint.Sprint, len(rep.Sprints))
	for i, s := range rep.Sprints {
		sprints[i] = s.Sprint
	}
	if activo, ok := sprint.Activo(sprints, hoy); ok {
		elegido = sprintDelReporte(rep, activo.Numero, elegido)
	}
	elegido = sprintDelReporte(rep, numeroSprint, elegido)

	historias, err := g.historias.ListarPorProyecto(ctx, rep.Resumen.Proyecto.ID)
	if err != nil {
		return nil, fmt.Errorf("listar las historias del proyecto %d: %w", rep.Resumen.Proyecto.ID, err)
	}
	completadaEn := map[int]time.Time{}
	for _, h := range historias {
		if h.CompletadaEn != nil {
			completadaEn[h.Numero] = *h.CompletadaEn
		}
	}
	var completadas []metrics.HistoriaCompletada
	for _, f := range elegido.Historias {
		if f.Completada && f.StoryPoints != nil {
			fecha, ok := completadaEn[f.Numero]
			if !ok {
				fecha = elegido.Sprint.FechaFin // completada sin fecha guardada: cuenta el último día
			}
			completadas = append(completadas, metrics.HistoriaCompletada{SP: float64(*f.StoryPoints), Fecha: fecha})
		}
	}
	dias, err := metrics.CalcularBurndown(elegido.Sprint.FechaInicio, elegido.Sprint.FechaFin, elegido.SP.Planificados, completadas, hoy)
	if err != nil {
		return nil, fmt.Errorf("burndown del Sprint %d: %w", elegido.Sprint.Numero, err)
	}
	b := &GraficoBurndown{Sprint: fmt.Sprintf("Sprint %d", elegido.Sprint.Numero), Numero: elegido.Sprint.Numero}
	for _, d := range dias {
		b.Dias = append(b.Dias, d.Fecha.Format("02/01"))
		b.Ideal = append(b.Ideal, unDecimalNumero(d.Ideal))
		var pendientes *float64
		if !d.Futuro {
			p := d.Pendientes
			pendientes = &p
		}
		b.Pendientes = append(b.Pendientes, pendientes)
	}
	return b, nil
}

// sprintDelReporte devuelve el reporte del Sprint numero, o el que se pasa si no existe.
func sprintDelReporte(rep ReporteProyecto, numero int, siNo ReporteSprint) ReporteSprint {
	for _, s := range rep.Sprints {
		if s.Sprint.Numero == numero {
			return s
		}
	}
	return siNo
}

// unDecimalNumero redondea a 1 decimal (RG-2).
func unDecimalNumero(v float64) float64 {
	return math.Round(v*10) / 10
}

// FilaGrafico es una fila de la tabla "Ver los datos" de un gráfico: el nombre y sus valores ya escritos ("6,7", o
// "—" si el día no pasó).
type FilaGrafico struct {
	Nombre  string
	Valores []string
}

// FilasStoryPoints devuelve, por sprint, los SP planificados y completados.
func (d DatosGraficos) FilasStoryPoints() []FilaGrafico {
	return filasDe(d.StoryPoints.Sprints, d.StoryPoints.Planificados, d.StoryPoints.Completados)
}

// FilasVelocidad devuelve, por sprint cerrado, los SP completados y la velocidad hasta ese sprint.
func (d DatosGraficos) FilasVelocidad() []FilaGrafico {
	return filasDe(d.Velocidad.Sprints, d.Velocidad.Completados, d.Velocidad.Velocidad)
}

// FilasHoras devuelve, por sprint, las horas estimadas y reales.
func (d DatosGraficos) FilasHoras() []FilaGrafico {
	return filasDe(d.Horas.Sprints, d.Horas.Estimadas, d.Horas.Reales)
}

// FilasDefectos devuelve, por severidad, los defectos detectados y los abiertos.
func (d DatosGraficos) FilasDefectos() []FilaGrafico {
	filas := make([]FilaGrafico, len(d.Defectos.Severidades))
	for i, sev := range d.Defectos.Severidades {
		filas[i] = FilaGrafico{Nombre: sev, Valores: []string{strconv.Itoa(d.Defectos.Detectados[i]), strconv.Itoa(d.Defectos.Abiertos[i])}}
	}
	return filas
}

// FilasBurndown devuelve, por día, la línea ideal y los SP pendientes ("—" si el día no pasó).
func (d DatosGraficos) FilasBurndown() []FilaGrafico {
	if d.Burndown == nil {
		return nil
	}
	filas := make([]FilaGrafico, len(d.Burndown.Dias))
	for i, dia := range d.Burndown.Dias {
		pendientes := "—"
		if p := d.Burndown.Pendientes[i]; p != nil {
			pendientes = effort.FormatearHoras(*p)
		}
		filas[i] = FilaGrafico{Nombre: dia, Valores: []string{effort.FormatearHoras(d.Burndown.Ideal[i]), pendientes}}
	}
	return filas
}

func filasDe(nombres []string, columnas ...[]float64) []FilaGrafico {
	filas := make([]FilaGrafico, len(nombres))
	for i, n := range nombres {
		filas[i] = FilaGrafico{Nombre: n}
		for _, c := range columnas {
			filas[i].Valores = append(filas[i].Valores, effort.FormatearHoras(c[i]))
		}
	}
	return filas
}
