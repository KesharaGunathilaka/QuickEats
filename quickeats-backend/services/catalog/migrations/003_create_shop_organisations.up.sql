-- A shop can serve several organisations (a merchant picks one or more codes).
CREATE TABLE shop_organisations (
    shop_id         UUID NOT NULL REFERENCES shops(id) ON DELETE CASCADE,
    organisation_id UUID NOT NULL,
    PRIMARY KEY (shop_id, organisation_id)
);
CREATE INDEX idx_shop_orgs_org ON shop_organisations(organisation_id);

-- Existing shops keep serving the organisation they were created in.
INSERT INTO shop_organisations (shop_id, organisation_id)
SELECT id, organisation_id FROM shops
ON CONFLICT DO NOTHING;
