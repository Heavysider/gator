-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
  INSERT INTO feed_follows (id, user_id, feed_id, created_at, updated_at)
  VALUES (
      $1,
      $2,
      $3,
      $4,
      $5
  )
  RETURNING *
)
SELECT inserted_feed_follow.*, users.name as user_name, feeds.name as feed_name
FROM inserted_feed_follow, users, feeds
WHERE inserted_feed_follow.user_id = users.id
AND inserted_feed_follow.feed_id = feeds.id;

-- name: GetFeedFollowsForUser :many
SELECT users.name as user_name, feeds.name as feed_name, feed_follows.* 
FROM feed_follows, users, feeds
WHERE feed_follows.user_id = users.id
AND feed_follows.feed_id = feeds.id
AND users.id = $1;

-- name: DeleteFeedFollowForUser :exec
DELETE FROM feed_follows
WHERE user_id = $1 AND feed_id = $2;