// Command ejemplo carga el proyecto de ejemplo (internal/ejemplo): tres sprints con historias, horas y defectos, para la
// demo y para probar la aplicación a mano.
//
//	go run ./cmd/ejemplo                                    levanta la app en memoria con el ejemplo, en http://localhost:8080
//	go run ./cmd/ejemplo -nombre "Demo"                     otro nombre de proyecto
//	go run ./cmd/ejemplo -emails jp@x.com,m@x.com,c@x.com   los emails de Juan Pablo, Mariano y Carolina
//	go run ./cmd/ejemplo -base                              lo carga en la base de DATABASE_URL y termina
//
// En memoria no hay login y los datos se pierden al cortar el programa. Con -base queda guardado: con los emails de
// las cuentas del equipo, cada uno queda vinculado con su integrante al ingresar.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/app"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/ejemplo"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
)

const ayuda = `Uso: go run ./cmd/ejemplo [-nombre "..."] [-emails a,b,c] [-base]

  -nombre   nombre del proyecto (por defecto "` + ejemplo.NombrePorDefecto + `")
  -emails   tres emails separados por coma: Juan Pablo, Mariano y Carolina
  -base     cargarlo en la base de DATABASE_URL y terminar, en lugar de levantar la app en memoria`

func main() {
	h, err := preparar(os.Args[1:], os.Getenv, os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	if h == nil {
		return
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{Addr: ":" + port, Handler: h, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("Servidor escuchando en http://localhost:%s (en memoria: los datos se pierden al cortarlo)", port)
	log.Fatal(srv.ListenAndServe())
}

// preparar carga el ejemplo según args. En memoria devuelve la aplicación con el ejemplo para servirla; con -base lo
// carga en la base de DATABASE_URL (getenv) y devuelve nil. Escribe en out dónde quedó el proyecto.
func preparar(args []string, getenv func(string) string, out io.Writer) (http.Handler, error) {
	opciones := flag.NewFlagSet("ejemplo", flag.ContinueOnError)
	opciones.SetOutput(io.Discard)
	nombre := opciones.String("nombre", "", "")
	emails := opciones.String("emails", "", "")
	enLaBase := opciones.Bool("base", false, "")
	if err := opciones.Parse(args); err != nil {
		return nil, fmt.Errorf("%w\n\n%s", err, ayuda)
	}
	datos := ejemplo.Datos{Nombre: *nombre}
	if *emails != "" {
		lista := strings.Split(*emails, ",")
		if len(lista) != len(datos.Emails) {
			return nil, fmt.Errorf("-emails lleva tres emails separados por coma (Juan Pablo, Mariano y Carolina)\n\n%s", ayuda)
		}
		for i, email := range lista {
			datos.Emails[i] = strings.TrimSpace(email)
		}
	}

	url := ""
	if *enLaBase {
		if url = getenv("DATABASE_URL"); url == "" {
			return nil, errors.New("-base necesita la variable DATABASE_URL con la base donde cargar el ejemplo")
		}
	}
	dep, err := app.DependenciasConBase(url)
	if err != nil {
		return nil, err
	}
	p, err := ejemplo.Cargar(context.Background(), dep, datos)
	if err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintf(out, "Proyecto de ejemplo %q cargado en /proyectos/%d\n", p.Nombre, p.ID); err != nil {
		return nil, err
	}
	if *enLaBase {
		return nil, nil
	}
	return server.NewRouter(dep), nil
}
