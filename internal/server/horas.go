package server

import (
	"net/http"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// horas agrupa el handler de la página de horas estimadas, reales y desviación de un proyecto (HU-33).
type horas struct {
	casos *service.Horas
}

// ver muestra la desviación del proyecto, de cada sprint y de cada historia.
func (h horas) ver(w http.ResponseWriter, r *http.Request) {
	proyectoID, ok := idDeProyecto(w, r)
	if !ok {
		return
	}
	vista, err := h.casos.Ver(r.Context(), proyectoID)
	if respondeErrorDeTareas(w, err, "ver las horas") {
		return
	}
	render(w, http.StatusOK, "horas.html", Pagina{Titulo: "Horas", Datos: vista})
}
