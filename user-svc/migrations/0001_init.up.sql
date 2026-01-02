-- +goose Up
-- +goose StatementBegin
CREATE TABLE users
(
    id            uuid PRIMARY KEY     DEFAULT gen_random_uuid(),
    email         text        NOT NULL UNIQUE,
    password_hash text        NOT NULL,
    status        text        NOT NULL DEFAULT 'ACTIVE',
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now(),
    last_login_at timestamptz,
    CONSTRAINT users_status_check CHECK (status IN ('ACTIVE', 'BLOCKED', 'DELETED'))
);

CREATE TABLE roles
(
    name       text PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE user_roles
(
    user_id    uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role_name  text        NOT NULL REFERENCES roles (name) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, role_name)
);

CREATE TABLE sessions
(
    id           uuid PRIMARY KEY     DEFAULT gen_random_uuid(),
    user_id      uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    refresh_hash text        NOT NULL UNIQUE,
    created_at   timestamptz NOT NULL DEFAULT now(),
    last_used_at timestamptz NOT NULL DEFAULT now(),
    revoked_at   timestamptz,
    ip           text,
    user_agent   text
);

CREATE TABLE resources
(
    resource_type text        NOT NULL,
    resource_id   uuid        NOT NULL,
    owner_user_id uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at    timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (resource_type, resource_id)
);

CREATE TABLE resource_grants
(
    id            uuid PRIMARY KEY     DEFAULT gen_random_uuid(),
    resource_type text        NOT NULL,
    resource_id   uuid        NOT NULL,
    user_id       uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    permission    text        NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE (resource_type, resource_id, user_id),
    CONSTRAINT resource_grants_resource_fk
        FOREIGN KEY (resource_type, resource_id)
            REFERENCES resources (resource_type, resource_id)
            ON DELETE CASCADE,
    CONSTRAINT resource_grants_permission_check CHECK (permission IN ('READ', 'EDIT'))
);

CREATE TABLE audit_log
(
    id            uuid PRIMARY KEY     DEFAULT gen_random_uuid(),
    actor_user_id uuid REFERENCES users (id),
    event_type    text        NOT NULL,
    resource_type text,
    resource_id   uuid,
    ip            text,
    user_agent    text,
    metadata      jsonb,
    created_at    timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_email_lower ON users (lower(email));
CREATE INDEX idx_sessions_user_id ON sessions (user_id);
CREATE INDEX idx_resources_owner_user_id ON resources (owner_user_id);
CREATE INDEX idx_resource_grants_user_id ON resource_grants (user_id);
CREATE INDEX idx_resource_grants_resource ON resource_grants (resource_type, resource_id);
CREATE INDEX idx_resource_grants_user_type ON resource_grants (user_id, resource_type);
CREATE INDEX idx_audit_log_actor_user_id ON audit_log (actor_user_id);
CREATE INDEX idx_audit_log_created_at ON audit_log (created_at);

INSERT INTO roles (name)
VALUES ('ADMIN'),
       ('USER')
ON CONFLICT DO NOTHING;
-- +goose StatementEnd
