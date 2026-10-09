// Command migrar aplica las migraciones de migrations/ a una base PostgreSQL (la de Supabase o la local de Docker).
//
//	go run ./cmd/migrar up                    aplica las migraciones que faltan
//	go run ./cmd/migrar status                muestra cuáles están aplicadas
//	go run ./cmd/migrar down --borrar-todo    las deshace todas (borra los datos)
//
// La base sale de MIGRATIONS_DATABASE_URL o, si no existe, de DATABASE_URL.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/pressly/goose/v3"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/postgres"
)

const ayuda = `Uso: go run ./cmd/migrar <comando>

  up                  aplica las migraciones que faltan
  status              muestra cuáles están aplicadas
  down --borrar-todo  deshace todas las migraciones (borra los datos)

La base sale de MIGRATIONS_DATABASE_URL o, si no existe, de DATABASE_URL.`

func main() {
	if err := run(context.Background(), os.Args[1:], os.Getenv, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run ejecuta el comando contra la base. getenv y out se reciben para poder probarlo.
func run(ctx context.Context, args []string, getenv func(string) string, out io.Writer) error {
	if len(args) == 0 {
		return errors.New(ayuda)
	}
	comando := args[0]
	if comando != "up" && comando != "status" && comando != "down" {
		return fmt.Errorf("comando desconocido %q\n\n%s", comando, ayuda)
	}
	// down borra los datos: se exige la confirmación antes de siquiera conectarse.
	if comando == "down" && (len(args) < 2 || args[1] != "--borrar-todo") {
		return errors.New("down deshace todas las migraciones y borra los datos: agregá --borrar-todo para confirmar")
	}

	url := getenv("MIGRATIONS_DATABASE_URL")
	if url == "" {
		url = getenv("DATABASE_URL")
	}
	if url == "" {
		return errors.New("falta la base de datos: definí MIGRATIONS_DATABASE_URL o DATABASE_URL")
	}
	pool, err := postgres.Conectar(ctx, url)
	if err != nil {
		return err
	}
	defer pool.Close()

	switch comando {
	case "up":
		if err := postgres.Subir(ctx, pool); err != nil {
			return err
		}
		fmt.Fprintln(out, "migraciones aplicadas")
	case "down":
		if err := postgres.Bajar(ctx, pool); err != nil {
			return err
		}
		fmt.Fprintln(out, "migraciones deshechas")
	case "status":
		estado, err := postgres.Estado(ctx, pool)
		if err != nil {
			return err
		}
		for _, m := range estado {
			if m.State == goose.StateApplied {
				fmt.Fprintf(out, "aplicada   %s  %s\n", m.AppliedAt.Format("2006-01-02 15:04"), m.Source.Path)
			} else {
				fmt.Fprintf(out, "pendiente  %17s  %s\n", "", m.Source.Path)
			}
		}
	}
	return nil
}
