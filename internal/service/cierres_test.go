package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// Los tests de HU-14 usan el escenario de HU-12 (Sprint 1 cerrado, Sprint 2 activo del 12/10 al 18/10, Sprint 3
// futuro; hoy es el 15/10) con tres historias en el Sprint 2: una Hecha (5 SP), una En progreso (3 SP) y una En sprint
// (5 SP), como en los casos límite de la spec.
type escenarioCierre struct {
	*escenarioAsignacion
	cierres *memory.Cierres
	hecha   int // números de las historias del Sprint 2
	enCurso int
	sinEmp  int
}

func nuevoEscenarioCierre(t *testing.T) *escenarioCierre {
	t.Helper()
	e := &escenarioCierre{escenarioAsignacion: nuevoEscenarioAsignacion(t, true)}
	e.cierres = memory.NuevoCierres(e.sprints, e.historias)
	e.hecha = e.enSprintCon(t, 2, 5, backlog.EstadoHecho, 1, 1)
	e.enCurso = e.enSprintCon(t, 2, 3, backlog.EstadoEnProgreso, 2, 1)
	e.sinEmp = e.enSprintCon(t, 2, 5, backlog.EstadoEnSprint, 1, 0)
	return e
}

// casosDeCierre arma el caso de uso con "hoy" en ese momento.
func (e *escenarioCierre) casosDeCierre(ahora time.Time) *service.Cierres {
	return service.NuevoCierres(e.proyectos, e.sprints, e.historias, e.cierres, func() time.Time { return ahora })
}

func (e *escenarioCierre) sprintGuardado(t *testing.T, numero int) sprint.Sprint {
	t.Helper()
	sprints, _ := e.sprints.ListarPorProyecto(context.Background(), e.proyectoID)
	for _, s := range sprints {
		if s.Numero == numero {
			return s
		}
	}
	t.Fatalf("no está el Sprint %d", numero)
	return sprint.Sprint{}
}

// CA-14.2 y CA-14.3 · cerrar el sprint activo guarda el sprint Cerrado con sus SP, la foto y las historias devueltas
func TestCerrarSprint_GuardaElCierre(t *testing.T) {
	e := nuevoEscenarioCierre(t)

	c, err := e.casosDeCierre(relojDelSprint()).Cerrar(context.Background(), e.proyectoID, 2)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if c.Sprint.Numero != 2 || *c.Sprint.SPPlanificados != 13 || *c.Sprint.SPCompletados != 5 || len(c.Devueltas) != 2 {
		t.Errorf("se esperaba el cierre del Sprint 2 con 5 de 13 SP y 2 historias devueltas, y se obtuvo %+v", c)
	}
	s := e.sprintGuardado(t, 2)
	if s.Estado != sprint.EstadoCerrado || s.SPPlanificados == nil || *s.SPPlanificados != 13 || *s.SPCompletados != 5 {
		t.Errorf("el Sprint 2 tenía que quedar guardado Cerrado con 5 de 13 SP y quedó %+v", s)
	}
	for _, numero := range []int{e.enCurso, e.sinEmp} {
		if h := e.guardada(t, numero); h.Estado != backlog.EstadoPendiente || h.SprintID != nil {
			t.Errorf("HU-%d tenía que quedar Pendiente y sin sprint, y quedó %+v", numero, h)
		}
	}
	if h := e.guardada(t, e.hecha); h.Estado != backlog.EstadoHecho || h.SprintID == nil || *h.SprintID != e.ids[2] {
		t.Errorf("la Hecha tenía que seguir Hecha en el Sprint 2 y quedó %+v", h)
	}
	if foto, _ := e.cierres.Foto(context.Background(), e.ids[2]); len(foto) != 3 {
		t.Errorf("la foto del Sprint 2 tenía que tener 3 historias y tiene %+v", foto)
	}
	// las del backlog no se tocan
	if h := e.guardada(t, 1); h.Estado != backlog.EstadoPendiente || h.SprintID != nil {
		t.Errorf("HU-1 del backlog no tenía que cambiar y quedó %+v", h)
	}
}

// Sección 3 · la confirmación dice qué va a pasar (SP, historias que vuelven con su estado y días que faltaban) y no
// guarda nada
func TestPrepararCierre_MuestraLoQueVaAPasar(t *testing.T) {
	e := nuevoEscenarioCierre(t)

	v, err := e.casosDeCierre(relojDelSprint()).Preparar(context.Background(), e.proyectoID, 2)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if v.Proyecto.ID != e.proyectoID || v.Sprint.Numero != 2 || v.Planificados != 13 || v.Completados != 5 || v.Historias != 3 || v.Hechas != 1 {
		t.Errorf("se esperaban 5 de 13 SP y 1 de 3 historias del Sprint 2, y se obtuvo %+v", v)
	}
	if len(v.Devueltas) != 2 || v.Devueltas[0].Numero != e.enCurso || v.Devueltas[0].Estado != backlog.EstadoEnProgreso ||
		v.Devueltas[1].Numero != e.sinEmp || v.Devueltas[1].Estado != backlog.EstadoEnSprint {
		t.Errorf("se esperaban las dos que vuelven al backlog con su estado de ahora, y se obtuvo %+v", v.Devueltas)
	}
	if v.DiasAntes != 3 {
		t.Errorf("el 15/10 faltan 3 días para el fin (18/10) y se obtuvo %d", v.DiasAntes)
	}
	if s := e.sprintGuardado(t, 2); s.Estado == sprint.EstadoCerrado {
		t.Error("mostrar la confirmación no tiene que cerrar el sprint")
	}
}

// RN1 y RN2 · el último día y con el sprint vencido no faltan días, y se puede cerrar igual
func TestPrepararCierre_UltimoDiaYVencido(t *testing.T) {
	for _, ahora := range []time.Time{
		time.Date(2026, 10, 18, 22, 0, 0, 0, time.UTC), // 19:00 del 18/10 en Argentina
		time.Date(2026, 10, 19, 2, 0, 0, 0, time.UTC),  // 23:00 del 18/10 en Argentina (en UTC ya es el 19)
		time.Date(2026, 10, 20, 15, 0, 0, 0, time.UTC), // vencido sin cerrar (con el Sprint 3 ya empezado)
	} {
		e := nuevoEscenarioCierre(t)
		v, err := e.casosDeCierre(ahora).Preparar(context.Background(), e.proyectoID, 2)
		if err != nil || v.DiasAntes != 0 {
			t.Errorf("%v: se esperaba poder cerrar sin días de anticipación y se obtuvo %d, %v", ahora, v.DiasAntes, err)
		}
	}
}

// CA-14.1 y CA-14.4 · sección 7 · los rechazos devuelven el error de la regla (con el sprint para mostrarlo) y no
// guardan nada
func TestCerrarSprint_Errores(t *testing.T) {
	casos := []struct {
		nombre  string
		numero  int
		base    error
		mensaje string
	}{
		{"un sprint futuro", 3, sprint.ErrNoEsElActivo, "solo se cierra el sprint activo (el Sprint 2)"},
		{"un sprint cerrado", 1, sprint.ErrYaCerrado, "el Sprint 1 ya está cerrado"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			e := nuevoEscenarioCierre(t)
			cierres := e.casosDeCierre(relojDelSprint())

			v, err := cierres.Preparar(context.Background(), e.proyectoID, c.numero)
			if !errors.Is(err, c.base) || err.Error() != c.mensaje || v.Sprint.Numero != c.numero {
				t.Errorf("Preparar: se esperaba %q con el Sprint %d, y se obtuvo %v con %+v", c.mensaje, c.numero, err, v.Sprint)
			}
			if _, err := cierres.Cerrar(context.Background(), e.proyectoID, c.numero); !errors.Is(err, c.base) {
				t.Errorf("Cerrar: se esperaba %v y se obtuvo %v", c.base, err)
			}
			if s := e.sprintGuardado(t, 2); s.Estado == sprint.EstadoCerrado {
				t.Error("un rechazo no tiene que cerrar ningún sprint")
			}
		})
	}
}

// Sección 7 · el proyecto o el sprint que no existen
func TestCerrarSprint_NoEncontrados(t *testing.T) {
	e := nuevoEscenarioCierre(t)
	casos := e.casosDeCierre(relojDelSprint())
	if _, err := casos.Cerrar(context.Background(), 99, 2); !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("proyecto que no existe: se esperaba project.ErrNoEncontrado y se obtuvo %v", err)
	}
	if _, err := casos.Preparar(context.Background(), e.proyectoID, 9); !errors.Is(err, sprint.ErrNoEncontrado) {
		t.Errorf("sprint que no existe: se esperaba sprint.ErrNoEncontrado y se obtuvo %v", err)
	}
}

// Sección 3 · la página de un sprint cerrado muestra los SP de la foto, aunque las no completadas ya no estén en él
func TestVerSprint_CerradoMuestraLosSPDelCierre(t *testing.T) {
	e := nuevoEscenarioCierre(t)
	if _, err := e.casosDeCierre(relojDelSprint()).Cerrar(context.Background(), e.proyectoID, 2); err != nil {
		t.Fatalf("cerrar: %v", err)
	}

	v, err := e.casos.Ver(context.Background(), e.proyectoID, 2)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if v.Abierto || v.Activo || v.Comprometidos != 13 || v.Completados != 5 || v.Aviso != "" {
		t.Errorf("se esperaba el Sprint 2 cerrado con 5 de 13 SP y sin aviso, y se obtuvo abierto=%v activo=%v %d de %d %q",
			v.Abierto, v.Activo, v.Completados, v.Comprometidos, v.Aviso)
	}
}
