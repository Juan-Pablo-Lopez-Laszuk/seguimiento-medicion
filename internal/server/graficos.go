package server

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// graficos agrupa los handlers de los gráficos de métricas de un proyecto (HU-36).
type graficos struct {
	casos *service.Graficos
}

// datos responde en JSON los datos de los cinco gráficos (CA-36.2). ?sprint= elige el sprint del burndown; un valor
// que no es un número es como no elegir ninguno (RN5).
func (h graficos) datos(w http.ResponseWriter, r *http.Request) {
	proyectoID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		responderJSON(w, http.StatusNotFound, map[string]string{"error": "No encontramos ese proyecto."})
		return
	}
	numeroSprint, _ := strconv.Atoi(r.URL.Query().Get("sprint"))
	d, err := h.casos.Datos(r.Context(), proyectoID, numeroSprint)
	switch {
	case errors.Is(err, project.ErrNoEncontrado):
		responderJSON(w, http.StatusNotFound, map[string]string{"error": "No encontramos ese proyecto."})
	case err != nil:
		log.Printf("datos de los gráficos: %v", err)
		responderJSON(w, http.StatusInternalServerError, map[string]string{"error": "No se pudieron armar los gráficos."})
	default:
		responderJSON(w, http.StatusOK, d)
	}
}

// responderJSON escribe v en JSON con el código indicado.
func responderJSON(w http.ResponseWriter, codigo int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codigo)
	_ = json.NewEncoder(w).Encode(v)
}
