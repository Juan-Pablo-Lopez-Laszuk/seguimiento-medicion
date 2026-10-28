package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
)

// RepositorioRevelarPoker es lo que HU-19 necesita del repositorio de Planning Poker. Lo implementa el tipo Poker de
// internal/store/memory e internal/store/postgres.
type RepositorioRevelarPoker interface {
	// Votos devuelve los votos de la ronda (los guarda HU-18).
	Votos(ctx context.Context, rondaID int64) ([]poker.Voto, error)
	// Revelar marca la ronda como revelada, o devuelve poker.ErrSesionNoEncontrada si no existe.
	Revelar(ctx context.Context, rondaID int64) error
}

// RevelarPoker son los casos de uso de HU-19: revelar la ronda sola cuando votaron todos y que el Agile Enabler la
// cierre antes. Usa la sesión de HU-17 (Poker) sin cambiarla.
type RevelarPoker struct {
	poker  *Poker
	rondas RepositorioRevelarPoker
}

// NuevoRevelarPoker arma el caso de uso.
func NuevoRevelarPoker(p *Poker, rondas RepositorioRevelarPoker) *RevelarPoker {
	return &RevelarPoker{poker: p, rondas: rondas}
}

// VistaRevelar es la sesión de HU-17 con los votos de la ronda. Votos queda vacío mientras la ronda no está revelada
// (RN6). ReveladaAhora es true si la ronda se reveló en este pedido, para mostrar "Ronda revelada.".
type VistaRevelar struct {
	VistaSesion
	Votos         []poker.VotoRevelado
	ReveladaAhora bool
}

// Revisar devuelve la sesión de la historia HU-N con sus votos y, si votaron todos los integrantes activos, revela la
// ronda (CA-19.1, RN1). Se llama al ver la sesión y después de cada voto. Si la historia no tiene sesión abierta,
// Abierta es false.
func (s *RevelarPoker) Revisar(ctx context.Context, proyectoID int64, numero int) (VistaRevelar, error) {
	v, integrantes, votos, err := s.sesion(ctx, proyectoID, numero)
	if err != nil || !v.Abierta {
		return v, err
	}
	if !v.Ronda.Revelada {
		if r := poker.RevelarSiCorresponde(v.Ronda, votos, integrantes); r.Revelada {
			if err := s.revelar(ctx, &v, r); err != nil {
				return VistaRevelar{}, err
			}
		}
	}
	return conVotos(v, votos, integrantes)
}

// Cerrar revela la ronda actual de la historia HU-N con los votos que haya (CA-19.1, RN2). quienID tiene que ser el
// Agile Enabler activo del proyecto (RN4). Los errores son poker.ErrSesionNoEncontrada y los de poker.Cerrar; si hay
// error, no se guarda nada.
func (s *RevelarPoker) Cerrar(ctx context.Context, proyectoID int64, numero int, quienID int64) (VistaRevelar, error) {
	v, integrantes, votos, err := s.sesion(ctx, proyectoID, numero)
	if err != nil {
		return VistaRevelar{}, err
	}
	if !v.Abierta {
		return VistaRevelar{}, poker.ErrSesionNoEncontrada
	}
	var quien member.Integrante
	for _, i := range integrantes {
		if i.ID == quienID {
			quien = i
		}
	}
	r, err := poker.Cerrar(v.Ronda, votos, quien)
	if err != nil {
		return VistaRevelar{}, err
	}
	if err := s.revelar(ctx, &v, r); err != nil {
		return VistaRevelar{}, err
	}
	return conVotos(v, votos, integrantes)
}

// sesion busca la historia con su sesión abierta, todos los integrantes del proyecto y los votos que cuentan de la
// ronda actual.
func (s *RevelarPoker) sesion(ctx context.Context, proyectoID int64, numero int) (VistaRevelar, []member.Integrante, []poker.Voto, error) {
	vs, integrantes, err := s.poker.historiaConIntegrantes(ctx, proyectoID, numero)
	if err != nil {
		return VistaRevelar{}, nil, nil, err
	}
	v := VistaRevelar{VistaSesion: vs}
	v.Sesion, v.Ronda, err = s.poker.sesiones.SesionAbierta(ctx, v.Historia.ID)
	switch {
	case errors.Is(err, poker.ErrSesionNoEncontrada):
		v.Sesion, v.Ronda = poker.Sesion{}, poker.Ronda{}
		return v, integrantes, nil, nil
	case err != nil:
		return VistaRevelar{}, nil, nil, fmt.Errorf("buscar la sesión de la historia %d: %w", v.Historia.ID, err)
	}
	v.Abierta = true
	votos, err := s.rondas.Votos(ctx, v.Ronda.ID)
	if err != nil {
		return VistaRevelar{}, nil, nil, fmt.Errorf("listar los votos de la ronda %d: %w", v.Ronda.ID, err)
	}
	return v, integrantes, votosQueCuentan(votos, v.Participantes), nil
}

// votosQueCuentan deja solo los votos de los participantes, los integrantes activos: el de alguien que se dio de baja
// no cuenta para revelar ni para cerrar (RN7, como RN2 de HU-18).
func votosQueCuentan(votos []poker.Voto, participantes []member.Integrante) []poker.Voto {
	activos := map[int64]bool{}
	for _, p := range participantes {
		activos[p.ID] = true
	}
	var cuentan []poker.Voto
	for _, v := range votos {
		if activos[v.IntegranteID] {
			cuentan = append(cuentan, v)
		}
	}
	return cuentan
}

// revelar guarda la ronda revelada y la deja en la vista.
func (s *RevelarPoker) revelar(ctx context.Context, v *VistaRevelar, r poker.Ronda) error {
	if err := s.rondas.Revelar(ctx, r.ID); err != nil {
		return fmt.Errorf("revelar la ronda %d: %w", r.ID, err)
	}
	v.Ronda, v.ReveladaAhora = r, true
	return nil
}

// conVotos agrega a la vista los votos con su autor si la ronda está revelada (CA-19.2, RN6).
func conVotos(v VistaRevelar, votos []poker.Voto, integrantes []member.Integrante) (VistaRevelar, error) {
	if !v.Ronda.Revelada {
		return v, nil
	}
	filas, err := poker.VotosRevelados(v.Ronda, votos, integrantes)
	if err != nil {
		return VistaRevelar{}, err
	}
	v.Votos = filas
	return v, nil
}
