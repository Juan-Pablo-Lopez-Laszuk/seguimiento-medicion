package metrics

import (
	"errors"
	"time"
)

// ErrSprintSinDias es el error del burndown de un sprint que termina antes de empezar.
var ErrSprintSinDias = errors.New("el sprint tiene que terminar el mismo día o después de empezar")

// HistoriaCompletada es una historia del sprint que se completó: sus SP y el día en que pasó a Hecho (HU-13).
type HistoriaCompletada struct {
	SP    float64
	Fecha time.Time
}

// DiaBurndown es un día del burndown: la línea ideal y los SP que faltaban al final del día. Futuro es true si el día
// todavía no pasó: entonces Pendientes no tiene valor (HU-36 RN4).
type DiaBurndown struct {
	Fecha      time.Time
	Ideal      float64
	Pendientes float64
	Futuro     bool
}

// CalcularBurndown devuelve un día por cada fecha del sprint, de inicio a fin (HU-36, docs/metricas.md sección 5):
// los SP pendientes son los planificados menos los de las historias completadas hasta ese día, sin bajar de 0 (RN2), y
// la línea ideal baja en línea recta de los planificados a 0 (RN3). Los días después de hoy son futuros (RN4). Solo
// cuenta la fecha, no la hora.
func CalcularBurndown(inicio, fin time.Time, planificados float64, completadas []HistoriaCompletada, hoy time.Time) ([]DiaBurndown, error) {
	inicio, fin, hoy = soloFecha(inicio), soloFecha(fin), soloFecha(hoy)
	if fin.Before(inicio) {
		return nil, ErrSprintSinDias
	}
	if planificados < 0 {
		return nil, ErrSPNegativos
	}
	for _, c := range completadas {
		if c.SP < 0 {
			return nil, ErrSPNegativos
		}
	}

	cantidad := int(fin.Sub(inicio).Hours()/24) + 1
	dias := make([]DiaBurndown, cantidad)
	for i := range dias {
		dia := inicio.AddDate(0, 0, i)
		d := DiaBurndown{Fecha: dia, Futuro: dia.After(hoy)}
		if cantidad > 1 {
			d.Ideal = planificados * float64(cantidad-1-i) / float64(cantidad-1)
		}
		if !d.Futuro {
			d.Pendientes = planificados
			for _, c := range completadas {
				if !soloFecha(c.Fecha).After(dia) {
					d.Pendientes -= c.SP
				}
			}
			d.Pendientes = max(d.Pendientes, 0)
		}
		dias[i] = d
	}
	return dias, nil
}

// soloFecha deja solo el día, sin la hora.
func soloFecha(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
