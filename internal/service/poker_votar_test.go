package service_test

import (
	"context"
	"errors"
	"maps"
	"reflect"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// nuevoEscenarioVotos es el de features/hu-18-votar-oculto.feature: el de HU-17 con los votos conectados y la sesión
// de HU-1 Login abierta por Juan.
func nuevoEscenarioVotos(t *testing.T) *escenarioPoker {
	t.Helper()
	e := nuevoEscenarioPoker(t)
	e.casos.ConVotos(e.sesiones)
	if _, err := e.casos.Abrir(context.Background(), e.proyecto, 1, e.porNombre["Juan"].ID); err != nil {
		t.Fatalf("abrir la sesión: %v", err)
	}
	return e
}

// votaron arma el conjunto de IDs de quienes votaron, a partir de los nombres.
func (e *escenarioPoker) votaron(nombres ...string) map[int64]bool {
	ids := map[int64]bool{}
	for _, n := range nombres {
		ids[e.porNombre[n].ID] = true
	}
	return ids
}

func (e *escenarioPoker) votar(t *testing.T, nombre, valor string) service.VistaSesion {
	t.Helper()
	v, err := e.casos.Votar(context.Background(), e.proyecto, 1, e.porNombre[nombre].ID, valor)
	if err != nil {
		t.Fatalf("%s vota %q: %v", nombre, valor, err)
	}
	return v
}

func verificarVotaron(t *testing.T, obtenido, quiero map[int64]bool) {
	t.Helper()
	if !maps.Equal(obtenido, quiero) {
		t.Errorf("votaron = %v, se esperaba %v", obtenido, quiero)
	}
}

// CA-18.1 y CA-18.2 · el voto se guarda y la vista dice quién votó
func TestVotarPoker_GuardaElVotoYMuestraQuienVoto(t *testing.T) {
	e := nuevoEscenarioVotos(t)

	v := e.votar(t, "Carolina", "5")

	if !v.Abierta || v.Ronda.Numero != 1 {
		t.Errorf("se esperaba la sesión abierta en la Ronda 1 y se obtuvo %+v", v)
	}
	verificarVotaron(t, v.Votaron, e.votaron("Carolina"))
	visto, err := e.casos.Ver(context.Background(), e.proyecto, 1)
	if err != nil {
		t.Fatalf("ver: %v", err)
	}
	verificarVotaron(t, visto.Votaron, e.votaron("Carolina"))
}

// CA-18.1 y RN1 · votar de nuevo reemplaza el voto; RN2 · el Agile Enabler también vota
func TestVotarPoker_DeNuevoYElAgileEnabler(t *testing.T) {
	e := nuevoEscenarioVotos(t)
	e.votar(t, "Carolina", "5")
	e.votar(t, "Juan", "3")

	v := e.votar(t, "Carolina", "8")

	verificarVotaron(t, v.Votaron, e.votaron("Juan", "Carolina"))
}

// CA-18.4, RN2 y RN4 · los errores de votar; no se guarda nada
func TestVotarPoker_Errores(t *testing.T) {
	casos := []struct {
		nombre string
		quien  string
		valor  string
		err    error
	}{
		{"fuera de la escala", "Carolina", "4", poker.ErrVotoFueraDeEscala},
		{"negativo", "Carolina", "-1", poker.ErrVotoFueraDeEscala},
		{"no es un número", "Carolina", "abc", poker.ErrVotoFueraDeEscala},
		{"vacío", "Carolina", "", poker.ErrVotoFueraDeEscala},
		{"una integrante dada de baja", "Lucía", "5", poker.ErrSoloParticipantes},
		{"nadie (no se eligió quién)", "", "5", poker.ErrSoloParticipantes},
		{"nadie y fuera de la escala: primero quién vota", "", "4", poker.ErrSoloParticipantes},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			e := nuevoEscenarioVotos(t)

			_, err := e.casos.Votar(context.Background(), e.proyecto, 1, e.porNombre[c.quien].ID, c.valor)

			if !errors.Is(err, c.err) {
				t.Errorf("error = %v, se esperaba %v", err, c.err)
			}
			v, _ := e.casos.Ver(context.Background(), e.proyecto, 1)
			verificarVotaron(t, v.Votaron, map[int64]bool{})
		})
	}
}

// RN2 · alguien de otro proyecto es como no haber elegido a nadie
func TestVotarPoker_IntegranteDeOtroProyecto(t *testing.T) {
	e := nuevoEscenarioVotos(t)
	otro, _ := e.integrantes.Guardar(context.Background(), member.Integrante{ProyectoID: 999, Nombre: "Otro", Rol: member.RolProductBuilder, Activo: true})

	if _, err := e.casos.Votar(context.Background(), e.proyecto, 1, otro.ID, "5"); !errors.Is(err, poker.ErrSoloParticipantes) {
		t.Errorf("error = %v, se esperaba ErrSoloParticipantes", err)
	}
}

// RN3 · sin sesión abierta no se vota; el error es el de la spec, tal cual, porque es el aviso que ve el usuario
func TestVotarPoker_SinSesion(t *testing.T) {
	e := nuevoEscenarioVotos(t)

	if _, err := e.casos.Votar(context.Background(), e.proyecto, 2, e.porNombre["Carolina"].ID, "5"); err == nil || err.Error() != poker.ErrSesionNoEncontrada.Error() {
		t.Errorf("error = %v, se esperaba el mensaje de ErrSesionNoEncontrada sin contexto", err)
	}
}

// sesionesQueSeRevelan dicen que la ronda está sin revelar, pero al guardar el voto el repositorio la encuentra
// revelada (se reveló entre que se buscó la sesión y se votó, RN3).
type sesionesQueSeRevelan struct {
	service.RepositorioSesionesPoker
}

func (sesionesQueSeRevelan) Votar(context.Context, int64, int64, int) error {
	return poker.ErrRondaRevelada
}

func (sesionesQueSeRevelan) QuienesVotaron(context.Context, int64) ([]int64, error) {
	return []int64{}, nil
}

// RN3 · si la ronda se revela antes de que llegue el voto, el error es el de la spec, tal cual
func TestVotarPoker_SeRevelaAntesDeGuardar(t *testing.T) {
	e := nuevoEscenarioPoker(t)
	_, _ = e.casos.Abrir(context.Background(), e.proyecto, 1, e.porNombre["Juan"].ID)
	e.casos.ConVotos(sesionesQueSeRevelan{e.sesiones})

	if _, err := e.casos.Votar(context.Background(), e.proyecto, 1, e.porNombre["Carolina"].ID, "5"); err == nil || err.Error() != poker.ErrRondaRevelada.Error() {
		t.Errorf("error = %v, se esperaba el mensaje de ErrRondaRevelada sin contexto", err)
	}
}

// sesionesReveladas es el repositorio en memoria con la ronda actual ya revelada (HU-19 todavía no la revela).
type sesionesReveladas struct {
	service.RepositorioSesionesPoker
}

func (s sesionesReveladas) SesionAbierta(ctx context.Context, historiaID int64) (poker.Sesion, poker.Ronda, error) {
	sesion, ronda, err := s.RepositorioSesionesPoker.SesionAbierta(ctx, historiaID)
	ronda.Revelada = true
	return sesion, ronda, err
}

// RN3 · en una ronda revelada no se vota
func TestVotarPoker_RondaRevelada(t *testing.T) {
	e := nuevoEscenarioPoker(t)
	_, _ = e.casos.Abrir(context.Background(), e.proyecto, 1, e.porNombre["Juan"].ID)
	casos := service.NuevoPoker(e.proyectos, e.historias, e.integrantes, sesionesReveladas{e.sesiones}).ConVotos(e.sesiones)

	if _, err := casos.Votar(context.Background(), e.proyecto, 1, e.porNombre["Carolina"].ID, "5"); !errors.Is(err, poker.ErrRondaRevelada) {
		t.Errorf("error = %v, se esperaba ErrRondaRevelada", err)
	}
}

// RN2 · el voto de un integrante que se da de baja no cuenta
func TestVotarPoker_ElVotoDeQuienSeDaDeBajaNoCuenta(t *testing.T) {
	e := nuevoEscenarioVotos(t)
	e.votar(t, "Carolina", "5")
	e.votar(t, "Mariano", "8")
	mariano, _ := e.porNombre["Mariano"].DarDeBaja()
	if err := e.integrantes.Actualizar(context.Background(), mariano); err != nil {
		t.Fatalf("dar de baja: %v", err)
	}

	v, err := e.casos.Ver(context.Background(), e.proyecto, 1)
	if err != nil {
		t.Fatalf("ver: %v", err)
	}

	verificarVotaron(t, v.Votaron, e.votaron("Carolina"))
}

// Sin votos conectados (como HU-17) o sin sesión, nadie votó: el conjunto está vacío (no nil)
func TestVerPoker_SinVotos(t *testing.T) {
	e := nuevoEscenarioPoker(t)
	sinSesion, _ := e.casos.Ver(context.Background(), e.proyecto, 1)
	_, _ = e.casos.Abrir(context.Background(), e.proyecto, 1, e.porNombre["Juan"].ID)
	sinConectar, _ := e.casos.Ver(context.Background(), e.proyecto, 1)

	for _, v := range []service.VistaSesion{sinSesion, sinConectar} {
		if v.Votaron == nil || len(v.Votaron) != 0 {
			t.Errorf("Votaron = %#v, se esperaba un conjunto vacío", v.Votaron)
		}
	}
}

// CA-18.2, CA-18.3 y RN5 · la vista de la sesión no tiene dónde llevar un voto ni su valor: ningún campo es un
// poker.Voto ni se llama Valor (salvo el valor acordado de la sesión, de HU-22)
func TestVistaSesion_NoTieneValoresDeVotos(t *testing.T) {
	var revisar func(tipo reflect.Type, camino string)
	revisar = func(tipo reflect.Type, camino string) {
		switch tipo.Kind() {
		case reflect.Pointer, reflect.Slice, reflect.Array, reflect.Map:
			if tipo.Kind() == reflect.Map {
				revisar(tipo.Key(), camino+"[clave]")
			}
			revisar(tipo.Elem(), camino+"[]")
		case reflect.Struct:
			if tipo == reflect.TypeFor[poker.Voto]() {
				t.Errorf("%s es un poker.Voto", camino)
				return
			}
			for i := range tipo.NumField() {
				campo := tipo.Field(i)
				if campo.Name == "Valor" {
					t.Errorf("%s.Valor es un valor", camino)
				}
				revisar(campo.Type, camino+"."+campo.Name)
			}
		default:
		}
	}
	revisar(reflect.TypeFor[service.VistaSesion](), "VistaSesion")
}

// votosCaidos es un repositorio de votos con la base caída.
type votosCaidos struct{}

func (votosCaidos) Votar(context.Context, int64, int64, int) error { return errBaseCaida }

func (votosCaidos) QuienesVotaron(context.Context, int64) ([]int64, error) { return nil, errBaseCaida }

func TestVotarPoker_FallaLaBase(t *testing.T) {
	e := nuevoEscenarioPoker(t)
	_, _ = e.casos.Abrir(context.Background(), e.proyecto, 1, e.porNombre["Juan"].ID)
	e.casos.ConVotos(votosCaidos{})

	if _, err := e.casos.Votar(context.Background(), e.proyecto, 1, e.porNombre["Carolina"].ID, "5"); !errors.Is(err, errBaseCaida) {
		t.Errorf("votar: error = %v, se esperaba el de la base", err)
	}
	if _, err := e.casos.Ver(context.Background(), e.proyecto, 1); !errors.Is(err, errBaseCaida) {
		t.Errorf("ver: error = %v, se esperaba el de la base", err)
	}
}

// Si falla la base al buscar la sesión, no se vota
func TestVotarPoker_FallaLaBaseAlBuscarLaSesion(t *testing.T) {
	e := nuevoEscenarioPoker(t)
	casos := service.NuevoPoker(e.proyectos, e.historias, e.integrantes, pokerCaido{}).ConVotos(e.sesiones)

	if _, err := casos.Votar(context.Background(), e.proyecto, 1, e.porNombre["Carolina"].ID, "5"); !errors.Is(err, errBaseCaida) {
		t.Errorf("error = %v, se esperaba el de la base", err)
	}
}

func TestVotarPoker_HistoriaInexistente(t *testing.T) {
	e := nuevoEscenarioVotos(t)
	if _, err := e.casos.Votar(context.Background(), e.proyecto, 99, e.porNombre["Carolina"].ID, "5"); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("error = %v, se esperaba ErrHistoriaNoEncontrada", err)
	}
}
