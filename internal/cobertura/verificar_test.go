package cobertura

import (
	"math"
	"strings"
	"testing"
)

func medicionDe(paquetes ...Paquete) Medicion {
	return Medicion{Controlados: controlados, Paquetes: paquetes}
}

func TestVerificar_CumpleSiLaCoberturaCombinadaLlegaAlUmbral(t *testing.T) {
	casos := []struct {
		nombre string
		m      Medicion
	}{
		{"cobertura total", medicionDe(Paquete{Ruta: "internal/metrics", Sentencias: 100, Cubiertas: 100})},
		{"muy por encima", medicionDe(
			Paquete{Ruta: "internal/domain/backlog", Sentencias: 141, Cubiertas: 138},
			Paquete{Ruta: "internal/metrics", Sentencias: 89, Cubiertas: 89},
		)},
		{"justo en el umbral", medicionDe(Paquete{Ruta: "internal/metrics", Sentencias: 100, Cubiertas: 80})},
		{"un paquete en cero pero la combinada llega", medicionDe(
			Paquete{Ruta: "internal/domain", Sentencias: 90, Cubiertas: 90},
			Paquete{Ruta: "internal/metrics", Sentencias: 10, Cubiertas: 0},
		)},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if err := Verificar(c.m, 80); err != nil {
				t.Errorf("Verificar() = %v, se esperaba que cumpliera", err)
			}
		})
	}
}

func TestVerificar_UnPuntoPorDebajoDelUmbralFalla(t *testing.T) {
	m := medicionDe(Paquete{Ruta: "internal/domain/backlog", Sentencias: 100, Cubiertas: 79})

	err := Verificar(m, 80)

	esperado := "la cobertura combinada de internal/domain e internal/metrics es 79,0 %, por debajo del mínimo de 80,0 %"
	if err == nil || err.Error() != esperado {
		t.Errorf("Verificar() = %v, se esperaba el error %q", err, esperado)
	}
}

func TestVerificar_ElMensajeNombraTodosLosPaquetesControlados(t *testing.T) {
	casos := []struct {
		nombre      string
		controlados []string
		esperado    string
	}{
		{"uno", []string{"internal/domain"}, "la cobertura combinada de internal/domain es 50,0 %"},
		{"dos", []string{"internal/domain", "internal/metrics"}, "la cobertura combinada de internal/domain e internal/metrics es 50,0 %"},
		{"tres", []string{"a", "b", "c"}, "la cobertura combinada de a, b e c es 50,0 %"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			m := Medicion{Controlados: c.controlados, Paquetes: []Paquete{{Ruta: c.controlados[0], Sentencias: 10, Cubiertas: 5}}}

			err := Verificar(m, 80)

			if err == nil || !strings.HasPrefix(err.Error(), c.esperado) {
				t.Errorf("Verificar() = %v, se esperaba que empezara con %q", err, c.esperado)
			}
		})
	}
}

func TestVerificar_SinSentenciasControladasFalla(t *testing.T) {
	err := Verificar(medicionDe(), 80)

	esperado := "el perfil no tiene sentencias de los paquetes controlados (internal/domain, internal/metrics): ¿cambió el nombre de algún paquete?"
	if err == nil || err.Error() != esperado {
		t.Errorf("Verificar() = %v, se esperaba el error %q", err, esperado)
	}
}

func TestVerificar_ElUmbralTieneQueEstarEntreCeroYCien(t *testing.T) {
	m := medicionDe(Paquete{Ruta: "internal/metrics", Sentencias: 10, Cubiertas: 10})

	for _, umbral := range []float64{-1, 100.5, 150, math.NaN(), math.Inf(1)} {
		err := Verificar(m, umbral)
		if err == nil || err.Error() != "el umbral tiene que estar entre 0 y 100" {
			t.Errorf("Verificar(umbral %v) = %v, se esperaba el error del umbral", umbral, err)
		}
	}
	for _, umbral := range []float64{0, 100} {
		if err := Verificar(m, umbral); err != nil {
			t.Errorf("Verificar(umbral %v) = %v, el borde es válido", umbral, err)
		}
	}
}

func TestInforme_TieneUnaLineaPorPaqueteYLaCombinada(t *testing.T) {
	m := medicionDe(
		Paquete{Ruta: "internal/domain/backlog", Sentencias: 141, Cubiertas: 138},
		Paquete{Ruta: "internal/metrics", Sentencias: 89, Cubiertas: 89},
	)

	esperado := "Cobertura de los paquetes controlados (sentencias cubiertas / totales)\n" +
		"  internal/domain/backlog  138/141   97,9 %\n" +
		"  internal/metrics" + strings.Repeat(" ", 11) + "89/89  100,0 %\n" +
		"  Combinada" + strings.Repeat(" ", 16) + "227/230" + "   98,7 %" + "   (mínimo 80,0 %)\n"

	if got := Informe(m, 80); got != esperado {
		t.Errorf("Informe() =\n%s\nse esperaba\n%s", got, esperado)
	}
}

func TestInforme_SinPaquetesMuestraSoloLaCombinada(t *testing.T) {
	got := Informe(medicionDe(), 80)

	if !strings.Contains(got, "Combinada") || !strings.Contains(got, "0/0") || !strings.Contains(got, "0,0 %") {
		t.Errorf("Informe() = %q, se esperaba la línea de la combinada con 0/0 y 0,0 %%", got)
	}
}
