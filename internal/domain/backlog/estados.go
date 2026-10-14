package backlog

import (
	"errors"
	"fmt"
	"slices"
	"strings"
)

// Los otros dos estados de una historia (EstadoPendiente y EstadoHecho están en backlog.go). Los textos son los que
// guarda la base (CHECK de historias.estado).
//
// Spec: specs/HU-08-estado-historia.md.
const (
	EstadoEnSprint   Estado = "En sprint"
	EstadoEnProgreso Estado = "En progreso"
)

// Estados son todos los estados de una historia, en el orden del camino: Pendiente, En sprint, En progreso, Hecho (RN1).
var Estados = []Estado{EstadoPendiente, EstadoEnSprint, EstadoEnProgreso, EstadoHecho}

// transiciones dice a qué estados puede pasar una historia desde cada uno (RN2). Hecho no pasa a ninguno (RN3).
var transiciones = map[Estado][]Estado{
	EstadoPendiente:  {EstadoEnSprint},
	EstadoEnSprint:   {EstadoEnProgreso, EstadoPendiente},
	EstadoEnProgreso: {EstadoHecho, EstadoPendiente},
}

// Errores de un cambio de estado (sección 7 de la spec de HU-08). Todos se devuelven con un mensaje descriptivo (RN6).
var (
	// ErrTransicionInvalida se devuelve al querer pasar a un estado que no está en la tabla de transiciones (RN2 y RN3).
	ErrTransicionInvalida = errors.New("transición de estado inválida")
	// ErrYaEstaEnEseEstado se devuelve al pedir el estado en el que ya está la historia (RN5).
	ErrYaEstaEnEseEstado = errors.New("la historia ya está en ese estado")
	// ErrEstadoDesconocido se devuelve al pedir un estado que no existe (RN1).
	ErrEstadoDesconocido = errors.New("el estado no existe")
	// ErrCriteriosSinCumplir se devuelve al querer pasar a Hecho sin criterios o con criterios sin cumplir (RN4).
	ErrCriteriosSinCumplir = errors.New("los criterios de aceptación no están todos cumplidos")
)

// errorDeEstado es un error con un mensaje descriptivo para el usuario (RN6) que sigue siendo reconocible con
// errors.Is contra su error base.
type errorDeEstado struct {
	base    error
	mensaje string
}

func (e errorDeEstado) Error() string { return e.mensaje }
func (e errorDeEstado) Unwrap() error { return e.base }

// CambiarEstado devuelve la historia con el estado nuevo, si el cambio vale; la original no cambia. Es la única
// regla de transición: HU-12 la usa para pasar a En sprint, HU-13 para pasar a Hecho y HU-14 para volver a Pendiente.
// progreso son los criterios de la historia (HU-07), que hacen falta para pasar a Hecho.
func (h Historia) CambiarEstado(nuevo Estado, progreso Progreso) (Historia, error) {
	if !slices.Contains(Estados, nuevo) {
		return Historia{}, errorDeEstado{base: ErrEstadoDesconocido, mensaje: mensajeDeEstadoDesconocido(nuevo)}
	}
	if nuevo == h.Estado {
		return Historia{}, errorDeEstado{base: ErrYaEstaEnEseEstado, mensaje: "la historia ya está " + string(nuevo)}
	}
	destinos := transiciones[h.Estado]
	if !slices.Contains(destinos, nuevo) {
		return Historia{}, errorDeEstado{base: ErrTransicionInvalida, mensaje: mensajeDeTransicion(h.Estado, nuevo, destinos)}
	}
	if nuevo == EstadoHecho && !progreso.Completo() {
		return Historia{}, errorDeEstado{base: ErrCriteriosSinCumplir, mensaje: mensajeDeCriteriosSinCumplir(progreso)}
	}
	h.Estado = nuevo
	return h, nil
}

// mensajeDeTransicion explica por qué no se puede pasar de un estado a otro y a cuáles sí (RN6).
func mensajeDeTransicion(desde, hacia Estado, destinos []Estado) string {
	if len(destinos) == 0 {
		return fmt.Sprintf("no se puede pasar de %s a %s: una historia Hecha no cambia de estado", desde, hacia)
	}
	nombres := make([]string, len(destinos))
	for i, d := range destinos {
		nombres[i] = string(d)
	}
	return fmt.Sprintf("no se puede pasar de %s a %s: desde %s solo se puede pasar a %s", desde, hacia, desde, strings.Join(nombres, " o "))
}

// mensajeDeEstadoDesconocido dice cuál estado no existe y cuáles sí.
func mensajeDeEstadoDesconocido(estado Estado) string {
	return fmt.Sprintf("el estado %s no existe: los estados son Pendiente, En sprint, En progreso y Hecho", estado)
}

// mensajeDeCriteriosSinCumplir explica por qué no pasa a Hecho: no tiene criterios, o cuántos tiene cumplidos.
func mensajeDeCriteriosSinCumplir(p Progreso) string {
	if !p.TieneCriterios() {
		return "no se puede pasar a Hecho: la historia no tiene criterios de aceptación"
	}
	return fmt.Sprintf("no se puede pasar a Hecho: la historia tiene %d de %d criterios cumplidos y hacen falta todos", p.Cumplidos, p.Total)
}
