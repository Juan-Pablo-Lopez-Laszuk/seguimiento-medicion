package features

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// hu03 guarda el estado de un escenario de HU-03 (registrar integrantes), con sus propios repositorios
// en memoria. El primer proyecto que se crea es "el proyecto" de los pasos que no nombran ninguno.
type hu03 struct {
	proyectos    *service.Proyectos
	integrantes  *service.Integrantes
	idProyecto   map[string]int64
	principal    int64
	idIntegrante map[string]int64 // por nombre, dentro del proyecto principal
	registrado   member.Integrante
	err          error
	errBaja      error
}

// registrarPasosHU03 conecta cada frase de features/hu-03-registrar-integrantes.feature con una función de Go.
func registrarPasosHU03(sc *godog.ScenarioContext) {
	repoProyectos := memory.NuevoProyectos()
	e := &hu03{
		proyectos:    service.NuevoProyectos(repoProyectos, time.Now),
		integrantes:  service.NuevoIntegrantes(repoProyectos, memory.NuevoIntegrantes()),
		idProyecto:   map[string]int64{},
		idIntegrante: map[string]int64{},
	}

	sc.Step(`^el proyecto "([^"]*)" para registrar integrantes$`, e.elProyectoParaRegistrarIntegrantes)
	sc.Step(`^que ya está registrado "([^"]*)" con el email "([^"]*)" como "([^"]*)"$`, e.queYaEstaRegistrado)
	sc.Step(`^"([^"]*)" está dado de baja$`, e.estaDadoDeBaja)
	sc.Step(`^registro a "([^"]*)" con el email "([^"]*)" como "([^"]*)"$`, e.registroA)
	sc.Step(`^registro en "([^"]*)" a "([^"]*)" con el email "([^"]*)" como "([^"]*)"$`, e.registroEnA)
	sc.Step(`^queda registrado "([^"]*)" con el email "([^"]*)" como "([^"]*)"$`, e.quedaRegistrado)
	sc.Step(`^el integrante está activo$`, e.elIntegranteEstaActivo)
	sc.Step(`^el proyecto tiene (\d+) integrantes activos$`, e.elProyectoTieneIntegrantesActivos)
	sc.Step(`^el registro da como resultado "([^"]*)"$`, e.elRegistroDaComoResultado)
	sc.Step(`^doy de baja a "([^"]*)"$`, e.doyDeBajaA)
	sc.Step(`^"([^"]*)" sigue en el proyecto pero dado de baja$`, e.sigueEnElProyectoPeroDadoDeBaja)
	sc.Step(`^la baja da el error "([^"]*)"$`, e.laBajaDaElError)
}

// rolDeEscenario convierte el rol como se escribe en el escenario ("Agile Enabler") en el valor del dominio.
// Un rol desconocido ("Scrum Master") se pasa igual, sin espacios, para que el dominio lo rechace.
func rolDeEscenario(texto string) member.Rol {
	return member.Rol(strings.ReplaceAll(texto, " ", ""))
}

func (e *hu03) elProyectoParaRegistrarIntegrantes(nombre string) error {
	p, err := e.proyectos.Crear(context.Background(), project.Datos{
		Nombre:      nombre,
		FechaInicio: fechaDeEscenario("2026-10-05"),
		FechaFin:    fechaDeEscenario("2026-11-01"),
	})
	if err != nil {
		return fmt.Errorf("no se pudo crear el proyecto %q: %w", nombre, err)
	}
	e.idProyecto[nombre] = p.ID
	if e.principal == 0 {
		e.principal = p.ID
	}
	return nil
}

func (e *hu03) registrar(proyectoID int64, nombre, email, rol string) {
	e.registrado, e.err = e.integrantes.Registrar(context.Background(), proyectoID, member.Datos{
		Nombre: nombre, Email: email, Rol: rolDeEscenario(rol),
	})
	if e.err == nil && proyectoID == e.principal {
		e.idIntegrante[nombre] = e.registrado.ID
	}
}

func (e *hu03) queYaEstaRegistrado(nombre, email, rol string) error {
	e.registrar(e.principal, nombre, email, rol)
	return e.err
}

func (e *hu03) estaDadoDeBaja(nombre string) error {
	_, err := e.integrantes.DarDeBaja(context.Background(), e.principal, e.idIntegrante[nombre])
	return err
}

func (e *hu03) registroA(nombre, email, rol string) error {
	e.registrar(e.principal, nombre, email, rol)
	return nil
}

func (e *hu03) registroEnA(proyecto, nombre, email, rol string) error {
	id, ok := e.idProyecto[proyecto]
	if !ok {
		return fmt.Errorf("el escenario no creó el proyecto %q", proyecto)
	}
	e.registrar(id, nombre, email, rol)
	return nil
}

func (e *hu03) quedaRegistrado(nombre, email, rol string) error {
	if e.err != nil {
		return fmt.Errorf("no se registró: %w", e.err)
	}
	i := e.registrado
	if i.Nombre != nombre || i.Email != email || i.Rol.Texto() != rol {
		return fmt.Errorf("se esperaba %q <%s> como %q y se obtuvo %q <%s> como %q",
			nombre, email, rol, i.Nombre, i.Email, i.Rol.Texto())
	}
	return nil
}

func (e *hu03) elIntegranteEstaActivo() error {
	if !e.registrado.Activo {
		return errors.New("se esperaba que el integrante esté activo")
	}
	return nil
}

func (e *hu03) elProyectoTieneIntegrantesActivos(cantidad int) error {
	_, lista, err := e.integrantes.Listar(context.Background(), e.principal)
	if err != nil {
		return err
	}
	activos := 0
	for _, i := range lista {
		if i.Activo {
			activos++
		}
	}
	if activos != cantidad {
		return fmt.Errorf("se esperaban %d integrantes activos y hay %d", cantidad, activos)
	}
	return nil
}

// elRegistroDaComoResultado acepta "registrado" o "campo: mensaje de error".
func (e *hu03) elRegistroDaComoResultado(resultado string) error {
	if resultado == "registrado" {
		if e.err != nil {
			return fmt.Errorf("se esperaba que se registre y se obtuvo: %w", e.err)
		}
		return nil
	}
	campo, mensaje, ok := strings.Cut(resultado, ": ")
	if !ok {
		return fmt.Errorf("resultado mal escrito en el escenario: %q (se espera \"registrado\" o \"campo: mensaje\")", resultado)
	}
	var errs domain.ErroresValidacion
	if !errors.As(e.err, &errs) {
		return fmt.Errorf("se esperaban errores de validación y se obtuvo: %v", e.err)
	}
	if errs[campo] == nil || errs[campo].Error() != mensaje {
		return fmt.Errorf("campo %q: se esperaba %q y se obtuvo %v", campo, mensaje, errs[campo])
	}
	return nil
}

func (e *hu03) doyDeBajaA(nombre string) error {
	_, e.errBaja = e.integrantes.DarDeBaja(context.Background(), e.principal, e.idIntegrante[nombre])
	return nil
}

func (e *hu03) sigueEnElProyectoPeroDadoDeBaja(nombre string) error {
	if e.errBaja != nil {
		return fmt.Errorf("no se dio de baja: %w", e.errBaja)
	}
	_, lista, err := e.integrantes.Listar(context.Background(), e.principal)
	if err != nil {
		return err
	}
	for _, i := range lista {
		if i.Nombre == nombre {
			if i.Activo {
				return fmt.Errorf("%q sigue activo", nombre)
			}
			return nil
		}
	}
	return fmt.Errorf("%q ya no está en el proyecto: se borró en lugar de darse de baja", nombre)
}

func (e *hu03) laBajaDaElError(mensaje string) error {
	if e.errBaja == nil || e.errBaja.Error() != mensaje {
		return fmt.Errorf("se esperaba el error %q y se obtuvo %v", mensaje, e.errBaja)
	}
	return nil
}
