package member_test

import (
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
)

const proyectoID int64 = 7

// datosValidos devuelve datos que cumplen todas las reglas; cada test cambia solo lo que quiere probar.
func datosValidos() member.Datos {
	return member.Datos{Nombre: "Juan Pablo", Email: "jp@mail.com", Rol: member.RolAgileEnabler}
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
