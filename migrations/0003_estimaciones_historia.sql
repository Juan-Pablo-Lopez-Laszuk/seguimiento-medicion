-- HU-16b · Historial de estimaciones (specs/HU-16b-historial-estimaciones.md).
-- Cada fila es un cambio de la estimación de una historia: el valor que tenía, el que quedó y cuándo cambió.
-- La aplicación solo agrega filas (no las edita ni las borra); se borran junto con la historia (HU-06), porque la
-- base no hace borrados en cascada. No guarda quién estimó: todavía no hay login, y cuando llegue TEC-05 se le suma
-- la columna. Una historia estimada antes de esta migración empieza con el historial vacío.

-- +goose Up

CREATE TABLE estimaciones_historia (
    id             bigserial   PRIMARY KEY,
    historia_id    bigint      NOT NULL REFERENCES historias (id),
    valor_anterior int         NOT NULL,
    valor_nuevo    int         NOT NULL,
    cambiada_en    timestamptz NOT NULL,
    CONSTRAINT estimaciones_historia_anterior_ck  CHECK (valor_anterior IN (0, 1, 2, 3, 5, 8, 13, 21)),
    CONSTRAINT estimaciones_historia_nuevo_ck     CHECK (valor_nuevo IN (0, 1, 2, 3, 5, 8, 13, 21)),
    CONSTRAINT estimaciones_historia_distintos_ck CHECK (valor_anterior <> valor_nuevo)
);

CREATE INDEX estimaciones_historia_historia_idx ON estimaciones_historia (historia_id);

-- Seguridad: igual que las demás tablas (ver 0001). Supabase publica cada tabla de "public" por su API; con RLS
-- activado y sin políticas esa API no puede leer ni escribir nada, y la app entra por DATABASE_URL con el usuario
-- postgres, que no está sujeto a RLS.
ALTER TABLE estimaciones_historia ENABLE ROW LEVEL SECURITY;

-- +goose Down

DROP TABLE estimaciones_historia;
