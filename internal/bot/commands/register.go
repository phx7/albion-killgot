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
func Register(s *discordgo.Session, settings *store.SettingsStore, tracking *store.TrackingStore, cr CrawlerSyncer) {
	cmds := []*discordgo.ApplicationCommand{
		trackCommand,
		untrackCommand,
		listCommand,
		settingsCommand,
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
			handleTrack(s, i, tracking, cr)
		case "untrack":
			handleUntrack(s, i, tracking)
		case "list":
			handleList(s, i, tracking)
		case "settings":
			handleSettings(s, i, settings)
		}
	})
}
