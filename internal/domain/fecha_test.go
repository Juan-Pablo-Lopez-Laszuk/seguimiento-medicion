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
