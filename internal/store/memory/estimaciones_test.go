package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// relojDeEjemplo devuelve un reloj que avanza una hora por cada vez que se lo consulta, empezando el 20/10/2026 a las
// 18:00, para saber qué hora tiene cada cambio.
func relojDeEjemplo() func() time.Time {
	hora := time.Date(2026, 10, 20, 17, 0, 0, 0, time.UTC)
	return func() time.Time {
		hora = hora.Add(time.Hour)
		return hora
	}
}

// historiaParaReestimar guarda una historia sin estimar en un repositorio con el reloj de ejemplo.
func historiaParaReestimar(t *testing.T, numero int) (*memory.Historias, backlog.Historia) {
	t.Helper()
	repo := memory.NuevoHistorias().ConReloj(relojDeEjemplo())
	h, err := repo.Guardar(context.Background(), backlog.Historia{ProyectoID: 1, Numero: numero, Titulo: "Historia", Estado: backlog.EstadoPendiente})
	if err != nil {
		t.Fatalf("no se esperaba error al guardar: %v", err)
	}
	return repo, h
}

// CA-16.3 y RN1 · re-estimar anota el valor anterior, el nuevo y la hora; la lista va de la más reciente a la más antigua
func TestHistorias_Reestimar_AnotaElValorAnterior(t *testing.T) {
	ctx := context.Background()
	repo, h := historiaParaReestimar(t, 1)
	for _, puntos := range []int{5, 8, 3} {
		if err := repo.ActualizarStoryPoints(ctx, h.ID, puntos); err != nil {
			t.Fatalf("%d story points: no se esperaba error: %v", puntos, err)
		}
	}

	historial, err := repo.ListarEstimaciones(ctx, h.ID)

	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	// el reloj se consulta una vez por cada estimación: 18:00, 19:00 y 20:00
	esperado := []backlog.Reestimacion{
		{Anterior: 8, Nuevo: 3, CambiadaEn: time.Date(2026, 10, 20, 20, 0, 0, 0, time.UTC)},
		{Anterior: 5, Nuevo: 8, CambiadaEn: time.Date(2026, 10, 20, 19, 0, 0, 0, time.UTC)},
	}
	if len(historial) != len(esperado) {
		t.Fatalf("se esperaban %d entradas y se obtuvieron %d: %+v", len(esperado), len(historial), historial)
	}
	for i := range esperado {
		if historial[i].Anterior != esperado[i].Anterior || historial[i].Nuevo != esperado[i].Nuevo || !historial[i].CambiadaEn.Equal(esperado[i].CambiadaEn) {
			t.Errorf("entrada %d: se esperaba %+v y se obtuvo %+v", i, esperado[i], historial[i])
		}
	}
}

// RN2, RN3 y RN4 · la primera estimación y pedir el mismo valor no se anotan; el 0 sí es un valor
func TestHistorias_Reestimar_SoloAnotaLosCambios(t *testing.T) {
	ctx := context.Background()
	repo, h := historiaParaReestimar(t, 1)

	for _, puntos := range []int{0, 0, 5, 5, 0} { // primera (0), igual (0), 0 → 5, igual (5), 5 → 0
		if err := repo.ActualizarStoryPoints(ctx, h.ID, puntos); err != nil {
			t.Fatalf("%d story points: no se esperaba error: %v", puntos, err)
		}
	}

	historial, _ := repo.ListarEstimaciones(ctx, h.ID)
	if len(historial) != 2 || historial[0].Anterior != 5 || historial[0].Nuevo != 0 || historial[1].Anterior != 0 || historial[1].Nuevo != 5 {
		t.Errorf("se esperaban 5 → 0 y 0 → 5, y se obtuvo %+v", historial)
	}
	if h, _ := repo.BuscarPorNumero(ctx, 1, 1); h.StoryPoints == nil || *h.StoryPoints != 0 {
		t.Errorf("la historia tenía que quedar con 0 story points y quedó con %v", h.StoryPoints)
	}
}

// RN7 · el historial es de cada historia
func TestHistorias_Reestimar_ElHistorialEsDeCadaHistoria(t *testing.T) {
	ctx := context.Background()
	repo, primera := historiaParaReestimar(t, 1)
	segunda, _ := repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 2, Titulo: "Otra", Estado: backlog.EstadoPendiente})
	for _, puntos := range []int{5, 8} {
		_ = repo.ActualizarStoryPoints(ctx, primera.ID, puntos)
	}
	_ = repo.ActualizarStoryPoints(ctx, segunda.ID, 2)

	deLaPrimera, _ := repo.ListarEstimaciones(ctx, primera.ID)
	deLaSegunda, _ := repo.ListarEstimaciones(ctx, segunda.ID)

	if len(deLaPrimera) != 1 {
		t.Errorf("la primera historia tenía que tener 1 entrada y tiene %d", len(deLaPrimera))
	}
	if len(deLaSegunda) != 0 {
		t.Errorf("la segunda historia no tenía que tener entradas y tiene %+v", deLaSegunda)
	}
}

// Una historia que no existe o nunca se re-estimó tiene el historial vacío, sin error
func TestHistorias_ListarEstimaciones_SinCambios_EstaVacio(t *testing.T) {
	ctx := context.Background()
	repo, h := historiaParaReestimar(t, 1)

	if historial, err := repo.ListarEstimaciones(ctx, h.ID); err != nil || len(historial) != 0 {
		t.Errorf("historia sin re-estimar: se esperaba vacío y se obtuvo %+v, %v", historial, err)
	}
	if historial, err := repo.ListarEstimaciones(ctx, 99); err != nil || len(historial) != 0 {
		t.Errorf("historia que no existe: se esperaba vacío y se obtuvo %+v, %v", historial, err)
	}
}

// lo que devuelve el repositorio no se comparte con lo guardado
func TestHistorias_ListarEstimaciones_DevuelveUnaCopia(t *testing.T) {
	ctx := context.Background()
	repo, h := historiaParaReestimar(t, 1)
	_ = repo.ActualizarStoryPoints(ctx, h.ID, 5)
	_ = repo.ActualizarStoryPoints(ctx, h.ID, 8)

	historial, _ := repo.ListarEstimaciones(ctx, h.ID)
	historial[0].Nuevo = 99

	if otraVez, _ := repo.ListarEstimaciones(ctx, h.ID); otraVez[0].Nuevo != 8 {
		t.Errorf("modificar lo devuelto no tenía que cambiar lo guardado: %+v", otraVez)
	}
}

// RN9 · al eliminar la historia se elimina su historial, y el de las demás queda
func TestHistorias_Eliminar_EliminaElHistorial(t *testing.T) {
	ctx := context.Background()
	repo, primera := historiaParaReestimar(t, 1)
	segunda, _ := repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 2, Titulo: "Otra", Estado: backlog.EstadoPendiente})
	for _, id := range []int64{primera.ID, segunda.ID} {
		_ = repo.ActualizarStoryPoints(ctx, id, 5)
		_ = repo.ActualizarStoryPoints(ctx, id, 8)
	}

	if err := repo.Eliminar(ctx, primera.ID); err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if historial, _ := repo.ListarEstimaciones(ctx, primera.ID); len(historial) != 0 {
		t.Errorf("el historial de la historia eliminada tenía que desaparecer y quedó %+v", historial)
	}
	if historial, _ := repo.ListarEstimaciones(ctx, segunda.ID); len(historial) != 1 {
		t.Errorf("el historial de la otra historia tenía que quedar y tiene %+v", historial)
	}
}

// Sin reloj propio, el repositorio usa la hora de verdad
func TestHistorias_Reestimar_SinReloj_UsaLaHoraActual(t *testing.T) {
	ctx := context.Background()
	repo := memory.NuevoHistorias()
	h, _ := repo.Guardar(ctx, backlog.Historia{ProyectoID: 1, Numero: 1, Titulo: "Historia", Estado: backlog.EstadoPendiente})
	antes := time.Now()

	_ = repo.ActualizarStoryPoints(ctx, h.ID, 5)
	_ = repo.ActualizarStoryPoints(ctx, h.ID, 8)

	historial, _ := repo.ListarEstimaciones(ctx, h.ID)
	if len(historial) != 1 || historial[0].CambiadaEn.Before(antes) || historial[0].CambiadaEn.After(time.Now()) {
		t.Errorf("se esperaba una entrada con la hora de ahora (desde %v) y se obtuvo %+v", antes, historial)
	}
}
