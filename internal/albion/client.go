package albion

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	defaultLimit = 51
	httpTimeout  = 60 * time.Second
	maxRetries   = 3
	retryDelay   = 5 * time.Second
)

type Client struct {
	http    *http.Client
	baseURL string
}

func NewClient(server Server) *Client {
	base, ok := serverBaseURLs[server.ID]
	if !ok {
		base = serverBaseURLs["americas"]
	}
	// Force IPv4 and HTTP/1.1 only (no HTTP/2 in ALPN) to avoid TLS fingerprint issues with Cloudflare
	dialer := &net.Dialer{Timeout: 30 * time.Second}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, "tcp4", addr)
		},
		TLSClientConfig: &tls.Config{
			NextProtos: []string{"http/1.1"},
		},
	}
	return &Client{
		http:    &http.Client{Timeout: httpTimeout, Transport: transport},
		baseURL: base,
	}
}

func (c *Client) get(ctx context.Context, path string, params url.Values, out any) error {
	u := c.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(retryDelay):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Accept-Language", "en-US,en;q=0.9")

		resp, err := c.http.Do(req)
		if err != nil {
			// Retry on EOF and temporary network errors
			if errors.Is(err, io.EOF) || isTemporary(err) {
				lastErr = err
				continue
			}
			return err
		}

		if resp.StatusCode == http.StatusGatewayTimeout {
			resp.Body.Close()
			lastErr = fmt.Errorf("status 504")
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return fmt.Errorf("albion API %s: status %d", path, resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		return json.Unmarshal(body, out)
	}
	return fmt.Errorf("after %d retries: %w", maxRetries, lastErr)
}

func isTemporary(err error) bool {
	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}

func (c *Client) GetEvents(ctx context.Context, offset int) ([]Event, error) {
	params := url.Values{
		"offset":    []string{strconv.Itoa(offset)},
		"limit":     []string{strconv.Itoa(defaultLimit)},
		"timestamp": []string{strconv.FormatInt(time.Now().Unix(), 10)},
	}
	var events []Event
	if err := c.get(ctx, "events", params, &events); err != nil {
		return nil, err
	}
	return events, nil
}

func (c *Client) GetBattles(ctx context.Context, offset int) ([]Battle, error) {
	params := url.Values{
		"offset":    []string{strconv.Itoa(offset)},
		"limit":     []string{strconv.Itoa(defaultLimit)},
		"sort":      []string{"recent"},
		"timestamp": []string{strconv.FormatInt(time.Now().Unix(), 10)},
	}
	var battles []Battle
	if err := c.get(ctx, "battles", params, &battles); err != nil {
		return nil, err
	}
	return battles, nil
}

func (c *Client) Search(ctx context.Context, q string) (*SearchResults, error) {
	params := url.Values{"q": []string{q}}
	var results SearchResults
	if err := c.get(ctx, "search", params, &results); err != nil {
		return nil, err
	}
	return &results, nil
}

func (c *Client) GetAlliance(ctx context.Context, id string) (*SearchEntity, error) {
	var entity SearchEntity
	if err := c.get(ctx, "alliances/"+id, nil, &entity); err != nil {
		return nil, err
	}
	return &entity, nil
}
