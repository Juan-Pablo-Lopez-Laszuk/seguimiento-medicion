package features

import (
	"context"
	"encoding/json"
	"fmt"
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

// hu36 guarda el estado de un escenario de HU-36 (gráficos de métricas), con sus propios repositorios en memoria. Los
// sprints se cierran con el caso de uso de HU-14 y el JSON se pide a la aplicación, como lo hace la página.
type hu36 struct {
	repoProyectos *memory.Proyectos
	repoSprints   *memory.Sprints
	repoHistorias *memory.Historias
	repoTareas    *memory.Tareas
	repoRegistros *memory.Registros
	repoDefectos  *memory.Defectos
	cierres       *service.Cierres
	graficos      *service.Graficos
	router        http.Handler
	ahora         time.Time
	proyectoID    int64
	sprintIDs     map[int]int64
	fines         map[int]time.Time
	datos         service.DatosGraficos
	codigo        int
	json          map[string]any
}

// registrarPasosHU36 conecta cada frase de features/hu-36-graficos.feature con una función de Go.
func registrarPasosHU36(sc *godog.ScenarioContext) {
	e := &hu36{
		repoProyectos: memory.NuevoProyectos(), repoSprints: memory.NuevoSprints(), repoHistorias: memory.NuevoHistorias(),
		repoTareas: memory.NuevoTareas(), repoRegistros: memory.NuevoRegistros(), repoDefectos: memory.NuevoDefectos(),
		sprintIDs: map[int]int64{}, fines: map[int]time.Time{},
	}
	reloj := func() time.Time { return e.ahora }
	e.cierres = service.NuevoCierres(e.repoProyectos, e.repoSprints, e.repoHistorias, memory.NuevoCierres(e.repoSprints, e.repoHistorias), reloj)
	horas := service.NuevoHoras(e.repoProyectos, e.repoSprints, e.repoHistorias, e.repoTareas, e.repoRegistros)
	defectos := service.NuevoDefectos(e.repoProyectos, e.repoSprints, e.repoHistorias, e.repoDefectos)
	panel := service.NuevoPanel(service.NuevoEstadoProyectos(e.repoProyectos, e.repoSprints, e.repoHistorias, reloj), e.repoSprints, horas, defectos)
	reportes := service.NuevoReportes(e.repoProyectos, e.repoSprints, e.repoHistorias, e.cierres, horas, e.repoDefectos, panel)
	e.graficos = service.NuevoGraficos(reportes, defectos, e.repoHistorias, reloj)
	e.router = server.NewRouter(server.Dependencias{Graficos: e.graficos})

	sc.Step(`^el proyecto "([^"]*)" con estos sprints para los gráficos:$`, e.elProyectoConSprints)
	sc.Step(`^el proyecto "([^"]*)" sin sprints para los gráficos$`, func(nombre string) error { return e.elProyecto(nombre) })
	sc.Step(`^para los gráficos estas historias:$`, e.estasHistorias)
	sc.Step(`^que el Sprint (\d+) se cerró para los gráficos$`, e.seCerro)
	sc.Step(`^para los gráficos estos defectos:$`, e.estosDefectos)
	sc.Step(`^hoy es (\d{4}-\d\d-\d\d) para los gráficos$`, func(fecha string) error {
		e.ahora = fechaDeEscenario(fecha).Add(18 * time.Hour) // las 15:00 de Argentina
		return nil
	})
	sc.Step(`^pido los datos de los gráficos$`, func() error { return e.pedirDatos(0) })
	sc.Step(`^pido los datos de los gráficos del Sprint (\d+)$`, e.pedirDatos)
	sc.Step(`^el gráfico de story points tiene:$`, func(t *godog.Table) error {
		return compararTabla(t, e.datos.FilasStoryPoints())
	})
	sc.Step(`^el gráfico de velocidad tiene:$`, func(t *godog.Table) error { return compararTabla(t, e.datos.FilasVelocidad()) })
	sc.Step(`^el gráfico de horas tiene:$`, func(t *godog.Table) error { return compararTabla(t, e.datos.FilasHoras()) })
	sc.Step(`^el gráfico de defectos tiene:$`, func(t *godog.Table) error { return compararTabla(t, e.datos.FilasDefectos()) })
	sc.Step(`^el burndown es del "([^"]*)" con:$`, e.elBurndown)
	sc.Step(`^pido el JSON de los gráficos$`, func() error { return e.pedirJSON(e.proyectoID) })
	sc.Step(`^pido el JSON de los gráficos del proyecto (\d+)$`, func(id int64) error { return e.pedirJSON(id) })
	sc.Step(`^el JSON responde (\d+) con "([^"]*)", "([^"]*)", "([^"]*)", "([^"]*)" y "([^"]*)"$`, func(codigo int, a, b, c, d, f string) error {
		return e.elJSONResponde(codigo, a, b, c, d, f)
	})
	sc.Step(`^el JSON dice que el burndown es del "([^"]*)" y su último día no tiene SP pendientes$`, e.elJSONTieneElBurndown)
	sc.Step(`^el JSON no tiene burndown$`, func() error {
		if b, ok := e.json["burndown"]; !ok || b != nil {
			return fmt.Errorf("burndown = %v, se esperaba null", b)
		}
		return nil
	})
	sc.Step(`^el JSON responde (\d+) con el error "([^"]*)"$`, func(codigo int, mensaje string) error {
		if e.codigo != codigo || e.json["error"] != mensaje {
			return fmt.Errorf("respondió %d con %v, se esperaba %d con %q", e.codigo, e.json, codigo, mensaje)
		}
		return nil
	})
}

func (e *hu36) elProyecto(nombre string) error {
	p, err := e.repoProyectos.Guardar(context.Background(), project.Proyecto{Nombre: nombre, FechaInicio: fechaDeEscenario("2026-10-05"), FechaFin: fechaDeEscenario("2026-11-30")})
	e.proyectoID = p.ID
	return err
}

func (e *hu36) elProyectoConSprints(nombre string, tabla *godog.Table) error {
	if err := e.elProyecto(nombre); err != nil {
		return err
	}
	for _, fila := range tabla.Rows[1:] {
		numero, err := strconv.Atoi(fila.Cells[0].Value)
		if err != nil {
			return err
		}
		s, err := e.repoSprints.Guardar(context.Background(), sprint.Sprint{ProyectoID: e.proyectoID, Numero: numero, Objetivo: fila.Cells[1].Value,
			FechaInicio: fechaDeEscenario(fila.Cells[2].Value), FechaFin: fechaDeEscenario(fila.Cells[3].Value), Estado: sprint.EstadoPlanificado})
		if err != nil {
			return err
		}
		e.sprintIDs[numero], e.fines[numero] = s.ID, s.FechaFin
	}
	return nil
}

// estasHistorias guarda cada historia ("HU-1 Login") en su sprint, Hecha el día de la columna "completada el" (o En
// sprint si está vacía), con una tarea de las horas estimadas y un registro de las reales.
func (e *hu36) estasHistorias(tabla *godog.Table) error {
	ctx := context.Background()
	for _, fila := range tabla.Rows[1:] {
		c := fila.Cells
		id, titulo, _ := strings.Cut(c[0].Value, " ")
		n, err := numeroDe(id)
		if err != nil {
			return err
		}
		numeroSprint, err := strconv.Atoi(c[1].Value)
		if err != nil {
			return err
		}
		sp, err := strconv.Atoi(c[2].Value)
		if err != nil {
			return err
		}
		sprintID := e.sprintIDs[numeroSprint]
		h := backlog.Historia{ProyectoID: e.proyectoID, Numero: n, Titulo: titulo, Prioridad: backlog.PrioridadAlta, Estado: backlog.EstadoEnSprint, StoryPoints: &sp, SprintID: &sprintID}
		if c[3].Value != "" {
			fecha := fechaDeEscenario(c[3].Value)
			h.Estado, h.CompletadaEn = backlog.EstadoHecho, &fecha
		}
		guardada, err := e.repoHistorias.Guardar(ctx, h)
		if err != nil {
			return err
		}
		if horas, err := effort.ParsearHoras(c[4].Value); err == nil && horas > 0 {
			if _, err := e.repoTareas.Guardar(ctx, effort.Tarea{HistoriaID: guardada.ID, Titulo: "Desarrollo", HorasEstimadas: horas}); err != nil {
				return err
			}
		}
		if horas, err := effort.ParsearHoras(c[5].Value); err == nil && horas > 0 {
			if _, err := e.repoRegistros.Guardar(ctx, effort.Registro{HistoriaID: guardada.ID, Fecha: fechaDeEscenario("2026-10-09"), Actividad: "Desarrollo", Horas: horas}); err != nil {
				return err
			}
		}
	}
	return nil
}

// seCerro cierra el sprint el día de su fecha de fin, con el caso de uso de HU-14.
func (e *hu36) seCerro(numero int) error {
	e.ahora = e.fines[numero]
	_, err := e.cierres.Cerrar(context.Background(), e.proyectoID, numero)
	return err
}

// estosDefectos guarda los defectos: Abiertos, o Resueltos si tienen sprint de resolución.
func (e *hu36) estosDefectos(tabla *godog.Table) error {
	for _, fila := range tabla.Rows[1:] {
		c := fila.Cells
		detectado, err := strconv.Atoi(c[2].Value)
		if err != nil {
			return err
		}
		d := defect.Defecto{ProyectoID: e.proyectoID, Descripcion: c[0].Value, Severidad: defect.Severidad(c[1].Value),
			Estado: defect.EstadoAbierto, SprintDeteccionID: e.sprintIDs[detectado]}
		if resuelto, err := strconv.Atoi(c[3].Value); err == nil {
			d.Estado, d.SprintResolucionID = defect.EstadoResuelto, e.sprintIDs[resuelto]
		}
		if _, err := e.repoDefectos.Guardar(context.Background(), d); err != nil {
			return err
		}
	}
	return nil
}

func (e *hu36) pedirDatos(numeroSprint int) error {
	var err error
	e.datos, err = e.graficos.Datos(context.Background(), e.proyectoID, numeroSprint)
	return err
}

// compararTabla compara las filas de un gráfico con la tabla del escenario (la primera columna es el nombre).
func compararTabla(tabla *godog.Table, filas []service.FilaGrafico) error {
	esperadas := tabla.Rows[1:]
	if len(filas) != len(esperadas) {
		return fmt.Errorf("el gráfico tiene %d filas (%+v), se esperaban %d", len(filas), filas, len(esperadas))
	}
	for i, fila := range esperadas {
		obtenida := append([]string{filas[i].Nombre}, filas[i].Valores...)
		for j, celda := range fila.Cells {
			if j >= len(obtenida) || obtenida[j] != celda.Value {
				return fmt.Errorf("fila %d = %v, se esperaba la columna %d en %q", i+1, obtenida, j+1, celda.Value)
			}
		}
	}
	return nil
}

func (e *hu36) elBurndown(nombre string, tabla *godog.Table) error {
	if e.datos.Burndown == nil || e.datos.Burndown.Sprint != nombre {
		return fmt.Errorf("el burndown es %+v, se esperaba el del %s", e.datos.Burndown, nombre)
	}
	return compararTabla(tabla, e.datos.FilasBurndown())
}

func (e *hu36) pedirJSON(proyectoID int64) error {
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, fmt.Sprintf("/proyectos/%d/graficos/datos", proyectoID), nil))
	e.codigo, e.json = rec.Code, nil
	return json.Unmarshal(rec.Body.Bytes(), &e.json)
}

func (e *hu36) elJSONResponde(codigo int, claves ...string) error {
	if e.codigo != codigo {
		return fmt.Errorf("el JSON respondió %d, se esperaba %d", e.codigo, codigo)
	}
	for _, c := range claves {
		if _, ok := e.json[c]; !ok {
			return fmt.Errorf("el JSON no tiene %q: %v", c, e.json)
		}
	}
	return nil
}

func (e *hu36) elJSONTieneElBurndown(nombre string) error {
	b, ok := e.json["burndown"].(map[string]any)
	if !ok || b["sprint"] != nombre {
		return fmt.Errorf("burndown = %v, se esperaba el del %s", e.json["burndown"], nombre)
	}
	pendientes, ok := b["pendientes"].([]any)
	if !ok || len(pendientes) == 0 || pendientes[len(pendientes)-1] != nil {
		return fmt.Errorf("pendientes = %v, el último día no tendría que tener dato", b["pendientes"])
	}
	return nil
}
