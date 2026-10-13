package sprint_test

import (
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
)

// sprintsDeEjemplo devuelve los sprints de la spec de HU-04: el Sprint 1 del 05/10 al 11/10 y el Sprint 2
// del 12/10 al 18/10, los dos sin cerrar.
func sprintsDeEjemplo(t *testing.T) []sprint.Sprint {
	return []sprint.Sprint{
		{ID: 1, Numero: 1, FechaInicio: fecha(t, "2026-10-05"), FechaFin: fecha(t, "2026-10-11"), Estado: sprint.EstadoPlanificado},
		{ID: 2, Numero: 2, FechaInicio: fecha(t, "2026-10-12"), FechaFin: fecha(t, "2026-10-18"), Estado: sprint.EstadoPlanificado},
	}
}

// HU-04 RN2 · el sprint activo es el no cerrado de menor número cuyo inicio ya llegó
func TestActivo(t *testing.T) {
	casos := []struct {
		nombre   string
		hoy      string
		cerrados []int // números de los sprints que están cerrados
		esperado int   // número del sprint activo; 0 si no hay
	}{
		{"antes del primer sprint no hay activo", "2026-10-04", nil, 0},
		{"el día que empieza el Sprint 1", "2026-10-05", nil, 1},
		{"el último día del Sprint 1", "2026-10-11", nil, 1},
		{"vencido sin cerrar sigue activo", "2026-10-13", nil, 1},
		{"cerrado el Sprint 1, el activo es el 2", "2026-10-13", []int{1}, 2},
		{"cerrado el Sprint 1 antes de que empiece el 2", "2026-10-11", []int{1}, 0},
		{"todos cerrados no hay activo", "2026-10-20", []int{1, 2}, 0},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			sprints := sprintsDeEjemplo(t)
			for i := range sprints {
				for _, n := range c.cerrados {
					if sprints[i].Numero == n {
						sprints[i].Estado = sprint.EstadoCerrado
					}
				}
			}
			activo, ok := sprint.Activo(sprints, fecha(t, c.hoy))
			switch {
			case c.esperado == 0 && ok:
				t.Errorf("no se esperaba sprint activo y se obtuvo el Sprint %d", activo.Numero)
			case c.esperado != 0 && (!ok || activo.Numero != c.esperado):
				t.Errorf("se esperaba el Sprint %d activo y se obtuvo %+v (ok=%v)", c.esperado, activo, ok)
			}
		})
	}
}

// HU-04 RN2 · el orden en que llegan los sprints no cambia el resultado
func TestActivo_NoDependeDelOrden(t *testing.T) {
	s := sprintsDeEjemplo(t)
	desordenados := []sprint.Sprint{s[1], s[0]}
	activo, ok := sprint.Activo(desordenados, fecha(t, "2026-10-13"))
	if !ok || activo.Numero != 1 {
		t.Errorf("se esperaba el Sprint 1 activo y se obtuvo %+v (ok=%v)", activo, ok)
	}
}

func TestActivo_SinSprints(t *testing.T) {
	if activo, ok := sprint.Activo(nil, fecha(t, "2026-10-13")); ok {
		t.Errorf("no se esperaba sprint activo y se obtuvo %+v", activo)
	}
}

// HU-04 sección 3 · sin sprint activo se muestra cuál es el próximo
func TestProximo(t *testing.T) {
	s := sprintsDeEjemplo(t)
	proximo, ok := sprint.Proximo(s, fecha(t, "2026-10-04"))
	if !ok || proximo.Numero != 1 {
		t.Errorf("se esperaba el Sprint 1 como próximo y se obtuvo %+v (ok=%v)", proximo, ok)
	}

	s[0].Estado = sprint.EstadoCerrado
	proximo, ok = sprint.Proximo(s, fecha(t, "2026-10-11"))
	if !ok || proximo.Numero != 2 {
		t.Errorf("con el Sprint 1 cerrado se esperaba el Sprint 2 como próximo y se obtuvo %+v (ok=%v)", proximo, ok)
	}

	if proximo, ok := sprint.Proximo(s, fecha(t, "2026-10-20")); ok {
		t.Errorf("después del último sprint no se esperaba próximo y se obtuvo %+v", proximo)
	}
}

// HU-04 RN3 · un sprint terminó por fecha cuando hoy es posterior a su fecha de fin
func TestTerminado(t *testing.T) {
	s := sprintsDeEjemplo(t)[0]
	casos := map[string]bool{"2026-10-10": false, "2026-10-11": false, "2026-10-12": true}
	for hoy, esperado := range casos {
		if got := s.Terminado(fecha(t, hoy)); got != esperado {
			t.Errorf("hoy %s: Terminado = %v, se esperaba %v", hoy, got, esperado)
		}
	}
}

// HU-04 RN1 · el estado del proyecto necesita el inicio del primer sprint (el de menor número)
func TestInicioDelPrimero(t *testing.T) {
	s := sprintsDeEjemplo(t)
	if got := sprint.InicioDelPrimero([]sprint.Sprint{s[1], s[0]}); !got.Equal(fecha(t, "2026-10-05")) {
		t.Errorf("se esperaba el 05/10/2026 y se obtuvo %v", got)
	}
	if got := sprint.InicioDelPrimero(nil); !got.IsZero() {
		t.Errorf("sin sprints se esperaba la fecha cero y se obtuvo %v", got)
	}
}
