package defect

import "testing"

// HU-29 · CA-29.1 · un defecto cumple el filtro si cumple todo lo que se eligió; lo vacío no filtra
func TestFiltro_Cumple(t *testing.T) {
	d := Defecto{Severidad: SeveridadAlta, Estado: EstadoAbierto, SprintDeteccionID: 3}
	for _, c := range []struct {
		filtro Filtro
		cumple bool
	}{
		{Filtro{}, true},
		{Filtro{Severidad: SeveridadAlta}, true},
		{Filtro{Severidad: SeveridadBaja}, false},
		{Filtro{Estado: EstadoAbierto, SprintID: 3}, true},
		{Filtro{Estado: EstadoCerrado}, false},
		{Filtro{Severidad: SeveridadAlta, Estado: EstadoAbierto, SprintID: 4}, false},
	} {
		if got := c.filtro.Cumple(d); got != c.cumple {
			t.Errorf("%+v.Cumple = %v, se esperaba %v", c.filtro, got, c.cumple)
		}
	}
}
