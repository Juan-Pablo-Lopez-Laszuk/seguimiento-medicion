package server

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/reportepdf"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// reportes agrupa los handlers del reporte de un sprint (HU-37), del reporte del proyecto (HU-38) y de sus PDF (HU-39).
type reportes struct {
	casos *service.Reportes
	ahora func() time.Time // la fecha del nombre y del pie del PDF
}

// sprint muestra el reporte del sprint para la Sprint Review.
func (h reportes) sprint(w http.ResponseWriter, r *http.Request) {
	proyectoID, numeroSprint, ok := idsDeSprint(w, r)
	if !ok {
		return
	}
	reporte, err := h.casos.Sprint(r.Context(), proyectoID, numeroSprint)
	if respondeErrorDeSprint(w, err, "ver el reporte del sprint") {
		return
	}
	render(w, http.StatusOK, "reporte_sprint.html", Pagina{Titulo: fmt.Sprintf("Reporte del Sprint %d", numeroSprint), Datos: reporte})
}

// proyecto muestra el reporte del proyecto para el cliente (HU-38).
func (h reportes) proyecto(w http.ResponseWriter, r *http.Request) {
	proyectoID, ok := idDeProyecto(w, r)
	if !ok {
		return
	}
	reporte, err := h.casos.Proyecto(r.Context(), proyectoID)
	if respondeError(w, err, "ver el reporte del proyecto") {
		return
	}
	render(w, http.StatusOK, "reporte_proyecto.html", Pagina{Titulo: "Reporte del proyecto", Datos: reporte})
}

// sprintPDF descarga el reporte del sprint como PDF (HU-39).
func (h reportes) sprintPDF(w http.ResponseWriter, r *http.Request) {
	proyectoID, numeroSprint, ok := idsDeSprint(w, r)
	if !ok {
		return
	}
	reporte, err := h.casos.Sprint(r.Context(), proyectoID, numeroSprint)
	if respondeErrorDeSprint(w, err, "descargar el PDF del reporte del sprint") {
		return
	}
	h.pdf(w, reportepdf.DocumentoSprint(reporte), reporte.Proyecto.Nombre, numeroSprint)
}

// proyectoPDF descarga el reporte del proyecto como PDF (HU-39).
func (h reportes) proyectoPDF(w http.ResponseWriter, r *http.Request) {
	proyectoID, ok := idDeProyecto(w, r)
	if !ok {
		return
	}
	reporte, err := h.casos.Proyecto(r.Context(), proyectoID)
	if respondeError(w, err, "descargar el PDF del reporte del proyecto") {
		return
	}
	h.pdf(w, reportepdf.DocumentoProyecto(reporte), reporte.Resumen.Proyecto.Nombre, 0)
}

// pdf genera el PDF en memoria y lo manda como archivo para descargar (CA-39.1). Se genera antes de escribir la
// respuesta, así una falla todavía puede responder 500.
func (h reportes) pdf(w http.ResponseWriter, documento reportepdf.Documento, proyecto string, sprint int) {
	ahora := h.ahora()
	var pdf bytes.Buffer
	if err := reportepdf.Escribir(&pdf, documento, ahora); err != nil {
		log.Printf("generar el PDF del reporte: %v", err)
		http.Error(w, "No se pudo completar la operación. Probá de nuevo en unos minutos.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, reportepdf.NombreArchivo(proyecto, sprint, ahora)))
	if _, err := w.Write(pdf.Bytes()); err != nil {
		log.Printf("enviar el PDF del reporte: %v", err)
	}
}
