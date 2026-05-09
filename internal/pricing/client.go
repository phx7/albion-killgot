package pricing

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var baseURLs = map[string]string{
	"americas": "https://west.albion-online-data.com/api/v2",
	"asia":     "https://east.albion-online-data.com/api/v2",
	"europe":   "https://europe.albion-online-data.com/api/v2",
}

type Client struct {
	http    *http.Client
	baseURL string
}

func NewClient(server string) *Client {
	base, ok := baseURLs[server]
	if !ok {
		base = baseURLs["americas"]
	}
	return &Client{
		http:    &http.Client{Timeout: 20 * time.Second},
		baseURL: base,
	}
}

type ItemSpec struct {
	ID      string
	Quality int // 1=Normal .. 5=Masterpiece; 0 treated as 1
	Count   int
}

type priceEntry struct {
	ItemID       string `json:"item_id"`
	Quality      int    `json:"quality"`
	SellPriceMin int64  `json:"sell_price_min"`
}

// GetPrices fetches best sell_price_min per (item_id, quality) across all cities.
func (c *Client) GetPrices(ctx context.Context, items []ItemSpec) (map[string]int64, error) {
	idSet := make(map[string]bool)
	for _, item := range items {
		if item.ID != "" {
			idSet[item.ID] = true
		}
	}
	if len(idSet) == 0 {
		return nil, nil
	}
	ids := make([]string, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}

	url := fmt.Sprintf("%s/stats/prices/%s", c.baseURL, strings.Join(ids, ","))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "albion-killgot/1.0")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pricing API: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var entries []priceEntry
	if err := json.Unmarshal(body, &entries); err != nil {
		return nil, err
	}

	result := make(map[string]int64)
	for _, e := range entries {
		if e.SellPriceMin <= 0 {
			continue
		}
		key := priceKey(e.ItemID, e.Quality)
		if e.SellPriceMin > result[key] {
			result[key] = e.SellPriceMin
		}
	}
	return result, nil
}

func priceKey(itemID string, quality int) string {
	if quality <= 0 {
		quality = 1
	}
	return fmt.Sprintf("%s:%d", itemID, quality)
}

// Total calculates total silver value for items using prices from GetPrices.
func Total(items []ItemSpec, prices map[string]int64) int64 {
	var total int64
	for _, item := range items {
		if item.ID == "" {
			continue
		}
		q := item.Quality
		if q <= 0 {
			q = 1
		}
		count := item.Count
		if count <= 0 {
			count = 1
		}
		price := prices[priceKey(item.ID, q)]
		if price == 0 {
			price = prices[priceKey(item.ID, 1)]
		}
		total += price * int64(count)
	}
	return total
}
