package server

import (
	"fmt"
	"net/http"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// reportes agrupa el handler del reporte de un sprint (HU-37).
type reportes struct {
	casos *service.Reportes
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
