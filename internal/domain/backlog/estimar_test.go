package backlog_test

import (
	"errors"
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

// CA-16.2 y RN2 · un valor fuera de la escala, o que no es un número entero, es un error del campo story_points
func TestEstimar_ValorInvalido_EsUnErrorDelCampo(t *testing.T) {
	for _, texto := range []string{"4", "22", "-1", "3,5", "3.5", "abc", "5 puntos", "99999999999999999999"} {
		_, err := historiaPendiente().Estimar(texto)

		verificarErrorDeCampo(t, err, backlog.CampoStoryPoints, backlog.ErrStoryPointsInvalidos)
	}
}

// RN3 · el valor es obligatorio: vacío no deja la historia "sin estimar" otra vez
func TestEstimar_Vacio_EsUnError(t *testing.T) {
	for _, texto := range []string{"", "   "} {
		_, err := historiaEstimadaEn(5).Estimar(texto)

		verificarErrorDeCampo(t, err, backlog.CampoStoryPoints, backlog.ErrStoryPointsInvalidos)
	}
}

// RN7 · si el valor no vale, la historia original conserva su estimación
func TestEstimar_ValorInvalido_ConservaLaEstimacionAnterior(t *testing.T) {
	original := historiaEstimadaEn(5)

	if _, err := original.Estimar("4"); err == nil {
		t.Fatal("se esperaba un error")
	}

	if original.StoryPoints == nil || *original.StoryPoints != 5 {
		t.Errorf("la historia tenía que seguir con 5 story points y quedó con %v", original.StoryPoints)
	}
}

// RN6 · una historia Hecha no se re-estima: los story points completados de su sprint cambiarían después de terminada
func TestEstimar_HistoriaHecha_NoSeReestima(t *testing.T) {
	hecha := historiaEn(backlog.EstadoHecho)
	puntos := 5
	hecha.StoryPoints = &puntos

	nueva, err := hecha.Estimar("8")

	if !errors.Is(err, backlog.ErrEstimacionDeHistoriaHecha) {
		t.Fatalf("se esperaba backlog.ErrEstimacionDeHistoriaHecha y se obtuvo: %v", err)
	}
	if err.Error() != "la historia está Hecha: no se puede cambiar su estimación" {
		t.Errorf("mensaje = %q", err.Error())
	}
	if nueva.StoryPoints != nil || *hecha.StoryPoints != 5 {
		t.Errorf("la historia Hecha no tenía que cambiar: %+v", nueva)
	}
}

// RN6 · el estado se revisa antes que el valor: no tiene sentido corregir un valor que igual no se va a poder guardar
func TestEstimar_HistoriaHechaConValorInvalido_InformaQueEstaHecha(t *testing.T) {
	_, err := historiaEn(backlog.EstadoHecho).Estimar("4")

	if !errors.Is(err, backlog.ErrEstimacionDeHistoriaHecha) {
		t.Errorf("se esperaba backlog.ErrEstimacionDeHistoriaHecha y se obtuvo: %v", err)
	}
}

// RN6 · en cualquier otro estado se puede estimar
func TestEstimar_PendienteEnSprintOEnProgreso_Admiten(t *testing.T) {
	for _, estado := range []backlog.Estado{backlog.EstadoPendiente, backlog.EstadoEnSprint, backlog.EstadoEnProgreso} {
		if _, err := historiaEn(estado).Estimar("8"); err != nil {
			t.Errorf("estado %q: no se esperaba error y se obtuvo %v", estado, err)
		}
	}
}
