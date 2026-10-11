package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// VerificadorBase es lo que /health necesita de la base de datos: saber si responde. El pool de pgx lo
// cumple con su método Ping, y en los tests se usa una base falsa (spec TEC-04).
type VerificadorBase interface {
	Ping(ctx context.Context) error
}

// EsperaBase es lo máximo que /health espera a la base antes de darla por caída (RN3).
const EsperaBase = 3 * time.Second

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
		status, codigo := estadoGeneral(estadoBase)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(codigo)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": status, "base": estadoBase})
	}
}

// verificarBase le hace ping a la base, con un tiempo máximo (RN3), y devuelve su estado. El error solo va al
// log del servidor: la respuesta no lo muestra porque puede traer el host o el usuario (RN4).
func verificarBase(ctx context.Context, base VerificadorBase) string {
	if base == nil {
		return BaseMemoria
	}
	ctx, cancelar := context.WithTimeout(ctx, EsperaBase)
	defer cancelar()
	if err := base.Ping(ctx); err != nil {
		log.Printf("health: la base de datos no responde: %v", err)
		return BaseSinConexion
	}
	return BaseOK
}

// estadoGeneral traduce el estado de la base al estado de la app: sin conexión la app está viva pero no puede
// guardar, por eso responde 503 (RN2). En memoria o con la base ok, responde 200 (RN1).
func estadoGeneral(estadoBase string) (status string, codigo int) {
	if estadoBase == BaseSinConexion {
		return "error", http.StatusServiceUnavailable
	}
	return "ok", http.StatusOK
}
