// Package member tiene la entidad Integrante y sus reglas de negocio (épica E1).
//
// Spec: specs/HU-03-registrar-integrantes.md. No usa base de datos ni HTTP: recibe los datos y los
// integrantes que ya tiene el proyecto, y devuelve el integrante o los errores de validación.
package member

import (
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
)

// Rol es el papel de un integrante en el equipo. Los valores son los mismos que guarda la base.
type Rol string

// Roles válidos (RN5).
const (
	RolAgileEnabler   Rol = "AgileEnabler"
	RolProductBuilder Rol = "ProductBuilder"
)

// Roles es la lista de roles válidos, en el orden en que se ofrecen en el formulario.
var Roles = []Rol{RolAgileEnabler, RolProductBuilder}

// Texto devuelve el rol como se muestra en pantalla ("Agile Enabler").
func (r Rol) Texto() string {
	switch r {
	case RolAgileEnabler:
		return "Agile Enabler"
	case RolProductBuilder:
		return "Product Builder"
	}
	return string(r)
}

func (r Rol) valido() bool {
	return r == RolAgileEnabler || r == RolProductBuilder
}

// Nombres de los campos, tal como se informan en los errores y en el formulario.
const (
	CampoNombre = "nombre"
	CampoEmail  = "email"
	CampoRol    = "rol"
)

// Límites, en caracteres: el del nombre (RN1) y el máximo de un email según el estándar (RN3).
const (
	NombreMax = 100
	EmailMax  = 254
)

// Errores de validación (sección 7 de la spec).
var (
	ErrNombreObligatorio = errors.New("el nombre es obligatorio")
	ErrNombreLargo       = errors.New("el nombre no puede superar los 100 caracteres")
	ErrEmailObligatorio  = errors.New("el email es obligatorio")
	ErrEmailInvalido     = errors.New("el email no tiene un formato válido")
	ErrRolInvalido       = errors.New("el rol tiene que ser Agile Enabler o Product Builder")
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
// listo para guardar. Si algún dato es inválido, devuelve domain.ErroresValidacion con todos los
// campos que fallaron (RN9).
func Nuevo(d Datos, proyectoID int64, _ []Integrante) (Integrante, error) {
	nombre, email := strings.TrimSpace(d.Nombre), NormalizarEmail(d.Email)

	errs := domain.ErroresValidacion{}
	errs.Agregar(CampoNombre, validarNombre(nombre))
	errs.Agregar(CampoEmail, validarEmail(email))
	if !d.Rol.valido() {
		errs.Agregar(CampoRol, ErrRolInvalido)
	}
	if len(errs) > 0 {
		return Integrante{}, errs
	}
	return Integrante{
		ProyectoID: proyectoID,
		Nombre:     nombre,
		Email:      email,
		Rol:        d.Rol,
		Activo:     true,
	}, nil
}

// NormalizarEmail quita los espacios de alrededor y pasa el email a minúsculas (RN2).
func NormalizarEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validarNombre(nombre string) error {
	switch largo := utf8.RuneCountInString(nombre); {
	case largo == 0:
		return ErrNombreObligatorio
	case largo > NombreMax:
		return ErrNombreLargo
	}
	return nil
}

// validarEmail controla el formato (RN3): una sola dirección, sin nombre delante, con usuario y con un
// dominio que tenga al menos un punto y no empiece ni termine con punto.
func validarEmail(email string) error {
	if email == "" {
		return ErrEmailObligatorio
	}
	if len(email) > EmailMax {
		return ErrEmailInvalido
	}
	dir, err := mail.ParseAddress(email)
	if err != nil || dir.Name != "" || dir.Address != email {
		return ErrEmailInvalido
	}
	_, dominio, _ := strings.Cut(email, "@")
	if !strings.Contains(dominio, ".") || strings.HasPrefix(dominio, ".") || strings.HasSuffix(dominio, ".") {
		return ErrEmailInvalido
	}
	return nil
}
