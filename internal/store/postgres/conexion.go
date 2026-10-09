package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Conectar abre un pool de conexiones a PostgreSQL con la URL de DATABASE_URL.
//
// Usa el protocolo simple de pgx porque Supabase conecta por el pooler en modo transacción (puerto 6543),
// que no admite sentencias preparadas. El pool no abre ninguna conexión hasta el primer pedido, así que
// arrancar la aplicación no depende de que la base esté disponible.
func Conectar(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("la URL de la base de datos no es válida: %w", err)
	}
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("crear el pool de conexiones: %w", err)
	}
	return pool, nil
}
