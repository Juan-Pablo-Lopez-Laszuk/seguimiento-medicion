package postgres_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/postgres"
)

// HU-17 RN3 · abrir guarda la sesión Abierta con su Ronda 1, sin revelar
func TestPoker_AbrirGuardaLaSesionConLaRonda1(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	h, _ := postgres.NuevoHistorias(pool).Guardar(ctx, historiaDeEjemplo(proyectoA, 1, "Login"))
	repo := postgres.NuevoPoker(pool)

	sesion, ronda, err := repo.Abrir(ctx, h.ID)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if sesion.ID == 0 || sesion.HistoriaID != h.ID || sesion.Estado != poker.EstadoAbierta || sesion.ValorAcordado != nil {
		t.Errorf("sesión = %+v, se esperaba una sesión Abierta de la historia %d con ID", sesion, h.ID)
	}
	if ronda.ID == 0 || ronda.SesionID != sesion.ID || ronda.Numero != 1 || ronda.Revelada {
		t.Errorf("ronda = %+v, se esperaba la Ronda 1 sin revelar de la sesión %d", ronda, sesion.ID)
	}
	abierta, actual, err := repo.SesionAbierta(ctx, h.ID)
	if err != nil || abierta != sesion || actual != ronda {
		t.Errorf("SesionAbierta = %+v, %+v, %v; se esperaba la sesión y la ronda guardadas", abierta, actual, err)
	}
}

// CA-17.2 · una sola sesión abierta por historia (índice único), aunque se abra dos veces a la vez; y no queda una
// ronda suelta del intento rechazado (RN3, transacción)
func TestPoker_AbrirDosVecesALaVez(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	h, _ := postgres.NuevoHistorias(pool).Guardar(ctx, historiaDeEjemplo(proyectoA, 1, "Login"))
	repo := postgres.NuevoPoker(pool)

	errs := make([]error, 2)
	var wg sync.WaitGroup
	for i := range errs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, errs[i] = repo.Abrir(ctx, h.ID)
		}()
	}
	wg.Wait()

	rechazados := 0
	for _, err := range errs {
		switch {
		case errors.Is(err, poker.ErrSesionYaAbierta):
			rechazados++
		case err != nil:
			t.Errorf("error inesperado: %v", err)
		}
	}
	if rechazados != 1 {
		t.Errorf("se esperaba un solo rechazo y hubo %d (%v)", rechazados, errs)
	}
	if n := cuantos(t, pool, "sesiones_poker", "historia_id = $1", h.ID); n != 1 {
		t.Errorf("se esperaba 1 sesión y hay %d", n)
	}
	if n := cuantos(t, pool, "rondas_poker", "true"); n != 1 {
		t.Errorf("se esperaba 1 ronda y hay %d", n)
	}
}

// RN2 · cuando la sesión se cierra (HU-22) se puede abrir otra; la ronda actual es la de número más alto
func TestPoker_DespuesDeCerrarSeAbreOtra(t *testing.T) {
	ctx := context.Background()
	pool := baseMigrada(t)
	proyectoA, _ := dosProyectos(t, pool)
	h, _ := postgres.NuevoHistorias(pool).Guardar(ctx, historiaDeEjemplo(proyectoA, 1, "Login"))
	repo := postgres.NuevoPoker(pool)
	primera, _, _ := repo.Abrir(ctx, h.ID)
	if _, err := pool.Exec(ctx, `UPDATE sesiones_poker SET estado = 'Cerrada', valor_acordado = 5 WHERE id = $1`, primera.ID); err != nil {
		t.Fatalf("cerrar la sesión: %v", err)
	}

	segunda, _, err := repo.Abrir(ctx, h.ID)
	if err != nil || segunda.ID == primera.ID {
		t.Fatalf("se esperaba una sesión nueva y se obtuvo %+v, %v", segunda, err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO rondas_poker (sesion_id, numero) VALUES ($1, 2)`, segunda.ID); err != nil {
		t.Fatalf("agregar la Ronda 2: %v", err)
	}

	abierta, ronda, err := repo.SesionAbierta(ctx, h.ID)
	if err != nil || abierta.ID != segunda.ID || ronda.Numero != 2 {
		t.Errorf("SesionAbierta = %+v, %+v, %v; se esperaba la sesión %d en la Ronda 2", abierta, ronda, err, segunda.ID)
	}
}

func TestPoker_SesionAbierta_SinSesion(t *testing.T) {
	pool := baseMigrada(t)
	if _, _, err := postgres.NuevoPoker(pool).SesionAbierta(context.Background(), 99); !errors.Is(err, poker.ErrSesionNoEncontrada) {
		t.Errorf("error = %v, se esperaba ErrSesionNoEncontrada", err)
	}
}
