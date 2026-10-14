package effort_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
)

// integrantesDelProyecto son los integrantes del ejemplo de la spec de HU-25: Carolina (2) y Mariano (3) activos, y
// Ana (4) dada de baja.
func integrantesDelProyecto() []member.Integrante {
	return []member.Integrante{
		{ID: 2, ProyectoID: 7, Nombre: "Carolina", Rol: member.RolProductBuilder, Activo: true},
		{ID: 3, ProyectoID: 7, Nombre: "Mariano", Rol: member.RolProductBuilder, Activo: true},
		{ID: 4, ProyectoID: 7, Nombre: "Ana", Rol: member.RolProductBuilder, Activo: false},
	}
}

func tareaValida() effort.DatosTarea {
	return effort.DatosTarea{Titulo: "Pantalla de login", ResponsableID: 2, Horas: "3"}
}

// errorDelCampo devuelve el error que NuevaTarea informó para el campo, o nil si no informó ninguno.
func errorDelCampo(t *testing.T, err error, campo string) error {
	t.Helper()
	var errs domain.ErroresValidacion
	if !errors.As(err, &errs) {
		t.Fatalf("se esperaba domain.ErroresValidacion, llegó %v", err)
	}
	return errs[campo]
}

// CA-25.1, CA-25.2 y RN5 · la tarea tiene título, responsable y horas, es de la historia y nace pendiente
func TestNuevaTarea_TieneSusDatosYNacePendiente(t *testing.T) {
	d := tareaValida()
	d.Titulo = "  Pantalla de login \n"
	d.Horas = " 2,5 "

	tarea, err := effort.NuevaTarea(d, 9, integrantesDelProyecto())
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	esperada := effort.Tarea{HistoriaID: 9, Titulo: "Pantalla de login", ResponsableID: 2, HorasEstimadas: 2.5}
	if tarea != esperada {
		t.Errorf("tarea = %+v, se esperaba %+v", tarea, esperada)
	}
}

// RN1 · el título es obligatorio y tiene hasta 120 caracteres, contados en caracteres y sin los espacios sobrantes
func TestNuevaTarea_ValidaElTitulo(t *testing.T) {
	casos := []struct {
		nombre string
		titulo string
		err    error
	}{
		{"vacío", "", effort.ErrTituloTareaObligatorio},
		{"solo espacios", "   \t ", effort.ErrTituloTareaObligatorio},
		{"120 caracteres", strings.Repeat("a", 120), nil},
		{"120 caracteres con tildes", strings.Repeat("ñ", 120), nil},
		{"121 caracteres", strings.Repeat("a", 121), effort.ErrTituloTareaLargo},
		{"120 caracteres con espacios alrededor", "  " + strings.Repeat("a", 120) + "  ", nil},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d := tareaValida()
			d.Titulo = c.titulo

			_, err := effort.NuevaTarea(d, 9, integrantesDelProyecto())

			if c.err == nil {
				if err != nil {
					t.Fatalf("no se esperaba error: %v", err)
				}
				return
			}
			if got := errorDelCampo(t, err, effort.CampoTitulo); !errors.Is(got, c.err) {
				t.Errorf("error del título = %v, se esperaba %v", got, c.err)
			}
		})
	}
}

// RN2 · el responsable es obligatorio y tiene que ser un integrante activo del proyecto
func TestNuevaTarea_ValidaElResponsable(t *testing.T) {
	casos := []struct {
		nombre        string
		responsableID int64
		err           error
	}{
		{"sin responsable", 0, effort.ErrResponsableObligatorio},
		{"dado de baja", 4, effort.ErrResponsableInvalido},
		{"no es del proyecto", 99, effort.ErrResponsableInvalido},
		{"activo", 3, nil},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d := tareaValida()
			d.ResponsableID = c.responsableID

			tarea, err := effort.NuevaTarea(d, 9, integrantesDelProyecto())

			if c.err == nil {
				if err != nil {
					t.Fatalf("no se esperaba error: %v", err)
				}
				if tarea.ResponsableID != c.responsableID {
					t.Errorf("responsable = %d, se esperaba %d", tarea.ResponsableID, c.responsableID)
				}
				return
			}
			if got := errorDelCampo(t, err, effort.CampoResponsable); !errors.Is(got, c.err) {
				t.Errorf("error del responsable = %v, se esperaba %v", got, c.err)
			}
		})
	}
}

// RN3 · las horas son las de HU-23: mayores a 0, con coma o punto decimal
func TestNuevaTarea_ValidaLasHoras(t *testing.T) {
	for _, horas := range []string{"", "0", "-1", "abc", "1,5,2"} {
		t.Run(horas, func(t *testing.T) {
			d := tareaValida()
			d.Horas = horas

			_, err := effort.NuevaTarea(d, 9, integrantesDelProyecto())

			if got := errorDelCampo(t, err, effort.CampoHoras); !errors.Is(got, effort.ErrHorasInvalidas) {
				t.Errorf("error de las horas = %v, se esperaba %v", got, effort.ErrHorasInvalidas)
			}
		})
	}
}

// Los errores de todos los campos se informan juntos, para mostrar cada uno debajo de su campo.
func TestNuevaTarea_InformaTodosLosErroresJuntos(t *testing.T) {
	_, err := effort.NuevaTarea(effort.DatosTarea{}, 9, integrantesDelProyecto())

	var errs domain.ErroresValidacion
	if !errors.As(err, &errs) {
		t.Fatalf("se esperaba domain.ErroresValidacion, llegó %v", err)
	}
	for _, campo := range []string{effort.CampoTitulo, effort.CampoResponsable, effort.CampoHoras} {
		if errs[campo] == nil {
			t.Errorf("falta el error del campo %q", campo)
		}
	}
}

func tareaGuardada() effort.Tarea {
	return effort.Tarea{ID: 5, HistoriaID: 9, Titulo: "Pantalla de login", ResponsableID: 2, HorasEstimadas: 3}
}

// CA-25.3 · mientras está pendiente, la estimación se cambia con la regla de horas de HU-23; la original no cambia
func TestCambiarEstimacion_TareaPendiente(t *testing.T) {
	original := tareaGuardada()

	cambiada, err := original.CambiarEstimacion(" 4,5 ")
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if cambiada.HorasEstimadas != 4.5 {
		t.Errorf("horas = %v, se esperaba 4.5", cambiada.HorasEstimadas)
	}
	if original.HorasEstimadas != 3 {
		t.Error("la tarea original no debería cambiar")
	}
}

// RN3 · las horas nuevas también tienen que ser mayores a 0
func TestCambiarEstimacion_HorasInvalidas(t *testing.T) {
	for _, horas := range []string{"", "0", "-2", "abc"} {
		t.Run(horas, func(t *testing.T) {
			if _, err := tareaGuardada().CambiarEstimacion(horas); !errors.Is(err, effort.ErrHorasInvalidas) {
				t.Errorf("error = %v, se esperaba %v", err, effort.ErrHorasInvalidas)
			}
		})
	}
}

// CA-25.3 (RN6) · una tarea completada no cambia su estimación, aunque las horas nuevas sean válidas
func TestCambiarEstimacion_TareaCompletada_EsUnError(t *testing.T) {
	completada := tareaGuardada()
	completada.Completada = true

	if _, err := completada.CambiarEstimacion("5"); !errors.Is(err, effort.ErrTareaCompletada) {
		t.Errorf("error = %v, se esperaba %v", err, effort.ErrTareaCompletada)
	}
}
