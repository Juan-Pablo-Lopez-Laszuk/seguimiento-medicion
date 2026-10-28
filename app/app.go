// Package app es la puerta de entrada pública a la aplicación.
//
// La función de Vercel (api/index.go) se compila fuera del módulo, y Go no deja importar paquetes
// "internal" desde afuera. Este paquete sí está dentro del módulo, así que puede usar internal/server
// y exponer el router para Vercel.
package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/auth"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/postgres"
)

// NewHandler arma la aplicación completa: repositorios, casos de uso y router. Guarda en la base de
// DATABASE_URL; si esa variable no está, guarda en memoria. Pide login si están SUPABASE_URL y SUPABASE_ANON_KEY
// (TEC-05). Si la base o la configuración del login no sirven, corta el arranque con un mensaje claro en lugar de
// seguir con datos que se perderían o con un login a medio configurar.
func NewHandler() http.Handler {
	h, err := NewHandlerDesdeEntorno(os.Getenv)
	if err != nil {
		log.Fatalf("no se pudo preparar la aplicación: %v", err)
	}
	return h
}

// NewHandlerDesdeEntorno arma la aplicación con las variables DATABASE_URL, SUPABASE_URL y SUPABASE_ANON_KEY que devuelve
// getenv. Sin las dos de Supabase no hay login; con una sola devuelve un error que dice cuál falta (TEC-05, RN1).
func NewHandlerDesdeEntorno(getenv func(string) string) (http.Handler, error) {
	cfg, loginActivo, err := auth.ConfigDesdeEntorno(getenv)
	if err != nil {
		return nil, fmt.Errorf("configuración del login: %w", err)
	}
	var opciones []Opcion
	if loginActivo {
		verificador, err := auth.NuevoVerificador(context.Background(), cfg)
		if err != nil {
			return nil, fmt.Errorf("preparar el login: %w", err)
		}
		opciones = append(opciones, ConLogin(verificador, auth.NuevoSupabase(cfg)))
	}
	return NewHandlerConBase(getenv("DATABASE_URL"), opciones...)
}

// Opcion cambia algo de cómo NewHandlerConBase arma la aplicación.
type Opcion func(*opciones)

type opciones struct {
	verificador auth.Verificador
	cuentas     auth.ServicioDeCuentas
}

// ConLogin activa el login (TEC-05): todas las rutas piden una sesión que el verificador acepte, y las cuentas se crean y se
// abren con el servicio de cuentas.
func ConLogin(verificador auth.Verificador, cuentas auth.ServicioDeCuentas) Opcion {
	return func(o *opciones) { o.verificador, o.cuentas = verificador, cuentas }
}

// NewHandlerConBase arma la aplicación con los repositorios de PostgreSQL si url no está vacía, o con los
// de memoria si lo está. Abrir el pool no conecta todavía: la primera conexión se hace en el primer pedido.
func NewHandlerConBase(url string, opcionesDeArmado ...Opcion) (http.Handler, error) {
	var o opciones
	for _, opcion := range opcionesDeArmado {
		opcion(&o)
	}
	var (
		proyectos interface {
			service.RepositorioProyectos
			service.ListadorProyectos
			service.RepositorioEdicionProyectos
		}
		sprints     service.RepositorioSprints
		integrantes interface {
			service.RepositorioIntegrantes
			service.RepositorioVinculos // quién es cada usuario en cada proyecto (TEC-05)
		}
		historias interface {
			service.RepositorioHistorias
			service.RepositorioEstados
			service.RepositorioAsignaciones
			service.RepositorioEstimaciones
			service.ListadorEstimaciones // el historial de estimaciones de la página de la historia (HU-16b)
			service.RepositorioEdicionHistorias
		}
		criterios service.RepositorioCriterios
		// Planning Poker: las sesiones (HU-17) y los votos (HU-18) van en el mismo repositorio.
		sesiones interface {
			service.RepositorioSesionesPoker
			service.RepositorioVotosPoker
		}
		tareas    service.RepositorioTareas
		registros service.RepositorioRegistros
		defectos  service.RepositorioDefectos
		cierres   service.RepositorioCierres
		base      server.VerificadorBase // queda en nil en modo memoria: /health informa "memoria"
	)
	if url == "" {
		proyectos, integrantes = memory.NuevoProyectos(), memory.NuevoIntegrantes()
		// Eliminar una historia borra también sus criterios y sus tareas (HU-06): en memoria, el repositorio de
		// historias conoce a los otros dos. Y no se elimina si tiene horas registradas, defectos o sesiones de Planning
		// Poker (RN10), como hacen las claves foráneas de PostgreSQL: por eso también conoce a esos tres.
		criteriosEnMemoria, tareasEnMemoria := memory.NuevoCriterios(), memory.NuevoTareas()
		registrosEnMemoria, defectosEnMemoria, pokerEnMemoria := memory.NuevoRegistros(), memory.NuevoDefectos(), memory.NuevoPoker()
		historiasEnMemoria := memory.NuevoHistorias(criteriosEnMemoria, tareasEnMemoria).ConRegistrosAsociados(registrosEnMemoria, defectosEnMemoria, pokerEnMemoria)
		sprintsEnMemoria := memory.NuevoSprints()
		historias, criterios, tareas, sprints = historiasEnMemoria, criteriosEnMemoria, tareasEnMemoria, sprintsEnMemoria
		// El cierre de un sprint (HU-14) cambia sprints e historias juntos, y le avisa al de historias qué sprints se
		// cerraron (para HU-06).
		cierres = memory.NuevoCierres(sprintsEnMemoria, historiasEnMemoria)
		registros, defectos, sesiones = registrosEnMemoria, defectosEnMemoria, pokerEnMemoria
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
		sesiones = postgres.NuevoPoker(pool)
		base = pool
	}
	estado := service.NuevoEstadoProyectos(proyectos, sprints, historias, time.Now)
	horas := service.NuevoHoras(proyectos, sprints, historias, tareas, registros)
	casosDefectos := service.NuevoDefectos(proyectos, sprints, historias, defectos)
	casosDeCierre := service.NuevoCierres(proyectos, sprints, historias, cierres, time.Now)
	panel := service.NuevoPanel(estado, sprints, horas, casosDefectos)
	dep := server.Dependencias{
		Proyectos:      service.NuevoProyectos(proyectos, time.Now),
		Sprints:        service.NuevoSprints(proyectos, sprints),
		Integrantes:    service.NuevoIntegrantes(proyectos, integrantes),
		Historias:      service.NuevoHistorias(proyectos, historias),
		Criterios:      service.NuevoCriterios(proyectos, historias, criterios).ConHistorial(historias),
		Asignaciones:   service.NuevoAsignaciones(proyectos, sprints, historias, service.NuevoCriterios(proyectos, historias, criterios), time.Now),
		CambioDeEstado: service.NuevoEstados(proyectos, historias, criterios),
		Estimaciones:   service.NuevoEstimaciones(proyectos, historias),
		Tareas:         service.NuevoTareas(proyectos, historias, integrantes, tareas),
		Esfuerzo:       service.NuevoEsfuerzo(proyectos, historias, integrantes, tareas, registros, time.Now),
		Defectos:       casosDefectos,
		Horas:          horas,
		Panel:          panel,
		Reportes:       service.NuevoReportes(proyectos, sprints, historias, casosDeCierre, horas, defectos, panel),
		Estado:         estado,
		Edicion:        service.NuevoEdicionProyectos(proyectos, sprints, time.Now),
		Cierres:        casosDeCierre,
		Base:           base,

		// Editar y eliminar una historia (HU-06).
		EdicionHistorias: service.NuevoEdicionHistorias(proyectos, historias, criterios, tareas),

		// Sesiones de Planning Poker (HU-17) y sus votos (HU-18).
		Poker: service.NuevoPoker(proyectos, historias, integrantes, sesiones).ConVotos(sesiones),

		// Gráficos de las métricas (HU-36): usan el reporte de cada sprint de HU-37.
		Graficos: service.NuevoGraficos(service.NuevoReportes(proyectos, sprints, historias, casosDeCierre, horas, defectos, panel), casosDefectos, historias, time.Now),

		// Consultar y corregir los registros de horas (HU-26).
		RegistrosEsfuerzo: service.NuevoRegistrosEsfuerzo(proyectos, sprints, historias, integrantes, tareas, registros, horas, time.Now),
	}
	if o.verificador != nil {
		dep.Verificador, dep.Cuentas = o.verificador, service.NuevoCuentas(o.cuentas, integrantes)
	}
	return server.NewRouter(dep), nil
}
