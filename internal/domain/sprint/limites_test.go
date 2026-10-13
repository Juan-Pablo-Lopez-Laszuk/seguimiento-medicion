package sprint_test

import (
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
)

// HU-02 RN3 y RN4 · lo que necesita project.Modificar: cuándo empieza el primer sprint y cuándo termina el último
func TestLimites(t *testing.T) {
	s := sprintsDeEjemplo(t)
	got := sprint.Limites([]sprint.Sprint{s[1], s[0]})

	esperado := project.LimitesSprints{
		PrimerNumero: 1, PrimerInicio: fecha(t, "2026-10-05"),
		UltimoNumero: 2, UltimoFin: fecha(t, "2026-10-18"),
	}
	if got != esperado {
		t.Errorf("se esperaba %+v y se obtuvo %+v", esperado, got)
	}
}

func TestLimites_SinSprints(t *testing.T) {
	if got := sprint.Limites(nil); got.HaySprints() || got != (project.LimitesSprints{}) {
		t.Errorf("sin sprints se esperaban límites vacíos y se obtuvo %+v", got)
	}
}
