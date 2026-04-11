-- name: GetFollowing :many
SELECT feed_follows.*,
    feeds.name AS feed_name
FROM feed_follows
INNER JOIN feeds ON feeds.id = feed_follows.feed_id
INNER JOIN users ON users.id = feed_follows.user_id
WHERE users.id = $1;
