package server

import (
	"encoding/json"
	"net/http"
)

// health indica que la aplicación está viva. Lo usan Vercel y el CI para verificar el deploy.
func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
