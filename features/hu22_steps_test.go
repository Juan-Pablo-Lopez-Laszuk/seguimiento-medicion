package features

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// hu22 guarda el estado de un escenario de HU-22 (registrar la estimación acordada). Usa los repositorios y los
// integrantes del escenario de HU-17, que abre la sesión con sus pasos.
type hu22 struct {
	*hu17
	reveladas map[int64]bool // las sesiones cuya ronda actual se reveló
	err       error
}

// pokerConRondaRevelada es el repositorio en memoria del escenario, con la ronda actual revelada en las sesiones que
// pidió el paso: revelar es de HU-19, y para HU-22 alcanza con que la ronda llegue revelada.
type pokerConRondaRevelada struct {
	*memory.Poker
	reveladas map[int64]bool
}

func (p pokerConRondaRevelada) SesionAbierta(ctx context.Context, historiaID int64) (poker.Sesion, poker.Ronda, error) {
	s, r, err := p.Poker.SesionAbierta(ctx, historiaID)
	r.Revelada = r.Revelada || p.reveladas[s.ID]
	return s, r, err
}

// registrarPasosHU22 conecta cada frase de features/hu-22-registrar-estimacion.feature con una función de Go.
func registrarPasosHU22(sc *godog.ScenarioContext, e17 *hu17) {
	e := &hu22{hu17: e17, reveladas: map[int64]bool{}}
	e.poker.ConAcordar(pokerConRondaRevelada{Poker: e.repoPoker.ConHistorias(e.repoHistorias), reveladas: e.reveladas})

	sc.Step(`^que se reveló la ronda actual de "([^"]*)"$`, e.seRevelo)
	sc.Step(`^que la historia "([^"]*)" está estimada en (\d+) story points$`, e.estaEstimadaEn)
	sc.Step(`^que la historia "([^"]*)" pasó a Hecha$`, e.pasoAHecha)
	sc.Step(`^"([^"]*)" registra (\d+) como estimación acordada de "([^"]*)"$`, e.registra)
	sc.Step(`^que "([^"]*)" registró (\d+) como estimación acordada de "([^"]*)"$`, e.registro)
	sc.Step(`^la historia "([^"]*)" tiene (\d+) story points$`, e.tienePuntos)
	sc.Step(`^la historia "([^"]*)" no tiene story points$`, e.noTienePuntos)
	sc.Step(`^la historia "([^"]*)" tiene estas sesiones cerradas:$`, e.tieneEstasSesionesCerradas)
	sc.Step(`^el historial de estimaciones de "([^"]*)" tiene el cambio de (\d+) a (\d+)$`, e.elHistorialTieneElCambio)
	sc.Step(`^el historial de estimaciones de "([^"]*)" está vacío$`, e.elHistorialEstaVacio)
	sc.Step(`^registrar la estimación da el error "([^"]*)"$`, e.registrarDaElError)
}

func (e *hu22) historia(titulo string) (backlog.Historia, error) {
	h, ok := e.porTitulo[titulo]
	if !ok {
		return backlog.Historia{}, fmt.Errorf("no hay una historia %q en el escenario", titulo)
	}
	return e.repoHistorias.BuscarPorNumero(context.Background(), e.proyectoID, h.Numero)
}

func (e *hu22) seRevelo(titulo string) error {
	h, err := e.historia(titulo)
	if err != nil {
		return err
	}
	s, _, err := e.repoPoker.SesionAbierta(context.Background(), h.ID)
	if err != nil {
		return fmt.Errorf("revelar la ronda de %q: %w", titulo, err)
	}
	e.reveladas[s.ID] = true
	return nil
}

func (e *hu22) estaEstimadaEn(titulo string, puntos int) error {
	h, err := e.historia(titulo)
	if err != nil {
		return err
	}
	return e.repoHistorias.ActualizarStoryPoints(context.Background(), h.ID, puntos)
}

func (e *hu22) pasoAHecha(titulo string) error {
	h, err := e.historia(titulo)
	if err != nil {
		return err
	}
	return e.repoHistorias.ActualizarEstado(context.Background(), h.ID, backlog.EstadoHecho)
}

// registra guarda el error para el paso que lo revisa.
func (e *hu22) registra(quien string, valor int, titulo string) error {
	h, err := e.historia(titulo)
	if err != nil {
		return err
	}
	_, e.err = e.poker.Acordar(context.Background(), e.proyectoID, h.Numero, e.porNombre[quien].ID, strconv.Itoa(valor))
	return nil
}

// registro es el Dado: el registro tiene que funcionar.
func (e *hu22) registro(quien string, valor int, titulo string) error {
	if err := e.registra(quien, valor, titulo); err != nil {
		return err
	}
	return e.err
}

func (e *hu22) tienePuntos(titulo string, puntos int) error {
	h, err := e.historia(titulo)
	if err != nil {
		return err
	}
	if h.StoryPoints == nil || *h.StoryPoints != puntos {
		return fmt.Errorf("la historia %q tiene %v story points, se esperaba %d (error al registrar: %v)", titulo, h.StoryPoints, puntos, e.err)
	}
	return nil
}

func (e *hu22) noTienePuntos(titulo string) error {
	h, err := e.historia(titulo)
	if err != nil {
		return err
	}
	if h.StoryPoints != nil {
		return fmt.Errorf("la historia %q no tenía que tener story points y tiene %d", titulo, *h.StoryPoints)
	}
	return nil
}

func (e *hu22) tieneEstasSesionesCerradas(titulo string, tabla *godog.Table) error {
	h, err := e.historia(titulo)
	if err != nil {
		return err
	}
	v, err := e.poker.Ver(context.Background(), e.proyectoID, h.Numero)
	if err != nil {
		return err
	}
	var quiero []poker.SesionCerrada
	for _, fila := range tabla.Rows[1:] {
		var c poker.SesionCerrada
		if _, err := fmt.Sscanf(fila.Cells[0].Value, "Sesión %d", &c.Numero); err != nil {
			return fmt.Errorf("sesión %q: %w", fila.Cells[0].Value, err)
		}
		if c.Rondas, err = strconv.Atoi(fila.Cells[1].Value); err != nil {
			return err
		}
		if c.ValorAcordado, err = strconv.Atoi(fila.Cells[2].Value); err != nil {
			return err
		}
		quiero = append(quiero, c)
	}
	if !slices.Equal(v.Cerradas, quiero) {
		return fmt.Errorf("sesiones cerradas = %+v, se esperaba %+v", v.Cerradas, quiero)
	}
	return nil
}

func (e *hu22) elHistorialTieneElCambio(titulo string, anterior, nuevo int) error {
	h, err := e.historia(titulo)
	if err != nil {
		return err
	}
	historial, err := e.repoHistorias.ListarEstimaciones(context.Background(), h.ID)
	if err != nil {
		return err
	}
	if len(historial) != 1 || historial[0].Anterior != anterior || historial[0].Nuevo != nuevo {
		return fmt.Errorf("historial = %+v, se esperaba solo el cambio de %d a %d", historial, anterior, nuevo)
	}
	return nil
}

func (e *hu22) elHistorialEstaVacio(titulo string) error {
	h, err := e.historia(titulo)
	if err != nil {
		return err
	}
	historial, err := e.repoHistorias.ListarEstimaciones(context.Background(), h.ID)
	if err != nil {
		return err
	}
	if len(historial) != 0 {
		return fmt.Errorf("el historial tenía que estar vacío y es %+v", historial)
	}
	return nil
}

// registrarDaElError compara el mensaje: el de un valor fuera de la escala es el del campo story_points.
func (e *hu22) registrarDaElError(mensaje string) error {
	if e.err == nil {
		return errors.New("se esperaba un error al registrar la estimación y no lo hubo")
	}
	obtenido := e.err.Error()
	var errs domain.ErroresValidacion
	if errors.As(e.err, &errs) {
		obtenido = errs[backlog.CampoStoryPoints].Error()
	}
	if obtenido != mensaje {
		return fmt.Errorf("error = %q, se esperaba %q", obtenido, mensaje)
	}
	return nil
}
