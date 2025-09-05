CREATE EXTENSION IF NOT EXISTS "pgcrypto"; 

CREATE TABLE IF NOT EXISTS audit_event (
    event_id     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp    TIMESTAMPTZ NOT NULL DEFAULT now(),
    action       TEXT NOT NULL,
    status       TEXT NOT NULL,
    actor_id     TEXT,
    actor_type   TEXT,
    ip_address   TEXT,
    user_agent   TEXT,
    resource     TEXT NOT NULL,
    resource_id  TEXT NOT NULL,
    details      JSONB,
    service_name TEXT NOT NULL
);

