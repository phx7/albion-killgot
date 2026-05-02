package albion

type Server struct {
	ID   string
	Name string
}

var (
	Americas = Server{ID: "americas", Name: "Albion Americas"}
	Asia     = Server{ID: "asia", Name: "Albion Asia"}
	Europe   = Server{ID: "europe", Name: "Albion Europe"}
	Servers  = []Server{Americas, Asia, Europe}
)

var serverBaseURLs = map[string]string{
	"americas": "https://gameinfo.albiononline.com/api/gameinfo/",
	"asia":     "https://gameinfo-sgp.albiononline.com/api/gameinfo/",
	"europe":   "https://gameinfo-ams.albiononline.com/api/gameinfo/",
}

// --- Kill events ---

type Event struct {
	EventID              int64       `json:"EventId"`
	TimeStamp            string      `json:"TimeStamp"`
	TotalVictimKillFame  int64       `json:"TotalVictimKillFame"`
	BattleID             int64       `json:"BattleId"`
	Killer               EventPlayer `json:"Killer"`
	Victim               EventPlayer `json:"Victim"`
	Participants         []EventPlayer `json:"Participants"`
	GroupMembers         []EventPlayer `json:"GroupMembers"`

	// Enriched fields (not from API)
	Server    string `json:"server,omitempty"`
	LootValue *LootValue `json:"lootValue,omitempty"`
}

type EventPlayer struct {
	ID              string    `json:"Id"`
	Name            string    `json:"Name"`
	GuildID         string    `json:"GuildId"`
	GuildName       string    `json:"GuildName"`
	AllianceID      string    `json:"AllianceId"`
	AllianceName    string    `json:"AllianceName"`
	AllianceTag     string    `json:"AllianceTag"`
	AverageItemPower float64  `json:"AverageItemPower"`
	KillFame        int64     `json:"KillFame"`
	DeathFame       int64     `json:"DeathFame"`
	DamageDone      float64   `json:"DamageDone"`
	Equipment       Equipment `json:"Equipment"`
	Inventory       []Item    `json:"Inventory"`
}

type Equipment struct {
	MainHand *Item `json:"MainHand"`
	OffHand  *Item `json:"OffHand"`
	Head     *Item `json:"Head"`
	Armor    *Item `json:"Armor"`
	Shoes    *Item `json:"Shoes"`
	Bag      *Item `json:"Bag"`
	Cape     *Item `json:"Cape"`
	Mount    *Item `json:"Mount"`
	Potion   *Item `json:"Potion"`
	Food     *Item `json:"Food"`
}

type Item struct {
	Type    string `json:"Type"`
	Count   int    `json:"Count"`
	Quality int    `json:"Quality"`
}

type LootValue struct {
	Equipment int64 `json:"equipment"`
	Inventory int64 `json:"inventory"`
}

// --- Battles ---

type Battle struct {
	ID         int64                     `json:"id"`
	StartTime  string                    `json:"startTime"`
	EndTime    string                    `json:"endTime"`
	TotalKills int                       `json:"totalKills"`
	TotalFame  int64                     `json:"totalFame"`
	Players    map[string]BattlePlayer   `json:"players"`
	Guilds     map[string]BattleGuild    `json:"guilds"`
	Alliances  map[string]BattleAlliance `json:"alliances"`

	// Enriched field
	Server string `json:"server,omitempty"`
}

type BattlePlayer struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	GuildID     string `json:"guildId"`
	AllianceID  string `json:"allianceId"`
	Kills       int    `json:"kills"`
	Deaths      int    `json:"deaths"`
	KillFame    int64  `json:"killFame"`
	AverageItemPower float64 `json:"averageItemPower"`
}

type BattleGuild struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	AllianceID string `json:"allianceId"`
	Kills      int    `json:"kills"`
	Deaths     int    `json:"deaths"`
	KillFame   int64  `json:"killFame"`
}

type BattleAlliance struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Kills    int    `json:"kills"`
	Deaths   int    `json:"deaths"`
	KillFame int64  `json:"killFame"`
}

// --- Search ---

type SearchResults struct {
	Players   []SearchEntity `json:"players"`
	Guilds    []SearchEntity `json:"guilds"`
	Alliances []SearchEntity `json:"alliances"`
}

type SearchEntity struct {
	ID   string `json:"Id"`
	Name string `json:"Name"`
	// For alliances the tag is in a different field
	AllianceID  string `json:"AllianceId"`
	AllianceTag string `json:"AllianceTag"`
}
