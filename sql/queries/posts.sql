-- name: CreatePost :one
INSERT INTO posts (id, title, url, description, feed_id, published_at, created_at, updated_at)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;

-- name: GetPostsForUser :many
SELECT posts.* 
FROM posts, feeds, feed_follows
WHERE posts.feed_id = feeds.id
AND feeds.id = feed_follows.feed_id
AND feed_follows.user_id = $1
ORDER BY posts.published_at DESC
LIMIT $2;