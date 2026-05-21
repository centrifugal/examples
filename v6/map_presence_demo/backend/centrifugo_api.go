package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

// CentrifugoAPI is a thin HTTP client for the Centrifugo server API.
type CentrifugoAPI struct {
	BaseURL string
	APIKey  string
	Client  *http.Client
}

func NewCentrifugoAPI(baseURL, apiKey string) *CentrifugoAPI {
	// Default Transport keeps only 2 idle conns per host — with many
	// parallel publishers that exhausts ephemeral ports because each
	// request opens a fresh socket. Tune for sustained throughput.
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          512,
		MaxIdleConnsPerHost:   256,
		MaxConnsPerHost:       256,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &CentrifugoAPI{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}
}

func (c *CentrifugoAPI) call(ctx context.Context, method string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/"+method, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("X-API-Key", c.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		// Drain so the connection is returned to the keep-alive pool.
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("centrifugo %s: %d %s", method, resp.StatusCode, string(b))
	}
	return nil
}

// MapPublish calls /api/map_publish. Data must be valid JSON or empty.
func (c *CentrifugoAPI) MapPublish(ctx context.Context, channel, key string, data []byte) error {
	payload := map[string]any{
		"channel": channel,
		"key":     key,
	}
	if len(data) > 0 {
		payload["data"] = json.RawMessage(data)
	}
	return c.call(ctx, "map_publish", payload)
}

// MapRemove calls /api/map_remove.
func (c *CentrifugoAPI) MapRemove(ctx context.Context, channel, key string) error {
	return c.call(ctx, "map_remove", map[string]any{
		"channel": channel,
		"key":     key,
	})
}

// WaitReady polls /api/info until Centrifugo accepts requests, or ctx is done.
func (c *CentrifugoAPI) WaitReady(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := c.call(ctx, "info", map[string]any{}); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
}
