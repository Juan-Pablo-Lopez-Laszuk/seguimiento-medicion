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
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// hu15 guarda el estado de un escenario de HU-15 (consultar sprints anteriores). Los sprints se cierran con el caso de
// uso de HU-14, con "hoy" en la fecha de fin de cada uno.
type hu15 struct {
	repoProyectos *memory.Proyectos
	repoSprints   *memory.Sprints
	repoHistorias *memory.Historias
	sprints       *service.Sprints
	cierres       *service.Cierres
	asignaciones  *service.Asignaciones
	proyectoID    int64
	sprintIDs     map[int]int64     // ID de cada sprint, por número
	fines         map[int]time.Time // fecha de fin de cada sprint, por número
	ahora         time.Time
	lista         []service.ResumenSprint
	detalle       service.DetalleCierre
	respuesta     string // lo que respondió la página del sprint, para el escenario de error
}

// registrarPasosHU15 conecta cada frase de features/hu-15-sprints-anteriores.feature con una función de Go.
func registrarPasosHU15(sc *godog.ScenarioContext) {
	e := &hu15{
		repoProyectos: memory.NuevoProyectos(), repoSprints: memory.NuevoSprints(), repoHistorias: memory.NuevoHistorias(),
		sprintIDs: map[int]int64{}, fines: map[int]time.Time{},
	}
	reloj := func() time.Time { return e.ahora }
	e.sprints = service.NuevoSprints(e.repoProyectos, e.repoSprints)
	e.cierres = service.NuevoCierres(e.repoProyectos, e.repoSprints, e.repoHistorias, memory.NuevoCierres(e.repoSprints, e.repoHistorias), reloj)
	criterios := service.NuevoCriterios(e.repoProyectos, e.repoHistorias, memory.NuevoCriterios())
	e.asignaciones = service.NuevoAsignaciones(e.repoProyectos, e.repoSprints, e.repoHistorias, criterios, reloj)

	sc.Step(`^el proyecto "([^"]*)" con estos sprints para consultar:$`, e.elProyectoConEstosSprints)
	sc.Step(`^que el Sprint (\d+) se cerró con estas historias:$`, e.queElSprintSeCerroCon)
	sc.Step(`^que el Sprint (\d+) se cerró sin historias$`, e.queElSprintSeCerroSinHistorias)
	sc.Step(`^que "([^"]*)" se re-estimó en (\d+) story points y entró al Sprint (\d+)$`, e.queSeReestimoYEntro)
	sc.Step(`^consulto los sprints del proyecto$`, e.consultoLosSprints)
	sc.Step(`^consulto el detalle del Sprint (\d+)$`, e.consultoElDetalle)
	sc.Step(`^la lista de sprints muestra, en este orden:$`, e.laListaMuestra)
	sc.Step(`^la lista muestra el Sprint (\d+) con "([^"]*)" y cumplimiento "([^"]*)"$`, e.laListaMuestraElSprint)
	sc.Step(`^el detalle dice "([^"]*)"$`, e.elDetalleDice)
	sc.Step(`^el detalle del Sprint (\d+) dice "([^"]*)"$`, e.elDetalleDelSprintDice)
	sc.Step(`^el detalle muestra estas historias al cerrar:$`, e.elDetalleMuestraEstasHistorias)
	sc.Step(`^el detalle muestra "([^"]*)" con (\d+) SP y "([^"]*)"$`, e.elDetalleMuestraLaHistoria)
	sc.Step(`^la consulta del sprint responde "([^"]*)"$`, e.laConsultaResponde)
}

// elProyectoConEstosSprints guarda el proyecto y sus sprints (todos sin cerrar) con las fechas de la tabla.
func (e *hu15) elProyectoConEstosSprints(nombre string, tabla *godog.Table) error {
	ctx := context.Background()
	p, err := e.repoProyectos.Guardar(ctx, project.Proyecto{
		Nombre: nombre, FechaInicio: fechaDeEscenario("2026-10-05"), FechaFin: fechaDeEscenario("2026-11-30"),
	})
	if err != nil {
		return err
	}
	e.proyectoID = p.ID
	for _, fila := range tabla.Rows[1:] {
		numero, err := strconv.Atoi(fila.Cells[0].Value)
		if err != nil {
			return err
		}
		s, err := e.repoSprints.Guardar(ctx, sprint.Sprint{
			ProyectoID: e.proyectoID, Numero: numero, Objetivo: fila.Cells[1].Value,
			FechaInicio: fechaDeEscenario(fila.Cells[2].Value), FechaFin: fechaDeEscenario(fila.Cells[3].Value),
			Estado: sprint.EstadoPlanificado,
		})
		if err != nil {
			return err
		}
		e.sprintIDs[numero], e.fines[numero] = s.ID, s.FechaFin
	}
	return nil
}

// queElSprintSeCerroCon guarda las historias de la tabla (por ejemplo "HU-1 Login") en el sprint, Hechas o En sprint
// según la columna completada, y cierra el sprint el día de su fecha de fin.
func (e *hu15) queElSprintSeCerroCon(numero int, tabla *godog.Table) error {
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
		estado := backlog.EstadoEnSprint
		if fila.Cells[2].Value == "sí" {
			estado = backlog.EstadoHecho
		}
		if _, err := e.repoHistorias.Guardar(context.Background(), backlog.Historia{
			ProyectoID: e.proyectoID, Numero: n, Titulo: titulo, Prioridad: backlog.PrioridadAlta,
			Estado: estado, StoryPoints: &sp, SprintID: &sprintID,
		}); err != nil {
			return err
		}
	}
	return e.queElSprintSeCerroSinHistorias(numero)
}

func (e *hu15) queElSprintSeCerroSinHistorias(numero int) error {
	e.ahora = e.fines[numero]
	_, err := e.cierres.Cerrar(context.Background(), e.proyectoID, numero)
	return err
}

func (e *hu15) queSeReestimoYEntro(id string, puntos, numero int) error {
	n, err := numeroDe(id)
	if err != nil {
		return err
	}
	h, err := e.repoHistorias.BuscarPorNumero(context.Background(), e.proyectoID, n)
	if err != nil {
		return err
	}
	if err := e.repoHistorias.ActualizarStoryPoints(context.Background(), h.ID, puntos); err != nil {
		return err
	}
	sprintID := e.sprintIDs[numero]
	return e.repoHistorias.ActualizarSprint(context.Background(), h.ID, backlog.EstadoEnSprint, &sprintID)
}

func (e *hu15) consultoLosSprints() error {
	var err error
	_, e.lista, err = e.sprints.Resumir(context.Background(), e.proyectoID)
	return err
}

// consultoElDetalle pide la página del sprint, como el enlace de la lista, y guarda el detalle del cierre.
func (e *hu15) consultoElDetalle(numero int) error {
	h := server.NewRouter(server.Dependencias{Asignaciones: e.asignaciones, Cierres: e.cierres})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, fmt.Sprintf("/proyectos/%d/sprints/%d", e.proyectoID, numero), nil))
	e.respuesta = rec.Body.String()
	if rec.Code != http.StatusOK {
		return nil // el escenario de error mira la respuesta
	}
	var err error
	e.detalle, err = e.cierres.Detalle(context.Background(), e.proyectoID, numero)
	return err
}

func (e *hu15) laListaMuestra(tabla *godog.Table) error {
	if len(e.lista) != len(tabla.Rows)-1 {
		return fmt.Errorf("la lista tiene %d sprints y se esperaban %d", len(e.lista), len(tabla.Rows)-1)
	}
	for i, fila := range tabla.Rows[1:] {
		r := e.lista[i]
		obtenido := []string{fmt.Sprintf("Sprint %d", r.Numero), r.Objetivo, r.FechaInicio.Format("02/01/2006"),
			r.FechaFin.Format("02/01/2006"), r.SPCompletadosTexto(), r.CumplimientoTexto()}
		for j, celda := range fila.Cells {
			if obtenido[j] != celda.Value {
				return fmt.Errorf("fila %d, columna %q: se esperaba %q y se obtuvo %q", i+1, tabla.Rows[0].Cells[j].Value, celda.Value, obtenido[j])
			}
		}
	}
	return nil
}

func (e *hu15) laListaMuestraElSprint(numero int, sp, cumplimiento string) error {
	if err := e.consultoLosSprints(); err != nil {
		return err
	}
	for _, r := range e.lista {
		if r.Numero == numero {
			if r.SPCompletadosTexto() != sp || r.CumplimientoTexto() != cumplimiento {
				return fmt.Errorf("el Sprint %d muestra %q y %q, y se esperaba %q y %q", numero, r.SPCompletadosTexto(), r.CumplimientoTexto(), sp, cumplimiento)
			}
			return nil
		}
	}
	return fmt.Errorf("la lista no tiene el Sprint %d", numero)
}

func (e *hu15) elDetalleDice(texto string) error {
	if e.detalle.Resumen() != texto {
		return fmt.Errorf("el detalle dice %q y se esperaba %q", e.detalle.Resumen(), texto)
	}
	if !strings.Contains(e.respuesta, texto) {
		return fmt.Errorf("la página del sprint no muestra %q", texto)
	}
	return nil
}

func (e *hu15) elDetalleDelSprintDice(numero int, texto string) error {
	if err := e.consultoElDetalle(numero); err != nil {
		return err
	}
	return e.elDetalleDice(texto)
}

func (e *hu15) elDetalleMuestraEstasHistorias(tabla *godog.Table) error {
	if len(e.detalle.Historias) != len(tabla.Rows)-1 {
		return fmt.Errorf("el detalle tiene %d historias y se esperaban %d", len(e.detalle.Historias), len(tabla.Rows)-1)
	}
	for i, fila := range tabla.Rows[1:] {
		f := e.detalle.Historias[i]
		obtenido := []string{f.Identificador(), f.Titulo, strconv.Itoa(*f.StoryPoints), f.Resultado()}
		for j, celda := range fila.Cells {
			if obtenido[j] != celda.Value {
				return fmt.Errorf("fila %d, columna %q: se esperaba %q y se obtuvo %q", i+1, tabla.Rows[0].Cells[j].Value, celda.Value, obtenido[j])
			}
		}
	}
	return nil
}

func (e *hu15) elDetalleMuestraLaHistoria(id string, sp int, resultado string) error {
	for _, f := range e.detalle.Historias {
		if f.Identificador() == id {
			if *f.StoryPoints != sp || f.Resultado() != resultado {
				return fmt.Errorf("%s figura con %d SP y %q, y se esperaba %d SP y %q", id, *f.StoryPoints, f.Resultado(), sp, resultado)
			}
			return nil
		}
	}
	return fmt.Errorf("el detalle no tiene a %s", id)
}

func (e *hu15) laConsultaResponde(mensaje string) error {
	if !strings.Contains(e.respuesta, mensaje) {
		return fmt.Errorf("la página respondió %q y se esperaba %q", e.respuesta, mensaje)
	}
	return nil
}
