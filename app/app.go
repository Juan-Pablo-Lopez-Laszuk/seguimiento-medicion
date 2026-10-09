// Package app es la puerta de entrada pública a la aplicación.
//
// La función de Vercel (api/index.go) se compila fuera del módulo, y Go no deja importar paquetes
// "internal" desde afuera. Este paquete sí está dentro del módulo, así que puede usar internal/server
// y exponer el router para Vercel.
package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/postgres"
)

// NewHandler arma la aplicación completa: repositorios, casos de uso y router. Guarda en la base de
// DATABASE_URL; si esa variable no está, guarda en memoria. Si la URL no sirve, corta el arranque con un
// mensaje claro en lugar de seguir con datos que se perderían.
func NewHandler() http.Handler {
	h, err := NewHandlerConBase(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("no se pudo preparar la base de datos: %v", err)
	}
	return h
}

// NewHandlerConBase arma la aplicación con los repositorios de PostgreSQL si url no está vacía, o con los
// de memoria si lo está. Abrir el pool no conecta todavía: la primera conexión se hace en el primer pedido.
func NewHandlerConBase(url string) (http.Handler, error) {
	var (
		proyectos interface {
			service.RepositorioProyectos
			service.BuscadorProyectos
		}
		sprints     service.RepositorioSprints
		integrantes service.RepositorioIntegrantes
	)
	if url == "" {
		proyectos, sprints, integrantes = memory.NuevoProyectos(), memory.NuevoSprints(), memory.NuevoIntegrantes()
	} else {
		pool, err := postgres.Conectar(context.Background(), url)
		if err != nil {
			return nil, err
		}
		proyectos, sprints, integrantes = postgres.NuevoProyectos(pool), postgres.NuevoSprints(pool), postgres.NuevoIntegrantes(pool)
	}
	return server.NewRouter(server.Dependencias{
		Proyectos:   service.NuevoProyectos(proyectos, time.Now),
		Sprints:     service.NuevoSprints(proyectos, sprints),
		Integrantes: service.NuevoIntegrantes(proyectos, integrantes),
	}), nil
}
