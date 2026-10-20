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

// hu34 guarda el estado de un escenario de HU-34 (defectos detectados y resueltos), con sus propios repositorios.
type hu34 struct {
	proyectos *service.Proyectos
	sprints   *service.Sprints
	defectos  *service.Defectos
	proyecto  int64
	numeros   map[string]int64 // "Sprint N" → ID
	resumen   service.ResumenDefectos
}

// caminos son los cambios de estado que llevan un defecto Abierto a cada estado de la tabla.
var caminos = map[string][]defect.Estado{
	"Abierto":   nil,
	"Resuelto":  {defect.EstadoEnProgreso, defect.EstadoResuelto},
	"Cerrado":   {defect.EstadoEnProgreso, defect.EstadoResuelto, defect.EstadoCerrado},
	"Reabierto": {defect.EstadoEnProgreso, defect.EstadoResuelto, defect.EstadoReabierto},
}

// registrarPasosHU34 conecta cada frase de features/hu-34-defectos.feature con una función de Go.
func registrarPasosHU34(sc *godog.ScenarioContext) {
	repoProyectos, repoSprints := memory.NuevoProyectos(), memory.NuevoSprints()
	e := &hu34{
		proyectos: service.NuevoProyectos(repoProyectos, time.Now),
		sprints:   service.NuevoSprints(repoProyectos, repoSprints),
		defectos:  service.NuevoDefectos(repoProyectos, repoSprints, memory.NuevoHistorias(), memory.NuevoDefectos()),
		numeros:   map[string]int64{},
	}
	sc.Step(`^el proyecto con los sprints 1 y 2 y estos defectos:$`, e.preparar)
	sc.Step(`^veo el resumen de defectos con todas las severidades$`, func() error { return e.ver("") })
	sc.Step(`^veo el resumen de defectos de severidad "([^"]*)"$`, e.ver)
	sc.Step(`^el Sprint (\d+) tiene (\d+) detectados? y (\d+) resueltos?$`, e.elSprintTiene)
	sc.Step(`^el proyecto tiene (\d+) detectados, (\d+) resueltos y (\d+) abiertos$`, e.elProyectoTiene)
}

func (e *hu34) preparar(tabla *godog.Table) error {
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
		sev, detectado, estado, resuelto := fila.Cells[0].Value, fila.Cells[1].Value, fila.Cells[2].Value, fila.Cells[3].Value
		d, err := e.defectos.Registrar(ctx, p.ID, defect.Datos{Descripcion: "Falla", Severidad: defect.Severidad(sev), SprintID: e.numeros[detectado]})
		if err != nil {
			return fmt.Errorf("no se pudo registrar el defecto: %w", err)
		}
		// Un reabierto se resuelve primero en su sprint de detección.
		resolucion := e.numeros[strings.TrimSpace(resuelto)]
		if resolucion == 0 {
			resolucion = e.numeros[detectado]
		}
		for _, paso := range caminos[estado] {
			if _, err := e.defectos.CambiarEstado(ctx, p.ID, d.ID, paso, resolucion); err != nil {
				return fmt.Errorf("no se pudo pasar el defecto a %s: %w", paso, err)
			}
		}
	}
	return nil
}

func (e *hu34) ver(severidad string) error {
	v, err := e.defectos.Ver(context.Background(), e.proyecto)
	if err != nil {
		return err
	}
	e.resumen = v.Resumen(defect.Severidad(severidad))
	return nil
}

func (e *hu34) elSprintTiene(n, detectados, resueltos int) error {
	nombre := "Sprint " + strconv.Itoa(n)
	for _, f := range e.resumen.Sprints {
		if f.Nombre == nombre {
			if f.Detectados != detectados || f.Resueltos != resueltos {
				return fmt.Errorf("el %s tiene %d detectados y %d resueltos", nombre, f.Detectados, f.Resueltos)
			}
			return nil
		}
	}
	return fmt.Errorf("el resumen no tiene el %s", nombre)
}

func (e *hu34) elProyectoTiene(detectados, resueltos, abiertos int) error {
	r := e.resumen
	if r.Detectados != detectados || r.Resueltos != resueltos || r.Abiertos != abiertos {
		return fmt.Errorf("el proyecto tiene %d detectados, %d resueltos y %d abiertos", r.Detectados, r.Resueltos, r.Abiertos)
	}
	return nil
}
