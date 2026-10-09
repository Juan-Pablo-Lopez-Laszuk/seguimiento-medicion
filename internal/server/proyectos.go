package server

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// proyectos agrupa los handlers de la épica E1 (HU-01).
type proyectos struct {
	casos *service.Proyectos
}

// listaProyectos son los datos de la página proyectos.html.
type listaProyectos struct {
	Proyectos []project.Proyecto
	Creado    bool // muestra el aviso "Proyecto creado" al volver del formulario
}

// formProyecto son los datos de la página proyecto_nuevo.html: lo que cargó el usuario (como texto,
// para volver a mostrarlo tal cual) y el error de cada campo.
type formProyecto struct {
	Nombre       string
	Descripcion  string
	FechaInicio  string
	FechaFin     string
	Errores      map[string]string
	ErrorGeneral string
}

// listar muestra todos los proyectos.
func (h proyectos) listar(w http.ResponseWriter, r *http.Request) {
	lista, err := h.casos.Listar(r.Context())
	if err != nil {
		log.Printf("listar proyectos: %v", err)
		http.Error(w, "No se pudieron cargar los proyectos. Probá de nuevo en unos minutos.", http.StatusInternalServerError)
		return
	}
	render(w, http.StatusOK, "proyectos.html", Pagina{
		Titulo: "Proyectos",
		Datos:  listaProyectos{Proyectos: lista, Creado: r.URL.Query().Get("creado") == "1"},
	})
}

// nuevo muestra el formulario vacío.
func (h proyectos) nuevo(w http.ResponseWriter, _ *http.Request) {
	render(w, http.StatusOK, "proyecto_nuevo.html", Pagina{Titulo: "Nuevo proyecto", Datos: formProyecto{}})
}

// crear recibe el formulario. Si los datos son válidos vuelve a la lista (redirección 303, así un
// F5 no crea el proyecto dos veces); si no, muestra el formulario con los errores y los datos cargados.
func (h proyectos) crear(w http.ResponseWriter, r *http.Request) {
	form := formProyecto{
		Nombre:      r.PostFormValue("nombre"),
		Descripcion: r.PostFormValue("descripcion"),
		FechaInicio: r.PostFormValue("fecha_inicio"),
		FechaFin:    r.PostFormValue("fecha_fin"),
	}
	_, err := h.casos.Crear(r.Context(), project.Datos{
		Nombre:      form.Nombre,
		Descripcion: form.Descripcion,
		FechaInicio: fechaDeFormulario(form.FechaInicio),
		FechaFin:    fechaDeFormulario(form.FechaFin),
	})

	var errs domain.ErroresValidacion
	switch {
	case err == nil:
		http.Redirect(w, r, "/proyectos?creado=1", http.StatusSeeOther)
		return
	case errors.As(err, &errs):
		form.Errores = make(map[string]string, len(errs))
		for campo, e := range errs {
			form.Errores[campo] = e.Error()
		}
		render(w, http.StatusUnprocessableEntity, "proyecto_nuevo.html", Pagina{Titulo: "Nuevo proyecto", Datos: form})
	default:
		log.Printf("crear proyecto: %v", err)
		form.ErrorGeneral = "No se pudo guardar el proyecto. Probá de nuevo en unos minutos."
		render(w, http.StatusInternalServerError, "proyecto_nuevo.html", Pagina{Titulo: "Nuevo proyecto", Datos: form})
	}
}

// fechaDeFormulario convierte "AAAA-MM-DD" (lo que envía un <input type="date">) en fecha.
// Vacío o inválido devuelve la fecha cero, que el dominio informa como fecha faltante.
func fechaDeFormulario(texto string) time.Time {
	f, err := time.Parse(time.DateOnly, texto)
	if err != nil {
		return time.Time{}
	}
	return f
}
