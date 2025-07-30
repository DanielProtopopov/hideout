BEGIN;

CREATE TABLE IF NOT EXISTS public.secrets
(
    id         SERIAL PRIMARY KEY,
    path_id    INTEGER        NOT NULL,
    uid        uuid           NOT NULL UNIQUE,
    name       VARCHAR(255)   NOT NULL DEFAULT '',
    value      VARCHAR(65535) NOT NULL DEFAULT '',
    type       VARCHAR(255)   NOT NULL DEFAULT '',
    created_at TIMESTAMP               DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP               DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP    NULL DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS public.paths
(
    id         SERIAL PRIMARY KEY,
    parent_id  INTEGER      NULL     DEFAULT NULL,
    uid        uuid         NOT NULL UNIQUE,
    name       VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMP             DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP             DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMP    NULL DEFAULT NULL
);

COMMIT;