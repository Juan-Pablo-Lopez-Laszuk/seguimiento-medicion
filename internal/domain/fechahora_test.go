package domain_test

import (
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
)

// HU-16b · la hora de un cambio se muestra en hora de Argentina (UTC−3): Vercel corre en UTC
func TestFechaHoraEnArgentina(t *testing.T) {
	casos := []struct {
		nombre string
		hora   time.Time
		quiere string
	}{
		{"de tarde", time.Date(2026, 10, 20, 21, 30, 0, 0, time.UTC), "20/10/2026 18:30"},
		{"cruzando la medianoche de UTC cuenta como el día anterior", time.Date(2026, 10, 21, 1, 15, 0, 0, time.UTC), "20/10/2026 22:15"},
		{"hora con minutos de un dígito", time.Date(2026, 11, 1, 12, 5, 0, 0, time.UTC), "01/11/2026 09:05"},
		{"una hora que ya viene con otra zona", time.Date(2026, 10, 20, 18, 30, 0, 0, time.FixedZone("ART", -3*60*60)), "20/10/2026 18:30"},
		{"medianoche en Argentina", time.Date(2026, 10, 20, 3, 0, 0, 0, time.UTC), "20/10/2026 00:00"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if obtenido := domain.FechaHoraEnArgentina(c.hora); obtenido != c.quiere {
				t.Errorf("se esperaba %q y se obtuvo %q", c.quiere, obtenido)
			}
		})
	}
}
