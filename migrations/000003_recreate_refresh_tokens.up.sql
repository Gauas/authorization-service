DROP TABLE IF EXISTS refresh_tokens;

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE refresh_tokens (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID        NOT NULL,
    device_id    TEXT        NOT NULL,
    permission   TEXT        NOT NULL,
    refresh_token TEXT       NOT NULL UNIQUE,
    issued_at    TIMESTAMPTZ NOT NULL,
    expires_at   TIMESTAMPTZ NOT NULL,
    revoked_at   TIMESTAMPTZ DEFAULT NULL,
    created_at   TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);
