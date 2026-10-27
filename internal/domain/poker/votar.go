package poker

import (
	"errors"
	"slices"
	"strconv"
	"strings"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
)

// Errores de votar (sección 7 de la spec de HU-18).
var (
	ErrVotoFueraDeEscala = errors.New("el voto tiene que ser un valor de la escala: 0, 1, 2, 3, 5, 8, 13 o 21")
	ErrSoloParticipantes = errors.New("solo votan los integrantes activos del proyecto")
	ErrRondaRevelada     = errors.New("la ronda ya se reveló: no se puede votar")
)

// ValidarVoto convierte el valor que llega del formulario: tiene que ser un número entero de la escala de story points
// (CA-18.4, RN4). A diferencia de la estimación de HU-16, vacío también es un error.
func ValidarVoto(texto string) (int, error) {
	texto = strings.TrimSpace(texto)
	valor, err := strconv.Atoi(texto)
	// Atoi acepta un signo adelante ("+5"); acá solo valen dígitos.
	if err != nil || texto[0] < '0' || texto[0] > '9' || !slices.Contains(backlog.EscalaStoryPoints, valor) {
		return 0, ErrVotoFueraDeEscala
	}
	return valor, nil
}

// PuedeVotar devuelve nil si quien puede votar en la ronda: es un integrante activo del proyecto, también el Agile
// Enabler (RN2), y la ronda no se reveló (RN3). Quién vota se revisa primero.
func PuedeVotar(quien member.Integrante, r Ronda) error {
	if quien.ID == 0 || !quien.Activo {
		return ErrSoloParticipantes
	}
	if r.Revelada {
		return ErrRondaRevelada
	}
	return nil
}

// Votaron devuelve el conjunto de IDs de los participantes que votaron, a partir de los IDs de quienes tienen un voto
// en la ronda. El voto de quien ya no participa (se dio de baja) no cuenta (RN2). Nunca tiene valores (RN5).
func Votaron(participantes []member.Integrante, votantes []int64) map[int64]bool {
	votaron := map[int64]bool{}
	for _, p := range participantes {
		if slices.Contains(votantes, p.ID) {
			votaron[p.ID] = true
		}
	}
	return votaron
}
