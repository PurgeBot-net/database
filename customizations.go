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
}

func (db *Database) GetCustomization(ctx context.Context, guildID int64) (*Customization, error) {
	var c Customization
	err := db.pool.QueryRow(ctx, `
		SELECT guild_id, bot_name, bot_avatar, remove_branding, updated_at
		FROM customizations WHERE guild_id = $1
	`, guildID).Scan(&c.GuildID, &c.BotName, &c.BotAvatar, &c.RemoveBranding, &c.UpdatedAt)
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
}

func (db *Database) DeleteCustomization(ctx context.Context, guildID int64) error {
	_, err := db.pool.Exec(ctx, `DELETE FROM customizations WHERE guild_id = $1`, guildID)
	return err
}

func (db *Database) UpsertCustomization(ctx context.Context, p UpsertCustomizationParams) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO customizations (guild_id, bot_name, bot_avatar, remove_branding, updated_at)
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT (guild_id) DO UPDATE SET
			bot_name        = EXCLUDED.bot_name,
			bot_avatar      = EXCLUDED.bot_avatar,
			remove_branding = EXCLUDED.remove_branding,
			updated_at      = now()
	`, p.GuildID, p.BotName, p.BotAvatar, p.RemoveBranding)
	return err
}
