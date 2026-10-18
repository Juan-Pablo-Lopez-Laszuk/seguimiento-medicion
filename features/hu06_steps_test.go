package features

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// historiasConSprintCerrado es el repositorio de historias en memoria con unas historias marcadas como "de un sprint
// cerrado": en memoria todavía no se pueden asignar a un sprint (HU-12), así que el escenario lo simula.
type historiasConSprintCerrado struct {
	*memory.Historias
	cerradas map[int64]bool
}

func (r historiasConSprintCerrado) EnSprintCerrado(_ context.Context, historiaID int64) (bool, error) {
	return r.cerradas[historiaID], nil
}

// hu06 guarda el estado de un escenario de HU-06 (editar y eliminar un ítem del backlog), con sus propios repositorios
// en memoria. "La historia" de todos los pasos es la del Antecedentes.
type hu06 struct {
	repoHistorias *memory.Historias
	repoCriterios *memory.Criterios
	repoTareas    *memory.Tareas
	cerradas      map[int64]bool
	historias     *service.Historias
	casos         *service.EdicionHistorias
	proyectoID    int64
	historia      backlog.Historia
	errEdicion    error
	errEliminar   error
	pedido        service.VistaEliminacion
}

// registrarPasosHU06 conecta cada frase de features/hu-06-editar-eliminar-item.feature con una función de Go.
func registrarPasosHU06(sc *godog.ScenarioContext) {
	repoProyectos := memory.NuevoProyectos()
	e := &hu06{cerradas: map[int64]bool{}}
	e.repoCriterios, e.repoTareas = memory.NuevoCriterios(), memory.NuevoTareas()
	e.repoHistorias = memory.NuevoHistorias(e.repoCriterios, e.repoTareas)
	e.historias = service.NuevoHistorias(repoProyectos, e.repoHistorias)
	e.casos = service.NuevoEdicionHistorias(repoProyectos, historiasConSprintCerrado{Historias: e.repoHistorias, cerradas: e.cerradas}, e.repoCriterios, e.repoTareas)
	proyectos := service.NuevoProyectos(repoProyectos, time.Now)

	sc.Step(`^el proyecto "([^"]*)" con la historia "([^"]*)" de (\d+) story points para editar o eliminar$`, func(nombre, titulo string, puntos int) error {
		return e.elProyectoConLaHistoria(proyectos, nombre, titulo, puntos)
	})
	sc.Step(`^que la historia a editar tiene (\d+) criterios?, (\d+) tareas? y está "([^"]*)"$`, e.queLaHistoriaTiene)
	sc.Step(`^que la historia a editar está "([^"]*)"$`, func(estado string) error { return e.queLaHistoriaTiene(0, 0, estado) })
	sc.Step(`^que la historia a editar es de un sprint cerrado$`, e.queLaHistoriaEsDeUnSprintCerrado)
	sc.Step(`^que el proyecto tiene además las historias "([^"]*)" y "([^"]*)"$`, e.queElProyectoTieneAdemas)

	sc.Step(`^edito la historia con título "([^"]*)", descripción "([^"]*)" y prioridad "([^"]*)"$`, e.editoLaHistoria)
	sc.Step(`^edito la historia con (título|descripción) de (\d+) caracteres$`, e.editoLaHistoriaConUnCampoDeLargo)
	sc.Step(`^pido eliminar la historia$`, e.pidoEliminarLaHistoria)
	sc.Step(`^confirmo la eliminación de la historia$`, func() error { return e.confirmoLaEliminacion(e.historia.Numero) })
	sc.Step(`^confirmo la eliminación de la historia "([^"]*)"$`, e.confirmoLaEliminacionDe)

	sc.Step(`^la edición da como resultado "([^"]*)"$`, e.laEdicionDaComoResultado)
	sc.Step(`^la historia es HU-(\d+) con título "([^"]*)", descripción "([^"]*)" y prioridad "([^"]*)"$`, e.laHistoriaEs)
	sc.Step(`^la historia conserva (\d+) story points, (\d+) criterios?, (\d+) tareas? y el estado "([^"]*)"$`, e.laHistoriaConserva)
	sc.Step(`^el sistema pide confirmar la eliminación de (\d+) criterios? y (\d+) tareas?$`, e.elSistemaPideConfirmar)
	sc.Step(`^el sistema no deja eliminarla: "([^"]*)"$`, e.elSistemaNoDejaEliminarla)
	sc.Step(`^la eliminación da como resultado "([^"]*)"$`, e.laEliminacionDaComoResultado)
	sc.Step(`^la historia sigue en el backlog$`, e.laHistoriaSigueEnElBacklog)
	sc.Step(`^la historia ya no está en el backlog$`, e.laHistoriaYaNoEstaEnElBacklog)
	sc.Step(`^no quedan criterios ni tareas de la historia$`, e.noQuedanCriteriosNiTareas)
	sc.Step(`^las historias del proyecto son "([^"]*)"$`, e.lasHistoriasDelProyectoSon)
	sc.Step(`^al crear una historia nueva es la HU-(\d+)$`, e.alCrearUnaHistoriaNuevaEs)
}

func (e *hu06) elProyectoConLaHistoria(proyectos *service.Proyectos, nombre, titulo string, puntos int) error {
	p, err := proyectos.Crear(context.Background(), project.Datos{
		Nombre:      nombre,
		FechaInicio: fechaDeEscenario("2026-10-05"),
		FechaFin:    fechaDeEscenario("2026-11-01"),
	})
	if err != nil {
		return fmt.Errorf("no se pudo crear el proyecto %q: %w", nombre, err)
	}
	e.proyectoID = p.ID
	e.historia, err = e.historias.Crear(context.Background(), p.ID, backlog.Datos{
		Titulo: titulo, Prioridad: backlog.PrioridadAlta, StoryPoints: fmt.Sprint(puntos),
	})
	if err != nil {
		return fmt.Errorf("no se pudo crear la historia %q: %w", titulo, err)
	}
	return nil
}

// queLaHistoriaTiene deja la historia en ese estado y con esos criterios y tareas, directo en los repositorios: los
// cambios de estado de HU-12 y HU-13 todavía no existen, y varios escenarios arrancan desde esos estados.
func (e *hu06) queLaHistoriaTiene(criterios, tareas int, estado string) error {
	ctx := context.Background()
	if err := e.repoHistorias.ActualizarEstado(ctx, e.historia.ID, backlog.Estado(estado)); err != nil {
		return fmt.Errorf("no se pudo dejar la historia en %q: %w", estado, err)
	}
	for i := 1; i <= criterios; i++ {
		if _, err := e.repoCriterios.Guardar(ctx, backlog.Criterio{HistoriaID: e.historia.ID, Descripcion: fmt.Sprintf("Criterio %d", i)}); err != nil {
			return fmt.Errorf("no se pudo guardar el criterio %d: %w", i, err)
		}
	}
	for i := 1; i <= tareas; i++ {
		if _, err := e.repoTareas.Guardar(ctx, effort.Tarea{HistoriaID: e.historia.ID, Titulo: fmt.Sprintf("Tarea %d", i)}); err != nil {
			return fmt.Errorf("no se pudo guardar la tarea %d: %w", i, err)
		}
	}
	return nil
}

func (e *hu06) queLaHistoriaEsDeUnSprintCerrado() error {
	e.cerradas[e.historia.ID] = true
	return nil
}

func (e *hu06) queElProyectoTieneAdemas(segunda, tercera string) error {
	for _, titulo := range []string{segunda, tercera} {
		if _, err := e.historias.Crear(context.Background(), e.proyectoID, backlog.Datos{Titulo: titulo, Prioridad: backlog.PrioridadMedia}); err != nil {
			return fmt.Errorf("no se pudo crear la historia %q: %w", titulo, err)
		}
	}
	return nil
}

func (e *hu06) editoLaHistoria(titulo, descripcion, prioridad string) error {
	_, e.errEdicion = e.casos.Editar(context.Background(), e.proyectoID, e.historia.Numero, backlog.Cambios{
		Titulo: titulo, Descripcion: descripcion, Prioridad: backlog.Prioridad(prioridad),
	})
	return nil
}

func (e *hu06) editoLaHistoriaConUnCampoDeLargo(campo string, largo int) error {
	cambios := backlog.Cambios{Titulo: "Título válido", Prioridad: backlog.PrioridadAlta}
	if campo == "título" {
		cambios.Titulo = strings.Repeat("a", largo)
	} else {
		cambios.Descripcion = strings.Repeat("a", largo)
	}
	_, e.errEdicion = e.casos.Editar(context.Background(), e.proyectoID, e.historia.Numero, cambios)
	return nil
}

func (e *hu06) pidoEliminarLaHistoria() error {
	var err error
	e.pedido, err = e.casos.PedirEliminacion(context.Background(), e.proyectoID, e.historia.Numero)
	return err
}

func (e *hu06) confirmoLaEliminacion(numero int) error {
	_, e.errEliminar = e.casos.Eliminar(context.Background(), e.proyectoID, numero)
	return nil
}

func (e *hu06) confirmoLaEliminacionDe(titulo string) error {
	_, lista, err := e.historias.Listar(context.Background(), e.proyectoID)
	if err != nil {
		return err
	}
	for _, h := range lista {
		if h.Titulo == titulo {
			return e.confirmoLaEliminacion(h.Numero)
		}
	}
	return fmt.Errorf("no hay una historia %q en el proyecto", titulo)
}

// laEdicionDaComoResultado acepta "ok" o "campo: mensaje de error".
func (e *hu06) laEdicionDaComoResultado(resultado string) error {
	if resultado == "ok" {
		if e.errEdicion != nil {
			return fmt.Errorf("se esperaba que la edición valga y se obtuvo: %w", e.errEdicion)
		}
		return nil
	}
	campo, mensaje, ok := strings.Cut(resultado, ": ")
	if !ok {
		return fmt.Errorf("resultado mal escrito en el escenario: %q (se espera \"ok\" o \"campo: mensaje\")", resultado)
	}
	var errs domain.ErroresValidacion
	if !errors.As(e.errEdicion, &errs) {
		return fmt.Errorf("se esperaban errores de validación y se obtuvo: %v", e.errEdicion)
	}
	if errs[campo] == nil || errs[campo].Error() != mensaje {
		return fmt.Errorf("campo %q: se esperaba %q y se obtuvo %v", campo, mensaje, errs[campo])
	}
	return nil
}

func (e *hu06) historiaGuardada() (backlog.Historia, error) {
	return e.repoHistorias.BuscarPorNumero(context.Background(), e.proyectoID, e.historia.Numero)
}

func (e *hu06) laHistoriaEs(numero int, titulo, descripcion, prioridad string) error {
	h, err := e.historiaGuardada()
	if err != nil {
		return err
	}
	if h.Numero != numero || h.Titulo != titulo || h.Descripcion != descripcion || string(h.Prioridad) != prioridad {
		return fmt.Errorf("la historia es HU-%d %q / %q / %q; se esperaba HU-%d %q / %q / %q",
			h.Numero, h.Titulo, h.Descripcion, h.Prioridad, numero, titulo, descripcion, prioridad)
	}
	return nil
}

func (e *hu06) laHistoriaConserva(puntos, criterios, tareas int, estado string) error {
	h, err := e.historiaGuardada()
	if err != nil {
		return err
	}
	if h.StoryPoints == nil || *h.StoryPoints != puntos {
		return fmt.Errorf("story points: se esperaba %d y se obtuvo %v", puntos, h.StoryPoints)
	}
	if string(h.Estado) != estado {
		return fmt.Errorf("estado: se esperaba %q y se obtuvo %q", estado, h.Estado)
	}
	cs, _ := e.repoCriterios.ListarPorHistoria(context.Background(), h.ID)
	ts, _ := e.repoTareas.ListarPorHistoria(context.Background(), h.ID)
	if len(cs) != criterios || len(ts) != tareas {
		return fmt.Errorf("se esperaban %d criterios y %d tareas y hay %d y %d", criterios, tareas, len(cs), len(ts))
	}
	return nil
}

func (e *hu06) elSistemaPideConfirmar(criterios, tareas int) error {
	if e.pedido.Criterios != criterios || e.pedido.Tareas != tareas {
		return fmt.Errorf("se esperaba avisar de %d criterios y %d tareas y se avisó de %d y %d",
			criterios, tareas, e.pedido.Criterios, e.pedido.Tareas)
	}
	if e.pedido.Impedimento != nil {
		return fmt.Errorf("se esperaba poder eliminar y se obtuvo: %w", e.pedido.Impedimento)
	}
	return nil
}

func (e *hu06) elSistemaNoDejaEliminarla(motivo string) error {
	if e.pedido.Impedimento == nil || e.pedido.Impedimento.Error() != motivo {
		return fmt.Errorf("se esperaba el impedimento %q y se obtuvo: %v", motivo, e.pedido.Impedimento)
	}
	return nil
}

// laEliminacionDaComoResultado acepta "ok" o el mensaje del error.
func (e *hu06) laEliminacionDaComoResultado(resultado string) error {
	if resultado == "ok" {
		if e.errEliminar != nil {
			return fmt.Errorf("se esperaba que se elimine y se obtuvo: %w", e.errEliminar)
		}
		return nil
	}
	if e.errEliminar == nil || e.errEliminar.Error() != resultado {
		return fmt.Errorf("se esperaba el error %q y se obtuvo: %v", resultado, e.errEliminar)
	}
	return nil
}

func (e *hu06) laHistoriaSigueEnElBacklog() error {
	if _, err := e.historiaGuardada(); err != nil {
		return fmt.Errorf("la historia tenía que seguir en el backlog: %w", err)
	}
	return nil
}

func (e *hu06) laHistoriaYaNoEstaEnElBacklog() error {
	if _, err := e.historiaGuardada(); !errors.Is(err, backlog.ErrHistoriaNoEncontrada) {
		return fmt.Errorf("la historia tenía que no estar más en el backlog y se obtuvo: %v", err)
	}
	return nil
}

func (e *hu06) noQuedanCriteriosNiTareas() error {
	cs, _ := e.repoCriterios.ListarPorHistoria(context.Background(), e.historia.ID)
	ts, _ := e.repoTareas.ListarPorHistoria(context.Background(), e.historia.ID)
	if len(cs) != 0 || len(ts) != 0 {
		return fmt.Errorf("quedaron %d criterios y %d tareas de la historia eliminada", len(cs), len(ts))
	}
	return nil
}

func (e *hu06) lasHistoriasDelProyectoSon(esperadas string) error {
	_, lista, err := e.historias.Listar(context.Background(), e.proyectoID)
	if err != nil {
		return err
	}
	identificadores := make([]string, 0, len(lista))
	for _, h := range lista {
		identificadores = append(identificadores, h.Identificador())
	}
	if obtenidas := strings.Join(identificadores, ", "); obtenidas != esperadas {
		return fmt.Errorf("las historias del proyecto son %q y se esperaban %q", obtenidas, esperadas)
	}
	return nil
}

func (e *hu06) alCrearUnaHistoriaNuevaEs(numero int) error {
	nueva, err := e.historias.Crear(context.Background(), e.proyectoID, backlog.Datos{Titulo: "Una historia nueva", Prioridad: backlog.PrioridadBaja})
	if err != nil {
		return err
	}
	if nueva.Numero != numero {
		return fmt.Errorf("la historia nueva es HU-%d y se esperaba HU-%d", nueva.Numero, numero)
	}
	return nil
}
