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
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// hu07 guarda el estado de un escenario de HU-07 (gestionar criterios de aceptación), con sus propios repositorios en
// memoria. "La historia" de los pasos que no nombran ninguna es la del Antecedentes; las demás se buscan por título.
type hu07 struct {
	repoHistorias *memory.Historias
	repoCriterios *memory.Criterios
	historias     *service.Historias
	criterios     *service.Criterios
	proyectoID    int64
	principal     backlog.Historia
	porTitulo     map[string]backlog.Historia
	err           error
}

// registrarPasosHU07 conecta cada frase de features/hu-07-criterios-aceptacion.feature con una función de Go.
func registrarPasosHU07(sc *godog.ScenarioContext) {
	repoProyectos := memory.NuevoProyectos()
	e := &hu07{
		repoHistorias: memory.NuevoHistorias(),
		repoCriterios: memory.NuevoCriterios(),
		porTitulo:     map[string]backlog.Historia{},
	}
	e.historias = service.NuevoHistorias(repoProyectos, e.repoHistorias)
	e.criterios = service.NuevoCriterios(repoProyectos, e.repoHistorias, e.repoCriterios)
	proyectos := service.NuevoProyectos(repoProyectos, time.Now)

	sc.Step(`^el proyecto "([^"]*)" con la historia "([^"]*)"$`, func(nombre, titulo string) error {
		return e.elProyectoConLaHistoria(proyectos, nombre, titulo)
	})
	sc.Step(`^que la historia tiene el criterio "([^"]*)"$`, e.queLaHistoriaTieneElCriterio)
	sc.Step(`^que la historia tiene (\d+) criterios?, (\d+) de ellos cumplidos$`, e.queLaHistoriaTieneCriterios)
	sc.Step(`^la historia "([^"]*)" con (\d+) criterios?, (\d+) de ellos cumplidos$`, e.otraHistoriaConCriterios)
	sc.Step(`^que la historia "([^"]*)" está Hecha y tiene (\d+) criterios cumplidos$`, e.queLaHistoriaEstaHecha)
	sc.Step(`^agrego a la historia el criterio "([^"]*)"$`, e.agregoElCriterio)
	sc.Step(`^agrego a la historia "([^"]*)" el criterio "([^"]*)"$`, e.agregoElCriterioALaHistoria)
	sc.Step(`^agrego a la historia un criterio de (\d+) caracteres$`, e.agregoUnCriterioDeLargo)
	sc.Step(`^marco como cumplido el criterio número (\d+)$`, func(n int) error { return e.marco(e.principal, n, true) })
	sc.Step(`^marco como no cumplido el criterio número (\d+)$`, func(n int) error { return e.marco(e.principal, n, false) })
	sc.Step(`^marco como no cumplido el criterio número (\d+) de la historia "([^"]*)"$`, func(n int, titulo string) error {
		return e.marco(e.porTitulo[titulo], n, false)
	})
	sc.Step(`^el criterio "([^"]*)" queda no cumplido$`, e.elCriterioQuedaNoCumplido)
	sc.Step(`^el criterio número (\d+) queda cumplido$`, e.elCriterioNumeroQuedaCumplido)
	sc.Step(`^los criterios de la historia, en orden, son "([^"]*)"$`, e.losCriteriosSon)
	sc.Step(`^el progreso de la historia es "([^"]*)"$`, func(esperado string) error { return e.elProgresoEs(e.principal, esperado) })
	sc.Step(`^el progreso de la historia "([^"]*)" es "([^"]*)"$`, func(titulo, esperado string) error {
		return e.elProgresoEs(e.porTitulo[titulo], esperado)
	})
	sc.Step(`^la historia (puede|no puede) entrar a un sprint por sus criterios$`, e.laHistoriaPuedeEntrarAUnSprint)
	sc.Step(`^la historia (está|no está) completa$`, e.laHistoriaEstaCompleta)
	sc.Step(`^el criterio da como resultado "([^"]*)"$`, e.elCriterioDaComoResultado)
}

func (e *hu07) elProyectoConLaHistoria(proyectos *service.Proyectos, nombre, titulo string) error {
	p, err := proyectos.Crear(context.Background(), project.Datos{
		Nombre:      nombre,
		FechaInicio: fechaDeEscenario("2026-10-05"),
		FechaFin:    fechaDeEscenario("2026-11-01"),
	})
	if err != nil {
		return fmt.Errorf("no se pudo crear el proyecto %q: %w", nombre, err)
	}
	e.proyectoID = p.ID
	e.principal, err = e.crearHistoria(titulo)
	return err
}

func (e *hu07) crearHistoria(titulo string) (backlog.Historia, error) {
	h, err := e.historias.Crear(context.Background(), e.proyectoID, backlog.Datos{Titulo: titulo, Prioridad: backlog.PrioridadAlta})
	if err != nil {
		return backlog.Historia{}, fmt.Errorf("no se pudo crear la historia %q: %w", titulo, err)
	}
	e.porTitulo[titulo] = h
	return h, nil
}

// agregarCriterios carga total criterios en la historia y deja cumplidos los primeros cumplidos.
func (e *hu07) agregarCriterios(h backlog.Historia, total, cumplidos int) error {
	for i := 1; i <= total; i++ {
		c, err := e.criterios.Agregar(context.Background(), e.proyectoID, h.Numero, fmt.Sprintf("Criterio %d", i))
		if err != nil {
			return fmt.Errorf("no se pudo cargar el criterio %d: %w", i, err)
		}
		if i <= cumplidos {
			if _, err := e.criterios.Marcar(context.Background(), e.proyectoID, h.Numero, c.ID, true); err != nil {
				return fmt.Errorf("no se pudo marcar el criterio %d: %w", i, err)
			}
		}
	}
	return nil
}

func (e *hu07) queLaHistoriaTieneElCriterio(descripcion string) error {
	e.agregar(e.principal, descripcion)
	return e.err
}

func (e *hu07) queLaHistoriaTieneCriterios(total, cumplidos int) error {
	return e.agregarCriterios(e.principal, total, cumplidos)
}

func (e *hu07) otraHistoriaConCriterios(titulo string, total, cumplidos int) error {
	h, err := e.crearHistoria(titulo)
	if err != nil {
		return err
	}
	return e.agregarCriterios(h, total, cumplidos)
}

// queLaHistoriaEstaHecha guarda una historia Hecha con criterios cumplidos directo en los repositorios: pasar una
// historia a Hecho es de HU-08, que todavía no existe.
func (e *hu07) queLaHistoriaEstaHecha(titulo string, criterios int) error {
	ctx := context.Background()
	h, err := e.repoHistorias.Guardar(ctx, backlog.Historia{
		ProyectoID: e.proyectoID, Numero: 99, Titulo: titulo,
		Prioridad: backlog.PrioridadAlta, Estado: backlog.EstadoHecho,
	})
	if err != nil {
		return fmt.Errorf("no se pudo guardar la historia Hecha: %w", err)
	}
	e.porTitulo[titulo] = h
	for i := 1; i <= criterios; i++ {
		if _, err := e.repoCriterios.Guardar(ctx, backlog.Criterio{HistoriaID: h.ID, Descripcion: fmt.Sprintf("Criterio %d", i), Cumplido: true}); err != nil {
			return fmt.Errorf("no se pudo guardar el criterio %d: %w", i, err)
		}
	}
	return nil
}

func (e *hu07) agregar(h backlog.Historia, descripcion string) {
	_, e.err = e.criterios.Agregar(context.Background(), e.proyectoID, h.Numero, descripcion)
}

func (e *hu07) agregoElCriterio(descripcion string) error {
	e.agregar(e.principal, descripcion)
	return nil
}

func (e *hu07) agregoElCriterioALaHistoria(titulo, descripcion string) error {
	h, ok := e.porTitulo[titulo]
	if !ok {
		return fmt.Errorf("el escenario no creó la historia %q", titulo)
	}
	e.agregar(h, descripcion)
	return nil
}

func (e *hu07) agregoUnCriterioDeLargo(largo int) error {
	e.agregar(e.principal, strings.Repeat("a", largo))
	return nil
}

// criteriosDe devuelve los criterios de la historia en el orden en que se cargaron.
func (e *hu07) criteriosDe(h backlog.Historia) ([]backlog.Criterio, error) {
	v, err := e.criterios.Ver(context.Background(), e.proyectoID, h.Numero)
	return v.Criterios, err
}

// marco pide marcar el criterio número n (el n-ésimo en orden) de la historia.
func (e *hu07) marco(h backlog.Historia, n int, cumplido bool) error {
	lista, err := e.criteriosDe(h)
	if err != nil {
		return err
	}
	if n < 1 || n > len(lista) {
		return fmt.Errorf("la historia %q no tiene el criterio número %d (tiene %d)", h.Titulo, n, len(lista))
	}
	_, e.err = e.criterios.Marcar(context.Background(), e.proyectoID, h.Numero, lista[n-1].ID, cumplido)
	return nil
}

func (e *hu07) elCriterioQuedaNoCumplido(descripcion string) error {
	lista, err := e.criteriosDe(e.principal)
	if err != nil {
		return err
	}
	for _, c := range lista {
		if c.Descripcion == descripcion {
			if c.Cumplido {
				return fmt.Errorf("el criterio %q tenía que quedar no cumplido y está cumplido", descripcion)
			}
			return nil
		}
	}
	return fmt.Errorf("la historia no tiene el criterio %q (error al agregarlo: %v)", descripcion, e.err)
}

func (e *hu07) elCriterioNumeroQuedaCumplido(n int) error {
	lista, err := e.criteriosDe(e.principal)
	if err != nil {
		return err
	}
	if n < 1 || n > len(lista) {
		return fmt.Errorf("la historia no tiene el criterio número %d (tiene %d)", n, len(lista))
	}
	if !lista[n-1].Cumplido {
		return fmt.Errorf("el criterio número %d tenía que quedar cumplido y no lo está", n)
	}
	return nil
}

func (e *hu07) losCriteriosSon(esperados string) error {
	lista, err := e.criteriosDe(e.principal)
	if err != nil {
		return err
	}
	descripciones := make([]string, 0, len(lista))
	for _, c := range lista {
		descripciones = append(descripciones, c.Descripcion)
	}
	if obtenidos := strings.Join(descripciones, "; "); obtenidos != esperados {
		return fmt.Errorf("criterios: se esperaba %q y se obtuvo %q", esperados, obtenidos)
	}
	return nil
}

func (e *hu07) progreso(h backlog.Historia) (backlog.Progreso, error) {
	v, err := e.criterios.Ver(context.Background(), e.proyectoID, h.Numero)
	return v.Progreso, err
}

func (e *hu07) elProgresoEs(h backlog.Historia, esperado string) error {
	p, err := e.progreso(h)
	if err != nil {
		return err
	}
	if p.Texto() != esperado {
		return fmt.Errorf("progreso de %q: se esperaba %q y se obtuvo %q", h.Titulo, esperado, p.Texto())
	}
	return nil
}

func (e *hu07) laHistoriaPuedeEntrarAUnSprint(puede string) error {
	p, err := e.progreso(e.principal)
	if err != nil {
		return err
	}
	if esperado := puede == "puede"; p.TieneCriterios() != esperado {
		return fmt.Errorf("con el progreso %s, TieneCriterios = %v y se esperaba %v", p.Texto(), p.TieneCriterios(), esperado)
	}
	return nil
}

func (e *hu07) laHistoriaEstaCompleta(esta string) error {
	p, err := e.progreso(e.principal)
	if err != nil {
		return err
	}
	if esperado := esta == "está"; p.Completo() != esperado {
		return fmt.Errorf("con el progreso %s, Completo = %v y se esperaba %v", p.Texto(), p.Completo(), esperado)
	}
	return nil
}

// elCriterioDaComoResultado acepta "agregado", "campo: mensaje" (error de validación) o el mensaje de otro error.
func (e *hu07) elCriterioDaComoResultado(resultado string) error {
	if resultado == "agregado" {
		if e.err != nil {
			return fmt.Errorf("se esperaba que se agregue y se obtuvo: %w", e.err)
		}
		return nil
	}
	var errs domain.ErroresValidacion
	if errors.As(e.err, &errs) {
		campo, mensaje, ok := strings.Cut(resultado, ": ")
		if !ok {
			return fmt.Errorf("resultado mal escrito en el escenario: %q (se espera \"campo: mensaje\")", resultado)
		}
		if errs[campo] == nil || errs[campo].Error() != mensaje {
			return fmt.Errorf("campo %q: se esperaba %q y se obtuvo %v", campo, mensaje, errs[campo])
		}
		return nil
	}
	if e.err == nil || e.err.Error() != resultado {
		return fmt.Errorf("se esperaba el error %q y se obtuvo: %v", resultado, e.err)
	}
	return nil
}
