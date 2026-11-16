package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

func (db *Database) GuildIDsWithData(ctx context.Context) ([]int64, error) {
	rows, err := db.pool.Query(ctx, `
		SELECT guild_id FROM customizations
		UNION SELECT guild_id FROM purge_events
		UNION SELECT guild_id FROM guild_settings
	`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[int64])
}

func (db *Database) GetDepartures(ctx context.Context) (map[int64]time.Time, error) {
	rows, err := db.pool.Query(ctx, `SELECT guild_id, left_at FROM guild_departures`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	departed := make(map[int64]time.Time)
	for rows.Next() {
		var id int64
		var leftAt time.Time
		if err := rows.Scan(&id, &leftAt); err != nil {
			return nil, err
		}
		departed[id] = leftAt
	}
	return departed, rows.Err()
}

type MarkDepartedParams struct {
	GuildIDs []int64
	LeftAt   time.Time
}

// DO NOTHING, so a repeated or concurrent run never restarts a guild's 30 days.
func (db *Database) MarkDeparted(ctx context.Context, p MarkDepartedParams) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO guild_departures (guild_id, left_at)
		SELECT unnest($1::bigint[]), $2::timestamptz
		ON CONFLICT (guild_id) DO NOTHING
	`, p.GuildIDs, p.LeftAt)
	return err
}

func (db *Database) ClearDepartures(ctx context.Context, guildIDs []int64) error {
	_, err := db.pool.Exec(ctx, `DELETE FROM guild_departures WHERE guild_id = ANY($1::bigint[])`, guildIDs)
	return err
}

type DeleteDepartedGuildsParams struct {
	GuildIDs   []int64
	LeftBefore time.Time
}

// Every guild-keyed table must be deleted here and listed in GuildIDsWithData.
func (db *Database) DeleteDepartedGuilds(ctx context.Context, p DeleteDepartedGuildsParams) ([]int64, error) {
	rows, err := db.pool.Query(ctx, `
		WITH due AS (
			DELETE FROM guild_departures
			WHERE guild_id = ANY($1::bigint[]) AND left_at < $2
			RETURNING guild_id
		), c AS (
			DELETE FROM customizations WHERE guild_id IN (SELECT guild_id FROM due)
		), e AS (
			DELETE FROM purge_events WHERE guild_id IN (SELECT guild_id FROM due)
		), s AS (
			DELETE FROM guild_settings WHERE guild_id IN (SELECT guild_id FROM due)
		)
		SELECT guild_id FROM due
	`, p.GuildIDs, p.LeftBefore)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[int64])
}
