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

// CA-03.2 · RN5: el rol es Agile Enabler o Product Builder
func TestNuevo_Rol(t *testing.T) {
	casos := map[member.Rol]bool{
		member.RolAgileEnabler:   true,
		member.RolProductBuilder: true,
		"":                       false,
		"ScrumMaster":            false,
		"agileenabler":           false, // los valores son exactos, como en la base
	}
	for rol, valido := range casos {
		d := datosValidos()
		d.Rol = rol

		_, err := member.Nuevo(d, proyectoID, nil)

		if valido && err != nil {
			t.Errorf("rol %q: no se esperaba error, se obtuvo %v", rol, err)
		}
		if !valido {
			verificarErrorDeCampo(t, err, member.CampoRol, member.ErrRolInvalido)
		}
	}
}

// En pantalla el rol se muestra con espacio: "Agile Enabler".
func TestRol_Texto(t *testing.T) {
	casos := map[member.Rol]string{
		member.RolAgileEnabler:   "Agile Enabler",
		member.RolProductBuilder: "Product Builder",
		"Otro":                   "Otro",
	}
	for rol, esperado := range casos {
		if got := rol.Texto(); got != esperado {
			t.Errorf("%q.Texto(): se esperaba %q y se obtuvo %q", rol, esperado, got)
		}
	}
}

// RN9 · si hay varios datos inválidos se informan todos juntos
func TestNuevo_InformaTodosLosCamposInvalidosJuntos(t *testing.T) {
	d := member.Datos{Nombre: "   ", Email: "ana@mail", Rol: "ScrumMaster"}

	_, err := member.Nuevo(d, proyectoID, nil)

	verificarErrorDeCampo(t, err, member.CampoNombre, member.ErrNombreObligatorio)
	verificarErrorDeCampo(t, err, member.CampoEmail, member.ErrEmailInvalido)
	verificarErrorDeCampo(t, err, member.CampoRol, member.ErrRolInvalido)
}

// existente arma un integrante ya guardado en el proyecto de ejemplo.
func existente(id int64, email string, rol member.Rol, activo bool) member.Integrante {
	return member.Integrante{ID: id, ProyectoID: proyectoID, Nombre: "Alguien", Email: email, Rol: rol, Activo: activo}
}

// CA-03.1 · RN4: el email no se repite en el proyecto, sin distinguir mayúsculas y contando a los dados de baja
func TestNuevo_EmailRepetido(t *testing.T) {
	casos := []struct {
		nombre     string
		existentes []member.Integrante
		repetido   bool
	}{
		{"mismo email", []member.Integrante{existente(1, "ana@mail.com", member.RolProductBuilder, true)}, true},
		{"dado de baja", []member.Integrante{existente(1, "ana@mail.com", member.RolProductBuilder, false)}, true},
		{"otro email", []member.Integrante{existente(1, "bea@mail.com", member.RolProductBuilder, true)}, false},
	}
	for _, c := range casos {
		d := datosValidos()
		d.Email = " ANA@Mail.com "
		d.Rol = member.RolProductBuilder

		_, err := member.Nuevo(d, proyectoID, c.existentes)

		if !c.repetido && err != nil {
			t.Errorf("%s: no se esperaba error, se obtuvo %v", c.nombre, err)
		}
		if c.repetido {
			verificarErrorDeCampo(t, err, member.CampoEmail, member.ErrEmailRepetido)
		}
	}
}

// CA-03.2 · RN6: un solo Agile Enabler activo; los dados de baja no cuentan y Product Builder puede haber varios
func TestNuevo_UnSoloAgileEnablerActivo(t *testing.T) {
	casos := []struct {
		nombre     string
		rol        member.Rol
		existentes []member.Integrante
		valido     bool
	}{
		{"segundo Agile Enabler", member.RolAgileEnabler,
			[]member.Integrante{existente(1, "jp@mail.com", member.RolAgileEnabler, true)}, false},
		{"el anterior está dado de baja", member.RolAgileEnabler,
			[]member.Integrante{existente(1, "jp@mail.com", member.RolAgileEnabler, false)}, true},
		{"varios Product Builder", member.RolProductBuilder,
			[]member.Integrante{
				existente(1, "jp@mail.com", member.RolAgileEnabler, true),
				existente(2, "mp@mail.com", member.RolProductBuilder, true),
			}, true},
	}
	for _, c := range casos {
		d := datosValidos()
		d.Email = "nuevo@mail.com"
		d.Rol = c.rol

		_, err := member.Nuevo(d, proyectoID, c.existentes)

		if c.valido && err != nil {
			t.Errorf("%s: no se esperaba error, se obtuvo %v", c.nombre, err)
		}
		if !c.valido {
			verificarErrorDeCampo(t, err, member.CampoRol, member.ErrAgileEnablerRepetido)
		}
	}
}

// CA-03.3 · RN7: dar de baja deja al integrante inactivo, con el resto de sus datos igual
func TestDarDeBaja_IntegranteActivo_QuedaInactivo(t *testing.T) {
	i := existente(1, "ana@mail.com", member.RolProductBuilder, true)

	baja, err := i.DarDeBaja()

	if err != nil {
		t.Fatalf("no se esperaba error, se obtuvo: %v", err)
	}
	if baja.Activo {
		t.Error("se esperaba que quede inactivo")
	}
	i.Activo = false
	if baja != i {
		t.Errorf("solo tenía que cambiar Activo: se esperaba %+v y se obtuvo %+v", i, baja)
	}
}

// RN8 · no se da de baja dos veces
func TestDarDeBaja_YaDadoDeBaja_DaError(t *testing.T) {
	i := existente(1, "ana@mail.com", member.RolProductBuilder, false)

	_, err := i.DarDeBaja()

	if !errors.Is(err, member.ErrYaDadoDeBaja) {
		t.Errorf("se esperaba %v y se obtuvo %v", member.ErrYaDadoDeBaja, err)
	}
}
