package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/phx7/albion-killgot/internal/albion"
	"github.com/phx7/albion-killgot/internal/store"
)


var trackCommand = &discordgo.ApplicationCommand{
	Name:                     "track",
	Description:              "Track a player, guild, or alliance",
	DefaultMemberPermissions: ptrInt64(0),
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "player",
			Description: "Track a player",
			Options:     trackOptions("player"),
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "guild",
			Description: "Track a guild",
			Options:     trackOptions("guild"),
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "alliance",
			Description: "Track an alliance by ID",
			Options:     trackAllianceOptions(),
		},
	},
}

func trackOptions(kind string) []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "server",
			Description: "Albion server",
			Required:    true,
			Choices:     serverChoices(),
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        kind,
			Description: fmt.Sprintf("%s name", strings.Title(kind)),
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionChannel,
			Name:        "kills_channel",
			Description: "Custom channel for kills (overrides default)",
		},
		{
			Type:        discordgo.ApplicationCommandOptionChannel,
			Name:        "deaths_channel",
			Description: "Custom channel for deaths (overrides default)",
		},
	}
}

func trackAllianceOptions() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "server",
			Description: "Albion server",
			Required:    true,
			Choices:     serverChoices(),
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "alliance",
			Description: "Alliance ID (from Albion)",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionChannel,
			Name:        "kills_channel",
			Description: "Custom channel for kills (overrides default)",
		},
		{
			Type:        discordgo.ApplicationCommandOptionChannel,
			Name:        "deaths_channel",
			Description: "Custom channel for deaths (overrides default)",
		},
	}
}

func handleTrack(s *discordgo.Session, i *discordgo.InteractionCreate, tracking *store.TrackingStore, cr CrawlerSyncer) {
	opts := i.ApplicationCommandData().Options
	if len(opts) == 0 {
		replyEphemeral(s, i, "Missing subcommand.")
		return
	}
	sub := opts[0]
	subOpts := optMap(sub.Options)

	serverID := subOpts["server"].StringValue()
	var albionServer albion.Server
	for _, sv := range albion.Servers {
		if sv.ID == serverID {
			albionServer = sv
		}
	}
	if albionServer.ID == "" {
		replyEphemeral(s, i, "Unknown server.")
		return
	}

	// Acknowledge immediately — Albion API search can take >3s

	// Acknowledge immediately — Albion API search can take >3s
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Flags: discordgo.MessageFlagsEphemeral},
	})

	editReply := func(content string) {
		s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{Content: &content})
	}

	ctx := context.Background()
	guildID := i.GuildID

	var entityType store.EntityType
	var entityID, entityName string

	client := albion.NewClient(albionServer)

	switch sub.Name {
	case "player":
		name := subOpts["player"].StringValue()
		results, err := client.Search(ctx, name)
		if err != nil || results == nil {
			editReply(fmt.Sprintf("Search failed: %v", err))
			return
		}
		for _, p := range results.Players {
			if strings.EqualFold(p.Name, name) {
				entityID = p.ID
				entityName = p.Name
				break
			}
		}
		if entityID == "" {
			editReply(fmt.Sprintf("Player `%s` not found on %s.", name, albionServer.Name))
			return
		}
		entityType = store.EntityPlayer

	case "guild":
		name := subOpts["guild"].StringValue()
		results, err := client.Search(ctx, name)
		if err != nil || results == nil {
			editReply(fmt.Sprintf("Search failed: %v", err))
			return
		}
		for _, g := range results.Guilds {
			if strings.EqualFold(g.Name, name) {
				entityID = g.ID
				entityName = g.Name
				break
			}
		}
		if entityID == "" {
			editReply(fmt.Sprintf("Guild `%s` not found on %s.", name, albionServer.Name))
			return
		}
		entityType = store.EntityGuild

	case "alliance":
		allianceID := subOpts["alliance"].StringValue()
		entity, err := client.GetAlliance(ctx, allianceID)
		if err != nil {
			editReply(fmt.Sprintf("Alliance not found: %v", err))
			return
		}
		entityID = entity.AllianceID
		entityName = entity.AllianceTag
		if entityID == "" {
			entityID = allianceID
		}
		entityType = store.EntityAlliance
	}

	var killsCh, deathsCh string
	if opt, ok := subOpts["kills_channel"]; ok {
		killsCh = opt.ChannelValue(s).ID
	}
	if opt, ok := subOpts["deaths_channel"]; ok {
		deathsCh = opt.ChannelValue(s).ID
	}

	err := tracking.Add(ctx, store.TrackedEntity{
		GuildID:       guildID,
		Type:          entityType,
		EntityID:      entityID,
		EntityName:    entityName,
		AlbionServer:  albionServer.ID,
		KillsChannel:  killsCh,
		DeathsChannel: deathsCh,
	})
	if err != nil {
		editReply(fmt.Sprintf("Failed to save: %v", err))
		return
	}

	go cr.Sync(ctx)

	editReply(fmt.Sprintf("Now tracking %s **%s** on %s.", string(entityType), entityName, albionServer.Name))
}
