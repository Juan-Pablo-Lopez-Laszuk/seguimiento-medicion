package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/app"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/ejemplo"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
)

// cargarEjemplo arma la aplicación con la base url (vacía = memoria) y le carga el proyecto de ejemplo.
func cargarEjemplo(t *testing.T, url string) (server.Dependencias, project.Proyecto) {
	t.Helper()
	dep, err := app.DependenciasConBase(url)
	if err != nil {
		t.Fatalf("armar la aplicación: %v", err)
	}
	p, err := ejemplo.Cargar(context.Background(), dep, ejemplo.Datos{})
	if err != nil {
		t.Fatalf("cargar el ejemplo: %v", err)
	}
	return dep, p
}

// verificarEjemplo controla lo que muestra la demo: dos sprints cerrados, el tercero activo, horas y defectos.
func verificarEjemplo(t *testing.T, dep server.Dependencias, p project.Proyecto) {
	t.Helper()
	ctx := context.Background()
	if p.Nombre != ejemplo.NombrePorDefecto {
		t.Errorf("nombre: se esperaba %q y se obtuvo %q", ejemplo.NombrePorDefecto, p.Nombre)
	}

	panel, err := dep.Panel.Ver(ctx, p.ID)
	if err != nil {
		t.Fatalf("panel: %v", err)
	}
	if panel.Estado != project.EstadoEnCurso {
		t.Errorf("estado: se esperaba En curso y se obtuvo %q", panel.Estado)
	}
	if panel.Activo == nil || panel.Activo.Numero != 3 {
		t.Errorf("se esperaba el Sprint 3 activo y se obtuvo %+v", panel.Activo)
	}
	if panel.Velocidad.Valor != 9.5 || panel.Velocidad.SprintsUsados != 2 {
		t.Errorf("velocidad: se esperaba 9,5 con 2 sprints y se obtuvo %+v", panel.Velocidad)
	}
	if panel.Horas.Estimadas != 54 || panel.Horas.Reales != 55 {
		t.Errorf("horas: se esperaban 54 estimadas y 55 reales y se obtuvo %+v", panel.Horas)
	}
	if d := panel.Defectos; d.Detectados != 5 || d.Resueltos != 3 || d.Abiertos != 2 {
		t.Errorf("defectos: se esperaban 5 detectados, 3 resueltos y 2 abiertos y se obtuvo %+v", d)
	}

	_, sprints, err := dep.Sprints.Resumir(ctx, p.ID)
	if err != nil {
		t.Fatalf("sprints: %v", err)
	}
	resultados := make([]string, len(sprints))
	for i, s := range sprints {
		resultados[i] = s.SPCompletadosTexto()
	}
	if got := strings.Join(resultados, " | "); got != "8 de 16 | 11 de 13 | — | —" {
		t.Errorf("SP completados por sprint: se obtuvo %q", got)
	}

	// En la demo se asigna una historia lista al sprint activo y se ve una sin estimar.
	vista, err := dep.Asignaciones.Ver(ctx, p.ID, 3)
	if err != nil {
		t.Fatalf("sprint 3: %v", err)
	}
	motivos := map[string]string{}
	for _, c := range vista.Candidatas {
		motivos[c.Titulo] = c.Motivo
	}
	if motivo, ok := motivos["Exportar la agenda"]; !ok || motivo != "" {
		t.Errorf("\"Exportar la agenda\" tendría que estar lista para asignar; candidatas: %v", motivos)
	}
	if motivos["Historial del paciente"] == "" {
		t.Errorf("\"Historial del paciente\" tendría que figurar sin estimar; candidatas: %v", motivos)
	}
}

func TestEjemplo_EnMemoria(t *testing.T) {
	dep, p := cargarEjemplo(t, "")
	verificarEjemplo(t, dep, p)
}

func TestEjemplo_EnPostgres(t *testing.T) {
	dep, p := cargarEjemplo(t, urlConEsquemaMigrado(t))
	verificarEjemplo(t, dep, p)
}

// Cargarlo dos veces en la misma base no duplica nada: el nombre del proyecto ya existe.
func TestEjemplo_NoSeCargaDosVeces(t *testing.T) {
	dep, _ := cargarEjemplo(t, "")

	_, err := ejemplo.Cargar(context.Background(), dep, ejemplo.Datos{})

	var errs domain.ErroresValidacion
	if !errors.As(err, &errs) || !errors.Is(errs[project.CampoNombre], project.ErrNombreRepetido) {
		t.Fatalf("se esperaba el error de nombre repetido y se obtuvo %v", err)
	}
	proyectos, err := dep.Proyectos.Listar(context.Background())
	if err != nil || len(proyectos) != 1 {
		t.Errorf("tendría que haber un solo proyecto y hay %d (%v)", len(proyectos), err)
	}
}

// Con otro nombre y los emails de las cuentas del equipo, cada uno queda como integrante con su email.
func TestEjemplo_ConOtroNombreYEmails(t *testing.T) {
	dep, err := app.DependenciasConBase("")
	if err != nil {
		t.Fatal(err)
	}
	emails := [3]string{"jp@facultad.test", "mariano@facultad.test", "caro@facultad.test"}

	p, err := ejemplo.Cargar(context.Background(), dep, ejemplo.Datos{Nombre: "Demo del 02/11", Emails: emails})

	if err != nil {
		t.Fatalf("cargar el ejemplo: %v", err)
	}
	if p.Nombre != "Demo del 02/11" {
		t.Errorf("se esperaba el nombre pedido y se obtuvo %q", p.Nombre)
	}
	_, integrantes, err := dep.Integrantes.Listar(context.Background(), p.ID)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, i := range integrantes {
		got = append(got, i.Email)
	}
	if strings.Join(got, ",") != strings.Join(emails[:], ",") {
		t.Errorf("emails de los integrantes: se esperaba %v y se obtuvo %v", emails, got)
	}
}
