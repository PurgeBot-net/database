package database

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type guildTable struct {
	name string
	// Narrows DeleteGuild to rows whose loss cannot undo a server's privacy choice; the sweeper deletes the rest once the bot is gone.
	erasable string
}

// guild_departures is the sweeper's bookkeeping and is handled by name. access_roles is fully
// erasable because losing grants only narrows access to the owner.
var guildTables = []guildTable{
	{name: "customizations"},
	{name: "purge_events"},
	{name: "guild_settings", erasable: "stats_enabled"},
	{name: "access_roles"},
}

func guildTableNames() []string {
	names := make([]string, len(guildTables))
	for i, t := range guildTables {
		names[i] = t.name
	}
	return names
}

func (db *Database) checkGuildTables(ctx context.Context) error {
	var unlisted *string
	err := db.pool.QueryRow(ctx, `
		SELECT string_agg(table_name::text, ', ')
		FROM information_schema.columns
		WHERE table_schema = current_schema() AND column_name = 'guild_id' AND table_name::text <> ALL($1::text[])
	`, append(guildTableNames(), "guild_departures")).Scan(&unlisted)
	if err != nil {
		return err
	}
	if unlisted != nil {
		return fmt.Errorf("tables with guild_id missing from guildTables: %s", *unlisted)
	}
	return nil
}

type GuildExport struct {
	Tables map[string]json.RawMessage
	Avatar []byte
}

func (db *Database) ExportGuild(ctx context.Context, guildID int64) (GuildExport, error) {
	var e GuildExport
	if err := db.checkGuildTables(ctx); err != nil {
		return e, err
	}
	var pairs []string
	for _, name := range append(guildTableNames(), "guild_departures") {
		// The avatar goes out as an image file instead.
		pairs = append(pairs, fmt.Sprintf(`'%s', (SELECT COALESCE(jsonb_agg(to_jsonb(t) - 'bot_avatar'), '[]') FROM %s t WHERE guild_id = $1)`, name, name))
	}
	err := db.pool.QueryRow(ctx, `
		SELECT jsonb_build_object(`+strings.Join(pairs, ", ")+`),
			(SELECT decode(bot_avatar, 'base64') FROM customizations WHERE guild_id = $1)
	`, guildID).Scan(&e.Tables, &e.Avatar)
	return e, err
}

type DeleteGuildParams struct {
	GuildID int64
	Confirm bool
}

type GuildRows struct {
	Table   string
	Deleted int64
	Kept    int64
}

func (db *Database) DeleteGuild(ctx context.Context, p DeleteGuildParams) ([]GuildRows, error) {
	if err := db.checkGuildTables(ctx); err != nil {
		return nil, err
	}
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	// Without Confirm everything rolls back, so a dry run reports exactly what a real run removes.
	defer tx.Rollback(ctx) //nolint:errcheck

	if err := lockGuildStats(ctx, tx, p.GuildID); err != nil {
		return nil, err
	}
	// 1970 makes the guild due at the next sweep, which keeps the opt-out if the bot is still there and clears the rest if not.
	// Written first, in the order DeleteDepartedGuilds locks rows.
	if _, err := tx.Exec(ctx, `
		INSERT INTO guild_departures (guild_id, left_at) VALUES ($1, 'epoch')
		ON CONFLICT (guild_id) DO UPDATE SET left_at = EXCLUDED.left_at
	`, p.GuildID); err != nil {
		return nil, fmt.Errorf("guild_departures: %w", err)
	}

	rows := make([]GuildRows, 0, len(guildTables))
	for _, t := range guildTables {
		sql := `DELETE FROM ` + t.name + ` WHERE guild_id = $1`
		if t.erasable != "" {
			sql += ` AND ` + t.erasable
		}
		tag, err := tx.Exec(ctx, sql, p.GuildID)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", t.name, err)
		}
		r := GuildRows{Table: t.name, Deleted: tag.RowsAffected()}
		if t.erasable != "" {
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM `+t.name+` WHERE guild_id = $1`, p.GuildID).Scan(&r.Kept); err != nil {
				return nil, fmt.Errorf("%s: %w", t.name, err)
			}
		}
		rows = append(rows, r)
	}
	// Committed even with no rows here, since the guild may still have Redis job keys only the sweeper can reach.
	if !p.Confirm {
		return rows, nil
	}
	return rows, tx.Commit(ctx)
}
