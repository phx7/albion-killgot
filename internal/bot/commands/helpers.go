package commands

import (
	"context"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/phx7/albion-killgot/internal/albion"
	"github.com/phx7/albion-killgot/internal/store"
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

// isOwner returns true if the interaction was sent by the guild owner.
func isOwner(s *discordgo.Session, i *discordgo.InteractionCreate) bool {
	guild, err := s.Guild(i.GuildID)
	if err != nil {
		slog.Warn("isOwner: fetch guild", "err", err)
		return false
	}
	return i.Member != nil && i.Member.User.ID == guild.OwnerID
}

// isAuthorized returns true if the user is the guild owner or has an explicit permission entry.
func isAuthorized(s *discordgo.Session, i *discordgo.InteractionCreate, perms *store.PermissionsStore) bool {
	if isOwner(s, i) {
		return true
	}
	if i.Member == nil {
		return false
	}
	allowed, err := perms.IsAllowed(context.Background(), i.GuildID, i.Member.User.ID, i.Member.Roles)
	if err != nil {
		slog.Warn("isAuthorized: db error", "err", err)
		return false
	}
	return allowed
}
