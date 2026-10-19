package memory_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// escenarioCierre tiene el Sprint 1 (12/10 al 18/10) con HU-1 Hecha (5 SP), HU-2 En progreso (3 SP) y HU-3 En sprint
// (5 SP), el Sprint 2 con HU-4 y HU-5 en el backlog, como en los casos límite de la spec de HU-14.
type escenarioCierre struct {
	sprints   *memory.Sprints
	historias *memory.Historias
	cierres   *memory.Cierres
	sprint1   sprint.Sprint
	todas     []backlog.Historia
}

func nuevoEscenarioCierre(t *testing.T) escenarioCierre {
	t.Helper()
	ctx := context.Background()
	e := escenarioCierre{sprints: memory.NuevoSprints(), historias: memory.NuevoHistorias()}
	e.cierres = memory.NuevoCierres(e.sprints, e.historias)
	dia := func(d int) time.Time { return time.Date(2026, time.October, d, 0, 0, 0, 0, time.UTC) }
	e.sprint1, _ = e.sprints.Guardar(ctx, sprint.Sprint{ProyectoID: 1, Numero: 1, FechaInicio: dia(12), FechaFin: dia(18), Estado: sprint.EstadoPlanificado})
	sprint2, _ := e.sprints.Guardar(ctx, sprint.Sprint{ProyectoID: 1, Numero: 2, FechaInicio: dia(19), FechaFin: dia(25), Estado: sprint.EstadoPlanificado})
	for _, h := range []struct {
		estado backlog.Estado
		sp     int
		sprint *int64
	}{
		{backlog.EstadoHecho, 5, &e.sprint1.ID},
		{backlog.EstadoEnProgreso, 3, &e.sprint1.ID},
		{backlog.EstadoEnSprint, 5, &e.sprint1.ID},
		{backlog.EstadoEnSprint, 2, &sprint2.ID},
		{backlog.EstadoPendiente, 8, nil},
	} {
		sp := h.sp
		guardada, _ := e.historias.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: len(e.todas) + 1, Titulo: "Historia", Estado: h.estado, StoryPoints: &sp, SprintID: h.sprint})
		e.todas = append(e.todas, guardada)
	}
	return e
}

// cierre arma el cierre del Sprint 1 con la regla de dominio, el 18/10.
func (e escenarioCierre) cierre(t *testing.T) sprint.Cierre {
	t.Helper()
	sprints, _ := e.sprints.ListarPorProyecto(context.Background(), 1)
	c, err := sprint.Cerrar(e.sprint1, sprints, e.todas, time.Date(2026, time.October, 18, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("armar el cierre: %v", err)
	}
	return c
}

// HU-14 RN3 y RN4: el cierre guarda el sprint Cerrado con sus SP, la foto y las historias que vuelven al backlog
func TestCierres_Cerrar(t *testing.T) {
	ctx := context.Background()
	e := nuevoEscenarioCierre(t)

	if err := e.cierres.Cerrar(ctx, e.cierre(t)); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	sprints, _ := e.sprints.ListarPorProyecto(ctx, 1)
	s1, s2 := sprints[0], sprints[1]
	if s1.Estado != sprint.EstadoCerrado || s1.SPPlanificados == nil || *s1.SPPlanificados != 13 || s1.SPCompletados == nil || *s1.SPCompletados != 5 {
		t.Errorf("se esperaba el Sprint 1 Cerrado con 5 de 13 SP y se obtuvo %+v", s1)
	}
	if s2.Estado != sprint.EstadoPlanificado || s2.SPPlanificados != nil {
		t.Errorf("el Sprint 2 no tenía que cambiar y quedó %+v", s2)
	}
	historias, _ := e.historias.ListarPorProyecto(ctx, 1)
	for i, esperada := range []struct {
		estado backlog.Estado
		sprint *int64
	}{
		{backlog.EstadoHecho, &e.sprint1.ID},
		{backlog.EstadoPendiente, nil},
		{backlog.EstadoPendiente, nil},
		{backlog.EstadoEnSprint, e.todas[3].SprintID},
		{backlog.EstadoPendiente, nil},
	} {
		h := historias[i]
		mismoSprint := (h.SprintID == nil && esperada.sprint == nil) || (h.SprintID != nil && esperada.sprint != nil && *h.SprintID == *esperada.sprint)
		if h.Estado != esperada.estado || !mismoSprint || h.StoryPoints == nil || *h.StoryPoints != *e.todas[i].StoryPoints {
			t.Errorf("HU-%d: se esperaba %s en el sprint %v con su estimación, y se obtuvo %+v", i+1, esperada.estado, esperada.sprint, h)
		}
	}

	foto, err := e.cierres.Foto(ctx, e.sprint1.ID)
	if err != nil {
		t.Fatalf("leer la foto: %v", err)
	}
	if len(foto) != 3 || foto[0].Numero != 1 || !foto[0].Completada || *foto[1].StoryPoints != 3 || foto[1].Completada || foto[2].HistoriaID != e.todas[2].ID {
		t.Errorf("la foto no tiene las tres historias del Sprint 1 con sus SP: %+v", foto)
	}
	// lo que devuelve el repositorio no se comparte con lo guardado
	*foto[1].StoryPoints = 21
	if otra, _ := e.cierres.Foto(ctx, e.sprint1.ID); *otra[1].StoryPoints != 3 {
		t.Errorf("modificar la foto devuelta no tenía que cambiar la guardada: %d", *otra[1].StoryPoints)
	}
	if vacia, err := e.cierres.Foto(ctx, s2.ID); err != nil || len(vacia) != 0 {
		t.Errorf("un sprint sin cerrar no tiene foto y se obtuvo %+v, %v", vacia, err)
	}
}

// HU-14 RN5 y RN6: un sprint cerrado no se cierra otra vez, y si algo falla no se guarda nada
func TestCierres_Rechazos_NoGuardanNada(t *testing.T) {
	ctx := context.Background()

	t.Run("sprint ya cerrado", func(t *testing.T) {
		e := nuevoEscenarioCierre(t)
		c := e.cierre(t)
		if err := e.cierres.Cerrar(ctx, c); err != nil {
			t.Fatalf("primer cierre: %v", err)
		}
		if err := e.cierres.Cerrar(ctx, c); !errors.Is(err, sprint.ErrYaCerrado) {
			t.Errorf("se esperaba sprint.ErrYaCerrado y se obtuvo %v", err)
		}
	})

	t.Run("historia que no existe", func(t *testing.T) {
		e := nuevoEscenarioCierre(t)
		c := e.cierre(t)
		c.Devueltas[1].ID = 99
		if err := e.cierres.Cerrar(ctx, c); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
			t.Fatalf("se esperaba backlog.ErrHistoriaNoEncontrada y se obtuvo %v", err)
		}
		sprints, _ := e.sprints.ListarPorProyecto(ctx, 1)
		historias, _ := e.historias.ListarPorProyecto(ctx, 1)
		foto, _ := e.cierres.Foto(ctx, e.sprint1.ID)
		if sprints[0].Estado != sprint.EstadoPlanificado || historias[1].Estado != backlog.EstadoEnProgreso || len(foto) != 0 {
			t.Errorf("un cierre que falla no tiene que guardar nada: sprint %q, HU-2 %q, foto %d", sprints[0].Estado, historias[1].Estado, len(foto))
		}
	})

	t.Run("sprint que no existe", func(t *testing.T) {
		e := nuevoEscenarioCierre(t)
		c := e.cierre(t)
		c.Sprint.ID = 99
		if err := e.cierres.Cerrar(ctx, c); !errors.Is(err, sprint.ErrNoEncontrado) {
			t.Errorf("se esperaba sprint.ErrNoEncontrado y se obtuvo %v", err)
		}
	})
}

// HU-06 RN6 con HU-14: una historia es de un sprint cerrado si su sprint está cerrado o si aparece en su foto
func TestHistorias_EnSprintCerrado(t *testing.T) {
	ctx := context.Background()
	e := nuevoEscenarioCierre(t)

	for _, h := range e.todas {
		if cerrado, _ := e.historias.EnSprintCerrado(ctx, h.ID); cerrado {
			t.Fatalf("antes de cerrar ninguna historia es de un sprint cerrado y HU-%d dio que sí", h.Numero)
		}
	}
	if err := e.cierres.Cerrar(ctx, e.cierre(t)); err != nil {
		t.Fatalf("cerrar: %v", err)
	}
	for i, esperado := range []bool{true, true, true, false, false} {
		cerrado, err := e.historias.EnSprintCerrado(ctx, e.todas[i].ID)
		if err != nil || cerrado != esperado {
			t.Errorf("HU-%d: EnSprintCerrado = %v (%v), se esperaba %v", i+1, cerrado, err, esperado)
		}
	}
	// sin el repositorio de cierres (por ejemplo en tests viejos) sigue diciendo que no
	solo := memory.NuevoHistorias()
	h, _ := solo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 1, Estado: backlog.EstadoPendiente})
	if cerrado, err := solo.EnSprintCerrado(ctx, h.ID); err != nil || cerrado {
		t.Errorf("sin cierres se esperaba false y se obtuvo %v, %v", cerrado, err)
	}
}
