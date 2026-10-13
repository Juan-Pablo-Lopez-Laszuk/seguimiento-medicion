package domain

import "time"

// SoloFecha deja el día de t a medianoche UTC, así las comparaciones entre fechas no dependen de la
// hora ni de la zona horaria. La fecha cero (campo vacío) se mantiene en cero.
func SoloFecha(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

// zonaArgentina es la hora de Argentina (UTC−3, sin horario de verano). Se arma fija y no con
// time.LoadLocation para no depender de la base de zonas horarias del servidor.
var zonaArgentina = time.FixedZone("ART", -3*60*60)

// HoyEnArgentina devuelve la fecha de hoy en Argentina, a medianoche UTC como el resto de las fechas del
// dominio (HU-04 RN6). Vercel corre en UTC: entre las 21:00 y las 24:00 de Argentina, allá ya es mañana.
func HoyEnArgentina(ahora time.Time) time.Time {
	return SoloFecha(ahora.In(zonaArgentina))
}

// DiasRestantes cuenta los días corridos desde hoy hasta fin, sin contar hoy: el último día da 0 y una
// fecha ya pasada también da 0, nunca un número negativo (HU-04 RN3).
func DiasRestantes(hoy, fin time.Time) int {
	dias := int(SoloFecha(fin).Sub(SoloFecha(hoy)).Hours() / 24)
	return max(dias, 0)
}
