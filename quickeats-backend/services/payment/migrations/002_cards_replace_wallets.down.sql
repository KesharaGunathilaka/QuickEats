DROP TABLE IF EXISTS cards;

CREATE TABLE wallets (
    id         UUID PRIMARY KEY,
    user_id    UUID UNIQUE NOT NULL,
    balance    BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0),
    currency   VARCHAR(3) NOT NULL DEFAULT 'LKR',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
