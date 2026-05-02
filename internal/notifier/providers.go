package notifier

import "fmt"

type Provider struct {
	ID       string
	Name     string
	EventURL func(id int64, server string) string
	BattleURL func(id int64, server string) string
}

var liveID = map[string]string{
	"americas": "live_us",
	"asia":     "live_sgp",
	"europe":   "live_ams",
}

var Providers = []Provider{
	{
		ID:   "albion-killboard",
		Name: "Albion Killboard",
		EventURL: func(id int64, server string) string {
			return fmt.Sprintf("https://albiononline.com/killboard/kill/%d?server=%s", id, liveID[server])
		},
		BattleURL: func(id int64, server string) string {
			return fmt.Sprintf("https://albiononline.com/killboard/battles/%d?server=%s", id, liveID[server])
		},
	},
	{
		ID:   "albion2d",
		Name: "Albion Online 2D",
		EventURL: func(id int64, server string) string {
			sub := ""
			if server == "asia" {
				sub = "sgp."
			} else if server == "europe" {
				sub = "ams."
			}
			return fmt.Sprintf("https://%salbiononline2d.com/en/scoreboard/events/%d", sub, id)
		},
		BattleURL: func(id int64, server string) string {
			sub := ""
			if server == "asia" {
				sub = "sgp."
			} else if server == "europe" {
				sub = "ams."
			}
			return fmt.Sprintf("https://%salbiononline2d.com/en/scoreboard/battles/%d", sub, id)
		},
	},
	{
		ID:   "albion-battles",
		Name: "Albion Battles",
		BattleURL: func(id int64, _ string) string {
			return fmt.Sprintf("https://albionbattles.com/battles/%d", id)
		},
	},
	{
		ID:   "kill-board",
		Name: "Kill-board",
		BattleURL: func(id int64, _ string) string {
			return fmt.Sprintf("https://kill-board.com/battles/%d", id)
		},
	},
}

func ProviderByID(id string) *Provider {
	for i := range Providers {
		if Providers[i].ID == id {
			return &Providers[i]
		}
	}
	return &Providers[0]
}

func EventURL(providerID string, eventID int64, server string) string {
	p := ProviderByID(providerID)
	if p.EventURL == nil {
		return ""
	}
	return p.EventURL(eventID, server)
}

func BattleURL(providerID string, battleID int64, server string) string {
	p := ProviderByID(providerID)
	if p.BattleURL == nil {
		return ""
	}
	return p.BattleURL(battleID, server)
}
