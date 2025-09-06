-- +goose Up
CREATE TABLE feeds(
  id uuid primary key default gen_random_uuid(),
  name varchar(255) unique not null,
  url varchar(1000) unique not null,
  user_id uuid not null references users(id) on delete cascade,
  created_at timestamp not null,
  updated_at timestamp not null
);

-- +goose Down
DROP TABLE feeds;
