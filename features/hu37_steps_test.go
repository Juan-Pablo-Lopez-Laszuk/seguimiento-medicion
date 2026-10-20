package features

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/defect"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// hu37 guarda el estado de un escenario de HU-37 (reporte de sprint). Los sprints se cierran con el caso de uso de
// HU-14 y el reporte se pide a la aplicación, como el botón de la página del sprint.
type hu37 struct {
	repoProyectos *memory.Proyectos
	repoSprints   *memory.Sprints
	repoHistorias *memory.Historias
	repoTareas    *memory.Tareas
	repoRegistros *memory.Registros
	repoDefectos  *memory.Defectos
	cierres       *service.Cierres
	reportes      *service.Reportes
	proyectoID    int64
	sprintIDs     map[int]int64     // ID de cada sprint, por número
	fines         map[int]time.Time // fecha de fin de cada sprint, por número
	ahora         time.Time
	reporte       service.ReporteSprint
	codigo        int    // código HTTP de la página del reporte
	pagina        string // la página del reporte, sin los escapes de HTML
}

// registrarPasosHU37 conecta cada frase de features/hu-37-reporte-sprint.feature con una función de Go.
func registrarPasosHU37(sc *godog.ScenarioContext) {
	e := &hu37{
		repoProyectos: memory.NuevoProyectos(), repoSprints: memory.NuevoSprints(), repoHistorias: memory.NuevoHistorias(),
		repoTareas: memory.NuevoTareas(), repoRegistros: memory.NuevoRegistros(), repoDefectos: memory.NuevoDefectos(),
		sprintIDs: map[int]int64{}, fines: map[int]time.Time{},
	}
	e.cierres = service.NuevoCierres(e.repoProyectos, e.repoSprints, e.repoHistorias, memory.NuevoCierres(e.repoSprints, e.repoHistorias), func() time.Time { return e.ahora })
	horas := service.NuevoHoras(e.repoProyectos, e.repoSprints, e.repoHistorias, e.repoTareas, e.repoRegistros)
	e.reportes = service.NuevoReportes(e.repoProyectos, e.repoSprints, e.repoHistorias, e.cierres, horas, e.repoDefectos)

	sc.Step(`^el proyecto "([^"]*)" con estos sprints para el reporte:$`, e.elProyectoConEstosSprints)
	sc.Step(`^en el Sprint (\d+) estas historias con sus horas:$`, e.enElSprintEstasHistorias)
	sc.Step(`^estos defectos:$`, e.estosDefectos)
	sc.Step(`^que el Sprint (\d+) se cerró para el reporte$`, e.queElSprintSeCerro)
	sc.Step(`^que "([^"]*)" se re-estimó en (\d+) story points y entró al Sprint (\d+) del reporte$`, e.queSeReestimoYEntro)
	sc.Step(`^el defecto "([^"]*)" se resolvió en el Sprint (\d+)$`, e.elDefectoSeResolvio)
	sc.Step(`^pido el reporte del Sprint (\d+)$`, e.pidoElReporte)
	sc.Step(`^el reporte muestra estas métricas:$`, e.elReporteMuestraEstasMetricas)
	sc.Step(`^el reporte muestra la métrica "([^"]*)" con "([^"]*)"$`, e.elReporteMuestraLaMetrica)
	sc.Step(`^el reporte muestra estas historias:$`, e.elReporteMuestraEstasHistorias)
	sc.Step(`^el reporte muestra "([^"]*)" con (\d+) SP y "([^"]*)"$`, e.elReporteMuestraLaHistoria)
	sc.Step(`^el reporte muestra estos defectos:$`, e.elReporteMuestraEstosDefectos)
	sc.Step(`^el reporte avisa "([^"]*)"$`, e.elReporteDice)
	sc.Step(`^el reporte dice "([^"]*)"$`, e.elReporteDice)
	sc.Step(`^el reporte responde (\d+) con "([^"]*)"$`, e.elReporteResponde)
}

// elProyectoConEstosSprints guarda el proyecto y sus sprints (todos sin cerrar) con las fechas de la tabla.
func (e *hu37) elProyectoConEstosSprints(nombre string, tabla *godog.Table) error {
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

// enElSprintEstasHistorias guarda las historias de la tabla (por ejemplo "HU-1 Login") en el sprint, Hechas o En
// sprint según la columna completada, con una tarea de las horas estimadas y un registro de las horas reales.
func (e *hu37) enElSprintEstasHistorias(numero int, tabla *godog.Table) error {
	ctx := context.Background()
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
		h, err := e.repoHistorias.Guardar(ctx, backlog.Historia{
			ProyectoID: e.proyectoID, Numero: n, Titulo: titulo, Prioridad: backlog.PrioridadAlta,
			Estado: estado, StoryPoints: &sp, SprintID: &sprintID,
		})
		if err != nil {
			return err
		}
		if err := e.conHoras(h.ID, fila.Cells[3].Value, fila.Cells[4].Value); err != nil {
			return err
		}
	}
	return nil
}

// conHoras guarda una tarea con las horas estimadas y un registro con las reales de la historia (nada si son 0).
func (e *hu37) conHoras(historiaID int64, estimadas, reales string) error {
	ctx := context.Background()
	if h, err := effort.ParsearHoras(estimadas); err == nil && h > 0 {
		if _, err := e.repoTareas.Guardar(ctx, effort.Tarea{HistoriaID: historiaID, Titulo: "Desarrollo", HorasEstimadas: h}); err != nil {
			return err
		}
	}
	if h, err := effort.ParsearHoras(reales); err == nil && h > 0 {
		if _, err := e.repoRegistros.Guardar(ctx, effort.Registro{HistoriaID: historiaID, Fecha: fechaDeEscenario("2026-10-15"), Actividad: "Desarrollo", Horas: h}); err != nil {
			return err
		}
	}
	return nil
}

// estosDefectos guarda los defectos de la tabla: Abiertos, o Resueltos si tienen sprint de resolución.
func (e *hu37) estosDefectos(tabla *godog.Table) error {
	ctx := context.Background()
	for _, fila := range tabla.Rows[1:] {
		n, err := numeroDe(fila.Cells[2].Value)
		if err != nil {
			return err
		}
		h, err := e.repoHistorias.BuscarPorNumero(ctx, e.proyectoID, n)
		if err != nil {
			return err
		}
		detectado, err := strconv.Atoi(fila.Cells[3].Value)
		if err != nil {
			return err
		}
		d := defect.Defecto{ProyectoID: e.proyectoID, HistoriaID: h.ID, Descripcion: fila.Cells[0].Value,
			Severidad: defect.Severidad(fila.Cells[1].Value), Estado: defect.EstadoAbierto, SprintDeteccionID: e.sprintIDs[detectado]}
		if resuelto, err := strconv.Atoi(fila.Cells[4].Value); err == nil {
			d.Estado, d.SprintResolucionID = defect.EstadoResuelto, e.sprintIDs[resuelto]
		}
		if _, err := e.repoDefectos.Guardar(ctx, d); err != nil {
			return err
		}
	}
	return nil
}

// queElSprintSeCerro cierra el sprint el día de su fecha de fin.
func (e *hu37) queElSprintSeCerro(numero int) error {
	e.ahora = e.fines[numero]
	_, err := e.cierres.Cerrar(context.Background(), e.proyectoID, numero)
	return err
}

func (e *hu37) queSeReestimoYEntro(id string, puntos, numero int) error {
	n, err := numeroDe(id)
	if err != nil {
		return err
	}
	ctx := context.Background()
	h, err := e.repoHistorias.BuscarPorNumero(ctx, e.proyectoID, n)
	if err != nil {
		return err
	}
	if err := e.repoHistorias.ActualizarStoryPoints(ctx, h.ID, puntos); err != nil {
		return err
	}
	sprintID := e.sprintIDs[numero]
	return e.repoHistorias.ActualizarSprint(ctx, h.ID, backlog.EstadoEnSprint, &sprintID)
}

func (e *hu37) elDefectoSeResolvio(descripcion string, numero int) error {
	defectos, err := e.repoDefectos.ListarPorProyecto(context.Background(), e.proyectoID)
	if err != nil {
		return err
	}
	for _, d := range defectos {
		if d.Descripcion == descripcion {
			d.Estado, d.SprintResolucionID = defect.EstadoResuelto, e.sprintIDs[numero]
			return e.repoDefectos.Actualizar(context.Background(), d)
		}
	}
	return fmt.Errorf("no está el defecto %q", descripcion)
}

// pidoElReporte pide la página del reporte y, si existe, guarda también el reporte del caso de uso.
func (e *hu37) pidoElReporte(numero int) error {
	h := server.NewRouter(server.Dependencias{Reportes: e.reportes})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, fmt.Sprintf("/proyectos/%d/sprints/%d/reporte", e.proyectoID, numero), nil))
	e.codigo, e.pagina = rec.Code, html.UnescapeString(rec.Body.String())
	if rec.Code != http.StatusOK {
		return nil // el escenario de error mira la respuesta
	}
	var err error
	e.reporte, err = e.reportes.Sprint(context.Background(), e.proyectoID, numero)
	return err
}

// metrica devuelve el texto de la tarjeta con ese nombre, como en la sección 3 de la spec.
func (e *hu37) metrica(nombre string) (string, error) {
	textos := map[string]func() string{
		"SP planificados y completados": e.reporte.SPTexto,
		"Cumplimiento":                  e.reporte.CumplimientoTexto,
		"Historias completadas":         e.reporte.CompletadasTexto,
		"Velocidad del equipo":          e.reporte.VelocidadTexto,
		"Horas estimadas y reales":      e.reporte.HorasTexto,
		"Desviación":                    e.reporte.DesviacionTexto,
		"Defectos":                      e.reporte.DefectosTexto,
	}
	texto, ok := textos[nombre]
	if !ok {
		return "", fmt.Errorf("el reporte no tiene la métrica %q", nombre)
	}
	return texto(), nil
}

func (e *hu37) elReporteMuestraLaMetrica(nombre, valor string) error {
	obtenido, err := e.metrica(nombre)
	if err != nil {
		return err
	}
	if obtenido != valor {
		return fmt.Errorf("%s: se esperaba %q y se obtuvo %q", nombre, valor, obtenido)
	}
	return e.elReporteDice(valor)
}

func (e *hu37) elReporteMuestraEstasMetricas(tabla *godog.Table) error {
	for _, fila := range tabla.Rows[1:] {
		if err := e.elReporteMuestraLaMetrica(fila.Cells[0].Value, fila.Cells[1].Value); err != nil {
			return err
		}
	}
	return nil
}

func (e *hu37) elReporteMuestraEstasHistorias(tabla *godog.Table) error {
	if len(e.reporte.Historias) != len(tabla.Rows)-1 {
		return fmt.Errorf("el reporte tiene %d historias y se esperaban %d", len(e.reporte.Historias), len(tabla.Rows)-1)
	}
	for i := range tabla.Rows[1:] {
		f := e.reporte.Historias[i]
		obtenido := []string{f.Identificador(), f.Titulo, strconv.Itoa(*f.StoryPoints), f.Resultado(),
			f.EstimadasTexto(), f.RealesTexto(), f.DesviacionTexto()}
		if err := compararFila(tabla, i, obtenido); err != nil {
			return err
		}
	}
	return nil
}

func (e *hu37) elReporteMuestraLaHistoria(id string, sp int, resultado string) error {
	for _, f := range e.reporte.Historias {
		if f.Identificador() == id {
			if *f.StoryPoints != sp || f.Resultado() != resultado {
				return fmt.Errorf("%s figura con %d SP y %q, y se esperaba %d SP y %q", id, *f.StoryPoints, f.Resultado(), sp, resultado)
			}
			return nil
		}
	}
	return fmt.Errorf("el reporte no tiene a %s", id)
}

func (e *hu37) elReporteMuestraEstosDefectos(tabla *godog.Table) error {
	if len(e.reporte.Defectos) != len(tabla.Rows)-1 {
		return fmt.Errorf("el reporte tiene %d defectos y se esperaban %d", len(e.reporte.Defectos), len(tabla.Rows)-1)
	}
	for i, fila := range tabla.Rows[1:] {
		f := e.reporte.Defectos[i]
		obtenido := []string{f.Descripcion, string(f.Severidad), string(f.Estado), f.DetectadoEn, f.ResueltoEn}
		if err := compararFila(tabla, i, obtenido); err != nil {
			return err
		}
		if err := e.elReporteDice(fila.Cells[0].Value); err != nil {
			return err
		}
	}
	return nil
}

// compararFila compara la fila i de la tabla (sin contar el encabezado) con lo obtenido, columna por columna.
func compararFila(tabla *godog.Table, i int, obtenido []string) error {
	for j, celda := range tabla.Rows[i+1].Cells {
		if obtenido[j] != celda.Value {
			return fmt.Errorf("fila %d, columna %q: se esperaba %q y se obtuvo %q", i+1, tabla.Rows[0].Cells[j].Value, celda.Value, obtenido[j])
		}
	}
	return nil
}

func (e *hu37) elReporteDice(texto string) error {
	if !strings.Contains(e.pagina, texto) {
		return fmt.Errorf("la página del reporte no muestra %q", texto)
	}
	return nil
}

func (e *hu37) elReporteResponde(codigo int, mensaje string) error {
	if e.codigo != codigo || !strings.Contains(e.pagina, mensaje) {
		return fmt.Errorf("se esperaba %d con %q y se obtuvo %d con %q", codigo, mensaje, e.codigo, e.pagina)
	}
	return nil
}
