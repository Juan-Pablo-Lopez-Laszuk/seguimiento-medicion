// Package handler es la función que Vercel ejecuta cuando alguien entra a la app.
// Por ahora solo responde "Hola mundo": sirve para probar que el deploy funciona.
// En el Sprint 1 (TEC-04) acá se conecta el router de la app.
package handler

import (
	"fmt"
	"net/http"
)

// Handler recibe cada pedido del navegador y devuelve una página simple.
func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, "<h1>Hola mundo</h1><p>Software Metrics &amp; Estimation · TPI ICSW 2026 · UTN FRSR</p>")
}
