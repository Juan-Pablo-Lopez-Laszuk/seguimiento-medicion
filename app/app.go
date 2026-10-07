// Package app es la puerta de entrada pública a la aplicación.
//
// La función de Vercel (api/index.go) se compila fuera del módulo, y Go no deja importar paquetes
// "internal" desde afuera. Este paquete sí está dentro del módulo, así que puede usar internal/server
// y exponer el router para Vercel.
package app

import (
	"net/http"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// NewHandler arma la aplicación completa: repositorios, casos de uso y router.
//
// Por ahora los datos se guardan en memoria; cuando esté la base (TEC-03) se cambia el repositorio
// por el de internal/store/postgres sin tocar nada más.
func NewHandler() http.Handler {
	return server.NewRouter(server.Dependencias{
		Proyectos: service.NuevoProyectos(memory.NuevoProyectos(), time.Now),
	})
}
