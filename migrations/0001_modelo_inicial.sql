-- Modelo inicial (TEC-03). Sale de docs/modelo-datos.md: si algo cambia, se agrega una migración nueva
-- y se actualiza ese documento. Nombres de tablas en plural y en español, columnas en snake_case.

-- +goose Up

CREATE TABLE proyectos (
    id           bigserial PRIMARY KEY,
    nombre       text        NOT NULL,
    descripcion  text        NOT NULL DEFAULT '',
    fecha_inicio date        NOT NULL,
    fecha_fin    date        NOT NULL,
    estado       text        NOT NULL,
    creado_en    timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT proyectos_fechas_ck CHECK (fecha_fin > fecha_inicio)
);
-- El nombre no se repite sin distinguir mayúsculas (HU-01).
CREATE UNIQUE INDEX proyectos_nombre_uq ON proyectos (lower(nombre));

CREATE TABLE integrantes (
    id           bigserial PRIMARY KEY,
    proyecto_id  bigint  NOT NULL REFERENCES proyectos (id),
    nombre       text    NOT NULL,
    email        text    NOT NULL,
    rol          text    NOT NULL,
    activo       boolean NOT NULL DEFAULT true,
    auth_user_id uuid,
    CONSTRAINT integrantes_email_uq UNIQUE (proyecto_id, email)
);
-- Un solo Agile Enabler ACTIVO por proyecto: si se da de baja, se puede registrar otro (HU-03, RN6).
CREATE UNIQUE INDEX integrantes_un_agile_enabler_uq ON integrantes (proyecto_id) WHERE rol = 'AgileEnabler' AND activo;

CREATE TABLE sprints (
    id              bigserial PRIMARY KEY,
    proyecto_id     bigint NOT NULL REFERENCES proyectos (id),
    numero          int    NOT NULL,
    objetivo        text   NOT NULL,
    fecha_inicio    date   NOT NULL,
    fecha_fin       date   NOT NULL,
    estado          text   NOT NULL,
    sp_planificados int,
    sp_completados  int,
    CONSTRAINT sprints_numero_uq UNIQUE (proyecto_id, numero)
);

CREATE TABLE historias (
    id              bigserial PRIMARY KEY,
    proyecto_id     bigint  NOT NULL REFERENCES proyectos (id),
    numero          int     NOT NULL,
    titulo          text    NOT NULL,
    descripcion     text    NOT NULL DEFAULT '',
    prioridad       text    NOT NULL,
    estado          text    NOT NULL DEFAULT 'Pendiente',
    story_points    int,
    orden           int     NOT NULL DEFAULT 0,
    sprint_id       bigint  REFERENCES sprints (id),
    horas_estimadas numeric,
    completada_en   date,
    CONSTRAINT historias_numero_uq UNIQUE (proyecto_id, numero),
    CONSTRAINT historias_prioridad_ck CHECK (prioridad IN ('Alta', 'Media', 'Baja')),
    CONSTRAINT historias_estado_ck CHECK (estado IN ('Pendiente', 'En sprint', 'En progreso', 'Hecho')),
    CONSTRAINT historias_story_points_ck CHECK (story_points IN (0, 1, 2, 3, 5, 8, 13, 21)),
    CONSTRAINT historias_horas_ck CHECK (horas_estimadas > 0)
);

CREATE TABLE criterios_aceptacion (
    id          bigserial PRIMARY KEY,
    historia_id bigint  NOT NULL REFERENCES historias (id),
    descripcion text    NOT NULL,
    cumplido    boolean NOT NULL DEFAULT false
);

CREATE TABLE tareas (
    id              bigserial PRIMARY KEY,
    historia_id     bigint  NOT NULL REFERENCES historias (id),
    titulo          text    NOT NULL,
    responsable_id  bigint  REFERENCES integrantes (id),
    horas_estimadas numeric,
    completada      boolean NOT NULL DEFAULT false,
    CONSTRAINT tareas_horas_ck CHECK (horas_estimadas > 0)
);

-- Foto de cierre: qué historias participaron de cada sprint cerrado y con qué SP.
CREATE TABLE sprint_historias (
    sprint_id    bigint  NOT NULL REFERENCES sprints (id),
    historia_id  bigint  NOT NULL REFERENCES historias (id),
    story_points int,
    completada   boolean NOT NULL DEFAULT false,
    PRIMARY KEY (sprint_id, historia_id),
    CONSTRAINT sprint_historias_story_points_ck CHECK (story_points IN (0, 1, 2, 3, 5, 8, 13, 21))
);

CREATE TABLE sesiones_poker (
    id             bigserial PRIMARY KEY,
    historia_id    bigint NOT NULL REFERENCES historias (id),
    estado         text   NOT NULL,
    valor_acordado int,
    CONSTRAINT sesiones_poker_valor_ck CHECK (valor_acordado IN (0, 1, 2, 3, 5, 8, 13, 21))
);
-- Una única sesión abierta por historia (CA-17.2).
CREATE UNIQUE INDEX sesiones_poker_una_abierta_uq ON sesiones_poker (historia_id) WHERE estado = 'Abierta';

CREATE TABLE rondas_poker (
    id        bigserial PRIMARY KEY,
    sesion_id bigint  NOT NULL REFERENCES sesiones_poker (id),
    numero    int     NOT NULL,
    revelada  boolean NOT NULL DEFAULT false,
    CONSTRAINT rondas_poker_numero_uq UNIQUE (sesion_id, numero)
);

CREATE TABLE votos (
    id            bigserial PRIMARY KEY,
    ronda_id      bigint NOT NULL REFERENCES rondas_poker (id),
    integrante_id bigint NOT NULL REFERENCES integrantes (id),
    valor         int    NOT NULL,
    -- Un integrante vota una sola vez por ronda; revotar es un UPDATE (CA-18.1).
    CONSTRAINT votos_uno_por_ronda_uq UNIQUE (ronda_id, integrante_id),
    CONSTRAINT votos_valor_ck CHECK (valor IN (0, 1, 2, 3, 5, 8, 13, 21))
);

CREATE TABLE registros_esfuerzo (
    id            bigserial PRIMARY KEY,
    integrante_id bigint  NOT NULL REFERENCES integrantes (id),
    historia_id   bigint  REFERENCES historias (id),
    tarea_id      bigint  REFERENCES tareas (id),
    fecha         date    NOT NULL,
    actividad     text    NOT NULL DEFAULT '',
    horas         numeric NOT NULL,
    CONSTRAINT registros_esfuerzo_horas_ck CHECK (horas > 0 AND horas <= 24),
    CONSTRAINT registros_esfuerzo_destino_ck CHECK (historia_id IS NOT NULL OR tarea_id IS NOT NULL)
);

CREATE TABLE defectos (
    id                   bigserial PRIMARY KEY,
    proyecto_id          bigint NOT NULL REFERENCES proyectos (id),
    historia_id          bigint REFERENCES historias (id),
    descripcion          text   NOT NULL,
    severidad            text   NOT NULL,
    estado               text   NOT NULL,
    sprint_deteccion_id  bigint NOT NULL REFERENCES sprints (id),
    sprint_resolucion_id bigint REFERENCES sprints (id)
);

-- Seguridad: Supabase publica cada tabla de "public" por su API (clave anon). Con RLS activado y sin
-- políticas esa API no puede leer ni escribir nada; la app entra por DATABASE_URL con el usuario
-- postgres, que no está sujeto a RLS.
ALTER TABLE proyectos            ENABLE ROW LEVEL SECURITY;
ALTER TABLE integrantes          ENABLE ROW LEVEL SECURITY;
ALTER TABLE sprints              ENABLE ROW LEVEL SECURITY;
ALTER TABLE historias            ENABLE ROW LEVEL SECURITY;
ALTER TABLE criterios_aceptacion ENABLE ROW LEVEL SECURITY;
ALTER TABLE tareas               ENABLE ROW LEVEL SECURITY;
ALTER TABLE sprint_historias     ENABLE ROW LEVEL SECURITY;
ALTER TABLE sesiones_poker       ENABLE ROW LEVEL SECURITY;
ALTER TABLE rondas_poker         ENABLE ROW LEVEL SECURITY;
ALTER TABLE votos                ENABLE ROW LEVEL SECURITY;
ALTER TABLE registros_esfuerzo   ENABLE ROW LEVEL SECURITY;
ALTER TABLE defectos             ENABLE ROW LEVEL SECURITY;

-- +goose Down

DROP TABLE defectos;
DROP TABLE registros_esfuerzo;
DROP TABLE votos;
DROP TABLE rondas_poker;
DROP TABLE sesiones_poker;
DROP TABLE sprint_historias;
DROP TABLE tareas;
DROP TABLE criterios_aceptacion;
DROP TABLE historias;
DROP TABLE sprints;
DROP TABLE integrantes;
DROP TABLE proyectos;
