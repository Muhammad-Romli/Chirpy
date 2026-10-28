-- +goose Up
ALTER TABLE refresh_tokens
ALTER COLUMN created_at SET NOT NULL;

ALTER TABLE refresh_tokens
ALTER COLUMN updated_at SET NOT NULL;

-- +goose Down 
ALTER TABLE refresh_tokens
ALTER COLUMN created_at DROP NOT NULL;

ALTER TABLE refresh_tokens
ALTER COLUMN updated_at DROP NOT NULL;