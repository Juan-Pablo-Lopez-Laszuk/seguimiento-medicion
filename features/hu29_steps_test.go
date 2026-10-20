package features

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/defect"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// hu29 guarda el estado de un escenario de HU-29 (consultar defectos), con sus propios repositorios en memoria.
type hu29 struct {
	proyectos *service.Proyectos
	sprints   *service.Sprints
	defectos  *service.Defectos
	proyecto  int64
	numeros   map[string]int64 // "Sprint N" → ID
	vista     service.VistaDefectos
}

// registrarPasosHU29 conecta cada frase de features/hu-29-consultar-defectos.feature con una función de Go.
func registrarPasosHU29(sc *godog.ScenarioContext) {
	repoProyectos, repoSprints := memory.NuevoProyectos(), memory.NuevoSprints()
	e := &hu29{
		proyectos: service.NuevoProyectos(repoProyectos, time.Now),
		sprints:   service.NuevoSprints(repoProyectos, repoSprints),
		defectos:  service.NuevoDefectos(repoProyectos, repoSprints, memory.NuevoHistorias(), memory.NuevoDefectos()),
		numeros:   map[string]int64{},
	}
	sc.Step(`^el proyecto con los sprints 1 y 2 y estos defectos para consultar:$`, e.preparar)
	sc.Step(`^consulto los defectos de severidad "([^"]*)", estado "([^"]*)" y sprint "([^"]*)"$`, func(sev, est, sp string) error {
		return e.consultar(defect.Filtro{Severidad: defect.Severidad(sev), Estado: defect.Estado(est), SprintID: e.numeros[sp]})
	})
	sc.Step(`^consulto los defectos de estado "([^"]*)"$`, func(est string) error {
		return e.consultar(defect.Filtro{Estado: defect.Estado(est)})
	})
	sc.Step(`^consulto los defectos de severidad "([^"]*)"$`, func(sev string) error {
		return e.consultar(defect.Filtro{Severidad: defect.Severidad(sev)})
	})
	sc.Step(`^veo solo los defectos "([^"]*)"$`, e.veoSolo)
	sc.Step(`^no veo ningún defecto$`, func() error { return e.veoSolo("") })
	sc.Step(`^veo los (\d+) defectos$`, func(n int) error {
		if len(e.vista.Filtrados) != n {
			return fmt.Errorf("se ven %d defectos, se esperaban %d", len(e.vista.Filtrados), n)
		}
		return nil
	})
	sc.Step(`^la cantidad por severidad es Crítica (\d+), Alta (\d+), Media (\d+) y Baja (\d+)$`, func(c, a, m, b int) error {
		var got []int
		for _, s := range e.vista.PorSeveridad() {
			got = append(got, s.Cantidad)
		}
		if fmt.Sprint(got) != fmt.Sprint([]int{c, a, m, b}) {
			return fmt.Errorf("la cantidad por severidad es %v", got)
		}
		return nil
	})
}

func (e *hu29) preparar(tabla *godog.Table) error {
	ctx := context.Background()
	p, err := e.proyectos.Crear(ctx, project.Datos{Nombre: "Tienda", FechaInicio: fechaDeEscenario("2026-10-05"), FechaFin: fechaDeEscenario("2026-11-01")})
	if err != nil {
		return fmt.Errorf("no se pudo crear el proyecto: %w", err)
	}
	e.proyecto = p.ID
	for _, f := range [][2]string{{"2026-10-05", "2026-10-11"}, {"2026-10-12", "2026-10-18"}} {
		s, err := e.sprints.Crear(ctx, p.ID, sprint.Datos{Objetivo: "MVP", FechaInicio: fechaDeEscenario(f[0]), FechaFin: fechaDeEscenario(f[1])})
		if err != nil {
			return fmt.Errorf("no se pudo crear el sprint: %w", err)
		}
		e.numeros["Sprint "+strconv.Itoa(s.Numero)] = s.ID
	}
	for _, fila := range tabla.Rows[1:] {
		desc, sev, detectado, estado := fila.Cells[0].Value, fila.Cells[1].Value, fila.Cells[2].Value, fila.Cells[3].Value
		d, err := e.defectos.Registrar(ctx, p.ID, defect.Datos{Descripcion: desc, Severidad: defect.Severidad(sev), SprintID: e.numeros[detectado]})
		if err != nil {
			return fmt.Errorf("no se pudo registrar %q: %w", desc, err)
		}
		if estado == string(defect.EstadoEnProgreso) {
			if _, err := e.defectos.CambiarEstado(ctx, p.ID, d.ID, defect.EstadoEnProgreso, 0); err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *hu29) consultar(f defect.Filtro) (err error) {
	e.vista, err = e.defectos.Consultar(context.Background(), e.proyecto, f)
	return err
}

// veoSolo compara las descripciones de la lista filtrada, en orden, con las esperadas separadas por coma.
func (e *hu29) veoSolo(esperadas string) error {
	var got []string
	for _, d := range e.vista.Filtrados {
		got = append(got, d.Descripcion)
	}
	if strings.Join(got, ", ") != esperadas {
		return fmt.Errorf("se ven %q, se esperaba %q", strings.Join(got, ", "), esperadas)
	}
	return nil
}
