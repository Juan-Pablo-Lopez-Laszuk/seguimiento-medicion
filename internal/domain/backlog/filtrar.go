package backlog

import (
	"slices"
	"strings"
)

// Filtro son los filtros de la página del backlog (HU-10): un estado, una prioridad y un texto que se busca en el
// título. Un campo vacío no filtra.
//
// Spec: specs/HU-10-filtrar-buscar-historias.md.
type Filtro struct {
	Estado    Estado
	Prioridad Prioridad
	Texto     string
}

// NuevoFiltro arma el filtro con los valores que llegan en la URL. Un estado o una prioridad que no existe se ignora
// (RN4) y al texto se le quitan los espacios de los extremos.
func NuevoFiltro(estado, prioridad, texto string) Filtro {
	var f Filtro
	if e := Estado(estado); slices.Contains(Estados, e) {
		f.Estado = e
	}
	if p := Prioridad(prioridad); p.valida() {
		f.Prioridad = p
	}
	f.Texto = strings.TrimSpace(texto)
	return f
}

// Activo dice si hay algún filtro puesto (RN5).
func (f Filtro) Activo() bool {
	return f.Estado != "" || f.Prioridad != "" || f.Texto != ""
}

// Filtrar devuelve, en una lista nueva y en el mismo orden, las historias que cumplen todos los filtros (RN1 y RN3).
// El texto se busca en el título sin distinguir mayúsculas ni tildes (RN2).
func Filtrar(historias []Historia, f Filtro) []Historia {
	texto := normalizar(f.Texto)
	var cumplen []Historia
	for _, h := range historias {
		if (f.Estado == "" || h.Estado == f.Estado) &&
			(f.Prioridad == "" || h.Prioridad == f.Prioridad) &&
			strings.Contains(normalizar(h.Titulo), texto) {
			cumplen = append(cumplen, h)
		}
	}
	return cumplen
}

// sinTildes cambia cada vocal con tilde o diéresis por la vocal sola. La ñ no se toca: "diseño" no es "diseno".
var sinTildes = strings.NewReplacer("á", "a", "é", "e", "í", "i", "ó", "o", "ú", "u", "ü", "u")

// normalizar pasa el texto a minúsculas y le saca las tildes, para comparar como lo haría una persona (RN2).
func normalizar(texto string) string {
	return sinTildes.Replace(strings.ToLower(texto))
}
