package account_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/account"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
)

// datosValidos devuelve un registro que cumple todas las reglas; cada test cambia solo lo que quiere probar.
func datosValidos() account.DatosRegistro {
	return account.DatosRegistro{Email: "ana@mail.com", Clave: "clave-segura-1", Repeticion: "clave-segura-1"}
}

// errorDeCampo devuelve el error del campo si err es un error de validación; falla el test si no lo es.
func errorDeCampo(t *testing.T, err error, campo string) error {
	t.Helper()
	var errs domain.ErroresValidacion
	if !errors.As(err, &errs) {
		t.Fatalf("se esperaba domain.ErroresValidacion y se obtuvo: %v", err)
	}
	return errs[campo]
}

// CA-05.1 · RN5: el email se guarda sin espacios y en minúsculas, y la contraseña tal cual se escribió
func TestNuevoRegistro_DatosValidos_NormalizaElEmailYConservaLaClave(t *testing.T) {
	d := datosValidos()
	d.Email = "  Ana@Mail.COM "

	c, err := account.NuevoRegistro(d)

	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo: %v", err)
	}
	if c.Email != "ana@mail.com" || c.Clave != "clave-segura-1" {
		t.Errorf("se esperaba ana@mail.com con la clave sin cambios y se obtuvo %q con %q", c.Email, c.Clave)
	}
}

// RN5 · el email tiene el mismo formato que el de un integrante (HU-03 RN2 y RN3)
func TestNuevoRegistro_Email(t *testing.T) {
	casos := []struct {
		email    string
		esperado error // nil si es válido
	}{
		{"", member.ErrEmailObligatorio},
		{"   ", member.ErrEmailObligatorio},
		{"ana@mail", member.ErrEmailInvalido},
		{"Ana <ana@mail.com>", member.ErrEmailInvalido},
		{"ana@mail.com", nil},
	}
	for _, c := range casos {
		d := datosValidos()
		d.Email = c.email

		_, err := account.NuevoRegistro(d)

		if c.esperado == nil {
			if err != nil {
				t.Errorf("email %q: no se esperaba error y se obtuvo %v", c.email, err)
			}
			continue
		}
		if got := errorDeCampo(t, err, account.CampoEmail); !errors.Is(got, c.esperado) {
			t.Errorf("email %q: se esperaba %v y se obtuvo %v", c.email, c.esperado, got)
		}
	}
}

// RN5 y decisión 4 · la contraseña tiene entre 8 y 72 caracteres y no más de 72 bytes (el límite de bcrypt)
func TestNuevoRegistro_LargoDeLaClave(t *testing.T) {
	casos := []struct {
		nombre   string
		clave    string
		esperado error // nil si es válida
	}{
		{"vacía", "", account.ErrClaveFueraDeRango},
		{"7 caracteres", strings.Repeat("a", 7), account.ErrClaveFueraDeRango},
		{"8 caracteres", strings.Repeat("a", 8), nil},
		{"72 caracteres", strings.Repeat("a", 72), nil},
		{"73 caracteres", strings.Repeat("a", 73), account.ErrClaveFueraDeRango},
		{"36 caracteres de 2 bytes (72 bytes)", strings.Repeat("ñ", 36), nil},
		{"37 caracteres de 2 bytes (74 bytes)", strings.Repeat("ñ", 37), account.ErrClaveFueraDeRango},
		{"4 caracteres de 2 bytes (8 bytes)", strings.Repeat("ñ", 4), account.ErrClaveFueraDeRango},
		{"8 espacios", strings.Repeat(" ", 8), nil},
	}
	for _, c := range casos {
		d := datosValidos()
		d.Clave, d.Repeticion = c.clave, c.clave

		_, err := account.NuevoRegistro(d)

		if c.esperado == nil {
			if err != nil {
				t.Errorf("%s: no se esperaba error y se obtuvo %v", c.nombre, err)
			}
			continue
		}
		if got := errorDeCampo(t, err, account.CampoClave); !errors.Is(got, c.esperado) {
			t.Errorf("%s: se esperaba %v y se obtuvo %v", c.nombre, c.esperado, got)
		}
	}
}

// RN5 · la repetición tiene que ser igual a la contraseña
func TestNuevoRegistro_LaRepeticionTieneQueCoincidir(t *testing.T) {
	d := datosValidos()
	d.Repeticion = "clave-segura-2"

	_, err := account.NuevoRegistro(d)

	if got := errorDeCampo(t, err, account.CampoRepeticion); !errors.Is(got, account.ErrClavesDistintas) {
		t.Errorf("se esperaba %v y se obtuvo %v", account.ErrClavesDistintas, got)
	}
	if got := errorDeCampo(t, err, account.CampoClave); got != nil {
		t.Errorf("la contraseña es válida y no se esperaba error en ese campo: %v", got)
	}
}

// RN5 · los errores de todos los campos se informan juntos
func TestNuevoRegistro_InformaTodosLosErroresJuntos(t *testing.T) {
	_, err := account.NuevoRegistro(account.DatosRegistro{Email: "ana@mail", Clave: "corta", Repeticion: "otra"})

	var errs domain.ErroresValidacion
	if !errors.As(err, &errs) || len(errs) != 3 {
		t.Fatalf("se esperaban los errores del email, la clave y la repetición y se obtuvo: %v", err)
	}
}

// RN5 · una repetición que no coincide también se informa cuando la contraseña es inválida
func TestNuevoRegistro_ClaveInvalidaYRepeticionDistinta_InformaLasDos(t *testing.T) {
	d := datosValidos()
	d.Clave, d.Repeticion = "corta", "corta2"

	_, err := account.NuevoRegistro(d)

	if got := errorDeCampo(t, err, account.CampoClave); !errors.Is(got, account.ErrClaveFueraDeRango) {
		t.Errorf("clave: se esperaba %v y se obtuvo %v", account.ErrClaveFueraDeRango, got)
	}
	if got := errorDeCampo(t, err, account.CampoRepeticion); !errors.Is(got, account.ErrClavesDistintas) {
		t.Errorf("repetición: se esperaba %v y se obtuvo %v", account.ErrClavesDistintas, got)
	}
}

// RN6 · al ingresar el email se normaliza y a la contraseña no se le exige largo (la valida Supabase)
func TestIngreso_NormalizaElEmailYNoExigeLargoDeClave(t *testing.T) {
	c, err := account.Ingreso(" Ana@Mail.com ", "x")

	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo: %v", err)
	}
	if c.Email != "ana@mail.com" || c.Clave != "x" {
		t.Errorf("se esperaba ana@mail.com con la clave x y se obtuvo %q con %q", c.Email, c.Clave)
	}
}

// RN6 · sin email o sin contraseña no hay nada que consultar: las credenciales están incompletas
func TestIngreso_FaltaElEmailOLaClave_Incompletas(t *testing.T) {
	for nombre, c := range map[string][2]string{"sin email": {"", "clave"}, "email en blanco": {"  ", "clave"}, "sin contraseña": {"ana@mail.com", ""}} {
		if _, err := account.Ingreso(c[0], c[1]); !errors.Is(err, account.ErrCredencialesIncompletas) {
			t.Errorf("%s: se esperaba %v y se obtuvo %v", nombre, account.ErrCredencialesIncompletas, err)
		}
	}
}
