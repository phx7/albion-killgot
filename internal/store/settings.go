package store

import (
	"context"
	"database/sql"
)

type GuildSettings struct {
	GuildID              string
	KillsEnabled         bool
	KillsChannel         string
	DeathsEnabled        bool
	DeathsChannel        string
	BattlesEnabled       bool
	BattlesChannel       string
	BattlesMinPlayers    int
	BattlesMinGuilds     int
	JuicyEnabledAmericas bool
	JuicyEnabledAsia     bool
	JuicyEnabledEurope   bool
	JuicyGoodChannel     string
	JuicyInsaneChannel   string
	Provider             string
	GuildTags            bool
}

func defaultSettings(guildID string) GuildSettings {
	return GuildSettings{
		GuildID:       guildID,
		KillsEnabled:  true,
		DeathsEnabled: true,
		Provider:      "albion-killboard",
	}
}

type SettingsStore struct {
	db *sql.DB
}

func NewSettingsStore(db *sql.DB) *SettingsStore {
	return &SettingsStore{db: db}
}

func (s *SettingsStore) Get(ctx context.Context, guildID string) (GuildSettings, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT guild_id, kills_enabled, kills_channel, deaths_enabled, deaths_channel,
       battles_enabled, battles_channel, battles_min_players, battles_min_guilds,
       juicy_enabled_americas, juicy_enabled_asia, juicy_enabled_europe,
       juicy_good_channel, juicy_insane_channel, provider, guild_tags
FROM guild_settings WHERE guild_id = ?`, guildID)

	var gs GuildSettings
	var killsCh, deathsCh, battlesCh, juicyGoodCh, juicyInsaneCh sql.NullString
	err := row.Scan(
		&gs.GuildID,
		&gs.KillsEnabled, &killsCh,
		&gs.DeathsEnabled, &deathsCh,
		&gs.BattlesEnabled, &battlesCh, &gs.BattlesMinPlayers, &gs.BattlesMinGuilds,
		&gs.JuicyEnabledAmericas, &gs.JuicyEnabledAsia, &gs.JuicyEnabledEurope,
		&juicyGoodCh, &juicyInsaneCh,
		&gs.Provider, &gs.GuildTags,
	)
	if err == sql.ErrNoRows {
		return defaultSettings(guildID), nil
	}
	if err != nil {
		return gs, err
	}
	gs.KillsChannel = killsCh.String
	gs.DeathsChannel = deathsCh.String
	gs.BattlesChannel = battlesCh.String
	gs.JuicyGoodChannel = juicyGoodCh.String
	gs.JuicyInsaneChannel = juicyInsaneCh.String
	return gs, nil
}

func (s *SettingsStore) Save(ctx context.Context, gs GuildSettings) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO guild_settings (
    guild_id, kills_enabled, kills_channel, deaths_enabled, deaths_channel,
    battles_enabled, battles_channel, battles_min_players, battles_min_guilds,
    juicy_enabled_americas, juicy_enabled_asia, juicy_enabled_europe,
    juicy_good_channel, juicy_insane_channel, provider, guild_tags
) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
ON CONFLICT(guild_id) DO UPDATE SET
    kills_enabled=excluded.kills_enabled,
    kills_channel=excluded.kills_channel,
    deaths_enabled=excluded.deaths_enabled,
    deaths_channel=excluded.deaths_channel,
    battles_enabled=excluded.battles_enabled,
    battles_channel=excluded.battles_channel,
    battles_min_players=excluded.battles_min_players,
    battles_min_guilds=excluded.battles_min_guilds,
    juicy_enabled_americas=excluded.juicy_enabled_americas,
    juicy_enabled_asia=excluded.juicy_enabled_asia,
    juicy_enabled_europe=excluded.juicy_enabled_europe,
    juicy_good_channel=excluded.juicy_good_channel,
    juicy_insane_channel=excluded.juicy_insane_channel,
    provider=excluded.provider,
    guild_tags=excluded.guild_tags`,
		gs.GuildID,
		gs.KillsEnabled, nullStr(gs.KillsChannel),
		gs.DeathsEnabled, nullStr(gs.DeathsChannel),
		gs.BattlesEnabled, nullStr(gs.BattlesChannel), gs.BattlesMinPlayers, gs.BattlesMinGuilds,
		gs.JuicyEnabledAmericas, gs.JuicyEnabledAsia, gs.JuicyEnabledEurope,
		nullStr(gs.JuicyGoodChannel), nullStr(gs.JuicyInsaneChannel),
		gs.Provider, gs.GuildTags,
	)
	return err
}

func nullStr(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
