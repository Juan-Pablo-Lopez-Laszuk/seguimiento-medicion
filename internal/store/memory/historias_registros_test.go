package memory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/defect"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// HU-06, RN10 · los registros de esfuerzo y los defectos guardan a qué historia pertenecen; los repositorios lo informan
func TestRegistros_TieneDeHistoria(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoRegistros()
	_, _ = repo.Guardar(ctx, effort.Registro{IntegranteID: 1, HistoriaID: 7, Horas: 2})

	if !repo.TieneDeHistoria(7) {
		t.Error("la historia 7 tiene un registro de esfuerzo")
	}
	for _, otra := range []int64{0, 1, 8} {
		if repo.TieneDeHistoria(otra) {
			t.Errorf("la historia %d no tiene registros de esfuerzo", otra)
		}
	}
}

func TestDefectos_TieneDeHistoria(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoDefectos()
	_, _ = repo.Guardar(ctx, defect.Defecto{ProyectoID: 1, HistoriaID: 7})
	_, _ = repo.Guardar(ctx, defect.Defecto{ProyectoID: 1}) // un defecto sin historia no es de ninguna

	if !repo.TieneDeHistoria(7) {
		t.Error("la historia 7 tiene un defecto")
	}
	for _, otra := range []int64{0, 1, 8} {
		if repo.TieneDeHistoria(otra) {
			t.Errorf("la historia %d no tiene defectos", otra)
		}
	}
}

// historiaConTodo arma un repositorio de historias en memoria conectado a los demás, con una historia a eliminar (con un
// criterio y una tarea) y otra que se queda.
func historiaConTodo(t *testing.T) (repo *memory.Historias, criterios *memory.Criterios, tareas *memory.Tareas, registros *memory.Registros, defectos *memory.Defectos, aEliminar, queSeQueda backlog.Historia) {
	t.Helper()
	ctx := context.Background()
	criterios, tareas, registros, defectos = memory.NuevoCriterios(), memory.NuevoTareas(), memory.NuevoRegistros(), memory.NuevoDefectos()
	repo = memory.NuevoHistorias(criterios, tareas).ConRegistrosAsociados(registros, defectos)
	aEliminar, _ = repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 1, Titulo: "A eliminar"})
	queSeQueda, _ = repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 2, Titulo: "Se queda"})
	_, _ = criterios.Guardar(ctx, backlog.Criterio{HistoriaID: aEliminar.ID, Descripcion: "Un criterio"})
	_, _ = tareas.Guardar(ctx, effort.Tarea{HistoriaID: aEliminar.ID, Titulo: "Una tarea"})
	return repo, criterios, tareas, registros, defectos, aEliminar, queSeQueda
}

// verificarNadaSeBorro comprueba que la historia, su criterio y su tarea siguen estando después de un borrado rechazado.
func verificarNadaSeBorro(t *testing.T, repo *memory.Historias, criterios *memory.Criterios, tareas *memory.Tareas, h backlog.Historia) {
	t.Helper()
	ctx := context.Background()
	if lista, _ := repo.ListarPorProyecto(ctx, h.ProyectoID); len(lista) != 2 {
		t.Errorf("la historia tenía que seguir en el repositorio y hay %d historias", len(lista))
	}
	if cs, _ := criterios.ListarPorHistoria(ctx, h.ID); len(cs) != 1 {
		t.Errorf("un borrado rechazado no borra nada: quedaron %d criterios y se esperaba 1", len(cs))
	}
	if ts, _ := tareas.ListarPorHistoria(ctx, h.ID); len(ts) != 1 {
		t.Errorf("un borrado rechazado no borra nada: quedaron %d tareas y se esperaba 1", len(ts))
	}
}

// HU-06, RN10 · una historia con horas registradas no se elimina, igual que con PostgreSQL: no se borra nada
func TestHistorias_Eliminar_ConEsfuerzoRegistrado_DevuelveErrHistoriaConRegistrosYNoBorraNada(t *testing.T) {
	ctx := context.Background()
	repo, criterios, tareas, registros, _, aEliminar, _ := historiaConTodo(t)
	_, _ = registros.Guardar(ctx, effort.Registro{IntegranteID: 1, HistoriaID: aEliminar.ID, Horas: 2})

	err := repo.Eliminar(ctx, aEliminar.ID)

	if !errors.Is(err, backlog.ErrHistoriaConRegistros) {
		t.Fatalf("se esperaba backlog.ErrHistoriaConRegistros y se obtuvo %v", err)
	}
	verificarNadaSeBorro(t, repo, criterios, tareas, aEliminar)
}

// HU-06, RN10 · lo mismo con un defecto de la historia
func TestHistorias_Eliminar_ConUnDefecto_DevuelveErrHistoriaConRegistrosYNoBorraNada(t *testing.T) {
	ctx := context.Background()
	repo, criterios, tareas, _, defectos, aEliminar, _ := historiaConTodo(t)
	_, _ = defectos.Guardar(ctx, defect.Defecto{ProyectoID: 1, HistoriaID: aEliminar.ID})

	err := repo.Eliminar(ctx, aEliminar.ID)

	if !errors.Is(err, backlog.ErrHistoriaConRegistros) {
		t.Fatalf("se esperaba backlog.ErrHistoriaConRegistros y se obtuvo %v", err)
	}
	verificarNadaSeBorro(t, repo, criterios, tareas, aEliminar)
}

// Los registros y defectos de otras historias, y los defectos sin historia, no impiden eliminar
func TestHistorias_Eliminar_ConRegistrosDeOtraHistoria_Elimina(t *testing.T) {
	ctx := context.Background()
	repo, _, _, registros, defectos, aEliminar, queSeQueda := historiaConTodo(t)
	_, _ = registros.Guardar(ctx, effort.Registro{IntegranteID: 1, HistoriaID: queSeQueda.ID, Horas: 2})
	_, _ = defectos.Guardar(ctx, defect.Defecto{ProyectoID: 1, HistoriaID: queSeQueda.ID})
	_, _ = defectos.Guardar(ctx, defect.Defecto{ProyectoID: 1})

	if err := repo.Eliminar(ctx, aEliminar.ID); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if lista, _ := repo.ListarPorProyecto(ctx, 1); len(lista) != 1 || lista[0].ID != queSeQueda.ID {
		t.Errorf("tenía que quedar solo la otra historia y quedó %+v", lista)
	}
}
