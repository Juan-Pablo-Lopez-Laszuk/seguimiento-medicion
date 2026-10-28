package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
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

// HU-19 CA-19.2 · Votos devuelve los votos de la ronda con su valor, por integrante; los de otra ronda no
func TestPoker_VotosDeLaRonda(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	h1, _ := postgres.NuevoHistorias(pool).Guardar(ctx, historiaDeEjemplo(proyectoA, 1, "Login"))
	h2, _ := postgres.NuevoHistorias(pool).Guardar(ctx, historiaDeEjemplo(proyectoA, 2, "Logout"))
	integrantes := postgres.NuevoIntegrantes(pool)
	var ids []int64
	for n, nombre := range []string{"Juan", "Carolina"} {
		i, err := integrantes.Guardar(ctx, integranteDeEjemplo(proyectoA, nombre, fmt.Sprintf("poker%d@mail.com", n), member.RolProductBuilder))
		if err != nil {
			t.Fatalf("guardar el integrante: %v", err)
		}
		ids = append(ids, i.ID)
	}
	repo := postgres.NuevoPoker(pool)
	_, ronda, _ := repo.Abrir(ctx, h1.ID)
	_, otra, _ := repo.Abrir(ctx, h2.ID)
	for _, v := range []poker.Voto{{RondaID: ronda.ID, IntegranteID: ids[1], Valor: 5}, {RondaID: ronda.ID, IntegranteID: ids[0], Valor: 8}, {RondaID: otra.ID, IntegranteID: ids[0], Valor: 13}} {
		if err := repo.Votar(ctx, v.RondaID, v.IntegranteID, v.Valor); err != nil {
			t.Fatalf("votar: %v", err)
		}
	}

	votos, err := repo.Votos(ctx, ronda.ID)

	if err != nil || len(votos) != 2 || votos[0].IntegranteID != ids[0] || votos[0].Valor != 8 || votos[1].IntegranteID != ids[1] || votos[1].Valor != 5 {
		t.Errorf("Votos = %+v, %v; se esperaban los votos 8 y 5 de los integrantes %v", votos, err, ids)
	}
}
