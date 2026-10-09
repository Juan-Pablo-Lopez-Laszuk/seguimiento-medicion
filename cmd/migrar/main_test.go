package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func sinVariables(string) string { return "" }

func TestRun_SinComandoMuestraLaAyuda(t *testing.T) {
	err := run(context.Background(), nil, sinVariables, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "Uso:") {
		t.Errorf("se esperaba un error con la ayuda y se obtuvo %v", err)
	}
}

func TestRun_ComandoDesconocido(t *testing.T) {
	err := run(context.Background(), []string{"borrar"}, sinVariables, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), `"borrar"`) {
		t.Errorf("se esperaba un error que nombre el comando desconocido y se obtuvo %v", err)
	}
}

// down deshace todo y borra los datos: sin confirmación explícita no se conecta ni a la base.
func TestRun_DownExigeConfirmacionExplicita(t *testing.T) {
	getenv := func(string) string { return "postgres://postgres:x@localhost:1/x" }
	err := run(context.Background(), []string{"down"}, getenv, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "--borrar-todo") {
		t.Errorf("se esperaba que pidiera --borrar-todo y se obtuvo %v", err)
	}
}

func TestRun_SinURLDeBaseDeDatos(t *testing.T) {
	err := run(context.Background(), []string{"up"}, sinVariables, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Errorf("se esperaba un error que pida DATABASE_URL y se obtuvo %v", err)
	}
}

// MIGRATIONS_DATABASE_URL (conexión directa o por session pooler) tiene prioridad sobre DATABASE_URL (pooler de transacciones).
func TestRun_MigrationsDatabaseURLTienePrioridad(t *testing.T) {
	getenv := func(clave string) string {
		if clave == "MIGRATIONS_DATABASE_URL" {
			return "no es una url %zz"
		}
		return "postgres://postgres:x@localhost:1/x"
	}
	err := run(context.Background(), []string{"up"}, getenv, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "URL de la base de datos no es válida") {
		t.Errorf("se esperaba el error de la URL de MIGRATIONS_DATABASE_URL y se obtuvo %v", err)
	}
}

func TestRun_UpStatusYDownContraUnaBase(t *testing.T) {
	base := os.Getenv("TEST_DATABASE_URL")
	if base == "" {
		t.Skip("TEST_DATABASE_URL no está definida: se omiten los tests con base de datos")
	}
	ctx := context.Background()
	esquema := fmt.Sprintf("test_migrar_%d", time.Now().UnixNano())
	separador := "?"
	if strings.Contains(base, "?") {
		separador = "&"
	}
	admin, err := pgxpool.New(ctx, base)
	if err != nil {
		t.Fatalf("conectar a la base de pruebas: %v", err)
	}
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+esquema); err != nil {
		t.Fatalf("crear el esquema de prueba: %v", err)
	}
	t.Cleanup(func() {
		_, _ = admin.Exec(ctx, "DROP SCHEMA "+esquema+" CASCADE")
		admin.Close()
	})
	getenv := func(clave string) string {
		if clave == "DATABASE_URL" {
			return base + separador + "search_path=" + esquema
		}
		return ""
	}

	estado := func() string {
		var salida bytes.Buffer
		if err := run(ctx, []string{"status"}, getenv, &salida); err != nil {
			t.Fatalf("status: %v", err)
		}
		return salida.String()
	}

	if s := estado(); !strings.Contains(s, "0001_modelo_inicial.sql") || !strings.Contains(s, "pendiente") {
		t.Errorf("antes de subir, la 0001 debería figurar como pendiente:\n%s", s)
	}
	if err := run(ctx, []string{"up"}, getenv, &bytes.Buffer{}); err != nil {
		t.Fatalf("up: %v", err)
	}
	if s := estado(); !strings.Contains(s, "aplicada") || strings.Contains(s, "pendiente") {
		t.Errorf("después de subir, la 0001 debería figurar como aplicada:\n%s", s)
	}
	if err := run(ctx, []string{"down", "--borrar-todo"}, getenv, &bytes.Buffer{}); err != nil {
		t.Fatalf("down: %v", err)
	}
	if s := estado(); !strings.Contains(s, "pendiente") {
		t.Errorf("después de bajar, la 0001 debería volver a figurar como pendiente:\n%s", s)
	}
}
