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
			service.ListadorProyectos
			service.RepositorioEdicionProyectos
		}
		sprints     service.RepositorioSprints
		integrantes service.RepositorioIntegrantes
		historias   interface {
			service.RepositorioHistorias
			service.RepositorioEstados
			service.RepositorioAsignaciones
			service.RepositorioEstimaciones
			service.RepositorioEdicionHistorias
		}
		criterios service.RepositorioCriterios
		tareas    service.RepositorioTareas
		registros service.RepositorioRegistros
		defectos  service.RepositorioDefectos
		cierres   service.RepositorioCierres
		base      server.VerificadorBase // queda en nil en modo memoria: /health informa "memoria"
	)
	if url == "" {
		proyectos, integrantes = memory.NuevoProyectos(), memory.NuevoIntegrantes()
		// Eliminar una historia borra también sus criterios y sus tareas (HU-06): en memoria, el repositorio de
		// historias conoce a los otros dos.
		criteriosEnMemoria, tareasEnMemoria := memory.NuevoCriterios(), memory.NuevoTareas()
		historiasEnMemoria, sprintsEnMemoria := memory.NuevoHistorias(criteriosEnMemoria, tareasEnMemoria), memory.NuevoSprints()
		historias, criterios, tareas, sprints = historiasEnMemoria, criteriosEnMemoria, tareasEnMemoria, sprintsEnMemoria
		// El cierre de un sprint (HU-14) cambia sprints e historias juntos, y le avisa al de historias qué sprints se
		// cerraron (para HU-06).
		cierres = memory.NuevoCierres(sprintsEnMemoria, historiasEnMemoria)
		registros = memory.NuevoRegistros()
		defectos = memory.NuevoDefectos()
	} else {
		pool, err := postgres.Conectar(context.Background(), url)
		if err != nil {
			return nil, err
		}
		proyectos, sprints, integrantes = postgres.NuevoProyectos(pool), postgres.NuevoSprints(pool), postgres.NuevoIntegrantes(pool)
		historias, criterios, tareas = postgres.NuevoHistorias(pool), postgres.NuevoCriterios(pool), postgres.NuevoTareas(pool)
		registros = postgres.NuevoRegistros(pool)
		defectos = postgres.NuevoDefectos(pool)
		cierres = postgres.NuevoCierres(pool)
		base = pool
	}
	estado := service.NuevoEstadoProyectos(proyectos, sprints, historias, time.Now)
	horas := service.NuevoHoras(proyectos, sprints, historias, tareas, registros)
	casosDefectos := service.NuevoDefectos(proyectos, sprints, historias, defectos)
	casosDeCierre := service.NuevoCierres(proyectos, sprints, historias, cierres, time.Now)
	return server.NewRouter(server.Dependencias{
		Proyectos:      service.NuevoProyectos(proyectos, time.Now),
		Sprints:        service.NuevoSprints(proyectos, sprints),
		Integrantes:    service.NuevoIntegrantes(proyectos, integrantes),
		Historias:      service.NuevoHistorias(proyectos, historias),
		Criterios:      service.NuevoCriterios(proyectos, historias, criterios),
		Asignaciones:   service.NuevoAsignaciones(proyectos, sprints, historias, service.NuevoCriterios(proyectos, historias, criterios), time.Now),
		CambioDeEstado: service.NuevoEstados(proyectos, historias, criterios),
		Estimaciones:   service.NuevoEstimaciones(proyectos, historias),
		Tareas:         service.NuevoTareas(proyectos, historias, integrantes, tareas),
		Esfuerzo:       service.NuevoEsfuerzo(proyectos, historias, integrantes, tareas, registros, time.Now),
		Defectos:       casosDefectos,
		Horas:          horas,
		Panel:          service.NuevoPanel(estado, sprints, horas, casosDefectos),
		Reportes:       service.NuevoReportes(proyectos, sprints, historias, casosDeCierre, horas, defectos),
		Estado:         estado,
		Edicion:        service.NuevoEdicionProyectos(proyectos, sprints, time.Now),
		Cierres:        casosDeCierre,
		Base:           base,

		// Editar y eliminar una historia (HU-06).
		EdicionHistorias: service.NuevoEdicionHistorias(proyectos, historias, criterios, tareas),
	}), nil
}
