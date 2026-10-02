// Package server arma el servidor web: el router con todas las rutas, los handlers y el renderizado
// de las páginas. Lo usan tanto el servidor local (cmd/server) como la función de Vercel (api/index.go).
package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// NewRouter arma el router con todas las rutas de la aplicación.
func NewRouter() http.Handler {
	r := chi.NewRouter()
	return r
}
