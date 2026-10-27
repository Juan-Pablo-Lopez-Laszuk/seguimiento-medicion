package poker

import (
	"errors"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
)

// Errores de registrar la estimación acordada (sección 7 de la spec de HU-22).
var (
	ErrRondaSinRevelar         = errors.New("para registrar la estimación, primero revelá la ronda")
	ErrSoloAgileEnablerAcuerda = errors.New("solo el Agile Enabler del proyecto registra la estimación acordada")
)

// SesionCerrada es una sesión ya cerrada de una historia, como se consulta después (CA-22.3). Numero es su lugar entre
// las sesiones de la historia, en el orden en que se abrieron ("Sesión 1", "Sesión 2"); la tabla no guarda fechas.
type SesionCerrada struct {
	Numero        int
	Rondas        int
	ValorAcordado int
}

// Acordar devuelve la historia con el valor acordado como story points (CA-22.2), si se puede registrar; la original
// no cambia. Las reglas se revisan en este orden: la sesión tiene que estar abierta (ErrSesionCerrada, RN5), quien
// registra tiene que ser el Agile Enabler activo (ErrSoloAgileEnablerAcuerda, RN3) y la ronda actual tiene que estar
// revelada (ErrRondaSinRevelar, RN4). El valor y la historia Hecha los decide backlog.Historia.Estimar, la misma regla
// que la estimación de HU-16 (CA-22.1 y RN7).
//
// Spec: specs/HU-22-registrar-estimacion.md.
func Acordar(h backlog.Historia, s Sesion, r Ronda, quien member.Integrante, valor string) (backlog.Historia, error) {
	if s.Estado != EstadoAbierta {
		return backlog.Historia{}, ErrSesionCerrada
	}
	if PuedeConducir(quien) != nil {
		return backlog.Historia{}, ErrSoloAgileEnablerAcuerda
	}
	if !r.Revelada {
		return backlog.Historia{}, ErrRondaSinRevelar
	}
	return h.Estimar(valor)
}
