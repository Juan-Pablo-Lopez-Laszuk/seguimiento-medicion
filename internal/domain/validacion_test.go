package domain_test

import (
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
)

// El texto del error junta todos los campos, ordenados, con su mensaje.
func TestErroresValidacion_TextoConTodosLosCamposOrdenados(t *testing.T) {
	errs := domain.ErroresValidacion{
		"nombre":       errors.New("el nombre es obligatorio"),
		"fecha_inicio": errors.New("la fecha de inicio es obligatoria y debe ser válida"),
	}

	esperado := "fecha_inicio: la fecha de inicio es obligatoria y debe ser válida; nombre: el nombre es obligatorio"
	if errs.Error() != esperado {
		t.Errorf("se esperaba %q y se obtuvo %q", esperado, errs.Error())
	}
}

// Agregar solo registra los campos que tienen error.
func TestErroresValidacion_AgregarIgnoraLosCamposSinError(t *testing.T) {
	errs := domain.ErroresValidacion{}
	falla := errors.New("falla")

	errs.Agregar("con_error", falla)
	errs.Agregar("sin_error", nil)

	if len(errs) != 1 || !errors.Is(errs["con_error"], falla) {
		t.Errorf("se esperaba solo el campo con error y se obtuvo %v", errs)
	}
}
