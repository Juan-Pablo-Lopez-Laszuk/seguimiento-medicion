package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/postgres"
)

// baseParaCerrar guarda el Sprint 1 con HU-1 Hecha (5 SP), HU-2 En progreso (3 SP) y HU-3 En sprint (5 SP), y el
// Sprint 2 con HU-4 En sprint, como en los casos límite de la spec de HU-14. Devuelve los sprints y las historias.
func baseParaCerrar(t *testing.T, pool *pgxpool.Pool) ([]sprint.Sprint, []backlog.Historia) {
	t.Helper()
	ctx := context.Background()
	proyecto, _ := dosProyectos(t, pool)
	repoSprints, repoHistorias := postgres.NuevoSprints(pool), postgres.NuevoHistorias(pool)
	var sprints []sprint.Sprint
	for n := 1; n <= 2; n++ {
		s, err := repoSprints.Guardar(ctx, sprintDeEjemplo(proyecto, n, "Sprint"))
		if err != nil {
			t.Fatalf("guardar el Sprint %d: %v", n, err)
		}
		sprints = append(sprints, s)
	}
	var historias []backlog.Historia
	for i, h := range []struct {
		estado backlog.Estado
		sp     int
		sprint int
	}{{backlog.EstadoHecho, 5, 0}, {backlog.EstadoEnProgreso, 3, 0}, {backlog.EstadoEnSprint, 5, 0}, {backlog.EstadoEnSprint, 2, 1}} {
		nueva := historiaDeEjemplo(proyecto, i+1, "Historia")
		sp := h.sp
		nueva.StoryPoints = &sp
		guardada, err := repoHistorias.Guardar(ctx, nueva)
		if err != nil {
			t.Fatalf("guardar HU-%d: %v", i+1, err)
		}
		if err := repoHistorias.ActualizarSprint(ctx, guardada.ID, h.estado, &sprints[h.sprint].ID); err != nil {
			t.Fatalf("poner HU-%d en su sprint: %v", i+1, err)
		}
		guardada.Estado, guardada.SprintID = h.estado, &sprints[h.sprint].ID
		historias = append(historias, guardada)
	}
	return sprints, historias
}

// cierreDelSprint1 es lo que arma la regla de dominio al cerrar el Sprint 1 de baseParaCerrar.
func cierreDelSprint1(sprints []sprint.Sprint, historias []backlog.Historia) sprint.Cierre {
	s := sprints[0]
	planificados, completados := 13, 5
	s.Estado, s.SPPlanificados, s.SPCompletados = sprint.EstadoCerrado, &planificados, &completados
	foto := make([]sprint.FotoHistoria, 3)
	for i, h := range historias[:3] {
		foto[i] = sprint.FotoHistoria{HistoriaID: h.ID, Numero: h.Numero, StoryPoints: h.StoryPoints, Completada: h.Estado == backlog.EstadoHecho}
	}
	devueltas := []backlog.Historia{historias[1], historias[2]}
	for i := range devueltas {
		devueltas[i].Estado, devueltas[i].SprintID = backlog.EstadoPendiente, nil
	}
	return sprint.Cierre{Sprint: s, Foto: foto, Devueltas: devueltas}
}

// HU-14 RN3 y RN4: el cierre guarda el sprint Cerrado con sus SP, la foto y las historias que vuelven al backlog, y se
// lee de vuelta
func TestCierres_Cerrar(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	sprints, historias := baseParaCerrar(t, pool)
	repo := postgres.NuevoCierres(pool)

	if err := repo.Cerrar(ctx, cierreDelSprint1(sprints, historias)); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	leidos, err := postgres.NuevoSprints(pool).ListarPorProyecto(ctx, sprints[0].ProyectoID)
	if err != nil {
		t.Fatalf("listar los sprints: %v", err)
	}
	s1, s2 := leidos[0], leidos[1]
	if s1.Estado != sprint.EstadoCerrado || s1.SPPlanificados == nil || *s1.SPPlanificados != 13 || s1.SPCompletados == nil || *s1.SPCompletados != 5 {
		t.Errorf("se esperaba el Sprint 1 Cerrado con 5 de 13 SP y se obtuvo %+v", s1)
	}
	if s2.Estado != sprint.EstadoPlanificado || s2.SPPlanificados != nil {
		t.Errorf("el Sprint 2 no tenía que cambiar y quedó %+v", s2)
	}
	repoHistorias := postgres.NuevoHistorias(pool)
	for i, esperado := range []backlog.Estado{backlog.EstadoHecho, backlog.EstadoPendiente, backlog.EstadoPendiente, backlog.EstadoEnSprint} {
		h, _ := repoHistorias.BuscarPorNumero(ctx, sprints[0].ProyectoID, i+1)
		conSprint := i == 0 || i == 3
		if h.Estado != esperado || (h.SprintID != nil) != conSprint || h.StoryPoints == nil {
			t.Errorf("HU-%d: se esperaba %s (con sprint: %v) y su estimación, y se obtuvo %+v", i+1, esperado, conSprint, h)
		}
	}

	foto, err := repo.Foto(ctx, s1.ID)
	if err != nil {
		t.Fatalf("leer la foto: %v", err)
	}
	if len(foto) != 3 || foto[0].Numero != 1 || !foto[0].Completada || foto[1].StoryPoints == nil || *foto[1].StoryPoints != 3 || foto[1].Completada || foto[2].HistoriaID != historias[2].ID {
		t.Errorf("la foto no tiene las tres historias del Sprint 1 con sus SP: %+v", foto)
	}
	if vacia, err := repo.Foto(ctx, s2.ID); err != nil || len(vacia) != 0 {
		t.Errorf("un sprint sin cerrar no tiene foto y se obtuvo %+v, %v", vacia, err)
	}
	// HU-06 RN6: la que volvió al backlog sigue siendo de un sprint cerrado, por la foto
	if cerrado, err := repoHistorias.EnSprintCerrado(ctx, historias[1].ID); err != nil || !cerrado {
		t.Errorf("HU-2 está en la foto del Sprint 1 cerrado y EnSprintCerrado dio %v, %v", cerrado, err)
	}
}

// HU-14 RN5 y RN6: un sprint cerrado no se cierra otra vez, y si algo falla no se guarda nada (una transacción)
func TestCierres_Rechazos_NoGuardanNada(t *testing.T) {
	ctx := context.Background()

	t.Run("sprint ya cerrado", func(t *testing.T) {
		pool := baseMigrada(t)
		sprints, historias := baseParaCerrar(t, pool)
		repo := postgres.NuevoCierres(pool)
		c := cierreDelSprint1(sprints, historias)
		if err := repo.Cerrar(ctx, c); err != nil {
			t.Fatalf("primer cierre: %v", err)
		}
		if err := repo.Cerrar(ctx, c); !errors.Is(err, sprint.ErrYaCerrado) {
			t.Errorf("se esperaba sprint.ErrYaCerrado y se obtuvo %v", err)
		}
	})

	t.Run("historia que no existe", func(t *testing.T) {
		pool := baseMigrada(t)
		sprints, historias := baseParaCerrar(t, pool)
		repo := postgres.NuevoCierres(pool)
		c := cierreDelSprint1(sprints, historias)
		c.Devueltas[1].ID = 99999
		if err := repo.Cerrar(ctx, c); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
			t.Fatalf("se esperaba backlog.ErrHistoriaNoEncontrada y se obtuvo %v", err)
		}
		leidos, _ := postgres.NuevoSprints(pool).ListarPorProyecto(ctx, sprints[0].ProyectoID)
		hu2, _ := postgres.NuevoHistorias(pool).BuscarPorNumero(ctx, sprints[0].ProyectoID, 2)
		foto, _ := repo.Foto(ctx, sprints[0].ID)
		if leidos[0].Estado != sprint.EstadoPlanificado || leidos[0].SPPlanificados != nil || hu2.Estado != backlog.EstadoEnProgreso || len(foto) != 0 {
			t.Errorf("un cierre que falla no tiene que guardar nada: sprint %+v, HU-2 %q, foto %d", leidos[0], hu2.Estado, len(foto))
		}
	})

	t.Run("sprint que no existe", func(t *testing.T) {
		pool := baseMigrada(t)
		sprints, historias := baseParaCerrar(t, pool)
		c := cierreDelSprint1(sprints, historias)
		c.Sprint.ID = 99999
		if err := postgres.NuevoCierres(pool).Cerrar(ctx, c); !errors.Is(err, sprint.ErrNoEncontrado) {
			t.Errorf("se esperaba sprint.ErrNoEncontrado y se obtuvo %v", err)
		}
	})
}
