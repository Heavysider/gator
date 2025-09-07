-- +goose Up
CREATE TABLE feed_follows(
  id uuid primary key default gen_random_uuid(),
  user_id uuid not null references users(id) on delete cascade,
  feed_id uuid not null references feeds(id) on delete cascade,
  created_at timestamp not null,
  updated_at timestamp not null,
  UNIQUE (user_id, feed_id)
);

-- +goose Down
DROP TABLE feed_follows;
