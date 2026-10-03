package bookingcom

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Config holds connection parameters for Booking.com Demand & Connectivity APIs
type Config struct {
	BaseURL     string
	APIKey      string
	AffiliateID string
	Sandbox     bool
	Timeout     time.Duration
}

// Client handles HTTP transport with Booking.com
type Client struct {
	cfg        Config
	httpClient *http.Client
}

// NewClient initializes a new Booking.com client
func NewClient(cfg Config) *Client {
	if cfg.BaseURL == "" {
		if cfg.Sandbox {
			cfg.BaseURL = "https://demandapi-sandbox.booking.com/3.1"
		} else {
			cfg.BaseURL = "https://demandapi.booking.com/3.1"
		}
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}

	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 25,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

// DoRequest performs authenticated HTTP requests
func (c *Client) DoRequest(ctx context.Context, method, path string, body interface{}, out interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	url := fmt.Sprintf("%s%s", c.cfg.BaseURL, path)
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.cfg.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.cfg.APIKey))
	}
	if c.cfg.AffiliateID != "" {
		req.Header.Set("X-Affiliate-Id", c.cfg.AffiliateID)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("execute request to Booking.com (%s): %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("booking.com api error (status %d): %s", resp.StatusCode, string(respBytes))
	}

	if out != nil {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}

	return nil
}
