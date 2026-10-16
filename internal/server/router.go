// Package server arma el servidor web: el router con todas las rutas, los handlers y el renderizado
// de las páginas. Lo usan tanto el servidor local (cmd/server) como la función de Vercel (api/index.go).
package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// Dependencias son los casos de uso que usan los handlers. Se arman afuera (en app) para que los
// tests puedan pasar repositorios en memoria.
type Dependencias struct {
	Proyectos   *service.Proyectos
	Sprints     *service.Sprints
	Integrantes *service.Integrantes
	Historias   *service.Historias
	Criterios   *service.Criterios
	// Asignaciones arma el Sprint Backlog: asignar y quitar historias de un sprint (HU-12).
	Asignaciones *service.Asignaciones
	// CambioDeEstado cambia el estado de una historia (HU-08); Estado es el del proyecto (HU-04).
	CambioDeEstado *service.Estados
	Tareas         *service.Tareas
	Estado         *service.EstadoProyectos
	Edicion        *service.EdicionProyectos
	// Base es la base de datos que chequea /health; nil si la app guarda en memoria (TEC-04).
	Base VerificadorBase
}

// NewRouter arma el router con todas las rutas de la aplicación.
func NewRouter(dep Dependencias) http.Handler {
	r := chi.NewRouter()
	r.Get("/health", health(dep.Base))
	r.Get("/", inicio)

	p := proyectos{casos: dep.Proyectos, estado: dep.Estado}
	r.Get("/proyectos", p.listar)
	r.Get("/proyectos/nuevo", p.nuevo)
	r.Post("/proyectos", p.crear)
	r.Get("/proyectos/{id}", p.ver)

	ed := edicion{casos: dep.Edicion}
	r.Get("/proyectos/{id}/editar", ed.formulario)
	r.Post("/proyectos/{id}/editar", ed.modificar)

	s := sprints{casos: dep.Sprints}
	r.Get("/proyectos/{id}/sprints", s.listar)
	r.Get("/proyectos/{id}/sprints/nuevo", s.nuevo)
	r.Post("/proyectos/{id}/sprints", s.crear)

	pl := planificacion{casos: dep.Asignaciones}
	r.Get("/proyectos/{id}/sprints/{sprint}", pl.ver)
	r.Post("/proyectos/{id}/sprints/{sprint}/historias", pl.asignar)
	r.Post("/proyectos/{id}/sprints/{sprint}/historias/{historia}/quitar", pl.accion((*service.Asignaciones).Quitar, "quitada"))
	r.Post("/proyectos/{id}/sprints/{sprint}/historias/{historia}/empezar", pl.accion((*service.Asignaciones).Empezar, "empezada"))
	r.Post("/proyectos/{id}/sprints/{sprint}/historias/{historia}/completar", pl.accion((*service.Asignaciones).Completar, "completada"))

	in := integrantes{casos: dep.Integrantes}
	r.Get("/proyectos/{id}/integrantes", in.listar)
	r.Get("/proyectos/{id}/integrantes/nuevo", in.nuevo)
	r.Post("/proyectos/{id}/integrantes", in.registrar)
	r.Post("/proyectos/{id}/integrantes/{integranteID}/baja", in.darDeBaja)

	hs := historias{casos: dep.Historias, criterios: dep.Criterios}
	r.Get("/proyectos/{id}/backlog", hs.listar)
	r.Get("/proyectos/{id}/backlog/nuevo", hs.nueva)
	r.Post("/proyectos/{id}/backlog", hs.crear)

	cr := criterios{casos: dep.Criterios}
	r.Get("/proyectos/{id}/backlog/{numero}", cr.ver)
	r.Post("/proyectos/{id}/backlog/{numero}/criterios", cr.agregar)
	r.Post("/proyectos/{id}/backlog/{numero}/criterios/{criterioID}/cumplido", cr.marcar)

	es := estados{casos: dep.CambioDeEstado, vista: dep.Criterios}
	r.Post("/proyectos/{id}/backlog/{numero}/estado", es.cambiar)

	ta := tareas{casos: dep.Tareas}
	r.Get("/proyectos/{id}/backlog/{numero}/tareas", ta.ver)
	r.Post("/proyectos/{id}/backlog/{numero}/tareas", ta.agregar)
	r.Post("/proyectos/{id}/backlog/{numero}/tareas/{tareaID}/horas", ta.cambiarHoras)
	r.Post("/proyectos/{id}/backlog/{numero}/tareas/{tareaID}/completada", ta.marcar)
	return r
}

// inicio muestra la página principal.
func inicio(w http.ResponseWriter, _ *http.Request) {
	render(w, http.StatusOK, "inicio.html", Pagina{Titulo: "Inicio"})
}
