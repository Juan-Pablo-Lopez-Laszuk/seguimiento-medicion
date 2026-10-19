package metrics

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Desviacion es el resultado de la métrica M6: cuánto se equivocó la estimación de horas (HU-33).
type Desviacion struct {
	Estimadas  float64 // horas estimadas (M4)
	Reales     float64 // horas reales (M5)
	Horas      float64 // Reales − Estimadas: positivo = se tardó más de lo estimado
	Porcentaje float64 // Horas / Estimadas × 100; solo vale si Aplica
	Aplica     bool    // false si no hay horas estimadas: el % es "no aplicable" (CA-33.2)
}

// CalcularDesviacion compara las horas estimadas con las reales (CA-33.1). Sin horas estimadas el % no aplica.
func CalcularDesviacion(estimadas, reales float64) Desviacion {
	d := Desviacion{Estimadas: estimadas, Reales: reales, Horas: reales - estimadas}
	if estimadas > 0 {
		d.Aplica = true
		d.Porcentaje = d.Horas / estimadas * 100
	}
	return d
}

// SumarDesviaciones devuelve la desviación de un sprint o del proyecto: suma las horas de sus historias y recién ahí
// calcula el % (CA-33.3), no promedia los % de cada historia.
func SumarDesviaciones(partes []Desviacion) Desviacion {
	var estimadas, reales float64
	for _, p := range partes {
		estimadas += p.Estimadas
		reales += p.Reales
	}
	return CalcularDesviacion(estimadas, reales)
}

// HorasTexto devuelve la desviación en horas con signo, como se muestra en pantalla ("+2 h", "-1,5 h", "0 h").
func (d Desviacion) HorasTexto() string {
	h := math.Round(d.Horas*100)/100 + 0 // + 0 deja -0 como 0, así no se muestra "-0 h"
	texto := strings.Replace(strconv.FormatFloat(h, 'f', -1, 64), ".", ",", 1)
	if h > 0 {
		texto = "+" + texto
	}
	return texto + " h"
}

// PorcentajeTexto devuelve la desviación en %, con signo y 1 decimal ("+20,0 %"), o "no aplicable" (CA-33.2).
func (d Desviacion) PorcentajeTexto() string {
	if !d.Aplica {
		return "no aplicable"
	}
	p := math.Round(d.Porcentaje*10)/10 + 0
	texto := strings.Replace(fmt.Sprintf("%.1f %%", p), ".", ",", 1)
	if p > 0 {
		texto = "+" + texto
	}
	return texto
}
