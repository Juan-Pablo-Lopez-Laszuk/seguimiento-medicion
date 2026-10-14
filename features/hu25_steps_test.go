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
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// hu25 guarda el estado de un escenario de HU-25 (descomponer una historia en tareas), con sus propios repositorios en
// memoria. "La historia" de los pasos que no nombran ninguna es la del Antecedentes; las demás se buscan por título.
type hu25 struct {
	proyectos   *service.Proyectos
	historias   *service.Historias
	integrantes *service.Integrantes
	tareas      *service.Tareas
	proyectoID  int64
	principal   backlog.Historia
	porTitulo   map[string]backlog.Historia
	porNombre   map[string]member.Integrante
	tareaPorTit map[string]effort.Tarea
	err         error
}

// registrarPasosHU25 conecta cada frase de features/hu-25-descomponer-tareas.feature con una función de Go.
func registrarPasosHU25(sc *godog.ScenarioContext) {
	repoProyectos, repoHistorias, repoIntegrantes := memory.NuevoProyectos(), memory.NuevoHistorias(), memory.NuevoIntegrantes()
	e := &hu25{
		proyectos:   service.NuevoProyectos(repoProyectos, time.Now),
		historias:   service.NuevoHistorias(repoProyectos, repoHistorias),
		integrantes: service.NuevoIntegrantes(repoProyectos, repoIntegrantes),
		tareas:      service.NuevoTareas(repoProyectos, repoHistorias, repoIntegrantes, memory.NuevoTareas()),
		porTitulo:   map[string]backlog.Historia{},
		porNombre:   map[string]member.Integrante{},
		tareaPorTit: map[string]effort.Tarea{},
	}

	sc.Step(`^el proyecto "([^"]*)" con la historia "([^"]*)" y los integrantes "([^"]*)"$`, e.elProyectoConLaHistoriaYLosIntegrantes)
	sc.Step(`^que "([^"]*)" está dado de baja$`, e.queEstaDadoDeBaja)
	sc.Step(`^que la historia tiene la tarea "([^"]*)" de "([^"]*)" con "([^"]*)" horas$`, func(titulo, responsable, horas string) error {
		return e.tareaPrevia(e.principal, titulo, responsable, horas)
	})
	sc.Step(`^la historia "([^"]*)" con la tarea "([^"]*)" de "([^"]*)" con "([^"]*)" horas$`, e.otraHistoriaConLaTarea)
	sc.Step(`^que la tarea "([^"]*)" está completada$`, func(titulo string) error { return e.marcar(titulo, true) })

	sc.Step(`^agrego la tarea "([^"]*)" para "([^"]*)" con "([^"]*)" horas$`, func(titulo, responsable, horas string) error {
		e.agregar(e.principal, titulo, responsable, horas)
		return nil
	})
	sc.Step(`^agrego a la historia una tarea con un título de (\d+) caracteres$`, func(largo int) error {
		e.agregar(e.principal, strings.Repeat("a", largo), "Carolina", "3")
		return nil
	})
	sc.Step(`^cambio a "([^"]*)" las horas de la tarea "([^"]*)"$`, func(horas, titulo string) error {
		return e.cambiarHoras(e.principal, titulo, horas)
	})
	sc.Step(`^cambio a "([^"]*)" las horas de la tarea "([^"]*)" desde la historia "([^"]*)"$`, func(horas, titulo, historia string) error {
		return e.cambiarHoras(e.porTitulo[historia], titulo, horas)
	})
	sc.Step(`^completo la tarea "([^"]*)"$`, func(titulo string) error { return e.marcar(titulo, true) })
	sc.Step(`^vuelvo a pendiente la tarea "([^"]*)"$`, func(titulo string) error { return e.marcar(titulo, false) })

	sc.Step(`^la historia tiene la tarea "([^"]*)" de "([^"]*)" con "([^"]*)" horas, pendiente$`, e.laHistoriaTieneLaTarea)
	sc.Step(`^las tareas de la historia, en orden, son "([^"]*)"$`, e.lasTareasSon)
	sc.Step(`^la historia suma "([^"]*)" horas estimadas$`, func(esperado string) error { return e.sumaHoras(e.principal, esperado) })
	sc.Step(`^la historia "([^"]*)" suma "([^"]*)" horas estimadas$`, func(titulo, esperado string) error {
		return e.sumaHoras(e.porTitulo[titulo], esperado)
	})
	sc.Step(`^la historia no tiene tareas$`, e.laHistoriaNoTieneTareas)
	sc.Step(`^la tarea se rechaza con el error "([^"]*)"$`, e.laTareaSeRechaza)
	sc.Step(`^la tarea da como resultado "([^"]*)"$`, e.laTareaDaComoResultado)
	sc.Step(`^no se encuentra la tarea$`, e.noSeEncuentraLaTarea)
	sc.Step(`^la tarea "([^"]*)" queda completada$`, e.laTareaQuedaCompletada)
}

func (e *hu25) elProyectoConLaHistoriaYLosIntegrantes(nombre, titulo, integrantes string) error {
	p, err := e.proyectos.Crear(context.Background(), project.Datos{
		Nombre:      nombre,
		FechaInicio: fechaDeEscenario("2026-10-05"),
		FechaFin:    fechaDeEscenario("2026-11-01"),
	})
	if err != nil {
		return fmt.Errorf("no se pudo crear el proyecto %q: %w", nombre, err)
	}
	e.proyectoID = p.ID
	if e.principal, err = e.crearHistoria(titulo); err != nil {
		return err
	}
	for _, n := range strings.Split(integrantes, ",") {
		n = strings.TrimSpace(n)
		d := member.Datos{Nombre: n, Email: strings.ToLower(n) + "@mail.com", Rol: member.RolProductBuilder}
		i, err := e.integrantes.Registrar(context.Background(), e.proyectoID, d)
		if err != nil {
			return fmt.Errorf("no se pudo registrar a %s: %w", n, err)
		}
		e.porNombre[n] = i
	}
	return nil
}

func (e *hu25) crearHistoria(titulo string) (backlog.Historia, error) {
	h, err := e.historias.Crear(context.Background(), e.proyectoID, backlog.Datos{Titulo: titulo, Prioridad: backlog.PrioridadAlta})
	if err != nil {
		return backlog.Historia{}, fmt.Errorf("no se pudo crear la historia %q: %w", titulo, err)
	}
	e.porTitulo[titulo] = h
	return h, nil
}

func (e *hu25) queEstaDadoDeBaja(nombre string) error {
	i, err := e.integrantes.DarDeBaja(context.Background(), e.proyectoID, e.porNombre[nombre].ID)
	if err != nil {
		return fmt.Errorf("no se pudo dar de baja a %s: %w", nombre, err)
	}
	e.porNombre[nombre] = i
	return nil
}

// agregar intenta agregar la tarea y guarda el error, para los pasos que lo verifican. Un responsable vacío es "sin
// responsable".
func (e *hu25) agregar(h backlog.Historia, titulo, responsable, horas string) {
	d := effort.DatosTarea{Titulo: titulo, ResponsableID: e.porNombre[responsable].ID, Horas: horas}
	var t effort.Tarea
	t, e.err = e.tareas.Agregar(context.Background(), e.proyectoID, h.Numero, d)
	if e.err == nil {
		e.tareaPorTit[t.Titulo] = t
	}
}

func (e *hu25) tareaPrevia(h backlog.Historia, titulo, responsable, horas string) error {
	e.agregar(h, titulo, responsable, horas)
	if e.err != nil {
		return fmt.Errorf("no se pudo cargar la tarea %q: %w", titulo, e.err)
	}
	return nil
}

func (e *hu25) otraHistoriaConLaTarea(historia, titulo, responsable, horas string) error {
	h, err := e.crearHistoria(historia)
	if err != nil {
		return err
	}
	return e.tareaPrevia(h, titulo, responsable, horas)
}

// cambiarHoras intenta cambiar las horas de la tarea desde la historia h y guarda el error.
func (e *hu25) cambiarHoras(h backlog.Historia, titulo, horas string) error {
	t, ok := e.tareaPorTit[titulo]
	if !ok {
		return fmt.Errorf("no hay una tarea %q en el escenario", titulo)
	}
	_, e.err = e.tareas.CambiarEstimacion(context.Background(), e.proyectoID, h.Numero, t.ID, horas)
	return nil
}

func (e *hu25) marcar(titulo string, completada bool) error {
	t, ok := e.tareaPorTit[titulo]
	if !ok {
		return fmt.Errorf("no hay una tarea %q en el escenario", titulo)
	}
	if _, err := e.tareas.Marcar(context.Background(), e.proyectoID, e.principal.Numero, t.ID, completada); err != nil {
		return fmt.Errorf("no se pudo marcar la tarea %q: %w", titulo, err)
	}
	return nil
}

func (e *hu25) vista(h backlog.Historia) (service.VistaTareas, error) {
	v, err := e.tareas.Ver(context.Background(), e.proyectoID, h.Numero)
	if err != nil {
		return service.VistaTareas{}, fmt.Errorf("no se pudo ver la historia %s: %w", h.Identificador(), err)
	}
	return v, nil
}

func (e *hu25) laHistoriaTieneLaTarea(titulo, responsable, horas string) error {
	if e.err != nil {
		return fmt.Errorf("no se esperaba error: %w", e.err)
	}
	v, err := e.vista(e.principal)
	if err != nil {
		return err
	}
	for _, t := range v.Tareas {
		if t.Titulo != titulo {
			continue
		}
		if got := v.NombreDe(t.ResponsableID); got != responsable {
			return fmt.Errorf("responsable = %q, se esperaba %q", got, responsable)
		}
		if got := effort.FormatearHoras(t.HorasEstimadas); got != horas {
			return fmt.Errorf("horas = %q, se esperaba %q", got, horas)
		}
		if t.Completada {
			return errors.New("la tarea nueva tendría que estar pendiente")
		}
		return nil
	}
	return fmt.Errorf("la historia no tiene la tarea %q", titulo)
}

func (e *hu25) lasTareasSon(esperadas string) error {
	v, err := e.vista(e.principal)
	if err != nil {
		return err
	}
	titulos := make([]string, 0, len(v.Tareas))
	for _, t := range v.Tareas {
		titulos = append(titulos, t.Titulo)
	}
	if got := strings.Join(titulos, "; "); got != esperadas {
		return fmt.Errorf("tareas = %q, se esperaba %q", got, esperadas)
	}
	return nil
}

func (e *hu25) sumaHoras(h backlog.Historia, esperado string) error {
	v, err := e.vista(h)
	if err != nil {
		return err
	}
	if got := effort.FormatearHoras(v.HorasEstimadas); got != esperado {
		return fmt.Errorf("la historia %s suma %q horas, se esperaba %q", h.Identificador(), got, esperado)
	}
	return nil
}

func (e *hu25) laHistoriaNoTieneTareas() error {
	v, err := e.vista(e.principal)
	if err != nil {
		return err
	}
	if len(v.Tareas) != 0 {
		return fmt.Errorf("no se esperaban tareas y hay %d", len(v.Tareas))
	}
	return nil
}

// mensajeDelError devuelve el texto que ve el usuario: el del campo si es un error de validación, o el error mismo.
func mensajeDelError(err error) string {
	var errs domain.ErroresValidacion
	if errors.As(err, &errs) {
		for _, e := range errs {
			return e.Error()
		}
	}
	return err.Error()
}

func (e *hu25) laTareaSeRechaza(mensaje string) error {
	if e.err == nil {
		return fmt.Errorf("se esperaba el error %q y no hubo error", mensaje)
	}
	var errs domain.ErroresValidacion
	if errors.As(e.err, &errs) {
		for _, err := range errs {
			if err.Error() == mensaje {
				return nil
			}
		}
		return fmt.Errorf("errores = %v, se esperaba %q", errs, mensaje)
	}
	if e.err.Error() != mensaje {
		return fmt.Errorf("error = %q, se esperaba %q", e.err.Error(), mensaje)
	}
	return nil
}

func (e *hu25) laTareaDaComoResultado(resultado string) error {
	if resultado == "agregada" {
		if e.err != nil {
			return fmt.Errorf("se esperaba que se agregue y hubo error: %w", e.err)
		}
		return nil
	}
	if e.err == nil {
		return fmt.Errorf("se esperaba el error %q y se agregó", resultado)
	}
	if got := mensajeDelError(e.err); got != resultado {
		return fmt.Errorf("error = %q, se esperaba %q", got, resultado)
	}
	return nil
}

func (e *hu25) noSeEncuentraLaTarea() error {
	if !errors.Is(e.err, effort.ErrTareaNoEncontrada) {
		return fmt.Errorf("se esperaba %v y se obtuvo %v", effort.ErrTareaNoEncontrada, e.err)
	}
	return nil
}

func (e *hu25) laTareaQuedaCompletada(titulo string) error {
	v, err := e.vista(e.principal)
	if err != nil {
		return err
	}
	for _, t := range v.Tareas {
		if t.Titulo == titulo {
			if !t.Completada {
				return fmt.Errorf("la tarea %q tendría que estar completada", titulo)
			}
			return nil
		}
	}
	return fmt.Errorf("la historia no tiene la tarea %q", titulo)
}
