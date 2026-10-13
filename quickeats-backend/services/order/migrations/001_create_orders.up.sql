CREATE TABLE orders (
    id                     UUID PRIMARY KEY,
    idempotency_key        VARCHAR(64) UNIQUE NOT NULL,
    customer_id            UUID NOT NULL,
    shop_id                UUID NOT NULL,
    status                 VARCHAR(30) NOT NULL DEFAULT 'PENDING_PAYMENT',
    pickup_code            VARCHAR(8) DEFAULT '',
    cancellation_deadline  TIMESTAMPTZ,
    total_amount           BIGINT NOT NULL,
    currency               VARCHAR(3) NOT NULL DEFAULT 'LKR',
    payment_method         VARCHAR(20) NOT NULL DEFAULT 'card',
    reject_reason          TEXT DEFAULT '',
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_customer ON orders(customer_id);
CREATE INDEX idx_orders_shop ON orders(shop_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_idempotency ON orders(idempotency_key);

CREATE TABLE order_items (
    id           UUID PRIMARY KEY,
    order_id     UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    menu_item_id UUID NOT NULL,
    item_name    VARCHAR(255) NOT NULL,
    quantity     INT NOT NULL CHECK (quantity > 0),
    unit_price   BIGINT NOT NULL,
    subtotal     BIGINT NOT NULL
);

CREATE INDEX idx_order_items_order ON order_items(order_id);

CREATE TABLE outbox (
    id             BIGSERIAL PRIMARY KEY,
    aggregate_type VARCHAR(50) NOT NULL,
    aggregate_id   UUID NOT NULL,
    event_type     VARCHAR(100) NOT NULL,
    payload        JSONB NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at   TIMESTAMPTZ
);

CREATE INDEX idx_outbox_unpublished ON outbox(created_at) WHERE published_at IS NULL;
