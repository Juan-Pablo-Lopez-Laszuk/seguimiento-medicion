package domain_test

import (
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
)

func TestSoloFecha_QuitaLaHoraYPasaAUTC(t *testing.T) {
	art := time.FixedZone("ART", -3*60*60)
	f := time.Date(2026, 10, 5, 22, 30, 0, 0, art)

	esperado := time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC)
	if got := domain.SoloFecha(f); !got.Equal(esperado) {
		t.Errorf("se esperaba %v y se obtuvo %v", esperado, got)
	}
}

func TestSoloFecha_LaFechaCeroSigueEnCero(t *testing.T) {
	if got := domain.SoloFecha(time.Time{}); !got.IsZero() {
		t.Errorf("se esperaba la fecha cero y se obtuvo %v", got)
	}
}

// fecha arma una fecha sin hora, en UTC, como las que guarda el dominio.
func fecha(t *testing.T, texto string) time.Time {
	t.Helper()
	f, err := time.Parse(time.DateOnly, texto)
	if err != nil {
		t.Fatalf("fecha de prueba inválida %q: %v", texto, err)
	}
	return f
}

// HU-04 RN3 · días corridos hasta el fin, sin contar hoy y nunca negativos
func TestDiasRestantes(t *testing.T) {
	casos := []struct {
		nombre   string
		hoy, fin string
		esperado int
	}{
		{"primer día del sprint", "2026-10-05", "2026-10-11", 6},
		{"día anterior al proyecto", "2026-10-04", "2026-11-01", 28},
		{"el último día da 0", "2026-11-01", "2026-11-01", 0},
		{"una fecha pasada da 0, no negativo", "2026-11-02", "2026-11-01", 0},
		{"cambio de mes", "2026-10-30", "2026-11-02", 3},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := domain.DiasRestantes(fecha(t, c.hoy), fecha(t, c.fin)); got != c.esperado {
				t.Errorf("hoy %s, fin %s: se esperaban %d días y se obtuvieron %d", c.hoy, c.fin, c.esperado, got)
			}
		})
	}
}

// HU-04 RN3 · la hora del día no cambia la cuenta
func TestDiasRestantes_IgnoraLaHora(t *testing.T) {
	hoy := time.Date(2026, 10, 5, 23, 59, 0, 0, time.UTC)
	fin := time.Date(2026, 10, 11, 0, 1, 0, 0, time.UTC)
	if got := domain.DiasRestantes(hoy, fin); got != 6 {
		t.Errorf("se esperaban 6 días y se obtuvieron %d", got)
	}
}

// HU-04 RN6 · "hoy" es la fecha de Argentina (UTC−3), aunque el servidor corra en UTC
func TestHoyEnArgentina(t *testing.T) {
	casos := []struct {
		nombre   string
		ahora    time.Time
		esperado string
	}{
		{"mediodía", time.Date(2026, 11, 1, 15, 0, 0, 0, time.UTC), "2026-11-01"},
		{"22:00 de Argentina: en UTC ya es el día siguiente", time.Date(2026, 11, 2, 1, 0, 0, 0, time.UTC), "2026-11-01"},
		{"medianoche de Argentina", time.Date(2026, 11, 2, 3, 0, 0, 0, time.UTC), "2026-11-02"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			got := domain.HoyEnArgentina(c.ahora)
			if !got.Equal(fecha(t, c.esperado)) {
				t.Errorf("ahora %v: se esperaba %s y se obtuvo %v", c.ahora, c.esperado, got)
			}
		})
	}
}
