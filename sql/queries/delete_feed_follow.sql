-- name: Unfollow :exec
DELETE FROM feed_follows
USING feeds, users
WHERE feeds.id = feed_follows.feed_id
AND users.id = feed_follows.user_id
AND users.id = $1;