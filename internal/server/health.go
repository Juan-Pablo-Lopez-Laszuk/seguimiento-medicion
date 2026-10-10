package server

import (
	"context"
	"encoding/json"
	"net/http"
)

// VerificadorBase es lo que /health necesita de la base de datos: saber si responde. El pool de pgx lo
// cumple con su método Ping, y en los tests se usa una base falsa (spec TEC-04).
type VerificadorBase interface {
	Ping(ctx context.Context) error
}

// Estados de la base que informa /health.
const (
	BaseMemoria = "memoria"
)

// health indica si la aplicación está viva y dónde guarda los datos. Lo usan Vercel, el CI y el equipo
// para verificar el deploy (spec TEC-04).
func health(base VerificadorBase) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok", "base": BaseMemoria})
	}
}
