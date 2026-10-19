package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/defect"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

type escenarioDefectos struct {
	casos      *service.Defectos
	proyectoID int64
	sprint     sprint.Sprint
	historia   backlog.Historia
}

func nuevoEscenarioDefectos(t *testing.T) escenarioDefectos {
	t.Helper()
	ctx := context.Background()
	proyectos, sprints, historias := memory.NuevoProyectos(), memory.NuevoSprints(), memory.NuevoHistorias()
	p, err := service.NuevoProyectos(proyectos, relojFijo).Crear(ctx, datosValidos())
	if err != nil {
		t.Fatalf("crear el proyecto: %v", err)
	}
	e := escenarioDefectos{proyectoID: p.ID}
	if e.sprint, err = service.NuevoSprints(proyectos, sprints).Crear(ctx, p.ID, datosDeSprint("2026-10-05", "2026-10-11")); err != nil {
		t.Fatalf("crear el sprint: %v", err)
	}
	if e.historia, err = service.NuevoHistorias(proyectos, historias).Crear(ctx, p.ID, datosDeHistoria("Iniciar sesión con Google")); err != nil {
		t.Fatalf("crear la historia: %v", err)
	}
	e.casos = service.NuevoDefectos(proyectos, sprints, historias, memory.NuevoDefectos())
	return e
}

// CA-27.1 a CA-27.3 · el defecto se guarda Abierto y la vista trae los defectos, sprints e historias del proyecto
func TestRegistrarDefecto_GuardaAbiertoYSeVeEnElProyecto(t *testing.T) {
	e := nuevoEscenarioDefectos(t)
	d := defect.Datos{Descripcion: "El login falla", Severidad: defect.SeveridadAlta, SprintID: e.sprint.ID, HistoriaID: e.historia.ID}

	guardado, err := e.casos.Registrar(context.Background(), e.proyectoID, d)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	v, err := e.casos.Ver(context.Background(), e.proyectoID)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if guardado.ID == 0 || guardado.Estado != defect.EstadoAbierto {
		t.Errorf("se esperaba un defecto con ID y Abierto, y se obtuvo %+v", guardado)
	}
	if len(v.Defectos) != 1 || len(v.Sprints) != 1 || len(v.Historias) != 1 {
		t.Errorf("la vista = %+v", v)
	}
	if v.SprintDe(e.sprint.ID) != "Sprint 1" || v.HistoriaDe(e.historia.ID) != "HU-1 · Iniciar sesión con Google" || v.HistoriaDe(0) != "" {
		t.Errorf("nombres: %q, %q", v.SprintDe(e.sprint.ID), v.HistoriaDe(e.historia.ID))
	}
}

// CA-27.3 · un sprint de otro proyecto se rechaza y no se guarda nada
func TestRegistrarDefecto_SprintDeOtroProyecto(t *testing.T) {
	e := nuevoEscenarioDefectos(t)
	_, err := e.casos.Registrar(context.Background(), e.proyectoID, defect.Datos{Descripcion: "x", Severidad: defect.SeveridadAlta, SprintID: e.sprint.ID + 99})

	if got := errorDeCampo(t, err, defect.CampoSprint); !errors.Is(got, defect.ErrSprintInvalido) {
		t.Errorf("se esperaba %v y se obtuvo %v", defect.ErrSprintInvalido, got)
	}
	if v, _ := e.casos.Ver(context.Background(), e.proyectoID); len(v.Defectos) != 0 {
		t.Error("no se debería haber guardado el defecto")
	}
}

func TestDefectos_ProyectoInexistente(t *testing.T) {
	e := nuevoEscenarioDefectos(t)
	if _, err := e.casos.Ver(context.Background(), 99); !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("ver: se esperaba %v y se obtuvo %v", project.ErrNoEncontrado, err)
	}
	if _, err := e.casos.Registrar(context.Background(), 99, defect.Datos{}); !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("registrar: se esperaba %v y se obtuvo %v", project.ErrNoEncontrado, err)
	}
}

// HU-28 · el estado se cambia y se guarda; la vista ofrece los estados siguientes de cada defecto
func TestCambiarEstadoDeDefecto(t *testing.T) {
	e := nuevoEscenarioDefectos(t)
	ctx := context.Background()
	d, err := e.casos.Registrar(ctx, e.proyectoID, defect.Datos{Descripcion: "Falla", Severidad: defect.SeveridadAlta, SprintID: e.sprint.ID})
	if err != nil {
		t.Fatalf("registrar: %v", err)
	}

	if _, err := e.casos.CambiarEstado(ctx, e.proyectoID, d.ID, defect.EstadoEnProgreso, 0); err != nil {
		t.Fatalf("a En progreso: %v", err)
	}
	resuelto, err := e.casos.CambiarEstado(ctx, e.proyectoID, d.ID, defect.EstadoResuelto, e.sprint.ID)
	if err != nil || resuelto.SprintResolucionID != e.sprint.ID {
		t.Fatalf("a Resuelto: %+v, %v", resuelto, err)
	}
	v, _ := e.casos.Ver(ctx, e.proyectoID)
	if v.Defectos[0].Estado != defect.EstadoResuelto || len(v.Siguientes(v.Defectos[0].Estado)) != 2 {
		t.Errorf("la vista = %+v", v.Defectos)
	}
	if _, err := e.casos.CambiarEstado(ctx, e.proyectoID, d.ID, defect.EstadoEnProgreso, 0); !errors.Is(err, defect.ErrTransicionInvalida) {
		t.Errorf("Resuelto → En progreso: se esperaba %v y se obtuvo %v", defect.ErrTransicionInvalida, err)
	}
	if _, err := e.casos.CambiarEstado(ctx, e.proyectoID, d.ID+99, defect.EstadoCerrado, 0); !errors.Is(err, defect.ErrDefectoNoEncontrado) {
		t.Errorf("inexistente: se esperaba %v y se obtuvo %v", defect.ErrDefectoNoEncontrado, err)
	}
}
