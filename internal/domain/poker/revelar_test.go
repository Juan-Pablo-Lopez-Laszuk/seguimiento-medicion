package poker_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
)

// equipo: Juan Pablo (1) Agile Enabler, Carolina (2) y Mariano (3) Product Builders, y Ana (4) dada de baja.
func equipo() []member.Integrante {
	return []member.Integrante{
		{ID: 1, Nombre: "Juan Pablo", Rol: member.RolAgileEnabler, Activo: true},
		{ID: 2, Nombre: "Carolina", Rol: member.RolProductBuilder, Activo: true},
		{ID: 3, Nombre: "Mariano", Rol: member.RolProductBuilder, Activo: true},
		{ID: 4, Nombre: "Ana", Rol: member.RolProductBuilder, Activo: false},
	}
}

func votos(pares ...int) []poker.Voto {
	var v []poker.Voto
	for i := 0; i < len(pares); i += 2 {
		v = append(v, poker.Voto{RondaID: 7, IntegranteID: int64(pares[i]), Valor: pares[i+1]})
	}
	return v
}

var ronda = poker.Ronda{ID: 7, SesionID: 3, Numero: 1}

// CA-19.1 · RN1 · se revela sola cuando votaron todos los integrantes activos; un dado de baja no cuenta
func TestRevelarSiCorresponde(t *testing.T) {
	casos := []struct {
		nombre   string
		votos    []poker.Voto
		revelada bool
	}{
		{"votaron todos", votos(1, 5, 2, 8, 3, 5), true},
		{"falta uno", votos(1, 5, 2, 8), false},
		{"sin votos", nil, false},
		{"votó también una dada de baja", votos(1, 5, 2, 8, 3, 5, 4, 13), true},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := poker.RevelarSiCorresponde(ronda, c.votos, equipo()); got.Revelada != c.revelada || got.ID != ronda.ID {
				t.Errorf("ronda = %+v, se esperaba revelada = %v", got, c.revelada)
			}
			if got := poker.TodosVotaron(c.votos, equipo()); got != c.revelada {
				t.Errorf("TodosVotaron = %v, se esperaba %v", got, c.revelada)
			}
		})
	}
	if poker.TodosVotaron(nil, nil) {
		t.Error("sin integrantes activos no se puede decir que votaron todos")
	}
}

// CA-19.1 y CA-19.3 · RN2 a RN5 · el Agile Enabler cierra la ronda si tiene votos y no está revelada
func TestCerrar(t *testing.T) {
	revelada := ronda
	revelada.Revelada = true
	ae, pb, aeDeBaja := equipo()[0], equipo()[1], member.Integrante{ID: 9, Rol: member.RolAgileEnabler}
	casos := []struct {
		nombre string
		ronda  poker.Ronda
		votos  []poker.Voto
		quien  member.Integrante
		err    error
	}{
		{"el Agile Enabler con votos", ronda, votos(2, 3), ae, nil},
		{"sin votos", ronda, nil, ae, poker.ErrRondaSinVotos},
		{"un Product Builder", ronda, votos(2, 3), pb, poker.ErrSoloAgileEnabler},
		{"un Agile Enabler dado de baja", ronda, votos(2, 3), aeDeBaja, poker.ErrSoloAgileEnabler},
		{"ya revelada", revelada, votos(2, 3), ae, poker.ErrRondaYaRevelada},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			r, err := poker.Cerrar(c.ronda, c.votos, c.quien)
			if !errors.Is(err, c.err) {
				t.Fatalf("error = %v, se esperaba %v", err, c.err)
			}
			if c.err == nil && !r.Revelada {
				t.Errorf("ronda = %+v, se esperaba revelada", r)
			}
		})
	}
}

// CA-19.2 · RN6 · RN7 · revelada, cada integrante activo con su voto (o "no votó"), por nombre; el voto de Ana, que se
// dio de baja, no se muestra; sin revelar, ningún valor
func TestVotosRevelados(t *testing.T) {
	revelada := ronda
	revelada.Revelada = true

	got, err := poker.VotosRevelados(revelada, votos(2, 3, 3, 5, 4, 13), equipo())
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	quiero := []poker.VotoRevelado{
		{Integrante: "Carolina", Valor: 3, Voto: true},
		{Integrante: "Juan Pablo"},
		{Integrante: "Mariano", Valor: 5, Voto: true},
	}
	if !reflect.DeepEqual(got, quiero) {
		t.Errorf("votos = %+v, se esperaba %+v", got, quiero)
	}

	if _, err := poker.VotosRevelados(ronda, votos(2, 3), equipo()); !errors.Is(err, poker.ErrRondaSinRevelar) {
		t.Errorf("sin revelar: error = %v, se esperaba ErrRondaSinRevelar", err)
	}
}

// Los mensajes de la sección 7 de la spec
func TestMensajesDeError(t *testing.T) {
	for err, quiero := range map[error]string{
		poker.ErrRondaSinVotos:   "la ronda no tiene votos: no se puede revelar",
		poker.ErrRondaYaRevelada: "la ronda ya está revelada",
		poker.ErrRondaSinRevelar: "los votos se ven cuando la ronda se revela",
	} {
		if err.Error() != quiero {
			t.Errorf("mensaje = %q, se esperaba %q", err.Error(), quiero)
		}
	}
}
