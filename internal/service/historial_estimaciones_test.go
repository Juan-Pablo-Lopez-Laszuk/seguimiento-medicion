package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// relojDeEstimaciones avanza una hora por cada consulta, desde las 18:00 del 20/10/2026.
func relojDeEstimaciones() func() time.Time {
	hora := time.Date(2026, 10, 20, 17, 0, 0, 0, time.UTC)
	return func() time.Time {
		hora = hora.Add(time.Hour)
		return hora
	}
}

// escenarioHistorial arma el proyecto de ejemplo con su HU-1 sin estimar, y devuelve los casos de uso de estimar y de
// ver la historia (este último con el historial conectado), más el ID del proyecto y el número de la historia.
func escenarioHistorial(t *testing.T) (*service.Estimaciones, *service.Criterios, int64, int) {
	t.Helper()
	ctx := context.Background()
	proyectos := memory.NuevoProyectos()
	historias := memory.NuevoHistorias().ConReloj(relojDeEstimaciones())
	p, err := service.NuevoProyectos(proyectos, relojFijo).Crear(ctx, datosValidos())
	if err != nil {
		t.Fatalf("no se pudo crear el proyecto de ejemplo: %v", err)
	}
	h, err := service.NuevoHistorias(proyectos, historias).Crear(ctx, p.ID, datosDeHistoria("Iniciar sesión con Google"))
	if err != nil {
		t.Fatalf("no se pudo crear la historia de ejemplo: %v", err)
	}
	ver := service.NuevoCriterios(proyectos, historias, memory.NuevoCriterios()).ConHistorial(historias)
	return service.NuevoEstimaciones(proyectos, historias), ver, p.ID, h.Numero
}

// CA-16.3 · la página de la historia trae su historial de estimaciones, el cambio más reciente primero
func TestVerHistoria_ConHistorial_TraeLasReestimaciones(t *testing.T) {
	ctx := context.Background()
	estimar, ver, proyectoID, numero := escenarioHistorial(t)
	for _, puntos := range []string{"5", "8", "3"} {
		if _, err := estimar.Estimar(ctx, proyectoID, numero, puntos); err != nil {
			t.Fatalf("estimar en %s: no se esperaba error: %v", puntos, err)
		}
	}

	vista, err := ver.Ver(ctx, proyectoID, numero)

	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	esperado := []backlog.Reestimacion{
		{Anterior: 8, Nuevo: 3, CambiadaEn: time.Date(2026, 10, 20, 20, 0, 0, 0, time.UTC)},
		{Anterior: 5, Nuevo: 8, CambiadaEn: time.Date(2026, 10, 20, 19, 0, 0, 0, time.UTC)},
	}
	if len(vista.Historial) != len(esperado) {
		t.Fatalf("se esperaban %d entradas y se obtuvieron %+v", len(esperado), vista.Historial)
	}
	for i := range esperado {
		if vista.Historial[i].Anterior != esperado[i].Anterior || vista.Historial[i].Nuevo != esperado[i].Nuevo || !vista.Historial[i].CambiadaEn.Equal(esperado[i].CambiadaEn) {
			t.Errorf("entrada %d: se esperaba %+v y se obtuvo %+v", i, esperado[i], vista.Historial[i])
		}
	}
}

// RN2 · una historia que nunca se re-estimó tiene el historial vacío (la primera estimación no se anota)
func TestVerHistoria_SinReestimar_TieneElHistorialVacio(t *testing.T) {
	ctx := context.Background()
	estimar, ver, proyectoID, numero := escenarioHistorial(t)
	if _, err := estimar.Estimar(ctx, proyectoID, numero, "5"); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	vista, err := ver.Ver(ctx, proyectoID, numero)

	if err != nil || len(vista.Historial) != 0 {
		t.Errorf("se esperaba el historial vacío y se obtuvo %+v, %v", vista.Historial, err)
	}
}

// Sin el historial conectado (como en los casos de uso que no lo necesitan) la vista no lo trae y no falla
func TestVerHistoria_SinHistorialConectado_NoLoTrae(t *testing.T) {
	ctx := context.Background()
	s, _, _, proyectoID, h := escenarioCriterios(t)

	vista, err := s.Ver(ctx, proyectoID, h.Numero)

	if err != nil || vista.Historial != nil {
		t.Errorf("se esperaba la vista sin historial y sin error, y se obtuvo %+v, %v", vista.Historial, err)
	}
}

// historialQueFalla es un repositorio de historial al que le falla la lectura.
type historialQueFalla struct{ err error }

func (h historialQueFalla) ListarEstimaciones(context.Context, int64) ([]backlog.Reestimacion, error) {
	return nil, h.err
}

// si falla la lectura del historial se informa el error, no se muestra la página a medias
func TestVerHistoria_FallaElHistorial_DevuelveElError(t *testing.T) {
	ctx := context.Background()
	proyectos := memory.NuevoProyectos()
	historias := memory.NuevoHistorias()
	p, _ := service.NuevoProyectos(proyectos, relojFijo).Crear(ctx, datosValidos())
	h, _ := service.NuevoHistorias(proyectos, historias).Crear(ctx, p.ID, datosDeHistoria("Historia"))
	falla := errors.New("base caída")
	ver := service.NuevoCriterios(proyectos, historias, memory.NuevoCriterios()).ConHistorial(historialQueFalla{err: falla})

	_, err := ver.Ver(ctx, p.ID, h.Numero)

	if !errors.Is(err, falla) {
		t.Errorf("se esperaba el error de la base y se obtuvo %v", err)
	}
}
