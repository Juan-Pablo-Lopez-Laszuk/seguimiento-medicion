package poker_test

import (
	"errors"
	"maps"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
)

// CA-18.4 y RN4 · el voto es un valor de la escala; los extremos valen
func TestValidarVoto(t *testing.T) {
	casos := []struct {
		texto string
		valor int
	}{
		{"0", 0}, {"1", 1}, {"2", 2}, {"3", 3}, {"5", 5}, {"8", 8}, {"13", 13}, {"21", 21}, {" 8 ", 8},
	}
	for _, c := range casos {
		t.Run(c.texto, func(t *testing.T) {
			valor, err := poker.ValidarVoto(c.texto)
			if err != nil || valor != c.valor {
				t.Errorf("ValidarVoto(%q) = %d, %v; se esperaba %d", c.texto, valor, err, c.valor)
			}
		})
	}
}

// CA-18.4 y RN4 · fuera de la escala, vacío o que no es un número entero es un error
func TestValidarVoto_FueraDeEscala(t *testing.T) {
	for _, texto := range []string{"4", "-1", "100", "22", "abc", "5.5", "+5", "", "   "} {
		t.Run(texto, func(t *testing.T) {
			if _, err := poker.ValidarVoto(texto); !errors.Is(err, poker.ErrVotoFueraDeEscala) {
				t.Errorf("ValidarVoto(%q): error = %v, se esperaba ErrVotoFueraDeEscala", texto, err)
			}
		})
	}
}

// RN2 · votan los integrantes activos, también el Agile Enabler; RN3 · solo antes de revelar la ronda
func TestPuedeVotar(t *testing.T) {
	enCurso := poker.Ronda{ID: 1, SesionID: 1, Numero: 1}
	revelada := poker.Ronda{ID: 1, SesionID: 1, Numero: 1, Revelada: true}
	casos := []struct {
		nombre string
		quien  member.Integrante
		ronda  poker.Ronda
		err    error
	}{
		{"un Product Builder activo", carolina, enCurso, nil},
		{"el Agile Enabler", juan, enCurso, nil},
		{"una integrante dada de baja", lucia, enCurso, poker.ErrSoloParticipantes},
		{"nadie (no se eligió quién)", member.Integrante{}, enCurso, poker.ErrSoloParticipantes},
		{"la ronda ya se reveló", carolina, revelada, poker.ErrRondaRevelada},
		{"dada de baja y la ronda revelada: primero quién vota", lucia, revelada, poker.ErrSoloParticipantes},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if err := poker.PuedeVotar(c.quien, c.ronda); !errors.Is(err, c.err) {
				t.Errorf("error = %v, se esperaba %v", err, c.err)
			}
		})
	}
}

// CA-18.2 y RN2 · quién votó, solo entre los participantes: el voto de quien se dio de baja no cuenta
func TestVotaron(t *testing.T) {
	participantes := []member.Integrante{juan, carolina, mariano}
	votaron := poker.Votaron(participantes, []int64{carolina.ID, lucia.ID, mariano.ID})
	esperado := map[int64]bool{carolina.ID: true, mariano.ID: true}
	if !maps.Equal(votaron, esperado) {
		t.Errorf("Votaron = %v, se esperaba %v", votaron, esperado)
	}
}

// CA-18.2 · si nadie votó, el conjunto está vacío (no nil)
func TestVotaron_Nadie(t *testing.T) {
	if votaron := poker.Votaron([]member.Integrante{juan, carolina}, nil); votaron == nil || len(votaron) != 0 {
		t.Errorf("Votaron = %#v, se esperaba un conjunto vacío", votaron)
	}
}

// Sección 7 de la spec de HU-18 · los mensajes que ve el usuario
func TestErroresDeVotar_Mensajes(t *testing.T) {
	casos := map[error]string{
		poker.ErrVotoFueraDeEscala: "el voto tiene que ser un valor de la escala: 0, 1, 2, 3, 5, 8, 13 o 21",
		poker.ErrSoloParticipantes: "solo votan los integrantes activos del proyecto",
		poker.ErrRondaRevelada:     "la ronda ya se reveló: no se puede votar",
	}
	for err, mensaje := range casos {
		if err.Error() != mensaje {
			t.Errorf("mensaje = %q, se esperaba %q", err.Error(), mensaje)
		}
	}
}
