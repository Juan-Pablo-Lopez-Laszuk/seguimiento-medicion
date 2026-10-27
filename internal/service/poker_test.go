package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// escenarioPoker es el de features/hu-17-abrir-sesion-poker.feature: Juan (Agile Enabler), Carolina y Mariano activos,
// Lucía dada de baja; HU-1 Login Pendiente y HU-2 Reportes Hecha.
type escenarioPoker struct {
	casos       *service.Poker
	proyectos   *memory.Proyectos
	historias   *memory.Historias
	integrantes *memory.Integrantes
	sesiones    *memory.Poker
	proyecto    int64
	porNombre   map[string]member.Integrante
}

func nuevoEscenarioPoker(t *testing.T) *escenarioPoker {
	t.Helper()
	ctx := context.Background()
	proyectos, historias, integrantes, sesiones := memory.NuevoProyectos(), memory.NuevoHistorias(), memory.NuevoIntegrantes(), memory.NuevoPoker()
	p, err := service.NuevoProyectos(proyectos, relojFijo).Crear(ctx, datosValidos())
	if err != nil {
		t.Fatalf("crear el proyecto: %v", err)
	}
	e := &escenarioPoker{
		casos:       service.NuevoPoker(proyectos, historias, integrantes, sesiones),
		proyectos:   proyectos,
		historias:   historias,
		integrantes: integrantes,
		sesiones:    sesiones,
		proyecto:    p.ID,
		porNombre:   map[string]member.Integrante{},
	}
	for _, i := range []member.Integrante{
		{Nombre: "Juan", Rol: member.RolAgileEnabler, Activo: true},
		{Nombre: "Carolina", Rol: member.RolProductBuilder, Activo: true},
		{Nombre: "Mariano", Rol: member.RolProductBuilder, Activo: true},
		{Nombre: "Lucía", Rol: member.RolProductBuilder},
	} {
		i.ProyectoID = p.ID
		guardado, err := integrantes.Guardar(ctx, i)
		if err != nil {
			t.Fatalf("guardar a %s: %v", i.Nombre, err)
		}
		e.porNombre[i.Nombre] = guardado
	}
	for n, h := range []backlog.Historia{
		{Titulo: "Login", Estado: backlog.EstadoPendiente},
		{Titulo: "Reportes", Estado: backlog.EstadoHecho},
	} {
		h.ProyectoID, h.Numero, h.Prioridad = p.ID, n+1, backlog.PrioridadAlta
		if _, err := historias.Guardar(ctx, h); err != nil {
			t.Fatalf("guardar %q: %v", h.Titulo, err)
		}
	}
	return e
}

func nombresDe(lista []member.Integrante) []string {
	nombres := []string{}
	for _, i := range lista {
		nombres = append(nombres, i.Nombre)
	}
	return nombres
}

// CA-17.1, CA-17.3 y RN3 · el Agile Enabler abre la sesión con la Ronda 1 y participan los integrantes activos
func TestAbrirPoker_ElAgileEnablerAbreLaSesion(t *testing.T) {
	e := nuevoEscenarioPoker(t)

	v, err := e.casos.Abrir(context.Background(), e.proyecto, 1, e.porNombre["Juan"].ID)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if !v.Abierta || v.Sesion.Estado != poker.EstadoAbierta || v.Ronda.Numero != 1 {
		t.Errorf("se esperaba la sesión abierta en la Ronda 1 y se obtuvo %+v", v)
	}
	if v.Historia.Titulo != "Login" || v.Proyecto.ID != e.proyecto || v.Sesion.HistoriaID != v.Historia.ID {
		t.Errorf("historia = %+v, proyecto = %+v", v.Historia, v.Proyecto)
	}
	verificarIgual(t, "participantes", nombresDe(v.Participantes), []string{"Juan", "Carolina", "Mariano"})
	if guardada, _, err := e.sesiones.SesionAbierta(context.Background(), v.Historia.ID); err != nil || guardada.ID != v.Sesion.ID {
		t.Errorf("la sesión no quedó guardada: %+v, %v", guardada, err)
	}
}

// Errores al abrir: no se guarda ninguna sesión
func TestAbrirPoker_Errores(t *testing.T) {
	casos := []struct {
		nombre string
		numero int
		quien  string
		err    error
	}{
		{"historia Hecha (CA-17.1)", 2, "Juan", poker.ErrHistoriaHecha},
		{"la abre un Product Builder (RN4)", 1, "Carolina", poker.ErrSoloAgileEnabler},
		{"la abre una integrante dada de baja (RN4)", 1, "Lucía", poker.ErrSoloAgileEnabler},
		{"no se eligió quién la abre (RN4)", 1, "", poker.ErrSoloAgileEnabler},
		{"historia inexistente", 99, "Juan", backlog.ErrHistoriaNoEncontrada},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			e := nuevoEscenarioPoker(t)

			_, err := e.casos.Abrir(context.Background(), e.proyecto, c.numero, e.porNombre[c.quien].ID)

			if !errors.Is(err, c.err) {
				t.Errorf("error = %v, se esperaba %v", err, c.err)
			}
			for _, historiaID := range []int64{1, 2} {
				if e.sesiones.TieneDeHistoria(historiaID) {
					t.Errorf("no se esperaba ninguna sesión guardada para la historia %d", historiaID)
				}
			}
		})
	}
}

// RN4 · un integrante de otro proyecto no abre la sesión aunque sea Agile Enabler allá
func TestAbrirPoker_IntegranteDeOtroProyecto(t *testing.T) {
	e := nuevoEscenarioPoker(t)
	otro, _ := e.integrantes.Guardar(context.Background(), member.Integrante{ProyectoID: e.proyecto + 1, Nombre: "Ana", Rol: member.RolAgileEnabler, Activo: true})

	if _, err := e.casos.Abrir(context.Background(), e.proyecto, 1, otro.ID); !errors.Is(err, poker.ErrSoloAgileEnabler) {
		t.Errorf("error = %v, se esperaba ErrSoloAgileEnabler", err)
	}
}

// CA-17.2 · una segunda sesión para la misma historia da el error y la primera sigue
func TestAbrirPoker_DosVeces(t *testing.T) {
	e := nuevoEscenarioPoker(t)
	primera, _ := e.casos.Abrir(context.Background(), e.proyecto, 1, e.porNombre["Juan"].ID)

	_, err := e.casos.Abrir(context.Background(), e.proyecto, 1, e.porNombre["Juan"].ID)

	if !errors.Is(err, poker.ErrSesionYaAbierta) {
		t.Errorf("error = %v, se esperaba ErrSesionYaAbierta", err)
	}
	if v, _ := e.casos.Ver(context.Background(), e.proyecto, 1); v.Sesion.ID != primera.Sesion.ID {
		t.Errorf("la sesión abierta tenía que seguir siendo la %d y es %+v", primera.Sesion.ID, v.Sesion)
	}
}

func TestAbrirPoker_ProyectoInexistente(t *testing.T) {
	e := nuevoEscenarioPoker(t)
	if _, err := e.casos.Abrir(context.Background(), 999, 1, e.porNombre["Juan"].ID); !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("error = %v, se esperaba project.ErrNoEncontrado", err)
	}
}

// RN5 · ver la sesión muestra los participantes de ese momento: quien se da de baja deja de participar
func TestVerPoker_LosParticipantesSonLosActivosDeCadaMomento(t *testing.T) {
	e := nuevoEscenarioPoker(t)
	_, _ = e.casos.Abrir(context.Background(), e.proyecto, 1, e.porNombre["Juan"].ID)
	mariano, _ := e.porNombre["Mariano"].DarDeBaja()
	if err := e.integrantes.Actualizar(context.Background(), mariano); err != nil {
		t.Fatalf("dar de baja: %v", err)
	}

	v, err := e.casos.Ver(context.Background(), e.proyecto, 1)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if !v.Abierta || v.Ronda.Numero != 1 {
		t.Errorf("se esperaba la sesión abierta en la Ronda 1 y se obtuvo %+v", v)
	}
	verificarIgual(t, "participantes", nombresDe(v.Participantes), []string{"Juan", "Carolina"})
}

// Sin sesión abierta, la vista lo dice (Abierta en false) y trae la historia y los integrantes activos, para ofrecer
// abrirla
func TestVerPoker_SinSesion(t *testing.T) {
	e := nuevoEscenarioPoker(t)

	v, err := e.casos.Ver(context.Background(), e.proyecto, 1)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if v.Abierta || v.Sesion.ID != 0 || v.Historia.Titulo != "Login" {
		t.Errorf("se esperaba la historia sin sesión y se obtuvo %+v", v)
	}
	verificarIgual(t, "participantes", nombresDe(v.Participantes), []string{"Juan", "Carolina", "Mariano"})
}

func TestVerPoker_HistoriaOProyectoInexistente(t *testing.T) {
	e := nuevoEscenarioPoker(t)
	if _, err := e.casos.Ver(context.Background(), e.proyecto, 99); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("historia inexistente: error = %v", err)
	}
	if _, err := e.casos.Ver(context.Background(), 999, 1); !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("proyecto inexistente: error = %v", err)
	}
}

// pokerCaido es un repositorio de sesiones con la base caída.
type pokerCaido struct{}

func (pokerCaido) Abrir(context.Context, int64) (poker.Sesion, poker.Ronda, error) {
	return poker.Sesion{}, poker.Ronda{}, errBaseCaida
}

func (pokerCaido) SesionAbierta(context.Context, int64) (poker.Sesion, poker.Ronda, error) {
	return poker.Sesion{}, poker.Ronda{}, errBaseCaida
}

func TestPoker_FallaLaBase(t *testing.T) {
	ctx := context.Background()
	proyectos, historias, integrantes := memory.NuevoProyectos(), memory.NuevoHistorias(), memory.NuevoIntegrantes()
	p, _ := service.NuevoProyectos(proyectos, relojFijo).Crear(ctx, datosValidos())
	_, _ = historias.Guardar(ctx, backlog.Historia{ProyectoID: p.ID, Numero: 1, Titulo: "Login", Estado: backlog.EstadoPendiente})
	juan, _ := integrantes.Guardar(ctx, member.Integrante{ProyectoID: p.ID, Nombre: "Juan", Rol: member.RolAgileEnabler, Activo: true})
	casos := service.NuevoPoker(proyectos, historias, integrantes, pokerCaido{})

	if _, err := casos.Abrir(ctx, p.ID, 1, juan.ID); !errors.Is(err, errBaseCaida) {
		t.Errorf("abrir: error = %v, se esperaba el de la base", err)
	}
	if _, err := casos.Ver(ctx, p.ID, 1); !errors.Is(err, errBaseCaida) {
		t.Errorf("ver: error = %v, se esperaba el de la base", err)
	}
}
