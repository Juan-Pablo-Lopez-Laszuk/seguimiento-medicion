package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const modulo = "github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/"

// perfilConCobertura: internal/domain/backlog 141 sentencias (138 cubiertas), internal/metrics 89 (todas) y un
// paquete que no se controla (internal/server) con poca cobertura.
const perfilConCobertura = "mode: atomic\n" +
	modulo + "internal/domain/backlog/historia.go:10.2,12.3 138 1\n" +
	modulo + "internal/domain/backlog/historia.go:14.2,16.3 3 0\n" +
	modulo + "internal/metrics/velocity.go:1.2,3.3 89 4\n" +
	modulo + "internal/server/render.go:1.2,3.3 200 0\n"

// perfilBajo: 79 de 100 sentencias de internal/metrics cubiertas.
const perfilBajo = "mode: atomic\n" +
	modulo + "internal/metrics/velocity.go:1.2,3.3 79 1\n" +
	modulo + "internal/metrics/velocity.go:5.2,7.3 21 0\n"

// escribirPerfil guarda el perfil en una carpeta temporal y devuelve su ruta.
func escribirPerfil(t *testing.T, contenido string) string {
	t.Helper()
	ruta := filepath.Join(t.TempDir(), "coverage.out")
	if err := os.WriteFile(ruta, []byte(contenido), 0o600); err != nil {
		t.Fatal(err)
	}
	return ruta
}

func TestRun_ConCoberturaSuficienteMuestraElInformeYNoFalla(t *testing.T) {
	ruta := escribirPerfil(t, perfilConCobertura)
	var salida bytes.Buffer

	err := run([]string{"-perfil", ruta}, &salida)

	if err != nil {
		t.Fatalf("run() = %v, se esperaba que cumpliera", err)
	}
	for _, esperado := range []string{"internal/domain/backlog", "138/141", "internal/metrics", "89/89", "Combinada", "227/230", "98,7 %", "mínimo 80,0 %"} {
		if !strings.Contains(salida.String(), esperado) {
			t.Errorf("el informe no contiene %q:\n%s", esperado, salida.String())
		}
	}
}

func TestRun_NoMideLosPaquetesQueNoSeControlan(t *testing.T) {
	ruta := escribirPerfil(t, perfilConCobertura)
	var salida bytes.Buffer

	_ = run([]string{"-perfil", ruta}, &salida)

	if strings.Contains(salida.String(), "internal/server") {
		t.Errorf("el informe no tendría que incluir internal/server:\n%s", salida.String())
	}
}

func TestRun_ConCoberturaBajaFallaPeroMuestraElInforme(t *testing.T) {
	ruta := escribirPerfil(t, perfilBajo)
	var salida bytes.Buffer

	err := run([]string{"-perfil", ruta}, &salida)

	esperado := "la cobertura combinada de internal/domain e internal/metrics es 79,0 %, por debajo del mínimo de 80,0 %"
	if err == nil || err.Error() != esperado {
		t.Errorf("run() = %v, se esperaba el error %q", err, esperado)
	}
	if !strings.Contains(salida.String(), "79/100") {
		t.Errorf("aunque falle, el informe tiene que mostrarse:\n%s", salida.String())
	}
}

func TestRun_ElUmbralSePuedeCambiar(t *testing.T) {
	ruta := escribirPerfil(t, perfilBajo)

	if err := run([]string{"-umbral", "79", "-perfil", ruta}, &bytes.Buffer{}); err != nil {
		t.Errorf("con umbral 79 la cobertura de 79 %% tendría que cumplir y dio: %v", err)
	}
	err := run([]string{"-umbral", "79.5", "-perfil", ruta}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "por debajo del mínimo de 79,5 %") {
		t.Errorf("con umbral 79.5 la cobertura de 79 %% tendría que fallar por debajo de 79,5 %% y dio: %v", err)
	}
}

func TestRun_ElUmbralFueraDeRangoDaError(t *testing.T) {
	ruta := escribirPerfil(t, perfilConCobertura)

	err := run([]string{"-umbral", "150", "-perfil", ruta}, &bytes.Buffer{})

	if err == nil || !strings.Contains(err.Error(), "el umbral tiene que estar entre 0 y 100") {
		t.Errorf("run() = %v, se esperaba el error del umbral", err)
	}
}

func TestRun_LosPaquetesControladosSePuedenElegir(t *testing.T) {
	ruta := escribirPerfil(t, perfilConCobertura)
	var salida bytes.Buffer

	err := run([]string{"-perfil", ruta, "internal/server"}, &salida)

	if err == nil || !strings.Contains(err.Error(), "la cobertura combinada de internal/server es 0,0 %") {
		t.Errorf("run() = %v, se esperaba que fallara por internal/server", err)
	}
	if strings.Contains(salida.String(), "internal/metrics") {
		t.Errorf("solo se controla internal/server:\n%s", salida.String())
	}
}

func TestRun_SinSentenciasControladasFalla(t *testing.T) {
	ruta := escribirPerfil(t, "mode: atomic\n"+modulo+"internal/server/render.go:1.2,3.3 200 4\n")

	err := run([]string{"-perfil", ruta}, &bytes.Buffer{})

	if err == nil || !strings.Contains(err.Error(), "el perfil no tiene sentencias de los paquetes controlados") {
		t.Errorf("run() = %v, se esperaba el error de que no hay sentencias controladas", err)
	}
}

func TestRun_SinPerfilUsaCoverageOutDeLaCarpetaActual(t *testing.T) {
	carpeta := t.TempDir()
	if err := os.WriteFile(filepath.Join(carpeta, "coverage.out"), []byte(perfilConCobertura), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(carpeta)

	if err := run(nil, &bytes.Buffer{}); err != nil {
		t.Errorf("run() sin argumentos = %v, se esperaba que leyera coverage.out y cumpliera", err)
	}
}

func TestRun_UnPerfilQueNoExisteDaError(t *testing.T) {
	err := run([]string{"-perfil", filepath.Join(t.TempDir(), "no-existe.out")}, &bytes.Buffer{})

	if err == nil || !strings.Contains(err.Error(), "no se pudo abrir el perfil de cobertura") {
		t.Errorf("run() = %v, se esperaba el error de que no se pudo abrir el perfil", err)
	}
}

func TestRun_UnPerfilMalFormadoDaError(t *testing.T) {
	ruta := escribirPerfil(t, "esto no es un perfil\n")

	err := run([]string{"-perfil", ruta}, &bytes.Buffer{})

	if err == nil || !strings.HasPrefix(err.Error(), "perfil de cobertura:") {
		t.Errorf("run() = %v, se esperaba un error del perfil de cobertura", err)
	}
}

func TestRun_UnaOpcionDesconocidaMuestraLaAyuda(t *testing.T) {
	err := run([]string{"-rapido"}, &bytes.Buffer{})

	if err == nil || !strings.Contains(err.Error(), "Uso:") {
		t.Errorf("run() = %v, se esperaba un error con la ayuda", err)
	}
}

func TestRun_PedirAyudaLaMuestraYNoFalla(t *testing.T) {
	var salida bytes.Buffer

	err := run([]string{"-h"}, &salida)

	if err != nil || !strings.Contains(salida.String(), "Uso:") {
		t.Errorf("run(-h) = %v con salida %q, se esperaba la ayuda sin error", err, salida.String())
	}
}
