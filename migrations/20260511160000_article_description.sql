-- +goose Up
ALTER TABLE articles ADD COLUMN description text NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE articles DROP COLUMN description;
