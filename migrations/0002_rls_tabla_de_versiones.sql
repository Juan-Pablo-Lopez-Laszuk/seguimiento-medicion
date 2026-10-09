-- Seguridad: Supabase le da al rol anon permisos sobre todo el esquema public. goose_db_version (donde goose
-- anota qué migraciones se aplicaron) no está en el modelo, así que 0001 no la cubría: sin RLS, la API pública
-- podría leerla y modificarla. Con RLS y sin políticas queda cerrada, igual que las tablas del modelo.

-- +goose Up

ALTER TABLE goose_db_version ENABLE ROW LEVEL SECURITY;

-- +goose Down

ALTER TABLE goose_db_version DISABLE ROW LEVEL SECURITY;
