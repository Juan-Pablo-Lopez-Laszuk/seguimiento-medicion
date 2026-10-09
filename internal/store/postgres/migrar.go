package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/migrations"
)

// Subir aplica todas las migraciones que todavía no se aplicaron. Si ya están todas, no hace nada.
func Subir(ctx context.Context, pool *pgxpool.Pool) error {
	return conMigrador(pool, func(p *goose.Provider) error {
		if _, err := p.Up(ctx); err != nil {
			return fmt.Errorf("aplicar las migraciones: %w", err)
		}
		return nil
	})
}

// Bajar deshace todas las migraciones aplicadas, de la última a la primera. Borra los datos: solo para tests
// y para bases de desarrollo.
func Bajar(ctx context.Context, pool *pgxpool.Pool) error {
	return conMigrador(pool, func(p *goose.Provider) error {
		if _, err := p.DownTo(ctx, 0); err != nil {
			return fmt.Errorf("deshacer las migraciones: %w", err)
		}
		return nil
	})
}

// Estado devuelve cada migración con su fecha de aplicación (la fecha es cero si todavía no se aplicó).
func Estado(ctx context.Context, pool *pgxpool.Pool) ([]*goose.MigrationStatus, error) {
	var estado []*goose.MigrationStatus
	err := conMigrador(pool, func(p *goose.Provider) error {
		var err error
		estado, err = p.Status(ctx)
		if err != nil {
			return fmt.Errorf("consultar el estado de las migraciones: %w", err)
		}
		return nil
	})
	return estado, err
}

// conMigrador arma el migrador de goose sobre el pool, ejecuta f y libera lo que abrió (no cierra el pool).
func conMigrador(pool *pgxpool.Pool, f func(*goose.Provider) error) error {
	db := stdlib.OpenDBFromPool(pool)
	defer closeQuietly(db)

	p, err := goose.NewProvider(goose.DialectPostgres, db, migrations.FS)
	if err != nil {
		return fmt.Errorf("preparar las migraciones: %w", err)
	}
	return f(p)
}

func closeQuietly(db *sql.DB) { _ = db.Close() }
