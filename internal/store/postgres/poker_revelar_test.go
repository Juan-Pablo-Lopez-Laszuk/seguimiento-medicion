package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/postgres"
)

// HU-19 CA-19.1 · revelar deja la ronda revelada; la de otra sesión no cambia
func TestPoker_RevelarGuardaLaRondaRevelada(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	h1, _ := postgres.NuevoHistorias(pool).Guardar(ctx, historiaDeEjemplo(proyectoA, 1, "Login"))
	h2, _ := postgres.NuevoHistorias(pool).Guardar(ctx, historiaDeEjemplo(proyectoA, 2, "Logout"))
	repo := postgres.NuevoPoker(pool)
	_, ronda, _ := repo.Abrir(ctx, h1.ID)
	_, otra, _ := repo.Abrir(ctx, h2.ID)

	if err := repo.Revelar(ctx, ronda.ID); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if _, actual, _ := repo.SesionAbierta(ctx, h1.ID); !actual.Revelada {
		t.Errorf("ronda = %+v, se esperaba revelada", actual)
	}
	if _, actual, _ := repo.SesionAbierta(ctx, h2.ID); actual != otra {
		t.Errorf("ronda de otra sesión = %+v, se esperaba sin cambios %+v", actual, otra)
	}
}

// HU-19 · revelar una ronda que no existe
func TestPoker_RevelarUnaRondaQueNoExiste(t *testing.T) {
	repo := postgres.NuevoPoker(baseMigrada(t))

	err := repo.Revelar(context.Background(), 99)

	if !errors.Is(err, poker.ErrSesionNoEncontrada) {
		t.Errorf("error = %v, se esperaba poker.ErrSesionNoEncontrada", err)
	}
}
