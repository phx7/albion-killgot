package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/phx7/albion-killgot/internal/albion"
)

func serverChoices() []*discordgo.ApplicationCommandOptionChoice {
	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, len(albion.Servers))
	for _, s := range albion.Servers {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  s.Name,
			Value: s.ID,
		})
	}
	return choices
}

func optMap(opts []*discordgo.ApplicationCommandInteractionDataOption) map[string]*discordgo.ApplicationCommandInteractionDataOption {
	m := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(opts))
	for _, o := range opts {
		m[o.Name] = o
	}
	return m
}

func replyEphemeral(s *discordgo.Session, i *discordgo.InteractionCreate, content string) {
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func ptrInt64(v int64) *int64 { return &v }
