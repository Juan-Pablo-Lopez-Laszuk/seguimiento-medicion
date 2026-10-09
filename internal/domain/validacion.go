package domain

import (
	"sort"
	"strings"
)

// ErroresValidacion junta los errores de todos los campos inválidos de un formulario, por nombre de
// campo, para informarlos todos juntos. La usan las entidades de cada épica (proyecto, sprint...).
type ErroresValidacion map[string]error

// Error arma un texto con todos los errores, ordenados por campo: "nombre: el nombre es obligatorio".
func (e ErroresValidacion) Error() string {
	campos := make([]string, 0, len(e))
	for campo := range e {
		campos = append(campos, campo)
	}
	sort.Strings(campos)
	partes := make([]string, 0, len(campos))
	for _, campo := range campos {
		partes = append(partes, campo+": "+e[campo].Error())
	}
	return strings.Join(partes, "; ")
}

// Agregar registra el error del campo solo si hay error.
func (e ErroresValidacion) Agregar(campo string, err error) {
	if err != nil {
		e[campo] = err
	}
}
