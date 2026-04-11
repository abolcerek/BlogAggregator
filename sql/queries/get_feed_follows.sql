-- name: GetFollowing :many
SELECT feed_follows.*,
    feeds.name AS feed_n