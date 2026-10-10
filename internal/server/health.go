package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
)

// VerificadorBase es lo que /health necesita de la base de datos: saber si responde. El pool de pgx lo
// cumple con su método Ping, y en los tests se usa una base falsa (spec TEC-04).
type VerificadorBase interface {
	Ping(ctx context.Context) error
}

// Estados de la base que informa /health.
const (
	BaseMemoria     = "memoria"
	BaseOK          = "ok"
	BaseSinConexion = "sin conexión"
)

// health indica si la aplicación está viva y si la base de datos responde. Lo usan Vercel, el CI y el equipo
// para verificar el deploy (spec TEC-04).
func health(base VerificadorBase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		estadoBase := verificarBase(r.Context(), base)
		status, codigo := "ok", http.StatusOK
		if estadoBase == BaseSinConexion {
			status, codigo = "error", http.StatusServiceUnavailable
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(codigo)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": status, "base": estadoBase})
	}
}

// verificarBase le hace ping a la base y devuelve su estado. El error solo va al log del servidor: la
// respuesta no lo muestra porque puede traer el host o el usuario (RN4).
func verificarBase(ctx context.Context, base VerificadorBase) string {
	if base == nil {
		return BaseMemoria
	}
	if err := base.Ping(ctx); err != nil {
		log.Printf("health: la base de datos no responde: %v", err)
		return BaseSinConexion
	}
	return BaseOK
}
