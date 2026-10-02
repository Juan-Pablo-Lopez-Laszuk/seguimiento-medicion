// Package server arma el servidor web: el router con todas las rutas, los handlers y el renderizado
// de las páginas. Lo usan tanto el servidor local (cmd/server) como la función de Vercel (api/index.go).
package server

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// NewRouter arma el router con todas las rutas de la aplicación.
func NewRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/health", health)
	r.Get("/", inicio)
	return r
}

// inicio muestra la página principal.
func inicio(w http.ResponseWriter, _ *http.Request) {
	render(w, "inicio.html", Pagina{Titulo: "Inicio"})
}

// health indica que la aplicación está viva. Lo usan Vercel y el CI para verificar el deploy.
func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
