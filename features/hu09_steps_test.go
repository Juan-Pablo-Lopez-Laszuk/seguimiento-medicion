package features

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// hu09 guarda el estado de un escenario de HU-09 (priorizar el backlog), con sus propios repositorios en memoria. Las
// historias se crean con el caso de uso, así toman su orden como en la aplicación (RN6).
type hu09 struct {
	repoProyectos *memory.Proyectos
	repoHistorias *memory.Historias
	repoSprints   *memory.Sprints
	historias     *service.Historias
	edicion       *service.EdicionHistorias
	asignaciones  *service.Asignaciones
	proyectoID    int64
	err           error
}

// registrarPasosHU09 conecta cada frase de features/hu-09-priorizar-backlog.feature con una función de Go.
func registrarPasosHU09(sc *godog.ScenarioContext) {
	e := &hu09{repoProyectos: memory.NuevoProyectos(), repoHistorias: memory.NuevoHistorias(), repoSprints: memory.NuevoSprints()}
	repoCriterios := memory.NuevoCriterios()
	e.historias = service.NuevoHistorias(e.repoProyectos, e.repoHistorias)
	e.edicion = service.NuevoEdicionHistorias(e.repoProyectos, e.repoHistorias, repoCriterios, memory.NuevoTareas())
	criterios := service.NuevoCriterios(e.repoProyectos, e.repoHistorias, repoCriterios)
	e.asignaciones = service.NuevoAsignaciones(e.repoProyectos, e.repoSprints, e.repoHistorias, criterios, func() time.Time {
		return time.Date(2026, 10, 15, 18, 0, 0, 0, time.UTC)
	})

	sc.Step(`^el proyecto "([^"]*)" con estas historias para priorizar:$`, e.elProyectoConEstasHistorias)
	sc.Step(`^(?:que )?se agrega al backlog la historia "([^"]*)" con prioridad "([^"]*)"$`, e.seAgrega)
	sc.Step(`^(?:que ya )?(subo|bajo|subí) la historia "([^"]*)"$`, e.muevo)
	sc.Step(`^intento (subir|bajar) la historia "([^"]*)"$`, e.intentoMover)
	sc.Step(`^intento subir la historia número (\d+)$`, e.intentoSubirElNumero)
	sc.Step(`^el movimiento falla con "([^"]*)"$`, e.elMovimientoFallaCon)
	sc.Step(`^el backlog (?:queda|sigue) ordenado así:$`, e.elBacklogQuedaOrdenado)
	sc.Step(`^al volver a cargar el backlog desde lo guardado queda ordenado así:$`, e.alVolverACargar)
	sc.Step(`^cambio la prioridad de "([^"]*)" a "([^"]*)"$`, e.cambioLaPrioridad)
	sc.Step(`^que "([^"]*)" está en el estado "([^"]*)"$`, e.estaEnElEstado)
	sc.Step(`^que "([^"]*)" está Hecha desde el "([^"]*)"$`, e.estaHechaDesde)
	sc.Step(`^las historias Hechas son:$`, e.lasHechasSon)
	sc.Step(`^que existe el Sprint (\d+) del proyecto$`, e.existeElSprint)
	sc.Step(`^las candidatas del Sprint (\d+) están en este orden:$`, e.lasCandidatasEstanEnOrden)
}

func (e *hu09) elProyectoConEstasHistorias(nombre string, tabla *godog.Table) error {
	p, err := e.repoProyectos.Guardar(context.Background(), project.Proyecto{
		Nombre: nombre, FechaInicio: fechaDeEscenario("2026-10-05"), FechaFin: fechaDeEscenario("2026-11-01"),
	})
	if err != nil {
		return err
	}
	e.proyectoID = p.ID
	for _, fila := range tabla.Rows[1:] {
		titulo, prioridad, estado := fila.Cells[0].Value, fila.Cells[1].Value, fila.Cells[2].Value
		if err := e.seAgrega(titulo, prioridad); err != nil {
			return err
		}
		if estado != string(backlog.EstadoPendiente) {
			if err := e.estaEnElEstado(titulo, estado); err != nil {
				return err
			}
		}
	}
	return nil
}

func (e *hu09) seAgrega(titulo, prioridad string) error {
	_, err := e.historias.Crear(context.Background(), e.proyectoID, backlog.Datos{Titulo: titulo, Prioridad: backlog.Prioridad(prioridad)})
	return err
}

// historia busca en el backlog la historia con ese título.
func (e *hu09) historia(titulo string) (backlog.Historia, error) {
	lista, err := e.repoHistorias.ListarPorProyecto(context.Background(), e.proyectoID)
	if err != nil {
		return backlog.Historia{}, err
	}
	i := slices.IndexFunc(lista, func(h backlog.Historia) bool { return h.Titulo == titulo })
	if i < 0 {
		return backlog.Historia{}, fmt.Errorf("no hay ninguna historia %q en el backlog", titulo)
	}
	return lista[i], nil
}

// movimientoDe traduce el verbo del escenario: "subo", "subí" y "subir" suben; "bajo" y "bajar" bajan.
func movimientoDe(verbo string) backlog.Movimiento {
	if verbo == "bajo" || verbo == "bajar" {
		return backlog.Bajar
	}
	return backlog.Subir
}

func (e *hu09) mover(numero int, m backlog.Movimiento) error {
	_, err := e.historias.Mover(context.Background(), e.proyectoID, numero, m)
	return err
}

func (e *hu09) muevo(verbo, titulo string) error {
	h, err := e.historia(titulo)
	if err != nil {
		return err
	}
	return e.mover(h.Numero, movimientoDe(verbo))
}

func (e *hu09) intentoMover(verbo, titulo string) error {
	h, err := e.historia(titulo)
	if err != nil {
		return err
	}
	e.err = e.mover(h.Numero, movimientoDe(verbo))
	return nil
}

func (e *hu09) intentoSubirElNumero(numero int) error {
	e.err = e.mover(numero, backlog.Subir)
	return nil
}

func (e *hu09) elMovimientoFallaCon(mensaje string) error {
	if e.err == nil || e.err.Error() != mensaje {
		return fmt.Errorf("se esperaba el error %q y se obtuvo %v", mensaje, e.err)
	}
	return nil
}

// verificarTitulos compara los títulos de la lista con la columna "título" de la tabla, en orden.
func verificarTitulos(que string, lista []backlog.Historia, tabla *godog.Table) error {
	var esperados, obtenidos []string
	for _, fila := range tabla.Rows[1:] {
		esperados = append(esperados, fila.Cells[0].Value)
	}
	for _, h := range lista {
		obtenidos = append(obtenidos, h.Titulo)
	}
	if !slices.Equal(obtenidos, esperados) {
		return fmt.Errorf("%s: se esperaba %v y se obtuvo %v", que, esperados, obtenidos)
	}
	return nil
}

func (e *hu09) elBacklogQuedaOrdenado(tabla *godog.Table) error {
	v, err := e.historias.Backlog(context.Background(), e.proyectoID)
	if err != nil {
		return err
	}
	return verificarTitulos("backlog", v.Pendientes, tabla)
}

// alVolverACargar lee el backlog con un caso de uso nuevo sobre los mismos repositorios, como al volver a abrir la
// aplicación: el orden sale de lo guardado (RN5).
func (e *hu09) alVolverACargar(tabla *godog.Table) error {
	v, err := service.NuevoHistorias(e.repoProyectos, e.repoHistorias).Backlog(context.Background(), e.proyectoID)
	if err != nil {
		return err
	}
	return verificarTitulos("backlog guardado", v.Pendientes, tabla)
}

func (e *hu09) cambioLaPrioridad(titulo, prioridad string) error {
	h, err := e.historia(titulo)
	if err != nil {
		return err
	}
	_, err = e.edicion.Editar(context.Background(), e.proyectoID, h.Numero,
		backlog.Cambios{Titulo: h.Titulo, Descripcion: h.Descripcion, Prioridad: backlog.Prioridad(prioridad)})
	return err
}

func (e *hu09) estaEnElEstado(titulo, estado string) error {
	h, err := e.historia(titulo)
	if err != nil {
		return err
	}
	return e.repoHistorias.ActualizarEstado(context.Background(), h.ID, backlog.Estado(estado))
}

func (e *hu09) estaHechaDesde(titulo, fecha string) error {
	h, err := e.historia(titulo)
	if err != nil {
		return err
	}
	dia := fechaDeEscenario(fecha)
	return e.repoHistorias.ActualizarCompletada(context.Background(), h.ID, backlog.EstadoHecho, &dia)
}

func (e *hu09) lasHechasSon(tabla *godog.Table) error {
	v, err := e.historias.Backlog(context.Background(), e.proyectoID)
	if err != nil {
		return err
	}
	return verificarTitulos("Hechas", v.Hechas, tabla)
}

func (e *hu09) existeElSprint(numero int) error {
	inicio := fechaDeEscenario("2026-10-12")
	_, err := e.repoSprints.Guardar(context.Background(), sprint.Sprint{
		ProyectoID: e.proyectoID, Numero: numero, Objetivo: "Primer incremento",
		FechaInicio: inicio, FechaFin: inicio.AddDate(0, 0, 6), Estado: sprint.EstadoPlanificado,
	})
	return err
}

func (e *hu09) lasCandidatasEstanEnOrden(numero int, tabla *godog.Table) error {
	v, err := e.asignaciones.Ver(context.Background(), e.proyectoID, numero)
	if err != nil {
		return err
	}
	var candidatas []backlog.Historia
	for _, c := range v.Candidatas {
		candidatas = append(candidatas, c.Historia)
	}
	return verificarTitulos(fmt.Sprintf("candidatas del Sprint %d", numero), candidatas, tabla)
}
