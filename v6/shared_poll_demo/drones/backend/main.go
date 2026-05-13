package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand/v2"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	hmacSecret = "demo-drones-secret"
	apiKey     = "demo-drones-api-key"
	cellSize   = 0.005 // ~550 m
	numDrones  = 500
	channel    = "drones:sf"

	// San Francisco — full peninsula plus bay edges.
	minLat = 37.700
	maxLat = 37.820
	minLng = -122.520
	maxLng = -122.360

	droneMinSpeed = 0.00015
	droneMaxSpeed = 0.00060
)

// DronePos is the JSON-visible part of a drone.
type DronePos struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
}

var (
	adjectives = []string{
		"Swift", "Silent", "Crimson", "Golden", "Shadow",
		"Iron", "Copper", "Storm", "Frost", "Ember",
		"Cobalt", "Jade", "Amber", "Onyx", "Silver",
		"Rusty", "Bright", "Dark", "Wild", "Calm",
		"Brave", "Lucky", "Dusty", "Misty", "Solar",
	}
	animals = []string{
		"Falcon", "Hawk", "Owl", "Raven", "Eagle",
		"Sparrow", "Heron", "Crane", "Condor", "Kite",
		"Osprey", "Wren", "Finch", "Tern", "Swift",
		"Robin", "Starling", "Jay", "Dove", "Lark",
	}
)

func droneName(i int) string {
	return adjectives[i%len(adjectives)] + " " + animals[i/len(adjectives)%len(animals)]
}

// drone adds bearing + speed for smooth curved flight paths.
type drone struct {
	DronePos
	bearing float64 // radians: 0=north, π/2=east
	speed   float64 // degrees per tick
}

var (
	mu           sync.RWMutex
	drones       []*drone
	cells        map[string][]DronePos // cellKey → sorted by ID
	cellVersions map[string]uint64
	version      uint64

	// channelEpoch is fresh per process startup. On restart Centrifugo
	// detects the change and unsubscribes connected clients with
	// insufficient-state code so they re-track from version 0 on
	// resubscribe — picture recovers cleanly without page reload.
	channelEpoch = uuid.NewString()

	centrifugoURL string
	httpClient    = &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
	}
)

func cellKey(lat, lng float64) string {
	cLat := math.Floor(lat/cellSize) * cellSize
	cLng := math.Floor(lng/cellSize) * cellSize
	return fmt.Sprintf("%.3f:%.3f", cLat, cLng)
}

// ---------- SharedPollPublish ----------

type pubItem struct {
	key     string
	data    []byte
	version uint64
}

// Buffered channel of capacity 1 — if the publisher is busy, the old
// batch is replaced by the newer one (fresher data wins).
var pubCh = make(chan []pubItem, 1)

func publishLoop() {
	for pubs := range pubCh {
		for _, p := range pubs {
			b64 := base64.StdEncoding.EncodeToString(p.data)
			body, _ := json.Marshal(map[string]any{
				"channel": channel,
				"key":     p.key,
				"b64data": b64,
				"version": p.version,
				"epoch":   channelEpoch,
			})
			req, _ := http.NewRequest("POST", centrifugoURL+"/api/shared_poll_publish", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "apikey "+apiKey)
			resp, err := httpClient.Do(req)
			if err != nil {
				log.Printf("shared_poll_publish %s: %v", p.key, err)
				continue
			}
			io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
	}
}

// ---------- Simulation ----------

func initDrones() {
	drones = make([]*drone, numDrones)
	for i := range numDrones {
		d := &drone{}
		d.ID = fmt.Sprintf("drone_%04d", i+1)
		d.Name = droneName(i)
		d.Lat = minLat + rand.Float64()*(maxLat-minLat)
		d.Lng = minLng + rand.Float64()*(maxLng-minLng)
		d.bearing = rand.Float64() * 2 * math.Pi
		d.speed = droneMinSpeed + rand.Float64()*(droneMaxSpeed-droneMinSpeed)
		drones[i] = d
	}
	cellVersions = make(map[string]uint64)
	version = 1
	cells = buildCells()
	for k := range cells {
		cellVersions[k] = version
	}
}

func buildCells() map[string][]DronePos {
	m := make(map[string][]DronePos)
	for _, d := range drones {
		k := cellKey(d.Lat, d.Lng)
		m[k] = append(m[k], d.DronePos)
	}
	for k, ds := range m {
		sort.Slice(ds, func(i, j int) bool { return ds[i].ID < ds[j].ID })
		m[k] = ds
	}
	return m
}

const numGroups = 50 // stagger drones into 50 groups, tick every 20ms

func moveDrone(d *drone) {
	// Gentle turn — smooth curved flight paths.
	d.bearing += (rand.Float64() - 0.5) * 0.4
	// Slight speed variation.
	d.speed += (rand.Float64() - 0.5) * 0.00004
	d.speed = max(droneMinSpeed, min(droneMaxSpeed, d.speed))
	// Move along bearing.
	d.Lat += math.Cos(d.bearing) * d.speed
	d.Lng += math.Sin(d.bearing) * d.speed
	// Bounce off city bounds.
	if d.Lat <= minLat || d.Lat >= maxLat {
		d.Lat = max(minLat, min(maxLat, d.Lat))
		d.bearing = math.Pi - d.bearing
	}
	if d.Lng <= minLng || d.Lng >= maxLng {
		d.Lng = max(minLng, min(maxLng, d.Lng))
		d.bearing = -d.bearing
	}
}

func simulationLoop() {
	ticker := time.NewTicker(time.Second / numGroups)
	groupIdx := 0
	for range ticker.C {
		mu.Lock()
		version++

		// Move only drones in this group.
		groupSize := numDrones / numGroups
		start := groupIdx * groupSize
		end := start + groupSize
		if groupIdx == numGroups-1 {
			end = numDrones
		}
		for i := start; i < end; i++ {
			moveDrone(drones[i])
		}
		groupIdx = (groupIdx + 1) % numGroups

		newCells := buildCells()
		changed := changedKeys(cells, newCells)
		cells = newCells
		for _, k := range changed {
			cellVersions[k] = version
		}
		// Clean up versions for cells with no drones.
		for k := range cellVersions {
			if _, ok := cells[k]; !ok {
				delete(cellVersions, k)
			}
		}

		// Collect publish payloads while holding the lock.
		pubs := make([]pubItem, 0, len(changed))
		for _, k := range changed {
			ds := cells[k]
			if ds == nil {
				ds = []DronePos{}
			}
			data, _ := json.Marshal(map[string]any{"drones": ds})
			pubs = append(pubs, pubItem{key: k, data: data, version: cellVersions[k]})
		}
		mu.Unlock()

		// Send to publisher; drop if it's still busy (next tick has fresher data).
		if len(pubs) > 0 {
			select {
			case pubCh <- pubs:
			default:
			}
		}
	}
}

func changedKeys(old, cur map[string][]DronePos) []string {
	all := make(map[string]struct{})
	for k := range old {
		all[k] = struct{}{}
	}
	for k := range cur {
		all[k] = struct{}{}
	}
	var out []string
	for k := range all {
		if !posSliceEqual(old[k], cur[k]) {
			out = append(out, k)
		}
	}
	return out
}

func posSliceEqual(a, b []DronePos) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ---------- HMAC ----------

func makeTrackSignature(secret, ch string, keys []string, user string, ttl int) string {
	now := time.Now().Unix()
	expiry := now + int64(ttl)
	keysHash := sha256.Sum256([]byte(strings.Join(keys, "\x00")))
	// Inner payload fields are NUL-separated to prevent colon-injection
	// ambiguity in (user_id, channel). Outer signature stays ':'-separated
	// because iat/exp/hmac_hex are colon-free by construction.
	payload := fmt.Sprintf("%d\x00%d\x00%s\x00%s\x00%x", now, expiry, user, ch, keysHash)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return fmt.Sprintf("%d:%d:%x", now, expiry, mac.Sum(nil))
}

// ---------- HTTP ----------

// GET /api/cells?keys=k1,k2,...
func handleGetCells(w http.ResponseWriter, r *http.Request) {
	keysParam := r.URL.Query().Get("keys")
	if keysParam == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"cells": map[string]any{}, "signature": "", "keys": []string{}})
		return
	}
	keys := strings.Split(keysParam, ",")

	mu.RLock()
	result := make(map[string][]DronePos, len(keys))
	versions := make(map[string]uint64, len(keys))
	for _, k := range keys {
		if ds, ok := cells[k]; ok {
			result[k] = ds
		} else {
			result[k] = []DronePos{}
		}
		versions[k] = cellVersions[k]
	}
	mu.RUnlock()

	sig := makeTrackSignature(hmacSecret, channel, keys, "", 30)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"cells":     result,
		"versions":  versions,
		"signature": sig,
		"keys":      keys,
	})
}

// POST /api/track_refresh
func handleTrackRefresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Keys    []string `json:"keys"`
		Channel string   `json:"channel"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	sig := makeTrackSignature(hmacSecret, req.Channel, req.Keys, "", 30)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"keys": req.Keys, "signature": sig})
}

// POST /centrifugo/refresh — shared poll refresh proxy (safety net).
func handleCentrifugoRefresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Channel string `json:"channel"`
		Items   []struct {
			Key string `json:"key"`
		} `json:"items"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if len(req.Items) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"items": []any{}}})
		return
	}

	mu.RLock()
	type respItem struct {
		Key     string `json:"key"`
		Data    any    `json:"data"`
		Version uint64 `json:"version"`
	}
	currentVersion := version
	items := make([]respItem, 0, len(req.Items))
	for _, it := range req.Items {
		ds := cells[it.Key]
		if ds == nil {
			ds = []DronePos{}
		}
		v := cellVersions[it.Key]
		if v == 0 {
			// Cell has no drones currently. Report the global current
			// version so Centrifugo's per-key version comparison advances
			// past whatever stale entry.version it may hold — otherwise
			// the empty-list response gets dropped as "unchanged" and
			// previously-rendered drones linger on subscribers' screens.
			// This matters specifically after backend restart: the new
			// publisher process has no memory of pre-restart cells, so
			// it never publishes "this cell is empty now" via the direct
			// publish path; the refresh proxy is the only path that can
			// communicate the absence, and only if its version advances.
			v = currentVersion
		}
		items = append(items, respItem{
			Key:     it.Key,
			Data:    map[string]any{"drones": ds},
			Version: v,
		})
	}
	mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{
		"items": items,
		"epoch": channelEpoch,
	}})
}

func main() {
	centrifugoURL = os.Getenv("CENTRIFUGO_URL")
	if centrifugoURL == "" {
		centrifugoURL = "http://localhost:8000"
	}

	initDrones()
	go publishLoop()
	go simulationLoop()

	fs := http.FileServer(http.Dir("../static"))
	http.Handle("/", fs)
	http.HandleFunc("/api/cells", handleGetCells)
	http.HandleFunc("/api/track_refresh", handleTrackRefresh)
	http.HandleFunc("/centrifugo/refresh", handleCentrifugoRefresh)

	log.Println("Backend listening on :3001")
	log.Fatal(http.ListenAndServe(":3001", nil))
}
