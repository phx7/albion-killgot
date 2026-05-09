package store

import (
	"context"
	"database/sql"
)

type Permission struct {
	PermissionID int64
	GuildID      string
	RoleID       string
	UserID       string
}

type PermissionsStore struct {
	db *sql.DB
}

func NewPermissionsStore(db *sql.DB) *PermissionsStore {
	return &PermissionsStore{db: db}
}

func (s *PermissionsStore) Grant(ctx context.Context, guildID, roleID, userID string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO permissions (guild_id, role_id, user_id) VALUES (?,?,?)`,
		guildID, nullStr(roleID), nullStr(userID))
	return err
}

func (s *PermissionsStore) Revoke(ctx context.Context, guildID, roleID, userID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM permissions WHERE guild_id=? AND (role_id IS ? OR role_id=?) AND (user_id IS ? OR user_id=?)`,
		guildID, nullStr(roleID), nullStr(roleID), nullStr(userID), nullStr(userID))
	return err
}

func (s *PermissionsStore) List(ctx context.Context, guildID string) ([]Permission, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT permission_id, guild_id, COALESCE(role_id,''), COALESCE(user_id,'') FROM permissions WHERE guild_id=? ORDER BY permission_id`,
		guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Permission
	for rows.Next() {
		var p Permission
		if err := rows.Scan(&p.PermissionID, &p.GuildID, &p.RoleID, &p.UserID); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// IsAllowed returns true if the user has an explicit permission (by user_id or any of their role_ids).
func (s *PermissionsStore) IsAllowed(ctx context.Context, guildID, userID string, roleIDs []string) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM permissions WHERE guild_id=? AND user_id=?`,
		guildID, userID).Scan(&n)
	if err != nil {
		return false, err
	}
	if n > 0 {
		return true, nil
	}

	for _, roleID := range roleIDs {
		err = s.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM permissions WHERE guild_id=? AND role_id=?`,
			guildID, roleID).Scan(&n)
		if err != nil {
			return false, err
		}
		if n > 0 {
			return true, nil
		}
	}
	return false, nil
}
