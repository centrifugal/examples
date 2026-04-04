package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const (
	pollSecret = "demo-poll-secret"
	numPosts   = 50
)

var (
	pool *pgxpool.Pool
	rdb  *redis.Client
)

var titles = []string{
	"Show HN: I built a real-time voting system with shared polling",
	"Why WebSockets still matter in 2026",
	"Ask HN: What's your favorite approach to real-time updates?",
	"The unexpected complexity of distributed counters",
	"Centrifugo: scalable real-time messaging server (open source)",
	"How we reduced our API polling traffic by 95%",
	"Server-Sent Events vs WebSockets vs Long Polling",
	"Ask HN: How do you handle optimistic UI updates?",
	"The architecture of real-time leaderboards at scale",
	"Why I switched from Firebase to a self-hosted solution",
	"Show HN: Fossil delta compression for real-time sync",
	"The hidden costs of real-time: lessons from production",
	"How Figma handles multiplayer editing",
	"Ask HN: Best practices for fan-out in pub/sub systems?",
	"We serve 1M concurrent WebSocket connections on a single box",
	"The case for server-driven UI updates",
	"Building a collaborative whiteboard with CRDTs",
	"Real-time search suggestions: the engineering behind autocomplete",
	"Ask HN: How do you test real-time features?",
	"Show HN: A tiny Go library for HMAC-signed track tokens",
	"Why eventual consistency is good enough for most apps",
	"The evolution of chat infrastructure at Discord",
	"Scaling notifications: from polling to push",
	"Ask HN: What's your real-time stack in 2026?",
	"How we built live sports scores for 10M users",
	"The surprising performance of HTTP/2 server push",
	"Show HN: Real-time collaborative spreadsheet in 500 lines",
	"Understanding backpressure in streaming systems",
	"Why we moved from GraphQL subscriptions to WebSockets",
	"The art of debouncing: UI responsiveness vs server load",
	"Ask HN: How do you handle reconnection and state recovery?",
	"Building a stock ticker with sub-second latency",
	"Show HN: Open-source alternative to Pusher/Ably",
	"Real-time analytics dashboards without the complexity",
	"The problem with polling (and how shared polling solves it)",
	"How Notion keeps everything in sync",
	"Ask HN: Do you use optimistic updates in production?",
	"Shared polling: a new primitive for real-time at scale",
	"Why your WebSocket reconnection logic is probably wrong",
	"Show HN: Live cursors and presence in 50 lines of JS",
	"The cost of real-time: when to poll vs push",
	"Building multiplayer games with authoritative servers",
	"Ask HN: How do you version your real-time data?",
	"Delta compression saved us 60% bandwidth",
	"The subtle bugs in distributed presence tracking",
	"Show HN: I added real-time to my CRUD app in an afternoon",
	"Why message ordering matters more than you think",
	"How Linear built their real-time sync engine",
	"The future of real-time web: predictions for 2027",
	"Ask HN: What's the hardest real-time problem you've solved?",
}

func initRedis() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("failed to parse REDIS_URL: %v", err)
	}
	rdb = redis.NewClient(opts)
	ctx := context.Background()
	for i := 0; i < 30; i++ {
		if err := rdb.Ping(ctx).Err(); err == nil {
			break
		}
		log.Printf("waiting for redis... (%v)", err)
		time.Sleep(time.Second)
	}
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	log.Println("connected to Redis")
}

func notifySharedPoll(ctx context.Context, channel string, keys []string) {
	type notification struct {
		Channel string `json:"channel"`
		Key     string `json:"key"`
	}
	items := make([]notification, len(keys))
	for i, key := range keys {
		items[i] = notification{Channel: channel, Key: key}
	}
	data, err := json.Marshal(map[string]any{"items": items})
	if err != nil {
		log.Printf("failed to marshal notification: %v", err)
		return
	}
	if err := rdb.Publish(ctx, "shared_poll_notify", data).Err(); err != nil {
		log.Printf("failed to publish notification: %v", err)
	}
}

func initDB() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://demo:demo@localhost:5432/shared_poll?sslmode=disable"
	}

	ctx := context.Background()
	var err error
	for i := 0; i < 30; i++ {
		pool, err = pgxpool.New(ctx, dsn)
		if err == nil {
			err = pool.Ping(ctx)
		}
		if err == nil {
			break
		}
		log.Printf("waiting for postgres... (%v)", err)
		time.Sleep(time.Second)
	}
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS posts (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			votes BIGINT NOT NULL DEFAULT 0,
			version BIGINT NOT NULL DEFAULT 1
		)
	`)
	if err != nil {
		log.Fatalf("failed to create table: %v", err)
	}

	// Seed posts if table is empty.
	var count int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM posts").Scan(&count)
	if count == 0 {
		tx, err := pool.Begin(ctx)
		if err != nil {
			log.Fatalf("failed to begin tx: %v", err)
		}
		for i := range numPosts {
			_, err = tx.Exec(ctx,
				"INSERT INTO posts (id, title, votes, version) VALUES ($1, $2, 0, 1)",
				fmt.Sprintf("post_%d", i+1), titles[i%len(titles)],
			)
			if err != nil {
				tx.Rollback(ctx)
				log.Fatalf("failed to seed post: %v", err)
			}
		}
		if err := tx.Commit(ctx); err != nil {
			log.Fatalf("failed to commit seed: %v", err)
		}
		log.Printf("seeded %d posts", numPosts)
	}
}

func makeTrackSignature(secret, channel string, keys []string, user string, ttl int) string {
	now := time.Now().Unix()
	expiry := now + int64(ttl)
	sort.Strings(keys)
	keysHash := sha256.Sum256([]byte(strings.Join(keys, "\x00")))
	payload := fmt.Sprintf("%d:%d:%s:%s:%x", now, expiry, user, channel, keysHash)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return fmt.Sprintf("%d:%d:%x", now, expiry, mac.Sum(nil))
}

type postRow struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Votes   int64  `json:"votes"`
	Version uint64 `json:"version"`
}

// GET /api/posts?page=N — returns 10 posts per page + track signature.
func handleGetPosts(w http.ResponseWriter, r *http.Request) {
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if page < 1 {
		page = 1
	}

	ctx := r.Context()

	var totalCount int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM posts").Scan(&totalCount)
	totalPages := (totalCount + 9) / 10

	offset := (page - 1) * 10
	rows, err := pool.Query(ctx, "SELECT id, title, votes, version FROM posts ORDER BY id LIMIT 10 OFFSET $1", offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var pagePosts []postRow
	var keys []string
	for rows.Next() {
		var p postRow
		if err := rows.Scan(&p.ID, &p.Title, &p.Votes, &p.Version); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		pagePosts = append(pagePosts, p)
		keys = append(keys, p.ID)
	}
	if pagePosts == nil {
		pagePosts = []postRow{}
		keys = []string{}
	}

	channel := "post_votes:main"
	signature := makeTrackSignature(pollSecret, channel, keys, "", 3)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"posts":       pagePosts,
		"page":        page,
		"total_pages": totalPages,
		"signature":   signature,
		"keys":        keys,
	})
}

// POST /api/posts/:id/vote — increments vote count.
func handleVote(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/posts/")
	id = strings.TrimSuffix(id, "/vote")

	ctx := r.Context()
	var p postRow
	err := pool.QueryRow(ctx,
		"UPDATE posts SET votes = votes + 1, version = version + 1 WHERE id = $1 RETURNING id, title, votes, version",
		id,
	).Scan(&p.ID, &p.Title, &p.Votes, &p.Version)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Notify Centrifugo about the changed key for near-instant updates.
	notifySharedPoll(ctx, "post_votes:main", []string{p.ID})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"id":      p.ID,
		"votes":   p.Votes,
		"version": p.Version,
	})
}

// POST /api/track_refresh — signature refresh callback.
func handleTrackRefresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Keys    []string `json:"keys"`
		Channel string   `json:"channel"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if req.Channel == "" {
		req.Channel = "post_votes:main"
	}

	ctx := r.Context()
	validKeys := make([]string, 0, len(req.Keys))
	if len(req.Keys) > 0 {
		args := make([]any, len(req.Keys))
		placeholders := make([]string, len(req.Keys))
		for i, k := range req.Keys {
			placeholders[i] = fmt.Sprintf("$%d", i+1)
			args[i] = k
		}
		rows, err := pool.Query(ctx,
			"SELECT id FROM posts WHERE id = ANY($1) ORDER BY id",
			req.Keys,
		)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var id string
				rows.Scan(&id)
				validKeys = append(validKeys, id)
			}
		}
	}

	signature := makeTrackSignature(pollSecret, req.Channel, validKeys, "", 3)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"keys":      validKeys,
		"signature": signature,
	})
}

// POST /centrifugo/refresh — Centrifugo shared poll refresh proxy endpoint.
//
// In versioned mode, Centrifugo includes known versions in the request.
// The backend can use them to skip items whose version hasn't changed,
// reducing response size. If the backend ignores versions, it simply
// returns all items every cycle (which is also fine).
func handleCentrifugoRefresh(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Channel string `json:"channel"`
		Items   []struct {
			Key     string `json:"key"`
			Version uint64 `json:"version,omitempty"`
		} `json:"items"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	if len(req.Items) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"result": map[string]any{"items": []any{}}})
		return
	}

	ctx := r.Context()

	// Build lookup of requested versions (sent in versioned mode).
	reqVersions := make(map[string]uint64, len(req.Items))
	keys := make([]string, len(req.Items))
	for i, item := range req.Items {
		reqVersions[item.Key] = item.Version
		keys[i] = item.Key
	}

	rows, err := pool.Query(ctx,
		"SELECT id, title, votes, version FROM posts WHERE id = ANY($1) ORDER BY id",
		keys,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type respItem struct {
		Key     string `json:"key"`
		Data    any    `json:"data"`
		Version uint64 `json:"version"`
	}
	var items []respItem
	for rows.Next() {
		var p postRow
		if err := rows.Scan(&p.ID, &p.Title, &p.Votes, &p.Version); err != nil {
			continue
		}
		// Versioned mode: if the request includes a version and it matches, skip this item.
		if v, ok := reqVersions[p.ID]; ok && v > 0 && v == p.Version {
			continue
		}
		items = append(items, respItem{
			Key: p.ID,
			Data: map[string]any{
				"id":    p.ID,
				"title": p.Title,
				"votes": p.Votes,
			},
			Version: p.Version,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"result": map[string]any{
			"items": items,
		},
	})
}

func main() {
	initDB()
	initRedis()

	// Serve static files.
	fs := http.FileServer(http.Dir("../static"))
	http.Handle("/", fs)

	// API endpoints.
	http.HandleFunc("/api/posts", handleGetPosts)
	http.HandleFunc("/api/posts/", handleVote)
	http.HandleFunc("/api/track_refresh", handleTrackRefresh)

	// Centrifugo proxy endpoint.
	http.HandleFunc("/centrifugo/refresh", handleCentrifugoRefresh)

	log.Println("Backend listening on :3001")
	log.Fatal(http.ListenAndServe(":3001", nil))
}
