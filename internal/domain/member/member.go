// Package member tiene la entidad Integrante y sus reglas de negocio (épica E1).
//
// Spec: specs/HU-03-registrar-integrantes.md. No usa base de datos ni HTTP: recibe los datos y los
// integrantes que ya tiene el proyecto, y devuelve el integrante o los errores de validación.
package member

import "strings"

// Rol es el papel de un integrante en el equipo. Los valores son los mismos que guarda la base.
type Rol string

// Roles válidos (RN5).
const (
	RolAgileEnabler   Rol = "AgileEnabler"
	RolProductBuilder Rol = "ProductBuilder"
)

// Datos son los campos que carga el usuario para registrar un integrante.
type Datos struct {
	Nombre string
	Email  string
	Rol    Rol
}

// Integrante es una persona del equipo de un proyecto. Nunca se borra: se da de baja (RN7).
type Integrante struct {
	ID         int64
	ProyectoID int64
	Nombre     string
	Email      string
	Rol        Rol
	Activo     bool
}

// Nuevo valida los datos contra los integrantes que ya tiene el proyecto y devuelve el integrante
// listo para guardar.
func Nuevo(d Datos, proyectoID int64, _ []Integrante) (Integrante, error) {
	return Integrante{
		ProyectoID: proyectoID,
		Nombre:     strings.TrimSpace(d.Nombre),
		Email:      NormalizarEmail(d.Email),
		Rol:        d.Rol,
		Activo:     true,
	}, nil
}

// NormalizarEmail quita los espacios de alrededor y pasa el email a minúsculas (RN2).
func NormalizarEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
