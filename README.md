# database

Shared database library for PurgeBot. Wraps a pgx connection pool and exposes typed query functions. One file per table.

## Usage

```go
import "github.com/PurgeBot-net/database"

db, err := database.New(ctx, "postgres://purgebot:password@postgres:5432/purgebot")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

c, err := db.GetCustomization(ctx, guildID)
```

## Schema

The schema is in `schema.sql` and is idempotent (`CREATE TABLE IF NOT EXISTS`). A copy of it lives in the docker repo at `init/01_schema.sql`, which `docker-compose.yml` mounts into the Postgres container on first start.

### Tables

| Table            | Description                                               |
| ---------------- | --------------------------------------------------------- |
| `customizations` | Per-guild bot name, avatar, and branding toggle (Premium) |
| `purge_events`   | Record of every completed purge job for stats             |
