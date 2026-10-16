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
