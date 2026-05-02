package store

import (
	"context"
	"database/sql"
)

type EntityType string

const (
	EntityPlayer   EntityType = "player"
	EntityGuild    EntityType = "guild"
	EntityAlliance EntityType = "alliance"
)

type TrackedEntity struct {
	ID            int64
	GuildID       string
	Type          EntityType
	EntityID      string
	EntityName    string
	AlbionServer  string
	KillsChannel  string
	DeathsChannel string
}

type TrackingStore struct {
	db *sql.DB
}

func NewTrackingStore(db *sql.DB) *TrackingStore {
	return &TrackingStore{db: db}
}

func (s *TrackingStore) List(ctx context.Context, guildID string) ([]TrackedEntity, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, guild_id, entity_type, entity_id, entity_name, albion_server, kills_channel, deaths_channel
FROM tracked_entities WHERE guild_id = ? ORDER BY entity_type, entity_name`, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEntities(rows)
}

func (s *TrackingStore) ListAll(ctx context.Context) ([]TrackedEntity, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, guild_id, entity_type, entity_id, entity_name, albion_server, kills_channel, deaths_channel
FROM tracked_entities`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanEntities(rows)
}

func (s *TrackingStore) Add(ctx context.Context, e TrackedEntity) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO tracked_entities (guild_id, entity_type, entity_id, entity_name, albion_server, kills_channel, deaths_channel)
VALUES (?,?,?,?,?,?,?)
ON CONFLICT(guild_id, entity_type, entity_id) DO UPDATE SET
    entity_name=excluded.entity_name,
    kills_channel=excluded.kills_channel,
    deaths_channel=excluded.deaths_channel`,
		e.GuildID, e.Type, e.EntityID, e.EntityName, e.AlbionServer,
		nullStr(e.KillsChannel), nullStr(e.DeathsChannel),
	)
	return err
}

func (s *TrackingStore) Remove(ctx context.Context, guildID string, t EntityType, entityID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM tracked_entities WHERE guild_id=? AND entity_type=? AND entity_id=?`,
		guildID, t, entityID)
	return err
}

func (s *TrackingStore) Count(ctx context.Context, guildID string, t EntityType) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM tracked_entities WHERE guild_id=? AND entity_type=?`, guildID, t).Scan(&n)
	return n, err
}

// ActiveServers returns distinct albion servers that have at least one tracked entity.
func (s *TrackingStore) ActiveServers(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT DISTINCT albion_server FROM tracked_entities`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var servers []string
	for rows.Next() {
		var sv string
		if err := rows.Scan(&sv); err != nil {
			return nil, err
		}
		servers = append(servers, sv)
	}
	return servers, rows.Err()
}

func scanEntities(rows *sql.Rows) ([]TrackedEntity, error) {
	var out []TrackedEntity
	for rows.Next() {
		var e TrackedEntity
		var killsCh, deathsCh sql.NullString
		if err := rows.Scan(&e.ID, &e.GuildID, &e.Type, &e.EntityID, &e.EntityName,
			&e.AlbionServer, &killsCh, &deathsCh); err != nil {
			return nil, err
		}
		e.KillsChannel = killsCh.String
		e.DeathsChannel = deathsCh.String
		out = append(out, e)
	}
	return out, rows.Err()
}
