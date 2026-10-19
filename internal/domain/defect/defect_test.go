package defect_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/defect"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
)

// El ejemplo de la spec de HU-27: el proyecto 7 con su Sprint 1 (id 3) y la HU-1 (id 4); el sprint 30 y la historia
// 40 son de otro proyecto.
func sprintsDelProyecto() []sprint.Sprint { return []sprint.Sprint{{ID: 3, ProyectoID: 7, Numero: 1}} }
func historiasDelProyecto() []backlog.Historia {
	return []backlog.Historia{{ID: 4, ProyectoID: 7, Numero: 1}}
}

func defectoValido() defect.Datos {
	return defect.Datos{Descripcion: "  El login falla con Google ", Severidad: defect.SeveridadAlta, SprintID: 3, HistoriaID: 4}
}

func nuevo(d defect.Datos) (defect.Defecto, error) {
	return defect.Nuevo(d, 7, sprintsDelProyecto(), historiasDelProyecto())
}

// CA-27.1 a CA-27.3 · el defecto guarda sus datos, sin espacios sobrantes, y arranca Abierto
func TestNuevo_ArrancaAbierto(t *testing.T) {
	d, err := nuevo(defectoValido())
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	esperado := defect.Defecto{ProyectoID: 7, HistoriaID: 4, Descripcion: "El login falla con Google",
		Severidad: defect.SeveridadAlta, Estado: defect.EstadoAbierto, SprintDeteccionID: 3}
	if d != esperado {
		t.Errorf("defecto = %+v, se esperaba %+v", d, esperado)
	}
}

// RN5 · la historia es opcional
func TestNuevo_SinHistoria(t *testing.T) {
	datos := defectoValido()
	datos.HistoriaID = 0
	if d, err := nuevo(datos); err != nil || d.HistoriaID != 0 {
		t.Errorf("se esperaba un defecto sin historia y se obtuvo %+v, %v", d, err)
	}
}

func TestNuevo_Validaciones(t *testing.T) {
	casos := []struct {
		nombre string
		cambio func(*defect.Datos)
		campo  string
		err    error
	}{
		{"descripción vacía", func(d *defect.Datos) { d.Descripcion = "   " }, defect.CampoDescripcion, defect.ErrDescripcionObligatoria},
		{"descripción de 1000", func(d *defect.Datos) { d.Descripcion = strings.Repeat("ñ", 1000) }, "", nil},
		{"descripción de 1001", func(d *defect.Datos) { d.Descripcion = strings.Repeat("a", 1001) }, defect.CampoDescripcion, defect.ErrDescripcionLarga},
		{"severidad inválida", func(d *defect.Datos) { d.Severidad = "Urgente" }, defect.CampoSeveridad, defect.ErrSeveridadInvalida},
		{"cada severidad válida", func(d *defect.Datos) { d.Severidad = defect.SeveridadCritica }, "", nil},
		{"sin sprint", func(d *defect.Datos) { d.SprintID = 0 }, defect.CampoSprint, defect.ErrSprintInvalido},
		{"sprint de otro proyecto", func(d *defect.Datos) { d.SprintID = 30 }, defect.CampoSprint, defect.ErrSprintInvalido},
		{"historia de otro proyecto", func(d *defect.Datos) { d.HistoriaID = 40 }, defect.CampoHistoria, defect.ErrHistoriaInvalida},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d := defectoValido()
			c.cambio(&d)
			_, err := nuevo(d)
			if c.err == nil {
				if err != nil {
					t.Fatalf("no se esperaba error: %v", err)
				}
				return
			}
			var errs domain.ErroresValidacion
			if !errors.As(err, &errs) || !errors.Is(errs[c.campo], c.err) {
				t.Errorf("error de %s = %v, se esperaba %v", c.campo, err, c.err)
			}
		})
	}
}

// RN6 · todos los errores juntos
func TestNuevo_InformaTodosLosErrores(t *testing.T) {
	_, err := nuevo(defect.Datos{HistoriaID: 40})
	var errs domain.ErroresValidacion
	if !errors.As(err, &errs) || len(errs) != 4 {
		t.Errorf("se esperaban 4 errores y se obtuvo %v", err)
	}
}

// El formulario ofrece las cuatro severidades, de la más grave a la más leve.
func TestSeveridades(t *testing.T) {
	if got := len(defect.Severidades); got != 4 || defect.Severidades[0] != defect.SeveridadCritica {
		t.Errorf("Severidades = %v", defect.Severidades)
	}
}
