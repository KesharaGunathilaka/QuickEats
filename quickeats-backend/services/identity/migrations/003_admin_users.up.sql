-- Platform admins do not belong to any organisation.
ALTER TABLE users ALTER COLUMN organisation_id DROP NOT NULL;
