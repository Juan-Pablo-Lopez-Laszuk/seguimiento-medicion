// Package account tiene las reglas de las cuentas de usuario (épica E4): qué email y qué contraseña se aceptan al
// registrarse y al ingresar.
//
// Spec: specs/TEC-05-autenticacion.md. No usa red ni base de datos: la cuenta la crea Supabase Auth (internal/auth);
// acá solo se validan los datos antes de pedírselo.
package account

import (
	"errors"
	"unicode/utf8"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
)

// Nombres de los campos del registro, tal como se informan en los errores y en el formulario.
const (
	CampoEmail      = "email"
	CampoClave      = "clave"
	CampoRepeticion = "repeticion"
)

// Largo de la contraseña (RN5): en caracteres, y en bytes el máximo, porque bcrypt (que usa Supabase) corta en 72.
const (
	ClaveMin = 8
	ClaveMax = 72
)

// Errores de validación (sección 7 de la spec).
var (
	ErrClaveFueraDeRango = errors.New("la contraseña tiene que tener entre 8 y 72 caracteres")
	ErrClavesDistintas   = errors.New("las contraseñas no coinciden")
	// ErrCredencialesIncompletas lo devuelve Ingreso cuando falta el email o la contraseña.
	ErrCredencialesIncompletas = errors.New("falta el email o la contraseña")
)

// DatosRegistro son los campos que carga el usuario para crear una cuenta.
type DatosRegistro struct {
	Email      string
	Clave      string
	Repeticion string
}

// Credenciales son el email (normalizado) y la contraseña con los que se crea una cuenta o se ingresa.
type Credenciales struct {
	Email string
	Clave string
}

// NuevoRegistro valida los datos de una cuenta nueva y devuelve sus credenciales. Si algún dato es inválido, devuelve
// domain.ErroresValidacion con todos los campos que fallaron (RN5).
func NuevoRegistro(d DatosRegistro) (Credenciales, error) {
	email := member.NormalizarEmail(d.Email)

	errs := domain.ErroresValidacion{}
	errs.Agregar(CampoEmail, member.ValidarFormatoEmail(email))
	errs.Agregar(CampoClave, validarLargoDeClave(d.Clave))
	if d.Repeticion != d.Clave {
		errs.Agregar(CampoRepeticion, ErrClavesDistintas)
	}
	if len(errs) > 0 {
		return Credenciales{}, errs
	}
	return Credenciales{Email: email, Clave: d.Clave}, nil
}

// Ingreso devuelve las credenciales para iniciar sesión con el email normalizado (RN6). No le exige largo a la
// contraseña: la valida Supabase. Si falta el email o la contraseña devuelve ErrCredencialesIncompletas.
func Ingreso(email, clave string) (Credenciales, error) {
	email = member.NormalizarEmail(email)
	if email == "" || clave == "" {
		return Credenciales{}, ErrCredencialesIncompletas
	}
	return Credenciales{Email: email, Clave: clave}, nil
}

func validarLargoDeClave(clave string) error {
	if caracteres := utf8.RuneCountInString(clave); caracteres < ClaveMin || caracteres > ClaveMax || len(clave) > ClaveMax {
		return ErrClaveFueraDeRango
	}
	return nil
}
