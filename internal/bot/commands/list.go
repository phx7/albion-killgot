package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/phx7/albion-killgot/internal/albion"
	"github.com/phx7/albion-killgot/internal/store"
)

var listCommand = &discordgo.ApplicationCommand{
	Name:                     "list",
	Description:              "Show all tracked players, guilds, and alliances",
	DefaultMemberPermissions: ptrInt64(0),
}

func handleList(s *discordgo.Session, i *discordgo.InteractionCreate, tracking *store.TrackingStore) {
	ctx := context.Background()
	entities, err := tracking.List(ctx, i.GuildID)
	if err != nil {
		replyEphemeral(s, i, "Database error.")
		return
	}

	byType := map[store.EntityType]map[string][]string{
		store.EntityPlayer:   {},
		store.EntityGuild:    {},
		store.EntityAlliance: {},
	}
	for _, sv := range albion.Servers {
		byType[store.EntityPlayer][sv.ID] = nil
		byType[store.EntityGuild][sv.ID] = nil
		byType[store.EntityAlliance][sv.ID] = nil
	}

	for _, e := range entities {
		byType[e.Type][e.AlbionServer] = append(byType[e.Type][e.AlbionServer], e.EntityName)
	}

	buildEmbed := func(t store.EntityType, color int) *discordgo.MessageEmbed {
		count := 0
		for _, sv := range albion.Servers {
			count += len(byType[t][sv.ID])
		}
		embed := &discordgo.MessageEmbed{
			Color: color,
			Title: fmt.Sprintf("Tracked %ss (%d)", string(t), count),
		}
		for _, sv := range albion.Servers {
			names := byType[t][sv.ID]
			value := "None"
			if len(names) > 0 {
				value = strings.Join(names, "\n")
			}
			embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
				Name:   sv.Name,
				Value:  value,
				Inline: true,
			})
		}
		return embed
	}

	embeds := []*discordgo.MessageEmbed{
		buildEmbed(store.EntityPlayer, 0xdcdfdf),
		buildEmbed(store.EntityGuild, 0x57ad65),
		buildEmbed(store.EntityAlliance, 0xed4f4f),
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Flags:  discordgo.MessageFlagsEphemeral,
			Embeds: embeds,
		},
	})
}
