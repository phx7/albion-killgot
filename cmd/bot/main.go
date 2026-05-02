package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/phx7/albion-killgot/internal/bot/commands"
	"github.com/phx7/albion-killgot/internal/config"
	"github.com/phx7/albion-killgot/internal/crawler"
	"github.com/phx7/albion-killgot/internal/notifier"
	"github.com/phx7/albion-killgot/internal/store"
)

// version is set at build time via -ldflags "-X main.version=v0.1.0"
var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-version" {
		fmt.Println("albion-killgot", version)
		return
	}
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	db, err := store.Open(cfg.DBPath)
	if err != nil {
		slog.Error("open db", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	settings := store.NewSettingsStore(db)
	tracking := store.NewTrackingStore(db)

	session, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		slog.Error("create discord session", "err", err)
		os.Exit(1)
	}
	session.Identify.Intents = discordgo.IntentsGuilds

	if err := session.Open(); err != nil {
		slog.Error("open discord session", "err", err)
		os.Exit(1)
	}
	defer session.Close()

	slog.Info("bot connected", "user", session.State.User.Username, "version", version)

	ntf := notifier.New(session, settings, tracking)
	cr := crawler.New(tracking.ActiveServers)

	commands.Register(session, settings, tracking, cr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cr.Run(ctx)

	go func() {
		for event := range cr.Events {
			ntf.HandleEvent(ctx, event)
		}
	}()

	go func() {
		for battle := range cr.Battles {
			ntf.HandleBattle(ctx, battle)
		}
	}()

	slog.Info("bot running, press Ctrl+C to stop")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	slog.Info("shutting down")
}

