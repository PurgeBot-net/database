package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

type Customization struct {
	GuildID        int64
	BotName        *string
	BotAvatar      *string
	RemoveBranding bool
	UpdatedAt      time.Time
	UpdatedBy      int64
}

func (db *Database) GetCustomization(ctx context.Context, guildID int64) (*Customization, error) {
	var c Customization
	err := db.pool.QueryRow(ctx, `
		SELECT guild_id, bot_name, bot_avatar, remove_branding, updated_at, updated_by
		FROM customizations WHERE guild_id = $1
	`, guildID).Scan(&c.GuildID, &c.BotName, &c.BotAvatar, &c.RemoveBranding, &c.UpdatedAt, &c.UpdatedBy)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &c, err
}

type UpsertCustomizationParams struct {
	GuildID        int64
	BotName        *string
	BotAvatar      *string
	RemoveBranding bool
	UpdatedBy      int64
}

func (db *Database) DeleteCustomization(ctx context.Context, guildID int64) error {
	_, err := db.pool.Exec(ctx, `DELETE FROM customizations WHERE guild_id = $1`, guildID)
	return err
}

func (db *Database) UpsertCustomization(ctx context.Context, p UpsertCustomizationParams) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO customizations (guild_id, bot_name, bot_avatar, remove_branding, updated_at, updated_by)
		VALUES ($1, $2, $3, $4, now(), $5)
		ON CONFLICT (guild_id) DO UPDATE SET
			bot_name        = EXCLUDED.bot_name,
			bot_avatar      = EXCLUDED.bot_avatar,
			remove_branding = EXCLUDED.remove_branding,
			updated_at      = now(),
			updated_by      = EXCLUDED.updated_by
	`, p.GuildID, p.BotName, p.BotAvatar, p.RemoveBranding, p.UpdatedBy)
	return err
}
