package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/phx7/albion-killgot/internal/store"
)

var untrackCommand = &discordgo.ApplicationCommand{
	Name:                     "untrack",
	Description:              "Stop tracking a player, guild, or alliance",
	DefaultMemberPermissions: ptrInt64(0),
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "player",
			Description: "Untrack a player",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionString, Name: "server", Description: "Albion server", Required: true, Choices: serverChoices()},
				{Type: discordgo.ApplicationCommandOptionString, Name: "player", Description: "Player name", Required: true},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "guild",
			Description: "Untrack a guild",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionString, Name: "server", Description: "Albion server", Required: true, Choices: serverChoices()},
				{Type: discordgo.ApplicationCommandOptionString, Name: "guild", Description: "Guild name", Required: true},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "alliance",
			Description: "Untrack an alliance",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionString, Name: "server", Description: "Albion server", Required: true, Choices: serverChoices()},
				{Type: discordgo.ApplicationCommandOptionString, Name: "alliance", Description: "Alliance name or ID", Required: true},
			},
		},
	},
}

func handleUntrack(s *discordgo.Session, i *discordgo.InteractionCreate, tracking *store.TrackingStore) {
	opts := i.ApplicationCommandData().Options
	if len(opts) == 0 {
		replyEphemeral(s, i, "Missing subcommand.")
		return
	}
	sub := opts[0]
	subOpts := optMap(sub.Options)

	serverID := subOpts["server"].StringValue()
	ctx := context.Background()
	guildID := i.GuildID

	var entityType store.EntityType
	var nameOrID string

	switch sub.Name {
	case "player":
		entityType = store.EntityPlayer
		nameOrID = subOpts["player"].StringValue()
	case "guild":
		entityType = store.EntityGuild
		nameOrID = subOpts["guild"].StringValue()
	case "alliance":
		entityType = store.EntityAlliance
		nameOrID = subOpts["alliance"].StringValue()
	}

	entities, err := tracking.List(ctx, guildID)
	if err != nil {
		replyEphemeral(s, i, "Database error.")
		return
	}

	var target *store.TrackedEntity
	for idx := range entities {
		e := &entities[idx]
		if e.Type != entityType || e.AlbionServer != serverID {
			continue
		}
		if strings.EqualFold(e.EntityName, nameOrID) || e.EntityID == nameOrID {
			target = e
			break
		}
	}

	if target == nil {
		replyEphemeral(s, i, fmt.Sprintf("Not found: `%s` on that server.", nameOrID))
		return
	}

	if err := tracking.Remove(ctx, guildID, entityType, target.EntityID); err != nil {
		replyEphemeral(s, i, fmt.Sprintf("Failed to remove: %v", err))
		return
	}

	replyEphemeral(s, i, fmt.Sprintf("Stopped tracking %s **%s**.", string(entityType), target.EntityName))
}
