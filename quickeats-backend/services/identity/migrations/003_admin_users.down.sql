DELETE FROM users WHERE organisation_id IS NULL;
ALTER TABLE users ALTER COLUMN organisation_id SET NOT NULL;
