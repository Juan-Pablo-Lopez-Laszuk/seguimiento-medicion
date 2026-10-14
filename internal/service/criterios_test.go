package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// escenarioCriterios arma los repositorios en memoria con el proyecto de ejemplo y su HU-1 ya creados, y devuelve el
// caso de uso, el repositorio de criterios, el repositorio de historias, el ID del proyecto y la HU-1.
func escenarioCriterios(t *testing.T) (*service.Criterios, *memory.Criterios, *memory.Historias, int64, backlog.Historia) {
	t.Helper()
	proyectos := memory.NuevoProyectos()
	p, err := service.NuevoProyectos(proyectos, relojFijo).Crear(context.Background(), datosValidos())
	if err != nil {
		t.Fatalf("no se pudo crear el proyecto de ejemplo: %v", err)
	}
	historias := memory.NuevoHistorias()
	h, err := service.NuevoHistorias(proyectos, historias).Crear(context.Background(), p.ID, datosDeHistoria("Iniciar sesión con Google"))
	if err != nil {
		t.Fatalf("no se pudo crear la historia de ejemplo: %v", err)
	}
	criterios := memory.NuevoCriterios()
	return service.NuevoCriterios(proyectos, historias, criterios), criterios, historias, p.ID, h
}

func criteriosDeLaHistoria(t *testing.T, repo *memory.Criterios, historiaID int64) []backlog.Criterio {
	t.Helper()
	lista, err := repo.ListarPorHistoria(context.Background(), historiaID)
	if err != nil {
		t.Fatalf("no se esperaba error al listar: %v", err)
	}
	return lista
}

// CA-07.2 · el criterio queda guardado en la historia, con ID, no cumplido y en el orden en que se cargó
func TestAgregarCriterio_DatosValidos_GuardaNoCumplidoYEnOrden(t *testing.T) {
	s, repo, _, proyectoID, h := escenarioCriterios(t)

	primero, err := s.Agregar(context.Background(), proyectoID, h.Numero, "Entra con su cuenta de Google")
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if _, err := s.Agregar(context.Background(), proyectoID, h.Numero, "Muestra su nombre"); err != nil {
		t.Fatalf("no se esperaba error en el segundo: %v", err)
	}

	if primero.ID == 0 || primero.HistoriaID != h.ID || primero.Cumplido {
		t.Errorf("se esperaba un criterio con ID, de la historia %d y no cumplido, y se obtuvo %+v", h.ID, primero)
	}
	lista := criteriosDeLaHistoria(t, repo, h.ID)
	if len(lista) != 2 || lista[0].Descripcion != "Entra con su cuenta de Google" || lista[1].Descripcion != "Muestra su nombre" {
		t.Errorf("se esperaban los dos criterios en orden y se obtuvo %+v", lista)
	}
}

// RN1 · una descripción inválida se informa en su campo y no se guarda nada
func TestAgregarCriterio_DescripcionInvalida_NoGuarda(t *testing.T) {
	s, repo, _, proyectoID, h := escenarioCriterios(t)

	_, err := s.Agregar(context.Background(), proyectoID, h.Numero, "   ")

	if got := errorDeCampo(t, err, backlog.CampoDescripcion); !errors.Is(got, backlog.ErrDescripcionCriterioObligatoria) {
		t.Errorf("descripcion: se esperaba %v y se obtuvo %v", backlog.ErrDescripcionCriterioObligatoria, got)
	}
	if n := len(criteriosDeLaHistoria(t, repo, h.ID)); n != 0 {
		t.Errorf("no se esperaba ningún criterio guardado y hay %d", n)
	}
}

// historiaHechaEnElRepositorio guarda una historia Hecha con un criterio cumplido, directo en los repositorios
// (pasar una historia a Hecho es de HU-08) y devuelve la historia y su criterio.
func historiaHechaEnElRepositorio(t *testing.T, historias *memory.Historias, criterios *memory.Criterios, proyectoID int64) (backlog.Historia, backlog.Criterio) {
	t.Helper()
	ctx := context.Background()
	h, err := historias.Guardar(ctx, backlog.Historia{
		ProyectoID: proyectoID, Numero: 2, Titulo: "Registrar horas",
		Prioridad: backlog.PrioridadAlta, Estado: backlog.EstadoHecho,
	})
	if err != nil {
		t.Fatalf("no se pudo guardar la historia Hecha: %v", err)
	}
	c, err := criterios.Guardar(ctx, backlog.Criterio{HistoriaID: h.ID, Descripcion: "Se guardan las horas", Cumplido: true})
	if err != nil {
		t.Fatalf("no se pudo guardar el criterio: %v", err)
	}
	return h, c
}

// RN7 · una historia Hecha no admite criterios nuevos ni que se desmarquen los que tiene
func TestCriterios_HistoriaHecha_NoSeCambian(t *testing.T) {
	s, repo, historias, proyectoID, _ := escenarioCriterios(t)
	hecha, criterio := historiaHechaEnElRepositorio(t, historias, repo, proyectoID)

	if _, err := s.Agregar(context.Background(), proyectoID, hecha.Numero, "Uno más"); !errors.Is(err, backlog.ErrHistoriaHecha) {
		t.Errorf("agregar: se esperaba backlog.ErrHistoriaHecha y se obtuvo %v", err)
	}
	if _, err := s.Marcar(context.Background(), proyectoID, hecha.Numero, criterio.ID, false); !errors.Is(err, backlog.ErrHistoriaHecha) {
		t.Errorf("desmarcar: se esperaba backlog.ErrHistoriaHecha y se obtuvo %v", err)
	}
	lista := criteriosDeLaHistoria(t, repo, hecha.ID)
	if len(lista) != 1 || !lista[0].Cumplido {
		t.Errorf("tenía que seguir el único criterio, cumplido, y se obtuvo %+v", lista)
	}
}

// El proyecto y la historia tienen que existir
func TestCriterios_ProyectoOHistoriaInexistente_DevuelveNoEncontrado(t *testing.T) {
	s, repo, _, proyectoID, h := escenarioCriterios(t)

	if _, err := s.Agregar(context.Background(), 99, h.Numero, "Un criterio"); !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("proyecto inexistente: se esperaba project.ErrNoEncontrado y se obtuvo %v", err)
	}
	if _, err := s.Agregar(context.Background(), proyectoID, 99, "Un criterio"); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("historia inexistente: se esperaba backlog.ErrHistoriaNoEncontrada y se obtuvo %v", err)
	}
	if n := len(criteriosDeLaHistoria(t, repo, h.ID)); n != 0 {
		t.Errorf("no se esperaba ningún criterio guardado y hay %d", n)
	}
}

// CA-07.2 y RN3 · marcar deja el criterio cumplido o no cumplido, y repetirlo lo deja igual
func TestMarcarCriterio_CumplidoYNoCumplido(t *testing.T) {
	s, repo, _, proyectoID, h := escenarioCriterios(t)
	c, _ := s.Agregar(context.Background(), proyectoID, h.Numero, "Entra con su cuenta")

	for _, esperado := range []bool{true, true, false, false} {
		marcado, err := s.Marcar(context.Background(), proyectoID, h.Numero, c.ID, esperado)
		if err != nil {
			t.Fatalf("marcar %v: no se esperaba error: %v", esperado, err)
		}
		if marcado.Cumplido != esperado {
			t.Errorf("marcar %v: el criterio devuelto quedó en %v", esperado, marcado.Cumplido)
		}
		if lista := criteriosDeLaHistoria(t, repo, h.ID); len(lista) != 1 || lista[0].Cumplido != esperado {
			t.Errorf("marcar %v: lo guardado es %+v", esperado, lista)
		}
	}
}

// RN8 · un criterio de otra historia, o que no existe, es un "no encontrado" y no se toca
func TestMarcarCriterio_DeOtraHistoriaOInexistente_DevuelveNoEncontrado(t *testing.T) {
	s, repo, _, proyectoID, h := escenarioCriterios(t)
	deOtraHistoria, _ := repo.Guardar(context.Background(), backlog.Criterio{HistoriaID: h.ID + 100, Descripcion: "De otra historia"})

	if _, err := s.Marcar(context.Background(), proyectoID, h.Numero, deOtraHistoria.ID, true); !errors.Is(err, backlog.ErrCriterioNoEncontrado) {
		t.Errorf("criterio de otra historia: se esperaba backlog.ErrCriterioNoEncontrado y se obtuvo %v", err)
	}
	if _, err := s.Marcar(context.Background(), proyectoID, h.Numero, 99, true); !errors.Is(err, backlog.ErrCriterioNoEncontrado) {
		t.Errorf("criterio inexistente: se esperaba backlog.ErrCriterioNoEncontrado y se obtuvo %v", err)
	}
	if _, err := s.Marcar(context.Background(), proyectoID, 99, deOtraHistoria.ID, true); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("historia inexistente: se esperaba backlog.ErrHistoriaNoEncontrada y se obtuvo %v", err)
	}
	if lista := criteriosDeLaHistoria(t, repo, h.ID+100); len(lista) != 1 || lista[0].Cumplido {
		t.Errorf("el criterio de la otra historia no tenía que cambiar y se obtuvo %+v", lista)
	}
}

// CA-07.3 · la página de la historia trae la historia, sus criterios en orden y el progreso
func TestVerHistoria_DevuelveLaHistoriaSusCriteriosYElProgreso(t *testing.T) {
	s, _, _, proyectoID, h := escenarioCriterios(t)
	a, _ := s.Agregar(context.Background(), proyectoID, h.Numero, "Primero")
	_, _ = s.Agregar(context.Background(), proyectoID, h.Numero, "Segundo")
	_, _ = s.Agregar(context.Background(), proyectoID, h.Numero, "Tercero")
	_, _ = s.Marcar(context.Background(), proyectoID, h.Numero, a.ID, true)

	v, err := s.Ver(context.Background(), proyectoID, h.Numero)

	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if v.Proyecto.ID != proyectoID || v.Historia.ID != h.ID || len(v.Criterios) != 3 || v.Criterios[0].Descripcion != "Primero" {
		t.Errorf("se esperaban el proyecto, la HU-1 y sus 3 criterios en orden, y se obtuvo %+v", v)
	}
	if v.Progreso.Texto() != "1/3" {
		t.Errorf("progreso = %s, se esperaba 1/3", v.Progreso.Texto())
	}
}

func TestVerHistoria_ProyectoOHistoriaInexistente_DevuelveNoEncontrado(t *testing.T) {
	s, _, _, proyectoID, _ := escenarioCriterios(t)

	if _, err := s.Ver(context.Background(), 99, 1); !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("proyecto inexistente: se esperaba project.ErrNoEncontrado y se obtuvo %v", err)
	}
	if _, err := s.Ver(context.Background(), proyectoID, 99); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("historia inexistente: se esperaba backlog.ErrHistoriaNoEncontrada y se obtuvo %v", err)
	}
}

// RN8 y CA-07.3 · el progreso de toda la lista de una vez: cada historia cuenta solo los suyos, y las que no tienen
// criterios figuran en 0/0
func TestProgresos_CadaHistoriaCuentaSusCriterios(t *testing.T) {
	s, _, _, proyectoID, primera := escenarioCriterios(t)
	_, _ = s.Agregar(context.Background(), proyectoID, primera.Numero, "Uno")
	dos, _ := s.Agregar(context.Background(), proyectoID, primera.Numero, "Dos")
	_, _ = s.Marcar(context.Background(), proyectoID, primera.Numero, dos.ID, true)
	sinCriterios := backlog.Historia{ID: 500}

	progresos, err := s.Progresos(context.Background(), []backlog.Historia{primera, sinCriterios})

	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if progresos[primera.ID].Texto() != "1/2" {
		t.Errorf("la HU-1 tenía que estar en 1/2 y está en %s", progresos[primera.ID].Texto())
	}
	if p, hay := progresos[sinCriterios.ID]; !hay || p.Texto() != "0/0" {
		t.Errorf("la historia sin criterios tenía que figurar en 0/0 y se obtuvo %+v (figura: %v)", p, hay)
	}
}

// buscadorDeHistoriasQueFalla y criteriosQueFallan simulan una base de datos caída.
type buscadorDeHistoriasQueFalla struct{ err error }

func (b buscadorDeHistoriasQueFalla) BuscarPorNumero(context.Context, int64, int) (backlog.Historia, error) {
	return backlog.Historia{}, b.err
}

// historiaDeEjemplo cumple la interfaz de búsqueda devolviendo siempre una historia Pendiente.
type historiaDeEjemplo struct{}

func (historiaDeEjemplo) BuscarPorNumero(_ context.Context, proyectoID int64, numero int) (backlog.Historia, error) {
	return backlog.Historia{ID: 4, ProyectoID: proyectoID, Numero: numero, Estado: backlog.EstadoPendiente}, nil
}

type criteriosQueFallan struct{ errListar, errGuardar, errBuscar, errActualizar error }

func (r criteriosQueFallan) ListarPorHistoria(context.Context, int64) ([]backlog.Criterio, error) {
	return nil, r.errListar
}
func (r criteriosQueFallan) ListarPorHistorias(context.Context, []int64) ([]backlog.Criterio, error) {
	return nil, r.errListar
}
func (r criteriosQueFallan) Guardar(_ context.Context, c backlog.Criterio) (backlog.Criterio, error) {
	return c, r.errGuardar
}
func (r criteriosQueFallan) BuscarPorID(_ context.Context, historiaID, id int64) (backlog.Criterio, error) {
	return backlog.Criterio{ID: id, HistoriaID: historiaID}, r.errBuscar
}
func (r criteriosQueFallan) Actualizar(context.Context, backlog.Criterio) error {
	return r.errActualizar
}

// Un error del repositorio no es de validación ni de "no encontrado": se devuelve para mostrar un error general.
func TestCriterios_ErrorDelRepositorio_SeDevuelve(t *testing.T) {
	falla := errors.New("base caída")
	ctx := context.Background()
	casos := map[string]func() error{
		"agregar, al buscar el proyecto": func() error {
			_, err := service.NuevoCriterios(proyectosQueFallan{err: falla}, historiaDeEjemplo{}, criteriosQueFallan{}).Agregar(ctx, 1, 1, "Uno")
			return err
		},
		"agregar, al buscar la historia": func() error {
			_, err := service.NuevoCriterios(proyectoDeEjemplo{}, buscadorDeHistoriasQueFalla{err: falla}, criteriosQueFallan{}).Agregar(ctx, 1, 1, "Uno")
			return err
		},
		"agregar, al guardar": func() error {
			_, err := service.NuevoCriterios(proyectoDeEjemplo{}, historiaDeEjemplo{}, criteriosQueFallan{errGuardar: falla}).Agregar(ctx, 1, 1, "Uno")
			return err
		},
		"marcar, al buscar el criterio": func() error {
			_, err := service.NuevoCriterios(proyectoDeEjemplo{}, historiaDeEjemplo{}, criteriosQueFallan{errBuscar: falla}).Marcar(ctx, 1, 1, 1, true)
			return err
		},
		"marcar, al actualizar": func() error {
			_, err := service.NuevoCriterios(proyectoDeEjemplo{}, historiaDeEjemplo{}, criteriosQueFallan{errActualizar: falla}).Marcar(ctx, 1, 1, 1, true)
			return err
		},
		"ver, al listar los criterios": func() error {
			_, err := service.NuevoCriterios(proyectoDeEjemplo{}, historiaDeEjemplo{}, criteriosQueFallan{errListar: falla}).Ver(ctx, 1, 1)
			return err
		},
		"progresos, al listar los criterios": func() error {
			_, err := service.NuevoCriterios(proyectoDeEjemplo{}, historiaDeEjemplo{}, criteriosQueFallan{errListar: falla}).Progresos(ctx, []backlog.Historia{{ID: 1}})
			return err
		},
	}
	for nombre, accion := range casos {
		if err := accion(); !errors.Is(err, falla) {
			t.Errorf("%s: se esperaba el error del repositorio y se obtuvo: %v", nombre, err)
		}
	}
}
