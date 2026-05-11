-- +goose Up
create table users (
  id bigserial primary key,
  name text not null,
  login text not null,
  description text not null,
  password_hash bytea not null
);

create index users_login on users (login);

-- +goose Down
drop table users;
