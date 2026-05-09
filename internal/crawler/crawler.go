package crawler

import (
	"context"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/phx7/albion-killgot/internal/albion"
)

const (
	eventInterval  = 30 * time.Second
	battleInterval = 2 * time.Minute
	syncInterval   = 5 * time.Minute
)

// ServerSource returns the set of albion server IDs that should be crawled.
type ServerSource func(ctx context.Context) ([]string, error)

type Crawler struct {
	source  ServerSource
	Events  chan albion.Event
	Battles chan albion.Battle

	mu      sync.Mutex
	running map[string]context.CancelFunc // serverID -> cancel for its goroutines
}

func New(source ServerSource) *Crawler {
	return &Crawler{
		source:  source,
		Events:  make(chan albion.Event, 1000),
		Battles: make(chan albion.Battle, 200),
		running: make(map[string]context.CancelFunc),
	}
}

func (c *Crawler) Run(ctx context.Context) {
	c.sync(ctx)
	go func() {
		ticker := time.NewTicker(syncInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.sync(ctx)
			}
		}
	}()
}

// Sync triggers an immediate server sync. Call it after tracking changes.
func (c *Crawler) Sync(ctx context.Context) {
	c.sync(ctx)
}

// sync starts pollers for newly active servers and stops pollers for inactive ones.
func (c *Crawler) sync(ctx context.Context) {
	ids, err := c.source(ctx)
	if err != nil {
		slog.Error("crawler: fetch active servers", "err", err)
		return
	}

	wanted := make(map[string]bool, len(ids))
	for _, id := range ids {
		wanted[id] = true
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Stop pollers for servers no longer needed
	for id, cancel := range c.running {
		if !wanted[id] {
			slog.Info("crawler: stopping server", "server", id)
			cancel()
			delete(c.running, id)
		}
	}

	// Start pollers for new servers
	for _, id := range ids {
		if _, ok := c.running[id]; ok {
			continue
		}
		var server albion.Server
		for _, s := range albion.Servers {
			if s.ID == id {
				server = s
				break
			}
		}
		if server.ID == "" {
			slog.Warn("crawler: unknown server id", "id", id)
			continue
		}
		slog.Info("crawler: starting server", "server", id)
		sCtx, cancel := context.WithCancel(ctx)
		c.running[id] = cancel
		client := albion.NewClient(server)
		go c.runEvents(sCtx, server, client)
		go c.runBattles(sCtx, server, client)
	}
}

func (c *Crawler) runEvents(ctx context.Context, server albion.Server, client *albion.Client) {
	var latestID int64
	ticker := time.NewTicker(eventInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			events, newLatestID, err := c.fetchEventsTo(ctx, server, client, latestID)
			if err != nil {
				slog.Error("fetch events", "server", server.ID, "err", err)
				continue
			}
			latestID = newLatestID
			for _, e := range events {
				if e.EventID > latestID {
					latestID = e.EventID
				}
				select {
				case c.Events <- e:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func (c *Crawler) runBattles(ctx context.Context, server albion.Server, client *albion.Client) {
	var latestID int64
	ticker := time.NewTicker(battleInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			battles, newLatestID, err := c.fetchBattlesTo(ctx, server, client, latestID)
			if err != nil {
				slog.Error("fetch battles", "server", server.ID, "err", err)
				continue
			}
			latestID = newLatestID
			for _, b := range battles {
				if b.ID > latestID {
					latestID = b.ID
				}
				select {
				case c.Battles <- b:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func (c *Crawler) fetchEventsTo(ctx context.Context, server albion.Server, client *albion.Client, latestID int64) ([]albion.Event, int64, error) {
	var result []albion.Event

	for offset := 0; offset < 1000; {
		batch, err := client.GetEvents(ctx, offset)
		if err != nil {
			return nil, latestID, err
		}
		if len(batch) == 0 {
			break
		}

		foundLatest := false
		for _, e := range batch {
			if latestID == 0 {
				// First run: just record latest, don't publish history
				latestID = e.EventID
				foundLatest = true
				break
			}
			if e.EventID <= latestID {
				foundLatest = true
				break
			}
			e.Server = server.ID
			result = append(result, e)
		}
		if foundLatest {
			break
		}
		offset += len(batch)
	}

	sort.Slice(result, func(i, j int) bool { return result[i].EventID < result[j].EventID })
	return result, latestID, nil
}

func (c *Crawler) fetchBattlesTo(ctx context.Context, server albion.Server, client *albion.Client, latestID int64) ([]albion.Battle, int64, error) {
	var result []albion.Battle

	for offset := 0; offset < 1000; {
		batch, err := client.GetBattles(ctx, offset)
		if err != nil {
			return nil, latestID, err
		}
		if len(batch) == 0 {
			break
		}

		foundLatest := false
		for _, b := range batch {
			if latestID == 0 {
				latestID = b.ID
				foundLatest = true
				break
			}
			if b.ID <= latestID {
				foundLatest = true
				break
			}
			b.Server = server.ID
			result = append(result, b)
		}
		if foundLatest {
			break
		}
		offset += len(batch)
	}

	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, latestID, nil
}
