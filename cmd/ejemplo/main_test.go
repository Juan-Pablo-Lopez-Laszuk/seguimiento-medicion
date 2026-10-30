package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/ejemplo"
)

// sinVariables simula un entorno sin variables definidas.
func sinVariables(string) string { return "" }

// Sin opciones devuelve la aplicación en memoria con el ejemplo ya cargado, lista para servir.
func TestPreparar_EnMemoriaDevuelveLaAppConElEjemplo(t *testing.T) {
	var salida bytes.Buffer

	h, err := preparar(nil, sinVariables, &salida)

	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if h == nil {
		t.Fatal("se esperaba la aplicación para servir")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/proyectos", nil))
	if !strings.Contains(rec.Body.String(), ejemplo.NombrePorDefecto) {
		t.Errorf("la lista de proyectos tendría que mostrar el ejemplo:\n%s", rec.Body.String())
	}
	if !strings.Contains(salida.String(), "/proyectos/1") {
		t.Errorf("tendría que decir dónde quedó el proyecto: %q", salida.String())
	}
}

// -base sin DATABASE_URL no hace nada: así nunca se carga en memoria creyendo que se cargó en la base.
func TestPreparar_BaseSinURLEsUnError(t *testing.T) {
	h, err := preparar([]string{"-base"}, sinVariables, &bytes.Buffer{})

	if err == nil || !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("se esperaba un error que nombre DATABASE_URL y se obtuvo %v", err)
	}
	if h != nil {
		t.Error("con error no hay aplicación para servir")
	}
}

func TestPreparar_EmailsTienenQueSerTres(t *testing.T) {
	_, err := preparar([]string{"-emails", "a@b.test,c@d.test"}, sinVariables, &bytes.Buffer{})

	if err == nil || !strings.Contains(err.Error(), "tres emails") {
		t.Fatalf("se esperaba el error de los tres emails y se obtuvo %v", err)
	}
}

func TestPreparar_OpcionDesconocidaEsUnError(t *testing.T) {
	if _, err := preparar([]string{"-borrar"}, sinVariables, &bytes.Buffer{}); err == nil {
		t.Error("se esperaba un error por la opción desconocida")
	}
}
