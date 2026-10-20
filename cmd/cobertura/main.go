// Command cobertura hace cumplir el umbral de cobertura del equipo (TEC-06): lee el perfil que escribe
// `go test -coverprofile`, muestra la cobertura de internal/domain e internal/metrics y falla si la combinada baja del
// mínimo. El CI la corre después de los tests.
//
//	go test -covermode=atomic -coverprofile=coverage.out ./...
//	go run ./cmd/cobertura                           perfil coverage.out, mínimo 80 %
//	go run ./cmd/cobertura -umbral 90                otro mínimo
//	go run ./cmd/cobertura internal/domain           otros paquetes (con sus subpaquetes)
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/cobertura"
)

const perfilPorDefecto = "coverage.out"

const ayuda = `Uso: go run ./cmd/cobertura [-perfil coverage.out] [-umbral 80] [paquete ...]

  -perfil   perfil de cobertura que escribió go test -coverprofile (por defecto coverage.out)
  -umbral   cobertura combinada mínima, de 0 a 100 (por defecto 80)
  paquete   ruta relativa al módulo, con sus subpaquetes (por defecto internal/domain e internal/metrics)`

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

// run lee el perfil, escribe el informe en out y devuelve error si la cobertura no llega al mínimo. out se recibe para
// poder probarlo.
func run(args []string, out io.Writer) error {
	opciones := flag.NewFlagSet("cobertura", flag.ContinueOnError)
	opciones.SetOutput(io.Discard)
	ruta := opciones.String("perfil", perfilPorDefecto, "")
	umbral := opciones.Float64("umbral", cobertura.MinimoPorDefecto, "")
	if err := opciones.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			_, err := fmt.Fprintln(out, ayuda)
			return err
		}
		return fmt.Errorf("%w\n\n%s", err, ayuda)
	}
	paquetes := opciones.Args()
	if len(paquetes) == 0 {
		paquetes = cobertura.PaquetesPorDefecto()
	}

	perfil, err := os.Open(*ruta)
	if err != nil {
		return fmt.Errorf("no se pudo abrir el perfil de cobertura %q: %w", *ruta, err)
	}
	defer func() { _ = perfil.Close() }()

	medicion, err := cobertura.Medir(perfil, paquetes)
	if err != nil {
		return err
	}
	// El informe se muestra aunque no se llegue al mínimo: es lo que el equipo necesita para ver qué bajó.
	if _, err := io.WriteString(out, cobertura.Informe(medicion, *umbral)); err != nil {
		return err
	}
	return cobertura.Verificar(medicion, *umbral)
}
