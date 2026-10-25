package metrics_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/metrics"
)

func octubre(d int) time.Time { return time.Date(2026, 10, d, 0, 0, 0, 0, time.UTC) }

// pendientesDe devuelve los SP pendientes de cada día, con -1 para los días que todavía no pasaron.
func pendientesDe(dias []metrics.DiaBurndown) []float64 {
	p := make([]float64, len(dias))
	for i, d := range dias {
		p[i] = -1
		if !d.Futuro {
			p[i] = d.Pendientes
		}
	}
	return p
}

func iguales(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if math.Abs(a[i]-b[i]) > 1e-9 {
			return false
		}
	}
	return true
}

// HU-36 · RN2 a RN4 · un día por fecha del sprint; pendientes = planificados − completados hasta ese día; la línea
// ideal baja recta a 0; los días que no pasaron no tienen pendientes
func TestCalcularBurndown_SprintEnCurso(t *testing.T) {
	completadas := []metrics.HistoriaCompletada{{SP: 8, Fecha: octubre(13)}, {SP: 2, Fecha: octubre(15)}}

	dias, err := metrics.CalcularBurndown(octubre(12), octubre(18), 13, completadas, octubre(15))
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if len(dias) != 7 || !dias[0].Fecha.Equal(octubre(12)) || !dias[6].Fecha.Equal(octubre(18)) {
		t.Fatalf("se esperaban los 7 días del 12 al 18 y se obtuvo %+v", dias)
	}
	if got, quiero := pendientesDe(dias), []float64{13, 5, 5, 3, -1, -1, -1}; !iguales(got, quiero) {
		t.Errorf("pendientes = %v, se esperaba %v", got, quiero)
	}
	ideal := make([]float64, len(dias))
	for i, d := range dias {
		ideal[i] = math.Round(d.Ideal*10) / 10
	}
	if quiero := []float64{13, 10.8, 8.7, 6.5, 4.3, 2.2, 0}; !iguales(ideal, quiero) {
		t.Errorf("ideal = %v, se esperaba %v", ideal, quiero)
	}
}

// RN2 · una historia completada antes de empezar cuenta desde el primer día, la hora del día no importa y los
// pendientes no bajan de 0; un sprint que ya terminó tiene todos los días con dato
func TestCalcularBurndown_CasosLimite(t *testing.T) {
	completadas := []metrics.HistoriaCompletada{
		{SP: 3, Fecha: octubre(1)},
		{SP: 5, Fecha: octubre(6).Add(20 * time.Hour)},
		{SP: 8, Fecha: octubre(7)},
	}

	dias, err := metrics.CalcularBurndown(octubre(5), octubre(7), 10, completadas, octubre(30))
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if got, quiero := pendientesDe(dias), []float64{7, 2, 0}; !iguales(got, quiero) {
		t.Errorf("pendientes = %v, se esperaba %v", got, quiero)
	}
}

// RN3 · un sprint de un solo día tiene el ideal en 0; todavía no empezado, ningún día tiene dato
func TestCalcularBurndown_UnSoloDiaYSinEmpezar(t *testing.T) {
	dias, err := metrics.CalcularBurndown(octubre(5), octubre(5), 5, nil, octubre(5))
	if err != nil || len(dias) != 1 || dias[0].Ideal != 0 || dias[0].Pendientes != 5 || dias[0].Futuro {
		t.Errorf("se esperaba un día con ideal 0 y 5 pendientes y se obtuvo %+v, %v", dias, err)
	}

	dias, err = metrics.CalcularBurndown(octubre(12), octubre(13), 5, nil, octubre(10))
	if err != nil || !dias[0].Futuro || !dias[1].Futuro {
		t.Errorf("antes de empezar ningún día tiene dato, y se obtuvo %+v, %v", dias, err)
	}
}

// RG-4 · datos inválidos devuelven un error con nombre y ningún resultado
func TestCalcularBurndown_Errores(t *testing.T) {
	casos := []struct {
		nombre      string
		inicio, fin time.Time
		planif      float64
		completadas []metrics.HistoriaCompletada
		err         error
	}{
		{"termina antes de empezar", octubre(12), octubre(11), 5, nil, metrics.ErrSprintSinDias},
		{"planificados negativos", octubre(12), octubre(18), -1, nil, metrics.ErrSPNegativos},
		{"una completada con SP negativos", octubre(12), octubre(18), 5, []metrics.HistoriaCompletada{{SP: -2, Fecha: octubre(12)}}, metrics.ErrSPNegativos},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			dias, err := metrics.CalcularBurndown(c.inicio, c.fin, c.planif, c.completadas, octubre(15))
			if !errors.Is(err, c.err) || dias != nil {
				t.Errorf("se esperaba %v y ningún día, y se obtuvo %+v, %v", c.err, dias, err)
			}
		})
	}
}
