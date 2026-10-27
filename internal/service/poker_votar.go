package service

import (
	"context"
	"fmt"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
)

// RepositorioVotosPoker es lo que HU-18 necesita del repositorio de Planning Poker. Lo implementan internal/store/memory
// e internal/store/postgres (tipo Poker). No devuelve los valores de los votos (CA-18.3, RN5): leerlos es de HU-19, con
// la ronda revelada.
type RepositorioVotosPoker interface {
	// Votar guarda el voto del integrante en la ronda, o lo reemplaza si ya votó. Si la ronda no existe, su sesión está
	// cerrada o ya se reveló, no guarda nada y devuelve poker.ErrSesionNoEncontrada, poker.ErrSesionCerrada o
	// poker.ErrRondaRevelada.
	Votar(ctx context.Context, rondaID, integranteID int64, valor int) error
	// QuienesVotaron devuelve los IDs de los integrantes que votaron en la ronda.
	QuienesVotaron(ctx context.Context, rondaID int64) ([]int64, error)
}

// ConVotos conecta los votos (HU-18): desde entonces se puede votar y la vista de la sesión dice quién votó. Se hace
// aparte de NuevoPoker, como Criterios.ConHistorial, para que HU-19 y HU-22 conecten lo suyo sin cambiarlo.
func (s *Poker) ConVotos(votos RepositorioVotosPoker) *Poker {
	s.votos = votos
	return s
}

// Votar guarda el voto de quienID en la ronda actual de la sesión abierta de la historia HU-N del proyecto, o lo
// reemplaza si ya votó (CA-18.1), y devuelve la vista de la sesión. valor es el texto del formulario. Se revisa en este
// orden: que haya una sesión abierta (poker.ErrSesionNoEncontrada), quién vota y que la ronda no esté revelada
// (poker.PuedeVotar) y el valor (poker.ValidarVoto). 0 o el ID de alguien de otro proyecto es como no haber elegido a
// nadie. Si no se puede votar, no se guarda nada.
func (s *Poker) Votar(ctx context.Context, proyectoID int64, numero int, quienID int64, valor string) (VistaSesion, error) {
	v, integrantes, err := s.historiaConIntegrantes(ctx, proyectoID, numero)
	if err != nil {
		return VistaSesion{}, err
	}
	v.Sesion, v.Ronda, err = s.sesiones.SesionAbierta(ctx, v.Historia.ID)
	if err != nil {
		return VistaSesion{}, fmt.Errorf("buscar la sesión de la historia %d: %w", v.Historia.ID, err)
	}
	v.Abierta = true
	if err := poker.PuedeVotar(integranteConID(integrantes, quienID), v.Ronda); err != nil {
		return VistaSesion{}, err
	}
	n, err := poker.ValidarVoto(valor)
	if err != nil {
		return VistaSesion{}, err
	}
	if err := s.votos.Votar(ctx, v.Ronda.ID, quienID, n); err != nil {
		return VistaSesion{}, fmt.Errorf("guardar el voto en la ronda %d: %w", v.Ronda.ID, err)
	}
	return s.conQuienesVotaron(ctx, v)
}

// conQuienesVotaron completa en la vista quiénes de los participantes votaron en la ronda actual (RN2). Sin sesión
// abierta o sin los votos conectados, no votó nadie.
func (s *Poker) conQuienesVotaron(ctx context.Context, v VistaSesion) (VistaSesion, error) {
	if !v.Abierta || s.votos == nil {
		return v, nil
	}
	ids, err := s.votos.QuienesVotaron(ctx, v.Ronda.ID)
	if err != nil {
		return VistaSesion{}, fmt.Errorf("listar quiénes votaron en la ronda %d: %w", v.Ronda.ID, err)
	}
	v.Votaron = poker.Votaron(v.Participantes, ids)
	return v, nil
}
