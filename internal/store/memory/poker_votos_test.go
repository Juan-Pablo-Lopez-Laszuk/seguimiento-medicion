package memory_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// CA-18.1 y CA-18.2 · se guarda el voto y el repositorio devuelve quiénes votaron, de menor a mayor ID
func TestPoker_VotarYQuienesVotaron(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoPoker()
	_, ronda, _ := repo.Abrir(ctx, 7)

	for _, v := range []struct {
		integrante int64
		valor      int
	}{{3, 8}, {1, 5}} {
		if err := repo.Votar(ctx, ronda.ID, v.integrante, v.valor); err != nil {
			t.Fatalf("votar: %v", err)
		}
	}

	votaron, err := repo.QuienesVotaron(ctx, ronda.ID)
	if err != nil || !slices.Equal(votaron, []int64{1, 3}) {
		t.Errorf("QuienesVotaron = %v, %v; se esperaba [1 3]", votaron, err)
	}
}

// CA-18.1 y RN1 · votar de nuevo en la misma ronda reemplaza el voto: sigue habiendo uno
func TestPoker_VotarDeNuevo_ReemplazaElVoto(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoPoker()
	_, ronda, _ := repo.Abrir(ctx, 7)
	_ = repo.Votar(ctx, ronda.ID, 2, 5)

	if err := repo.Votar(ctx, ronda.ID, 2, 8); err != nil {
		t.Fatalf("votar de nuevo: %v", err)
	}

	if votaron, _ := repo.QuienesVotaron(ctx, ronda.ID); !slices.Equal(votaron, []int64{2}) {
		t.Errorf("QuienesVotaron = %v, se esperaba un solo voto de 2", votaron)
	}
}

// Los votos son de cada ronda: los de otra historia no se mezclan; sin votos, la lista está vacía (no nil)
func TestPoker_QuienesVotaron_SoloLaRonda(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoPoker()
	_, login, _ := repo.Abrir(ctx, 7)
	_, perfil, _ := repo.Abrir(ctx, 8)
	_ = repo.Votar(ctx, login.ID, 2, 5)

	votaron, err := repo.QuienesVotaron(ctx, perfil.ID)
	if err != nil || votaron == nil || len(votaron) != 0 {
		t.Errorf("QuienesVotaron = %#v, %v; se esperaba una lista vacía", votaron, err)
	}
}

// RN3 · no se vota en una ronda que no existe
func TestPoker_Votar_RondaInexistente(t *testing.T) {
	if err := memory.NuevoPoker().Votar(context.Background(), 99, 2, 5); !errors.Is(err, poker.ErrSesionNoEncontrada) {
		t.Errorf("error = %v, se esperaba ErrSesionNoEncontrada", err)
	}
}
