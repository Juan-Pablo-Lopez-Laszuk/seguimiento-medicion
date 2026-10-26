package reportepdf_test

import (
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/reportepdf"
)

// CA-39.1, RN3 · el archivo se llama proyecto-sprint-fecha.pdf, con el nombre del proyecto sin tildes ni espacios y la
// fecha de hoy en la Argentina
func TestNombreArchivo(t *testing.T) {
	tarde := time.Date(2026, 10, 29, 18, 0, 0, 0, time.UTC)         // 15:00 en la Argentina
	casiMedianoche := time.Date(2026, 10, 30, 2, 0, 0, 0, time.UTC) // 23:00 del 29/10 en la Argentina

	casos := []struct {
		nombre   string
		proyecto string
		sprint   int
		ahora    time.Time
		esperado string
	}{
		{"de un sprint", "Software Metrics", 2, tarde, "software-metrics-sprint-2-2026-10-29.pdf"},
		{"del proyecto", "Software Metrics", 0, tarde, "software-metrics-2026-10-29.pdf"},
		{"con la fecha de la Argentina", "Software Metrics", 0, casiMedianoche, "software-metrics-2026-10-29.pdf"},
		{"sin tildes", "Gestión Ágil 2026", 1, tarde, "gestion-agil-2026-sprint-1-2026-10-29.pdf"},
		{"con eñes y signos", "  Año  Ñandú -- v2! ", 0, tarde, "ano-nandu-v2-2026-10-29.pdf"},
		{"sin letras ni números", "¿¡!?", 3, tarde, "proyecto-sprint-3-2026-10-29.pdf"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := reportepdf.NombreArchivo(c.proyecto, c.sprint, c.ahora); got != c.esperado {
				t.Errorf("NombreArchivo(%q, %d) = %q, se esperaba %q", c.proyecto, c.sprint, got, c.esperado)
			}
		})
	}
}
