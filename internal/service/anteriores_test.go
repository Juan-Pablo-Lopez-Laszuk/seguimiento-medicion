package service_test

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// Los tests de HU-15 usan el escenario de HU-14 con el Sprint 2 ya cerrado (5 de 13 SP, una de tres historias
// completada). El Sprint 1 del escenario está cerrado sin SP planificados (se cargó así, sin la foto de HU-14) y el
// Sprint 3 sigue abierto.
func nuevoEscenarioAnteriores(t *testing.T) *escenarioCierre {
	t.Helper()
	e := nuevoEscenarioCierre(t)
	if _, err := e.casosDeCierre(relojDelSprint()).Cerrar(context.Background(), e.proyectoID, 2); err != nil {
		t.Fatalf("cerrar el Sprint 2: %v", err)
	}
	return e
}

// errRepo es la falla de un repositorio en los tests de HU-15.
var errRepo = errors.New("base caída")

// sprintsDesordenados devuelve los sprints al revés, para probar que la lista no depende del orden del repositorio.
type sprintsDesordenados struct{ *memory.Sprints }

func (r sprintsDesordenados) ListarPorProyecto(ctx context.Context, proyectoID int64) ([]sprint.Sprint, error) {
	lista, err := r.Sprints.ListarPorProyecto(ctx, proyectoID)
	slices.Reverse(lista)
	return lista, err
}

// CA-15.1 · la lista va ordenada por número y los cerrados muestran SP completados y cumplimiento de la foto (RN1,
// RN2 y RN4)
func TestResumirSprints_OrdenadosConElResultadoDeLosCerrados(t *testing.T) {
	e := nuevoEscenarioAnteriores(t)
	casos := service.NuevoSprints(e.proyectos, sprintsDesordenados{e.sprints})

	p, lista, err := casos.Resumir(context.Background(), e.proyectoID)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if p.ID != e.proyectoID {
		t.Errorf("se esperaba el proyecto %d y se obtuvo %d", e.proyectoID, p.ID)
	}
	quiero := []struct {
		numero           int
		sp, cumplimiento string
	}{
		{1, "—", "—"}, // cerrado sin SP planificados: no hay resultado que mostrar
		{2, "5 de 13", "38,5 %"},
		{3, "—", "—"}, // abierto (RN1)
	}
	if len(lista) != len(quiero) {
		t.Fatalf("se esperaban %d sprints y se obtuvieron %d", len(quiero), len(lista))
	}
	for i, q := range quiero {
		r := lista[i]
		if r.Numero != q.numero || r.SPCompletadosTexto() != q.sp || r.CumplimientoTexto() != q.cumplimiento {
			t.Errorf("fila %d: se esperaba Sprint %d con %q y %q, y se obtuvo Sprint %d con %q y %q",
				i+1, q.numero, q.sp, q.cumplimiento, r.Numero, r.SPCompletadosTexto(), r.CumplimientoTexto())
		}
	}
}

// RN3 · un sprint cerrado sin historias muestra 0 de 0 y cumplimiento —
func TestResumirSprints_CerradoSinHistorias(t *testing.T) {
	e := &escenarioCierre{escenarioAsignacion: nuevoEscenarioAsignacion(t, true)}
	e.cierres = memory.NuevoCierres(e.sprints, e.historias)
	if _, err := e.casosDeCierre(relojDelSprint()).Cerrar(context.Background(), e.proyectoID, 2); err != nil {
		t.Fatalf("cerrar el Sprint 2: %v", err)
	}

	_, lista, err := service.NuevoSprints(e.proyectos, e.sprints).Resumir(context.Background(), e.proyectoID)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if r := lista[1]; r.SPCompletadosTexto() != "0 de 0" || r.CumplimientoTexto() != "—" {
		t.Errorf("se esperaba el Sprint 2 con \"0 de 0\" y \"—\", y se obtuvo %q y %q", r.SPCompletadosTexto(), r.CumplimientoTexto())
	}
	d, err := e.casosDeCierre(relojDelSprint()).Detalle(context.Background(), e.proyectoID, 2)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if !d.Cerrado || len(d.Historias) != 0 || d.Resumen() != "El sprint se cerró sin historias" {
		t.Errorf("se esperaba el detalle cerrado y sin historias, y se obtuvo %+v con resumen %q", d, d.Resumen())
	}
}

// Errores · el proyecto que no existe y la falla del repositorio se devuelven
func TestResumirSprints_Errores(t *testing.T) {
	e := nuevoEscenarioAnteriores(t)
	if _, _, err := service.NuevoSprints(e.proyectos, e.sprints).Resumir(context.Background(), 99); !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("proyecto que no existe: se esperaba project.ErrNoEncontrado y se obtuvo %v", err)
	}
	if _, _, err := service.NuevoSprints(e.proyectos, sprintsQueFallan{errListar: errRepo}).Resumir(context.Background(), e.proyectoID); !errors.Is(err, errRepo) {
		t.Errorf("repositorio que falla: se esperaba errRepo y se obtuvo %v", err)
	}
}

// CA-15.2 · el detalle de un sprint cerrado muestra la foto ordenada por número, con el título de hoy (RN5 y RN6)
func TestDetalleDeCierre_MuestraLaFoto(t *testing.T) {
	e := nuevoEscenarioAnteriores(t)
	h := e.guardada(t, e.hecha)
	h.Titulo = "Login corregido"
	if err := e.historias.ActualizarDatos(context.Background(), h); err != nil {
		t.Fatalf("cambiar el título: %v", err)
	}

	d, err := e.casosDeCierre(relojDelSprint()).Detalle(context.Background(), e.proyectoID, 2)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	if !d.Cerrado || d.Resumen() != "5 de 13 SP completados · 38,5 % de cumplimiento · 1 de 3 historias" {
		t.Errorf("se esperaba el Sprint 2 cerrado con 5 de 13 SP, y se obtuvo cerrado=%v %q", d.Cerrado, d.Resumen())
	}
	quiero := []service.FilaFoto{
		{Numero: e.hecha, Titulo: "Login corregido", StoryPoints: puntos(5), Completada: true},
		{Numero: e.enCurso, Titulo: "En el sprint", StoryPoints: puntos(3)},
		{Numero: e.sinEmp, Titulo: "En el sprint", StoryPoints: puntos(5)},
	}
	if len(d.Historias) != len(quiero) {
		t.Fatalf("se esperaban %d historias en la foto y se obtuvieron %+v", len(quiero), d.Historias)
	}
	for i, q := range quiero {
		f := d.Historias[i]
		if f.Numero != q.Numero || f.Titulo != q.Titulo || f.StoryPoints == nil || *f.StoryPoints != *q.StoryPoints || f.Completada != q.Completada {
			t.Errorf("fila %d: se esperaba %+v y se obtuvo %+v", i+1, q, f)
		}
	}
	if d.Historias[0].Identificador() != "HU-6" || d.Historias[0].Resultado() != "Completada" || d.Historias[1].Resultado() != "No completada" {
		t.Errorf("identificador o resultado inesperados: %q, %q, %q", d.Historias[0].Identificador(), d.Historias[0].Resultado(), d.Historias[1].Resultado())
	}
}

// RN6 · re-estimar una historia de la foto y pasarla a otro sprint no cambia el detalle
func TestDetalleDeCierre_NoCambiaDespuesDelCierre(t *testing.T) {
	e := nuevoEscenarioAnteriores(t)
	h := e.guardada(t, e.enCurso)
	if err := e.historias.ActualizarStoryPoints(context.Background(), h.ID, 8); err != nil {
		t.Fatalf("re-estimar: %v", err)
	}
	tercero := e.ids[3]
	if err := e.historias.ActualizarSprint(context.Background(), h.ID, backlog.EstadoEnSprint, &tercero); err != nil {
		t.Fatalf("asignar al Sprint 3: %v", err)
	}

	d, err := e.casosDeCierre(relojDelSprint()).Detalle(context.Background(), e.proyectoID, 2)
	if err != nil {
		t.Fatalf("error inesperado: %v", err)
	}
	f := d.Historias[1]
	if f.Numero != e.enCurso || *f.StoryPoints != 3 || f.Completada {
		t.Errorf("la foto tenía que seguir con HU-%d en 3 SP y no completada, y se obtuvo %+v", e.enCurso, f)
	}
	if d.Resumen() != "5 de 13 SP completados · 38,5 % de cumplimiento · 1 de 3 historias" {
		t.Errorf("el resumen no tenía que cambiar y es %q", d.Resumen())
	}
}

// Sección 3 · un sprint abierto no tiene detalle de cierre, y no es un error
func TestDetalleDeCierre_SprintAbierto(t *testing.T) {
	e := nuevoEscenarioAnteriores(t)
	d, err := e.casosDeCierre(relojDelSprint()).Detalle(context.Background(), e.proyectoID, 3)
	if err != nil || d.Cerrado {
		t.Errorf("se esperaba el Sprint 3 sin detalle de cierre y sin error, y se obtuvo cerrado=%v, %v", d.Cerrado, err)
	}
}

// fotosQueFallan es un repositorio de cierres cuya foto no se puede leer.
type fotosQueFallan struct{ *memory.Cierres }

func (fotosQueFallan) Foto(context.Context, int64) ([]sprint.FotoHistoria, error) {
	return nil, errRepo
}

// Errores · proyecto o sprint que no existen, y la falla al leer la foto
func TestDetalleDeCierre_Errores(t *testing.T) {
	e := nuevoEscenarioAnteriores(t)
	casos := e.casosDeCierre(relojDelSprint())
	if _, err := casos.Detalle(context.Background(), 99, 2); !errors.Is(err, project.ErrNoEncontrado) {
		t.Errorf("proyecto que no existe: se esperaba project.ErrNoEncontrado y se obtuvo %v", err)
	}
	if _, err := casos.Detalle(context.Background(), e.proyectoID, 9); !errors.Is(err, sprint.ErrNoEncontrado) {
		t.Errorf("sprint que no existe: se esperaba sprint.ErrNoEncontrado y se obtuvo %v", err)
	}
	conFalla := service.NuevoCierres(e.proyectos, e.sprints, e.historias, fotosQueFallan{e.cierres}, relojDelSprint)
	if _, err := conFalla.Detalle(context.Background(), e.proyectoID, 2); !errors.Is(err, errRepo) {
		t.Errorf("foto que falla: se esperaba errRepo y se obtuvo %v", err)
	}
}
