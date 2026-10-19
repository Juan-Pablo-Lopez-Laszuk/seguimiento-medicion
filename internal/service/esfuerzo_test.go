package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// el18deOctubre son las 15:00 de Argentina del 18/10/2026: "hoy" de la spec de HU-24.
func el18deOctubre() time.Time { return time.Date(2026, 10, 18, 18, 0, 0, 0, time.UTC) }

type escenarioEsfuerzo struct {
	casos      *service.Esfuerzo
	proyectoID int64
	hu1, hu2   backlog.Historia
	carolina   member.Integrante
	tarea      effort.Tarea
}

func nuevoEscenarioEsfuerzo(t *testing.T) escenarioEsfuerzo {
	t.Helper()
	ctx := context.Background()
	proyectos, historias, integrantes, tareas := memory.NuevoProyectos(), memory.NuevoHistorias(), memory.NuevoIntegrantes(), memory.NuevoTareas()
	p, err := service.NuevoProyectos(proyectos, relojFijo).Crear(ctx, datosValidos())
	if err != nil {
		t.Fatalf("crear el proyecto: %v", err)
	}
	e := escenarioEsfuerzo{proyectoID: p.ID}
	casosHistorias := service.NuevoHistorias(proyectos, historias)
	if e.hu1, err = casosHistorias.Crear(ctx, p.ID, datosDeHistoria("Iniciar sesión con Google")); err != nil {
		t.Fatalf("crear la HU-1: %v", err)
	}
	if e.hu2, err = casosHistorias.Crear(ctx, p.ID, datosDeHistoria("Ver el tablero")); err != nil {
		t.Fatalf("crear la HU-2: %v", err)
	}
	if e.carolina, err = service.NuevoIntegrantes(proyectos, integrantes).Registrar(ctx, p.ID, member.Datos{Nombre: "Carolina", Email: "caro@mail.com", Rol: member.RolProductBuilder}); err != nil {
		t.Fatalf("registrar a Carolina: %v", err)
	}
	if e.tarea, err = service.NuevoTareas(proyectos, historias, integrantes, tareas).Agregar(ctx, p.ID, e.hu2.Numero, effort.DatosTarea{Titulo: "Armar la tabla", ResponsableID: e.carolina.ID, Horas: "4"}); err != nil {
		t.Fatalf("agregar la tarea de la HU-2: %v", err)
	}
	e.casos = service.NuevoEsfuerzo(proyectos, historias, integrantes, tareas, memory.NuevoRegistros(), el18deOctubre)
	return e
}

func (e escenarioEsfuerzo) registrar(h backlog.Historia, fecha, horas string) error {
	_, err := e.casos.Registrar(context.Background(), e.proyectoID, h.Numero,
		effort.DatosRegistro{IntegranteID: e.carolina.ID, Fecha: fecha, Actividad: "Trabajo", Horas: horas})
	return err
}

// CA-24.1 y M5 · lo registrado queda en la historia y suma sus horas reales; hoy es el de Argentina
func TestRegistrarEsfuerzo_GuardaYSumaLasHorasReales(t *testing.T) {
	e := nuevoEscenarioEsfuerzo(t)
	if err := e.registrar(e.hu1, "2026-10-18", "3"); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if err := e.registrar(e.hu1, "2026-10-17", "1,5"); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	v, err := e.casos.Ver(context.Background(), e.proyectoID, e.hu1.Numero)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if len(v.Registros) != 2 || v.HorasReales != 4.5 || v.NombreDe(e.carolina.ID) != "Carolina" || len(v.Activos()) != 1 {
		t.Errorf("se esperaban 2 registros, 4,5 h y Carolina para elegir, y se obtuvo %+v", v)
	}
	if !v.Hoy.Equal(time.Date(2026, 10, 18, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("hoy = %v, se esperaba el 18/10/2026", v.Hoy)
	}
}

// CA-24.1 (RN2) · el total del día cuenta los registros del integrante en cualquier historia
func TestRegistrarEsfuerzo_TotalDelDiaEntreHistorias(t *testing.T) {
	e := nuevoEscenarioEsfuerzo(t)
	if err := e.registrar(e.hu2, "2026-10-17", "20"); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	err := e.registrar(e.hu1, "2026-10-17", "5")

	if got := errorDeCampo(t, err, effort.CampoHoras); !errors.Is(got, effort.ErrMasDe24HorasEnElDia) {
		t.Errorf("se esperaba %v y se obtuvo %v", effort.ErrMasDe24HorasEnElDia, got)
	}
}

// RN5 · la tarea tiene que ser de la historia: una de la HU-2 no se usa en la HU-1
func TestRegistrarEsfuerzo_TareaDeOtraHistoria(t *testing.T) {
	e := nuevoEscenarioEsfuerzo(t)
	d := effort.DatosRegistro{IntegranteID: e.carolina.ID, TareaID: e.tarea.ID, Fecha: "2026-10-18", Actividad: "Trabajo", Horas: "1"}

	_, err := e.casos.Registrar(context.Background(), e.proyectoID, e.hu1.Numero, d)
	if got := errorDeCampo(t, err, effort.CampoTarea); !errors.Is(got, effort.ErrTareaDeOtraHistoria) {
		t.Errorf("se esperaba %v y se obtuvo %v", effort.ErrTareaDeOtraHistoria, got)
	}
	if _, err := e.casos.Registrar(context.Background(), e.proyectoID, e.hu2.Numero, d); err != nil {
		t.Errorf("en su historia no se esperaba error: %v", err)
	}
}

func TestEsfuerzo_ProyectoOHistoriaInexistente(t *testing.T) {
	e := nuevoEscenarioEsfuerzo(t)
	if _, err := e.casos.Ver(context.Background(), 99, 1); !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("proyecto: se esperaba %v y se obtuvo %v", project.ErrNoEncontrado, err)
	}
	if err := e.registrar(backlog.Historia{Numero: 99}, "2026-10-18", "1"); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("historia: se esperaba %v y se obtuvo %v", backlog.ErrHistoriaNoEncontrada, err)
	}
}
