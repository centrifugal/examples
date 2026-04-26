package main

import (
	"bytes"
	"context"
	"encoding/base64"
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
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          64,
		MaxIdleConnsPerHost:   32,
		MaxConnsPerHost:       32,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &CentrifugoAPI{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Client:  &http.Client{Timeout: 5 * time.Second, Transport: transport},
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
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("centrifugo %s: %d %s", method, resp.StatusCode, string(b))
	}
	return nil
}

// PublishBinary publishes a raw byte payload via /api/publish using b64data.
// The bytes ride to Protobuf-WS clients as binary; JSON-WS clients receive
// the same bytes as a base64 string in publication.data.
func (c *CentrifugoAPI) PublishBinary(ctx context.Context, channel string, data []byte) error {
	return c.call(ctx, "publish", map[string]any{
		"channel": channel,
		"b64data": base64.StdEncoding.EncodeToString(data),
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
