package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
)

// Revelar la ronda (HU-19) y cerrar la sesión (HU-22) todavía no tienen método: estos tests las cambian por dentro,
// como los de PostgreSQL lo hacen con SQL.

// RN3 · en una ronda revelada no se vota, y el voto no se guarda
func TestPoker_Votar_RondaRevelada(t *testing.T) {
	ctx := context.Background()
	repo := NuevoPoker()
	_, ronda, _ := repo.Abrir(ctx, 7)
	repo.rondas[0].Revelada = true

	if err := repo.Votar(ctx, ronda.ID, 2, 5); !errors.Is(err, poker.ErrRondaRevelada) {
		t.Errorf("error = %v, se esperaba ErrRondaRevelada", err)
	}
	if votaron, _ := repo.QuienesVotaron(ctx, ronda.ID); len(votaron) != 0 {
		t.Errorf("QuienesVotaron = %v, no se tenía que guardar el voto", votaron)
	}
}

// En una sesión cerrada no se vota
func TestPoker_Votar_SesionCerrada(t *testing.T) {
	ctx := context.Background()
	repo := NuevoPoker()
	_, ronda, _ := repo.Abrir(ctx, 7)
	repo.sesiones[0].Estado = poker.EstadoCerrada

	if err := repo.Votar(ctx, ronda.ID, 2, 5); !errors.Is(err, poker.ErrSesionCerrada) {
		t.Errorf("error = %v, se esperaba ErrSesionCerrada", err)
	}
}
