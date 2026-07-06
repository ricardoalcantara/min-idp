-- +goose Up
ALTER TABLE oidc_clients ADD COLUMN post_logout_redirect_uris TEXT NOT NULL DEFAULT '[]';

-- +goose Down
ALTER TABLE oidc_clients DROP COLUMN post_logout_redirect_uris;
