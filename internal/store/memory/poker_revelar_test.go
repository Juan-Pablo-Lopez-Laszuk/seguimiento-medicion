package memory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// HU-19 CA-19.1 · revelar deja la ronda revelada; la de otra sesión no cambia
func TestPoker_RevelarGuardaLaRondaRevelada(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoPoker()
	_, ronda, _ := repo.Abrir(ctx, 7)
	_, otra, _ := repo.Abrir(ctx, 8)

	if err := repo.Revelar(ctx, ronda.ID); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if _, actual, _ := repo.SesionAbierta(ctx, 7); !actual.Revelada {
		t.Errorf("ronda = %+v, se esperaba revelada", actual)
	}
	if _, actual, _ := repo.SesionAbierta(ctx, 8); actual != otra {
		t.Errorf("ronda de otra sesión = %+v, se esperaba sin cambios %+v", actual, otra)
	}
}

// HU-19 · revelar una ronda que no existe
func TestPoker_RevelarUnaRondaQueNoExiste(t *testing.T) {
	repo := memory.NuevoPoker()

	err := repo.Revelar(context.Background(), 99)

	if !errors.Is(err, poker.ErrSesionNoEncontrada) {
		t.Errorf("error = %v, se esperaba poker.ErrSesionNoEncontrada", err)
	}
}
