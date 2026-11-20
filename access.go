package database

import (
	"context"

	"github.com/jackc/pgx/v5"
)

func (db *Database) GetAccessRoles(ctx context.Context, guildID int64) ([]int64, error) {
	rows, err := db.pool.Query(ctx, `SELECT role_id FROM access_roles WHERE guild_id = $1 ORDER BY role_id`, guildID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[int64])
}

type AccessRoleParams struct {
	GuildID int64
	RoleID  int64
}

func (db *Database) AddAccessRole(ctx context.Context, p AccessRoleParams) error {
	_, err := db.pool.Exec(ctx, `
		INSERT INTO access_roles (guild_id, role_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, p.GuildID, p.RoleID)
	return err
}

func (db *Database) RemoveAccessRole(ctx context.Context, p AccessRoleParams) error {
	_, err := db.pool.Exec(ctx, `DELETE FROM access_roles WHERE guild_id = $1 AND role_id = $2`, p.GuildID, p.RoleID)
	return err
}
