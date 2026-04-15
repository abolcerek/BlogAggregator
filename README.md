# Gator

Gator is a CLI blog aggregator built in Go. It lets you subscribe to RSS feeds, automatically fetch new posts on a schedule, and browse the latest content all from your terminal.

## Prerequisites

Make sure you have the following installed before getting started:

- **Go** (1.22 or later) — [https://go.dev/dl/](https://go.dev/dl/)
- **PostgreSQL** (14 or later) — [https://www.postgresql.org/download/](https://www.postgresql.org/download/)

You'll also need a running Postgres instance with a database created for Gator to use:

```sql
CREATE DATABASE gator;
```

## Installation

Install the `gator` CLI directly using `go install`:

```bash
go install github.com/aaronbolcerek/BlogAggregator@latest
```

## Configuration

Gator reads its config from a JSON file located at `~/.gatorconfig.json`. Create that file with the following structure:

```json
{
  "db_url": "postgres://username:password@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

Replace `username` and `password` with your Postgres credentials. The `current_user_name` field will be updated automatically when you log in.

## Running the Program

Once installed and configured, you run Gator commands in this format:

```bash
gator <command> [arguments]
```

## Commands

| Command | Description |
|---|---|
| `gator register <name>` | Create a new user and log in as them |
| `gator login <name>` | Switch to an existing user |
| `gator users` | List all registered users (current user marked) |
| `gator reset` | Delete all users from the database |

| Command | Description |
|---|---|
| `gator addfeed <name> <url>` | Add a new RSS feed and automatically follow it |
| `gator feeds` | List all feeds in the database |
| `gator follow <url>` | Follow an existing feed by URL |
| `gator following` | List feeds the current user follows |
| `gator unfollow <url>` | Unfollow a feed |

| Command | Description |
|---|---|
| `gator agg <interval>` | Start fetching feeds on a schedule (e.g. `30s`, `1m`, `5m`) |
| `gator browse [limit]` | Browse the latest posts for the current user (default: 2) |

## Example Workflow

```bash
# Register a new user
gator register alice

# Add some RSS feeds
gator addfeed "Hacker News" "https://news.ycombinator.com/rss"
gator addfeed "Boot.dev Blog" "https://blog.boot.dev/index.xml"

# Start the aggregator (fetches new posts every minute)
gator agg 1m

# In a separate terminal, browse your latest posts
gator browse 5
```