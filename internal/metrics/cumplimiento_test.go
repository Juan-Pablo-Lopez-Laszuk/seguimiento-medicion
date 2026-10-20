package metrics_test

import (
	"math"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/metrics"
)

// HU-15 RN2: cumplimiento = SP completados ÷ SP planificados × 100, con 1 decimal y coma para mostrar.
func TestCumplimiento_CompletadosSobrePlanificados(t *testing.T) {
	casos := []struct {
		nombre      string
		sp          metrics.StoryPoints
		quiero      float64
		quieroTexto string
	}{
		{"5 de 13", metrics.StoryPoints{Planificados: 13, Completados: 5}, 5.0 / 13 * 100, "38,5 %"},
		{"6 de 8", metrics.StoryPoints{Planificados: 8, Completados: 6}, 75, "75,0 %"},
		{"todo completado", metrics.StoryPoints{Planificados: 5, Completados: 5}, 100, "100,0 %"},
		{"nada completado", metrics.StoryPoints{Planificados: 13}, 0, "0,0 %"},
		{"periódico", metrics.StoryPoints{Planificados: 3, Completados: 2}, 2.0 / 3 * 100, "66,7 %"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			valor, ok := c.sp.Cumplimiento()
			if !ok {
				t.Fatal("con SP planificados el cumplimiento se tiene que poder calcular")
			}
			if math.Abs(valor-c.quiero) > 1e-9 {
				t.Errorf("cumplimiento = %v, se esperaba %v", valor, c.quiero)
			}
			if texto := c.sp.CumplimientoTexto(); texto != c.quieroTexto {
				t.Errorf("texto = %q, se esperaba %q", texto, c.quieroTexto)
			}
		})
	}
}

// HU-15 RN3: con 0 SP planificados no hay cumplimiento y se muestra "—": nunca se divide por cero.
func TestCumplimiento_SinSPPlanificadosNoDividePorCero(t *testing.T) {
	for _, sp := range []metrics.StoryPoints{{}, {SinEstimar: 2}} {
		valor, ok := sp.Cumplimiento()
		if ok || valor != 0 {
			t.Errorf("Cumplimiento() de %+v = %v, %v; se esperaba 0, false", sp, valor, ok)
		}
		if texto := sp.CumplimientoTexto(); texto != "—" {
			t.Errorf("CumplimientoTexto() de %+v = %q, se esperaba \"—\"", sp, texto)
		}
	}
}
