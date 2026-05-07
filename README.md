# database

Shared database library for PurgeBot. Wraps a pgx connection pool and exposes typed query functions. One file per table.

## Usage

```go
import "github.com/PurgeBot-net/database"

db, err := database.New(ctx, os.Getenv("DATABASE_URL"))
if err != nil {
    log.Fatal(err)
}
defer db.Close()

c, err := db.GetCustomization(ctx, guildID)
```

## Schema

The schema is in `schema.sql` and is idempotent (`CREATE TABLE IF NOT EXISTS`). It is mounted into the Postgres container on first start via the docker repo's `docker-compose.yml`.

### Tables

| Table            | Description                                               |
| ---------------- | --------------------------------------------------------- |
| `customizations` | Per-guild bot name, avatar, and branding toggle (Premium) |
| `purge_events`   | Record of every completed purge job for stats             |
