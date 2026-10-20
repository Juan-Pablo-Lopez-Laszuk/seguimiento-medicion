package server

import (
	"net/http"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// panel agrupa el handler del panel de indicadores de un proyecto (HU-35).
type panel struct {
	casos *service.Panel
}

// ver muestra las seis tarjetas del panel.
func (h panel) ver(w http.ResponseWriter, r *http.Request) {
	proyectoID, ok := idDeProyecto(w, r)
	if !ok {
		return
	}
	vista, err := h.casos.Ver(r.Context(), proyectoID)
	if respondeErrorDeTareas(w, err, "ver el panel") {
		return
	}
	render(w, http.StatusOK, "panel.html", Pagina{Titulo: "Panel", Datos: vista})
}
