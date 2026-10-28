package service_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// pokerConRevelar es el repositorio en memoria con la ronda actual revelada o no, a pedido: revelar es de HU-19, y
// para HU-22 alcanza con que la ronda llegue revelada.
type pokerConRevelar struct {
	*memory.Poker
	revelada bool
}

func (p *pokerConRevelar) SesionAbierta(ctx context.Context, historiaID int64) (poker.Sesion, poker.Ronda, error) {
	s, r, err := p.Poker.SesionAbierta(ctx, historiaID)
	r.Revelada = r.Revelada || p.revelada
	return s, r, err
}

// nuevoEscenarioAcordar es el de HU-17 con la sesión de HU-1 Login abierta por Juan y el caso de uso de HU-22
// conectado. La ronda empieza sin revelar.
func nuevoEscenarioAcordar(t *testing.T) (*escenarioPoker, *pokerConRevelar) {
	t.Helper()
	e := nuevoEscenarioPoker(t)
	repo := &pokerConRevelar{Poker: e.sesiones.ConHistorias(e.historias)}
	e.casos.ConAcordar(repo)
	if _, err := e.casos.Abrir(context.Background(), e.proyecto, 1, e.porNombre["Juan"].ID); err != nil {
		t.Fatalf("abrir la sesión: %v", err)
	}
	return e, repo
}

// CA-22.2 CA-22.3 · con la ronda revelada, el Agile Enabler registra el valor: la historia queda estimada y la sesión
// cerrada y consultable
func TestAcordarPoker_ElAgileEnablerRegistraLaEstimacion(t *testing.T) {
	ctx := context.Background()
	e, repo := nuevoEscenarioAcordar(t)
	repo.revelada = true

	h, err := e.casos.Acordar(ctx, e.proyecto, 1, e.porNombre["Juan"].ID, "5")
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if h.StoryPoints == nil || *h.StoryPoints != 5 {
		t.Errorf("historia devuelta con %v story points, se esperaba 5", h.StoryPoints)
	}

	v, err := e.casos.Ver(ctx, e.proyecto, 1)
	if err != nil {
		t.Fatal(err)
	}
	if v.Abierta {
		t.Error("la sesión tenía que quedar cerrada")
	}
	if v.Historia.StoryPoints == nil || *v.Historia.StoryPoints != 5 {
		t.Errorf("la historia guardada tiene %v story points, se esperaba 5", v.Historia.StoryPoints)
	}
	if quiero := []poker.SesionCerrada{{Numero: 1, Rondas: 1, ValorAcordado: 5}}; !slices.Equal(v.Cerradas, quiero) {
		t.Errorf("sesiones cerradas = %+v, se esperaba %+v", v.Cerradas, quiero)
	}
}

// CA-22.1 · un valor fuera de la escala no se registra: la sesión sigue abierta y la historia sin estimar
func TestAcordarPoker_ValorFueraDeLaEscala(t *testing.T) {
	ctx := context.Background()
	e, repo := nuevoEscenarioAcordar(t)
	repo.revelada = true

	_, err := e.casos.Acordar(ctx, e.proyecto, 1, e.porNombre["Juan"].ID, "4")
	var errs domain.ErroresValidacion
	if !errors.As(err, &errs) || !errors.Is(errs[backlog.CampoStoryPoints], backlog.ErrStoryPointsInvalidos) {
		t.Fatalf("error = %v, se esperaba ErrStoryPointsInvalidos en story_points", err)
	}
	quedaComoEstaba(t, e)
}

// RN3 RN4 · los errores de las reglas (quién registra y la ronda revelada) no guardan nada
func TestAcordarPoker_ErroresDeLasReglas(t *testing.T) {
	casos := []struct {
		nombre   string
		quien    string
		quienID  int64
		revelada bool
		err      error
	}{
		{"un Product Builder", "Carolina", 0, true, poker.ErrSoloAgileEnablerAcuerda},
		{"una Product Builder dada de baja", "Lucía", 0, true, poker.ErrSoloAgileEnablerAcuerda},
		{"nadie (no se eligió quién)", "", 0, true, poker.ErrSoloAgileEnablerAcuerda},
		{"alguien de otro proyecto", "", 999, true, poker.ErrSoloAgileEnablerAcuerda},
		{"la ronda sin revelar", "Juan", 0, false, poker.ErrAcordarSinRevelar},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			e, repo := nuevoEscenarioAcordar(t)
			repo.revelada = c.revelada
			quienID := c.quienID
			if c.quien != "" {
				quienID = e.porNombre[c.quien].ID
			}
			if _, err := e.casos.Acordar(context.Background(), e.proyecto, 1, quienID, "5"); !errors.Is(err, c.err) {
				t.Fatalf("error = %v, se esperaba %v", err, c.err)
			}
			quedaComoEstaba(t, e)
		})
	}
}

// RN5 · sin una sesión abierta no hay nada que registrar: tampoco dos veces
func TestAcordarPoker_SinSesionAbierta(t *testing.T) {
	ctx := context.Background()
	e, repo := nuevoEscenarioAcordar(t)
	repo.revelada = true
	if _, err := e.casos.Acordar(ctx, e.proyecto, 1, e.porNombre["Juan"].ID, "5"); err != nil {
		t.Fatal(err)
	}

	if _, err := e.casos.Acordar(ctx, e.proyecto, 1, e.porNombre["Juan"].ID, "8"); !errors.Is(err, poker.ErrSesionNoEncontrada) {
		t.Errorf("error = %v, se esperaba ErrSesionNoEncontrada", err)
	}
	if v, _ := e.casos.Ver(ctx, e.proyecto, 1); *v.Historia.StoryPoints != 5 || len(v.Cerradas) != 1 {
		t.Errorf("la historia tenía que seguir con 5 y una sesión cerrada: %v, %+v", *v.Historia.StoryPoints, v.Cerradas)
	}
}

// RN7 · la historia Hecha no se estima, aunque tenga la sesión abierta
func TestAcordarPoker_HistoriaHecha(t *testing.T) {
	ctx := context.Background()
	e, repo := nuevoEscenarioAcordar(t)
	repo.revelada = true
	h, _ := e.historias.BuscarPorNumero(ctx, e.proyecto, 1)
	if err := e.historias.ActualizarEstado(ctx, h.ID, backlog.EstadoHecho); err != nil {
		t.Fatalf("pasar la historia a Hecha: %v", err)
	}

	if _, err := e.casos.Acordar(ctx, e.proyecto, 1, e.porNombre["Juan"].ID, "5"); !errors.Is(err, backlog.ErrEstimacionDeHistoriaHecha) {
		t.Fatalf("error = %v, se esperaba ErrEstimacionDeHistoriaHecha", err)
	}
	if v, _ := e.casos.Ver(ctx, e.proyecto, 1); !v.Abierta {
		t.Error("la sesión tenía que seguir abierta")
	}
}

// El proyecto y la historia tienen que existir
func TestAcordarPoker_ProyectoOHistoriaInexistente(t *testing.T) {
	ctx := context.Background()
	e, _ := nuevoEscenarioAcordar(t)
	if _, err := e.casos.Acordar(ctx, 999, 1, e.porNombre["Juan"].ID, "5"); !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("proyecto inexistente: error = %v", err)
	}
	if _, err := e.casos.Acordar(ctx, e.proyecto, 99, e.porNombre["Juan"].ID, "5"); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("historia inexistente: error = %v", err)
	}
}

// CA-22.3 · la vista de la sesión trae las sesiones cerradas de la historia, la más reciente primero, también con otra
// abierta
func TestVerPoker_TraeLasSesionesCerradas(t *testing.T) {
	ctx := context.Background()
	e, repo := nuevoEscenarioAcordar(t)
	repo.revelada = true
	juan := e.porNombre["Juan"].ID
	if _, err := e.casos.Acordar(ctx, e.proyecto, 1, juan, "5"); err != nil {
		t.Fatal(err)
	}
	if _, err := e.casos.Abrir(ctx, e.proyecto, 1, juan); err != nil {
		t.Fatal(err)
	}
	if _, err := e.casos.Acordar(ctx, e.proyecto, 1, juan, "8"); err != nil {
		t.Fatal(err)
	}
	if _, err := e.casos.Abrir(ctx, e.proyecto, 1, juan); err != nil {
		t.Fatal(err)
	}

	v, err := e.casos.Ver(ctx, e.proyecto, 1)
	quiero := []poker.SesionCerrada{{Numero: 2, Rondas: 1, ValorAcordado: 8}, {Numero: 1, Rondas: 1, ValorAcordado: 5}}
	if err != nil || !v.Abierta || !slices.Equal(v.Cerradas, quiero) {
		t.Errorf("Ver = abierta %v, cerradas %+v, %v; se esperaba la tercera abierta y %+v", v.Abierta, v.Cerradas, err, quiero)
	}
}

// quedaComoEstaba comprueba que la sesión de Login sigue abierta y la historia sin estimar.
func quedaComoEstaba(t *testing.T, e *escenarioPoker) {
	t.Helper()
	v, err := e.casos.Ver(context.Background(), e.proyecto, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !v.Abierta || v.Historia.StoryPoints != nil || len(v.Cerradas) != 0 {
		t.Errorf("tenía que quedar como estaba: abierta %v, story points %v, cerradas %+v", v.Abierta, v.Historia.StoryPoints, v.Cerradas)
	}
}
