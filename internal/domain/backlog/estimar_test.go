package backlog_test

import (
	"strconv"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
)

// historiaEstimadaEn devuelve la historia de ejemplo ya estimada con esos story points.
func historiaEstimadaEn(puntos int) backlog.Historia {
	h := historiaPendiente()
	h.StoryPoints = &puntos
	return h
}

// CA-16.1 · todos los valores de la escala valen; la historia nueva queda estimada y la original no cambia
func TestEstimar_ValoresDeLaEscala(t *testing.T) {
	for _, puntos := range backlog.EscalaStoryPoints {
		original := historiaPendiente()

		nueva, err := original.Estimar(strconv.Itoa(puntos))
		if err != nil {
			t.Errorf("%d story points: no se esperaba error: %v", puntos, err)
			continue
		}

		if nueva.StoryPoints == nil || *nueva.StoryPoints != puntos {
			t.Errorf("%d story points: la historia quedó con %v", puntos, nueva.StoryPoints)
		}
		if original.StoryPoints != nil {
			t.Errorf("%d story points: la historia original cambió a %d", puntos, *original.StoryPoints)
		}
	}
}

// RN4 · 0 es una estimación: la historia deja de estar "sin estimar"
func TestEstimar_Cero_CuentaComoEstimada(t *testing.T) {
	nueva, err := historiaPendiente().Estimar("0")

	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if nueva.StoryPoints == nil || *nueva.StoryPoints != 0 {
		t.Errorf("se esperaba la historia estimada en 0 y quedó con %v", nueva.StoryPoints)
	}
}

// RN5 · re-estimar reemplaza el valor anterior, sin tocar el valor de la historia original
func TestEstimar_ReemplazaLaEstimacionAnterior(t *testing.T) {
	original := historiaEstimadaEn(5)

	nueva, err := original.Estimar("8")
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if nueva.StoryPoints == nil || *nueva.StoryPoints != 8 {
		t.Errorf("se esperaban 8 story points y quedó con %v", nueva.StoryPoints)
	}
	if *original.StoryPoints != 5 {
		t.Errorf("la historia original tenía que seguir con 5 y quedó con %d", *original.StoryPoints)
	}
}

// RN5 · pedir el mismo valor que ya tiene no es un error
func TestEstimar_ElMismoValor_NoEsUnError(t *testing.T) {
	nueva, err := historiaEstimadaEn(5).Estimar("5")

	if err != nil || nueva.StoryPoints == nil || *nueva.StoryPoints != 5 {
		t.Errorf("se esperaba que siga con 5 sin error y se obtuvo %v, %v", nueva.StoryPoints, err)
	}
}

// RN8 · estimar solo cambia los story points
func TestEstimar_NoCambiaLosDemasDatos(t *testing.T) {
	original := historiaEn(backlog.EstadoEnProgreso)
	original.Descripcion = "Para no recordar otra contraseña."

	nueva, err := original.Estimar("13")
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	nueva.StoryPoints = nil // lo único que tiene que ser distinto
	if nueva != original {
		t.Errorf("cambió algo más que los story points: antes %+v y después %+v", original, nueva)
	}
}

// RN1 · los espacios alrededor del número no importan (igual que al crear la historia)
func TestEstimar_ConEspaciosAlrededor(t *testing.T) {
	nueva, err := historiaPendiente().Estimar(" 8 ")

	if err != nil || nueva.StoryPoints == nil || *nueva.StoryPoints != 8 {
		t.Errorf("se esperaban 8 story points y se obtuvo %v, %v", nueva.StoryPoints, err)
	}
}
