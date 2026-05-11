-- +goose Up

-- subjects is a unified principals table: roles, groups, and users register here.
-- access_rules.subject_id references this table, eliminating the need for subject_type on rules.
CREATE TABLE IF NOT EXISTS subjects (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    type      TEXT NOT NULL,      -- "role" | "group" | "user"
    entity_id INTEGER NOT NULL,
    UNIQUE(type, entity_id)
);

CREATE TABLE IF NOT EXISTS users (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME,
    uuid          TEXT NOT NULL UNIQUE,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    status        TEXT NOT NULL DEFAULT 'active'
);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

CREATE TABLE IF NOT EXISTS roles (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at  DATETIME,
    updated_at  DATETIME,
    deleted_at  DATETIME,
    uuid        TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL UNIQUE,
    description TEXT,
    system      INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_roles_deleted_at ON roles(deleted_at);

CREATE TABLE IF NOT EXISTS permissions (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME,
    uuid       TEXT NOT NULL UNIQUE,
    name       TEXT NOT NULL UNIQUE
);
CREATE INDEX IF NOT EXISTS idx_permissions_deleted_at ON permissions(deleted_at);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id       INTEGER NOT NULL,
    permission_id INTEGER NOT NULL,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id INTEGER NOT NULL,
    role_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE IF NOT EXISTS groups (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at  DATETIME,
    updated_at  DATETIME,
    deleted_at  DATETIME,
    uuid        TEXT NOT NULL UNIQUE,
    name        TEXT NOT NULL UNIQUE,
    description TEXT
);
CREATE INDEX IF NOT EXISTS idx_groups_deleted_at ON groups(deleted_at);

CREATE TABLE IF NOT EXISTS user_groups (
    user_id  INTEGER NOT NULL,
    group_id INTEGER NOT NULL,
    PRIMARY KEY (user_id, group_id)
);

CREATE TABLE IF NOT EXISTS sessions (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at   DATETIME,
    updated_at   DATETIME,
    deleted_at   DATETIME,
    uuid         TEXT NOT NULL UNIQUE,
    user_id      INTEGER NOT NULL,
    expires_at   DATETIME NOT NULL,
    last_seen_at DATETIME,
    ip           TEXT,
    user_agent   TEXT,
    revoked_at   DATETIME
);
CREATE INDEX IF NOT EXISTS idx_sessions_deleted_at ON sessions(deleted_at);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id    ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);

CREATE TABLE IF NOT EXISTS sp_sessions (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME,
    session_id INTEGER NOT NULL,
    sp_id      INTEGER NOT NULL,
    sub        TEXT NOT NULL,
    UNIQUE (session_id, sp_id)
);
CREATE INDEX IF NOT EXISTS idx_sp_sessions_deleted_at ON sp_sessions(deleted_at);

CREATE TABLE IF NOT EXISTS service_providers (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME,
    uuid       TEXT NOT NULL UNIQUE,
    slug       TEXT NOT NULL UNIQUE,
    name       TEXT NOT NULL,
    protocol   TEXT NOT NULL,
    enabled    INTEGER NOT NULL DEFAULT 1
);
CREATE INDEX IF NOT EXISTS idx_service_providers_deleted_at ON service_providers(deleted_at);

CREATE TABLE IF NOT EXISTS oidc_clients (
    id                     INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at             DATETIME,
    updated_at             DATETIME,
    deleted_at             DATETIME,
    sp_id                  INTEGER NOT NULL UNIQUE,
    client_id              TEXT NOT NULL UNIQUE,
    client_secret_hash     TEXT NOT NULL,
    redirect_uris          TEXT NOT NULL DEFAULT '[]',
    grant_types            TEXT NOT NULL DEFAULT '["authorization_code"]',
    response_types         TEXT NOT NULL DEFAULT '["code"]',
    scopes                 TEXT NOT NULL DEFAULT '["openid"]',
    token_endpoint_auth    TEXT NOT NULL DEFAULT 'client_secret_basic',
    pkce_required          INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_oidc_clients_deleted_at ON oidc_clients(deleted_at);

CREATE TABLE IF NOT EXISTS saml_clients (
    id                     INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at             DATETIME,
    updated_at             DATETIME,
    deleted_at             DATETIME,
    sp_id                  INTEGER NOT NULL UNIQUE,
    entity_id              TEXT NOT NULL UNIQUE,
    acs_urls               TEXT NOT NULL DEFAULT '[]',
    slo_url                TEXT,
    name_id_format         TEXT NOT NULL DEFAULT 'urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress',
    sp_certificate         TEXT,
    want_signed_requests   INTEGER NOT NULL DEFAULT 0,
    want_signed_assertions INTEGER NOT NULL DEFAULT 1
);
CREATE INDEX IF NOT EXISTS idx_saml_clients_deleted_at ON saml_clients(deleted_at);

CREATE TABLE IF NOT EXISTS access_rules (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at   DATETIME,
    updated_at   DATETIME,
    deleted_at   DATETIME,
    sp_id        INTEGER NOT NULL,
    rule_type    TEXT NOT NULL,
    subject_id   INTEGER NOT NULL REFERENCES subjects(id),
    priority     INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_access_rules_deleted_at ON access_rules(deleted_at);
CREATE INDEX IF NOT EXISTS idx_access_rules_sp_id      ON access_rules(sp_id);

CREATE TABLE IF NOT EXISTS signing_keys (
    id                     INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at             DATETIME,
    updated_at             DATETIME,
    deleted_at             DATETIME,
    protocol               TEXT NOT NULL,
    kid                    TEXT NOT NULL,
    algorithm              TEXT NOT NULL,
    private_key_encrypted  BLOB NOT NULL,
    public_key             TEXT NOT NULL,
    certificate            TEXT,
    status                 TEXT NOT NULL DEFAULT 'active',
    activated_at           DATETIME,
    retired_at             DATETIME
);
CREATE INDEX IF NOT EXISTS idx_signing_keys_deleted_at       ON signing_keys(deleted_at);
CREATE INDEX IF NOT EXISTS idx_signing_keys_protocol_status  ON signing_keys(protocol, status);

CREATE TABLE IF NOT EXISTS events (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at    DATETIME,
    updated_at    DATETIME,
    deleted_at    DATETIME,
    timestamp     DATETIME NOT NULL,
    actor_user_id INTEGER,
    action        TEXT NOT NULL,
    target_type   TEXT,
    target_id     INTEGER,
    sp_id         INTEGER,
    ip            TEXT,
    user_agent    TEXT,
    result        TEXT NOT NULL DEFAULT 'ok',
    metadata_json TEXT
);
CREATE INDEX IF NOT EXISTS idx_events_deleted_at ON events(deleted_at);
CREATE INDEX IF NOT EXISTS idx_events_timestamp  ON events(timestamp);
CREATE INDEX IF NOT EXISTS idx_events_action     ON events(action);

CREATE TABLE IF NOT EXISTS kv_store (
    key        TEXT PRIMARY KEY,
    value      BLOB NOT NULL,
    expires_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_kv_store_expires_at ON kv_store(expires_at);

CREATE TABLE IF NOT EXISTS bootstrap_states (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- +goose Down

DROP TABLE IF EXISTS bootstrap_states;
DROP TABLE IF EXISTS kv_store;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS signing_keys;
DROP TABLE IF EXISTS access_rules;
DROP TABLE IF EXISTS subjects;
DROP TABLE IF EXISTS saml_clients;
DROP TABLE IF EXISTS oidc_clients;
DROP TABLE IF EXISTS service_providers;
DROP TABLE IF EXISTS sp_sessions;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS user_groups;
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;
