package app_test

import (
	"net/http"
	neturl "net/url"
	"strings"
	"testing"
)

// eliminarUnaHistoriaConRegistros crea el proyecto 1 con la HU-1, la HU-2 (sin nada) y Carolina, le agrega a la HU-1 horas
// registradas o un defecto según pida el test, y comprueba que eliminarla da 409 con el motivo, que no se borra nada y que la
// HU-2 sí se puede eliminar (HU-06, RN10).
func eliminarUnaHistoriaConRegistros(t *testing.T, h http.Handler, conHoras bool) {
	t.Helper()
	crearProyecto(h, "Con registros asociados")
	post(h, "/proyectos/1/backlog", neturl.Values{"titulo": {"Historia con registros"}, "prioridad": {"Alta"}})
	post(h, "/proyectos/1/backlog", neturl.Values{"titulo": {"Historia sin nada"}, "prioridad": {"Baja"}})
	post(h, "/proyectos/1/backlog/1/criterios", neturl.Values{"descripcion": {"Un criterio que no se tiene que perder"}})
	if conHoras {
		post(h, "/proyectos/1/integrantes", neturl.Values{"nombre": {"Carolina"}, "email": {"caro@mail.com"}, "rol": {"ProductBuilder"}})
		campos := neturl.Values{"integrante": {"1"}, "fecha": {"2026-10-05"}, "actividad": {"Trabajo ya registrado"}, "horas": {"2"}}
		if rec := post(h, "/proyectos/1/backlog/1/esfuerzo", campos); rec.Code != http.StatusSeeOther {
			t.Fatalf("registrar esfuerzo: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
		}
	} else {
		post(h, "/proyectos/1/sprints", neturl.Values{"objetivo": {"MVP"}, "fecha_inicio": {"2026-10-05"}, "fecha_fin": {"2026-10-11"}})
		campos := neturl.Values{"descripcion": {"Defecto de la historia"}, "severidad": {"Alta"}, "sprint": {"1"}, "historia": {"1"}}
		if rec := post(h, "/proyectos/1/defectos", campos); rec.Code != http.StatusSeeOther {
			t.Fatalf("registrar defecto: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
		}
	}

	rec := post(h, "/proyectos/1/backlog/1/eliminar", neturl.Values{})

	if rec.Code != http.StatusConflict || !strings.Contains(rec.Body.String(), "tiene registros asociados") {
		t.Fatalf("eliminar una historia con registros: se esperaba %d con el motivo y se obtuvo %d:\n%s", http.StatusConflict, rec.Code, rec.Body.String())
	}
	backlog := get(h, "/proyectos/1/backlog")
	if !strings.Contains(backlog, "Historia con registros") {
		t.Error("la historia con registros tendría que seguir en el backlog")
	}
	if !strings.Contains(get(h, "/proyectos/1/backlog/1"), "Un criterio que no se tiene que perder") {
		t.Error("un borrado rechazado no borra nada: la historia tendría que conservar su criterio")
	}
	if rec := post(h, "/proyectos/1/backlog/2/eliminar", neturl.Values{}); rec.Code != http.StatusSeeOther {
		t.Errorf("la historia sin registros sí se elimina: se esperaba %d y se obtuvo %d", http.StatusSeeOther, rec.Code)
	}
}

// HU-06, RN10: sin base, una historia con horas registradas tampoco se elimina (antes se borraba y dejaba los registros
// sueltos, mientras que con PostgreSQL daba 409).
func TestNewHandlerConBase_SinURLNoSeEliminaUnaHistoriaConHorasRegistradas(t *testing.T) {
	eliminarUnaHistoriaConRegistros(t, nuevaApp(t, ""), true)
}

func TestNewHandlerConBase_SinURLNoSeEliminaUnaHistoriaConUnDefecto(t *testing.T) {
	eliminarUnaHistoriaConRegistros(t, nuevaApp(t, ""), false)
}

// HU-06, RN10: con PostgreSQL pasa lo mismo, con el mismo mensaje y sin borrar nada.
func TestNewHandlerConBase_ConURLNoSeEliminaUnaHistoriaConHorasRegistradas(t *testing.T) {
	eliminarUnaHistoriaConRegistros(t, nuevaApp(t, urlConEsquemaMigrado(t)), true)
}

func TestNewHandlerConBase_ConURLNoSeEliminaUnaHistoriaConUnDefecto(t *testing.T) {
	eliminarUnaHistoriaConRegistros(t, nuevaApp(t, urlConEsquemaMigrado(t)), false)
}
