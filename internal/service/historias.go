package service

import (
	"context"
	"fmt"
	"slices"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
)

// RepositorioHistorias es lo que el caso de uso necesita de la base de datos. Lo implementan
// internal/store/memory (tests) e internal/store/postgres (Supabase).
type RepositorioHistorias interface {
	// ListarPorProyecto devuelve las historias del proyecto en el orden en que se crearon.
	ListarPorProyecto(ctx context.Context, proyectoID int64) ([]backlog.Historia, error)
	// Guardar guarda la historia y la devuelve con su ID.
	Guardar(ctx context.Context, h backlog.Historia) (backlog.Historia, error)
	// ActualizarOrden guarda el orden de cada historia del proyecto (por ID), todas o ninguna; si alguna no existe en
	// el proyecto devuelve backlog.ErrHistoriaNoEncontrada (HU-09).
	ActualizarOrden(ctx context.Context, proyectoID int64, ordenes map[int64]int) error
}

// Historias agrupa los casos de uso del Product Backlog (épica E2).
type Historias struct {
	proyectos BuscadorProyectos
	repo      RepositorioHistorias
}

// NuevoHistorias arma el caso de uso.
func NuevoHistorias(proyectos BuscadorProyectos, repo RepositorioHistorias) *Historias {
	return &Historias{proyectos: proyectos, repo: repo}
}

// Crear valida los datos contra las historias del proyecto (HU-05) y guarda la historia. Los datos inválidos se
// devuelven juntos en domain.ErroresValidacion; si el proyecto no existe, el error es project.ErrNoEncontrado;
// cualquier otro error viene de los repositorios.
func (s *Historias) Crear(ctx context.Context, proyectoID int64, d backlog.Datos) (backlog.Historia, error) {
	_, existentes, err := s.Listar(ctx, proyectoID)
	if err != nil {
		return backlog.Historia{}, err
	}
	nueva, err := backlog.Nueva(d, proyectoID, existentes)
	if err != nil {
		return backlog.Historia{}, err
	}
	guardada, err := s.repo.Guardar(ctx, nueva)
	if err != nil {
		return backlog.Historia{}, fmt.Errorf("guardar la historia: %w", err)
	}
	return guardada, nil
}

// VistaBacklog es lo que muestra la página del backlog (HU-09): las historias que no están Hechas, en el orden del
// backlog, y las Hechas aparte, la más reciente primero.
type VistaBacklog struct {
	Proyecto   project.Proyecto
	Pendientes []backlog.Historia
	Hechas     []backlog.Historia
}

// Backlog devuelve el proyecto con su backlog ordenado (backlog.Ordenar). Si el proyecto no existe, el error es
// project.ErrNoEncontrado.
func (s *Historias) Backlog(ctx context.Context, proyectoID int64) (VistaBacklog, error) {
	p, lista, err := s.Listar(ctx, proyectoID)
	if err != nil {
		return VistaBacklog{}, err
	}
	pendientes, hechas := backlog.Ordenar(lista)
	return VistaBacklog{Proyecto: p, Pendientes: pendientes, Hechas: hechas}, nil
}

// Mover sube o baja la historia HU-numero un lugar dentro de su prioridad (HU-09), guarda el orden nuevo de su
// prioridad y devuelve la historia. Si la regla no lo deja (backlog.Mover) devuelve su error y no guarda nada; si el
// proyecto no existe, project.ErrNoEncontrado.
func (s *Historias) Mover(ctx context.Context, proyectoID int64, numero int, m backlog.Movimiento) (backlog.Historia, error) {
	_, lista, err := s.Listar(ctx, proyectoID)
	if err != nil {
		return backlog.Historia{}, err
	}
	ordenes, err := backlog.Mover(lista, numero, m)
	if err != nil {
		return backlog.Historia{}, err
	}
	if err := s.repo.ActualizarOrden(ctx, proyectoID, ordenes); err != nil {
		return backlog.Historia{}, fmt.Errorf("guardar el orden del backlog: %w", err)
	}
	i := slices.IndexFunc(lista, func(h backlog.Historia) bool { return h.Numero == numero })
	movida := lista[i]
	movida.Orden = ordenes[movida.ID]
	return movida, nil
}

// Listar devuelve el proyecto y sus historias en el orden en que se crearon.
func (s *Historias) Listar(ctx context.Context, proyectoID int64) (project.Proyecto, []backlog.Historia, error) {
	p, err := s.proyectos.BuscarPorID(ctx, proyectoID)
	if err != nil {
		return project.Proyecto{}, nil, fmt.Errorf("buscar el proyecto %d: %w", proyectoID, err)
	}
	lista, err := s.repo.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return project.Proyecto{}, nil, fmt.Errorf("listar las historias del proyecto %d: %w", proyectoID, err)
	}
	return p, lista, nil
}
