-- +goose Up
create table users (
  id bigserial primary key,
  name text not null,
  description text not null,
  password_hash bytea not null
);

-- +goose Down
drop table users;
