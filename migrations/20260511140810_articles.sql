-- +goose Up
create type article_status as enum ('pending', 'published');

create table articles (
  id bigserial primary key,
  document_name text not null,
  document_fid text not null,
  author_id bigint not null,
  status article_status not null default 'pending'
);

-- +goose Down
drop table articles;
drop type article_status;