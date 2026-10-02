// Package handler es la función que Vercel ejecuta en cada pedido. vercel.json redirige todas las
// rutas a esta función, y el router de la aplicación (el mismo que usa el servidor local) decide qué responder.
package handler

import (
	"net/http"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
)

// router se arma una sola vez y se reutiliza entre pedidos mientras la función siga activa.
var router = server.NewRouter()

// Handler es el punto de entrada que invoca Vercel.
func Handler(w http.ResponseWriter, r *http.Request) {
	router.ServeHTTP(w, r)
}
