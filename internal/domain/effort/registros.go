package effort

import (
	"errors"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
)

// Límites de un registro de esfuerzo (spec de HU-24).
const (
	HorasPorDiaMax = 24
	ActividadMax   = 200
)

// Nombres de los campos de un registro, además de CampoHoras.
const (
	CampoIntegrante = "integrante"
	CampoTarea      = "tarea"
	CampoFecha      = "fecha"
	CampoActividad  = "actividad"
)

// Errores de validación de un registro (sección 7 de la spec de HU-24).
var (
	ErrHorasTrabajadasInvalidas = errors.New("las horas tienen que ser mayores a 0 y hasta 24")
	ErrMasDe24HorasEnElDia      = errors.New("el integrante no puede superar 24 h en un día")
	ErrFechaInvalida            = errors.New("la fecha es obligatoria")
	ErrFechaFutura              = errors.New("la fecha no puede ser futura")
	ErrFechaFueraDelProyecto    = errors.New("la fecha tiene que estar dentro de las fechas del proyecto")
	ErrIntegranteInvalido       = errors.New("el integrante tiene que ser un integrante activo del proyecto")
	ErrTareaDeOtraHistoria      = errors.New("la tarea tiene que ser de esta historia")
	ErrActividadObligatoria     = errors.New("la actividad es obligatoria")
	ErrActividadLarga           = errors.New("la actividad no puede superar los 200 caracteres")
)

// DatosRegistro son los campos que carga el usuario para registrar horas trabajadas.
type DatosRegistro struct {
	IntegranteID int64
	TareaID      int64 // 0 si el registro es de toda la historia
	Fecha        string
	Actividad    string
	Horas        string
}

// ContextoRegistro es lo que hace falta saber para validar un registro: la historia, las fechas del proyecto, hoy (en
// Argentina), los integrantes del proyecto, las tareas de la historia y los registros que ya tiene el integrante.
type ContextoRegistro struct {
	HistoriaID     int64
	InicioProyecto time.Time
	FinProyecto    time.Time
	Hoy            time.Time
	Integrantes    []member.Integrante
	Tareas         []Tarea
	DelIntegrante  []Registro
}

// Registro son las horas que un integrante trabajó un día en una historia o en una de sus tareas.
//
// Spec: specs/HU-24-registro-esfuerzo.md.
type Registro struct {
	ID           int64
	IntegranteID int64
	HistoriaID   int64
	TareaID      int64 // 0 si es de toda la historia
	Fecha        time.Time
	Actividad    string
	Horas        float64
}

// NuevoRegistro valida los datos (RN1 a RN6) y devuelve el registro listo para guardar. Si algún dato es inválido,
// devuelve domain.ErroresValidacion con todos los campos que fallaron (RN8).
func NuevoRegistro(d DatosRegistro, c ContextoRegistro) (Registro, error) {
	actividad := strings.TrimSpace(d.Actividad)
	fecha, errFecha := validarFecha(d.Fecha, c)
	horas, errHoras := parsearHorasTrabajadas(d.Horas)
	errIntegrante := validarIntegrante(d.IntegranteID, c.Integrantes)
	if errHoras == nil && errFecha == nil && errIntegrante == nil && horasDelDia(c.DelIntegrante, d.IntegranteID, fecha)+horas > HorasPorDiaMax {
		errHoras = ErrMasDe24HorasEnElDia
	}

	errs := domain.ErroresValidacion{}
	errs.Agregar(CampoIntegrante, errIntegrante)
	errs.Agregar(CampoTarea, validarTarea(d.TareaID, c))
	errs.Agregar(CampoFecha, errFecha)
	errs.Agregar(CampoActividad, validarActividad(actividad))
	errs.Agregar(CampoHoras, errHoras)
	if len(errs) > 0 {
		return Registro{}, errs
	}
	return Registro{
		IntegranteID: d.IntegranteID, HistoriaID: c.HistoriaID, TareaID: d.TareaID,
		Fecha: fecha, Actividad: actividad, Horas: horas,
	}, nil
}

// HorasReales suma las horas de los registros (M5, RN7).
func HorasReales(registros []Registro) float64 {
	total := 0.0
	for _, r := range registros {
		total += r.Horas
	}
	return total
}

func parsearHorasTrabajadas(texto string) (float64, error) {
	horas, err := ParsearHoras(texto)
	if err != nil || horas > HorasPorDiaMax {
		return 0, ErrHorasTrabajadasInvalidas
	}
	return horas, nil
}

func validarFecha(texto string, c ContextoRegistro) (time.Time, error) {
	t, err := time.Parse(time.DateOnly, strings.TrimSpace(texto))
	if err != nil {
		return time.Time{}, ErrFechaInvalida
	}
	fecha := domain.SoloFecha(t)
	switch {
	case fecha.Before(domain.SoloFecha(c.InicioProyecto)) || fecha.After(domain.SoloFecha(c.FinProyecto)):
		return fecha, ErrFechaFueraDelProyecto
	case fecha.After(domain.SoloFecha(c.Hoy)):
		return fecha, ErrFechaFutura
	}
	return fecha, nil
}

func validarIntegrante(id int64, integrantes []member.Integrante) error {
	for _, i := range integrantes {
		if i.ID == id && i.Activo {
			return nil
		}
	}
	return ErrIntegranteInvalido
}

func validarTarea(id int64, c ContextoRegistro) error {
	if id == 0 {
		return nil
	}
	for _, t := range c.Tareas {
		if t.ID == id && t.HistoriaID == c.HistoriaID {
			return nil
		}
	}
	return ErrTareaDeOtraHistoria
}

func validarActividad(actividad string) error {
	switch largo := utf8.RuneCountInString(actividad); {
	case largo == 0:
		return ErrActividadObligatoria
	case largo > ActividadMax:
		return ErrActividadLarga
	}
	return nil
}

// horasDelDia suma lo que el integrante ya registró ese día, redondeado a centésimos para que 20 + 4 dé 24 justo.
func horasDelDia(registros []Registro, integranteID int64, fecha time.Time) float64 {
	total := 0.0
	for _, r := range registros {
		if r.IntegranteID == integranteID && domain.SoloFecha(r.Fecha).Equal(fecha) {
			total += r.Horas
		}
	}
	return math.Round(total*100) / 100
}
