package commands

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/phx7/albion-killgot/internal/albion"
	"github.com/phx7/albion-killgot/internal/notifier"
	"github.com/phx7/albion-killgot/internal/store"
)

var testCommand = &discordgo.ApplicationCommand{
	Name:                     "test",
	Description:              "Send a test notification to verify bot is working",
	DefaultMemberPermissions: ptrInt64(0),
	Options: []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "kill",
			Description: "Send a test kill notification",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "death",
			Description: "Send a test death notification",
		},
	},
}

func handleTest(s *discordgo.Session, i *discordgo.InteractionCreate, settings *store.SettingsStore) {
	opts := i.ApplicationCommandData().Options
	if len(opts) == 0 {
		replyEphemeral(s, i, "Missing subcommand.")
		return
	}
	isKill := opts[0].Name == "kill"

	gs, err := settings.Get(context.Background(), i.GuildID)
	if err != nil {
		replyEphemeral(s, i, "Failed to load settings.")
		return
	}

	ch := gs.KillsChannel
	if !isKill {
		ch = gs.DeathsChannel
	}
	if ch == "" {
		ch = i.ChannelID
	}

	event := fakeEvent(isKill)
	msg := notifier.EmbedKill(event, gs, isKill)

	_, err = s.ChannelMessageSendComplex(ch, msg)
	if err != nil {
		replyEphemeral(s, i, "Failed to send: "+err.Error())
		return
	}

	kind := "kill"
	if !isKill {
		kind = "death"
	}
	replyEphemeral(s, i, "Test "+kind+" notification sent to <#"+ch+">.")
}

func fakeEvent(isKill bool) albion.Event {
	killer := albion.EventPlayer{
		ID:               "test-killer-id",
		Name:             "TestKiller",
		GuildName:        "Test Guild",
		AverageItemPower: 1200,
	}
	victim := albion.EventPlayer{
		ID:               "test-victim-id",
		Name:             "TestVictim",
		GuildName:        "Enemy Guild",
		AverageItemPower: 1100,
	}
	if !isKill {
		killer, victim = victim, killer
	}
	return albion.Event{
		EventID:             999999999,
		TimeStamp:           time.Now().UTC().Format(time.RFC3339),
		TotalVictimKillFame: 1_234_567,
		Server:              "europe",
		Killer:              killer,
		Victim:              victim,
		Participants:        []albion.EventPlayer{killer},
		LootValue:           &albion.LootValue{Equipment: 500_000, Inventory: 200_000},
	}
}
