// Package server arma el servidor web: el router con todas las rutas, los handlers y el renderizado
// de las páginas. Lo usan tanto el servidor local (cmd/server) como la función de Vercel (api/index.go).
package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// Dependencias son los casos de uso que usan los handlers. Se arman afuera (en app) para que los
// tests puedan pasar repositorios en memoria.
type Dependencias struct {
	Proyectos *service.Proyectos
	Sprints   *service.Sprints
}

// NewRouter arma el router con todas las rutas de la aplicación.
func NewRouter(dep Dependencias) http.Handler {
	r := chi.NewRouter()
	r.Get("/health", health)
	r.Get("/", inicio)

	p := proyectos{casos: dep.Proyectos}
	r.Get("/proyectos", p.listar)
	r.Get("/proyectos/nuevo", p.nuevo)
	r.Post("/proyectos", p.crear)

	s := sprints{casos: dep.Sprints}
	r.Get("/proyectos/{id}/sprints", s.listar)
	r.Get("/proyectos/{id}/sprints/nuevo", s.nuevo)
	r.Post("/proyectos/{id}/sprints", s.crear)
	return r
}

// inicio muestra la página principal.
func inicio(w http.ResponseWriter, _ *http.Request) {
	render(w, http.StatusOK, "inicio.html", Pagina{Titulo: "Inicio"})
}

// health indica que la aplicación está viva. Lo usan Vercel y el CI para verificar el deploy.
func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
