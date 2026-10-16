package memory_test

import (
	"context"
	"errors"
	"testing"

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

// HU-06 · ActualizarDatos cambia título, descripción y prioridad de esa historia, y deja el estado y los story points
func TestHistorias_ActualizarDatos(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoHistorias()
	puntos := 5
	guardada, _ := repo.Guardar(ctx, backlog.Historia{
		ProyectoID: 1, Numero: 1, Titulo: "Login", Descripcion: "Vieja", Prioridad: backlog.PrioridadAlta,
		Estado: backlog.EstadoEnProgreso, StoryPoints: &puntos,
	})
	otra, _ := repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 2, Titulo: "Otra", Prioridad: backlog.PrioridadMedia})

	err := repo.ActualizarDatos(ctx, backlog.Historia{
		ID: guardada.ID, Titulo: "Login con Google", Descripcion: "Nueva", Prioridad: backlog.PrioridadBaja,
	})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	h, _ := repo.BuscarPorNumero(ctx, 1, 1)
	if h.Titulo != "Login con Google" || h.Descripcion != "Nueva" || h.Prioridad != backlog.PrioridadBaja {
		t.Errorf("la historia quedó con %q / %q / %q", h.Titulo, h.Descripcion, h.Prioridad)
	}
	if h.Numero != 1 || h.Estado != backlog.EstadoEnProgreso || h.StoryPoints == nil || *h.StoryPoints != 5 {
		t.Errorf("cambió lo que no se editaba: HU-%d, estado %q, story points %v", h.Numero, h.Estado, h.StoryPoints)
	}
	o, _ := repo.BuscarPorNumero(ctx, 1, 2)
	if o.Titulo != otra.Titulo || o.Prioridad != backlog.PrioridadMedia {
		t.Errorf("la otra historia cambió: %+v", o)
	}
}

func TestHistorias_ActualizarDatos_HistoriaInexistente(t *testing.T) {
	err := memory.NuevoHistorias().ActualizarDatos(context.Background(), backlog.Historia{ID: 99, Titulo: "X"})

	if !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("se esperaba ErrHistoriaNoEncontrada y se obtuvo: %v", err)
	}
}

// HU-06 · Eliminar borra esa historia y las demás conservan su número (RN11)
func TestHistorias_Eliminar_BorraLaHistoriaYLasDemasConservanSuNumero(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoHistorias()
	a, _ := repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 1, Titulo: "Primera"})
	b, _ := repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 2, Titulo: "Segunda"})
	c, _ := repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 3, Titulo: "Tercera"})
	deOtroProyecto, _ := repo.Guardar(ctx, backlog.Historia{ProyectoID: 2, Numero: 2, Titulo: "De otro proyecto"})

	if err := repo.Eliminar(ctx, b.ID); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	lista, _ := repo.ListarPorProyecto(ctx, 1)
	if len(lista) != 2 || lista[0].ID != a.ID || lista[0].Numero != 1 || lista[1].ID != c.ID || lista[1].Numero != 3 {
		t.Errorf("se esperaban HU-1 y HU-3, y se obtuvo %+v", lista)
	}
	if _, err := repo.BuscarPorNumero(ctx, 1, 2); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("HU-2 tenía que no existir más y se obtuvo: %v", err)
	}
	if _, err := repo.BuscarPorNumero(ctx, 2, 2); err != nil {
		t.Errorf("la HU-2 del otro proyecto (ID %d) tenía que seguir: %v", deOtroProyecto.ID, err)
	}
}

func TestHistorias_Eliminar_HistoriaInexistente(t *testing.T) {
	err := memory.NuevoHistorias().Eliminar(context.Background(), 99)

	if !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("se esperaba ErrHistoriaNoEncontrada y se obtuvo: %v", err)
	}
}

// HU-06 · en memoria no hay forma de asignar una historia a un sprint (HU-12), así que nunca es de un sprint cerrado
func TestHistorias_EnSprintCerrado_EnMemoriaEsFalso(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoHistorias()
	h, _ := repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 1, Titulo: "Primera"})

	cerrado, err := repo.EnSprintCerrado(ctx, h.ID)

	if err != nil || cerrado {
		t.Errorf("EnSprintCerrado = %v, %v; se esperaba false sin error", cerrado, err)
	}
}
