// Package picsum is the library behind the picsum command line:
// the HTTP client, request shaping, and the typed data models for the
// Lorem Picsum public placeholder-image API.
//
// The API requires no authentication. A polite User-Agent and 200 ms pacing
// between requests keeps the client well within the public rate limits.
package picsum

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Host is the site this client talks to.
const Host = "picsum.photos"

// Config holds all tunable parameters for the Client.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://picsum.photos",
		UserAgent: "picsum-cli/0.1 (tamnd87@gmail.com)",
		Rate:      200 * time.Millisecond,
		Timeout:   10 * time.Second,
		Retries:   3,
	}
}

// Image is one Lorem Picsum image record.
type Image struct {
	ID          string `kit:"id" json:"id"`
	Author      string `json:"author"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	URL         string `json:"url"`
	DownloadURL string `json:"download_url"`
}

// Client talks to picsum.photos over HTTP.
type Client struct {
	cfg  Config
	http *http.Client
	mu   sync.Mutex
	last time.Time
}

// NewClient returns a Client configured with cfg.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// List returns a paginated slice of images.
func (c *Client) List(ctx context.Context, page, limit int) ([]Image, error) {
	url := fmt.Sprintf("%s/v2/list?page=%d&limit=%d", c.cfg.BaseURL, page, limit)
	body, err := c.get(ctx, url)
	if err != nil {
		return nil, err
	}
	var images []Image
	if err := json.Unmarshal(body, &images); err != nil {
		return nil, fmt.Errorf("decode image list: %w", err)
	}
	return images, nil
}

// Info returns the metadata for the image with the given id.
func (c *Client) Info(ctx context.Context, id string) (Image, error) {
	url := fmt.Sprintf("%s/id/%s/info", c.cfg.BaseURL, id)
	body, err := c.get(ctx, url)
	if err != nil {
		return Image{}, err
	}
	var img Image
	if err := json.Unmarshal(body, &img); err != nil {
		return Image{}, fmt.Errorf("decode image info: %w", err)
	}
	return img, nil
}

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, url)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	return b, err != nil, err
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}
