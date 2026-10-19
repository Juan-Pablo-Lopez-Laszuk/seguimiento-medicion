package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/postgres"
)

func fechaDeRegistro(d int) time.Time { return time.Date(2026, 10, d, 0, 0, 0, 0, time.UTC) }

// HU-24 · los registros se guardan con todos sus campos y se listan por historia y por integrante
func TestRegistros_GuardarYListar(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	historiaA, historiaB, integrante := historiasConResponsable(t, pool)
	tarea, err := postgres.NuevoTareas(pool).Guardar(ctx, effort.Tarea{HistoriaID: historiaA, Titulo: "T", ResponsableID: integrante, HorasEstimadas: 2})
	if err != nil {
		t.Fatalf("guardar la tarea: %v", err)
	}
	repo := postgres.NuevoRegistros(pool)

	a, err := repo.Guardar(ctx, effort.Registro{IntegranteID: integrante, HistoriaID: historiaA, Fecha: fechaDeRegistro(16), Actividad: "A", Horas: 3})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	b, err := repo.Guardar(ctx, effort.Registro{IntegranteID: integrante, HistoriaID: historiaA, TareaID: tarea.ID, Fecha: fechaDeRegistro(17), Actividad: "B", Horas: 2.5})
	if err != nil {
		t.Fatalf("no se esperaba error con tarea: %v", err)
	}
	if _, err := repo.Guardar(ctx, effort.Registro{IntegranteID: integrante, HistoriaID: historiaB, Fecha: fechaDeRegistro(17), Actividad: "C", Horas: 1}); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	lista, err := repo.ListarPorHistoria(ctx, historiaA)
	if err != nil || len(lista) != 2 || lista[0] != b || lista[1] != a {
		t.Errorf("de la historia A se esperaba [b, a] con todos sus campos y se obtuvo %+v, %v", lista, err)
	}
	delIntegrante, err := repo.ListarPorIntegrante(ctx, integrante)
	if err != nil || len(delIntegrante) != 3 {
		t.Errorf("de la integrante se esperaban 3 registros y se obtuvo %d, %v", len(delIntegrante), err)
	}
}

// La base controla lo mismo que el dominio: horas > 0 y ≤ 24, y un integrante que exista.
func TestRegistros_LaBaseRechazaDatosInvalidos(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	historia, _, integrante := historiasConResponsable(t, pool)
	repo := postgres.NuevoRegistros(pool)

	for nombre, r := range map[string]effort.Registro{
		"25 horas":                 {IntegranteID: integrante, HistoriaID: historia, Fecha: fechaDeRegistro(16), Actividad: "A", Horas: 25},
		"integrante que no existe": {IntegranteID: integrante + 99, HistoriaID: historia, Fecha: fechaDeRegistro(16), Actividad: "A", Horas: 1},
	} {
		if _, err := repo.Guardar(ctx, r); err == nil {
			t.Errorf("%s: se esperaba que la base lo rechace", nombre)
		}
	}
}
