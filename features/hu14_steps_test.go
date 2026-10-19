package features

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/metrics"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// hu14 guarda el estado de un escenario de HU-14 (cerrar sprint), con sus propios repositorios en memoria y un reloj
// que fija cada escenario.
type hu14 struct {
	repoProyectos *memory.Proyectos
	repoSprints   *memory.Sprints
	repoHistorias *memory.Historias
	repoCierres   *memory.Cierres
	cierres       *service.Cierres
	asignaciones  *service.Asignaciones
	estimaciones  *service.Estimaciones
	proyectoID    int64
	sprintIDs     map[int]int64 // ID de cada sprint, por número
	ahora         time.Time
	confirmacion  string // la pantalla de confirmación, para los escenarios que la miran
	err           error
}

// registrarPasosHU14 conecta cada frase de features/hu-14-cerrar-sprint.feature con una función de Go.
func registrarPasosHU14(sc *godog.ScenarioContext) {
	e := &hu14{
		repoProyectos: memory.NuevoProyectos(), repoSprints: memory.NuevoSprints(), repoHistorias: memory.NuevoHistorias(),
		sprintIDs: map[int]int64{},
	}
	e.repoCierres = memory.NuevoCierres(e.repoSprints, e.repoHistorias)
	reloj := func() time.Time { return e.ahora }
	e.cierres = service.NuevoCierres(e.repoProyectos, e.repoSprints, e.repoHistorias, e.repoCierres, reloj)
	criterios := service.NuevoCriterios(e.repoProyectos, e.repoHistorias, memory.NuevoCriterios())
	e.asignaciones = service.NuevoAsignaciones(e.repoProyectos, e.repoSprints, e.repoHistorias, criterios, reloj)
	e.estimaciones = service.NuevoEstimaciones(e.repoProyectos, e.repoHistorias)

	sc.Step(`^el proyecto "([^"]*)" con el Sprint (\d+) del "([^"]*)" al "([^"]*)" para cerrar$`, e.elProyectoConElSprint)
	sc.Step(`^que el Sprint (\d+) va del "([^"]*)" al "([^"]*)"$`, e.queElSprintVa)
	sc.Step(`^el Sprint (\d+) tiene estas historias al cerrar:$`, e.elSprintTieneEstasHistorias)
	sc.Step(`^que hoy es el "([^"]*)" para cerrar$`, e.queHoyEs)
	sc.Step(`^que el Sprint (\d+) no tiene historias$`, e.queElSprintNoTieneHistorias)
	sc.Step(`^que cerré el Sprint (\d+)$`, e.queCerreElSprint)
	sc.Step(`^cierro el Sprint (\d+)$`, e.cierro)
	sc.Step(`^pido cerrar el Sprint (\d+)$`, e.pidoCerrar)
	sc.Step(`^re-estimo "([^"]*)" en (\d+) story points$`, e.reestimo)
	sc.Step(`^asigno "([^"]*)" al Sprint (\d+) cerrado$`, e.asignoAlSprintCerrado)
	sc.Step(`^el Sprint (\d+) queda Cerrado con (\d+) SP planificados y (\d+) SP completados$`, e.quedaCerrado)
	sc.Step(`^la foto del Sprint (\d+) tiene estas historias:$`, e.laFotoTiene)
	sc.Step(`^la foto del Sprint (\d+) sigue con (\d+) SP para "([^"]*)"$`, e.laFotoSigueCon)
	sc.Step(`^el Sprint (\d+) sigue con (\d+) SP planificados$`, e.sigueConPlanificados)
	sc.Step(`^"([^"]*)" queda "([^"]*)" y sin sprint$`, e.quedaSinSprint)
	sc.Step(`^"([^"]*)" sigue "([^"]*)" en el Sprint (\d+)$`, e.sigueEnElSprint)
	sc.Step(`^la velocidad del equipo es de (\d+) SP$`, e.laVelocidadEs)
	sc.Step(`^la confirmación avisa "([^"]*)"$`, e.laConfirmacionAvisa)
	sc.Step(`^la confirmación dice que vuelven al backlog "([^"]*)" y "([^"]*)"$`, e.laConfirmacionDiceQueVuelven)
	sc.Step(`^cerrar da el error "([^"]*)"$`, e.cerrarDaElError)
	sc.Step(`^asignar da el error "([^"]*)"$`, e.cerrarDaElError)
	sc.Step(`^el Sprint (\d+) sigue sin cerrar$`, e.sigueSinCerrar)
}

func (e *hu14) elProyectoConElSprint(nombre string, numero int, inicio, fin string) error {
	p, err := e.repoProyectos.Guardar(context.Background(), project.Proyecto{
		Nombre: nombre, FechaInicio: fechaDeEscenario("2026-10-05"), FechaFin: fechaDeEscenario("2026-11-30"),
	})
	if err != nil {
		return err
	}
	e.proyectoID = p.ID
	return e.queElSprintVa(numero, inicio, fin)
}

func (e *hu14) queElSprintVa(numero int, inicio, fin string) error {
	s, err := e.repoSprints.Guardar(context.Background(), sprint.Sprint{
		ProyectoID: e.proyectoID, Numero: numero, Objetivo: "Objetivo", FechaInicio: fechaDeEscenario(inicio),
		FechaFin: fechaDeEscenario(fin), Estado: sprint.EstadoPlanificado,
	})
	if err != nil {
		return err
	}
	e.sprintIDs[numero] = s.ID
	return nil
}

// elSprintTieneEstasHistorias guarda las historias de la tabla (por ejemplo "HU-1 Login") ya en el sprint, en su estado.
func (e *hu14) elSprintTieneEstasHistorias(numero int, tabla *godog.Table) error {
	sprintID := e.sprintIDs[numero]
	for _, fila := range tabla.Rows[1:] {
		id, titulo, _ := strings.Cut(fila.Cells[0].Value, " ")
		n, err := numeroDe(id)
		if err != nil {
			return err
		}
		sp, err := strconv.Atoi(fila.Cells[1].Value)
		if err != nil {
			return err
		}
		if _, err := e.repoHistorias.Guardar(context.Background(), backlog.Historia{
			ProyectoID: e.proyectoID, Numero: n, Titulo: titulo, Prioridad: backlog.PrioridadAlta,
			Estado: backlog.Estado(fila.Cells[2].Value), StoryPoints: &sp, SprintID: &sprintID,
		}); err != nil {
			return err
		}
	}
	return nil
}

// queHoyEs fija "ahora" a las 15:00 de ese día en Argentina.
func (e *hu14) queHoyEs(dia string) error {
	f := fechaDeEscenario(dia)
	e.ahora = time.Date(f.Year(), f.Month(), f.Day(), 15, 0, 0, 0, zonaArgentina)
	return nil
}

// queElSprintNoTieneHistorias devuelve al backlog las historias de los antecedentes.
func (e *hu14) queElSprintNoTieneHistorias(int) error {
	historias, err := e.repoHistorias.ListarPorProyecto(context.Background(), e.proyectoID)
	if err != nil {
		return err
	}
	for _, h := range historias {
		if err := e.repoHistorias.ActualizarSprint(context.Background(), h.ID, backlog.EstadoPendiente, nil); err != nil {
			return err
		}
	}
	return nil
}

func (e *hu14) queCerreElSprint(numero int) error {
	_, err := e.cierres.Cerrar(context.Background(), e.proyectoID, numero)
	return err
}

func (e *hu14) cierro(numero int) error {
	_, e.err = e.cierres.Cerrar(context.Background(), e.proyectoID, numero)
	return nil
}

// pidoCerrar abre la pantalla de confirmación, como lo hace el botón Cerrar sprint.
func (e *hu14) pidoCerrar(numero int) error {
	h := server.NewRouter(server.Dependencias{Cierres: e.cierres})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, fmt.Sprintf("/proyectos/%d/sprints/%d/cerrar", e.proyectoID, numero), nil))
	if rec.Code != http.StatusOK {
		return fmt.Errorf("la confirmación respondió %d:\n%s", rec.Code, rec.Body.String())
	}
	e.confirmacion = rec.Body.String()
	return nil
}

func (e *hu14) reestimo(id string, puntos int) error {
	n, err := numeroDe(id)
	if err != nil {
		return err
	}
	_, err = e.estimaciones.Estimar(context.Background(), e.proyectoID, n, strconv.Itoa(puntos))
	return err
}

func (e *hu14) asignoAlSprintCerrado(id string, numero int) error {
	n, err := numeroDe(id)
	if err != nil {
		return err
	}
	_, e.err = e.asignaciones.Asignar(context.Background(), e.proyectoID, numero, n)
	return nil
}

func (e *hu14) sprintGuardado(numero int) (sprint.Sprint, error) {
	sprints, err := e.repoSprints.ListarPorProyecto(context.Background(), e.proyectoID)
	if err != nil {
		return sprint.Sprint{}, err
	}
	for _, s := range sprints {
		if s.Numero == numero {
			return s, nil
		}
	}
	return sprint.Sprint{}, fmt.Errorf("no está el Sprint %d", numero)
}

func (e *hu14) quedaCerrado(numero, planificados, completados int) error {
	if e.err != nil {
		return fmt.Errorf("no se esperaba error al cerrar: %v", e.err)
	}
	s, err := e.sprintGuardado(numero)
	if err != nil {
		return err
	}
	if s.Estado != sprint.EstadoCerrado || s.SPPlanificados == nil || *s.SPPlanificados != planificados ||
		s.SPCompletados == nil || *s.SPCompletados != completados {
		return fmt.Errorf("se esperaba el Sprint %d Cerrado con %d de %d SP y quedó %+v", numero, completados, planificados, s)
	}
	return nil
}

func (e *hu14) foto(numero int) ([]sprint.FotoHistoria, error) {
	return e.repoCierres.Foto(context.Background(), e.sprintIDs[numero])
}

func (e *hu14) laFotoTiene(numero int, tabla *godog.Table) error {
	foto, err := e.foto(numero)
	if err != nil {
		return err
	}
	var obtenida []string
	for _, f := range foto {
		completada := "no"
		if f.Completada {
			completada = "sí"
		}
		obtenida = append(obtenida, fmt.Sprintf("HU-%d %d %s", f.Numero, *f.StoryPoints, completada))
	}
	var esperada []string
	for _, fila := range tabla.Rows[1:] {
		esperada = append(esperada, fila.Cells[0].Value+" "+fila.Cells[1].Value+" "+fila.Cells[2].Value)
	}
	if strings.Join(obtenida, ", ") != strings.Join(esperada, ", ") {
		return fmt.Errorf("la foto tiene %v y se esperaba %v", obtenida, esperada)
	}
	return nil
}

func (e *hu14) laFotoSigueCon(numero, puntos int, id string) error {
	n, err := numeroDe(id)
	if err != nil {
		return err
	}
	foto, err := e.foto(numero)
	if err != nil {
		return err
	}
	for _, f := range foto {
		if f.Numero == n {
			if f.StoryPoints == nil || *f.StoryPoints != puntos {
				return fmt.Errorf("la foto tiene %v SP para %s y se esperaban %d", f.StoryPoints, id, puntos)
			}
			return nil
		}
	}
	return fmt.Errorf("%s no está en la foto del Sprint %d", id, numero)
}

func (e *hu14) sigueConPlanificados(numero, planificados int) error {
	s, err := e.sprintGuardado(numero)
	if err != nil {
		return err
	}
	if s.SPPlanificados == nil || *s.SPPlanificados != planificados {
		return fmt.Errorf("el Sprint %d tiene %v SP planificados y se esperaban %d", numero, s.SPPlanificados, planificados)
	}
	return nil
}

func (e *hu14) historia(id string) (backlog.Historia, error) {
	n, err := numeroDe(id)
	if err != nil {
		return backlog.Historia{}, err
	}
	return e.repoHistorias.BuscarPorNumero(context.Background(), e.proyectoID, n)
}

func (e *hu14) quedaSinSprint(id, estado string) error {
	h, err := e.historia(id)
	if err != nil {
		return err
	}
	if string(h.Estado) != estado || h.SprintID != nil {
		return fmt.Errorf("se esperaba %s %s y sin sprint, y quedó %q en el sprint %v", id, estado, h.Estado, h.SprintID)
	}
	return nil
}

func (e *hu14) sigueEnElSprint(id, estado string, numero int) error {
	h, err := e.historia(id)
	if err != nil {
		return err
	}
	if string(h.Estado) != estado || h.SprintID == nil || *h.SprintID != e.sprintIDs[numero] {
		return fmt.Errorf("se esperaba %s %s en el Sprint %d, y quedó %q en el sprint %v", id, estado, numero, h.Estado, h.SprintID)
	}
	return nil
}

// laVelocidadEs calcula la velocidad de HU-31 con los sprints guardados, como lo hace el aviso de HU-12.
func (e *hu14) laVelocidadEs(esperada int) error {
	sprints, err := e.repoSprints.ListarPorProyecto(context.Background(), e.proyectoID)
	if err != nil {
		return err
	}
	var historial []metrics.Sprint
	for _, s := range sprints {
		m := metrics.Sprint{Numero: s.Numero, Estado: metrics.SprintPlanificado}
		if s.Estado == sprint.EstadoCerrado {
			m.Estado = metrics.SprintCerrado
		}
		if s.SPCompletados != nil {
			m.SPCompletados = float64(*s.SPCompletados)
		}
		historial = append(historial, m)
	}
	v, err := metrics.CalcularVelocidad(historial, metrics.VentanaPorDefecto)
	if err != nil {
		return err
	}
	if v.SprintsUsados == 0 || v.Valor != float64(esperada) {
		return fmt.Errorf("se esperaba una velocidad de %d SP y se obtuvo %v (con %d sprints)", esperada, v.Valor, v.SprintsUsados)
	}
	return nil
}

func (e *hu14) laConfirmacionAvisa(aviso string) error {
	if !strings.Contains(e.confirmacion, aviso) {
		return fmt.Errorf("la confirmación no avisa %q:\n%s", aviso, e.confirmacion)
	}
	return nil
}

func (e *hu14) laConfirmacionDiceQueVuelven(primera, segunda string) error {
	for _, texto := range []string{"Vuelven al backlog como Pendientes", primera + " ·", segunda + " ·"} {
		if !strings.Contains(e.confirmacion, texto) {
			return fmt.Errorf("la confirmación no dice %q:\n%s", texto, e.confirmacion)
		}
	}
	return nil
}

func (e *hu14) cerrarDaElError(mensaje string) error {
	if e.err == nil {
		return fmt.Errorf("se esperaba el error %q y no hubo error", mensaje)
	}
	if e.err.Error() != mensaje {
		return fmt.Errorf("se esperaba el error %q y se obtuvo %q", mensaje, e.err.Error())
	}
	return nil
}

func (e *hu14) sigueSinCerrar(numero int) error {
	s, err := e.sprintGuardado(numero)
	if err != nil {
		return err
	}
	if s.Estado == sprint.EstadoCerrado {
		return fmt.Errorf("el Sprint %d no se tenía que cerrar", numero)
	}
	return nil
}
