// Package reportepdf pasa a PDF el reporte de un sprint (HU-37) y el del proyecto (HU-38) (spec:
// specs/HU-39-exportar-pdf.md). Es presentación, como las plantillas HTML: arma un Documento con los mismos datos y
// los mismos textos que la página (RN1) y lo dibuja con fpdf.
package reportepdf

import (
	"fmt"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// Tipo es la clase de un bloque del documento.
type Tipo int

const (
	Titulo      Tipo = iota // el título del reporte
	Subtitulo               // el título del bloque de un sprint en el reporte del proyecto
	Seccion                 // "Métricas", "Historias del sprint", ...
	Parrafo                 // texto común
	Detalle                 // texto chico y gris: las fechas y el estado
	Aviso                   // el aviso de reporte parcial o de historias sin estimar
	Tabla                   // una tabla con encabezados
	SaltoDeHoja             // lo que sigue arranca en una hoja nueva
)

// Bloque es una parte del documento. Las tablas usan Encabezados, Filas y Anchos (proporciones del ancho de cada
// columna); el resto usa Texto.
type Bloque struct {
	Tipo        Tipo
	Texto       string
	Encabezados []string
	Filas       [][]string
	Anchos      []float64
}

// Documento es lo que se escribe en el PDF, en orden. Titulo va en las propiedades del archivo.
type Documento struct {
	Titulo  string
	Bloques []Bloque
}

// Textos devuelve todos los textos del documento en orden: los de los bloques, los encabezados y las celdas.
func (d Documento) Textos() []string {
	var textos []string
	for _, b := range d.Bloques {
		if b.Texto != "" {
			textos = append(textos, b.Texto)
		}
		textos = append(textos, b.Encabezados...)
		for _, fila := range b.Filas {
			textos = append(textos, fila...)
		}
	}
	return textos
}

const (
	avisoParcial = "El sprint no está cerrado: el reporte muestra cómo va hasta hoy."
	formatoFecha = "02/01/2006"
)

// DocumentoSprint arma el PDF del reporte de un sprint, con el encabezado y el cuerpo de reporte_sprint.html.
func DocumentoSprint(r service.ReporteSprint) Documento {
	titulo := fmt.Sprintf("Reporte del Sprint %d · %s", r.Sprint.Numero, r.Proyecto.Nombre)
	bloques := []Bloque{
		{Tipo: Titulo, Texto: titulo},
		{Tipo: Parrafo, Texto: r.Sprint.Objetivo},
		{Tipo: Detalle, Texto: fechasSprint(r)},
	}
	return Documento{Titulo: titulo, Bloques: append(bloques, cuerpoSprint(r)...)}
}

// DocumentoProyecto arma el PDF del reporte del proyecto, como reporte_proyecto.html: el resumen, la evolución de la
// velocidad y un bloque por sprint, cada uno en una hoja nueva.
func DocumentoProyecto(r service.ReporteProyecto) Documento {
	p := r.Resumen.Proyecto
	titulo := "Reporte del proyecto · " + p.Nombre
	bloques := []Bloque{{Tipo: Titulo, Texto: titulo}}
	if p.Descripcion != "" {
		bloques = append(bloques, Bloque{Tipo: Parrafo, Texto: p.Descripcion})
	}
	detalle := fmt.Sprintf("Del %s al %s", p.FechaInicio.Format(formatoFecha), p.FechaFin.Format(formatoFecha))
	if r.Resumen.Estado != "" {
		detalle += " · " + string(r.Resumen.Estado)
	}
	bloques = append(bloques,
		Bloque{Tipo: Detalle, Texto: detalle},
		Bloque{Tipo: Seccion, Texto: "Resumen del proyecto"},
		metricas([][]string{
			{"Sprints", r.SprintsTexto()},
			{"SP completados", r.SPTexto()},
			{"Velocidad del equipo", r.VelocidadTexto()},
			{"Horas estimadas y reales", r.HorasTexto()},
			{"Desviación", r.DesviacionTexto()},
			{"Defectos", r.DefectosTexto()},
		}),
		Bloque{Tipo: Seccion, Texto: "Evolución de la velocidad"},
	)
	bloques = append(bloques, evolucion(r.Evolucion()))
	if len(r.Sprints) == 0 {
		bloques = append(bloques, Bloque{Tipo: Parrafo, Texto: "Todavía no hay sprints en este proyecto."})
	}
	for _, s := range r.Sprints {
		bloques = append(bloques,
			Bloque{Tipo: SaltoDeHoja},
			Bloque{Tipo: Subtitulo, Texto: fmt.Sprintf("Sprint %d · %s", s.Sprint.Numero, s.Sprint.Objetivo)},
			Bloque{Tipo: Detalle, Texto: fechasSprint(s)},
		)
		bloques = append(bloques, cuerpoSprint(s)...)
	}
	return Documento{Titulo: titulo, Bloques: bloques}
}

func fechasSprint(r service.ReporteSprint) string {
	return fmt.Sprintf("Del %s al %s · %s", r.Sprint.FechaInicio.Format(formatoFecha), r.Sprint.FechaFin.Format(formatoFecha), r.Sprint.Estado)
}

// cuerpoSprint es el bloque "reporteSprint" de componentes.html: los avisos, las métricas, las historias y los
// defectos.
func cuerpoSprint(r service.ReporteSprint) []Bloque {
	var bloques []Bloque
	if !r.Cerrado() {
		bloques = append(bloques, Bloque{Tipo: Aviso, Texto: avisoParcial})
	}
	if aviso := r.Advertencia(); aviso != "" {
		bloques = append(bloques, Bloque{Tipo: Aviso, Texto: aviso})
	}
	bloques = append(bloques,
		Bloque{Tipo: Seccion, Texto: "Métricas"},
		metricas([][]string{
			{"SP planificados y completados", r.SPTexto()},
			{"Cumplimiento", r.CumplimientoTexto()},
			{"Historias completadas", r.CompletadasTexto()},
			{"Velocidad del equipo", r.VelocidadTexto()},
			{"Horas estimadas y reales", r.HorasTexto()},
			{"Desviación", r.DesviacionTexto()},
			{"Defectos", r.DefectosTexto()},
		}),
		Bloque{Tipo: Seccion, Texto: "Historias del sprint"},
		historias(r.Historias),
		Bloque{Tipo: Seccion, Texto: "Defectos del sprint"},
		defectos(r.Defectos),
	)
	return bloques
}

// metricas es la tabla de dos columnas que en la página son tarjetas.
func metricas(filas [][]string) Bloque {
	return Bloque{Tipo: Tabla, Encabezados: []string{"Métrica", "Valor"}, Filas: filas, Anchos: []float64{2, 3}}
}

func evolucion(cerrados []service.ReporteSprint) Bloque {
	if len(cerrados) == 0 {
		return Bloque{Tipo: Parrafo, Texto: "Aún no hay sprints cerrados"}
	}
	filas := make([][]string, 0, len(cerrados))
	for _, s := range cerrados {
		// fmt.Sprint escribe los SP igual que la plantilla ({{.SP.Planificados}})
		filas = append(filas, []string{fmt.Sprintf("Sprint %d", s.Sprint.Numero), fmt.Sprint(s.SP.Planificados),
			fmt.Sprint(s.SP.Completados), s.CumplimientoTexto(), s.VelocidadPorSprint()})
	}
	return Bloque{Tipo: Tabla, Encabezados: []string{"Sprint", "SP planificados", "SP completados", "Cumplimiento", "Velocidad hasta ese sprint"},
		Filas: filas, Anchos: []float64{2, 2, 2, 2, 3}}
}

func historias(filas []service.FilaReporte) Bloque {
	if len(filas) == 0 {
		return Bloque{Tipo: Parrafo, Texto: "El sprint no tiene historias."}
	}
	tabla := make([][]string, 0, len(filas))
	for _, h := range filas {
		sp := "Sin estimar"
		if h.StoryPoints != nil {
			sp = fmt.Sprint(*h.StoryPoints)
		}
		tabla = append(tabla, []string{h.Identificador(), h.Titulo, sp, h.Resultado(), h.EstimadasTexto(), h.RealesTexto(), h.DesviacionTexto()})
	}
	return Bloque{Tipo: Tabla, Encabezados: []string{"ID", "Título", "SP", "Resultado", "Estimadas", "Reales", "Desviación"},
		Filas: tabla, Anchos: []float64{1.3, 4.5, 1.5, 2.3, 1.7, 1.5, 2.6}}
}

func defectos(filas []service.FilaDefecto) Bloque {
	if len(filas) == 0 {
		return Bloque{Tipo: Parrafo, Texto: "No hubo defectos en este sprint."}
	}
	tabla := make([][]string, 0, len(filas))
	for _, d := range filas {
		historia := d.Historia
		if historia == "" {
			historia = "—"
		}
		tabla = append(tabla, []string{fmt.Sprintf("#%d %s", d.ID, d.Descripcion), string(d.Severidad), string(d.Estado), historia, d.DetectadoEn, d.ResueltoEn})
	}
	return Bloque{Tipo: Tabla, Encabezados: []string{"Defecto", "Severidad", "Estado", "Historia", "Detectado en", "Resuelto en"},
		Filas: tabla, Anchos: []float64{4.5, 1.8, 1.8, 2.8, 1.8, 1.8}}
}
