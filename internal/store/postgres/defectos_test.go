package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/defect"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/postgres"
)

// HU-27 · los defectos se guardan con todos sus campos (historia opcional) y se listan por proyecto
func TestDefectos_GuardarYListarPorProyecto(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyecto, _ := dosProyectos(t, pool)
	h, err := postgres.NuevoHistorias(pool).Guardar(ctx, historiaDeEjemplo(proyecto, 1, "Primera"))
	if err != nil {
		t.Fatalf("guardar la historia: %v", err)
	}
	historia := h.ID
	s, err := postgres.NuevoSprints(pool).Guardar(ctx, sprint.Sprint{ProyectoID: proyecto, Numero: 1, Objetivo: "MVP",
		FechaInicio: time.Date(2026, 10, 5, 0, 0, 0, 0, time.UTC), FechaFin: time.Date(2026, 10, 11, 0, 0, 0, 0, time.UTC), Estado: sprint.EstadoPlanificado})
	if err != nil {
		t.Fatalf("guardar el sprint: %v", err)
	}
	repo := postgres.NuevoDefectos(pool)

	a, err := repo.Guardar(ctx, defect.Defecto{ProyectoID: proyecto, Descripcion: "Sin historia", Severidad: defect.SeveridadCritica, Estado: defect.EstadoAbierto, SprintDeteccionID: s.ID})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	b, err := repo.Guardar(ctx, defect.Defecto{ProyectoID: proyecto, HistoriaID: historia, Descripcion: "Con historia", Severidad: defect.SeveridadAlta, Estado: defect.EstadoAbierto, SprintDeteccionID: s.ID})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	lista, err := repo.ListarPorProyecto(ctx, proyecto)
	if err != nil || len(lista) != 2 || lista[0] != b || lista[1] != a {
		t.Errorf("se esperaba [b, a] con todos sus campos y se obtuvo %+v, %v", lista, err)
	}
	if _, err := repo.Guardar(ctx, defect.Defecto{ProyectoID: proyecto, Descripcion: "x", Severidad: defect.SeveridadAlta, Estado: defect.EstadoAbierto, SprintDeteccionID: s.ID + 99}); err == nil {
		t.Error("la base debería rechazar un sprint que no existe")
	}
}
