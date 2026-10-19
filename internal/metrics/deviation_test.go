package metrics

import "testing"

// M6 · CA-33.1 y CA-33.2 · desviación en horas y en %, con "no aplicable" si no hay horas estimadas
func TestCalcularDesviacion(t *testing.T) {
	casos := []struct {
		nombre            string
		estimadas, reales float64
		horas, porcentaje string
		aplica            bool
	}{
		{"se tardó más", 10, 12, "+2 h", "+20,0 %", true},
		{"se tardó menos", 10, 8, "-2 h", "-20,0 %", true},
		{"justo", 5, 5, "0 h", "0,0 %", true},
		{"con decimales", 3, 4.5, "+1,5 h", "+50,0 %", true},
		{"sin estimadas", 0, 3, "+3 h", "no aplicable", false},
		{"sin nada", 0, 0, "0 h", "no aplicable", false},
		{"diferencia mínima", 10, 9.999, "0 h", "0,0 %", true},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d := CalcularDesviacion(c.estimadas, c.reales)
			if d.Estimadas != c.estimadas || d.Reales != c.reales || d.Aplica != c.aplica {
				t.Errorf("CalcularDesviacion(%v, %v) = %+v", c.estimadas, c.reales, d)
			}
			if d.HorasTexto() != c.horas || d.PorcentajeTexto() != c.porcentaje {
				t.Errorf("textos = %q y %q, se esperaba %q y %q", d.HorasTexto(), d.PorcentajeTexto(), c.horas, c.porcentaje)
			}
		})
	}
}

// M6 · CA-33.3 · el sprint y el proyecto suman horas y recién ahí calculan el %, no promedian los %
func TestSumarDesviaciones(t *testing.T) {
	d := SumarDesviaciones([]Desviacion{CalcularDesviacion(10, 12), CalcularDesviacion(5, 4), CalcularDesviacion(0, 1)})
	if d.Estimadas != 15 || d.Reales != 17 || d.HorasTexto() != "+2 h" || d.PorcentajeTexto() != "+13,3 %" {
		t.Errorf("SumarDesviaciones = %+v (%s, %s)", d, d.HorasTexto(), d.PorcentajeTexto())
	}
	if v := SumarDesviaciones(nil); v.Aplica || v.HorasTexto() != "0 h" || v.PorcentajeTexto() != "no aplicable" {
		t.Errorf("sin historias = %+v", v)
	}
}
