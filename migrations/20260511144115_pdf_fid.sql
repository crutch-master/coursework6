-- +goose Up
ALTER TYPE article_status ADD VALUE 'failed';
ALTER TABLE articles ADD COLUMN pdf_fid text;

-- +goose Down
ALTER TABLE articles DROP COLUMN pdf_fid;