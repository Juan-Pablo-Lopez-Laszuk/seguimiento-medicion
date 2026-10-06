// Command server levanta la aplicación en la computadora local: go run ./cmd/server
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/app"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           app.NewHandler(),
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("Servidor escuchando en http://localhost:%s", port)
	log.Fatal(srv.ListenAndServe())
}
