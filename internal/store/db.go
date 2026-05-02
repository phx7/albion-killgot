package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	path = absPath(path)
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if _, err := db.Exec(`PRAGMA journal_mode=WAL`); err != nil {
		return nil, fmt.Errorf("pragma journal_mode: %w", err)
	}
	if _, err := db.Exec(`PRAGMA foreign_keys=ON`); err != nil {
		return nil, fmt.Errorf("pragma foreign_keys: %w", err)
	}
	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

// absPath returns an absolute path. Relative paths are resolved from the
// directory of the running executable so the db file ends up next to it.
func absPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	if exe, err := os.Executable(); err == nil {
		return filepath.Join(filepath.Dir(exe), path)
	}
	return path
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS guild_settings (
    guild_id              TEXT PRIMARY KEY,
    kills_enabled         INTEGER NOT NULL DEFAULT 1,
    kills_channel         TEXT,
    deaths_enabled        INTEGER NOT NULL DEFAULT 1,
    deaths_channel        TEXT,
    battles_enabled       INTEGER NOT NULL DEFAULT 1,
    battles_channel       TEXT,
    battles_min_players   INTEGER NOT NULL DEFAULT 0,
    battles_min_guilds    INTEGER NOT NULL DEFAULT 0,
    juicy_enabled_americas INTEGER NOT NULL DEFAULT 0,
    juicy_enabled_asia    INTEGER NOT NULL DEFAULT 0,
    juicy_enabled_europe  INTEGER NOT NULL DEFAULT 0,
    juicy_good_channel    TEXT,
    juicy_insane_channel  TEXT,
    provider              TEXT NOT NULL DEFAULT 'albion-killboard',
    guild_tags            INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS tracked_entities (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    guild_id       TEXT NOT NULL,
    entity_type    TEXT NOT NULL CHECK(entity_type IN ('player','guild','alliance')),
    entity_id      TEXT NOT NULL,
    entity_name    TEXT NOT NULL,
    albion_server  TEXT NOT NULL CHECK(albion_server IN ('americas','asia','europe')),
    kills_channel  TEXT,
    deaths_channel TEXT,
    UNIQUE(guild_id, entity_type, entity_id)
);
`)
	return err
}
