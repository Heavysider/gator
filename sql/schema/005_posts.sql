-- +goose Up
CREATE TABLE posts(
  id uuid primary key default gen_random_uuid(),
  title varchar(255) not null,
  url varchar(1000) unique not null,
  description text,
  feed_id uuid not null references feeds(id) on delete cascade,
  published_at timestamp not null,
  created_at timestamp not null,
  updated_at timestamp not null
);

-- +goose Down
DROP TABLE posts;