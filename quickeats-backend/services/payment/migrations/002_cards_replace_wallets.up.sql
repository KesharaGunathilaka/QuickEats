-- Card payments replace the wallet. One saved card per customer. Only what the app
-- shows is stored: brand, last four digits, holder name and expiry; never the full
-- number or the CVV. balance is the demo card's available credit, never returned
-- by the API.
CREATE TABLE cards (
    id          UUID PRIMARY KEY,
    user_id     UUID UNIQUE NOT NULL,
    brand       VARCHAR(20) NOT NULL,
    last4       CHAR(4) NOT NULL,
    holder_name VARCHAR(100) NOT NULL,
    exp_month   SMALLINT NOT NULL CHECK (exp_month BETWEEN 1 AND 12),
    exp_year    SMALLINT NOT NULL,
    balance     BIGINT NOT NULL CHECK (balance >= 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

DROP TABLE IF EXISTS wallets;
