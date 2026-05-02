package notifier

import (
	"context"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/phx7/albion-killgot/internal/albion"
	"github.com/phx7/albion-killgot/internal/store"
)

const (
	juicyMinFame      = 2_000_000
	juicyGoodLoot     = 15_000_000
	juicyInsaneLoot   = 30_000_000
)

type Notifier struct {
	session  *discordgo.Session
	settings *store.SettingsStore
	tracking *store.TrackingStore
}

func New(session *discordgo.Session, settings *store.SettingsStore, tracking *store.TrackingStore) *Notifier {
	return &Notifier{session: session, settings: settings, tracking: tracking}
}

func (n *Notifier) HandleEvent(ctx context.Context, event albion.Event) {
	if event.TotalVictimKillFame <= 0 {
		return
	}

	entities, err := n.tracking.ListAll(ctx)
	if err != nil {
		slog.Error("list tracking", "err", err)
		return
	}

	// Group tracked entities by guild
	byGuild := make(map[string][]store.TrackedEntity)
	for _, e := range entities {
		if e.AlbionServer == event.Server {
			byGuild[e.GuildID] = append(byGuild[e.GuildID], e)
		}
	}

	for guildID, tracked := range byGuild {
		gs, err := n.settings.Get(ctx, guildID)
		if err != nil {
			slog.Error("get settings", "guild", guildID, "err", err)
			continue
		}

		matched, isKill, channel := matchEvent(event, tracked, gs)
		if matched {
			msg := EmbedKill(event, gs, isKill)
			n.send(guildID, channel, msg)
		}

		// Juicy kills — separate channel, no tracking required
		if tier, juicyCh := juicyTier(event, gs); tier != "" && juicyCh != "" {
			msg := EmbedJuicy(event, gs, tier)
			n.send(guildID, juicyCh, msg)
		}
	}
}

func (n *Notifier) HandleBattle(ctx context.Context, battle albion.Battle) {
	if battle.TotalFame <= 0 {
		return
	}

	entities, err := n.tracking.ListAll(ctx)
	if err != nil {
		slog.Error("list tracking", "err", err)
		return
	}

	byGuild := make(map[string][]store.TrackedEntity)
	for _, e := range entities {
		if e.AlbionServer == battle.Server {
			byGuild[e.GuildID] = append(byGuild[e.GuildID], e)
		}
	}

	for guildID, tracked := range byGuild {
		gs, err := n.settings.Get(ctx, guildID)
		if err != nil {
			slog.Error("get settings", "guild", guildID, "err", err)
			continue
		}

		if !gs.BattlesEnabled || gs.BattlesChannel == "" {
			continue
		}
		if !battleMatches(battle, tracked) {
			continue
		}
		if !battleMeetsThreshold(battle, gs) {
			continue
		}

		msg := EmbedBattle(battle, gs)
		n.send(guildID, gs.BattlesChannel, msg)
	}
}

func (n *Notifier) send(guildID, channelID string, msg *discordgo.MessageSend) {
	_, err := n.session.ChannelMessageSendComplex(channelID, msg)
	if err != nil {
		slog.Error("send message", "guild", guildID, "channel", channelID, "err", err)
	}
}

// matchEvent checks if the event involves any tracked entity and returns
// the channel to send to and whether it's a kill (true) or death (false).
func matchEvent(event albion.Event, tracked []store.TrackedEntity, gs store.GuildSettings) (matched bool, isKill bool, channel string) {
	killer := event.Killer
	victim := event.Victim

	for _, t := range tracked {
		// Check if killer is tracked -> it's a kill for us
		if matchesPlayer(killer, t) {
			if !gs.KillsEnabled {
				continue
			}
			ch := gs.KillsChannel
			if t.KillsChannel != "" {
				ch = t.KillsChannel
			}
			if ch == "" {
				continue
			}
			return true, true, ch
		}
		// Check if victim is tracked -> it's a death for us
		if matchesPlayer(victim, t) {
			if !gs.DeathsEnabled {
				continue
			}
			ch := gs.DeathsChannel
			if t.DeathsChannel != "" {
				ch = t.DeathsChannel
			}
			if ch == "" {
				continue
			}
			return true, false, ch
		}
	}
	return false, false, ""
}

func matchesPlayer(p albion.EventPlayer, t store.TrackedEntity) bool {
	switch t.Type {
	case store.EntityPlayer:
		return p.ID == t.EntityID
	case store.EntityGuild:
		return p.GuildID == t.EntityID
	case store.EntityAlliance:
		return p.AllianceID == t.EntityID
	}
	return false
}

func battleMatches(battle albion.Battle, tracked []store.TrackedEntity) bool {
	for _, t := range tracked {
		switch t.Type {
		case store.EntityPlayer:
			if _, ok := battle.Players[t.EntityID]; ok {
				return true
			}
		case store.EntityGuild:
			if _, ok := battle.Guilds[t.EntityID]; ok {
				return true
			}
		case store.EntityAlliance:
			if _, ok := battle.Alliances[t.EntityID]; ok {
				return true
			}
		}
	}
	return false
}

func battleMeetsThreshold(battle albion.Battle, gs store.GuildSettings) bool {
	if gs.BattlesMinPlayers > 0 && len(battle.Players) < gs.BattlesMinPlayers {
		return false
	}
	if gs.BattlesMinGuilds > 0 && len(battle.Guilds) < gs.BattlesMinGuilds {
		return false
	}
	return true
}

func juicyTier(event albion.Event, gs store.GuildSettings) (tier, channel string) {
	juicyEnabled := false
	switch event.Server {
	case "americas":
		juicyEnabled = gs.JuicyEnabledAmericas
	case "asia":
		juicyEnabled = gs.JuicyEnabledAsia
	case "europe":
		juicyEnabled = gs.JuicyEnabledEurope
	}
	if !juicyEnabled {
		return "", ""
	}
	if event.TotalVictimKillFame < juicyMinFame {
		return "", ""
	}

	loot := int64(0)
	if event.LootValue != nil {
		loot = event.LootValue.Equipment + event.LootValue.Inventory
	}

	if loot >= juicyInsaneLoot {
		return "insane", gs.JuicyInsaneChannel
	}
	if loot >= juicyGoodLoot {
		return "good", gs.JuicyGoodChannel
	}
	return "", ""
}
