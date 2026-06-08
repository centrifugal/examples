// Command backend is a tiny, dependency-free demo backend for Centrifugo PRO native Web Push.
//
// It exposes three endpoints (served behind nginx under /api/):
//   - GET  /api/vapid    — returns the VAPID public key for the frontend,
//   - POST /api/register — proxies to Centrifugo device_register,
//   - POST /api/send     — proxies to Centrifugo send_push_notification,
//
// adding the "Authorization: apikey <key>" header so the API key stays on the backend.
//
// It is intentionally self-contained (Go standard library only). Run `go run . -genvapid`
// locally to generate a VAPID key pair for centrifugo.json.
package main

import (
	"bytes"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	genVAPID := flag.Bool("genvapid", false, "generate a VAPID key pair and exit")
	flag.Parse()
	if *genVAPID {
		generateVAPIDKeys()
		return
	}

	configPath := getenv("CONFIG_PATH", "centrifugo.json")
	centrifugoURL := getenv("CENTRIFUGO_URL", "http://localhost:8000")
	addr := getenv("ADDR", ":5000")

	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("error loading %s: %v", configPath, err)
	}
	if err := validateVAPIDPublicKey(cfg.vapidPublicKey); err != nil {
		log.Fatalf("invalid push_notifications.webpush.vapid_public_key in %s: %v\nRun `go run . -genvapid` and paste the generated keys into %s.", configPath, err, configPath)
	}

	proxy := &centrifugoProxy{baseURL: centrifugoURL, apiKey: cfg.apiKey, client: &http.Client{Timeout: 10 * time.Second}}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/vapid", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"publicKey": cfg.vapidPublicKey})
	})
	mux.HandleFunc("/api/register", proxy.register)
	mux.HandleFunc("/api/send", proxy.send)

	log.Printf("backend listening on %s, proxying to Centrifugo at %s", addr, centrifugoURL)
	log.Fatal(http.ListenAndServe(addr, mux))
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// --- centrifugo.json reading (only the bits we need) ---

type demoConfig struct {
	vapidPublicKey string
	apiKey         string
}

func loadConfig(path string) (demoConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return demoConfig{}, err
	}
	var raw struct {
		HTTPAPI struct {
			Key string `json:"key"`
		} `json:"http_api"`
		PushNotifications struct {
			WebPush struct {
				VAPIDPublicKey string `json:"vapid_public_key"`
			} `json:"webpush"`
		} `json:"push_notifications"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return demoConfig{}, err
	}
	return demoConfig{
		vapidPublicKey: raw.PushNotifications.WebPush.VAPIDPublicKey,
		apiKey:         raw.HTTPAPI.Key,
	}, nil
}

// validateVAPIDPublicKey ensures the configured key is a real base64url-encoded uncompressed
// P-256 point (65 bytes, leading 0x04), not the placeholder. This catches the common mistake
// of forgetting to replace the placeholder, which would otherwise reach the browser and make
// pushManager.subscribe throw "InvalidCharacterError" on atob.
func validateVAPIDPublicKey(key string) error {
	if key == "" {
		return fmt.Errorf("key is empty")
	}
	if strings.Contains(key, "REPLACE") {
		return fmt.Errorf("key is still the placeholder value")
	}
	raw, err := base64.RawURLEncoding.DecodeString(key)
	if err != nil {
		return fmt.Errorf("key is not valid base64url: %w", err)
	}
	if len(raw) != 65 || raw[0] != 0x04 {
		return fmt.Errorf("key must be a 65-byte uncompressed P-256 point (got %d bytes)", len(raw))
	}
	return nil
}

// --- Centrifugo HTTP API proxy ---

type centrifugoProxy struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// register accepts {"subscription": <PushSubscription>, "user": "..."} from the browser and
// calls Centrifugo device_register, storing the subscription JSON as the device token.
func (p *centrifugoProxy) register(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Subscription json.RawMessage `json:"subscription"`
		User         string          `json:"user"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	apiReq := map[string]any{
		"provider": "webpush",
		"platform": "web",
		"token":    string(body.Subscription), // the whole PushSubscription JSON is the token
		"user":     body.User,
	}
	p.forward(w, "device_register", apiReq)
}

// send accepts {"user": "...", "title": "...", "body": "..."} and calls send_push_notification
// targeting all webpush devices (optionally narrowed to a user) via a DeviceFilter recipient.
func (p *centrifugoProxy) send(w http.ResponseWriter, r *http.Request) {
	var body struct {
		User  string `json:"user"`
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if body.Title == "" {
		body.Title = "Hello from Centrifugo"
	}

	filter := map[string]any{"providers": []string{"webpush"}}
	if body.User != "" {
		filter["users"] = []string{body.User}
	}
	// This payload is exactly what the service worker receives via event.data.json().
	// Use json.RawMessage so it is embedded as a JSON object, not base64-encoded (which is
	// what a plain []byte would become when marshaled).
	payload, _ := json.Marshal(map[string]string{"title": body.Title, "body": body.Body})

	apiReq := map[string]any{
		"recipient": map[string]any{"filter": filter},
		"notification": map[string]any{
			"expire_at": time.Now().Add(time.Hour).Unix(),
			"webpush": map[string]any{
				"payload": json.RawMessage(payload),
				"headers": map[string]string{"TTL": "300", "Urgency": "high"},
			},
		},
	}
	p.forward(w, "send_push_notification", apiReq)
}

func (p *centrifugoProxy) forward(w http.ResponseWriter, method string, apiReq map[string]any) {
	reqBody, _ := json.Marshal(apiReq)
	httpReq, err := http.NewRequest(http.MethodPost, p.baseURL+"/api/"+method, bytes.NewReader(reqBody))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "apikey "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, _ := io.ReadAll(resp.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(respBody)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// generateVAPIDKeys prints a fresh VAPID key pair in the base64url (raw, unpadded) format that
// Centrifugo and browsers expect. Uses the standard library only.
func generateVAPIDKeys() {
	priv, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		log.Fatalf("error generating key: %v", err)
	}
	public := base64.RawURLEncoding.EncodeToString(priv.PublicKey().Bytes())
	private := base64.RawURLEncoding.EncodeToString(priv.Bytes())
	fmt.Println("VAPID key pair generated. Put these into centrifugo.json under push_notifications.webpush:")
	fmt.Printf("\n  \"vapid_public_key\": %q,\n  \"vapid_private_key\": %q\n\n", public, private)
}
