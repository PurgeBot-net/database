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

### `cmd/`

Operator tools for server data requests, run by hand and never built into an image. Each takes `-db`, a libpq connection string that defaults to `service=purgebot-local`; the services are defined in the operator's own `.pg_service.conf` and `pgpass.conf`.

- `go run ./cmd/exportguild -guildid <id>` writes `exports/guild-<id>.zip` (gitignored) with one JSON file per table and the avatar as an image. Built on `ExportGuild`.
- `go run ./cmd/deleteguild -guildid <id> [-confirm]` deletes the server's rows, role grants included, so the server is back to owner-only. It keeps a stats opt-out and dates its departure 1970 so the next sweep removes the rest once Discord confirms the bot has left. Without `-confirm` it rolls back. Built on `DeleteGuild`.

### `queries/`

Read-only SQL to run by hand in a client like HeidiSQL. `stats-*` gives bot-wide numbers, `guild-*` shows what is stored for one server, and `ops-*` covers housekeeping. Every file starts with the question it answers, what to watch out for when reading the result, and which value to change. The `guild-stats-*` files reuse the SQL from `GetGuildStats`, so they give the same numbers as `/stats` as long as the session uses the server's default time zone.

## Schema

The schema is in `schema.sql` and is idempotent (`CREATE TABLE IF NOT EXISTS`). A copy of it lives in the docker repo at `init/01_schema.sql`, which `docker-compose.yml` mounts into the Postgres container on first start.

### Tables

| Table              | Description                                                     |
| ------------------ | --------------------------------------------------------------- |
| `customizations`   | Per-guild bot name, avatar, and branding toggle (Premium)       |
| `purge_events`     | Record of every completed purge job for stats                   |
| `guild_settings`   | Per-guild settings: whether purge stats are kept                |
| `access_roles`     | Roles the server owner has allowed to use the bot               |
| `guild_departures` | When the bot left a guild, so its data is deleted 30 days later |
