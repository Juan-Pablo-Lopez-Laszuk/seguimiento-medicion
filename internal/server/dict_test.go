package server

import "testing"

// dict es la ayuda de las plantillas para pasarle varios datos a un componente (TEC-07).
func TestDict(t *testing.T) {
	m, err := dict("ID", "confirmar-baja-1", "Boton", "Dar de baja")
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}
	if m["ID"] != "confirmar-baja-1" || m["Boton"] != "Dar de baja" {
		t.Errorf("dict = %v", m)
	}

	if _, err := dict("ID"); err == nil {
		t.Error("una clave sin valor debería dar error")
	}
	if _, err := dict(1, "uno"); err == nil {
		t.Error("una clave que no es texto debería dar error")
	}
}
