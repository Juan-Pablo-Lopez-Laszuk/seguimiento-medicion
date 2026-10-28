package features

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// hu19 guarda el estado de un escenario de HU-19 (revelar estimaciones), con sus propios repositorios en memoria.
type hu19 struct {
	repoProyectos   *memory.Proyectos
	repoHistorias   *memory.Historias
	repoIntegrantes *memory.Integrantes
	repoPoker       *memory.Poker
	poker           *service.Poker
	revelar         *service.RevelarPoker
	agileEnabler    member.Integrante
	proyectoID      int64
	numero          int
	ronda           poker.Ronda
	porNombre       map[string]member.Integrante
	err             error
}

// unVoto es cada "Nombre" N de la frase de los votos.
var unVoto = regexp.MustCompile(`"([^"]+)" (\d+)`)

// registrarPasosHU19 conecta cada frase de features/hu-19-revelar-estimaciones.feature con una función de Go.
func registrarPasosHU19(sc *godog.ScenarioContext) {
	e := &hu19{
		repoProyectos: memory.NuevoProyectos(), repoHistorias: memory.NuevoHistorias(),
		repoIntegrantes: memory.NuevoIntegrantes(),
		repoPoker:       memory.NuevoPoker(),
		porNombre:       map[string]member.Integrante{},
	}
	e.poker = service.NuevoPoker(e.repoProyectos, e.repoHistorias, e.repoIntegrantes, e.repoPoker).ConVotos(e.repoPoker)
	e.revelar = service.NuevoRevelarPoker(e.poker, e.repoPoker)

	sc.Step(`^el equipo de Planning Poker con "([^"]*)" de Agile Enabler y "([^"]*)" y "([^"]*)" de Product Builder$`, e.elEquipo)
	sc.Step(`^una sesión de Planning Poker abierta para "HU-(\d+) ([^"]*)" con la ronda 1$`, e.unaSesionAbierta)
	sc.Step(`^que en la ronda votaron (.+)$`, e.enLaRondaVotaron)
	sc.Step(`^que "([^"]*)" deja el equipo de Planning Poker$`, e.dejaElEquipo)
	sc.Step(`^"([^"]*)" cierra la ronda$`, e.cierraLaRonda)
	sc.Step(`^la ronda está revelada$`, e.laRondaEstaRevelada)
	sc.Step(`^la ronda no está revelada$`, e.laRondaNoEstaRevelada)
	sc.Step(`^se ven estos votos:$`, e.seVenEstosVotos)
	sc.Step(`^no se ve ningún voto$`, e.noSeVeNingunVoto)
	sc.Step(`^revelar da el error "([^"]*)"$`, e.revelarDaElError)
}

func (e *hu19) elEquipo(agileEnabler, pb1, pb2 string) error {
	ctx := context.Background()
	p, err := e.repoProyectos.Guardar(ctx, project.Proyecto{
		Nombre: "Seguimiento", FechaInicio: fechaDeEscenario("2026-10-05"), FechaFin: fechaDeEscenario("2026-11-01"),
	})
	if err != nil {
		return err
	}
	e.proyectoID = p.ID
	for n, nombre := range []string{agileEnabler, pb1, pb2} {
		rol := member.RolProductBuilder
		if n == 0 {
			rol = member.RolAgileEnabler
		}
		i, err := e.repoIntegrantes.Guardar(ctx, member.Integrante{
			ProyectoID: p.ID, Nombre: nombre, Email: fmt.Sprintf("poker%d@mail.com", n+1), Rol: rol, Activo: true,
		})
		if err != nil {
			return err
		}
		e.porNombre[nombre] = i
		if rol == member.RolAgileEnabler {
			e.agileEnabler = i
		}
	}
	return nil
}

// unaSesionAbierta guarda la historia y el Agile Enabler le abre la sesión, como en HU-17.
func (e *hu19) unaSesionAbierta(numero int, titulo string) error {
	ctx := context.Background()
	if _, err := e.repoHistorias.Guardar(ctx, backlog.Historia{
		ProyectoID: e.proyectoID, Numero: numero, Titulo: titulo, Prioridad: backlog.PrioridadAlta,
		Estado: backlog.EstadoPendiente,
	}); err != nil {
		return err
	}
	e.numero = numero
	v, err := e.poker.Abrir(ctx, e.proyectoID, numero, e.agileEnabler.ID)
	if err != nil {
		return err
	}
	e.ronda = v.Ronda
	return nil
}

// enLaRondaVotaron vota con Poker.Votar de HU-18 y revisa la ronda, como después de cada voto (RN1).
func (e *hu19) enLaRondaVotaron(votos string) error {
	for _, m := range unVoto.FindAllStringSubmatch(votos, -1) {
		i, ok := e.porNombre[m[1]]
		if !ok {
			return fmt.Errorf("no hay un integrante %q en el escenario", m[1])
		}
		if _, err := e.poker.Votar(context.Background(), e.proyectoID, e.numero, i.ID, m[2]); err != nil {
			return err
		}
	}
	_, err := e.revelar.Revisar(context.Background(), e.proyectoID, e.numero)
	return err
}

// dejaElEquipo da de baja al integrante con el caso de uso de HU-03, como en la aplicación (RN7).
func (e *hu19) dejaElEquipo(nombre string) error {
	i, ok := e.porNombre[nombre]
	if !ok {
		return fmt.Errorf("no hay un integrante %q en el escenario", nombre)
	}
	_, err := service.NuevoIntegrantes(e.repoProyectos, e.repoIntegrantes).DarDeBaja(context.Background(), e.proyectoID, i.ID)
	return err
}

// cierraLaRonda guarda el error para el paso que lo revisa.
func (e *hu19) cierraLaRonda(quien string) error {
	_, e.err = e.revelar.Cerrar(context.Background(), e.proyectoID, e.numero, e.porNombre[quien].ID)
	return nil
}

func (e *hu19) vista() (service.VistaRevelar, error) {
	if e.err != nil {
		return service.VistaRevelar{}, fmt.Errorf("no se esperaba error: %w", e.err)
	}
	return e.revelar.Revisar(context.Background(), e.proyectoID, e.numero)
}

func (e *hu19) laRondaEstaRevelada() error {
	v, err := e.vista()
	if err != nil {
		return err
	}
	if !v.Ronda.Revelada {
		return fmt.Errorf("la ronda %d no está revelada", v.Ronda.Numero)
	}
	return nil
}

func (e *hu19) laRondaNoEstaRevelada() error {
	v, err := e.vista()
	if err != nil {
		return err
	}
	if v.Ronda.Revelada {
		return fmt.Errorf("la ronda %d está revelada y no se esperaba", v.Ronda.Numero)
	}
	return nil
}

func (e *hu19) seVenEstosVotos(tabla *godog.Table) error {
	v, err := e.vista()
	if err != nil {
		return err
	}
	var quiere, hay []string
	for _, fila := range tabla.Rows[1:] {
		quiere = append(quiere, fila.Cells[0].Value+": "+fila.Cells[1].Value)
	}
	for _, f := range v.Votos {
		voto := "no votó"
		if f.Voto {
			voto = strconv.Itoa(f.Valor)
		}
		hay = append(hay, f.Integrante+": "+voto)
	}
	if fmt.Sprint(hay) != fmt.Sprint(quiere) {
		return fmt.Errorf("votos = %v, se esperaba %v", hay, quiere)
	}
	return nil
}

func (e *hu19) noSeVeNingunVoto() error {
	v, err := e.vista()
	if err != nil {
		return err
	}
	if len(v.Votos) > 0 {
		return fmt.Errorf("se ven %d votos y no se esperaba ninguno", len(v.Votos))
	}
	return nil
}

func (e *hu19) revelarDaElError(mensaje string) error {
	if e.err == nil || e.err.Error() != mensaje {
		return fmt.Errorf("error = %v, se esperaba %q", e.err, mensaje)
	}
	return nil
}
