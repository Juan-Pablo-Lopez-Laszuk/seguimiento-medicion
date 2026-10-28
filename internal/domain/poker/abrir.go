package poker

import (
	"errors"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
)

// Errores de abrir una sesión (sección 7 de la spec de HU-17).
var (
	ErrHistoriaHecha    = errors.New("la historia está Hecha: ya no se estima")
	ErrSesionYaAbierta  = errors.New("la historia ya tiene una sesión de Planning Poker abierta")
	ErrSoloAgileEnabler = errors.New("solo el Agile Enabler del proyecto abre una sesión de Planning Poker")
)

// PuedeConducir devuelve nil si quien es el Agile Enabler activo del proyecto, el que conduce la sesión: la abre
// (HU-17), cierra la ronda (HU-19) y registra el valor acordado (HU-22). Si no, devuelve ErrSoloAgileEnabler (RN4).
func PuedeConducir(quien member.Integrante) error {
	if quien.ID == 0 || !quien.Activo || quien.Rol != member.RolAgileEnabler {
		return ErrSoloAgileEnabler
	}
	return nil
}

// PuedeAbrir devuelve nil si quien puede abrir una sesión para la historia: la historia no está Hecha (CA-17.1, RN1)
// y quien la abre es el Agile Enabler activo (RN4). La historia se revisa primero. Que no haya otra sesión abierta
// (CA-17.2) lo controla el repositorio al guardar.
func PuedeAbrir(h backlog.Historia, quien member.Integrante) error {
	if h.Estado == backlog.EstadoHecho {
		return ErrHistoriaHecha
	}
	return PuedeConducir(quien)
}

// Participantes devuelve los integrantes activos, en el orden en que se registraron (CA-17.3, RN5).
func Participantes(integrantes []member.Integrante) []member.Integrante {
	activos := []member.Integrante{}
	for _, i := range integrantes {
		if i.Activo {
			activos = append(activos, i)
		}
	}
	return activos
}
