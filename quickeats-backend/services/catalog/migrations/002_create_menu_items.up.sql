CREATE TABLE menu_items (
    id          UUID PRIMARY KEY,
    shop_id     UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    price       BIGINT NOT NULL,
    category    VARCHAR(100) DEFAULT 'General',
    image_url   VARCHAR(512) DEFAULT '',
    stock       INT NOT NULL DEFAULT 0,
    available   BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_menu_items_shop ON menu_items(shop_id);

CREATE TABLE stock_reservations (
    id           UUID PRIMARY KEY,
    order_id     UUID NOT NULL,
    menu_item_id UUID NOT NULL REFERENCES menu_items(id),
    quantity     INT NOT NULL,
    expires_at   TIMESTAMPTZ NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_stock_res_order ON stock_reservations(order_id);
CREATE INDEX idx_stock_res_expires ON stock_reservations(expires_at) WHERE expires_at > now();
