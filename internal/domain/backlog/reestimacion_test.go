package backlog_test

import (
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
)

// cambioDeEstimacion es la hora de ejemplo de los cambios de estos tests.
var cambioDeEstimacion = time.Date(2026, 10, 20, 18, 30, 0, 0, time.UTC)

// enteroPtr devuelve un puntero a puntos, como el que guarda Historia.StoryPoints.
func enteroPtr(puntos int) *int { return &puntos }

// CA-16.3 y RN1 · re-estimar con un valor distinto anota el valor anterior, el nuevo y cuándo cambió
func TestNuevaReestimacion_ConUnValorDistinto_AnotaElCambio(t *testing.T) {
	entrada, anota := backlog.NuevaReestimacion(enteroPtr(5), 8, cambioDeEstimacion)

	if !anota {
		t.Fatal("se esperaba que el cambio de 5 a 8 se anotara")
	}
	if entrada.Anterior != 5 || entrada.Nuevo != 8 {
		t.Errorf("se esperaba 5 → 8 y quedó %d → %d", entrada.Anterior, entrada.Nuevo)
	}
	if !entrada.CambiadaEn.Equal(cambioDeEstimacion) {
		t.Errorf("se esperaba la hora %v y quedó %v", cambioDeEstimacion, entrada.CambiadaEn)
	}
}

// RN1 · cualquier par de valores distintos de la escala se anota
func TestNuevaReestimacion_TodosLosParesDistintosDeLaEscala_Anotan(t *testing.T) {
	for _, antes := range backlog.EscalaStoryPoints {
		for _, despues := range backlog.EscalaStoryPoints {
			if antes == despues {
				continue
			}

			entrada, anota := backlog.NuevaReestimacion(enteroPtr(antes), despues, cambioDeEstimacion)

			if !anota {
				t.Errorf("%d → %d: se esperaba que se anotara", antes, despues)
				continue
			}
			if entrada.Anterior != antes || entrada.Nuevo != despues {
				t.Errorf("%d → %d: quedó %d → %d", antes, despues, entrada.Anterior, entrada.Nuevo)
			}
		}
	}
}

// RN2 · la primera estimación no tiene valor anterior: no se anota
func TestNuevaReestimacion_PrimeraEstimacion_NoSeAnota(t *testing.T) {
	for _, puntos := range backlog.EscalaStoryPoints {
		if _, anota := backlog.NuevaReestimacion(nil, puntos, cambioDeEstimacion); anota {
			t.Errorf("de «sin estimar» a %d: no se esperaba que se anotara", puntos)
		}
	}
}

// RN3 · pedir el mismo valor que ya tiene no cambia nada: no se anota
func TestNuevaReestimacion_ElMismoValor_NoSeAnota(t *testing.T) {
	for _, puntos := range backlog.EscalaStoryPoints {
		if _, anota := backlog.NuevaReestimacion(enteroPtr(puntos), puntos, cambioDeEstimacion); anota {
			t.Errorf("%d → %d: no se esperaba que se anotara", puntos, puntos)
		}
	}
}

// RN4 · 0 es una estimación: pasar de 0 a 5, o de 5 a 0, se anota
func TestNuevaReestimacion_ElCeroEsUnValor_Anota(t *testing.T) {
	casos := []struct{ antes, despues int }{{0, 5}, {5, 0}}
	for _, c := range casos {
		entrada, anota := backlog.NuevaReestimacion(enteroPtr(c.antes), c.despues, cambioDeEstimacion)

		if !anota {
			t.Errorf("%d → %d: se esperaba que se anotara", c.antes, c.despues)
			continue
		}
		if entrada.Anterior != c.antes || entrada.Nuevo != c.despues {
			t.Errorf("%d → %d: quedó %d → %d", c.antes, c.despues, entrada.Anterior, entrada.Nuevo)
		}
	}
}

// RN1 · la función no toca la estimación anterior que recibe
func TestNuevaReestimacion_NoCambiaLaEstimacionAnterior(t *testing.T) {
	anterior := enteroPtr(5)

	_, _ = backlog.NuevaReestimacion(anterior, 8, cambioDeEstimacion)

	if *anterior != 5 {
		t.Errorf("la estimación anterior tenía que seguir en 5 y quedó en %d", *anterior)
	}
}
