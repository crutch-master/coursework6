-- +goose Up
create table reviews (
  id bigserial primary key,
  article_id bigint not null,
  reviewer_id bigint not null,
  text text not null
);

-- +goose Down
drop table reviews;
