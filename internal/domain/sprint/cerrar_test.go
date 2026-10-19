package sprint_test

import (
	"errors"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
)

// Los tests de HU-14 usan los casos límite de la spec: el Sprint 1 del 12/10 al 18/10 y el Sprint 2 del 19/10 al
// 25/10, los dos sin cerrar.
func sprintsParaCerrar(t *testing.T) []sprint.Sprint {
	return []sprint.Sprint{
		{ID: 10, ProyectoID: 1, Numero: 1, FechaInicio: fecha(t, "2026-10-12"), FechaFin: fecha(t, "2026-10-18"), Estado: sprint.EstadoPlanificado},
		{ID: 20, ProyectoID: 1, Numero: 2, FechaInicio: fecha(t, "2026-10-19"), FechaFin: fecha(t, "2026-10-25"), Estado: sprint.EstadoPlanificado},
	}
}

func puntos(n int) *int { return &n }

func idDe(n int64) *int64 { return &n }

// historiasParaCerrar son las del Sprint 1 de la spec (HU-1 Hecha, HU-2 En progreso, HU-3 En sprint), más una del
// Sprint 2 y una del backlog, que el cierre no tiene que tocar.
func historiasParaCerrar() []backlog.Historia {
	return []backlog.Historia{
		{ID: 1, ProyectoID: 1, Numero: 1, Titulo: "Login", Estado: backlog.EstadoHecho, StoryPoints: puntos(5), SprintID: idDe(10)},
		{ID: 2, ProyectoID: 1, Numero: 2, Titulo: "Registro", Estado: backlog.EstadoEnProgreso, StoryPoints: puntos(3), SprintID: idDe(10)},
		{ID: 3, ProyectoID: 1, Numero: 3, Titulo: "Perfil", Estado: backlog.EstadoEnSprint, StoryPoints: puntos(5), SprintID: idDe(10)},
		{ID: 4, ProyectoID: 1, Numero: 4, Titulo: "Ayuda", Estado: backlog.EstadoEnSprint, StoryPoints: puntos(2), SprintID: idDe(20)},
		{ID: 5, ProyectoID: 1, Numero: 5, Titulo: "Reportes", Estado: backlog.EstadoPendiente, StoryPoints: puntos(8)},
	}
}

func cerrarElSprint1(t *testing.T, hoy string, historias []backlog.Historia) sprint.Cierre {
	t.Helper()
	sprints := sprintsParaCerrar(t)
	cierre, err := sprint.Cerrar(sprints[0], sprints, historias, fecha(t, hoy))
	if err != nil {
		t.Fatalf("no se esperaba error al cerrar el Sprint 1: %v", err)
	}
	return cierre
}

// CA-14.3 · RN3 · el cierre deja el sprint Cerrado con los SP planificados y completados, y la foto de sus historias
func TestCerrar_GuardaLaFotoYLosSP(t *testing.T) {
	cierre := cerrarElSprint1(t, "2026-10-18", historiasParaCerrar())

	s := cierre.Sprint
	if s.ID != 10 || s.Estado != sprint.EstadoCerrado {
		t.Errorf("se esperaba el Sprint 1 (ID 10) Cerrado y se obtuvo ID %d %q", s.ID, s.Estado)
	}
	if s.SPPlanificados == nil || *s.SPPlanificados != 13 || s.SPCompletados == nil || *s.SPCompletados != 5 {
		t.Errorf("se esperaban 13 SP planificados y 5 completados, y se obtuvo %v y %v", s.SPPlanificados, s.SPCompletados)
	}
	esperada := []struct {
		historiaID int64
		numero     int
		sp         int
		completada bool
	}{{1, 1, 5, true}, {2, 2, 3, false}, {3, 3, 5, false}}
	if len(cierre.Foto) != len(esperada) {
		t.Fatalf("la foto tiene %d historias y se esperaban %d: %+v", len(cierre.Foto), len(esperada), cierre.Foto)
	}
	for i, e := range esperada {
		f := cierre.Foto[i]
		if f.HistoriaID != e.historiaID || f.Numero != e.numero || f.StoryPoints == nil || *f.StoryPoints != e.sp || f.Completada != e.completada {
			t.Errorf("foto[%d]: se esperaba HU-%d con %d SP y completada=%v, y se obtuvo %+v", i, e.numero, e.sp, e.completada, f)
		}
	}
}

// CA-14.2 · RN4 · las historias no completadas vuelven al backlog Pendientes y sin sprint, con el resto de sus datos;
// las Hechas y las de otros sprints no se tocan
func TestCerrar_LasNoCompletadasVuelvenAlBacklog(t *testing.T) {
	historias := historiasParaCerrar()
	cierre := cerrarElSprint1(t, "2026-10-18", historias)

	if len(cierre.Devueltas) != 2 {
		t.Fatalf("se esperaban 2 historias devueltas al backlog y se obtuvo %+v", cierre.Devueltas)
	}
	for i, numero := range []int{2, 3} {
		h := cierre.Devueltas[i]
		if h.Numero != numero || h.Estado != backlog.EstadoPendiente || h.SprintID != nil {
			t.Errorf("se esperaba HU-%d Pendiente y sin sprint, y se obtuvo %+v", numero, h)
		}
		if h.StoryPoints == nil || *h.StoryPoints != *historias[numero-1].StoryPoints || h.Titulo != historias[numero-1].Titulo {
			t.Errorf("HU-%d tenía que conservar su estimación y su título, y se obtuvo %+v", numero, h)
		}
	}
	// lo que recibe Cerrar no cambia: el cierre devuelve copias
	if historias[1].Estado != backlog.EstadoEnProgreso || historias[1].SprintID == nil {
		t.Errorf("Cerrar no tiene que modificar las historias que recibe y HU-2 quedó %+v", historias[1])
	}
}

// RN3 · la foto guarda los story points de ese momento: cambiar después la historia no la cambia
func TestCerrar_LaFotoNoComparteLosStoryPoints(t *testing.T) {
	historias := historiasParaCerrar()
	cierre := cerrarElSprint1(t, "2026-10-18", historias)

	*historias[1].StoryPoints = 8 // se re-estima HU-2 después del cierre
	*cierre.Devueltas[0].StoryPoints = 13

	if *cierre.Foto[1].StoryPoints != 3 {
		t.Errorf("la foto de HU-2 tenía que seguir en 3 SP y quedó en %d", *cierre.Foto[1].StoryPoints)
	}
}

// CA-14.1 · RN1 y RN2 · el sprint activo se cierra vencido y también antes de su fecha de fin
func TestCerrar_SprintActivoEnCualquierDia(t *testing.T) {
	for _, hoy := range []string{"2026-10-12", "2026-10-16", "2026-10-18", "2026-10-20"} {
		t.Run(hoy, func(t *testing.T) {
			cierre := cerrarElSprint1(t, hoy, historiasParaCerrar())
			if cierre.Sprint.Estado != sprint.EstadoCerrado {
				t.Errorf("el %s el Sprint 1 es el activo y se tenía que cerrar", hoy)
			}
		})
	}
}

// RN7 y casos límite · sin historias, todas Hechas, una en 0 SP y una sin estimar
func TestCerrar_CasosLimite(t *testing.T) {
	enCero := []backlog.Historia{
		{ID: 1, Numero: 1, Estado: backlog.EstadoHecho, StoryPoints: puntos(0), SprintID: idDe(10)},
		{ID: 2, Numero: 2, Estado: backlog.EstadoHecho, StoryPoints: puntos(3), SprintID: idDe(10)},
	}
	sinEstimar := []backlog.Historia{
		{ID: 1, Numero: 1, Estado: backlog.EstadoHecho, StoryPoints: puntos(5), SprintID: idDe(10)},
		{ID: 2, Numero: 2, Estado: backlog.EstadoEnSprint, SprintID: idDe(10)},
	}
	casos := []struct {
		nombre                    string
		historias                 []backlog.Historia
		planificados, completados int
		foto, devueltas           int
	}{
		{"sprint sin historias", nil, 0, 0, 0, 0},
		{"todas Hechas, una en 0 SP", enCero, 3, 3, 2, 0},
		{"una sin estimar suma 0", sinEstimar, 5, 5, 2, 1},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			cierre := cerrarElSprint1(t, "2026-10-18", c.historias)
			s := cierre.Sprint
			if s.SPPlanificados == nil || *s.SPPlanificados != c.planificados || s.SPCompletados == nil || *s.SPCompletados != c.completados {
				t.Errorf("se esperaban %d de %d SP y se obtuvo %v de %v", c.completados, c.planificados, s.SPCompletados, s.SPPlanificados)
			}
			if len(cierre.Foto) != c.foto || len(cierre.Devueltas) != c.devueltas {
				t.Errorf("se esperaban %d en la foto y %d devueltas, y se obtuvo %d y %d", c.foto, c.devueltas, len(cierre.Foto), len(cierre.Devueltas))
			}
		})
	}
	// la de 0 SP queda en la foto con 0, y la sin estimar sin story points
	if cierre := cerrarElSprint1(t, "2026-10-18", enCero); cierre.Foto[0].StoryPoints == nil || *cierre.Foto[0].StoryPoints != 0 {
		t.Errorf("la historia de 0 SP tenía que quedar en la foto con 0, y quedó %v", cierre.Foto[0].StoryPoints)
	}
	if cierre := cerrarElSprint1(t, "2026-10-18", sinEstimar); cierre.Foto[1].StoryPoints != nil {
		t.Errorf("la historia sin estimar tenía que quedar en la foto sin story points, y quedó %d", *cierre.Foto[1].StoryPoints)
	}
}

// CA-14.1 y CA-14.4 · sección 7 · solo se cierra el sprint activo, y uno cerrado no se cierra otra vez
func TestCerrar_Errores(t *testing.T) {
	casos := []struct {
		nombre   string
		numero   int // sprint que se quiere cerrar
		hoy      string
		cerrado1 bool // el Sprint 1 ya está cerrado
		base     error
		mensaje  string
	}{
		{"un sprint futuro", 2, "2026-10-18", false, sprint.ErrNoEsElActivo, "solo se cierra el sprint activo (el Sprint 1)"},
		{"uno posterior a otro abierto", 2, "2026-10-20", false, sprint.ErrNoEsElActivo, "solo se cierra el sprint activo (el Sprint 1)"},
		{"hoy no hay sprint activo", 1, "2026-10-10", false, sprint.ErrNoEsElActivo, "solo se cierra el sprint activo: hoy no hay ninguno"},
		{"cerrado, sin activo después", 1, "2026-10-18", true, sprint.ErrYaCerrado, "el Sprint 1 ya está cerrado"},
		{"cerrado, con otro activo", 1, "2026-10-20", true, sprint.ErrYaCerrado, "el Sprint 1 ya está cerrado"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			sprints := sprintsParaCerrar(t)
			if c.cerrado1 {
				sprints[0].Estado = sprint.EstadoCerrado
			}
			_, err := sprint.Cerrar(sprints[c.numero-1], sprints, historiasParaCerrar(), fecha(t, c.hoy))
			if !errors.Is(err, c.base) {
				t.Fatalf("se esperaba %v y se obtuvo %v", c.base, err)
			}
			if err.Error() != c.mensaje {
				t.Errorf("mensaje = %q, se esperaba %q", err.Error(), c.mensaje)
			}
		})
	}
}

// RN1 · el sprint activo se decide con la fecha de hoy, sin la hora
func TestCerrar_IgnoraLaHora(t *testing.T) {
	sprints := sprintsParaCerrar(t)
	hoy := time.Date(2026, 10, 12, 0, 30, 0, 0, time.UTC)
	if _, err := sprint.Cerrar(sprints[0], sprints, nil, hoy); err != nil {
		t.Errorf("el 12/10 a las 00:30 el Sprint 1 ya es el activo y se obtuvo %v", err)
	}
}
