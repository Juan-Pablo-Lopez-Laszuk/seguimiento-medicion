package server_test

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// Los tests de HU-02 usan el proyecto de ejemplo (ID 1, del 05/10/2026 al 01/11/2026) y, salvo que digan otra cosa,
// el reloj de nuevoRouter: hoy es el 13/10/2026.

// proyectoFinalizado arma la aplicación con "hoy" en el 02/11/2026 y el proyecto de ejemplo con un sprint: como ya
// pasó su fecha de fin, está Finalizado.
func proyectoFinalizado(t *testing.T) http.Handler {
	t.Helper()
	h := nuevoRouterConReloj(func() time.Time { return time.Date(2026, 11, 2, 18, 0, 0, 0, time.UTC) })
	if rec := post(h, "/proyectos", formularioValido()); rec.Code != http.StatusSeeOther {
		t.Fatalf("no se pudo crear el proyecto de ejemplo: código %d", rec.Code)
	}
	if rec := post(h, "/proyectos/1/sprints", sprintValido()); rec.Code != http.StatusSeeOther {
		t.Fatalf("no se pudo crear el sprint: código %d", rec.Code)
	}
	return h
}

// Sección 3 · la vista del proyecto tiene el botón para modificarlo
func TestVistaDelProyecto_TieneElBotonModificar(t *testing.T) {
	verificarContiene(t, get(conProyecto(t), "/proyectos/1").Body.String(), `href="/proyectos/1/editar"`)
}

// Sección 3 · el formulario aparece con los datos actuales ya cargados
func TestFormularioModificar_CargaLosDatosActuales(t *testing.T) {
	rec := get(conProyecto(t), "/proyectos/1/editar")

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusOK)
	}
	verificarContiene(t, rec.Body.String(),
		"<title>Modificar proyecto · Software Metrics</title>",
		`action="/proyectos/1/editar"`,
		`value="Software Metrics"`, "TPI de ICS", `value="2026-10-05"`, `value="2026-11-01"`,
	)
}

// CA-02.1 · con datos válidos vuelve a la vista del proyecto con el mensaje y los datos nuevos
func TestModificar_DatosValidos_VuelveALaVistaConElMensaje(t *testing.T) {
	h := conProyecto(t)
	campos := formularioValido()
	campos.Set("descripcion", "TPI de ICS 2026")

	rec := post(h, "/proyectos/1/editar", campos)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("código = %d, se esperaba %d (redirección)", rec.Code, http.StatusSeeOther)
	}
	destino := rec.Header().Get("Location")
	if destino != "/proyectos/1?modificado=1" {
		t.Fatalf("redirige a %q, se esperaba /proyectos/1?modificado=1", destino)
	}
	verificarContiene(t, get(h, destino).Body.String(), "Proyecto modificado", "TPI de ICS 2026")
}

// CA-02.1 y CA-02.2 · los errores aparecen en su campo y se conservan los datos cargados
func TestModificar_Errores_MuestraElFormularioConLosErrores(t *testing.T) {
	h := conSprint(t, "2026-10-05", "2026-10-11")
	campos := url.Values{
		"nombre":       {"ab"},
		"descripcion":  {"Otra descripción"},
		"fecha_inicio": {"2026-10-05"},
		"fecha_fin":    {"2026-10-10"},
	}

	rec := post(h, "/proyectos/1/editar", campos)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("código = %d, se esperaba %d", rec.Code, http.StatusUnprocessableEntity)
	}
	verificarContiene(t, rec.Body.String(),
		"el nombre debe tener entre 3 y 100 caracteres",
		"la fecha de finalización no puede ser anterior al fin del último sprint (el Sprint 1 termina el 11/10/2026)",
		`value="ab"`, "Otra descripción", `value="2026-10-10"`,
	)
	verificarContiene(t, get(h, "/proyectos/1").Body.String(), "TPI de ICS") // no se guardó nada
}

// CA-02.3 · un proyecto Finalizado no muestra el formulario, rechaza el pedido y no tiene el botón
func TestModificar_ProyectoFinalizado(t *testing.T) {
	h := proyectoFinalizado(t)
	campos := formularioValido()
	campos.Set("descripcion", "Cambio tardío")

	formulario := get(h, "/proyectos/1/editar")
	envio := post(h, "/proyectos/1/editar", campos)

	for nombre, rec := range map[string]int{"formulario": formulario.Code, "envío": envio.Code} {
		if rec != http.StatusConflict {
			t.Errorf("%s: código = %d, se esperaba %d", nombre, rec, http.StatusConflict)
		}
	}
	verificarContiene(t, formulario.Body.String(), "un proyecto Finalizado no se puede modificar")
	if strings.Contains(formulario.Body.String(), "<form") {
		t.Error("a un proyecto Finalizado no se le muestra el formulario")
	}
	vista := get(h, "/proyectos/1").Body.String()
	if strings.Contains(vista, "Cambio tardío") {
		t.Error("un proyecto Finalizado no tenía que guardarse")
	}
	if strings.Contains(vista, `href="/proyectos/1/editar"`) {
		t.Error("la vista de un proyecto Finalizado no tiene que tener el botón Modificar")
	}
}

// Sección 7 · un proyecto que no existe (o un ID que no es un número) da 404
func TestModificar_ProyectoInexistente(t *testing.T) {
	h := conProyecto(t)
	for _, ruta := range []string{"/proyectos/999/editar", "/proyectos/abc/editar"} {
		if rec := get(h, ruta); rec.Code != http.StatusNotFound {
			t.Errorf("GET %s: código = %d, se esperaba %d", ruta, rec.Code, http.StatusNotFound)
		}
		if rec := post(h, ruta, formularioValido()); rec.Code != http.StatusNotFound {
			t.Errorf("POST %s: código = %d, se esperaba %d", ruta, rec.Code, http.StatusNotFound)
		}
	}
}

// Sección 7 · si falla la base se muestra un mensaje general, sin detalles internos
func TestModificar_ErrorDeLaBase(t *testing.T) {
	proyectos := memory.NuevoProyectos()
	if _, err := proyectos.Guardar(context.Background(), project.Proyecto{Nombre: "Software Metrics"}); err != nil {
		t.Fatalf("no se pudo guardar el proyecto: %v", err)
	}
	h := server.NewRouter(server.Dependencias{
		Edicion: service.NuevoEdicionProyectos(proyectos, sprintsCaidos{}, relojDeLaVista),
	})

	formulario := get(h, "/proyectos/1/editar")
	envio := post(h, "/proyectos/1/editar", formularioValido())

	for nombre, rec := range map[string]string{"formulario": formulario.Body.String(), "envío": envio.Body.String()} {
		if strings.Contains(rec, "detalle interno") {
			t.Errorf("%s: la página muestra el error interno", nombre)
		}
	}
	if formulario.Code != http.StatusInternalServerError || envio.Code != http.StatusInternalServerError {
		t.Errorf("códigos = %d y %d, se esperaba %d", formulario.Code, envio.Code, http.StatusInternalServerError)
	}
	verificarContiene(t, envio.Body.String(), "No se pudo guardar el proyecto", `value="Software Metrics"`)
}
