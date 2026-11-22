# database

Shared database library for PurgeBot. Wraps a pgx connection pool and exposes typed query functions. One file per table, plus `guilds.go` for exporting or deleting one server's data.

## Usage

```go
import "github.com/PurgeBot-net/database"

db, err := database.New(ctx, cfg.Database.URL())
if err != nil {
    log.Fatal(err)
}
defer db.Close()

c, err := db.GetCustomization(ctx, guildID)
```

`database.Config` holds the env tags `DATABASE_HOST` (default `localhost`), `DATABASE_PORT` (`5432`), `DATABASE_NAME` (`purgebot`), `DATABASE_USER` (`purgebot`) and `DATABASE_PASSWORD`. Each service keeps it as the `Database` field of its own config. `URL()` builds the `postgres://` URL with user and password escaped, so any password works. `New` takes any pgx connection string, which is how the `cmd/` tools pass a libpq service.

`GetCustomization` returns only the bot name and the branding toggle. The avatar stays in the table for the data export, and `UpsertCustomization` keeps the stored one when `BotAvatar` is nil.

### `cmd/`

Operator tools for server data requests, run by hand and never built into an image. Each takes `-db`, a libpq connection string that defaults to `service=purgebot-local`; the services are defined in the operator's own `.pg_service.conf` and `pgpass.conf`.

- `go run ./cmd/exportguild -guildid <id>` writes `exports/guild-<id>.zip` (gitignored) with one JSON file per table and the avatar as an image. Built on `ExportGuild`.
- `go run ./cmd/deleteguild -guildid <id> [-confirm]` deletes the server's rows, role grants included, so the server is back to owner-only. It keeps a stats opt-out and dates its departure 1970 so the next sweep removes the rest once Discord confirms the bot has left. Without `-confirm` it rolls back. Built on `DeleteGuild`.

### `queries/`

Read-only SQL to run by hand in a client like HeidiSQL. `stats-*` gives bot-wide numbers, `guild-*` shows what is stored for one server, and `ops-*` covers housekeeping. Every file starts with the question it answers, what to watch out for when reading the result, and which value to change. The `guild-stats-*` files reuse the SQL from `GetGuildStats`, so they give the same numbers as `/stats` as long as the session uses the server's default time zone.

## Schema

The schema is in `schema.sql` and is idempotent (`CREATE TABLE IF NOT EXISTS`). A copy of it lives in the docker repo at `init/01_schema.sql`, which `docker-compose.yml` mounts into the Postgres container on first start. A `DO` block also gives a legacy `customizations.updated_by` column a default of `0` on older databases that still have one; no code writes it.

### Tables

| Table              | Columns                                                                               | Description                                                                              |
| ------------------ | ------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------- |
| `customizations`   | `guild_id`, `bot_name`, `bot_avatar`, `remove_branding`, `updated_at`                 | Per-guild bot name, avatar image, and branding toggle (Premium)                          |
| `purge_events`     | `id`, `guild_id`, `purge_type`, `target_type`, `deleted`, `duration_ms`, `created_at` | Record of each completed purge job for stats, unless the server turned stats off         |
| `guild_settings`   | `guild_id`, `stats_enabled`                                                           | Per-guild settings: whether purge stats are kept                                         |
| `access_roles`     | `guild_id`, `role_id`                                                                 | Roles the server owner has allowed to use the bot                                        |
| `guild_departures` | `guild_id`, `left_at`                                                                 | When a sweep first found the bot gone from a guild, so its data is deleted after 30 days |
