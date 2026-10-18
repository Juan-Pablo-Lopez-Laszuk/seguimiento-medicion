package memory_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

func TestHistorias_GuardarAsignaIDsYListarPorProyecto(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoHistorias()

	a, err := repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 1, Titulo: "Primera"})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	_, _ = repo.Guardar(ctx, backlog.Historia{ProyectoID: 2, Numero: 1, Titulo: "De otro proyecto"})
	c, _ := repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 2, Titulo: "Segunda"})

	if a.ID != 1 || c.ID != 3 {
		t.Errorf("IDs: se esperaba 1 y 3 y se obtuvo %d y %d", a.ID, c.ID)
	}
	lista, err := repo.ListarPorProyecto(ctx, 1)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if len(lista) != 2 || lista[0].Titulo != "Primera" || lista[1].Titulo != "Segunda" {
		t.Errorf("se esperaban solo las historias del proyecto 1, en orden, y se obtuvo %+v", lista)
	}
}

func TestHistorias_ProyectoSinHistorias_DevuelveListaVacia(t *testing.T) {
	lista, err := memory.NuevoHistorias().ListarPorProyecto(context.Background(), 1)
	if err != nil || len(lista) != 0 {
		t.Errorf("se esperaba una lista vacía sin error y se obtuvo %+v, %v", lista, err)
	}
}

// Los story points son un puntero: lo que guarda el repositorio no tiene que cambiar si quien lo usa modifica
// el valor que recibió o que le devolvieron.
func TestHistorias_LosStoryPointsGuardadosNoSeComparten(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoHistorias()
	puntos := 5
	guardada, _ := repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 1, StoryPoints: &puntos})

	puntos = 8
	*guardada.StoryPoints = 13
	lista, _ := repo.ListarPorProyecto(ctx, 1)
	*lista[0].StoryPoints = 21
	otraVez, _ := repo.ListarPorProyecto(ctx, 1)

	if otraVez[0].StoryPoints == nil || *otraVez[0].StoryPoints != 5 {
		t.Errorf("se esperaban 5 story points guardados y se obtuvo %v", otraVez[0].StoryPoints)
	}
}

// HU-07: la página de una historia se pide por su número dentro del proyecto (HU-N)
func TestHistorias_BuscarPorNumero(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoHistorias()
	_, _ = repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 1, Titulo: "HU-1 del proyecto 1"})
	_, _ = repo.Guardar(ctx, backlog.Historia{ProyectoID: 2, Numero: 1, Titulo: "HU-1 del proyecto 2"})

	h, err := repo.BuscarPorNumero(ctx, 2, 1)
	if err != nil || h.Titulo != "HU-1 del proyecto 2" || h.ID != 2 {
		t.Errorf("se esperaba la HU-1 del proyecto 2 y se obtuvo %+v, %v", h, err)
	}
	if _, err := repo.BuscarPorNumero(ctx, 1, 2); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("número que no existe: se esperaba backlog.ErrHistoriaNoEncontrada y se obtuvo %v", err)
	}
	if _, err := repo.BuscarPorNumero(ctx, 3, 1); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("proyecto sin historias: se esperaba backlog.ErrHistoriaNoEncontrada y se obtuvo %v", err)
	}
}

// HU-08: el estado de una historia se guarda sin tocar los demás datos
func TestHistorias_ActualizarEstado(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoHistorias()
	puntos := 5
	guardada, _ := repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 1, Titulo: "Primera", Estado: backlog.EstadoPendiente, StoryPoints: &puntos})
	otra, _ := repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 2, Titulo: "Segunda", Estado: backlog.EstadoPendiente})

	if err := repo.ActualizarEstado(ctx, guardada.ID, backlog.EstadoEnSprint); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	h, _ := repo.BuscarPorNumero(ctx, 1, 1)
	if h.Estado != backlog.EstadoEnSprint || h.Titulo != "Primera" || h.StoryPoints == nil || *h.StoryPoints != 5 {
		t.Errorf("se esperaba la HU-1 En sprint con el resto de sus datos y se obtuvo %+v", h)
	}
	if o, _ := repo.BuscarPorNumero(ctx, 1, 2); o.ID != otra.ID || o.Estado != backlog.EstadoPendiente {
		t.Errorf("la HU-2 no tenía que cambiar y se obtuvo %+v", o)
	}
	if err := repo.ActualizarEstado(ctx, 99, backlog.EstadoEnSprint); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("historia que no existe: se esperaba backlog.ErrHistoriaNoEncontrada y se obtuvo %v", err)
	}
}

// HU-12: al asignar o quitar, el estado y el sprint se guardan juntos, sin tocar los demás datos
func TestHistorias_ActualizarSprint(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoHistorias()
	puntos := 5
	guardada, _ := repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 1, Titulo: "Primera", Estado: backlog.EstadoPendiente, StoryPoints: &puntos})
	sprintID := int64(7)

	if err := repo.ActualizarSprint(ctx, guardada.ID, backlog.EstadoEnSprint, &sprintID); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	sprintID = 8 // el repositorio no comparte el valor con quien lo llamó
	h, _ := repo.BuscarPorNumero(ctx, 1, 1)
	if h.Estado != backlog.EstadoEnSprint || h.SprintID == nil || *h.SprintID != 7 || h.Titulo != "Primera" {
		t.Errorf("se esperaba la HU-1 En sprint en el sprint 7 con el resto de sus datos y se obtuvo %+v", h)
	}
	*h.SprintID = 9
	if lista, _ := repo.ListarPorProyecto(ctx, 1); *lista[0].SprintID != 7 {
		t.Errorf("el sprint guardado cambió a %d desde afuera del repositorio", *lista[0].SprintID)
	}

	if err := repo.ActualizarSprint(ctx, guardada.ID, backlog.EstadoPendiente, nil); err != nil {
		t.Fatalf("no se esperaba error al quitar: %v", err)
	}
	if h, _ := repo.BuscarPorNumero(ctx, 1, 1); h.Estado != backlog.EstadoPendiente || h.SprintID != nil {
		t.Errorf("se esperaba la HU-1 Pendiente y sin sprint y se obtuvo %+v", h)
	}
	if err := repo.ActualizarSprint(ctx, 99, backlog.EstadoEnSprint, &sprintID); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("historia que no existe: se esperaba backlog.ErrHistoriaNoEncontrada y se obtuvo %v", err)
	}
}

// HU-13: al completar, el estado y la fecha de completado se guardan juntos, sin tocar los demás datos (ni el sprint)
func TestHistorias_ActualizarCompletada(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoHistorias()
	sprintID := int64(7)
	guardada, _ := repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 1, Titulo: "Primera", Estado: backlog.EstadoEnProgreso, SprintID: &sprintID})
	fecha := time.Date(2026, 10, 15, 0, 0, 0, 0, time.UTC)

	if err := repo.ActualizarCompletada(ctx, guardada.ID, backlog.EstadoHecho, &fecha); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	fecha = fecha.AddDate(0, 0, 1) // el repositorio no comparte el valor con quien lo llamó
	h, _ := repo.BuscarPorNumero(ctx, 1, 1)
	if h.Estado != backlog.EstadoHecho || h.CompletadaEn == nil || h.CompletadaEn.Day() != 15 || h.SprintID == nil || *h.SprintID != 7 {
		t.Errorf("se esperaba la HU-1 Hecha el 15/10 en el sprint 7 y se obtuvo %+v", h)
	}
	*h.CompletadaEn = h.CompletadaEn.AddDate(0, 0, 3)
	if lista, _ := repo.ListarPorProyecto(ctx, 1); lista[0].CompletadaEn.Day() != 15 {
		t.Errorf("la fecha guardada cambió a %v desde afuera del repositorio", lista[0].CompletadaEn)
	}
	if err := repo.ActualizarCompletada(ctx, 99, backlog.EstadoHecho, &fecha); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("historia que no existe: se esperaba backlog.ErrHistoriaNoEncontrada y se obtuvo %v", err)
	}
}

// HU-16: los story points de una historia se guardan sin tocar los demás datos
func TestHistorias_ActualizarStoryPoints(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoHistorias()
	guardada, _ := repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 1, Titulo: "Primera", Estado: backlog.EstadoEnSprint})
	otra, _ := repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 2, Titulo: "Segunda", Estado: backlog.EstadoPendiente})

	if err := repo.ActualizarStoryPoints(ctx, guardada.ID, 8); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if err := repo.ActualizarStoryPoints(ctx, guardada.ID, 0); err != nil {
		t.Fatalf("no se esperaba error al estimar en 0: %v", err)
	}
	if err := repo.ActualizarStoryPoints(ctx, guardada.ID, 13); err != nil {
		t.Fatalf("no se esperaba error al re-estimar: %v", err)
	}

	h, _ := repo.BuscarPorNumero(ctx, 1, 1)
	if h.StoryPoints == nil || *h.StoryPoints != 13 || h.Titulo != "Primera" || h.Estado != backlog.EstadoEnSprint {
		t.Errorf("se esperaba la HU-1 con 13 story points y el resto de sus datos, y se obtuvo %+v", h)
	}
	if o, _ := repo.BuscarPorNumero(ctx, 1, 2); o.ID != otra.ID || o.StoryPoints != nil {
		t.Errorf("la HU-2 no tenía que cambiar y se obtuvo %+v", o)
	}
	// lo que devuelve el repositorio no se comparte con lo guardado
	*h.StoryPoints = 21
	if otraVez, _ := repo.BuscarPorNumero(ctx, 1, 1); *otraVez.StoryPoints != 13 {
		t.Errorf("modificar lo devuelto no tenía que cambiar lo guardado: %d", *otraVez.StoryPoints)
	}
	if err := repo.ActualizarStoryPoints(ctx, 99, 5); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("historia que no existe: se esperaba backlog.ErrHistoriaNoEncontrada y se obtuvo %v", err)
	}
}
