package backlog

import (
	"cmp"
	"errors"
	"slices"
)

// Movimiento es hacia dónde se mueve una historia dentro de su prioridad (HU-09). Los valores son los de la URL.
//
// Spec: specs/HU-09-priorizar-backlog.md.
type Movimiento string

// Los dos movimientos: un lugar hacia arriba o hacia abajo (RN2).
const (
	Subir Movimiento = "subir"
	Bajar Movimiento = "bajar"
)

// Errores de un movimiento (sección 7 de la spec de HU-09). ErrYaEsLaPrimera y ErrYaEsLaUltima se devuelven con un
// mensaje que nombra la historia.
var (
	ErrYaEsLaPrimera      = errors.New("la historia ya es la primera de su prioridad")
	ErrYaEsLaUltima       = errors.New("la historia ya es la última de su prioridad")
	ErrHechaNoSeReordena  = errors.New("una historia Hecha no se reordena")
	ErrMovimientoInvalido = errors.New("el movimiento tiene que ser subir o bajar")
)

// rangoPrioridad da la posición de cada prioridad en el backlog: Alta primero (RN1).
var rangoPrioridad = map[Prioridad]int{PrioridadAlta: 0, PrioridadMedia: 1, PrioridadBaja: 2}

// Ordenar separa el backlog en las historias que no están Hechas y las Hechas, cada parte en una lista nueva (la que
// recibe no cambia). Las pendientes van por prioridad (Alta, Media, Baja), después por el orden manual y después por
// número (RN1); las Hechas, la completada más recientemente primero y, con la misma fecha, por número (RN4).
func Ordenar(historias []Historia) (pendientes, hechas []Historia) {
	for _, h := range historias {
		if h.Estado == EstadoHecho {
			hechas = append(hechas, h)
		} else {
			pendientes = append(pendientes, h)
		}
	}
	slices.SortFunc(pendientes, func(a, b Historia) int {
		return cmp.Or(cmp.Compare(rangoPrioridad[a.Prioridad], rangoPrioridad[b.Prioridad]),
			cmp.Compare(a.Orden, b.Orden), cmp.Compare(a.Numero, b.Numero))
	})
	slices.SortFunc(hechas, func(a, b Historia) int {
		return cmp.Or(compararFechaDescendente(a, b), cmp.Compare(a.Numero, b.Numero))
	})
	return pendientes, hechas
}

// compararFechaDescendente pone primero la completada más recientemente; una sin fecha va al final.
func compararFechaDescendente(a, b Historia) int {
	switch {
	case a.CompletadaEn == nil && b.CompletadaEn == nil:
		return 0
	case a.CompletadaEn == nil:
		return 1
	case b.CompletadaEn == nil:
		return -1
	}
	return b.CompletadaEn.Compare(*a.CompletadaEn)
}

// Mover sube o baja la historia HU-numero un lugar dentro de su prioridad, intercambiándola con la vecina que no está
// Hecha (RN2), y devuelve el orden nuevo de cada historia de esa prioridad que no está Hecha, por ID: 1, 2, 3… (RN5).
// Las demás historias no cambian. Si no se puede mover devuelve el error y ningún orden: la primera no sube y la
// última no baja (RN3), una Hecha no se reordena (RN4) y una que no está en la lista es ErrHistoriaNoEncontrada.
func Mover(historias []Historia, numero int, m Movimiento) (map[int64]int, error) {
	if m != Subir && m != Bajar {
		return nil, ErrMovimientoInvalido
	}
	i := slices.IndexFunc(historias, func(h Historia) bool { return h.Numero == numero })
	if i < 0 {
		return nil, errorDeEstado{base: ErrHistoriaNoEncontrada, mensaje: "no existe la historia " + Historia{Numero: numero}.Identificador()}
	}
	elegida := historias[i]
	if elegida.Estado == EstadoHecho {
		return nil, ErrHechaNoSeReordena
	}

	pendientes, _ := Ordenar(historias)
	grupo := slices.DeleteFunc(pendientes, func(h Historia) bool { return h.Prioridad != elegida.Prioridad })
	pos := slices.IndexFunc(grupo, func(h Historia) bool { return h.ID == elegida.ID })
	switch {
	case m == Subir && pos == 0:
		return nil, errorDeEstado{base: ErrYaEsLaPrimera, mensaje: elegida.Identificador() + " ya es la primera de su prioridad"}
	case m == Bajar && pos == len(grupo)-1:
		return nil, errorDeEstado{base: ErrYaEsLaUltima, mensaje: elegida.Identificador() + " ya es la última de su prioridad"}
	case m == Subir:
		grupo[pos-1], grupo[pos] = grupo[pos], grupo[pos-1]
	default:
		grupo[pos], grupo[pos+1] = grupo[pos+1], grupo[pos]
	}

	ordenes := make(map[int64]int, len(grupo))
	for n, h := range grupo {
		ordenes[h.ID] = n + 1
	}
	return ordenes, nil
}

// siguienteOrden devuelve el orden más alto de las historias del proyecto más uno, para que una historia nueva quede
// al final de su prioridad, aunque ya se haya reordenado a mano (RN6).
func siguienteOrden(existentes []Historia) int {
	mayor := 0
	for _, h := range existentes {
		mayor = max(mayor, h.Orden)
	}
	return mayor + 1
}
