package memory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// HU-17 RN3 · abrir guarda la sesión Abierta con su Ronda 1, sin revelar
func TestPoker_AbrirGuardaLaSesionConLaRonda1(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoPoker()

	sesion, ronda, err := repo.Abrir(ctx, 7)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if sesion.ID == 0 || sesion.HistoriaID != 7 || sesion.Estado != poker.EstadoAbierta || sesion.ValorAcordado != nil {
		t.Errorf("sesión = %+v, se esperaba una sesión Abierta de la historia 7 con ID", sesion)
	}
	if ronda.ID == 0 || ronda.SesionID != sesion.ID || ronda.Numero != 1 || ronda.Revelada {
		t.Errorf("ronda = %+v, se esperaba la Ronda 1 sin revelar de la sesión %d", ronda, sesion.ID)
	}
	abierta, actual, err := repo.SesionAbierta(ctx, 7)
	if err != nil || abierta != sesion || actual != ronda {
		t.Errorf("SesionAbierta = %+v, %+v, %v; se esperaba la sesión y la ronda guardadas", abierta, actual, err)
	}
}

// CA-17.2 · una sola sesión abierta por historia; la de otra historia no molesta
func TestPoker_AbrirDosVecesLaMismaHistoria(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoPoker()
	primera, _, _ := repo.Abrir(ctx, 7)

	_, _, err := repo.Abrir(ctx, 7)

	if !errors.Is(err, poker.ErrSesionYaAbierta) {
		t.Errorf("error = %v, se esperaba ErrSesionYaAbierta", err)
	}
	if abierta, _, _ := repo.SesionAbierta(ctx, 7); abierta.ID != primera.ID {
		t.Errorf("la sesión abierta tenía que seguir siendo la %d y es %+v", primera.ID, abierta)
	}
	if _, _, err := repo.Abrir(ctx, 8); err != nil {
		t.Errorf("otra historia sí abre su sesión: %v", err)
	}
}

func TestPoker_SesionAbierta_SinSesion(t *testing.T) {
	if _, _, err := memory.NuevoPoker().SesionAbierta(context.Background(), 7); !errors.Is(err, poker.ErrSesionNoEncontrada) {
		t.Errorf("error = %v, se esperaba ErrSesionNoEncontrada", err)
	}
}

// HU-06 RN10 · una historia con una sesión de Planning Poker no se elimina, como con la clave foránea de PostgreSQL
func TestPoker_TieneDeHistoria(t *testing.T) {
	ctx := context.Background()
	sesiones := memory.NuevoPoker()
	historias := memory.NuevoHistorias().ConRegistrosAsociados(sesiones)
	conSesion, _ := historias.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 1, Titulo: "Con sesión"})
	sinSesion, _ := historias.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 2, Titulo: "Sin sesión"})
	_, _, _ = sesiones.Abrir(ctx, conSesion.ID)

	if !sesiones.TieneDeHistoria(conSesion.ID) || sesiones.TieneDeHistoria(sinSesion.ID) {
		t.Errorf("solo la historia %d tiene una sesión", conSesion.ID)
	}
	if err := historias.Eliminar(ctx, conSesion.ID); !errors.Is(err, backlog.ErrHistoriaConRegistros) {
		t.Errorf("error = %v, se esperaba ErrHistoriaConRegistros", err)
	}
}
