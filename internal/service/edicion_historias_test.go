package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// historiasDeSprintCerrado es el repositorio de historias en memoria con unas historias marcadas como "de un sprint
// cerrado": en memoria todavía no se pueden asignar a un sprint (HU-12), así que la prueba lo simula.
type historiasDeSprintCerrado struct {
	*memory.Historias
	cerradas map[int64]bool
}

func (r historiasDeSprintCerrado) EnSprintCerrado(_ context.Context, historiaID int64) (bool, error) {
	return r.cerradas[historiaID], nil
}

// escenarioEdicionDeHistoria arma los repositorios en memoria con el proyecto de ejemplo y su HU-1 (Pendiente) ya creados.
type escenarioEdicionDeHistoria struct {
	casos      *service.EdicionHistorias
	historias  *memory.Historias
	criterios  *memory.Criterios
	tareas     *memory.Tareas
	cerradas   map[int64]bool
	proyectoID int64
	historia   backlog.Historia
}

func nuevoEscenarioEdicionDeHistoria(t *testing.T) *escenarioEdicionDeHistoria {
	t.Helper()
	proyectos := memory.NuevoProyectos()
	p, err := service.NuevoProyectos(proyectos, relojFijo).Crear(context.Background(), datosValidos())
	if err != nil {
		t.Fatalf("no se pudo crear el proyecto de ejemplo: %v", err)
	}
	e := &escenarioEdicionDeHistoria{cerradas: map[int64]bool{}, proyectoID: p.ID}
	e.criterios, e.tareas = memory.NuevoCriterios(), memory.NuevoTareas()
	e.historias = memory.NuevoHistorias(e.criterios, e.tareas)
	e.historia, err = service.NuevoHistorias(proyectos, e.historias).Crear(context.Background(), p.ID, datosDeHistoria("Iniciar sesión con Google"))
	if err != nil {
		t.Fatalf("no se pudo crear la historia de ejemplo: %v", err)
	}
	repo := historiasDeSprintCerrado{Historias: e.historias, cerradas: e.cerradas}
	e.casos = service.NuevoEdicionHistorias(proyectos, repo, e.criterios, e.tareas)
	return e
}

// conCriteriosYTareas guarda criterios y tareas en la historia, directo en los repositorios.
func (e *escenarioEdicionDeHistoria) conCriteriosYTareas(t *testing.T, historiaID int64, criterios, tareas int) {
	t.Helper()
	for i := 0; i < criterios; i++ {
		if _, err := e.criterios.Guardar(context.Background(), backlog.Criterio{HistoriaID: historiaID, Descripcion: "Criterio"}); err != nil {
			t.Fatalf("no se pudo guardar un criterio: %v", err)
		}
	}
	for i := 0; i < tareas; i++ {
		if _, err := e.tareas.Guardar(context.Background(), effort.Tarea{HistoriaID: historiaID, Titulo: "Tarea"}); err != nil {
			t.Fatalf("no se pudo guardar una tarea: %v", err)
		}
	}
}

func cambiosDeHistoriaDeEjemplo() backlog.Cambios {
	return backlog.Cambios{Titulo: "Iniciar sesión con Google o GitHub", Descripcion: "Con OAuth.", Prioridad: backlog.PrioridadBaja}
}

// CA-06.1 · editar guarda los datos nuevos y la historia sigue siendo la misma (mismo ID y mismo número)
func TestEditarHistoria_Valido_GuardaLosDatosNuevos(t *testing.T) {
	e := nuevoEscenarioEdicionDeHistoria(t)

	editada, err := e.casos.Editar(context.Background(), e.proyectoID, e.historia.Numero, cambiosDeHistoriaDeEjemplo())
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if editada.ID != e.historia.ID || editada.Numero != e.historia.Numero || editada.Titulo != "Iniciar sesión con Google o GitHub" {
		t.Errorf("se esperaba la misma historia con el título nuevo y se obtuvo %+v", editada)
	}
	guardada, err := e.historias.BuscarPorNumero(context.Background(), e.proyectoID, e.historia.Numero)
	if err != nil {
		t.Fatalf("no se esperaba error al buscar la historia: %v", err)
	}
	if guardada.Titulo != "Iniciar sesión con Google o GitHub" || guardada.Descripcion != "Con OAuth." || guardada.Prioridad != backlog.PrioridadBaja {
		t.Errorf("la historia guardada quedó con %q / %q / %q", guardada.Titulo, guardada.Descripcion, guardada.Prioridad)
	}
	if guardada.Estado != e.historia.Estado || guardada.Numero != e.historia.Numero {
		t.Errorf("cambió lo que no se editaba: HU-%d en %q", guardada.Numero, guardada.Estado)
	}
}

// RN1 · con datos inválidos se devuelven los errores de campo y no se guarda nada
func TestEditarHistoria_DatosInvalidos_NoGuardaNada(t *testing.T) {
	e := nuevoEscenarioEdicionDeHistoria(t)
	c := cambiosDeHistoriaDeEjemplo()
	c.Titulo, c.Prioridad = "", "Urgente"

	_, err := e.casos.Editar(context.Background(), e.proyectoID, e.historia.Numero, c)

	var errs domain.ErroresValidacion
	if !errors.As(err, &errs) || errs[backlog.CampoTitulo] == nil || errs[backlog.CampoPrioridad] == nil {
		t.Fatalf("se esperaban errores de título y de prioridad y se obtuvo: %v", err)
	}
	guardada, _ := e.historias.BuscarPorNumero(context.Background(), e.proyectoID, e.historia.Numero)
	if guardada.Titulo != e.historia.Titulo || guardada.Prioridad != e.historia.Prioridad {
		t.Errorf("la historia cambió: %q / %q", guardada.Titulo, guardada.Prioridad)
	}
}

// RN3 · se puede editar una historia Hecha
func TestEditarHistoria_HistoriaHecha_Vale(t *testing.T) {
	e := nuevoEscenarioEdicionDeHistoria(t)
	if err := e.historias.ActualizarEstado(context.Background(), e.historia.ID, backlog.EstadoHecho); err != nil {
		t.Fatalf("no se pudo dejar la historia Hecha: %v", err)
	}

	if _, err := e.casos.Editar(context.Background(), e.proyectoID, e.historia.Numero, cambiosDeHistoriaDeEjemplo()); err != nil {
		t.Errorf("no se esperaba error: %v", err)
	}
}

// Buscar trae el proyecto y la historia para armar el formulario de editar
func TestBuscarHistoriaParaEditar_DevuelveElProyectoYLaHistoria(t *testing.T) {
	e := nuevoEscenarioEdicionDeHistoria(t)

	p, h, err := e.casos.Buscar(context.Background(), e.proyectoID, e.historia.Numero)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if p.ID != e.proyectoID || h.ID != e.historia.ID {
		t.Errorf("se esperaba el proyecto %d y la historia %d, y se obtuvo %d y %d", e.proyectoID, e.historia.ID, p.ID, h.ID)
	}
}

// Sección 7 · proyecto o historia que no existen
func TestEdicionDeHistorias_ProyectoOHistoriaInexistente_DevuelveNoEncontrado(t *testing.T) {
	e := nuevoEscenarioEdicionDeHistoria(t)
	ctx := context.Background()

	_, _, errBuscar := e.casos.Buscar(ctx, 999, e.historia.Numero)
	_, errProyecto := e.casos.Editar(ctx, 999, e.historia.Numero, cambiosDeHistoriaDeEjemplo())
	_, errHistoria := e.casos.Editar(ctx, e.proyectoID, 999, cambiosDeHistoriaDeEjemplo())

	if !errors.Is(errBuscar, project.ErrNoEncontrado) || !errors.Is(errProyecto, project.ErrNoEncontrado) {
		t.Errorf("proyecto inexistente: se esperaba project.ErrNoEncontrado y se obtuvo %v / %v", errBuscar, errProyecto)
	}
	if !errors.Is(errHistoria, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("historia inexistente: se esperaba backlog.ErrHistoriaNoEncontrada y se obtuvo %v", errHistoria)
	}
}

// RN8 · pedir la eliminación muestra qué se pierde (cuántos criterios y tareas) y no elimina nada
func TestPedirEliminacion_CuentaLosCriteriosYLasTareas_YNoEliminaNada(t *testing.T) {
	e := nuevoEscenarioEdicionDeHistoria(t)
	e.conCriteriosYTareas(t, e.historia.ID, 2, 1)

	vista, err := e.casos.PedirEliminacion(context.Background(), e.proyectoID, e.historia.Numero)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if vista.Proyecto.ID != e.proyectoID || vista.Historia.ID != e.historia.ID {
		t.Errorf("se esperaba el proyecto %d y la historia %d, y se obtuvo %d y %d", e.proyectoID, e.historia.ID, vista.Proyecto.ID, vista.Historia.ID)
	}
	if vista.Criterios != 2 || vista.Tareas != 1 {
		t.Errorf("se esperaban 2 criterios y 1 tarea y se obtuvo %d y %d", vista.Criterios, vista.Tareas)
	}
	if vista.Impedimento != nil {
		t.Errorf("no se esperaba impedimento y se obtuvo: %v", vista.Impedimento)
	}
	if _, err := e.historias.BuscarPorNumero(context.Background(), e.proyectoID, e.historia.Numero); err != nil {
		t.Errorf("pedir la eliminación no tenía que eliminar la historia: %v", err)
	}
}

// CA-06.2 · la pantalla de confirmación sabe por qué no se puede eliminar
func TestPedirEliminacion_HistoriaHecha_InformaElImpedimento(t *testing.T) {
	e := nuevoEscenarioEdicionDeHistoria(t)
	if err := e.historias.ActualizarEstado(context.Background(), e.historia.ID, backlog.EstadoHecho); err != nil {
		t.Fatalf("no se pudo dejar la historia Hecha: %v", err)
	}

	vista, err := e.casos.PedirEliminacion(context.Background(), e.proyectoID, e.historia.Numero)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if !errors.Is(vista.Impedimento, backlog.ErrEliminarHistoriaHecha) {
		t.Errorf("se esperaba ErrEliminarHistoriaHecha y se obtuvo: %v", vista.Impedimento)
	}
}

func TestPedirEliminacion_DeUnSprintCerrado_InformaElImpedimento(t *testing.T) {
	e := nuevoEscenarioEdicionDeHistoria(t)
	e.cerradas[e.historia.ID] = true

	vista, err := e.casos.PedirEliminacion(context.Background(), e.proyectoID, e.historia.Numero)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if !errors.Is(vista.Impedimento, backlog.ErrEliminarEnSprintCerrado) {
		t.Errorf("se esperaba ErrEliminarEnSprintCerrado y se obtuvo: %v", vista.Impedimento)
	}
}

// CA-06.3 y RN9 · confirmar elimina la historia con sus criterios y sus tareas, y devuelve la historia eliminada
func TestEliminarHistoria_Valido_EliminaLaHistoriaConSusCriteriosYTareas(t *testing.T) {
	e := nuevoEscenarioEdicionDeHistoria(t)
	e.conCriteriosYTareas(t, e.historia.ID, 2, 1)

	eliminada, err := e.casos.Eliminar(context.Background(), e.proyectoID, e.historia.Numero)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if eliminada.ID != e.historia.ID || eliminada.Numero != e.historia.Numero || eliminada.Titulo != e.historia.Titulo {
		t.Errorf("se esperaba la historia eliminada y se obtuvo %+v", eliminada)
	}
	if _, err := e.historias.BuscarPorNumero(context.Background(), e.proyectoID, e.historia.Numero); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("la historia tenía que no existir más y se obtuvo: %v", err)
	}
	if cs, _ := e.criterios.ListarPorHistoria(context.Background(), e.historia.ID); len(cs) != 0 {
		t.Errorf("quedaron %d criterios", len(cs))
	}
	if ts, _ := e.tareas.ListarPorHistoria(context.Background(), e.historia.ID); len(ts) != 0 {
		t.Errorf("quedaron %d tareas", len(ts))
	}
}

// RN7 · se elimina en Pendiente, En sprint y En progreso
func TestEliminarHistoria_PendienteEnSprintOEnProgreso_Vale(t *testing.T) {
	for _, estado := range []backlog.Estado{backlog.EstadoPendiente, backlog.EstadoEnSprint, backlog.EstadoEnProgreso} {
		e := nuevoEscenarioEdicionDeHistoria(t)
		if err := e.historias.ActualizarEstado(context.Background(), e.historia.ID, estado); err != nil {
			t.Fatalf("no se pudo dejar la historia en %q: %v", estado, err)
		}

		if _, err := e.casos.Eliminar(context.Background(), e.proyectoID, e.historia.Numero); err != nil {
			t.Errorf("estado %q: no se esperaba error: %v", estado, err)
		}
	}
}

// verificarQueNoSeBorroNada comprueba que la historia de ejemplo sigue con sus 2 criterios y su tarea.
func (e *escenarioEdicionDeHistoria) verificarQueNoSeBorroNada(t *testing.T) {
	t.Helper()
	if _, err := e.historias.BuscarPorNumero(context.Background(), e.proyectoID, e.historia.Numero); err != nil {
		t.Errorf("la historia tenía que seguir: %v", err)
	}
	if cs, _ := e.criterios.ListarPorHistoria(context.Background(), e.historia.ID); len(cs) != 2 {
		t.Errorf("los criterios tenían que seguir (2) y hay %d", len(cs))
	}
	if ts, _ := e.tareas.ListarPorHistoria(context.Background(), e.historia.ID); len(ts) != 1 {
		t.Errorf("la tarea tenía que seguir (1) y hay %d", len(ts))
	}
}

// CA-06.2 y RN5 · una historia Hecha no se elimina, y no se borra nada
func TestEliminarHistoria_HistoriaHecha_NoSeElimina(t *testing.T) {
	e := nuevoEscenarioEdicionDeHistoria(t)
	e.conCriteriosYTareas(t, e.historia.ID, 2, 1)
	if err := e.historias.ActualizarEstado(context.Background(), e.historia.ID, backlog.EstadoHecho); err != nil {
		t.Fatalf("no se pudo dejar la historia Hecha: %v", err)
	}

	_, err := e.casos.Eliminar(context.Background(), e.proyectoID, e.historia.Numero)

	if !errors.Is(err, backlog.ErrEliminarHistoriaHecha) {
		t.Fatalf("se esperaba ErrEliminarHistoriaHecha y se obtuvo: %v", err)
	}
	e.verificarQueNoSeBorroNada(t)
}

// CA-06.2 y RN6 · una historia de un sprint cerrado no se elimina, y no se borra nada
func TestEliminarHistoria_DeUnSprintCerrado_NoSeElimina(t *testing.T) {
	e := nuevoEscenarioEdicionDeHistoria(t)
	e.conCriteriosYTareas(t, e.historia.ID, 2, 1)
	e.cerradas[e.historia.ID] = true

	_, err := e.casos.Eliminar(context.Background(), e.proyectoID, e.historia.Numero)

	if !errors.Is(err, backlog.ErrEliminarEnSprintCerrado) {
		t.Fatalf("se esperaba ErrEliminarEnSprintCerrado y se obtuvo: %v", err)
	}
	e.verificarQueNoSeBorroNada(t)
}

// historiasQueNoSeBorran simula que la base rechaza el borrado por otros registros asociados (RN10).
type historiasQueNoSeBorran struct{ *memory.Historias }

func (historiasQueNoSeBorran) EnSprintCerrado(context.Context, int64) (bool, error) {
	return false, nil
}
func (historiasQueNoSeBorran) Eliminar(context.Context, int64) error {
	return backlog.ErrHistoriaConRegistros
}

// RN10 · si el repositorio informa que la historia tiene registros asociados, el error llega tal cual
func TestEliminarHistoria_ConRegistrosAsociados_DevuelveElError(t *testing.T) {
	proyectos := memory.NuevoProyectos()
	p, _ := service.NuevoProyectos(proyectos, relojFijo).Crear(context.Background(), datosValidos())
	historias := memory.NuevoHistorias()
	h, _ := service.NuevoHistorias(proyectos, historias).Crear(context.Background(), p.ID, datosDeHistoria("Con registros"))
	casos := service.NuevoEdicionHistorias(proyectos, historiasQueNoSeBorran{historias}, memory.NuevoCriterios(), memory.NuevoTareas())

	_, err := casos.Eliminar(context.Background(), p.ID, h.Numero)

	if !errors.Is(err, backlog.ErrHistoriaConRegistros) {
		t.Errorf("se esperaba ErrHistoriaConRegistros y se obtuvo: %v", err)
	}
}

// Sección 7 · proyecto o historia que no existen
func TestEliminarHistoria_ProyectoOHistoriaInexistente_DevuelveNoEncontrado(t *testing.T) {
	e := nuevoEscenarioEdicionDeHistoria(t)
	ctx := context.Background()

	_, errProyecto := e.casos.Eliminar(ctx, 999, e.historia.Numero)
	_, errHistoria := e.casos.Eliminar(ctx, e.proyectoID, 999)
	_, errPedirProyecto := e.casos.PedirEliminacion(ctx, 999, e.historia.Numero)
	_, errPedirHistoria := e.casos.PedirEliminacion(ctx, e.proyectoID, 999)

	for _, err := range []error{errProyecto, errPedirProyecto} {
		if !errors.Is(err, project.ErrNoEncontrado) {
			t.Errorf("proyecto inexistente: se esperaba project.ErrNoEncontrado y se obtuvo %v", err)
		}
	}
	for _, err := range []error{errHistoria, errPedirHistoria} {
		if !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
			t.Errorf("historia inexistente: se esperaba backlog.ErrHistoriaNoEncontrada y se obtuvo %v", err)
		}
	}
}
