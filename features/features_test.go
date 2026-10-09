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
	registrarPasosHU23(sc)
	registrarPasosHU30(sc)
	registrarPasosHU31(sc)
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
