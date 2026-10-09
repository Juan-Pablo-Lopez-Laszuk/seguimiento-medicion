package postgres_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/postgres"
)

// nuevaBase devuelve una conexión a un esquema vacío y exclusivo de este test, que se borra al terminar.
// Así los tests no se pisan entre sí y no dejan tablas sueltas. Necesita TEST_DATABASE_URL (una base
// PostgreSQL descartable, por ejemplo la de Docker de CONTRIBUTING.md); si no está, el test se omite.
func nuevaBase(t *testing.T) *pgxpool.Pool {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL no está definida: se omiten los tests con base de datos")
	}
	ctx := context.Background()
	esquema := fmt.Sprintf("test_%d", time.Now().UnixNano())

	admin, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatalf("conectar a la base de pruebas: %v", err)
	}
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+esquema); err != nil {
		admin.Close()
		t.Fatalf("crear el esquema de prueba: %v", err)
	}

	separador := "?"
	if strings.Contains(url, "?") {
		separador = "&"
	}
	pool, err := postgres.Conectar(ctx, url+separador+"search_path="+esquema)
	if err != nil {
		t.Fatalf("conectar al esquema de prueba: %v", err)
	}
	t.Cleanup(func() {
		pool.Close()
		_, _ = admin.Exec(ctx, "DROP SCHEMA "+esquema+" CASCADE")
		admin.Close()
	})
	return pool
}

// baseMigrada es una nuevaBase con todas las migraciones aplicadas.
func baseMigrada(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool := nuevaBase(t)
	if err := postgres.Subir(context.Background(), pool); err != nil {
		t.Fatalf("aplicar las migraciones: %v", err)
	}
	return pool
}
