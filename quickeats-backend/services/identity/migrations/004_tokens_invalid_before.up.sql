-- Lets a password change or logout revoke every refresh token issued so far:
-- a refresh token with an iat older than this is rejected (see auth_handler.go).
ALTER TABLE users ADD COLUMN tokens_invalid_before TIMESTAMPTZ;
