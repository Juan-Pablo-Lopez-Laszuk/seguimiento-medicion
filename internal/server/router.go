// Package server arma el servidor web: el router con todas las rutas, los handlers y el renderizado
// de las páginas. Lo usan tanto el servidor local (cmd/server) como la función de Vercel (api/index.go).
package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/auth"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
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
	// Cierres cierra el sprint activo (HU-14).
	Cierres      *service.Cierres
	Estimaciones *service.Estimaciones
	// EdicionHistorias edita y elimina una historia (HU-06).
	EdicionHistorias *service.EdicionHistorias
	// CambioDeEstado cambia el estado de una historia (HU-08); Estado es el del proyecto (HU-04).
	CambioDeEstado *service.Estados
	Tareas         *service.Tareas
	// Esfuerzo registra las horas trabajadas en una historia o tarea (HU-24).
	Esfuerzo *service.Esfuerzo
	// RegistrosEsfuerzo consulta y corrige los registros de horas de un proyecto (HU-26).
	RegistrosEsfuerzo *service.RegistrosEsfuerzo
	// Graficos arma los datos de los gráficos de métricas de un proyecto (HU-36).
	Graficos *service.Graficos
	// Defectos registra los defectos de un proyecto (HU-27).
	Defectos *service.Defectos
	// Horas compara las horas estimadas y reales de un proyecto (HU-33).
	Horas *service.Horas
	// Panel junta los indicadores del proyecto (HU-35).
	Panel *service.Panel
	// Reportes arma el reporte de un sprint (HU-37) y el del proyecto (HU-38).
	Reportes *service.Reportes
	Estado   *service.EstadoProyectos
	Edicion  *service.EdicionProyectos
	// Base es la base de datos que chequea /health; nil si la app guarda en memoria (TEC-04).
	Base VerificadorBase

	// Verificador valida la sesión de cada pedido (TEC-05). Si es nil el login no está configurado y la aplicación no pide
	// sesión; si no, hace falta Cuentas.
	Verificador auth.Verificador
	// Cuentas registra e ingresa cuentas de usuario (TEC-05).
	Cuentas *service.Cuentas
}

// NewRouter arma el router con todas las rutas de la aplicación. /health queda siempre abierto. Con el login activado
// (dep.Verificador) también quedan abiertas /login, /registro y /salir, y todas las demás rutas piden una sesión válida;
// sin login no se pide nada (TEC-05, RN1).
func NewRouter(dep Dependencias) http.Handler {
	r := chi.NewRouter()
	r.Get("/health", health(dep.Base))
	if dep.Verificador == nil {
		rutas(r, dep)
		return r
	}
	if dep.Cuentas == nil {
		panic("server: el login activado necesita Dependencias.Cuentas")
	}
	c := cuentas{casos: dep.Cuentas}
	r.Get("/login", c.formularioDeLogin)
	r.Post("/login", c.ingresar)
	r.Get("/registro", c.formularioDeRegistro)
	r.Post("/registro", c.registrar)
	r.Post("/salir", c.salir)
	r.Group(func(r chi.Router) {
		r.Use(requiereSesion(dep.Verificador))
		rutas(r, dep)
	})
	return r
}

// rutas registra las rutas de las historias de usuario: todas menos /health y las de las cuentas.
func rutas(r chi.Router, dep Dependencias) {
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

	pl := planificacion{casos: dep.Asignaciones, cierres: dep.Cierres}
	r.Get("/proyectos/{id}/sprints/{sprint}", pl.ver)
	r.Post("/proyectos/{id}/sprints/{sprint}/historias", pl.asignar)
	r.Post("/proyectos/{id}/sprints/{sprint}/historias/{historia}/quitar", pl.accion((*service.Asignaciones).Quitar, "quitada"))
	r.Post("/proyectos/{id}/sprints/{sprint}/historias/{historia}/empezar", pl.accion((*service.Asignaciones).Empezar, "empezada"))
	r.Post("/proyectos/{id}/sprints/{sprint}/historias/{historia}/completar", pl.accion((*service.Asignaciones).Completar, "completada"))

	r.Get("/proyectos/{id}/sprints/{sprint}/reporte", reportes{casos: dep.Reportes}.sprint)
	r.Get("/proyectos/{id}/reporte", reportes{casos: dep.Reportes}.proyecto)

	ci := cierres{casos: dep.Cierres}
	r.Get("/proyectos/{id}/sprints/{sprint}/cerrar", ci.confirmar)
	r.Post("/proyectos/{id}/sprints/{sprint}/cerrar", ci.cerrar)

	in := integrantes{casos: dep.Integrantes}
	r.Get("/proyectos/{id}/integrantes", in.listar)
	r.Get("/proyectos/{id}/integrantes/nuevo", in.nuevo)
	r.Post("/proyectos/{id}/integrantes", in.registrar)
	r.Post("/proyectos/{id}/integrantes/{integranteID}/baja", in.darDeBaja)

	hs := historias{casos: dep.Historias, criterios: dep.Criterios}
	r.Get("/proyectos/{id}/backlog", hs.listar)
	r.Get("/proyectos/{id}/backlog/nuevo", hs.nueva)
	r.Post("/proyectos/{id}/backlog", hs.crear)
	r.Post("/proyectos/{id}/backlog/{numero}/subir", hs.mover(backlog.Subir))
	r.Post("/proyectos/{id}/backlog/{numero}/bajar", hs.mover(backlog.Bajar))

	cr := criterios{casos: dep.Criterios}
	r.Get("/proyectos/{id}/backlog/{numero}", cr.ver)
	r.Post("/proyectos/{id}/backlog/{numero}/criterios", cr.agregar)
	r.Post("/proyectos/{id}/backlog/{numero}/criterios/{criterioID}/cumplido", cr.marcar)

	es := estados{casos: dep.CambioDeEstado, vista: dep.Criterios}
	r.Post("/proyectos/{id}/backlog/{numero}/estado", es.cambiar)

	est := estimaciones{casos: dep.Estimaciones, vista: dep.Criterios}
	r.Post("/proyectos/{id}/backlog/{numero}/estimacion", est.estimar)

	eh := edicionHistorias{casos: dep.EdicionHistorias}
	r.Get("/proyectos/{id}/backlog/{numero}/editar", eh.formulario)
	r.Post("/proyectos/{id}/backlog/{numero}/editar", eh.editar)
	r.Get("/proyectos/{id}/backlog/{numero}/eliminar", eh.confirmar)
	r.Post("/proyectos/{id}/backlog/{numero}/eliminar", eh.eliminar)

	ta := tareas{casos: dep.Tareas}
	r.Get("/proyectos/{id}/backlog/{numero}/tareas", ta.ver)
	r.Post("/proyectos/{id}/backlog/{numero}/tareas", ta.agregar)
	r.Post("/proyectos/{id}/backlog/{numero}/tareas/{tareaID}/horas", ta.cambiarHoras)
	r.Post("/proyectos/{id}/backlog/{numero}/tareas/{tareaID}/completada", ta.marcar)

	ef := esfuerzo{casos: dep.Esfuerzo}
	r.Get("/proyectos/{id}/backlog/{numero}/esfuerzo", ef.ver)
	r.Post("/proyectos/{id}/backlog/{numero}/esfuerzo", ef.registrar)

	re := registrosEsfuerzo{casos: dep.RegistrosEsfuerzo}
	r.Get("/proyectos/{id}/esfuerzo", re.consultar)
	r.Get("/proyectos/{id}/esfuerzo/{registro}/corregir", re.verCorreccion)
	r.Post("/proyectos/{id}/esfuerzo/{registro}/corregir", re.corregir)

	df := defectos{casos: dep.Defectos}
	r.Get("/proyectos/{id}/defectos", df.ver)
	r.Post("/proyectos/{id}/defectos", df.registrar)
	r.Post("/proyectos/{id}/defectos/{defecto}/estado", df.cambiarEstado)
	r.Get("/proyectos/{id}/horas", horas{casos: dep.Horas}.ver)
	gr := graficos{casos: dep.Graficos}
	r.Get("/proyectos/{id}/graficos", gr.ver)
	r.Get("/proyectos/{id}/graficos/datos", gr.datos)
	r.Get("/proyectos/{id}/panel", panel{casos: dep.Panel}.ver)
}

// inicio muestra la página principal.
func inicio(w http.ResponseWriter, _ *http.Request) {
	render(w, http.StatusOK, "inicio.html", Pagina{Titulo: "Inicio"})
}
