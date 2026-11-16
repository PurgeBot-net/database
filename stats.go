package database

import (
	"context"
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
	var s GuildStats

	if days == 0 {
		s.Monthly = true
		err := db.pool.QueryRow(ctx, `
			SELECT COUNT(*), COALESCE(SUM(deleted), 0), MAX(created_at)
			FROM purge_events
			WHERE guild_id = $1
		`, guildID).Scan(&s.TotalPurges, &s.TotalDeleted, &s.LastPurgeAt)
		if err != nil {
			return s, err
		}
		if err := db.fillMonthlyBars(ctx, guildID, &s); err != nil {
			return s, err
		}
	} else {
		err := db.pool.QueryRow(ctx, `
			SELECT COUNT(*), COALESCE(SUM(deleted), 0), MAX(created_at)
			FROM purge_events
			WHERE guild_id = $1 AND created_at >= NOW() - ($2 * INTERVAL '1 day')
		`, guildID, days).Scan(&s.TotalPurges, &s.TotalDeleted, &s.LastPurgeAt)
		if err != nil {
			return s, err
		}
		if err := db.fillDailyBars(ctx, guildID, days, &s); err != nil {
			return s, err
		}
	}

	return s, nil
}

func (db *Database) fillDailyBars(ctx context.Context, guildID int64, days int, s *GuildStats) error {
	rows, err := db.pool.Query(ctx, `
		SELECT
			DATE_TRUNC('day', created_at AT TIME ZONE 'UTC')::DATE AS day,
			COALESCE(SUM(deleted), 0)                             AS deleted
		FROM purge_events
		WHERE guild_id = $1 AND created_at >= NOW() - ($2 * INTERVAL '1 day')
		GROUP BY 1
		ORDER BY 1
	`, guildID, days)
	if err != nil {
		return err
	}
	defer rows.Close()

	byDate := make(map[string]DailyPurgeStat)
	for rows.Next() {
		var d DailyPurgeStat
		if err := rows.Scan(&d.Date, &d.Deleted); err != nil {
			return err
		}
		byDate[d.Date.Format("2006-01-02")] = d
	}
	if err := rows.Err(); err != nil {
		return err
	}

	now := time.Now().UTC().Truncate(24 * time.Hour)
	s.ChartBars = make([]DailyPurgeStat, days)
	for i := 0; i < days; i++ {
		day := now.AddDate(0, 0, -(days - 1 - i))
		if d, ok := byDate[day.Format("2006-01-02")]; ok {
			s.ChartBars[i] = d
		} else {
			s.ChartBars[i] = DailyPurgeStat{Date: day}
		}
	}
	return nil
}

func (db *Database) fillMonthlyBars(ctx context.Context, guildID int64, s *GuildStats) error {
	rows, err := db.pool.Query(ctx, `
		SELECT
			DATE_TRUNC('month', created_at AT TIME ZONE 'UTC')::DATE AS month,
			COALESCE(SUM(deleted), 0)                                AS deleted
		FROM purge_events
		WHERE guild_id = $1 AND created_at >= NOW() - INTERVAL '12 months'
		GROUP BY 1
		ORDER BY 1
	`, guildID)
	if err != nil {
		return err
	}
	defer rows.Close()

	byMonth := make(map[string]DailyPurgeStat)
	for rows.Next() {
		var d DailyPurgeStat
		if err := rows.Scan(&d.Date, &d.Deleted); err != nil {
			return err
		}
		byMonth[d.Date.Format("2006-01")] = d
	}
	if err := rows.Err(); err != nil {
		return err
	}

	now := time.Now().UTC()
	firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	s.ChartBars = make([]DailyPurgeStat, 12)
	for i := 0; i < 12; i++ {
		month := firstOfMonth.AddDate(0, -(11 - i), 0)
		if d, ok := byMonth[month.Format("2006-01")]; ok {
			s.ChartBars[i] = d
		} else {
			s.ChartBars[i] = DailyPurgeStat{Date: month}
		}
	}
	return nil
}
