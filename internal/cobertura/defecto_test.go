package cobertura

import "testing"

func TestPaquetesPorDefecto_SonElDominioYLasMetricas(t *testing.T) {
	got := PaquetesPorDefecto()

	if len(got) != 2 || got[0] != "internal/domain" || got[1] != "internal/metrics" {
		t.Errorf("PaquetesPorDefecto() = %v, se esperaban internal/domain e internal/metrics", got)
	}
}

func TestPaquetesPorDefecto_DevuelveUnaCopiaQueNoSePuedePisar(t *testing.T) {
	got := PaquetesPorDefecto()
	got[0] = "otro"

	if otra := PaquetesPorDefecto(); otra[0] != "internal/domain" {
		t.Errorf("modificar el resultado cambió los paquetes por defecto: %v", otra)
	}
}

func TestMinimoPorDefecto_EsElOchentaPorCiento(t *testing.T) {
	if MinimoPorDefecto != 80 {
		t.Errorf("MinimoPorDefecto = %v, se esperaba 80 (meta interna del equipo, Definition of Done)", MinimoPorDefecto)
	}
}
