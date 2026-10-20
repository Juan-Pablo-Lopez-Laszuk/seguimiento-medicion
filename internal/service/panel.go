package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/metrics"
)

// sinDatos es lo que muestra una tarjeta del panel cuando todavía no hay con qué calcularla (HU-35 RN2).
const sinDatos = "sin datos"

// Panel arma el panel de indicadores de un proyecto (HU-35) con las métricas que ya calculan otros casos de uso.
// Solo lee: no cambia nada.
type Panel struct {
	estado   *EstadoProyectos
	sprints  ListadorSprints
	horas    *Horas
	defectos *Defectos
}

// NuevoPanel arma el caso de uso.
func NuevoPanel(estado *EstadoProyectos, sprints ListadorSprints, horas *Horas, defectos *Defectos) *Panel {
	return &Panel{estado: estado, sprints: sprints, horas: horas, defectos: defectos}
}

// VistaPanel son los datos del panel: el estado del proyecto (HU-04), la velocidad (M3), la desviación de horas del
// proyecto (M6) y el resumen de defectos (M8 y M9).
type VistaPanel struct {
	EstadoProyecto
	Velocidad metrics.Velocidad
	Horas     metrics.Desviacion
	Defectos  ResumenDefectos
}

// Tarjeta es una tarjeta del panel: Valor es el número grande y Detalle la línea de abajo.
type Tarjeta struct {
	Titulo  string
	Valor   string
	Detalle string
}

// Ver junta las métricas del proyecto, o devuelve project.ErrNoEncontrado.
func (p *Panel) Ver(ctx context.Context, proyectoID int64) (VistaPanel, error) {
	estado, err := p.estado.Consultar(ctx, proyectoID)
	if err != nil {
		return VistaPanel{}, err
	}
	sprints, err := p.sprints.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return VistaPanel{}, fmt.Errorf("listar los sprints del proyecto %d: %w", proyectoID, err)
	}
	velocidad, err := velocidadDe(sprints)
	if err != nil {
		return VistaPanel{}, err
	}
	horas, err := p.horas.Ver(ctx, proyectoID)
	if err != nil {
		return VistaPanel{}, err
	}
	defectos, err := p.defectos.Ver(ctx, proyectoID)
	if err != nil {
		return VistaPanel{}, err
	}
	return VistaPanel{EstadoProyecto: estado, Velocidad: velocidad, Horas: horas.Total, Defectos: defectos.Resumen("")}, nil
}

// Tarjetas devuelve las seis tarjetas del panel (CA-35.1); la que no tiene datos dice "sin datos" (CA-35.2).
func (v VistaPanel) Tarjetas() []Tarjeta {
	return []Tarjeta{
		{Titulo: "Estado", Valor: string(v.Estado), Detalle: fmt.Sprintf("Quedan %d días del proyecto", v.DiasRestantes)},
		v.tarjetaSprint(), v.tarjetaVelocidad(), v.tarjetaCompletado(), v.tarjetaHoras(), v.tarjetaDefectos(),
	}
}

func (v VistaPanel) tarjetaSprint() Tarjeta {
	t := Tarjeta{Titulo: "Sprint activo", Valor: sinDatos, Detalle: "No hay un sprint en curso"}
	if a := v.Activo; a != nil {
		t.Valor, t.Detalle = fmt.Sprintf("Sprint %d", a.Numero), fmt.Sprintf("Le quedan %d días", a.DiasRestantes)
		if a.Terminado {
			t.Detalle = "Terminó: falta cerrarlo"
		}
	}
	return t
}

func (v VistaPanel) tarjetaVelocidad() Tarjeta {
	t := Tarjeta{Titulo: "Velocidad", Valor: sinDatos, Detalle: "Todavía no hay sprints cerrados"}
	if n := v.Velocidad.SprintsUsados; n > 0 {
		t.Valor = numero(v.Velocidad.Valor) + " SP por sprint"
		t.Detalle = fmt.Sprintf("Promedio de los últimos %d sprints cerrados", n)
		if n == 1 {
			t.Detalle = "Del único sprint cerrado"
		}
	}
	return t
}

func (v VistaPanel) tarjetaCompletado() Tarjeta {
	t := Tarjeta{Titulo: "% completado", Valor: sinDatos, Detalle: "Todavía no hay historias estimadas"}
	if sp := v.StoryPoints; sp.Planificados > 0 {
		t.Valor = v.PorcentajeSP()
		t.Detalle = fmt.Sprintf("%s de %s SP", numero(sp.Completados), numero(sp.Planificados))
	}
	return t
}

func (v VistaPanel) tarjetaHoras() Tarjeta {
	t := Tarjeta{Titulo: "Desviación de horas", Valor: sinDatos, Detalle: "Todavía no hay horas estimadas ni trabajadas"}
	if h := v.Horas; h.Estimadas > 0 || h.Reales > 0 {
		t.Valor = h.HorasTexto()
		t.Detalle = fmt.Sprintf("%s (%s h estimadas, %s h reales)", h.PorcentajeTexto(), numero(h.Estimadas), numero(h.Reales))
	}
	return t
}

func (v VistaPanel) tarjetaDefectos() Tarjeta {
	t := Tarjeta{Titulo: "Defectos abiertos", Valor: sinDatos, Detalle: "Todavía no hay defectos registrados"}
	if d := v.Defectos; d.Detectados > 0 {
		t.Valor, t.Detalle = strconv.Itoa(d.Abiertos), fmt.Sprintf("De %d detectados", d.Detectados)
	}
	return t
}

// numero muestra un número con coma decimal y sin ceros de más ("8", "7,5").
func numero(x float64) string {
	return strings.Replace(strconv.FormatFloat(x, 'f', -1, 64), ".", ",", 1)
}

// velocidadDe calcula la velocidad del equipo (M3, HU-31) con los SP completados de los últimos sprints cerrados.
func velocidadDe(sprints []sprint.Sprint) (metrics.Velocidad, error) {
	historial := make([]metrics.Sprint, 0, len(sprints))
	for _, s := range sprints {
		m := metrics.Sprint{Numero: s.Numero, Estado: metrics.SprintPlanificado}
		if s.Estado == sprint.EstadoCerrado {
			m.Estado = metrics.SprintCerrado
		}
		if s.SPCompletados != nil {
			m.SPCompletados = float64(*s.SPCompletados)
		}
		historial = append(historial, m)
	}
	v, err := metrics.CalcularVelocidad(historial, metrics.VentanaPorDefecto)
	if err != nil {
		return metrics.Velocidad{}, fmt.Errorf("calcular la velocidad: %w", err)
	}
	return v, nil
}
