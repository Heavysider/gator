-- +goose Up
CREATE TABLE users(
  id uuid primary key default gen_random_uuid(),
  name varchar(255) unique not null,
  created_at timestamp not null,
  updated_at timestamp not null
);

-- +goose Down
DROP TABLE users;
