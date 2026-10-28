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

// HU-19 CA-19.2 · Votos devuelve los votos de la ronda con su valor, por integrante; los de otra ronda no
func TestPoker_VotosDeLaRonda(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoPoker()
	_, ronda, _ := repo.Abrir(ctx, 7)
	_, otra, _ := repo.Abrir(ctx, 8)
	for _, v := range []poker.Voto{{RondaID: ronda.ID, IntegranteID: 3, Valor: 5}, {RondaID: ronda.ID, IntegranteID: 1, Valor: 8}, {RondaID: otra.ID, IntegranteID: 2, Valor: 13}} {
		if err := repo.Votar(ctx, v.RondaID, v.IntegranteID, v.Valor); err != nil {
			t.Fatalf("votar: %v", err)
		}
	}

	votos, err := repo.Votos(ctx, ronda.ID)

	if err != nil || len(votos) != 2 || votos[0].IntegranteID != 1 || votos[0].Valor != 8 || votos[1].IntegranteID != 3 || votos[1].Valor != 5 {
		t.Errorf("Votos = %+v, %v; se esperaban los votos 8 (integrante 1) y 5 (integrante 3)", votos, err)
	}
}
