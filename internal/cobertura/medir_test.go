package cobertura

import (
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

const modulo = "github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/"

var controlados = []string{"internal/domain", "internal/metrics"}

// perfil arma un perfil de cobertura de prueba con el encabezado "mode: atomic" y las líneas dadas.
func perfil(lineas ...string) string {
	return "mode: atomic\n" + strings.Join(lineas, "\n") + "\n"
}

func medir(t *testing.T, texto string) Medicion {
	t.Helper()
	m, err := Medir(strings.NewReader(texto), controlados)
	if err != nil {
		t.Fatalf("Medir() dio error: %v", err)
	}
	return m
}

func TestMedir_SumaLasSentenciasDeLosPaquetesControlados(t *testing.T) {
	m := medir(t, perfil(
		modulo+"internal/metrics/velocity.go:10.2,12.3 4 1",
		modulo+"internal/metrics/velocity.go:14.2,16.3 6 0",
		modulo+"internal/domain/fecha.go:5.2,7.3 3 2",
	))

	if m.Sentencias() != 13 || m.Cubiertas() != 7 {
		t.Errorf("sentencias y cubiertas = %d y %d, se esperaban 13 y 7", m.Sentencias(), m.Cubiertas())
	}
}

func TestMedir_AgrupaPorPaqueteYLosOrdenaPorRuta(t *testing.T) {
	m := medir(t, perfil(
		modulo+"internal/metrics/velocity.go:10.2,12.3 4 1",
		modulo+"internal/domain/fecha.go:5.2,7.3 3 0",
		modulo+"internal/metrics/otro.go:1.2,3.3 2 1",
	))

	esperado := []Paquete{
		{Ruta: "internal/domain", Sentencias: 3, Cubiertas: 0},
		{Ruta: "internal/metrics", Sentencias: 6, Cubiertas: 6},
	}
	if len(m.Paquetes) != len(esperado) {
		t.Fatalf("paquetes = %+v, se esperaba %+v", m.Paquetes, esperado)
	}
	for i, p := range esperado {
		if m.Paquetes[i] != p {
			t.Errorf("paquete %d = %+v, se esperaba %+v", i, m.Paquetes[i], p)
		}
	}
}

func TestMedir_LosSubpaquetesCuentanComoParteDelPaqueteControlado(t *testing.T) {
	m := medir(t, perfil(
		modulo+"internal/domain/backlog/historia.go:10.2,12.3 5 5",
		modulo+"internal/domain/sprint/sprint.go:3.2,5.3 5 0",
	))

	if m.Sentencias() != 10 || m.Cubiertas() != 5 {
		t.Errorf("sentencias y cubiertas = %d y %d, se esperaban 10 y 5", m.Sentencias(), m.Cubiertas())
	}
	if len(m.Paquetes) != 2 || m.Paquetes[0].Ruta != "internal/domain/backlog" || m.Paquetes[1].Ruta != "internal/domain/sprint" {
		t.Errorf("paquetes = %+v, se esperaban internal/domain/backlog e internal/domain/sprint", m.Paquetes)
	}
}

func TestMedir_IgnoraLosPaquetesQueNoSeControlan(t *testing.T) {
	m := medir(t, perfil(
		modulo+"cmd/server/main.go:13.13,15.16 50 0",
		modulo+"internal/server/render.go:3.2,5.3 200 40",
		modulo+"internal/domainx/otro.go:3.2,5.3 7 0", // parecido a internal/domain, pero es otro paquete
		modulo+"internal/metrics/velocity.go:10.2,12.3 4 4",
	))

	if m.Sentencias() != 4 || m.Cubiertas() != 4 {
		t.Errorf("sentencias y cubiertas = %d y %d, se esperaban 4 y 4", m.Sentencias(), m.Cubiertas())
	}
}

func TestMedir_UnBloqueRepetidoCuentaUnaSolaVezYEstaCubiertoSiAlgunaAparicionLoCubre(t *testing.T) {
	m := medir(t, perfil(
		modulo+"internal/metrics/velocity.go:10.2,12.3 10 0",
		modulo+"internal/metrics/velocity.go:10.2,12.3 10 3",
		modulo+"internal/metrics/velocity.go:20.2,22.3 5 0",
		modulo+"internal/metrics/velocity.go:20.2,22.3 5 0",
	))

	if m.Sentencias() != 15 || m.Cubiertas() != 10 {
		t.Errorf("sentencias y cubiertas = %d y %d, se esperaban 15 y 10", m.Sentencias(), m.Cubiertas())
	}
}

func TestMedir_AceptaLosTresModosDeCobertura(t *testing.T) {
	for _, modo := range []string{"set", "count", "atomic"} {
		texto := "mode: " + modo + "\n" + modulo + "internal/metrics/velocity.go:10.2,12.3 4 1\n"
		m, err := Medir(strings.NewReader(texto), controlados)
		if err != nil {
			t.Errorf("modo %q: Medir() dio error: %v", modo, err)
			continue
		}
		if m.Sentencias() != 4 || m.Cubiertas() != 4 {
			t.Errorf("modo %q: sentencias y cubiertas = %d y %d, se esperaban 4 y 4", modo, m.Sentencias(), m.Cubiertas())
		}
	}
}

func TestMedir_SinSentenciasControladasDevuelveUnaMedicionVacia(t *testing.T) {
	m := medir(t, perfil(modulo+"internal/server/render.go:3.2,5.3 200 40"))

	if len(m.Paquetes) != 0 || m.Sentencias() != 0 || m.Cubiertas() != 0 {
		t.Errorf("medición = %+v, se esperaba vacía", m)
	}
}

func TestMedir_UnPerfilIlegibleDaError(t *testing.T) {
	casos := []struct {
		nombre   string
		texto    string
		contiene string
	}{
		{"vacío", "", "perfil de cobertura: está vacío"},
		{"sin encabezado mode", modulo + "internal/metrics/velocity.go:10.2,12.3 4 1\n", "perfil de cobertura: la primera línea tiene que empezar con \"mode:\""},
		{"modo desconocido", "mode: raro\n", "perfil de cobertura: modo desconocido \"raro\""},
		{"línea sin sentencias ni veces", perfil(modulo + "internal/metrics/velocity.go:10.2,12.3"), "perfil de cobertura: línea 2 mal formada"},
		{"línea sin bloque", perfil("internal/metrics/velocity.go 4 1"), "perfil de cobertura: línea 2 mal formada"},
		{"sentencias que no son un número", perfil(modulo + "internal/metrics/velocity.go:10.2,12.3 cuatro 1"), "perfil de cobertura: línea 2 mal formada"},
		{"veces que no son un número", perfil(modulo + "internal/metrics/velocity.go:10.2,12.3 4 uno"), "perfil de cobertura: línea 2 mal formada"},
		{"sentencias negativas", perfil(modulo + "internal/metrics/velocity.go:10.2,12.3 -4 1"), "perfil de cobertura: línea 2 mal formada"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := Medir(strings.NewReader(c.texto), controlados)
			if err == nil {
				t.Fatalf("se esperaba el error %q y no hubo ninguno", c.contiene)
			}
			if !strings.Contains(err.Error(), c.contiene) {
				t.Errorf("error = %q, se esperaba que contuviera %q", err.Error(), c.contiene)
			}
		})
	}
}

func TestMedir_IgnoraLasLineasEnBlanco(t *testing.T) {
	m := medir(t, "mode: atomic\n\n"+modulo+"internal/metrics/velocity.go:10.2,12.3 4 1\n\n")

	if m.Sentencias() != 4 {
		t.Errorf("sentencias = %d, se esperaban 4", m.Sentencias())
	}
}

func TestMedicion_Porcentaje(t *testing.T) {
	casos := []struct {
		nombre   string
		m        Medicion
		esperado float64
	}{
		{"sin sentencias", Medicion{}, 0},
		{"todo cubierto", Medicion{Paquetes: []Paquete{{Ruta: "a", Sentencias: 10, Cubiertas: 10}}}, 100},
		{"nada cubierto", Medicion{Paquetes: []Paquete{{Ruta: "a", Sentencias: 10, Cubiertas: 0}}}, 0},
		{"se pondera por sentencias", Medicion{Paquetes: []Paquete{
			{Ruta: "a", Sentencias: 90, Cubiertas: 90},
			{Ruta: "b", Sentencias: 10, Cubiertas: 0},
		}}, 90},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := c.m.Porcentaje(); got != c.esperado {
				t.Errorf("Porcentaje() = %v, se esperaba %v", got, c.esperado)
			}
		})
	}
}

func TestMedir_UnErrorDeLecturaSeInformaConElPrefijoDelPerfil(t *testing.T) {
	falla := errors.New("disco roto")
	casos := []struct {
		nombre string
		lector io.Reader
	}{
		{"antes de la primera línea", iotest.ErrReader(falla)},
		{"a mitad del perfil", io.MultiReader(strings.NewReader(perfil(modulo+"internal/metrics/velocity.go:10.2,12.3 4 1")), iotest.ErrReader(falla))},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			_, err := Medir(c.lector, controlados)

			if !errors.Is(err, falla) || !strings.HasPrefix(err.Error(), "perfil de cobertura: ") {
				t.Errorf("Medir() = %v, se esperaba el error de lectura con el prefijo \"perfil de cobertura: \"", err)
			}
		})
	}
}
