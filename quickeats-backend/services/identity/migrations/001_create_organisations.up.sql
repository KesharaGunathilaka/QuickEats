CREATE TABLE organisations (
    id            UUID PRIMARY KEY,
    name          VARCHAR(255) NOT NULL,
    slug          VARCHAR(100) UNIQUE NOT NULL,
    verify_method VARCHAR(50) NOT NULL DEFAULT 'email_domain',
    verify_config JSONB NOT NULL DEFAULT '{}',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
