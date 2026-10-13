CREATE TABLE shops (
    id              UUID PRIMARY KEY,
    organisation_id UUID NOT NULL,
    name            VARCHAR(255) NOT NULL,
    description     TEXT DEFAULT '',
    image_url       VARCHAR(512) DEFAULT '',
    is_open         BOOLEAN NOT NULL DEFAULT false,
    owner_id        UUID NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_shops_org ON shops(organisation_id);
