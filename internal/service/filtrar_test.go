package service_test

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
)

// CA-10.1 y RN3 · Buscar filtra las dos secciones del backlog, conserva su orden y cuenta el total del proyecto
func TestBuscar_FiltraLasDosSeccionesYCuentaElTotal(t *testing.T) {
	s, repo, proyectoID := backlogParaPriorizar(t) // HU-1 Login Baja, HU-2 Registro Alta, HU-3 Perfil Media, HU-4 Reportes Alta
	login, _ := repo.BuscarPorNumero(context.Background(), proyectoID, 1)
	completada := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	if err := repo.ActualizarCompletada(context.Background(), login.ID, backlog.EstadoHecho, &completada); err != nil {
		t.Fatalf("completar: %v", err)
	}

	v, err := s.Buscar(context.Background(), proyectoID, backlog.NuevoFiltro("", "", "r"))
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	// "r" está en Registro, Reportes y Perfil; Login no la tiene.
	if got := numerosDe(v.Pendientes); !slices.Equal(got, []int{2, 4, 3}) {
		t.Errorf("pendientes = HU %v, se esperaba HU [2 4 3]", got)
	}
	if len(v.Hechas) != 0 {
		t.Errorf("Hechas = HU %v, se esperaba ninguna", numerosDe(v.Hechas))
	}
	if v.Total != 4 || v.Filtro.Texto != "r" {
		t.Errorf("total = %d y filtro %+v, se esperaba 4 y el filtro usado", v.Total, v.Filtro)
	}
}

// RN1 · sin filtros es el backlog completo, igual que Backlog
func TestBuscar_SinFiltrosEsElBacklogCompleto(t *testing.T) {
	s, _, proyectoID := backlogParaPriorizar(t)

	v, err := s.Buscar(context.Background(), proyectoID, backlog.Filtro{})
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	completo, _ := s.Backlog(context.Background(), proyectoID)

	if !slices.Equal(numerosDe(v.Pendientes), numerosDe(completo.Pendientes)) || v.Total != 4 || completo.Total != 4 {
		t.Errorf("Buscar sin filtros = HU %v (total %d), Backlog = HU %v (total %d)", numerosDe(v.Pendientes), v.Total, numerosDe(completo.Pendientes), completo.Total)
	}
}

func TestBuscar_ProyectoInexistente(t *testing.T) {
	s, _, _ := backlogParaPriorizar(t)

	if _, err := s.Buscar(context.Background(), 999, backlog.Filtro{}); !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("error = %v, se esperaba project.ErrNoEncontrado", err)
	}
}
