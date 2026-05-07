package database

import (
	"context"
	"time"
)

type RecordPurgeEventParams struct {
	GuildID    int64
	PurgeType  string
	TargetType string
	Deleted    int
	DurationMs int
}

func (db *Database) RecordPurgeEvent(ctx context.Context, p RecordPurgeEventParams) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO purge_events (guild_id, purge_type, target_type, deleted, duration_ms)
		VALUES ($1, $2, $3, $4, $5)
	`, p.GuildID, p.PurgeType, p.TargetType, p.Deleted, p.DurationMs)
	return err
}

type GuildStats struct {
	TotalPurges  int
	TotalDeleted int64
	LastPurgeAt  *time.Time
}

func (db *Database) GetGuildStats(ctx context.Context, guildID int64) (GuildStats, error) {
	var s GuildStats
	err := db.pool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(deleted), 0), MAX(created_at)
		FROM purge_events
		WHERE guild_id = $1
	`, guildID).Scan(&s.TotalPurges, &s.TotalDeleted, &s.LastPurgeAt)
	return s, err
}
