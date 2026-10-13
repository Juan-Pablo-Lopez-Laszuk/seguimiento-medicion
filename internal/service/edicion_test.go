package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// escenarioEdicion es el escenario de la spec de HU-02: "Software Metrics" (del 05/10 al 01/11) con el Sprint 1 del
// 05/10 al 11/10 y el Sprint 2 del 12/10 al 18/10, y además "Otro proyecto". Reutiliza escenarioEstado (HU-04).
func escenarioEdicion(t *testing.T) *escenarioEstado {
	t.Helper()
	e := nuevoEscenarioEstado(t)
	e.sprint(1, "2026-10-05", "2026-10-11", sprint.EstadoCerrado)
	e.sprint(2, "2026-10-12", "2026-10-18", sprint.EstadoPlanificado)
	otro := datosValidos()
	otro.Nombre = "Otro proyecto"
	if _, err := service.NuevoProyectos(e.proyectos, relojFijo).Crear(context.Background(), otro); err != nil {
		t.Fatalf("no se pudo crear el segundo proyecto: %v", err)
	}
	return e
}

// edicion arma el caso de uso con "hoy" a las 15:00 de Argentina de ese día.
func (e *escenarioEstado) edicion(hoy string) *service.EdicionProyectos {
	ahora := fechaISO(e.t, hoy).Add(18 * time.Hour)
	return service.NuevoEdicionProyectos(e.proyectos, e.sprints, func() time.Time { return ahora })
}

// datosDeEdicion devuelve los datos actuales del proyecto de ejemplo con el nombre y las fechas indicadas.
func datosDeEdicion(t *testing.T, nombre, inicio, fin string) project.Datos {
	return project.Datos{Nombre: nombre, Descripcion: "TPI de ICS", FechaInicio: fechaISO(t, inicio), FechaFin: fechaISO(t, fin)}
}

// guardado devuelve el proyecto de ejemplo tal como quedó en el repositorio.
func (e *escenarioEstado) guardado() project.Proyecto {
	e.t.Helper()
	p, err := e.proyectos.BuscarPorID(context.Background(), e.proyectoID)
	if err != nil {
		e.t.Fatalf("no se pudo buscar el proyecto: %v", err)
	}
	return p
}

// CA-02.1 · los datos nuevos quedan guardados
func TestModificarProyecto_DatosValidos_QuedaGuardado(t *testing.T) {
	e := escenarioEdicion(t)
	d := datosDeEdicion(t, "Software Metrics", "2026-10-05", "2026-11-01")
	d.Descripcion = "TPI de ICS 2026"

	p, err := e.edicion("2026-10-13").Modificar(context.Background(), e.proyectoID, d)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if p.Descripcion != "TPI de ICS 2026" || e.guardado().Descripcion != "TPI de ICS 2026" {
		t.Errorf("se esperaba la descripción nueva guardada y se obtuvo %+v", e.guardado())
	}
}

// RN2 · el propio nombre con otras mayúsculas no es un nombre repetido
func TestModificarProyecto_PropioNombreConOtrasMayusculas(t *testing.T) {
	e := escenarioEdicion(t)
	_, err := e.edicion("2026-10-13").Modificar(context.Background(), e.proyectoID, datosDeEdicion(t, "software METRICS", "2026-10-05", "2026-11-01"))
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if e.guardado().Nombre != "software METRICS" {
		t.Errorf("se esperaba el nombre \"software METRICS\" y se obtuvo %q", e.guardado().Nombre)
	}
}

// CA-02.1 y RN2 · el nombre de otro proyecto se informa junto con los demás errores y no se guarda nada
func TestModificarProyecto_NombreDeOtroYFechasInvalidas(t *testing.T) {
	e := escenarioEdicion(t)
	_, err := e.edicion("2026-10-13").Modificar(context.Background(), e.proyectoID, datosDeEdicion(t, "OTRO PROYECTO", "2026-10-05", "2026-10-05"))

	if got := errorDeCampo(t, err, project.CampoNombre); !errors.Is(got, project.ErrNombreRepetido) {
		t.Errorf("nombre: se esperaba ErrNombreRepetido y se obtuvo %v", got)
	}
	if got := errorDeCampo(t, err, project.CampoFechaFin); !errors.Is(got, project.ErrFechasInvalidas) {
		t.Errorf("fecha_fin: se esperaba ErrFechasInvalidas y se obtuvo %v", got)
	}
	if e.guardado().Nombre != "Software Metrics" {
		t.Errorf("no tenía que guardarse nada y el nombre quedó %q", e.guardado().Nombre)
	}
}

// RN1 · con el nombre inválido no se busca si está repetido: ya tiene su error
func TestModificarProyecto_NombreInvalido_NoSeBuscaRepetido(t *testing.T) {
	e := escenarioEdicion(t)
	_, err := e.edicion("2026-10-13").Modificar(context.Background(), e.proyectoID, datosDeEdicion(t, "ab", "2026-10-05", "2026-11-01"))
	if got := errorDeCampo(t, err, project.CampoNombre); !errors.Is(got, project.ErrNombreLargo) {
		t.Errorf("nombre: se esperaba ErrNombreLargo y se obtuvo %v", got)
	}
}

// CA-02.2 · la fecha de fin no puede quedar antes del fin del último sprint
func TestModificarProyecto_FinAntesDelUltimoSprint(t *testing.T) {
	e := escenarioEdicion(t)
	_, err := e.edicion("2026-10-13").Modificar(context.Background(), e.proyectoID, datosDeEdicion(t, "Software Metrics", "2026-10-05", "2026-10-17"))
	if got := errorDeCampo(t, err, project.CampoFechaFin); !errors.Is(got, project.ErrFinAntesDelUltimoSprint) {
		t.Errorf("fecha_fin: se esperaba ErrFinAntesDelUltimoSprint y se obtuvo %v", got)
	}
	if !e.guardado().FechaFin.Equal(fechaISO(t, "2026-11-01")) {
		t.Errorf("no tenía que guardarse nada y la fecha de fin quedó %v", e.guardado().FechaFin)
	}
}

// CA-02.3 · un proyecto Finalizado no se modifica
func TestModificarProyecto_Finalizado(t *testing.T) {
	e := escenarioEdicion(t)
	d := datosDeEdicion(t, "Software Metrics", "2026-10-05", "2026-11-01")
	d.Descripcion = "Cambio tardío"

	_, err := e.edicion("2026-11-02").Modificar(context.Background(), e.proyectoID, d)

	if !errors.Is(err, project.ErrProyectoFinalizado) {
		t.Fatalf("se esperaba ErrProyectoFinalizado y se obtuvo %v", err)
	}
	if e.guardado().Descripcion == "Cambio tardío" {
		t.Error("un proyecto Finalizado no tenía que guardarse")
	}
}

// RN5 · "hoy" es la fecha de Argentina: a las 22:00 del 01/11 (en UTC ya es 02/11) todavía se puede modificar
func TestModificarProyecto_HoyEsLaFechaDeArgentina(t *testing.T) {
	e := escenarioEdicion(t)
	ahora := time.Date(2026, 11, 2, 1, 0, 0, 0, time.UTC)
	caso := service.NuevoEdicionProyectos(e.proyectos, e.sprints, func() time.Time { return ahora })

	if _, err := caso.Modificar(context.Background(), e.proyectoID, datosDeEdicion(t, "Software Metrics", "2026-10-05", "2026-11-01")); err != nil {
		t.Errorf("no se esperaba error: %v", err)
	}
}

// Sección 7 · el proyecto no existe
func TestModificarProyecto_Inexistente(t *testing.T) {
	e := escenarioEdicion(t)
	caso := e.edicion("2026-10-13")
	if _, err := caso.Modificar(context.Background(), 999, datosDeEdicion(t, "Software Metrics", "2026-10-05", "2026-11-01")); !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("Modificar: se esperaba project.ErrNoEncontrado y se obtuvo %v", err)
	}
	if _, err := caso.Preparar(context.Background(), 999); !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("Preparar: se esperaba project.ErrNoEncontrado y se obtuvo %v", err)
	}
}

// Sección 3 y RN5 · Preparar devuelve el proyecto para cargar el formulario, salvo que esté Finalizado
func TestPrepararModificacion(t *testing.T) {
	e := escenarioEdicion(t)

	p, err := e.edicion("2026-10-13").Preparar(context.Background(), e.proyectoID)
	if err != nil || p.Nombre != "Software Metrics" {
		t.Errorf("se esperaba el proyecto de ejemplo y se obtuvo %+v (error %v)", p, err)
	}
	if _, err := e.edicion("2026-11-02").Preparar(context.Background(), e.proyectoID); !errors.Is(err, project.ErrProyectoFinalizado) {
		t.Errorf("Finalizado: se esperaba ErrProyectoFinalizado y se obtuvo %v", err)
	}
}

// Sección 7 · si falla la base, el error llega al handler
func TestModificarProyecto_ErrorDeLaBase(t *testing.T) {
	e := escenarioEdicion(t)
	caso := service.NuevoEdicionProyectos(e.proyectos, sprintsCaidos{}, relojFijo)

	if _, err := caso.Modificar(context.Background(), e.proyectoID, datosDeEdicion(t, "Software Metrics", "2026-10-05", "2026-11-01")); !errors.Is(err, errSprintsCaidos) {
		t.Errorf("Modificar: se esperaba el error de la base y se obtuvo %v", err)
	}
	if _, err := caso.Preparar(context.Background(), e.proyectoID); !errors.Is(err, errSprintsCaidos) {
		t.Errorf("Preparar: se esperaba el error de la base y se obtuvo %v", err)
	}
}
