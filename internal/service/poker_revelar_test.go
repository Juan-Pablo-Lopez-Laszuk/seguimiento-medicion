package service_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// escenarioRevelar es el de HU-17 con la sesión de HU-1 abierta por Juan y votos en la ronda 1, guardados con Votar
// de HU-18.
type escenarioRevelar struct {
	*escenarioPoker
	casos    *service.RevelarPoker
	ronda    poker.Ronda
	historia int64
}

func nuevoEscenarioRevelar(t *testing.T, votos map[string]int) *escenarioRevelar {
	t.Helper()
	e := nuevoEscenarioPoker(t)
	v, err := e.casos.Abrir(context.Background(), e.proyecto, 1, e.porNombre["Juan"].ID)
	if err != nil {
		t.Fatalf("abrir la sesión: %v", err)
	}
	for nombre, valor := range votos {
		if err := e.sesiones.Votar(context.Background(), v.Ronda.ID, e.porNombre[nombre].ID, valor); err != nil {
			t.Fatalf("votar: %v", err)
		}
	}
	return &escenarioRevelar{escenarioPoker: e, casos: service.NuevoRevelarPoker(e.casos, e.sesiones), ronda: v.Ronda, historia: v.Historia.ID}
}

func (e *escenarioRevelar) rondaGuardada(t *testing.T) poker.Ronda {
	t.Helper()
	_, r, err := e.sesiones.SesionAbierta(context.Background(), e.historia)
	if err != nil {
		t.Fatalf("buscar la ronda: %v", err)
	}
	return r
}

// CA-19.1 y RN1 · si votaron todos los integrantes activos, la ronda se revela sola al ver la sesión
func TestRevelarPoker_SeRevelaSolaCuandoVotaronTodos(t *testing.T) {
	e := nuevoEscenarioRevelar(t, map[string]int{"Juan": 5, "Carolina": 3, "Mariano": 5})

	v, err := e.casos.Revisar(context.Background(), e.proyecto, 1)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if !v.Ronda.Revelada || !v.ReveladaAhora || !e.rondaGuardada(t).Revelada {
		t.Errorf("vista = %+v, se esperaba la ronda revelada y guardada", v)
	}
	quiere := []poker.VotoRevelado{{Integrante: "Carolina", Valor: 3, Voto: true}, {Integrante: "Juan", Valor: 5, Voto: true}, {Integrante: "Mariano", Valor: 5, Voto: true}}
	if !reflect.DeepEqual(v.Votos, quiere) {
		t.Errorf("votos = %+v, se esperaba %+v", v.Votos, quiere)
	}
}

// RN1 y RN6 · si falta votar alguien no se revela ni se ven los valores
func TestRevelarPoker_NoSeRevelaSiFaltaAlguien(t *testing.T) {
	e := nuevoEscenarioRevelar(t, map[string]int{"Juan": 5, "Carolina": 3})

	v, err := e.casos.Revisar(context.Background(), e.proyecto, 1)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if v.Ronda.Revelada || v.ReveladaAhora || v.Votos != nil || e.rondaGuardada(t).Revelada {
		t.Errorf("vista = %+v, se esperaba la ronda sin revelar y sin votos a la vista", v)
	}
}

// RN5 · una ronda ya revelada muestra los votos pero no avisa de nuevo
func TestRevelarPoker_RevisarUnaRondaYaRevelada(t *testing.T) {
	e := nuevoEscenarioRevelar(t, map[string]int{"Juan": 5, "Carolina": 3, "Mariano": 5})
	if _, err := e.casos.Revisar(context.Background(), e.proyecto, 1); err != nil {
		t.Fatalf("primera revisión: %v", err)
	}

	v, err := e.casos.Revisar(context.Background(), e.proyecto, 1)

	if err != nil || !v.Ronda.Revelada || v.ReveladaAhora || len(v.Votos) != 3 {
		t.Errorf("vista = %+v, %v; se esperaba la ronda revelada con 3 votos y sin aviso", v, err)
	}
}

// HU-19 · sin sesión abierta no hay nada que revelar
func TestRevelarPoker_RevisarSinSesion(t *testing.T) {
	e := nuevoEscenarioPoker(t)
	casos := service.NuevoRevelarPoker(e.casos, e.sesiones)

	v, err := casos.Revisar(context.Background(), e.proyecto, 1)

	if err != nil || v.Abierta || v.Votos != nil {
		t.Errorf("vista = %+v, %v; se esperaba la historia sin sesión", v, err)
	}
}

// CA-19.1 y RN2 · el Agile Enabler cierra la ronda con los votos que haya; los que no votaron aparecen igual
func TestRevelarPoker_ElAgileEnablerCierraLaRonda(t *testing.T) {
	e := nuevoEscenarioRevelar(t, map[string]int{"Carolina": 8})

	v, err := e.casos.Cerrar(context.Background(), e.proyecto, 1, e.porNombre["Juan"].ID)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if !v.Ronda.Revelada || !v.ReveladaAhora || !e.rondaGuardada(t).Revelada {
		t.Errorf("vista = %+v, se esperaba la ronda revelada y guardada", v)
	}
	quiere := []poker.VotoRevelado{{Integrante: "Carolina", Valor: 8, Voto: true}, {Integrante: "Juan"}, {Integrante: "Mariano"}}
	if !reflect.DeepEqual(v.Votos, quiere) {
		t.Errorf("votos = %+v, se esperaba %+v", v.Votos, quiere)
	}
}

// CA-19.3, RN4, RN5 y RN7 · cuándo no se puede cerrar la ronda; no se guarda nada
func TestRevelarPoker_CerrarConError(t *testing.T) {
	casos := []struct {
		nombre string
		votos  map[string]int
		quien  string
		antes  bool
		quiere error
	}{
		{"un Product Builder", map[string]int{"Carolina": 8}, "Carolina", false, poker.ErrSoloAgileEnabler},
		{"una integrante dada de baja", map[string]int{"Carolina": 8}, "Lucía", false, poker.ErrSoloAgileEnabler},
		{"alguien de otro proyecto", map[string]int{"Carolina": 8}, "", false, poker.ErrSoloAgileEnabler},
		{"la ronda sin votos", nil, "Juan", false, poker.ErrRondaSinVotos},
		{"solo votó una integrante dada de baja (RN7)", map[string]int{"Lucía": 8}, "Juan", false, poker.ErrRondaSinVotos},
		{"la ronda ya revelada", map[string]int{"Carolina": 8}, "Juan", true, poker.ErrRondaYaRevelada},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			e := nuevoEscenarioRevelar(t, c.votos)
			if c.antes {
				if _, err := e.casos.Cerrar(context.Background(), e.proyecto, 1, e.porNombre["Juan"].ID); err != nil {
					t.Fatalf("primer cierre: %v", err)
				}
			}

			_, err := e.casos.Cerrar(context.Background(), e.proyecto, 1, e.porNombre[c.quien].ID)

			if !errors.Is(err, c.quiere) {
				t.Errorf("error = %v, se esperaba %v", err, c.quiere)
			}
			if !c.antes && e.rondaGuardada(t).Revelada {
				t.Error("la ronda quedó revelada y no se esperaba")
			}
		})
	}
}

// HU-19 · cerrar sin sesión abierta
func TestRevelarPoker_CerrarSinSesion(t *testing.T) {
	e := nuevoEscenarioPoker(t)
	casos := service.NuevoRevelarPoker(e.casos, e.sesiones)

	_, err := casos.Cerrar(context.Background(), e.proyecto, 1, e.porNombre["Juan"].ID)

	if !errors.Is(err, poker.ErrSesionNoEncontrada) {
		t.Errorf("error = %v, se esperaba poker.ErrSesionNoEncontrada", err)
	}
}
