package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type RecordPurgeEventParams struct {
	GuildID    int64
	PurgeType  string
	TargetType string
	Deleted    int
	DurationMs int
}

// RecordPurgeEvent records nothing for a guild that turned stats off.
func (db *Database) RecordPurgeEvent(ctx context.Context, p RecordPurgeEventParams) error {
	return pgx.BeginFunc(ctx, db.pool, func(tx pgx.Tx) error {
		if err := lockGuildStats(ctx, tx, p.GuildID); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO purge_events (guild_id, purge_type, target_type, deleted, duration_ms)
			SELECT $1::bigint, $2::varchar, $3::varchar, $4::int, $5::int
			WHERE NOT EXISTS (
				SELECT 1 FROM guild_settings WHERE guild_id = $1 AND NOT stats_enabled
			)
		`, p.GuildID, p.PurgeType, p.TargetType, p.Deleted, p.DurationMs)
		return err
	})
}

// Orders RecordPurgeEvent and DisableStats per guild, so an opt-out's DELETE never misses a row committed meanwhile.
func lockGuildStats(ctx context.Context, tx pgx.Tx, guildID int64) error {
	_, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock($1::bigint)`, guildID)
	return err
}

type DailyPurgeStat struct {
	Date    time.Time
	Deleted int64
}

type GuildStats struct {
	TotalPurges  int
	TotalDeleted int64
	LastPurgeAt  *time.Time
	ChartBars    []DailyPurgeStat
	Monthly      bool
}

func (db *Database) GetGuildStats(ctx context.Context, guildID int64, days int) (GuildStats, error) {
	s := GuildStats{Monthly: days == 0}
	err := db.pool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(deleted), 0), MAX(created_at)
		FROM purge_events
		WHERE guild_id = $1 AND ($2 = 0 OR created_at >= NOW() - ($2 * INTERVAL '1 day'))
	`, guildID, days).Scan(&s.TotalPurges, &s.TotalDeleted, &s.LastPurgeAt)
	if err != nil {
		return s, err
	}

	unit, n := "day", days
	if s.Monthly {
		unit, n = "month", 12
	}
	rows, err := db.pool.Query(ctx, `
		SELECT
			DATE_TRUNC($2, created_at AT TIME ZONE 'UTC')::DATE AS bar,
			COALESCE(SUM(deleted), 0)                         AS deleted
		FROM purge_events
		WHERE guild_id = $1 AND created_at >= NOW() - $3::interval
		GROUP BY 1
		ORDER BY 1
	`, guildID, unit, fmt.Sprintf("%d %ss", n, unit))
	if err != nil {
		return s, err
	}
	defer rows.Close()

	byDate := make(map[string]DailyPurgeStat)
	for rows.Next() {
		var d DailyPurgeStat
		if err := rows.Scan(&d.Date, &d.Deleted); err != nil {
			return s, err
		}
		byDate[d.Date.Format(time.DateOnly)] = d
	}
	if err := rows.Err(); err != nil {
		return s, err
	}
	s.ChartBars = fillBars(byDate, time.Now().UTC(), n, s.Monthly)
	return s, nil
}

func fillBars(byDate map[string]DailyPurgeStat, now time.Time, n int, monthly bool) []DailyPurgeStat {
	last := now.Truncate(24 * time.Hour)
	if monthly {
		last = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	}
	bars := make([]DailyPurgeStat, n)
	for i := range bars {
		back := -(n - 1 - i)
		bar := last.AddDate(0, 0, back)
		if monthly {
			bar = last.AddDate(0, back, 0)
		}
		bars[i] = DailyPurgeStat{Date: bar}
		if d, ok := byDate[bar.Format(time.DateOnly)]; ok {
			bars[i] = d
		}
	}
	return bars
}
