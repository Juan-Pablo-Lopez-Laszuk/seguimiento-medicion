package reportepdf_test

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/defect"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/metrics"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/reportepdf"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

func fecha(dia, mes int) time.Time {
	return time.Date(2026, time.Month(mes), dia, 0, 0, 0, 0, time.UTC)
}

func sp(n int) *int { return &n }

// El proyecto del ejemplo de la spec (sección 6).
var softwareMetrics = project.Proyecto{ID: 1, Nombre: "Software Metrics", Descripcion: "TPI de ICS", FechaInicio: fecha(5, 10), FechaFin: fecha(30, 11)}

// reporteDelSprint2 es el reporte del Sprint 2 del ejemplo: cerrado con HU-2 (completada) y HU-3, y dos defectos.
func reporteDelSprint2() service.ReporteSprint {
	return service.ReporteSprint{
		Proyecto:    softwareMetrics,
		Sprint:      sprint.Sprint{ProyectoID: 1, Numero: 2, Objetivo: "Planificar sprints", FechaInicio: fecha(19, 10), FechaFin: fecha(25, 10), Estado: sprint.EstadoCerrado},
		SP:          metrics.StoryPoints{Planificados: 8, Completados: 3},
		Completadas: metrics.Porcentaje{Hechas: 1, Planificadas: 2, Valor: 50},
		Velocidad:   metrics.Velocidad{Valor: 4, SprintsUsados: 2},
		Horas:       metrics.Desviacion{Estimadas: 6, Reales: 3, Horas: -3, Porcentaje: -50, Aplica: true},
		Historias: []service.FilaReporte{
			{FilaFoto: service.FilaFoto{ProyectoID: 1, Numero: 2, Titulo: "Registro", StoryPoints: sp(3), Completada: true},
				Horas: metrics.Desviacion{Estimadas: 4, Reales: 3, Horas: -1, Porcentaje: -25, Aplica: true}},
			{FilaFoto: service.FilaFoto{ProyectoID: 1, Numero: 3, Titulo: "Perfil"},
				Horas: metrics.Desviacion{Estimadas: 2, Reales: 0, Horas: -2, Porcentaje: -100, Aplica: true}},
		},
		Conteo: metrics.ConteoSprint{Detectados: 1, Resueltos: 1},
		Defectos: []service.FilaDefecto{
			{Defecto: defect.Defecto{ID: 2, Descripcion: "El registro tarda mucho", Severidad: defect.SeveridadMedia, Estado: defect.EstadoResuelto},
				Historia: "HU-2 · Registro", DetectadoEn: "Sprint 1", ResueltoEn: "Sprint 2"},
			{Defecto: defect.Defecto{ID: 3, Descripcion: "El perfil no carga la foto", Severidad: defect.SeveridadBaja, Estado: defect.EstadoAbierto},
				DetectadoEn: "Sprint 2", ResueltoEn: "—"},
		},
	}
}

// reporteDelSprint3 es el del Sprint 3 del ejemplo: abierto, con HU-3 sin estimar y sin defectos.
func reporteDelSprint3() service.ReporteSprint {
	return service.ReporteSprint{
		Proyecto:  softwareMetrics,
		Sprint:    sprint.Sprint{ProyectoID: 1, Numero: 3, Objetivo: "Ver métricas", FechaInicio: fecha(26, 10), FechaFin: fecha(1, 11), Estado: sprint.EstadoPlanificado},
		SP:        metrics.StoryPoints{SinEstimar: 1},
		Velocidad: metrics.Velocidad{Valor: 4, SprintsUsados: 2},
		Historias: []service.FilaReporte{{FilaFoto: service.FilaFoto{ProyectoID: 1, Numero: 3, Titulo: "Perfil"}}},
	}
}

func reporteDelSprint1() service.ReporteSprint {
	r := reporteDelSprint2()
	r.Sprint = sprint.Sprint{ProyectoID: 1, Numero: 1, Objetivo: "Gestionar el backlog", FechaInicio: fecha(12, 10), FechaFin: fecha(18, 10), Estado: sprint.EstadoCerrado}
	r.SP = metrics.StoryPoints{Planificados: 13, Completados: 5}
	r.Velocidad = metrics.Velocidad{Valor: 5, SprintsUsados: 1}
	return r
}

func reporteDelProyecto() service.ReporteProyecto {
	resumen := service.VistaPanel{
		EstadoProyecto: service.EstadoProyecto{Proyecto: softwareMetrics, Estado: project.EstadoEnCurso, StoryPoints: metrics.StoryPoints{Planificados: 13, Completados: 8}},
		Velocidad:      metrics.Velocidad{Valor: 4, SprintsUsados: 2},
		Horas:          metrics.Desviacion{Estimadas: 12, Reales: 11, Horas: -1, Porcentaje: -8.3333, Aplica: true},
		Defectos:       service.ResumenDefectos{Detectados: 3, Resueltos: 2, Abiertos: 1},
	}
	return service.ReporteProyecto{Resumen: resumen, Sprints: []service.ReporteSprint{reporteDelSprint1(), reporteDelSprint2(), reporteDelSprint3()}}
}

// buscar devuelve la posición del primer bloque del tipo y el texto pedidos, o -1.
func buscar(d reportepdf.Documento, tipo reportepdf.Tipo, texto string) int {
	for i, b := range d.Bloques {
		if b.Tipo == tipo && b.Texto == texto {
			return i
		}
	}
	return -1
}

// tablaDespuesDe devuelve la primera tabla que sigue a la sección pedida.
func tablaDespuesDe(t *testing.T, d reportepdf.Documento, seccion string) reportepdf.Bloque {
	t.Helper()
	i := buscar(d, reportepdf.Seccion, seccion)
	if i < 0 {
		t.Fatalf("el documento no tiene la sección %q", seccion)
	}
	if i+1 >= len(d.Bloques) || d.Bloques[i+1].Tipo != reportepdf.Tabla {
		t.Fatalf("después de la sección %q tiene que venir una tabla", seccion)
	}
	return d.Bloques[i+1]
}

func verificarTabla(t *testing.T, tabla reportepdf.Bloque, encabezados []string, filas [][]string) {
	t.Helper()
	if !reflect.DeepEqual(tabla.Encabezados, encabezados) {
		t.Errorf("encabezados = %q, se esperaba %q", tabla.Encabezados, encabezados)
	}
	if !reflect.DeepEqual(tabla.Filas, filas) {
		t.Errorf("filas = %q, se esperaba %q", tabla.Filas, filas)
	}
	if len(tabla.Anchos) != len(encabezados) {
		t.Errorf("la tabla tiene %d anchos para %d columnas", len(tabla.Anchos), len(encabezados))
	}
}

// Sección 3 · el PDF de un sprint arranca con el título, el objetivo, las fechas y el estado, como la página
func TestDocumentoSprint_Encabezado(t *testing.T) {
	d := reportepdf.DocumentoSprint(reporteDelSprint2())

	if d.Titulo != "Reporte del Sprint 2 · Software Metrics" {
		t.Errorf("Titulo = %q", d.Titulo)
	}
	esperados := []reportepdf.Bloque{
		{Tipo: reportepdf.Titulo, Texto: "Reporte del Sprint 2 · Software Metrics"},
		{Tipo: reportepdf.Parrafo, Texto: "Planificar sprints"},
		{Tipo: reportepdf.Detalle, Texto: "Del 19/10/2026 al 25/10/2026 · Cerrado"},
	}
	if len(d.Bloques) < 3 || !reflect.DeepEqual(d.Bloques[:3], esperados) {
		t.Errorf("el documento tiene que empezar con %v", esperados)
	}
}

// CA-39.2, RN1 · las métricas son las siete de la página, con los mismos textos
func TestDocumentoSprint_Metricas(t *testing.T) {
	r := reporteDelSprint2()
	d := reportepdf.DocumentoSprint(r)

	verificarTabla(t, tablaDespuesDe(t, d, "Métricas"), []string{"Métrica", "Valor"}, [][]string{
		{"SP planificados y completados", "3 de 8 SP"},
		{"Cumplimiento", r.CumplimientoTexto()},
		{"Historias completadas", r.CompletadasTexto()},
		{"Velocidad del equipo", r.VelocidadTexto()},
		{"Horas estimadas y reales", "6 h estimadas · 3 h reales"},
		{"Desviación", r.DesviacionTexto()},
		{"Defectos", "1 detectado · 1 resuelto"},
	})
}

// CA-39.2 · la tabla de historias tiene las columnas de la página; una historia sin SP dice "Sin estimar"
func TestDocumentoSprint_Historias(t *testing.T) {
	r := reporteDelSprint2()
	d := reportepdf.DocumentoSprint(r)

	verificarTabla(t, tablaDespuesDe(t, d, "Historias del sprint"),
		[]string{"ID", "Título", "SP", "Resultado", "Estimadas", "Reales", "Desviación"}, [][]string{
			{"HU-2", "Registro", "3", "Completada", "4 h", "3 h", r.Historias[0].DesviacionTexto()},
			{"HU-3", "Perfil", "Sin estimar", "No completada", "2 h", "0 h", r.Historias[1].DesviacionTexto()},
		})
}

// CA-39.2 · la tabla de defectos tiene las columnas de la página; un defecto sin historia dice "—"
func TestDocumentoSprint_Defectos(t *testing.T) {
	d := reportepdf.DocumentoSprint(reporteDelSprint2())

	verificarTabla(t, tablaDespuesDe(t, d, "Defectos del sprint"),
		[]string{"Defecto", "Severidad", "Estado", "Historia", "Detectado en", "Resuelto en"}, [][]string{
			{"#2 El registro tarda mucho", "Media", "Resuelto", "HU-2 · Registro", "Sprint 1", "Sprint 2"},
			{"#3 El perfil no carga la foto", "Baja", "Abierto", "—", "Sprint 2", "—"},
		})
}

// RN2 · un sprint abierto lleva el aviso de reporte parcial y el de historias sin estimar, antes de las métricas
func TestDocumentoSprint_Abierto(t *testing.T) {
	r := reporteDelSprint3()
	d := reportepdf.DocumentoSprint(r)

	parcial := buscar(d, reportepdf.Aviso, "El sprint no está cerrado: el reporte muestra cómo va hasta hoy.")
	sinEstimar := buscar(d, reportepdf.Aviso, r.Advertencia())
	metricas := buscar(d, reportepdf.Seccion, "Métricas")
	if r.Advertencia() == "" {
		t.Fatal("el ejemplo tiene que tener una historia sin estimar")
	}
	if parcial < 0 || sinEstimar < 0 || parcial > metricas || sinEstimar > metricas {
		t.Errorf("los avisos tienen que ir antes de las métricas: parcial en %d, sin estimar en %d, métricas en %d", parcial, sinEstimar, metricas)
	}
	if buscar(d, reportepdf.Detalle, "Del 26/10/2026 al 01/11/2026 · Planificado") < 0 {
		t.Error("falta el detalle con las fechas y el estado del sprint abierto")
	}
}

// Sección 6 · un sprint cerrado no lleva avisos
func TestDocumentoSprint_CerradoSinAvisos(t *testing.T) {
	for _, b := range reportepdf.DocumentoSprint(reporteDelSprint2()).Bloques {
		if b.Tipo == reportepdf.Aviso {
			t.Errorf("un sprint cerrado no lleva el aviso %q", b.Texto)
		}
	}
}

// Sección 6 · sin historias ni defectos dice lo mismo que la página
func TestDocumentoSprint_SinHistoriasNiDefectos(t *testing.T) {
	r := reporteDelSprint2()
	r.Historias, r.Defectos = nil, nil
	d := reportepdf.DocumentoSprint(r)

	for _, texto := range []string{"El sprint no tiene historias.", "No hubo defectos en este sprint."} {
		if buscar(d, reportepdf.Parrafo, texto) < 0 {
			t.Errorf("el documento tiene que decir %q", texto)
		}
	}
}

// Sección 3 · el PDF del proyecto arranca con el título, la descripción, las fechas y el estado, y sigue con el resumen
// del panel (HU-38 RN2)
func TestDocumentoProyecto_EncabezadoYResumen(t *testing.T) {
	r := reporteDelProyecto()
	d := reportepdf.DocumentoProyecto(r)

	if d.Titulo != "Reporte del proyecto · Software Metrics" {
		t.Errorf("Titulo = %q", d.Titulo)
	}
	esperados := []reportepdf.Bloque{
		{Tipo: reportepdf.Titulo, Texto: "Reporte del proyecto · Software Metrics"},
		{Tipo: reportepdf.Parrafo, Texto: "TPI de ICS"},
		{Tipo: reportepdf.Detalle, Texto: "Del 05/10/2026 al 30/11/2026 · En curso"},
	}
	if len(d.Bloques) < 3 || !reflect.DeepEqual(d.Bloques[:3], esperados) {
		t.Errorf("el documento tiene que empezar con %v", esperados)
	}
	verificarTabla(t, tablaDespuesDe(t, d, "Resumen del proyecto"), []string{"Métrica", "Valor"}, [][]string{
		{"Sprints", "2 cerrados de 3"},
		{"SP completados", r.SPTexto()},
		{"Velocidad del equipo", "4,0 SP por sprint (2 sprints cerrados)"},
		{"Horas estimadas y reales", "12 h estimadas · 11 h reales"},
		{"Desviación", r.DesviacionTexto()},
		{"Defectos", r.DefectosTexto()},
	})
}

// CA-39.2, HU-38 RN3 · la evolución de la velocidad tiene solo los sprints cerrados
func TestDocumentoProyecto_Evolucion(t *testing.T) {
	d := reportepdf.DocumentoProyecto(reporteDelProyecto())

	verificarTabla(t, tablaDespuesDe(t, d, "Evolución de la velocidad"),
		[]string{"Sprint", "SP planificados", "SP completados", "Cumplimiento", "Velocidad hasta ese sprint"}, [][]string{
			{"Sprint 1", "13", "5", "38,5 %", "5,0 SP por sprint"},
			{"Sprint 2", "8", "3", "37,5 %", "4,0 SP por sprint"},
		})
}

// CA-39.2, HU-38 RN1 · cada sprint va en una hoja nueva, con su título y el mismo cuerpo que el PDF de ese sprint
func TestDocumentoProyecto_UnBloquePorSprint(t *testing.T) {
	r := reporteDelProyecto()
	d := reportepdf.DocumentoProyecto(r)

	for _, s := range r.Sprints {
		titulo := fmt.Sprintf("Sprint %d · %s", s.Sprint.Numero, s.Sprint.Objetivo)
		i := buscar(d, reportepdf.Subtitulo, titulo)
		if i < 1 || d.Bloques[i-1].Tipo != reportepdf.SaltoDeHoja {
			t.Errorf("el bloque %q tiene que empezar en una hoja nueva", titulo)
			continue
		}
		cuerpo := reportepdf.DocumentoSprint(s).Bloques[3:] // sin el título, el objetivo y el detalle
		bloque := d.Bloques[i+2:]
		if len(bloque) < len(cuerpo) || !reflect.DeepEqual(bloque[:len(cuerpo)], cuerpo) {
			t.Errorf("el bloque %q tiene que tener el mismo cuerpo que el PDF de ese sprint", titulo)
		}
		if d.Bloques[i+1].Tipo != reportepdf.Detalle {
			t.Errorf("después del título %q van las fechas y el estado", titulo)
		}
	}
}

// RN5 de HU-38 · un proyecto sin sprints dice lo mismo que la página
func TestDocumentoProyecto_SinSprints(t *testing.T) {
	r := reporteDelProyecto()
	r.Sprints = nil
	r.Resumen.StoryPoints = metrics.StoryPoints{}
	d := reportepdf.DocumentoProyecto(r)

	for _, texto := range []string{"Aún no hay sprints cerrados", "Todavía no hay sprints en este proyecto."} {
		if buscar(d, reportepdf.Parrafo, texto) < 0 {
			t.Errorf("el documento tiene que decir %q", texto)
		}
	}
	if buscar(d, reportepdf.SaltoDeHoja, "") >= 0 {
		t.Error("sin sprints no hay hojas nuevas")
	}
	verificarTabla(t, tablaDespuesDe(t, d, "Resumen del proyecto"), []string{"Métrica", "Valor"}, [][]string{
		{"Sprints", "sin datos"}, {"SP completados", "sin datos"}, {"Velocidad del equipo", r.VelocidadTexto()},
		{"Horas estimadas y reales", r.HorasTexto()}, {"Desviación", r.DesviacionTexto()}, {"Defectos", r.DefectosTexto()},
	})
}

// Textos junta todo lo que se escribe en el PDF, en orden, para compararlo con la página
func TestDocumento_Textos(t *testing.T) {
	d := reportepdf.Documento{Bloques: []reportepdf.Bloque{
		{Tipo: reportepdf.Titulo, Texto: "Título"},
		{Tipo: reportepdf.SaltoDeHoja},
		{Tipo: reportepdf.Tabla, Encabezados: []string{"A", "B"}, Filas: [][]string{{"1", "2"}, {"3", "4"}}, Anchos: []float64{1, 1}},
	}}

	if got, want := d.Textos(), []string{"Título", "A", "B", "1", "2", "3", "4"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Textos() = %q, se esperaba %q", got, want)
	}
}

// CA-39.3 · un proyecto de 12 sprints con 30 historias cada uno se genera en menos de 1 segundo (margen sobre los 10 s
// de Vercel)
func TestEscribir_ProyectoGrandeEnMenosDeUnSegundo(t *testing.T) {
	r := reporteDelProyecto()
	r.Sprints = nil
	for n := 1; n <= 12; n++ {
		s := reporteDelSprint2()
		s.Sprint.Numero = n
		s.Historias = nil
		for h := 1; h <= 30; h++ {
			s.Historias = append(s.Historias, service.FilaReporte{FilaFoto: service.FilaFoto{
				Numero: h, Titulo: fmt.Sprintf("Historia %d con un título largo para que ocupe más de una línea en la tabla del PDF", h),
				StoryPoints: sp(5), Completada: h%2 == 0}})
		}
		r.Sprints = append(r.Sprints, s)
	}

	inicio := time.Now()
	var pdf bytes.Buffer
	if err := reportepdf.Escribir(&pdf, reportepdf.DocumentoProyecto(r), time.Date(2026, 10, 29, 15, 0, 0, 0, time.UTC)); err != nil {
		t.Fatal(err)
	}
	if demora := time.Since(inicio); demora > time.Second {
		t.Errorf("tardó %v, se esperaba menos de 1 segundo", demora)
	}
	if !bytes.HasPrefix(pdf.Bytes(), []byte("%PDF-")) {
		t.Error("lo que se escribió no es un PDF")
	}
}
