package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
)

// RepositorioAcordarPoker es lo que HU-22 necesita del repositorio de Planning Poker: la sesión abierta con su ronda
// actual (para saber si está revelada), cerrarla con el valor acordado y listar las cerradas. Lo implementan
// internal/store/memory e internal/store/postgres (tipo Poker).
type RepositorioAcordarPoker interface {
	// SesionAbierta devuelve la sesión abierta de la historia y su ronda actual, o poker.ErrSesionNoEncontrada.
	SesionAbierta(ctx context.Context, historiaID int64) (poker.Sesion, poker.Ronda, error)
	// Cerrar deja la sesión Cerrada con el valor acordado y guarda los story points de la historia (con su entrada del
	// historial de HU-16b si cambian), en una transacción. Si la sesión ya no está abierta devuelve
	// poker.ErrSesionCerrada.
	Cerrar(ctx context.Context, sesionID, historiaID int64, valor int) error
	// SesionesCerradas devuelve las sesiones cerradas de la historia, la más reciente primero, con su cantidad de rondas.
	SesionesCerradas(ctx context.Context, historiaID int64) ([]poker.SesionCerrada, error)
}

// ConAcordar conecta el repositorio de HU-22: registrar la estimación acordada y consultar las sesiones cerradas.
// Devuelve el mismo caso de uso para armarlo en una línea.
func (s *Poker) ConAcordar(repo RepositorioAcordarPoker) *Poker {
	s.acordar = repo
	return s
}

// Acordar registra el valor acordado de la sesión abierta de la historia HU-N del proyecto (HU-22): la sesión queda
// Cerrada con ese valor (CA-22.3) y la historia con esos story points (CA-22.2), en una sola operación. quienID es el
// integrante que lo registra; 0 o el ID de alguien de otro proyecto es como no haber elegido a nadie. Devuelve la
// historia estimada. Los errores son los de poker.Acordar (domain.ErroresValidacion si el valor no es de la escala,
// CA-22.1), poker.ErrSesionNoEncontrada si no hay una sesión abierta, poker.ErrSesionCerrada si se cerró mientras
// tanto, project.ErrNoEncontrado o backlog.ErrHistoriaNoEncontrada; si no se puede registrar, no se guarda nada.
func (s *Poker) Acordar(ctx context.Context, proyectoID int64, numero int, quienID int64, valor string) (backlog.Historia, error) {
	if s.acordar == nil {
		return backlog.Historia{}, errors.New("el caso de uso Poker no tiene conectado el repositorio de HU-22 (ConAcordar)")
	}
	v, integrantes, err := s.historiaConIntegrantes(ctx, proyectoID, numero)
	if err != nil {
		return backlog.Historia{}, err
	}
	var quien member.Integrante
	for _, i := range integrantes {
		if i.ID == quienID {
			quien = i
		}
	}
	sesion, ronda, err := s.acordar.SesionAbierta(ctx, v.Historia.ID)
	switch {
	case errors.Is(err, poker.ErrSesionNoEncontrada):
		return backlog.Historia{}, err
	case err != nil:
		return backlog.Historia{}, fmt.Errorf("buscar la sesión de la historia %d: %w", v.Historia.ID, err)
	}
	estimada, err := poker.Acordar(v.Historia, sesion, ronda, quien, valor)
	if err != nil {
		return backlog.Historia{}, err
	}
	err = s.acordar.Cerrar(ctx, sesion.ID, v.Historia.ID, *estimada.StoryPoints)
	switch {
	case errors.Is(err, poker.ErrSesionCerrada):
		return backlog.Historia{}, err
	case err != nil:
		return backlog.Historia{}, fmt.Errorf("cerrar la sesión %d de la historia %d: %w", sesion.ID, v.Historia.ID, err)
	}
	return estimada, nil
}

// sesionesCerradas devuelve las sesiones cerradas de la historia para la vista (CA-22.3), o ninguna si el repositorio
// de HU-22 no está conectado.
func (s *Poker) sesionesCerradas(ctx context.Context, historiaID int64) ([]poker.SesionCerrada, error) {
	if s.acordar == nil {
		return nil, nil
	}
	cerradas, err := s.acordar.SesionesCerradas(ctx, historiaID)
	if err != nil {
		return nil, fmt.Errorf("listar las sesiones cerradas de la historia %d: %w", historiaID, err)
	}
	return cerradas, nil
}
