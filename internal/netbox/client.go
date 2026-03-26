package netbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SyncPayload is an illustrative webhook body after apply (see docs/integrations.md).
type SyncPayload struct {
	Action string `json:"action"`
	IP     string `json:"ip"`
	Role   string `json:"role,omitempty"`
}

// Client posts JSON payloads to a configured HTTP endpoint.
type Client struct {
	HTTPClient *http.Client
}

func (c *Client) httpClient() *http.Client {
	if c != nil && c.HTTPClient != nil {
		return c.HTTPClient
	}
	return &http.Client{Timeout: 30 * time.Second}
}

// PostJSON sends an arbitrary JSON body to url. Optional headers are merged (e.g. idempotency keys).
func (c *Client) PostJSON(ctx context.Context, url string, body []byte, headers map[string]string) error {
	if url == "" {
		return fmt.Errorf("netbox: empty url")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		if k != "" && v != "" {
			req.Header.Set(k, v)
		}
	}
	resp, err := c.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("netbox: webhook %s: %s", resp.Status, string(body))
	}
	return nil
}

// PostWebhook sends legacy SyncPayload as JSON (no apiVersion field).
func (c *Client) PostWebhook(ctx context.Context, url string, p SyncPayload) error {
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return c.PostJSON(ctx, url, b, nil)
}
