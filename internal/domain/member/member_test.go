package member_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
)

const proyectoID int64 = 7

// datosValidos devuelve datos que cumplen todas las reglas; cada test cambia solo lo que quiere probar.
func datosValidos() member.Datos {
	return member.Datos{Nombre: "Juan Pablo", Email: "jp@mail.com", Rol: member.RolAgileEnabler}
}

// verificarErrorDeCampo comprueba que err sea un error de validación con el error esperado en ese campo.
func verificarErrorDeCampo(t *testing.T, err error, campo string, esperado error) {
	t.Helper()
	var errs domain.ErroresValidacion
	if !errors.As(err, &errs) {
		t.Fatalf("se esperaba domain.ErroresValidacion y se obtuvo: %v", err)
	}
	if !errors.Is(errs[campo], esperado) {
		t.Errorf("campo %q: se esperaba %v y se obtuvo %v (todos: %v)", campo, esperado, errs[campo], errs)
	}
}

// CA-03.1 y CA-03.2 · RN1 y RN2: se registra activo, con el nombre sin espacios y el email en minúsculas
func TestNuevo_DatosValidos_RegistraAlIntegranteActivoYNormalizado(t *testing.T) {
	d := datosValidos()
	d.Nombre = "  Juan Pablo  "
	d.Email = " JuanPablo@Mail.COM "

	i, err := member.Nuevo(d, proyectoID, nil)

	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo: %v", err)
	}
	if i.Nombre != "Juan Pablo" || i.Email != "juanpablo@mail.com" {
		t.Errorf("se esperaba %q <%s> y se obtuvo %q <%s>", "Juan Pablo", "juanpablo@mail.com", i.Nombre, i.Email)
	}
	if i.ProyectoID != proyectoID || i.Rol != member.RolAgileEnabler || !i.Activo {
		t.Errorf("se esperaba proyecto %d, Agile Enabler y activo, y se obtuvo %+v", proyectoID, i)
	}
}

// RN1 · el nombre es obligatorio y tiene hasta 100 caracteres (se cuentan caracteres, no bytes)
func TestNuevo_Nombre(t *testing.T) {
	casos := []struct {
		nombre   string
		esperado error // nil si es válido
	}{
		{"", member.ErrNombreObligatorio},
		{"   ", member.ErrNombreObligatorio},
		{"Ñ", nil},
		{strings.Repeat("ñ", member.NombreMax), nil}, // 100 caracteres, 200 bytes
		{strings.Repeat("a", member.NombreMax+1), member.ErrNombreLargo},
	}
	for _, c := range casos {
		d := datosValidos()
		d.Nombre = c.nombre

		_, err := member.Nuevo(d, proyectoID, nil)

		if c.esperado == nil && err != nil {
			t.Errorf("nombre de %d caracteres: no se esperaba error, se obtuvo %v", len([]rune(c.nombre)), err)
		}
		if c.esperado != nil {
			verificarErrorDeCampo(t, err, member.CampoNombre, c.esperado)
		}
	}
}

// CA-03.1 · RN3: el email es obligatorio y tiene formato válido
func TestNuevo_FormatoDelEmail(t *testing.T) {
	largo := strings.Repeat("a", 64) + "@" + strings.Repeat("b", 185) + ".com" // 254 caracteres
	casos := []struct {
		email    string
		esperado error // nil si es válido
	}{
		{"ana@mail.com", nil},
		{"ana.perez+tpi@frsr.utn.edu.ar", nil},
		{largo, nil},
		{"", member.ErrEmailObligatorio},
		{"   ", member.ErrEmailObligatorio},
		{"ana@mail", member.ErrEmailInvalido},                   // dominio sin punto
		{"ana.mail.com", member.ErrEmailInvalido},               // sin @
		{"@mail.com", member.ErrEmailInvalido},                  // sin usuario
		{"ana@.com", member.ErrEmailInvalido},                   // dominio que empieza con punto
		{"ana@mail.", member.ErrEmailInvalido},                  // dominio que termina con punto
		{"Ana <ana@mail.com>", member.ErrEmailInvalido},         // con nombre delante
		{"<ana@mail.com>", member.ErrEmailInvalido},             // entre < >
		{"ana@mail.com, bea@mail.com", member.ErrEmailInvalido}, // dos direcciones
		{"a" + largo, member.ErrEmailInvalido},                  // 255 caracteres
	}
	for _, c := range casos {
		d := datosValidos()
		d.Email = c.email

		_, err := member.Nuevo(d, proyectoID, nil)

		if c.esperado == nil && err != nil {
			t.Errorf("email %q: no se esperaba error, se obtuvo %v", c.email, err)
		}
		if c.esperado != nil {
			verificarErrorDeCampo(t, err, member.CampoEmail, c.esperado)
		}
	}
}
