package notifier

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/phx7/albion-killgot/internal/albion"
	"github.com/phx7/albion-killgot/internal/store"
)

const (
	colorGrey     = 0xdcdfdf
	colorGold     = 0xF1C40F
	colorYellow   = 0xFFFF00
	colorDarkGreen = 52224
	colorRed      = 13369344
	colorBattle   = 16752981

	maxFieldValue = 1024
	maxFields     = 25
)

func fmtNumber(n int64) string {
	if n == 0 {
		return "0"
	}
	s := fmt.Sprintf("%d", n)
	out := make([]byte, 0, len(s)+len(s)/3)
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, byte(c))
	}
	return string(out)
}

func playerLabel(p albion.EventPlayer, guildTags bool) string {
	var b strings.Builder
	if guildTags && p.GuildName != "" {
		b.WriteString("[")
		b.WriteString(p.GuildName)
		b.WriteString("] ")
	}
	b.WriteString(p.Name)
	return b.String()
}

func guildField(p albion.EventPlayer) string {
	if p.GuildName == "" {
		return "No guild"
	}
	if p.AllianceName != "" {
		return fmt.Sprintf("[%s] %s", p.AllianceName, p.GuildName)
	}
	return p.GuildName
}

// EmbedKill builds a text embed for a kill or death event.
func EmbedKill(event albion.Event, gs store.GuildSettings, isKill bool) *discordgo.MessageSend {
	lootSum := int64(0)
	if event.LootValue != nil {
		lootSum = event.LootValue.Equipment + event.LootValue.Inventory
	}

	color := colorGrey
	if isKill {
		color = colorDarkGreen
	} else {
		color = colorRed
	}

	killerLabel := playerLabel(event.Killer, gs.GuildTags)
	victimLabel := playerLabel(event.Victim, gs.GuildTags)

	var title string
	if isKill {
		title = fmt.Sprintf("%s killed %s", killerLabel, victimLabel)
	} else {
		title = fmt.Sprintf("%s was killed by %s", victimLabel, killerLabel)
	}

	var desc string
	if len(event.Participants) == 1 {
		desc = "Solo kill!"
	} else {
		var assists []string
		totalDmg := 0.0
		for _, p := range event.Participants {
			totalDmg += p.DamageDone
		}
		for _, p := range event.Participants {
			if p.ID == event.Victim.ID {
				continue
			}
			pct := 0
			if totalDmg > 0 {
				pct = int(math.Round(p.DamageDone / totalDmg * 100))
			}
			assists = append(assists, fmt.Sprintf("%s (%d%%)", p.Name, pct))
		}
		if len(assists) > 0 {
			desc = "Assisted by: " + strings.Join(assists, " / ")
		}
	}

	ts, _ := time.Parse(time.RFC3339, event.TimeStamp)

	embed := &discordgo.MessageEmbed{
		Color:       color,
		Title:       title,
		URL:         EventURL(gs.Provider, event.EventID, event.Server),
		Description: desc,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://user-images.githubusercontent.com/13356774/76129825-ee15b580-5fde-11ea-9f77-7ae16bd65368.png",
		},
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Fame", Value: fmtNumber(event.TotalVictimKillFame), Inline: true},
			{Name: "Loot Value", Value: fmtNumber(lootSum), Inline: true},
			{Name: "​", Value: "​", Inline: true},
			{Name: "Killer Guild", Value: guildField(event.Killer), Inline: true},
			{Name: "Victim Guild", Value: guildField(event.Victim), Inline: true},
			{Name: "​", Value: "​", Inline: true},
			{Name: "Killer IP", Value: fmt.Sprintf("%d", int(math.Round(event.Killer.AverageItemPower))), Inline: true},
			{Name: "Victim IP", Value: fmt.Sprintf("%d", int(math.Round(event.Victim.AverageItemPower))), Inline: true},
			{Name: "​", Value: "​", Inline: true},
		},
		Footer:    &discordgo.MessageEmbedFooter{Text: "Powered by albion-killgot"},
		Timestamp: ts.Format(time.RFC3339),
	}

	return &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{embed}}
}

// EmbedJuicy builds an embed for a juicy kill (high fame/loot).
func EmbedJuicy(event albion.Event, gs store.GuildSettings, tier string) *discordgo.MessageSend {
	msg := EmbedKill(event, gs, true)
	if len(msg.Embeds) > 0 {
		switch tier {
		case "good":
			msg.Embeds[0].Color = colorGold
			msg.Embeds[0].Title = ":moneybag: " + msg.Embeds[0].Title + " :moneybag:"
		case "insane":
			msg.Embeds[0].Color = colorYellow
			msg.Embeds[0].Title = ":star: " + msg.Embeds[0].Title + " :star:"
		}
	}
	return msg
}

// EmbedBattle builds a text embed for a battle report.
func EmbedBattle(battle albion.Battle, gs store.GuildSettings) *discordgo.MessageSend {
	start, _ := time.Parse(time.RFC3339, battle.StartTime)
	end, _ := time.Parse(time.RFC3339, battle.EndTime)
	dur := end.Sub(start).Round(time.Minute)

	durStr := fmt.Sprintf("%dm", int(dur.Minutes()))
	if dur >= time.Hour {
		durStr = fmt.Sprintf("%dh %dm", int(dur.Hours()), int(dur.Minutes())%60)
	}

	desc := fmt.Sprintf(
		"%d players, %d kills, %s fame, %s duration",
		len(battle.Players),
		battle.TotalKills,
		fmtNumber(battle.TotalFame),
		durStr,
	)

	line := func(name string, kills, deaths int, fame int64, total int) string {
		return fmt.Sprintf("**%s** — %d players, %d kills, %d deaths, %s fame",
			name, total, kills, deaths, fmtNumber(fame))
	}

	var fields []*discordgo.MessageEmbedField
	players := make([]albion.BattlePlayer, 0, len(battle.Players))
	for _, p := range battle.Players {
		players = append(players, p)
	}

	for allianceID, alliance := range battle.Alliances {
		total := 0
		for _, p := range players {
			if p.AllianceID == allianceID {
				total++
			}
		}
		name := line(alliance.Name, alliance.Kills, alliance.Deaths, alliance.KillFame, total)

		var guildLines []string
		for _, guild := range battle.Guilds {
			if guild.AllianceID != allianceID {
				continue
			}
			gTotal := 0
			for _, p := range players {
				if p.GuildID == guild.ID {
					gTotal++
				}
			}
			guildLines = append(guildLines, line(guild.Name, guild.Kills, guild.Deaths, guild.KillFame, gTotal))
		}

		value := strings.Join(guildLines, "\n")
		if len(value) > maxFieldValue {
			value = value[:maxFieldValue]
		}
		if value == "" {
			value = "​"
		}
		fields = append(fields, &discordgo.MessageEmbedField{Name: name, Value: value})
	}

	// Guilds without alliance
	var soloLines []string
	for _, guild := range battle.Guilds {
		if guild.AllianceID != "" {
			continue
		}
		gTotal := 0
		for _, p := range players {
			if p.GuildID == guild.ID {
				gTotal++
			}
		}
		soloLines = append(soloLines, line(guild.Name, guild.Kills, guild.Deaths, guild.KillFame, gTotal))
	}
	if len(soloLines) > 0 {
		value := strings.Join(soloLines, "\n")
		if len(value) > maxFieldValue {
			value = value[:maxFieldValue]
		}
		fields = append(fields, &discordgo.MessageEmbedField{Name: "No Alliance", Value: value})
	}

	if len(fields) > maxFields {
		fields = fields[:maxFields]
	}

	embed := &discordgo.MessageEmbed{
		Color:       colorBattle,
		Title:       fmt.Sprintf("Battle Event (%d guilds)", len(battle.Guilds)),
		URL:         BattleURL(gs.Provider, battle.ID, battle.Server),
		Description: desc,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://user-images.githubusercontent.com/13356774/76130049-b9eec480-5fdf-11ea-95c0-7de130a705a3.png",
		},
		Fields:    fields,
		Footer:    &discordgo.MessageEmbedFooter{Text: "Powered by albion-killgot"},
		Timestamp: end.Format(time.RFC3339),
	}

	return &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{embed}}
}
