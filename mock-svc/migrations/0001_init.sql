-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS mocks
(
    id                uuid PRIMARY KEY     DEFAULT gen_random_uuid(),
    name              text        NOT NULL,
    description       text,
    request_match     jsonb       NOT NULL,
    response_template jsonb       NOT NULL,
    enabled           boolean     NOT NULL DEFAULT true,
    family_id         uuid        NULL,
    created_by        uuid        NOT NULL,
    generated_by      uuid        NULL,
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now(),
    deleted_at        timestamptz NULL
);

CREATE TABLE IF NOT EXISTS families
(
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    primary_mock_id uuid NOT NULL
);

ALTER TABLE families
    ADD CONSTRAINT fk_families_primary_mock
        FOREIGN KEY (primary_mock_id) REFERENCES mocks (id);

ALTER TABLE mocks
    ADD CONSTRAINT fk_mocks_family
        FOREIGN KEY (family_id) REFERENCES families (id);

CREATE TABLE IF NOT EXISTS generations
(
    id              uuid PRIMARY KEY     DEFAULT gen_random_uuid(),
    base_mock_id    uuid        NOT NULL REFERENCES mocks (id),
    dsl_script_id   uuid        NOT NULL,
    created_by      uuid        NOT NULL,
    status          text        NOT NULL CHECK (status IN ('QUEUED', 'RUNNING', 'DONE', 'FAILED')),
    params          jsonb,
    result_mock_ids text[],
    error           text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS derived_relationships
(
    id              uuid PRIMARY KEY     DEFAULT gen_random_uuid(),
    base_mock_id    uuid        NOT NULL REFERENCES mocks (id),
    derived_mock_id uuid        NOT NULL REFERENCES mocks (id),
    dsl_script_id   uuid        NOT NULL,
    generation_id   uuid        NOT NULL REFERENCES generations (id),
    params          jsonb,
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS mock_search
(
    mock_id     uuid PRIMARY KEY,
    family_id   uuid        NULL,
    enabled     boolean     NOT NULL,
    deleted_at  timestamptz NULL,
    updated_at  timestamptz NOT NULL,
    method      text        NULL,
    path        text        NULL,
    name        text        NOT NULL,
    description text        NULL,
    search_tsv  tsvector    NOT NULL
);

CREATE TABLE IF NOT EXISTS outbox
(
    id              uuid PRIMARY KEY     DEFAULT gen_random_uuid(),
    event_type      text        NOT NULL,
    payload         jsonb       NOT NULL,
    attempts        int         NOT NULL DEFAULT 0,
    max_attempts    int         NOT NULL,
    last_error      text,
    next_attempt_at timestamptz NOT NULL DEFAULT now(),
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_mocks_family_active ON mocks (family_id) WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX ux_families_primary_mock_id ON families (primary_mock_id);

CREATE UNIQUE INDEX ux_derived_rel_derived ON derived_relationships (derived_mock_id);
CREATE INDEX idx_derived_rel_base ON derived_relationships (base_mock_id);

CREATE INDEX idx_generations_base_created ON generations (base_mock_id, created_at DESC);

CREATE INDEX idx_mock_search_family_enabled_updated ON mock_search (family_id, enabled, updated_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_mock_search_method_path ON mock_search (method, path) WHERE deleted_at IS NULL;
CREATE INDEX idx_mock_search_tsv ON mock_search USING GIN (search_tsv) WHERE deleted_at IS NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS outbox;
DROP TABLE IF EXISTS mock_search;
DROP TABLE IF EXISTS derived_relationships;
DROP TABLE IF EXISTS generations;
DROP TABLE IF EXISTS families;
DROP TABLE IF EXISTS mocks;
-- +goose StatementEnd
