package backlog_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
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

// RN1 · el título es obligatorio, también si son solo espacios
func TestEditar_TituloVacio_EsUnErrorDelCampo(t *testing.T) {
	for _, titulo := range []string{"", "   ", "\t\n"} {
		c := cambiosValidos()
		c.Titulo = titulo

		_, err := historiaCompleta().Editar(c)

		verificarErrorDeCampo(t, err, backlog.CampoTitulo, backlog.ErrTituloObligatorio)
	}
}

// RN1 · el título llega a 120 caracteres (se cuentan caracteres, no bytes); con uno más es un error
func TestEditar_LargoDelTitulo(t *testing.T) {
	casos := []struct {
		nombre   string
		titulo   string
		esperado error
	}{
		{"120 caracteres", strings.Repeat("a", 120), nil},
		{"121 caracteres", strings.Repeat("a", 121), backlog.ErrTituloLargo},
		{"120 caracteres con tildes", strings.Repeat("ñ", 120), nil},
		{"121 caracteres con tildes", strings.Repeat("ñ", 121), backlog.ErrTituloLargo},
	}
	for _, c := range casos {
		cambios := cambiosValidos()
		cambios.Titulo = c.titulo

		_, err := historiaCompleta().Editar(cambios)

		if c.esperado == nil {
			if err != nil {
				t.Errorf("%s: no se esperaba error: %v", c.nombre, err)
			}
			continue
		}
		verificarErrorDeCampo(t, err, backlog.CampoTitulo, c.esperado)
	}
}

// RN1 · la descripción llega a 2000 caracteres; con uno más es un error
func TestEditar_LargoDeLaDescripcion(t *testing.T) {
	casos := []struct {
		nombre      string
		descripcion string
		esperado    error
	}{
		{"2000 caracteres", strings.Repeat("a", 2000), nil},
		{"2001 caracteres", strings.Repeat("a", 2001), backlog.ErrDescripcionLarga},
	}
	for _, c := range casos {
		cambios := cambiosValidos()
		cambios.Descripcion = c.descripcion

		_, err := historiaCompleta().Editar(cambios)

		if c.esperado == nil {
			if err != nil {
				t.Errorf("%s: no se esperaba error: %v", c.nombre, err)
			}
			continue
		}
		verificarErrorDeCampo(t, err, backlog.CampoDescripcion, c.esperado)
	}
}

// RN1 · la prioridad tiene que ser Alta, Media o Baja
func TestEditar_PrioridadInvalida_EsUnErrorDelCampo(t *testing.T) {
	for _, prioridad := range []backlog.Prioridad{"", "Urgente", "alta"} {
		c := cambiosValidos()
		c.Prioridad = prioridad

		_, err := historiaCompleta().Editar(c)

		verificarErrorDeCampo(t, err, backlog.CampoPrioridad, backlog.ErrPrioridadInvalida)
	}
}

// RN1 · si hay varios datos inválidos se informan todos juntos
func TestEditar_VariosErrores_SeInformanJuntos(t *testing.T) {
	c := backlog.Cambios{Titulo: "", Descripcion: strings.Repeat("a", 2001), Prioridad: "Urgente"}

	_, err := historiaCompleta().Editar(c)

	verificarErrorDeCampo(t, err, backlog.CampoTitulo, backlog.ErrTituloObligatorio)
	verificarErrorDeCampo(t, err, backlog.CampoDescripcion, backlog.ErrDescripcionLarga)
	verificarErrorDeCampo(t, err, backlog.CampoPrioridad, backlog.ErrPrioridadInvalida)
	var errs domain.ErroresValidacion
	if errors.As(err, &errs) && len(errs) != 3 {
		t.Errorf("se esperaban 3 errores y hay %d: %v", len(errs), errs)
	}
}

// RN1 · con datos inválidos no se devuelve ninguna historia, y la original queda como estaba
func TestEditar_DatosInvalidos_NoCambianNada(t *testing.T) {
	original := historiaCompleta()
	c := cambiosValidos()
	c.Titulo = ""

	editada, err := original.Editar(c)

	if err == nil {
		t.Fatal("se esperaba un error")
	}
	if editada.Numero != 0 || editada.Titulo != "" {
		t.Errorf("con un error se esperaba la historia vacía y se obtuvo %+v", editada)
	}
	if original.Titulo != "Iniciar sesión con Google" || original.Prioridad != backlog.PrioridadAlta {
		t.Errorf("la historia original cambió: %q / %q", original.Titulo, original.Prioridad)
	}
}
