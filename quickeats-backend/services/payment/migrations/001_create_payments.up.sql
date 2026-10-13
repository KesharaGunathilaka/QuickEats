CREATE TABLE payments (
    id          UUID PRIMARY KEY,
    order_id    UUID UNIQUE NOT NULL,
    method      VARCHAR(20) NOT NULL,
    status      VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    amount      BIGINT NOT NULL,
    currency    VARCHAR(3) NOT NULL DEFAULT 'LKR',
    gateway_ref VARCHAR(255) DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_payments_order ON payments(order_id);

CREATE TABLE ledger_entries (
    id          UUID PRIMARY KEY,
    payment_id  UUID REFERENCES payments(id),
    account     VARCHAR(100) NOT NULL,
    entry_type  VARCHAR(10) NOT NULL,
    amount      BIGINT NOT NULL,
    description VARCHAR(255) DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_ledger_payment ON ledger_entries(payment_id);

CREATE TABLE wallets (
    id         UUID PRIMARY KEY,
    user_id    UUID UNIQUE NOT NULL,
    balance    BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0),
    currency   VARCHAR(3) NOT NULL DEFAULT 'LKR',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE outbox (
    id             BIGSERIAL PRIMARY KEY,
    aggregate_type VARCHAR(50) NOT NULL,
    aggregate_id   UUID NOT NULL,
    event_type     VARCHAR(100) NOT NULL,
    payload        JSONB NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at   TIMESTAMPTZ
);
CREATE INDEX idx_payment_outbox_unpub ON outbox(created_at) WHERE published_at IS NULL;
