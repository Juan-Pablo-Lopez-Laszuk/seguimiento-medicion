package defect_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/defect"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
)

// Sprints del ejemplo de la spec de HU-28: Sprint 1 (id 3, detección) y Sprint 2 (id 5) del proyecto 7; el 30 es de
// otro proyecto.
func sprintsUnoYDos() []sprint.Sprint {
	return []sprint.Sprint{{ID: 3, ProyectoID: 7, Numero: 1}, {ID: 5, ProyectoID: 7, Numero: 2}, {ID: 30, ProyectoID: 8, Numero: 3}}
}

func defectoEn(estado defect.Estado, detectadoEn, resueltoEn int64) defect.Defecto {
	return defect.Defecto{ID: 1, ProyectoID: 7, Descripcion: "Falla", Severidad: defect.SeveridadAlta, Estado: estado,
		SprintDeteccionID: detectadoEn, SprintResolucionID: resueltoEn}
}

// CA-28.1 y CA-28.2 · el camino completo hasta Cerrado; Cerrado conserva el sprint de resolución (RN4)
func TestCambiarEstado_CaminoCompleto(t *testing.T) {
	d := defectoEn(defect.EstadoAbierto, 3, 0)
	pasos := []struct {
		estado defect.Estado
		sprint int64
	}{{defect.EstadoEnProgreso, 0}, {defect.EstadoResuelto, 5}, {defect.EstadoCerrado, 0}}
	for _, p := range pasos {
		var err error
		if d, err = d.CambiarEstado(p.estado, p.sprint, sprintsUnoYDos()); err != nil {
			t.Fatalf("a %s: no se esperaba error: %v", p.estado, err)
		}
	}
	if d.Estado != defect.EstadoCerrado || d.SprintResolucionID != 5 {
		t.Errorf("defecto = %+v, se esperaba Cerrado con el Sprint 2 (id 5) de resolución", d)
	}
}

// CA-28.2 · el sprint de resolución es del proyecto y no anterior al de detección; el mismo vale
func TestCambiarEstado_SprintDeResolucion(t *testing.T) {
	casos := []struct {
		nombre     string
		detectado  int64
		resolucion int64
		err        error
	}{
		{"el mismo sprint", 3, 3, nil},
		{"un sprint posterior", 3, 5, nil},
		{"un sprint anterior", 5, 3, defect.ErrSprintResolucionAnterior},
		{"sin sprint", 3, 0, defect.ErrSprintResolucionInvalido},
		{"de otro proyecto", 3, 30, defect.ErrSprintResolucionInvalido},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			d, err := defectoEn(defect.EstadoEnProgreso, c.detectado, 0).CambiarEstado(defect.EstadoResuelto, c.resolucion, sprintsUnoYDos())
			if !errors.Is(err, c.err) {
				t.Fatalf("error = %v, se esperaba %v", err, c.err)
			}
			if c.err == nil && (d.Estado != defect.EstadoResuelto || d.SprintResolucionID != c.resolucion) {
				t.Errorf("defecto = %+v", d)
			}
		})
	}
}

// CA-28.3 · reabrir borra el sprint de resolución; un reabierto vuelve a En progreso
func TestCambiarEstado_Reabrir(t *testing.T) {
	d, err := defectoEn(defect.EstadoResuelto, 3, 5).CambiarEstado(defect.EstadoReabierto, 0, sprintsUnoYDos())
	if err != nil || d.Estado != defect.EstadoReabierto || d.SprintResolucionID != 0 {
		t.Fatalf("se esperaba Reabierto sin sprint de resolución y se obtuvo %+v, %v", d, err)
	}
	if d, err = d.CambiarEstado(defect.EstadoEnProgreso, 0, sprintsUnoYDos()); err != nil || d.Estado != defect.EstadoEnProgreso {
		t.Errorf("de Reabierto a En progreso: %+v, %v", d, err)
	}
}

// RN1 · las transiciones que no están permitidas son un error que dice de dónde a dónde
func TestCambiarEstado_TransicionesInvalidas(t *testing.T) {
	casos := []struct{ de, a defect.Estado }{
		{defect.EstadoAbierto, defect.EstadoResuelto}, {defect.EstadoAbierto, defect.EstadoCerrado},
		{defect.EstadoAbierto, defect.EstadoReabierto}, {defect.EstadoEnProgreso, defect.EstadoCerrado},
		{defect.EstadoCerrado, defect.EstadoReabierto}, {defect.EstadoAbierto, defect.EstadoAbierto},
	}
	for _, c := range casos {
		_, err := defectoEn(c.de, 3, 0).CambiarEstado(c.a, 5, sprintsUnoYDos())
		if !errors.Is(err, defect.ErrTransicionInvalida) {
			t.Errorf("%s → %s: error = %v, se esperaba %v", c.de, c.a, err, defect.ErrTransicionInvalida)
			continue
		}
		if esperado := "un defecto en estado " + string(c.de) + " no puede pasar a " + string(c.a); err.Error() != esperado {
			t.Errorf("mensaje = %q, se esperaba %q", err, esperado)
		}
	}
	if _, err := defectoEn(defect.EstadoAbierto, 3, 0).CambiarEstado("Perdido", 0, nil); !errors.Is(err, defect.ErrEstadoInvalido) {
		t.Errorf("estado que no existe: error = %v, se esperaba %v", err, defect.ErrEstadoInvalido)
	}
}

// La lista muestra solo los botones de las transiciones que valen.
func TestSiguientes(t *testing.T) {
	casos := map[defect.Estado][]defect.Estado{
		defect.EstadoAbierto:    {defect.EstadoEnProgreso},
		defect.EstadoEnProgreso: {defect.EstadoResuelto},
		defect.EstadoResuelto:   {defect.EstadoCerrado, defect.EstadoReabierto},
		defect.EstadoReabierto:  {defect.EstadoEnProgreso},
		defect.EstadoCerrado:    nil,
	}
	for de, esperado := range casos {
		if got := defect.Siguientes(de); !slices.Equal(got, esperado) {
			t.Errorf("Siguientes(%s) = %v, se esperaba %v", de, got, esperado)
		}
	}
}
