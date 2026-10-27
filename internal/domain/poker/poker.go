// Package poker tiene las entidades de Planning Poker y sus reglas de negocio (épica E4): la sesión de una historia,
// sus rondas y los votos de cada integrante.
//
// Spec: specs/HU-17-abrir-sesion-poker.md (sección "Modelo común de Planning Poker"). No usa base de datos ni HTTP.
// Cada historia de la épica agrega sus reglas en su propio archivo: abrir.go (HU-17), votar.go (HU-18), revelar.go
// (HU-19) y acordar.go (HU-22).
package poker

import "errors"

// EstadoSesion es el estado de una sesión de Planning Poker. Los valores son los mismos que guarda la base.
type EstadoSesion string

// Estados de una sesión: se abre en HU-17 y se cierra al registrar el valor acordado en HU-22.
const (
	EstadoAbierta EstadoSesion = "Abierta"
	EstadoCerrada EstadoSesion = "Cerrada"
)

// Sesion es una sesión de Planning Poker para estimar una historia (tabla sesiones_poker). Una historia tiene una sola
// sesión abierta a la vez (CA-17.2).
type Sesion struct {
	ID         int64
	HistoriaID int64
	Estado     EstadoSesion
	// ValorAcordado es la estimación que registra el Agile Enabler al cerrar la sesión (HU-22); nil mientras está
	// abierta.
	ValorAcordado *int
}

// Ronda es una vuelta de votación de una sesión (tabla rondas_poker). Se numeran desde 1 dentro de la sesión; la
// Ronda 1 se crea al abrir la sesión (HU-17).
type Ronda struct {
	ID       int64
	SesionID int64
	Numero   int
	// Revelada es true cuando la ronda se reveló (HU-19): recién ahí se pueden ver los valores de los votos.
	Revelada bool
}

// Voto es la estimación de un integrante en una ronda (tabla votos): uno por integrante y por ronda; revotar lo
// reemplaza (CA-18.1). Valor es de la escala de story points (backlog.EscalaStoryPoints). No se muestra ni se devuelve
// antes de revelar la ronda (CA-18.2 y CA-18.3).
type Voto struct {
	ID           int64
	RondaID      int64
	IntegranteID int64
	Valor        int
}

// Errores comunes de la épica. Los de cada historia van en su archivo.
var (
	// ErrSesionNoEncontrada lo devuelven los repositorios cuando la historia no tiene una sesión abierta.
	ErrSesionNoEncontrada = errors.New("esta historia no tiene una sesión de Planning Poker abierta")
	// ErrSesionCerrada es el error de querer votar, revelar o acordar en una sesión que ya se cerró (HU-22).
	ErrSesionCerrada = errors.New("la sesión de Planning Poker ya está cerrada")
)
