package backlog_test

import (
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
)

// historiaCompleta devuelve una historia ya trabajada: estimada, En progreso y con descripción, para probar que editar
// no le toca nada más que título, descripción y prioridad.
func historiaCompleta() backlog.Historia {
	puntos := 5
	h := historiaEn(backlog.EstadoEnProgreso)
	h.Descripcion = "Para no tener que recordar otra contraseña."
	h.StoryPoints = &puntos
	return h
}

// cambiosValidos devuelve cambios que cumplen todas las reglas; cada test cambia solo lo que quiere probar.
func cambiosValidos() backlog.Cambios {
	return backlog.Cambios{
		Titulo:      "Iniciar sesión con Google o GitHub",
		Descripcion: "Con OAuth.",
		Prioridad:   backlog.PrioridadBaja,
	}
}

// CA-06.1 · editar cambia título, descripción y prioridad, y la historia sigue siendo la misma (mismo número e ID)
func TestEditar_CambiaLosDatosYConservaElIdentificador(t *testing.T) {
	original := historiaCompleta()

	editada, err := original.Editar(cambiosValidos())
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if editada.Titulo != "Iniciar sesión con Google o GitHub" || editada.Descripcion != "Con OAuth." || editada.Prioridad != backlog.PrioridadBaja {
		t.Errorf("la historia quedó con %q / %q / %q", editada.Titulo, editada.Descripcion, editada.Prioridad)
	}
	if editada.ID != original.ID || editada.ProyectoID != original.ProyectoID || editada.Numero != original.Numero {
		t.Errorf("el identificador cambió: era %d/%d/HU-%d y quedó %d/%d/HU-%d",
			original.ID, original.ProyectoID, original.Numero, editada.ID, editada.ProyectoID, editada.Numero)
	}
	if editada.Identificador() != "HU-1" {
		t.Errorf("Identificador() = %q, se esperaba HU-1", editada.Identificador())
	}
}

// RN2 · editar no toca el estado ni los story points, y la historia original no cambia
func TestEditar_NoTocaElEstadoNiLosStoryPoints(t *testing.T) {
	original := historiaCompleta()

	editada, err := original.Editar(cambiosValidos())
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if editada.Estado != backlog.EstadoEnProgreso {
		t.Errorf("el estado cambió a %q", editada.Estado)
	}
	if editada.StoryPoints == nil || *editada.StoryPoints != 5 {
		t.Errorf("los story points cambiaron a %v", editada.StoryPoints)
	}
	if original.Titulo != "Iniciar sesión con Google" || original.Prioridad != backlog.PrioridadAlta {
		t.Errorf("la historia original cambió: %q / %q", original.Titulo, original.Prioridad)
	}
}

// RN4 · guardar sin haber cambiado nada vale, y la historia queda igual
func TestEditar_SinCambios_NoEsUnError(t *testing.T) {
	original := historiaCompleta()
	sinCambios := backlog.Cambios{Titulo: original.Titulo, Descripcion: original.Descripcion, Prioridad: original.Prioridad}

	editada, err := original.Editar(sinCambios)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if editada.Titulo != original.Titulo || editada.Descripcion != original.Descripcion || editada.Prioridad != original.Prioridad {
		t.Errorf("la historia cambió: %+v", editada)
	}
}

// RN1 · el título y la descripción se guardan sin los espacios del inicio y del final
func TestEditar_QuitaLosEspaciosDeLosExtremos(t *testing.T) {
	c := cambiosValidos()
	c.Titulo, c.Descripcion = "  Login con Google  ", "\t Con OAuth \n"

	editada, err := historiaCompleta().Editar(c)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if editada.Titulo != "Login con Google" || editada.Descripcion != "Con OAuth" {
		t.Errorf("la historia quedó con %q / %q", editada.Titulo, editada.Descripcion)
	}
}

// RN1 · la descripción es opcional: vacía queda como texto vacío
func TestEditar_DescripcionVacia_QuedaVacia(t *testing.T) {
	c := cambiosValidos()
	c.Descripcion = "   "

	editada, err := historiaCompleta().Editar(c)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if editada.Descripcion != "" {
		t.Errorf("la descripción quedó %q, se esperaba vacía", editada.Descripcion)
	}
}

// RN3 · se puede editar en cualquier estado, también una historia Hecha
func TestEditar_EnCualquierEstado_Vale(t *testing.T) {
	for _, estado := range backlog.Estados {
		editada, err := historiaEn(estado).Editar(cambiosValidos())
		if err != nil {
			t.Errorf("estado %q: no se esperaba error: %v", estado, err)
			continue
		}
		if editada.Estado != estado {
			t.Errorf("estado %q: el estado cambió a %q", estado, editada.Estado)
		}
	}
}
