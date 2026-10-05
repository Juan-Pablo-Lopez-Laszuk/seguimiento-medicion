// Command server levanta la aplicación en la computadora local: go run ./cmd/server
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           server.NewRouter(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("Servidor escuchando en http://localhost:%s", port)
	log.Fatal(srv.ListenAndServe())
}
