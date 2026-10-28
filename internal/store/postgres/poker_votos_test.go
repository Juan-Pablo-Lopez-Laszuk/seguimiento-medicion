package postgres_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/postgres"
)

// sesionParaVotar abre una sesión para una historia nueva y registra tres integrantes en el proyecto. Devuelve el
// repositorio, la sesión, la Ronda 1 y los IDs de los integrantes.
func sesionParaVotar(t *testing.T, pool *pgxpool.Pool) (*postgres.Poker, poker.Sesion, poker.Ronda, []int64) {
	t.Helper()
	ctx := context.Background()
	proyecto, _ := dosProyectos(t, pool)
	h, err := postgres.NuevoHistorias(pool).Guardar(ctx, historiaDeEjemplo(proyecto, 1, "Login"))
	if err != nil {
		t.Fatalf("guardar la historia: %v", err)
	}
	ids := []int64{}
	for _, nombre := range []string{"Juan", "Carolina", "Mariano"} {
		i, err := postgres.NuevoIntegrantes(pool).Guardar(ctx,
			integranteDeEjemplo(proyecto, nombre, nombre+"@mail.com", member.RolProductBuilder))
		if err != nil {
			t.Fatalf("guardar a %s: %v", nombre, err)
		}
		ids = append(ids, i.ID)
	}
	repo := postgres.NuevoPoker(pool)
	sesion, ronda, err := repo.Abrir(ctx, h.ID)
	if err != nil {
		t.Fatalf("abrir la sesión: %v", err)
	}
	return repo, sesion, ronda, ids
}

// CA-18.1 y CA-18.2 · se guarda el voto y el repositorio devuelve quiénes votaron, de menor a mayor ID
func TestPoker_VotarYQuienesVotaron(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	repo, _, ronda, ids := sesionParaVotar(t, pool)

	for _, i := range []int{2, 0} {
		if err := repo.Votar(ctx, ronda.ID, ids[i], 5); err != nil {
			t.Fatalf("votar: %v", err)
		}
	}

	votaron, err := repo.QuienesVotaron(ctx, ronda.ID)
	if esperado := []int64{ids[0], ids[2]}; err != nil || !slices.Equal(votaron, esperado) {
		t.Errorf("QuienesVotaron = %v, %v; se esperaba %v", votaron, err, esperado)
	}
}

// CA-18.1 y RN1 · votar de nuevo en la misma ronda reemplaza el voto (UPDATE): queda una fila con el último valor
func TestPoker_VotarDeNuevo_ReemplazaElVoto(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	repo, _, ronda, ids := sesionParaVotar(t, pool)
	_ = repo.Votar(ctx, ronda.ID, ids[1], 5)

	if err := repo.Votar(ctx, ronda.ID, ids[1], 8); err != nil {
		t.Fatalf("votar de nuevo: %v", err)
	}

	if n := cuantos(t, pool, "votos", "ronda_id = $1", ronda.ID); n != 1 {
		t.Errorf("se esperaba 1 voto y hay %d", n)
	}
	if n := cuantos(t, pool, "votos", "ronda_id = $1 AND integrante_id = $2 AND valor = 8", ronda.ID, ids[1]); n != 1 {
		t.Errorf("el voto tenía que quedar en 8")
	}
}

// Sin votos, la lista está vacía (no nil)
func TestPoker_QuienesVotaron_Nadie(t *testing.T) {
	pool := baseMigrada(t)
	repo, _, ronda, _ := sesionParaVotar(t, pool)

	votaron, err := repo.QuienesVotaron(context.Background(), ronda.ID)
	if err != nil || votaron == nil || len(votaron) != 0 {
		t.Errorf("QuienesVotaron = %#v, %v; se esperaba una lista vacía", votaron, err)
	}
}

// RN3 · en una ronda revelada no se vota, y el voto no se guarda (se controla en la misma sentencia que guarda)
func TestPoker_Votar_RondaRevelada(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	repo, _, ronda, ids := sesionParaVotar(t, pool)
	if _, err := pool.Exec(ctx, `UPDATE rondas_poker SET revelada = true WHERE id = $1`, ronda.ID); err != nil {
		t.Fatalf("revelar la ronda: %v", err)
	}

	if err := repo.Votar(ctx, ronda.ID, ids[1], 5); !errors.Is(err, poker.ErrRondaRevelada) {
		t.Errorf("error = %v, se esperaba ErrRondaRevelada", err)
	}
	if n := cuantos(t, pool, "votos", "true"); n != 0 {
		t.Errorf("no se tenía que guardar el voto y hay %d", n)
	}
}

// En una sesión cerrada no se vota
func TestPoker_Votar_SesionCerrada(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	repo, sesion, ronda, ids := sesionParaVotar(t, pool)
	if _, err := pool.Exec(ctx, `UPDATE sesiones_poker SET estado = 'Cerrada', valor_acordado = 5 WHERE id = $1`, sesion.ID); err != nil {
		t.Fatalf("cerrar la sesión: %v", err)
	}

	if err := repo.Votar(ctx, ronda.ID, ids[1], 5); !errors.Is(err, poker.ErrSesionCerrada) {
		t.Errorf("error = %v, se esperaba ErrSesionCerrada", err)
	}
}

// RN3 · no se vota en una ronda que no existe
func TestPoker_Votar_RondaInexistente(t *testing.T) {
	pool := baseMigrada(t)
	_, _, _, ids := sesionParaVotar(t, pool)
	if err := postgres.NuevoPoker(pool).Votar(context.Background(), 9999, ids[1], 5); !errors.Is(err, poker.ErrSesionNoEncontrada) {
		t.Errorf("error = %v, se esperaba ErrSesionNoEncontrada", err)
	}
}
