package commands

import (
	"context"
	"log/slog"

	"github.com/bwmarrin/discordgo"
	"github.com/phx7/albion-killgot/internal/store"
)

// CrawlerSyncer is implemented by crawler.Crawler.
type CrawlerSyncer interface {
	Sync(ctx context.Context)
}

// Register registers all slash commands and their handlers with the session.
func Register(s *discordgo.Session, settings *store.SettingsStore, tracking *store.TrackingStore, perms *store.PermissionsStore, cr CrawlerSyncer) {
	cmds := []*discordgo.ApplicationCommand{
		trackCommand,
		untrackCommand,
		listCommand,
		settingsCommand,
		testCommand,
	}

	registered, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, "", cmds)
	if err != nil {
		slog.Error("register commands", "err", err)
		return
	}
	for _, cmd := range registered {
		slog.Info("registered command", "name", cmd.Name)
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type != discordgo.InteractionApplicationCommand {
			return
		}
		name := i.ApplicationCommandData().Name
		switch name {
		case "track":
			if !isAuthorized(s, i, perms) {
				replyEphemeral(s, i, "You don't have permission to use this command.")
				return
			}
			handleTrack(s, i, tracking, cr)
		case "untrack":
			if !isAuthorized(s, i, perms) {
				replyEphemeral(s, i, "You don't have permission to use this command.")
				return
			}
			handleUntrack(s, i, tracking)
		case "list":
			if !isAuthorized(s, i, perms) {
				replyEphemeral(s, i, "You don't have permission to use this command.")
				return
			}
			handleList(s, i, tracking)
		case "settings":
			handleSettings(s, i, settings, perms)
		case "test":
			if !isAuthorized(s, i, perms) {
				replyEphemeral(s, i, "You don't have permission to use this command.")
				return
			}
			handleTest(s, i, settings)
		}
	})
}
