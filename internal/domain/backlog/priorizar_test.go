package backlog_test

import (
	"errors"
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
)

// historiaParaPriorizar arma la HU-N del proyecto 7 (con ID N) con esa prioridad y ese orden manual, Pendiente.
func historiaParaPriorizar(numero int, prioridad backlog.Prioridad, orden int) backlog.Historia {
	return backlog.Historia{ID: int64(numero), ProyectoID: 7, Numero: numero, Titulo: "Historia", Prioridad: prioridad,
		Estado: backlog.EstadoPendiente, Orden: orden}
}

// hecha devuelve la historia Hecha y completada ese día.
func hecha(h backlog.Historia, fecha string) backlog.Historia {
	dia, err := time.Parse(time.DateOnly, fecha)
	if err != nil {
		panic(err)
	}
	h.Estado, h.CompletadaEn = backlog.EstadoHecho, &dia
	return h
}

// numeros devuelve los números de las historias, en el orden de la lista.
func numeros(lista []backlog.Historia) []int {
	var n []int
	for _, h := range lista {
		n = append(n, h.Numero)
	}
	return n
}

// CA-09.1 y RN1 · primero la prioridad (Alta, Media, Baja); dentro de una prioridad, el orden manual y después el número
func TestOrdenar_PorPrioridadDespuesPorOrdenYDespuesPorNumero(t *testing.T) {
	lista := []backlog.Historia{
		historiaParaPriorizar(1, backlog.PrioridadBaja, 0),
		historiaParaPriorizar(2, backlog.PrioridadAlta, 2),
		historiaParaPriorizar(3, backlog.PrioridadMedia, 0),
		historiaParaPriorizar(4, backlog.PrioridadAlta, 1),
		historiaParaPriorizar(5, backlog.PrioridadAlta, 2),
		historiaParaPriorizar(6, backlog.PrioridadMedia, 0),
	}

	pendientes, hechas := backlog.Ordenar(lista)

	if got, want := numeros(pendientes), []int{4, 2, 5, 3, 6, 1}; !slices.Equal(got, want) {
		t.Errorf("pendientes = HU %v, se esperaba HU %v", got, want)
	}
	if len(hechas) != 0 {
		t.Errorf("no se esperaban Hechas y se obtuvo HU %v", numeros(hechas))
	}
}

// CA-09.3 y RN4 · las Hechas van aparte, la completada más recientemente primero; con la misma fecha, por número
func TestOrdenar_LasHechasVanAparteLaMasRecientePrimero(t *testing.T) {
	lista := []backlog.Historia{
		hecha(historiaParaPriorizar(1, backlog.PrioridadAlta, 0), "2026-10-12"),
		historiaParaPriorizar(2, backlog.PrioridadBaja, 0),
		hecha(historiaParaPriorizar(3, backlog.PrioridadBaja, 0), "2026-10-14"),
		hecha(historiaParaPriorizar(4, backlog.PrioridadMedia, 0), "2026-10-12"),
		historiaParaPriorizar(5, backlog.PrioridadAlta, 0),
	}

	pendientes, hechas := backlog.Ordenar(lista)

	if got, want := numeros(pendientes), []int{5, 2}; !slices.Equal(got, want) {
		t.Errorf("pendientes = HU %v, se esperaba HU %v", got, want)
	}
	if got, want := numeros(hechas), []int{3, 1, 4}; !slices.Equal(got, want) {
		t.Errorf("Hechas = HU %v, se esperaba HU %v", got, want)
	}
}

// Ordenar devuelve listas nuevas: la que recibe queda como estaba
func TestOrdenar_NoCambiaLaListaQueRecibe(t *testing.T) {
	lista := []backlog.Historia{historiaParaPriorizar(1, backlog.PrioridadBaja, 0), historiaParaPriorizar(2, backlog.PrioridadAlta, 0)}

	backlog.Ordenar(lista)

	if got := numeros(lista); !slices.Equal(got, []int{1, 2}) {
		t.Errorf("la lista quedó HU %v, se esperaba que no cambie", got)
	}
}

// backlogDelEscenario es el de la feature: HU-1 Baja, HU-2 Alta, HU-3 Media, HU-4 Alta, todas con orden 0.
func backlogDelEscenario() []backlog.Historia {
	return []backlog.Historia{
		historiaParaPriorizar(1, backlog.PrioridadBaja, 0),
		historiaParaPriorizar(2, backlog.PrioridadAlta, 0),
		historiaParaPriorizar(3, backlog.PrioridadMedia, 0),
		historiaParaPriorizar(4, backlog.PrioridadAlta, 0),
	}
}

// CA-09.1, RN2 y RN5 · subir intercambia la historia con la de arriba y numera 1..n las de su prioridad
func TestMover_SubirIntercambiaConLaDeArribaYNumeraLasDeSuPrioridad(t *testing.T) {
	ordenes, err := backlog.Mover(backlogDelEscenario(), 4, backlog.Subir)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	// Solo cambian las de Alta: HU-4 queda primera y HU-2 segunda.
	if want := map[int64]int{4: 1, 2: 2}; !maps.Equal(ordenes, want) {
		t.Errorf("órdenes = %v, se esperaba %v", ordenes, want)
	}
}

// CA-09.1 y RN2 · bajar intercambia la historia con la de abajo
func TestMover_BajarIntercambiaConLaDeAbajo(t *testing.T) {
	lista := []backlog.Historia{
		historiaParaPriorizar(1, backlog.PrioridadMedia, 1),
		historiaParaPriorizar(2, backlog.PrioridadMedia, 2),
		historiaParaPriorizar(3, backlog.PrioridadMedia, 3),
	}

	ordenes, err := backlog.Mover(lista, 1, backlog.Bajar)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if want := map[int64]int{2: 1, 1: 2, 3: 3}; !maps.Equal(ordenes, want) {
		t.Errorf("órdenes = %v, se esperaba %v", ordenes, want)
	}
}

// RN3 · la primera de su prioridad no sube y la última no baja, aunque haya otra prioridad al lado; también la única
func TestMover_LaPrimeraNoSubeYLaUltimaNoBaja(t *testing.T) {
	casos := []struct {
		nombre     string
		numero     int
		movimiento backlog.Movimiento
		base       error
		mensaje    string
	}{
		{"la primera de Alta no sube", 2, backlog.Subir, backlog.ErrYaEsLaPrimera, "HU-2 ya es la primera de su prioridad"},
		{"la última de Alta no baja a Media", 4, backlog.Bajar, backlog.ErrYaEsLaUltima, "HU-4 ya es la última de su prioridad"},
		{"la única de Media no sube", 3, backlog.Subir, backlog.ErrYaEsLaPrimera, "HU-3 ya es la primera de su prioridad"},
		{"la única de Media no baja", 3, backlog.Bajar, backlog.ErrYaEsLaUltima, "HU-3 ya es la última de su prioridad"},
		{"la única de Baja no sube a Media", 1, backlog.Subir, backlog.ErrYaEsLaPrimera, "HU-1 ya es la primera de su prioridad"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			ordenes, err := backlog.Mover(backlogDelEscenario(), c.numero, c.movimiento)

			if !errors.Is(err, c.base) || err.Error() != c.mensaje {
				t.Errorf("error = %v, se esperaba %q (%v)", err, c.mensaje, c.base)
			}
			if ordenes != nil {
				t.Errorf("no se esperaban órdenes nuevos y se obtuvo %v", ordenes)
			}
		})
	}
}

// RN4 · una historia Hecha no se reordena
func TestMover_UnaHechaNoSeReordena(t *testing.T) {
	lista := backlogDelEscenario()
	lista[3] = hecha(lista[3], "2026-10-12")

	_, err := backlog.Mover(lista, 4, backlog.Subir)

	if !errors.Is(err, backlog.ErrHechaNoSeReordena) || err.Error() != "una historia Hecha no se reordena" {
		t.Errorf("error = %v, se esperaba ErrHechaNoSeReordena", err)
	}
}

// RN4 · las Hechas no cuentan para el orden: se intercambia con la pendiente de arriba aunque haya una Hecha en el medio
func TestMover_SaltaLasHechasDeSuPrioridad(t *testing.T) {
	lista := []backlog.Historia{
		historiaParaPriorizar(1, backlog.PrioridadAlta, 1),
		hecha(historiaParaPriorizar(2, backlog.PrioridadAlta, 2), "2026-10-12"),
		historiaParaPriorizar(3, backlog.PrioridadAlta, 3),
	}

	ordenes, err := backlog.Mover(lista, 3, backlog.Subir)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if want := map[int64]int{3: 1, 1: 2}; !maps.Equal(ordenes, want) {
		t.Errorf("órdenes = %v, se esperaba %v (la Hecha no cambia)", ordenes, want)
	}
}

// RN8 · una historia En sprint o En progreso se reordena igual
func TestMover_UnaHistoriaDeUnSprintSeReordena(t *testing.T) {
	lista := backlogDelEscenario()
	lista[1].Estado = backlog.EstadoEnSprint
	lista[3].Estado = backlog.EstadoEnProgreso

	ordenes, err := backlog.Mover(lista, 2, backlog.Bajar)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if want := map[int64]int{4: 1, 2: 2}; !maps.Equal(ordenes, want) {
		t.Errorf("órdenes = %v, se esperaba %v", ordenes, want)
	}
}

// Sección 7 · una historia que no existe en el proyecto
func TestMover_UnaHistoriaQueNoExiste(t *testing.T) {
	_, err := backlog.Mover(backlogDelEscenario(), 99, backlog.Subir)

	if !errors.Is(err, backlog.ErrHistoriaNoEncontrada) || err.Error() != "no existe la historia HU-99" {
		t.Errorf("error = %v, se esperaba %q", err, "no existe la historia HU-99")
	}
}

// Sección 2 · el movimiento es subir o bajar
func TestMover_UnMovimientoQueNoEsSubirNiBajar(t *testing.T) {
	_, err := backlog.Mover(backlogDelEscenario(), 4, backlog.Movimiento("arriba"))

	if !errors.Is(err, backlog.ErrMovimientoInvalido) {
		t.Errorf("error = %v, se esperaba ErrMovimientoInvalido", err)
	}
}

// RN6 · una historia nueva va al final de su prioridad: su orden es el más alto del proyecto más uno
func TestNueva_VaAlFinalDeSuPrioridad(t *testing.T) {
	existentes := []backlog.Historia{
		historiaParaPriorizar(1, backlog.PrioridadBaja, 7),
		historiaParaPriorizar(2, backlog.PrioridadAlta, 2),
	}

	h, err := backlog.Nueva(datosValidos(), 7, existentes)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if h.Orden != 8 {
		t.Errorf("orden = %d, se esperaba 8", h.Orden)
	}
	pendientes, _ := backlog.Ordenar(append(existentes, h))
	if got := numeros(pendientes); !slices.Equal(got, []int{2, 3, 1}) {
		t.Errorf("backlog = HU %v, se esperaba la nueva al final de Alta", got)
	}
}
