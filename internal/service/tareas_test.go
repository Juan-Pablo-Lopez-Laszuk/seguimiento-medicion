package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// escenarioTareas tiene el ejemplo de la spec de HU-25: el proyecto, su HU-1 "Iniciar sesión con Google", la HU-2
// "Ver el tablero" y los integrantes Carolina y Mariano (activos) y Ana (dada de baja).
type escenarioTareas struct {
	casos      *service.Tareas
	repo       *memory.Tareas
	proyectoID int64
	hu1, hu2   backlog.Historia
	carolina   member.Integrante
	mariano    member.Integrante
	ana        member.Integrante
}

func nuevoEscenarioTareas(t *testing.T) escenarioTareas {
	t.Helper()
	ctx := context.Background()
	proyectos, historias, integrantes := memory.NuevoProyectos(), memory.NuevoHistorias(), memory.NuevoIntegrantes()
	p, err := service.NuevoProyectos(proyectos, relojFijo).Crear(ctx, datosValidos())
	if err != nil {
		t.Fatalf("no se pudo crear el proyecto de ejemplo: %v", err)
	}
	e := escenarioTareas{repo: memory.NuevoTareas(), proyectoID: p.ID}

	casosHistorias := service.NuevoHistorias(proyectos, historias)
	for _, h := range []struct {
		destino *backlog.Historia
		titulo  string
	}{{&e.hu1, "Iniciar sesión con Google"}, {&e.hu2, "Ver el tablero"}} {
		if *h.destino, err = casosHistorias.Crear(ctx, p.ID, datosDeHistoria(h.titulo)); err != nil {
			t.Fatalf("no se pudo crear la historia %q: %v", h.titulo, err)
		}
	}

	casosIntegrantes := service.NuevoIntegrantes(proyectos, integrantes)
	for _, i := range []struct {
		destino      *member.Integrante
		nombre, mail string
	}{{&e.carolina, "Carolina", "caro@mail.com"}, {&e.mariano, "Mariano", "mariano@mail.com"}, {&e.ana, "Ana", "ana@mail.com"}} {
		d := member.Datos{Nombre: i.nombre, Email: i.mail, Rol: member.RolProductBuilder}
		if *i.destino, err = casosIntegrantes.Registrar(ctx, p.ID, d); err != nil {
			t.Fatalf("no se pudo registrar a %s: %v", i.nombre, err)
		}
	}
	if e.ana, err = casosIntegrantes.DarDeBaja(ctx, p.ID, e.ana.ID); err != nil {
		t.Fatalf("no se pudo dar de baja a Ana: %v", err)
	}

	e.casos = service.NuevoTareas(proyectos, historias, integrantes, e.repo)
	return e
}

func (e escenarioTareas) tareaDe(responsable member.Integrante, titulo, horas string) effort.DatosTarea {
	return effort.DatosTarea{Titulo: titulo, ResponsableID: responsable.ID, Horas: horas}
}

// agregar agrega una tarea válida a la historia y la devuelve.
func (e escenarioTareas) agregar(t *testing.T, h backlog.Historia, d effort.DatosTarea) effort.Tarea {
	t.Helper()
	tarea, err := e.casos.Agregar(context.Background(), e.proyectoID, h.Numero, d)
	if err != nil {
		t.Fatalf("no se pudo agregar la tarea %q: %v", d.Titulo, err)
	}
	return tarea
}

func (e escenarioTareas) tareasDe(t *testing.T, h backlog.Historia) []effort.Tarea {
	t.Helper()
	lista, err := e.repo.ListarPorHistoria(context.Background(), h.ID)
	if err != nil {
		t.Fatalf("no se esperaba error al listar: %v", err)
	}
	return lista
}

// CA-25.1, CA-25.2 y RN5 · la tarea queda guardada en la historia de la URL, con ID, sus datos y pendiente, en orden
func TestAgregarTarea_DatosValidos_GuardaEnLaHistoriaYEnOrden(t *testing.T) {
	e := nuevoEscenarioTareas(t)

	primera := e.agregar(t, e.hu1, e.tareaDe(e.carolina, " Pantalla de login ", "3"))
	e.agregar(t, e.hu1, e.tareaDe(e.mariano, "Guardar la sesión", "2,5"))

	esperada := effort.Tarea{ID: primera.ID, HistoriaID: e.hu1.ID, Titulo: "Pantalla de login", ResponsableID: e.carolina.ID, HorasEstimadas: 3}
	if primera.ID == 0 || primera != esperada {
		t.Errorf("tarea = %+v, se esperaba %+v con un ID", primera, esperada)
	}
	lista := e.tareasDe(t, e.hu1)
	if len(lista) != 2 || lista[0].Titulo != "Pantalla de login" || lista[1].Titulo != "Guardar la sesión" {
		t.Errorf("se esperaban las dos tareas en orden y se obtuvo %+v", lista)
	}
}

// RN2 · el responsable se valida contra los integrantes del proyecto: uno dado de baja no puede serlo, y no se guarda
func TestAgregarTarea_ResponsableDadoDeBaja_NoGuarda(t *testing.T) {
	e := nuevoEscenarioTareas(t)

	_, err := e.casos.Agregar(context.Background(), e.proyectoID, e.hu1.Numero, e.tareaDe(e.ana, "Pantalla de login", "3"))

	if got := errorDeCampo(t, err, effort.CampoResponsable); !errors.Is(got, effort.ErrResponsableInvalido) {
		t.Errorf("responsable: se esperaba %v y se obtuvo %v", effort.ErrResponsableInvalido, got)
	}
	if n := len(e.tareasDe(t, e.hu1)); n != 0 {
		t.Errorf("no se esperaba ninguna tarea guardada y hay %d", n)
	}
}

// RN8 y CA-23.2 · la vista trae las tareas de la historia, sus horas sumadas y los integrantes para elegir responsable
func TestVerTareas_TraeLasTareasLaSumaYLosIntegrantes(t *testing.T) {
	e := nuevoEscenarioTareas(t)
	e.agregar(t, e.hu1, e.tareaDe(e.carolina, "Pantalla de login", "3"))
	e.agregar(t, e.hu1, e.tareaDe(e.mariano, "Guardar la sesión", "2,5"))
	e.agregar(t, e.hu2, e.tareaDe(e.mariano, "Armar la tabla", "4"))

	v, err := e.casos.Ver(context.Background(), e.proyectoID, e.hu1.Numero)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if v.Proyecto.ID != e.proyectoID || v.Historia.ID != e.hu1.ID {
		t.Errorf("proyecto %d e historia %d, se esperaba %d y %d", v.Proyecto.ID, v.Historia.ID, e.proyectoID, e.hu1.ID)
	}
	if len(v.Tareas) != 2 || v.HorasEstimadas != 5.5 {
		t.Errorf("se esperaban las 2 tareas de la HU-1 y 5,5 horas, y se obtuvo %d tareas y %v horas", len(v.Tareas), v.HorasEstimadas)
	}
	if activos := v.Activos(); len(activos) != 2 || activos[0].Nombre != "Carolina" || activos[1].Nombre != "Mariano" {
		t.Errorf("para elegir responsable se esperaban Carolina y Mariano, y se obtuvo %+v", activos)
	}
	if v.NombreDe(e.mariano.ID) != "Mariano" || v.NombreDe(e.ana.ID) != "Ana" {
		t.Error("el nombre de cada responsable se busca entre todos los integrantes, aun los dados de baja")
	}
}

// CA-25.3 · la estimación de una tarea pendiente se cambia; la de una completada no, y queda como estaba
func TestCambiarEstimacion_PendienteSiCompletadaNo(t *testing.T) {
	e := nuevoEscenarioTareas(t)
	ctx := context.Background()
	tarea := e.agregar(t, e.hu1, e.tareaDe(e.carolina, "Pantalla de login", "3"))

	if _, err := e.casos.CambiarEstimacion(ctx, e.proyectoID, e.hu1.Numero, tarea.ID, "4"); err != nil {
		t.Fatalf("no se esperaba error al cambiar una pendiente: %v", err)
	}
	if _, err := e.casos.Marcar(ctx, e.proyectoID, e.hu1.Numero, tarea.ID, true); err != nil {
		t.Fatalf("no se esperaba error al completarla: %v", err)
	}
	_, err := e.casos.CambiarEstimacion(ctx, e.proyectoID, e.hu1.Numero, tarea.ID, "5")

	if !errors.Is(err, effort.ErrTareaCompletada) {
		t.Errorf("se esperaba %v y se obtuvo %v", effort.ErrTareaCompletada, err)
	}
	lista := e.tareasDe(t, e.hu1)
	if len(lista) != 1 || lista[0].HorasEstimadas != 4 || !lista[0].Completada {
		t.Errorf("se esperaba la tarea completada con 4 horas y se obtuvo %+v", lista)
	}
}

// RN7 · marcar es explícito: se completa y se vuelve a pendiente
func TestMarcarTarea_CompletadaYPendiente(t *testing.T) {
	e := nuevoEscenarioTareas(t)
	ctx := context.Background()
	tarea := e.agregar(t, e.hu1, e.tareaDe(e.carolina, "Pantalla de login", "3"))

	for _, completada := range []bool{true, true, false} {
		marcada, err := e.casos.Marcar(ctx, e.proyectoID, e.hu1.Numero, tarea.ID, completada)
		if err != nil {
			t.Fatalf("no se esperaba error: %v", err)
		}
		if guardada := e.tareasDe(t, e.hu1)[0]; marcada.Completada != completada || guardada.Completada != completada {
			t.Errorf("completada = %v (guardada %v), se esperaba %v", marcada.Completada, guardada.Completada, completada)
		}
	}
}

// CA-25.2 · una tarea no se cambia desde otra historia: es un "no encontrado" y queda como estaba
func TestTareas_DeOtraHistoria_DevuelveNoEncontrada(t *testing.T) {
	e := nuevoEscenarioTareas(t)
	ctx := context.Background()
	deLaHU2 := e.agregar(t, e.hu2, e.tareaDe(e.mariano, "Armar la tabla", "4"))

	if _, err := e.casos.CambiarEstimacion(ctx, e.proyectoID, e.hu1.Numero, deLaHU2.ID, "6"); !errors.Is(err, effort.ErrTareaNoEncontrada) {
		t.Errorf("cambiar horas: se esperaba %v y se obtuvo %v", effort.ErrTareaNoEncontrada, err)
	}
	if _, err := e.casos.Marcar(ctx, e.proyectoID, e.hu1.Numero, deLaHU2.ID, true); !errors.Is(err, effort.ErrTareaNoEncontrada) {
		t.Errorf("marcar: se esperaba %v y se obtuvo %v", effort.ErrTareaNoEncontrada, err)
	}
	if lista := e.tareasDe(t, e.hu2); lista[0] != deLaHU2 {
		t.Errorf("la tarea de la HU-2 no debería cambiar y quedó %+v", lista[0])
	}
}

// Sección 7 · un proyecto o una historia que no existen son un "no encontrado"
func TestTareas_ProyectoOHistoriaInexistente_DevuelveNoEncontrado(t *testing.T) {
	e := nuevoEscenarioTareas(t)
	ctx := context.Background()

	if _, err := e.casos.Ver(ctx, 99, 1); !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("proyecto inexistente: se esperaba %v y se obtuvo %v", project.ErrNoEncontrado, err)
	}
	if _, err := e.casos.Ver(ctx, e.proyectoID, 99); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("historia inexistente: se esperaba %v y se obtuvo %v", backlog.ErrHistoriaNoEncontrada, err)
	}
	if _, err := e.casos.Agregar(ctx, e.proyectoID, 99, e.tareaDe(e.carolina, "Pantalla de login", "3")); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("agregar a una historia inexistente: se esperaba %v y se obtuvo %v", backlog.ErrHistoriaNoEncontrada, err)
	}
}
