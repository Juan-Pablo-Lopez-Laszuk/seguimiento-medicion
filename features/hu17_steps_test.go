package features

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// hu17 guarda el estado de un escenario de HU-17 (abrir una sesión de Planning Poker), con sus propios repositorios en
// memoria. Los integrantes se dan de baja y las sesiones se abren con los casos de uso, como en la aplicación.
type hu17 struct {
	repoProyectos   *memory.Proyectos
	repoHistorias   *memory.Historias
	repoIntegrantes *memory.Integrantes
	repoPoker       *memory.Poker
	integrantes     *service.Integrantes
	poker           *service.Poker
	proyectoID      int64
	porNombre       map[string]member.Integrante
	porTitulo       map[string]backlog.Historia
	err             error
}

// registrarPasosHU17 conecta cada frase de features/hu-17-abrir-sesion-poker.feature con una función de Go. Devuelve el
// estado del escenario para que las otras historias de Planning Poker (HU-22) usen los mismos repositorios.
func registrarPasosHU17(sc *godog.ScenarioContext) *hu17 {
	e := &hu17{
		repoProyectos: memory.NuevoProyectos(), repoHistorias: memory.NuevoHistorias(),
		repoIntegrantes: memory.NuevoIntegrantes(), repoPoker: memory.NuevoPoker(),
		porNombre: map[string]member.Integrante{}, porTitulo: map[string]backlog.Historia{},
	}
	e.integrantes = service.NuevoIntegrantes(e.repoProyectos, e.repoIntegrantes)
	e.poker = service.NuevoPoker(e.repoProyectos, e.repoHistorias, e.repoIntegrantes, e.repoPoker)

	sc.Step(`^el proyecto "([^"]*)" con estos integrantes para Planning Poker:$`, e.elProyectoConEstosIntegrantes)
	sc.Step(`^estas historias para estimar:$`, e.estasHistoriasParaEstimar)
	sc.Step(`^que "([^"]*)" abrió una sesión de Planning Poker para "([^"]*)"$`, e.abrioUnaSesion)
	sc.Step(`^"([^"]*)" abre una sesión de Planning Poker para "([^"]*)"$`, e.abreUnaSesion)
	sc.Step(`^la historia "([^"]*)" tiene una sesión de Planning Poker abierta en la ronda (\d+)$`, e.tieneUnaSesionAbiertaEnLaRonda)
	sc.Step(`^la historia "([^"]*)" no tiene una sesión de Planning Poker abierta$`, e.noTieneUnaSesionAbierta)
	sc.Step(`^la historia "([^"]*)" tiene una sola sesión de Planning Poker abierta$`, e.tieneUnaSolaSesionAbierta)
	sc.Step(`^participan estos integrantes:$`, e.participanEstosIntegrantes)
	sc.Step(`^(?:que )?se da de baja a "([^"]*)"$`, e.seDaDeBajaA)
	sc.Step(`^se registra a "([^"]*)" como integrante del proyecto$`, e.seRegistraComoIntegrante)
	sc.Step(`^abrir la sesión da el error "([^"]*)"$`, e.abrirLaSesionDaElError)
	return e
}

func (e *hu17) elProyectoConEstosIntegrantes(nombre string, tabla *godog.Table) error {
	ctx := context.Background()
	p, err := e.repoProyectos.Guardar(ctx, project.Proyecto{
		Nombre: nombre, FechaInicio: fechaDeEscenario("2026-10-05"), FechaFin: fechaDeEscenario("2026-11-01"),
	})
	if err != nil {
		return err
	}
	e.proyectoID = p.ID
	for _, fila := range tabla.Rows[1:] {
		nombre, rol, activo := fila.Cells[0].Value, fila.Cells[1].Value, fila.Cells[2].Value
		i, err := e.repoIntegrantes.Guardar(ctx, member.Integrante{
			ProyectoID: p.ID, Nombre: nombre, Email: fmt.Sprintf("integrante%d@mail.com", len(e.porNombre)+1),
			Rol: member.Rol(rol), Activo: activo == "sí",
		})
		if err != nil {
			return err
		}
		e.porNombre[nombre] = i
	}
	return nil
}

func (e *hu17) estasHistoriasParaEstimar(tabla *godog.Table) error {
	for n, fila := range tabla.Rows[1:] {
		h, err := e.repoHistorias.Guardar(context.Background(), backlog.Historia{
			ProyectoID: e.proyectoID, Numero: n + 1, Titulo: fila.Cells[0].Value,
			Prioridad: backlog.PrioridadAlta, Estado: backlog.Estado(fila.Cells[1].Value),
		})
		if err != nil {
			return err
		}
		e.porTitulo[h.Titulo] = h
	}
	return nil
}

// abrioUnaSesion es el Dado: la sesión tiene que abrirse.
func (e *hu17) abrioUnaSesion(quien, titulo string) error {
	if err := e.abreUnaSesion(quien, titulo); err != nil {
		return err
	}
	return e.err
}

// abreUnaSesion guarda el error para el paso que lo revisa.
func (e *hu17) abreUnaSesion(quien, titulo string) error {
	h, ok := e.porTitulo[titulo]
	if !ok {
		return fmt.Errorf("no hay una historia %q en el escenario", titulo)
	}
	_, e.err = e.poker.Abrir(context.Background(), e.proyectoID, h.Numero, e.porNombre[quien].ID)
	return nil
}

func (e *hu17) vista(titulo string) (service.VistaSesion, error) {
	h, ok := e.porTitulo[titulo]
	if !ok {
		return service.VistaSesion{}, fmt.Errorf("no hay una historia %q en el escenario", titulo)
	}
	return e.poker.Ver(context.Background(), e.proyectoID, h.Numero)
}

func (e *hu17) tieneUnaSesionAbiertaEnLaRonda(titulo string, ronda int) error {
	if e.err != nil {
		return fmt.Errorf("no se esperaba error al abrir: %w", e.err)
	}
	v, err := e.vista(titulo)
	if err != nil {
		return err
	}
	if !v.Abierta || v.Ronda.Numero != ronda {
		return fmt.Errorf("se esperaba la sesión abierta en la ronda %d y se obtuvo %+v", ronda, v)
	}
	return nil
}

func (e *hu17) noTieneUnaSesionAbierta(titulo string) error {
	v, err := e.vista(titulo)
	if err != nil {
		return err
	}
	if v.Abierta {
		return fmt.Errorf("no se esperaba una sesión abierta y está la %d", v.Sesion.ID)
	}
	return nil
}

// tieneUnaSolaSesionAbierta comprueba que la abierta es la primera: la segunda no se guardó (CA-17.2).
func (e *hu17) tieneUnaSolaSesionAbierta(titulo string) error {
	v, err := e.vista(titulo)
	if err != nil {
		return err
	}
	if !v.Abierta || v.Sesion.ID != 1 {
		return fmt.Errorf("se esperaba abierta solo la primera sesión y se obtuvo %+v", v.Sesion)
	}
	return nil
}

func (e *hu17) participanEstosIntegrantes(tabla *godog.Table) error {
	v, err := e.vista("Login")
	if err != nil {
		return err
	}
	var quiero, obtenidos []string
	for _, fila := range tabla.Rows[1:] {
		quiero = append(quiero, fila.Cells[0].Value)
	}
	for _, i := range v.Participantes {
		obtenidos = append(obtenidos, i.Nombre)
	}
	if !slices.Equal(obtenidos, quiero) {
		return fmt.Errorf("participan %v, se esperaba %v", obtenidos, quiero)
	}
	return nil
}

func (e *hu17) seDaDeBajaA(nombre string) error {
	_, err := e.integrantes.DarDeBaja(context.Background(), e.proyectoID, e.porNombre[nombre].ID)
	return err
}

// seRegistraComoIntegrante registra un Product Builder con el caso de uso de HU-03, como en la aplicación.
func (e *hu17) seRegistraComoIntegrante(nombre string) error {
	_, err := e.integrantes.Registrar(context.Background(), e.proyectoID, member.Datos{
		Nombre: nombre, Email: strings.ToLower(nombre) + "@mail.com", Rol: member.RolProductBuilder,
	})
	return err
}

func (e *hu17) abrirLaSesionDaElError(mensaje string) error {
	if e.err == nil {
		return errors.New("se esperaba un error al abrir la sesión y no lo hubo")
	}
	if e.err.Error() != mensaje {
		return fmt.Errorf("error = %q, se esperaba %q", e.err.Error(), mensaje)
	}
	return nil
}
