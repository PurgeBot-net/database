package database

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type GuildSettings struct {
	GuildID      int64
	StatsEnabled bool
}

func (db *Database) GetGuildSettings(ctx context.Context, guildID int64) (*GuildSettings, error) {
	var s GuildSettings
	err := db.pool.QueryRow(ctx, `
		SELECT guild_id, stats_enabled FROM guild_settings WHERE guild_id = $1
	`, guildID).Scan(&s.GuildID, &s.StatsEnabled)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (db *Database) DisableStats(ctx context.Context, guildID int64) error {
	return pgx.BeginFunc(ctx, db.pool, func(tx pgx.Tx) error {
		if err := lockGuildStats(ctx, tx, guildID); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `
			WITH s AS (
				INSERT INTO guild_settings (guild_id, stats_enabled) VALUES ($1, false)
				ON CONFLICT (guild_id) DO UPDATE SET stats_enabled = false
			)
			DELETE FROM purge_events WHERE guild_id = $1
		`, guildID)
		return err
	})
}

// EnableStats never inserts, so a guild that never opted out gets no settings row.
func (db *Database) EnableStats(ctx context.Context, guildID int64) error {
	_, err := db.pool.Exec(ctx, `UPDATE guild_settings SET stats_enabled = true WHERE guild_id = $1`, guildID)
	return err
}
