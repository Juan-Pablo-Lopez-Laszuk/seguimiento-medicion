package features

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/defect"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// hu38 guarda el estado de un escenario de HU-38 (reporte de proyecto). Usa los repositorios y los casos de uso de
// HU-37; el reporte se pide a la aplicación, como el botón de la página del proyecto.
type hu38 struct {
	*hu37
	proyectos map[string]int64 // ID de cada proyecto, por nombre
	reporte   service.ReporteProyecto
	codigo    int    // código HTTP de la página del reporte
	pagina    string // la página del reporte, sin los escapes de HTML
}

// registrarPasosHU38 conecta cada frase de features/hu-38-reporte-proyecto.feature con una función de Go.
func registrarPasosHU38(sc *godog.ScenarioContext) {
	e := &hu38{hu37: nuevoHU37(), proyectos: map[string]int64{}}

	sc.Step(`^el proyecto "([^"]*)" con estos sprints para el reporte del proyecto:$`, e.elProyectoConEstosSprints)
	sc.Step(`^estas historias del proyecto con sus horas:$`, e.estasHistoriasDelProyecto)
	sc.Step(`^que el Sprint (\d+) se cerró con "([^"]*)" y completó "([^"]*)"$`, e.queElSprintSeCerroConYCompleto)
	sc.Step(`^que "([^"]*)" está en el Sprint (\d+)$`, e.queEstaEnElSprint)
	sc.Step(`^estos defectos del proyecto:$`, e.estosDefectosDelProyecto)
	sc.Step(`^que "([^"]*)" se re-estimó en (\d+) story points para el reporte del proyecto$`, e.queSeReestimo)
	sc.Step(`^el proyecto "([^"]*)" sin sprints$`, e.elProyectoSinSprints)
	sc.Step(`^pido el reporte del proyecto$`, e.pidoElReporteDelProyecto)
	sc.Step(`^pido el reporte del proyecto "([^"]*)"$`, e.pidoElReporteDelProyectoLlamado)
	sc.Step(`^pido el reporte del proyecto (\d+)$`, e.pidoElReporteDelProyectoConID)
	sc.Step(`^el reporte del proyecto muestra este resumen:$`, e.muestraEsteResumen)
	sc.Step(`^el reporte del proyecto tiene los bloques "([^"]*)"$`, e.tieneLosBloques)
	sc.Step(`^cada bloque muestra lo mismo que el reporte de ese sprint$`, e.cadaBloqueEsSuReporte)
	sc.Step(`^el bloque del "([^"]*)" avisa "([^"]*)"$`, e.elBloqueAvisa)
	sc.Step(`^el reporte del proyecto muestra esta evolución de la velocidad:$`, e.muestraEstaEvolucion)
	sc.Step(`^la fila del "([^"]*)" en la evolución tiene (\d+) SP planificados y (\d+) completados$`, e.laFilaTiene)
	sc.Step(`^el bloque del "([^"]*)" muestra "([^"]*)" con (\d+) SP$`, e.elBloqueMuestraLaHistoria)
	sc.Step(`^la evolución de la velocidad dice "([^"]*)"$`, e.elReporteDelProyectoDice)
	sc.Step(`^el reporte del proyecto dice "([^"]*)"$`, e.elReporteDelProyectoDice)
	sc.Step(`^el reporte del proyecto responde (\d+) con "([^"]*)"$`, e.elReporteDelProyectoResponde)
}

func (e *hu38) elProyectoConEstosSprints(nombre string, tabla *godog.Table) error {
	if err := e.hu37.elProyectoConEstosSprints(nombre, tabla); err != nil {
		return err
	}
	e.proyectos[nombre] = e.proyectoID
	return nil
}

// estasHistoriasDelProyecto guarda las historias de la tabla en el backlog (Pendientes, sin sprint) con una tarea de
// las horas estimadas y un registro de las horas reales.
func (e *hu38) estasHistoriasDelProyecto(tabla *godog.Table) error {
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
		h, err := e.repoHistorias.Guardar(context.Background(), backlog.Historia{
			ProyectoID: e.proyectoID, Numero: n, Titulo: titulo, Prioridad: backlog.PrioridadAlta,
			Estado: backlog.EstadoPendiente, StoryPoints: &sp,
		})
		if err != nil {
			return err
		}
		if err := e.conHoras(h.ID, fila.Cells[2].Value, fila.Cells[3].Value); err != nil {
			return err
		}
	}
	return nil
}

// ponerEnSprint deja la historia (por ejemplo "HU-1") en el sprint con ese estado.
func (e *hu38) ponerEnSprint(id string, numero int, estado backlog.Estado) error {
	n, err := numeroDe(id)
	if err != nil {
		return err
	}
	ctx := context.Background()
	h, err := e.repoHistorias.BuscarPorNumero(ctx, e.proyectoID, n)
	if err != nil {
		return err
	}
	sprintID := e.sprintIDs[numero]
	return e.repoHistorias.ActualizarSprint(ctx, h.ID, estado, &sprintID)
}

// queElSprintSeCerroConYCompleto pone las historias en el sprint (Hechas las completadas) y lo cierra el día de su fin.
func (e *hu38) queElSprintSeCerroConYCompleto(numero int, historias, completadas string) error {
	hechas := map[string]bool{}
	for _, id := range strings.Split(completadas, ", ") {
		hechas[id] = true
	}
	for _, id := range strings.Split(historias, ", ") {
		estado := backlog.EstadoEnSprint
		if hechas[id] {
			estado = backlog.EstadoHecho
		}
		if err := e.ponerEnSprint(id, numero, estado); err != nil {
			return err
		}
	}
	return e.queElSprintSeCerro(numero)
}

func (e *hu38) queEstaEnElSprint(id string, numero int) error {
	return e.ponerEnSprint(id, numero, backlog.EstadoEnSprint)
}

// estosDefectosDelProyecto guarda los defectos de la tabla, de una historia o de ninguna: Abiertos, o Resueltos si
// tienen sprint de resolución.
func (e *hu38) estosDefectosDelProyecto(tabla *godog.Table) error {
	ctx := context.Background()
	for _, fila := range tabla.Rows[1:] {
		var historiaID int64
		if id := fila.Cells[2].Value; id != "" {
			n, err := numeroDe(id)
			if err != nil {
				return err
			}
			h, err := e.repoHistorias.BuscarPorNumero(ctx, e.proyectoID, n)
			if err != nil {
				return err
			}
			historiaID = h.ID
		}
		detectado, err := strconv.Atoi(fila.Cells[3].Value)
		if err != nil {
			return err
		}
		d := defect.Defecto{ProyectoID: e.proyectoID, HistoriaID: historiaID, Descripcion: fila.Cells[0].Value,
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

func (e *hu38) queSeReestimo(id string, puntos int) error {
	n, err := numeroDe(id)
	if err != nil {
		return err
	}
	ctx := context.Background()
	h, err := e.repoHistorias.BuscarPorNumero(ctx, e.proyectoID, n)
	if err != nil {
		return err
	}
	return e.repoHistorias.ActualizarStoryPoints(ctx, h.ID, puntos)
}

func (e *hu38) elProyectoSinSprints(nombre string) error {
	p, err := e.repoProyectos.Guardar(context.Background(), project.Proyecto{
		Nombre: nombre, FechaInicio: fechaDeEscenario("2026-10-05"), FechaFin: fechaDeEscenario("2026-11-30"),
	})
	if err != nil {
		return err
	}
	e.proyectos[nombre] = p.ID
	return nil
}

func (e *hu38) pidoElReporteDelProyecto() error {
	return e.pedir(e.proyectoID)
}

func (e *hu38) pidoElReporteDelProyectoLlamado(nombre string) error {
	id, ok := e.proyectos[nombre]
	if !ok {
		return fmt.Errorf("no está el proyecto %q", nombre)
	}
	return e.pedir(id)
}

func (e *hu38) pidoElReporteDelProyectoConID(id int) error {
	return e.pedir(int64(id))
}

// pedir pide la página del reporte del proyecto y, si existe, guarda también el reporte del caso de uso.
func (e *hu38) pedir(proyectoID int64) error {
	h := server.NewRouter(server.Dependencias{Reportes: e.reportes})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, fmt.Sprintf("/proyectos/%d/reporte", proyectoID), nil))
	e.codigo, e.pagina = rec.Code, html.UnescapeString(rec.Body.String())
	if rec.Code != http.StatusOK {
		return nil // el escenario de error mira la respuesta
	}
	var err error
	e.reporte, err = e.reportes.Proyecto(context.Background(), proyectoID)
	return err
}

// muestraEsteResumen compara cada tarjeta del resumen con la tabla y revisa que la página la muestre.
func (e *hu38) muestraEsteResumen(tabla *godog.Table) error {
	textos := map[string]func() string{
		"Sprints":                  e.reporte.SprintsTexto,
		"SP completados":           e.reporte.SPTexto,
		"Velocidad del equipo":     e.reporte.VelocidadTexto,
		"Horas estimadas y reales": e.reporte.HorasTexto,
		"Desviación":               e.reporte.DesviacionTexto,
		"Defectos":                 e.reporte.DefectosTexto,
	}
	for _, fila := range tabla.Rows[1:] {
		nombre, valor := fila.Cells[0].Value, fila.Cells[1].Value
		texto, ok := textos[nombre]
		if !ok {
			return fmt.Errorf("el resumen no tiene la tarjeta %q", nombre)
		}
		if obtenido := texto(); obtenido != valor {
			return fmt.Errorf("%s: se esperaba %q y se obtuvo %q", nombre, valor, obtenido)
		}
		if err := e.elReporteDelProyectoDice(valor); err != nil {
			return err
		}
	}
	return nil
}

func (e *hu38) tieneLosBloques(lista string) error {
	quiero := strings.Split(lista, ", ")
	if len(e.reporte.Sprints) != len(quiero) {
		return fmt.Errorf("el reporte tiene %d bloques y se esperaban %d", len(e.reporte.Sprints), len(quiero))
	}
	for i, nombre := range quiero {
		if obtenido := fmt.Sprintf("Sprint %d", e.reporte.Sprints[i].Sprint.Numero); obtenido != nombre {
			return fmt.Errorf("bloque %d: se esperaba %q y se obtuvo %q", i+1, nombre, obtenido)
		}
		if _, err := e.bloque(nombre); err != nil {
			return err
		}
	}
	return nil
}

func (e *hu38) cadaBloqueEsSuReporte() error {
	for _, bloque := range e.reporte.Sprints {
		reporte, err := e.reportes.Sprint(context.Background(), e.proyectoID, bloque.Sprint.Numero)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(bloque, reporte) {
			return fmt.Errorf("el bloque del Sprint %d no es igual a su reporte", bloque.Sprint.Numero)
		}
	}
	return nil
}

// bloque devuelve el pedazo de la página con el bloque del sprint ("Sprint 3"), hasta el bloque siguiente.
func (e *hu38) bloque(nombre string) (string, error) {
	inicio := strings.Index(e.pagina, `id="`+strings.ToLower(strings.ReplaceAll(nombre, " ", "-"))+`"`)
	if inicio < 0 {
		return "", fmt.Errorf("la página no tiene el bloque del %s", nombre)
	}
	resto := e.pagina[inicio+1:]
	if fin := strings.Index(resto, `id="sprint-`); fin >= 0 {
		resto = resto[:fin]
	}
	return resto, nil
}

func (e *hu38) elBloqueAvisa(nombre, texto string) error {
	b, err := e.bloque(nombre)
	if err != nil {
		return err
	}
	if !strings.Contains(b, texto) {
		return fmt.Errorf("el bloque del %s no dice %q", nombre, texto)
	}
	return nil
}

func (e *hu38) muestraEstaEvolucion(tabla *godog.Table) error {
	evolucion := e.reporte.Evolucion()
	if len(evolucion) != len(tabla.Rows)-1 {
		return fmt.Errorf("la evolución tiene %d filas y se esperaban %d", len(evolucion), len(tabla.Rows)-1)
	}
	for i := range tabla.Rows[1:] {
		f := evolucion[i]
		obtenido := []string{fmt.Sprintf("Sprint %d", f.Sprint.Numero), fmt.Sprintf("%g", f.SP.Planificados),
			fmt.Sprintf("%g", f.SP.Completados), f.CumplimientoTexto(), f.VelocidadPorSprint()}
		if err := compararFila(tabla, i, obtenido); err != nil {
			return err
		}
	}
	return e.elReporteDelProyectoDice("Evolución de la velocidad")
}

func (e *hu38) laFilaTiene(nombre string, planificados, completados int) error {
	for _, f := range e.reporte.Evolucion() {
		if fmt.Sprintf("Sprint %d", f.Sprint.Numero) == nombre {
			if f.SP.Planificados != float64(planificados) || f.SP.Completados != float64(completados) {
				return fmt.Errorf("la fila del %s tiene %g de %g SP", nombre, f.SP.Completados, f.SP.Planificados)
			}
			return nil
		}
	}
	return fmt.Errorf("la evolución no tiene al %s", nombre)
}

func (e *hu38) elBloqueMuestraLaHistoria(nombre, id string, sp int) error {
	for _, bloque := range e.reporte.Sprints {
		if fmt.Sprintf("Sprint %d", bloque.Sprint.Numero) != nombre {
			continue
		}
		for _, f := range bloque.Historias {
			if f.Identificador() == id {
				if *f.StoryPoints != sp {
					return fmt.Errorf("en el bloque del %s, %s tiene %d SP y se esperaban %d", nombre, id, *f.StoryPoints, sp)
				}
				return nil
			}
		}
		return fmt.Errorf("el bloque del %s no tiene a %s", nombre, id)
	}
	return fmt.Errorf("el reporte no tiene el bloque del %s", nombre)
}

func (e *hu38) elReporteDelProyectoDice(texto string) error {
	if !strings.Contains(e.pagina, texto) {
		return fmt.Errorf("la página del reporte del proyecto no muestra %q", texto)
	}
	return nil
}

func (e *hu38) elReporteDelProyectoResponde(codigo int, mensaje string) error {
	if e.codigo != codigo || !strings.Contains(e.pagina, mensaje) {
		return fmt.Errorf("se esperaba %d con %q y se obtuvo %d con %q", codigo, mensaje, e.codigo, e.pagina)
	}
	return nil
}
