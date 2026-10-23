package service_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// backlogParaPriorizar crea, con el caso de uso, las historias de la feature de HU-09: HU-1 Login Baja, HU-2 Registro
// Alta, HU-3 Perfil Media y HU-4 Reportes Alta.
func backlogParaPriorizar(t *testing.T) (*service.Historias, *memory.Historias, int64) {
	t.Helper()
	s, repo, proyectoID := escenarioHistorias(t)
	for _, d := range []backlog.Datos{
		{Titulo: "Login", Prioridad: backlog.PrioridadBaja},
		{Titulo: "Registro", Prioridad: backlog.PrioridadAlta},
		{Titulo: "Perfil", Prioridad: backlog.PrioridadMedia},
		{Titulo: "Reportes", Prioridad: backlog.PrioridadAlta},
	} {
		if _, err := s.Crear(context.Background(), proyectoID, d); err != nil {
			t.Fatalf("crear %q: %v", d.Titulo, err)
		}
	}
	return s, repo, proyectoID
}

// numerosDe devuelve los números de las historias, en el orden de la lista.
func numerosDe(lista []backlog.Historia) []int {
	var n []int
	for _, h := range lista {
		n = append(n, h.Numero)
	}
	return n
}

// pendientesDelBacklog devuelve los números de las historias pendientes, en el orden del backlog.
func pendientesDelBacklog(t *testing.T, s *service.Historias, proyectoID int64) []int {
	t.Helper()
	v, err := s.Backlog(context.Background(), proyectoID)
	if err != nil {
		t.Fatalf("ver el backlog: %v", err)
	}
	return numerosDe(v.Pendientes)
}

// CA-09.1 y CA-09.3 · el backlog viene ordenado por prioridad, con las Hechas aparte
func TestBacklog_OrdenadoPorPrioridadConLasHechasAparte(t *testing.T) {
	s, repo, proyectoID := backlogParaPriorizar(t)
	login, _ := repo.BuscarPorNumero(context.Background(), proyectoID, 1)
	completada := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	if err := repo.ActualizarCompletada(context.Background(), login.ID, backlog.EstadoHecho, &completada); err != nil {
		t.Fatalf("completar: %v", err)
	}

	v, err := s.Backlog(context.Background(), proyectoID)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if v.Proyecto.ID != proyectoID {
		t.Errorf("proyecto = %d, se esperaba %d", v.Proyecto.ID, proyectoID)
	}
	if got := numerosDe(v.Pendientes); !slices.Equal(got, []int{2, 4, 3}) {
		t.Errorf("pendientes = HU %v, se esperaba HU [2 4 3]", got)
	}
	if got := numerosDe(v.Hechas); !slices.Equal(got, []int{1}) {
		t.Errorf("Hechas = HU %v, se esperaba HU [1]", got)
	}
}

func TestBacklog_ProyectoInexistente(t *testing.T) {
	s, _, _ := escenarioHistorias(t)

	_, err := s.Backlog(context.Background(), 999)

	if !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("error = %v, se esperaba project.ErrNoEncontrado", err)
	}
}

// CA-09.1, CA-09.2 y RN5 · mover guarda el orden nuevo y devuelve la historia movida
func TestMover_GuardaElOrdenNuevo(t *testing.T) {
	s, _, proyectoID := backlogParaPriorizar(t)

	movida, err := s.Mover(context.Background(), proyectoID, 4, backlog.Subir)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if movida.Numero != 4 || movida.Titulo != "Reportes" {
		t.Errorf("se esperaba la HU-4 Reportes y se obtuvo %s %s", movida.Identificador(), movida.Titulo)
	}
	if got := pendientesDelBacklog(t, s, proyectoID); !slices.Equal(got, []int{4, 2, 3, 1}) {
		t.Errorf("backlog = HU %v, se esperaba HU [4 2 3 1]", got)
	}
}

// RN6 · después de reordenar, una historia nueva va al final de su prioridad
func TestMover_UnaNuevaDespuesDeReordenarVaAlFinal(t *testing.T) {
	s, _, proyectoID := backlogParaPriorizar(t)
	if _, err := s.Mover(context.Background(), proyectoID, 4, backlog.Subir); err != nil {
		t.Fatalf("mover: %v", err)
	}

	if _, err := s.Crear(context.Background(), proyectoID, backlog.Datos{Titulo: "Cuentas", Prioridad: backlog.PrioridadAlta}); err != nil {
		t.Fatalf("crear: %v", err)
	}

	if got := pendientesDelBacklog(t, s, proyectoID); !slices.Equal(got, []int{4, 2, 5, 3, 1}) {
		t.Errorf("backlog = HU %v, se esperaba HU [4 2 5 3 1]", got)
	}
}

// RN3 · si la regla no deja mover, no se guarda nada
func TestMover_LaPrimeraNoSubeYNoCambiaNada(t *testing.T) {
	s, _, proyectoID := backlogParaPriorizar(t)

	_, err := s.Mover(context.Background(), proyectoID, 2, backlog.Subir)

	if !errors.Is(err, backlog.ErrYaEsLaPrimera) {
		t.Errorf("error = %v, se esperaba ErrYaEsLaPrimera", err)
	}
	if got := pendientesDelBacklog(t, s, proyectoID); !slices.Equal(got, []int{2, 4, 3, 1}) {
		t.Errorf("backlog = HU %v, se esperaba que no cambie", got)
	}
}

func TestMover_HistoriaOProyectoInexistente(t *testing.T) {
	s, _, proyectoID := backlogParaPriorizar(t)

	if _, err := s.Mover(context.Background(), proyectoID, 99, backlog.Subir); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		t.Errorf("historia inexistente: error = %v, se esperaba ErrHistoriaNoEncontrada", err)
	}
	if _, err := s.Mover(context.Background(), 999, 1, backlog.Subir); !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("proyecto inexistente: error = %v, se esperaba project.ErrNoEncontrado", err)
	}
}

// historiasQueNoGuardanElOrden es el repositorio en memoria con una base que falla al guardar el orden.
type historiasQueNoGuardanElOrden struct{ *memory.Historias }

var errBaseCaida = errors.New("la base no responde")

func (historiasQueNoGuardanElOrden) ActualizarOrden(context.Context, int64, map[int64]int) error {
	return errBaseCaida
}

func TestMover_FallaAlGuardar(t *testing.T) {
	_, repo, proyectoID := backlogParaPriorizar(t)
	s := service.NuevoHistorias(proyectosQueEncuentran{proyectoID}, historiasQueNoGuardanElOrden{repo})

	_, err := s.Mover(context.Background(), proyectoID, 4, backlog.Subir)

	if !errors.Is(err, errBaseCaida) {
		t.Errorf("error = %v, se esperaba el de la base", err)
	}
}

// proyectosQueEncuentran encuentra siempre el proyecto con ese ID.
type proyectosQueEncuentran struct{ id int64 }

func (p proyectosQueEncuentran) BuscarPorID(_ context.Context, id int64) (project.Proyecto, error) {
	if id != p.id {
		return project.Proyecto{}, project.ErrNoEncontrado
	}
	return project.Proyecto{ID: id, Nombre: "Software Metrics"}, nil
}

// HU-09 RN9 · las candidatas de un sprint siguen el orden del backlog
func TestVerSprint_LasCandidatasSiguenElOrdenDelBacklog(t *testing.T) {
	e := nuevoEscenarioAsignacion(t, true)
	ids := map[int]int64{}
	for numero := 1; numero <= 4; numero++ {
		ids[numero] = e.guardada(t, numero).ID
	}
	// Todas son Alta: Reportes (HU-4) queda primera y después las demás por número.
	if err := e.historias.ActualizarOrden(context.Background(), e.proyectoID, map[int64]int{ids[4]: 1, ids[1]: 2, ids[2]: 3, ids[3]: 4}); err != nil {
		t.Fatalf("ordenar: %v", err)
	}

	v, err := e.casos.Ver(context.Background(), e.proyectoID, 2)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	var candidatas []int
	for _, c := range v.Candidatas {
		candidatas = append(candidatas, c.Numero)
	}
	if !slices.Equal(candidatas, []int{4, 1, 2, 3}) {
		t.Errorf("candidatas = HU %v, se esperaba HU [4 1 2 3]", candidatas)
	}
}
