package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
)

// RepositorioSesionesPoker es lo que HU-17 necesita del repositorio de Planning Poker. Lo implementan
// internal/store/memory e internal/store/postgres (tipo Poker). HU-18, HU-19 y HU-22 agregan cada una su interfaz en su
// archivo (spec de HU-17, "Modelo común de Planning Poker").
type RepositorioSesionesPoker interface {
	// Abrir guarda una sesión Abierta de la historia con su Ronda 1, en una transacción. Si ya hay una abierta devuelve
	// poker.ErrSesionYaAbierta.
	Abrir(ctx context.Context, historiaID int64) (poker.Sesion, poker.Ronda, error)
	// SesionAbierta devuelve la sesión abierta de la historia y su ronda actual (la de número más alto), o
	// poker.ErrSesionNoEncontrada.
	SesionAbierta(ctx context.Context, historiaID int64) (poker.Sesion, poker.Ronda, error)
}

// Poker agrupa los casos de uso de Planning Poker (épica E4). HU-17 abre la sesión y la muestra; HU-18, HU-19 y HU-22
// le agregan sus métodos en sus archivos (poker_votar.go, poker_revelar.go y poker_acordar.go).
type Poker struct {
	proyectos   BuscadorProyectos
	historias   BuscadorHistorias
	integrantes ListadorIntegrantes
	sesiones    RepositorioSesionesPoker
	votos       RepositorioVotosPoker
}

// NuevoPoker arma el caso de uso.
func NuevoPoker(proyectos BuscadorProyectos, historias BuscadorHistorias, integrantes ListadorIntegrantes, sesiones RepositorioSesionesPoker) *Poker {
	return &Poker{proyectos: proyectos, historias: historias, integrantes: integrantes, sesiones: sesiones}
}

// VistaSesion es lo que muestra la página de la sesión de una historia. Si la historia no tiene una sesión abierta,
// Abierta es false y Sesion y Ronda quedan vacías. Participantes son los integrantes activos del proyecto (CA-17.3).
// Votaron es el conjunto de IDs de los participantes que votaron en la ronda actual (HU-18, CA-18.2), nunca los
// valores (RN5); vacío si no hay sesión abierta.
type VistaSesion struct {
	Proyecto      project.Proyecto
	Historia      backlog.Historia
	Abierta       bool
	Sesion        poker.Sesion
	Ronda         poker.Ronda
	Participantes []member.Integrante
	Votaron       map[int64]bool
}

// Ver devuelve la historia HU-N del proyecto con su sesión abierta, si tiene, y los participantes de ese momento
// (RN5). Si el proyecto o la historia no existen, el error es project.ErrNoEncontrado o backlog.ErrHistoriaNoEncontrada.
func (s *Poker) Ver(ctx context.Context, proyectoID int64, numero int) (VistaSesion, error) {
	v, _, err := s.historiaConIntegrantes(ctx, proyectoID, numero)
	if err != nil {
		return VistaSesion{}, err
	}
	v.Sesion, v.Ronda, err = s.sesiones.SesionAbierta(ctx, v.Historia.ID)
	switch {
	case errors.Is(err, poker.ErrSesionNoEncontrada):
		v.Sesion, v.Ronda = poker.Sesion{}, poker.Ronda{}
	case err != nil:
		return VistaSesion{}, fmt.Errorf("buscar la sesión de la historia %d: %w", v.Historia.ID, err)
	default:
		v.Abierta = true
	}
	return s.conQuienesVotaron(ctx, v)
}

// Abrir abre una sesión de Planning Poker para la historia HU-N del proyecto, con la Ronda 1, y devuelve la vista de la
// sesión. quienID es el integrante que la abre: tiene que ser el Agile Enabler activo del proyecto (RN4); 0 o el ID de
// alguien de otro proyecto es como no haber elegido a nadie. Los errores son los de poker.PuedeAbrir,
// poker.ErrSesionYaAbierta (CA-17.2), project.ErrNoEncontrado o backlog.ErrHistoriaNoEncontrada; si no se puede abrir,
// no se guarda nada.
func (s *Poker) Abrir(ctx context.Context, proyectoID int64, numero int, quienID int64) (VistaSesion, error) {
	v, integrantes, err := s.historiaConIntegrantes(ctx, proyectoID, numero)
	if err != nil {
		return VistaSesion{}, err
	}
	if err := poker.PuedeAbrir(v.Historia, integranteConID(integrantes, quienID)); err != nil {
		return VistaSesion{}, err
	}
	v.Sesion, v.Ronda, err = s.sesiones.Abrir(ctx, v.Historia.ID)
	switch {
	case errors.Is(err, poker.ErrSesionYaAbierta):
		return VistaSesion{}, err
	case err != nil:
		return VistaSesion{}, fmt.Errorf("abrir la sesión de la historia %d: %w", v.Historia.ID, err)
	}
	v.Abierta = true
	return v, nil
}

// historiaConIntegrantes busca el proyecto, la historia y los integrantes del proyecto, y arma la vista sin sesión con
// los participantes. También devuelve todos los integrantes, para buscar quién abre.
func (s *Poker) historiaConIntegrantes(ctx context.Context, proyectoID int64, numero int) (VistaSesion, []member.Integrante, error) {
	p, err := s.proyectos.BuscarPorID(ctx, proyectoID)
	if err != nil {
		return VistaSesion{}, nil, fmt.Errorf("buscar el proyecto %d: %w", proyectoID, err)
	}
	h, err := s.historias.BuscarPorNumero(ctx, proyectoID, numero)
	if err != nil {
		return VistaSesion{}, nil, fmt.Errorf("buscar la historia HU-%d del proyecto %d: %w", numero, proyectoID, err)
	}
	integrantes, err := s.integrantes.ListarPorProyecto(ctx, proyectoID)
	if err != nil {
		return VistaSesion{}, nil, fmt.Errorf("listar los integrantes del proyecto %d: %w", proyectoID, err)
	}
	return VistaSesion{Proyecto: p, Historia: h, Participantes: poker.Participantes(integrantes), Votaron: map[int64]bool{}}, integrantes, nil
}

// integranteConID busca el integrante entre los del proyecto. Si no está (0 o alguien de otro proyecto) devuelve un
// integrante vacío, que es como no haber elegido a nadie.
func integranteConID(integrantes []member.Integrante, id int64) member.Integrante {
	for _, i := range integrantes {
		if i.ID == id {
			return i
		}
	}
	return member.Integrante{}
}
