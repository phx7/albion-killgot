package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/phx7/albion-killgot/internal/notifier"
	"github.com/phx7/albion-killgot/internal/store"
)

var settingsCommand = &discordgo.ApplicationCommand{
	Name:                     "settings",
	Description:              "Configure bot settings",
	DefaultMemberPermissions: ptrInt64(0),
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "kills",
			Description: "Configure kill notifications",
			Options:     eventSettingOptions(),
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "deaths",
			Description: "Configure death notifications",
			Options:     eventSettingOptions(),
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "battles",
			Description: "Configure battle notifications",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionBoolean, Name: "enabled", Description: "Enable or disable"},
				{Type: discordgo.ApplicationCommandOptionChannel, Name: "channel", Description: "Channel to post to"},
				{Type: discordgo.ApplicationCommandOptionInteger, Name: "min_players", Description: "Minimum players threshold"},
				{Type: discordgo.ApplicationCommandOptionInteger, Name: "min_guilds", Description: "Minimum guilds threshold"},
				providerOption(true),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "juicy",
			Description: "Configure juicy kill notifications",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionBoolean, Name: "americas", Description: "Enable for Americas server"},
				{Type: discordgo.ApplicationCommandOptionBoolean, Name: "asia", Description: "Enable for Asia server"},
				{Type: discordgo.ApplicationCommandOptionBoolean, Name: "europe", Description: "Enable for Europe server"},
				{Type: discordgo.ApplicationCommandOptionChannel, Name: "good_channel", Description: "Channel for good kills (15M+ loot)"},
				{Type: discordgo.ApplicationCommandOptionChannel, Name: "insane_channel", Description: "Channel for insane kills (30M+ loot)"},
				providerOption(false),
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "show",
			Description: "Show current settings",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "permit",
			Description: "Grant a role or user permission to use bot commands",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionRole, Name: "role", Description: "Role to grant"},
				{Type: discordgo.ApplicationCommandOptionUser, Name: "user", Description: "User to grant"},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "revoke",
			Description: "Revoke a role or user permission to use bot commands",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionRole, Name: "role", Description: "Role to revoke"},
				{Type: discordgo.ApplicationCommandOptionUser, Name: "user", Description: "User to revoke"},
			},
		},
	},
}

func eventSettingOptions() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{Type: discordgo.ApplicationCommandOptionBoolean, Name: "enabled", Description: "Enable or disable"},
		{Type: discordgo.ApplicationCommandOptionChannel, Name: "channel", Description: "Channel to post to"},
		providerOption(false),
	}
}

func providerOption(battlesOnly bool) *discordgo.ApplicationCommandOption {
	var choices []*discordgo.ApplicationCommandOptionChoice
	for _, p := range notifier.Providers {
		if battlesOnly && p.BattleURL == nil {
			continue
		}
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  p.Name,
			Value: p.ID,
		})
	}
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        "provider",
		Description: "Link provider for kill/battle reports",
		Choices:     choices,
	}
}

func handleSettings(s *discordgo.Session, i *discordgo.InteractionCreate, settings *store.SettingsStore, perms *store.PermissionsStore) {
	if !isOwner(s, i) {
		replyEphemeral(s, i, "Only the server owner can change settings.")
		return
	}

	opts := i.ApplicationCommandData().Options
	if len(opts) == 0 {
		replyEphemeral(s, i, "Missing subcommand.")
		return
	}
	sub := opts[0]
	ctx := context.Background()
	guildID := i.GuildID

	gs, err := settings.Get(ctx, guildID)
	if err != nil {
		replyEphemeral(s, i, "Failed to load settings.")
		return
	}

	subOpts := optMap(sub.Options)

	switch sub.Name {
	case "kills":
		if opt, ok := subOpts["enabled"]; ok {
			gs.KillsEnabled = opt.BoolValue()
		}
		if opt, ok := subOpts["channel"]; ok {
			gs.KillsChannel = opt.ChannelValue(s).ID
		}
		if opt, ok := subOpts["provider"]; ok {
			gs.Provider = opt.StringValue()
		}

	case "deaths":
		if opt, ok := subOpts["enabled"]; ok {
			gs.DeathsEnabled = opt.BoolValue()
		}
		if opt, ok := subOpts["channel"]; ok {
			gs.DeathsChannel = opt.ChannelValue(s).ID
		}
		if opt, ok := subOpts["provider"]; ok {
			gs.Provider = opt.StringValue()
		}

	case "battles":
		if opt, ok := subOpts["enabled"]; ok {
			gs.BattlesEnabled = opt.BoolValue()
		}
		if opt, ok := subOpts["channel"]; ok {
			gs.BattlesChannel = opt.ChannelValue(s).ID
		}
		if opt, ok := subOpts["min_players"]; ok {
			gs.BattlesMinPlayers = int(opt.IntValue())
		}
		if opt, ok := subOpts["min_guilds"]; ok {
			gs.BattlesMinGuilds = int(opt.IntValue())
		}
		if opt, ok := subOpts["provider"]; ok {
			gs.Provider = opt.StringValue()
		}

	case "juicy":
		if opt, ok := subOpts["americas"]; ok {
			gs.JuicyEnabledAmericas = opt.BoolValue()
		}
		if opt, ok := subOpts["asia"]; ok {
			gs.JuicyEnabledAsia = opt.BoolValue()
		}
		if opt, ok := subOpts["europe"]; ok {
			gs.JuicyEnabledEurope = opt.BoolValue()
		}
		if opt, ok := subOpts["good_channel"]; ok {
			gs.JuicyGoodChannel = opt.ChannelValue(s).ID
		}
		if opt, ok := subOpts["insane_channel"]; ok {
			gs.JuicyInsaneChannel = opt.ChannelValue(s).ID
		}
		if opt, ok := subOpts["provider"]; ok {
			gs.Provider = opt.StringValue()
		}

	case "show":
		replyEphemeral(s, i, formatSettings(gs))
		return

	case "permit", "revoke":
		subOpts := optMap(sub.Options)
		var roleID, userID string
		if opt, ok := subOpts["role"]; ok {
			roleID = opt.RoleValue(s, guildID).ID
		}
		if opt, ok := subOpts["user"]; ok {
			userID = opt.UserValue(s).ID
		}
		if roleID == "" && userID == "" {
			replyEphemeral(s, i, "Specify a role or user.")
			return
		}
		var err error
		if sub.Name == "permit" {
			err = perms.Grant(ctx, guildID, roleID, userID)
		} else {
			err = perms.Revoke(ctx, guildID, roleID, userID)
		}
		if err != nil {
			replyEphemeral(s, i, fmt.Sprintf("Failed: %v", err))
			return
		}
		action := "granted"
		if sub.Name == "revoke" {
			action = "revoked"
		}
		target := ""
		if roleID != "" {
			target = "<@&" + roleID + ">"
		} else {
			target = "<@" + userID + ">"
		}
		replyEphemeral(s, i, fmt.Sprintf("Permission %s for %s.", action, target))
		return
	}

	if err := settings.Save(ctx, gs); err != nil {
		replyEphemeral(s, i, fmt.Sprintf("Failed to save settings: %v", err))
		return
	}
	replyEphemeral(s, i, "Settings saved.\n\n"+formatSettings(gs))
}

func formatSettings(gs store.GuildSettings) string {
	var b strings.Builder
	onoff := func(v bool) string {
		if v {
			return "enabled"
		}
		return "disabled"
	}
	ch := func(id string) string {
		if id == "" {
			return "not set"
		}
		return "<#" + id + ">"
	}

	fmt.Fprintf(&b, "**Kills:** %s, channel: %s\n", onoff(gs.KillsEnabled), ch(gs.KillsChannel))
	fmt.Fprintf(&b, "**Deaths:** %s, channel: %s\n", onoff(gs.DeathsEnabled), ch(gs.DeathsChannel))
	fmt.Fprintf(&b, "**Battles:** %s, channel: %s, min players: %d, min guilds: %d\n",
		onoff(gs.BattlesEnabled), ch(gs.BattlesChannel), gs.BattlesMinPlayers, gs.BattlesMinGuilds)
	fmt.Fprintf(&b, "**Juicy kills:** americas=%s asia=%s europe=%s\n",
		onoff(gs.JuicyEnabledAmericas), onoff(gs.JuicyEnabledAsia), onoff(gs.JuicyEnabledEurope))
	fmt.Fprintf(&b, "  good channel: %s, insane channel: %s\n", ch(gs.JuicyGoodChannel), ch(gs.JuicyInsaneChannel))
	fmt.Fprintf(&b, "**Provider:** %s\n", gs.Provider)
	fmt.Fprintf(&b, "**Guild tags:** %s\n", onoff(gs.GuildTags))
	return b.String()
}
