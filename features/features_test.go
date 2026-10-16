// Package features ejecuta los escenarios BDD (*.feature, en español) como tests de Go con godog.
//
// Cada historia agrega sus pasos en un archivo propio (por ejemplo hu31_steps_test.go) y los registra
// en InitializeScenario. Mientras un paso no esté implementado, godog lo informa como "undefined".
package features

import (
	"testing"

	"github.com/cucumber/godog"
)

// InitializeScenario registra los pasos de todas las historias.
func InitializeScenario(sc *godog.ScenarioContext) {
	registrarPasosHU01(sc)
	registrarPasosHU02(sc)
	registrarPasosHU03(sc)
	registrarPasosHU04(sc)
	registrarPasosHU05(sc)
	registrarPasosHU06(sc)
	registrarPasosHU07(sc)
	registrarPasosHU08(sc)
	registrarPasosHU11(sc)
	registrarPasosHU23(sc)
	registrarPasosHU25(sc)
	registrarPasosHU30(sc)
	registrarPasosHU31(sc)
	registrarPasosHU32(sc)
	registrarPasosTEC04(sc)
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"."},
			Tags:     "~@HU-XX", // excluye la plantilla (features/_plantilla.feature)
			TestingT: t,
		},
	}
	if suite.Run() != 0 {
		t.Fatal("fallaron escenarios BDD")
	}
}
