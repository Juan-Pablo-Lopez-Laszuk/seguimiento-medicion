// Package migrations tiene los archivos SQL que arman la base de datos, versionados con goose.
//
// Cada archivo se llama NNNN_descripcion.sql y tiene una parte "-- +goose Up" y otra "-- +goose Down".
// Una migración que ya se aplicó nunca se edita: cualquier cambio es una migración nueva (y se actualiza
// docs/modelo-datos.md). Se embeben en el programa para que el comando de migraciones no dependa de dónde se ejecute.
package migrations

import "embed"

// FS contiene todos los archivos .sql de esta carpeta.
//
//go:embed *.sql
var FS embed.FS
