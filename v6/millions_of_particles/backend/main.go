// 2 million particle multiplayer simulation served over Centrifugo.
//
// Two modes selectable via the MODE env var:
//
//   - fanout (default) — one shared bitmap of the world is published to
//     a single channel; every viewer receives the same bytes. Browser
//     window clips the centered canvas. No panning.
//
//   - shared_poll — the world is split into a 16×16 tile grid; each tile
//     is one key on a shared-poll subscription. Each viewer tracks the
//     tiles its viewport intersects (with a prefetch margin). Backend
//     publishes all tiles each tick via the batch API; Centrifugo fans
//     out only the tiles each viewer tracks. Pan/zoom over the world.
//
// Adapted from https://dgerrells.com/blog/how-fast-is-go-simulating-millions-of-particles-on-a-smart-tv
// — see sim.go for the simulation code (lifted with minor changes from
// https://github.com/dgerrells/how-fast-is-it).
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	channelFanout     = "particles:frame"
	channelSharedPoll = "tiles:world"
)

func main() {
	apiURL := envOr("CENTRIFUGO_API_URL", "http://localhost:8000/api")
	apiKey := envOr("CENTRIFUGO_API_KEY", "particles-demo-api-key")
	mode := envOr("MODE", "fanout")
	hmacSecret := envOr("HMAC_SECRET", "particles-demo-hmac-secret")

	if mode != "fanout" && mode != "shared_poll" {
		log.Fatalf("MODE must be 'fanout' or 'shared_poll', got %q", mode)
	}

	cfg := SimConfig{
		WorldWidth:         envInt("WORLD_WIDTH", 2200),
		WorldHeight:        envInt("WORLD_HEIGHT", 2200),
		ParticleCount:      envInt("PARTICLE_COUNT", 2_000_000),
		ViewportX:          envInt("VIEWPORT_X", 0),
		ViewportY:          envInt("VIEWPORT_Y", 0),
		ViewportW:          envInt("VIEWPORT_W", 2200),
		ViewportH:          envInt("VIEWPORT_H", 2200),
		Downsample:         envInt("DOWNSAMPLE", 1),
		FPS:                60,
		PublishEveryNTicks: envInt("PUBLISH_EVERY_N_TICKS", 2), // sim 60 Hz, publish 30 Hz
	}

	api := NewCentrifugoAPI(apiURL, apiKey)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Printf("waiting for Centrifugo at %s ...", apiURL)
	if err := api.WaitReady(ctx); err != nil {
		log.Fatalf("Centrifugo not reachable: %v", err)
	}
	log.Printf("Centrifugo ready (mode=%s)", mode)

	sim := NewSim(cfg)
	mux := http.NewServeMux()
	mux.HandleFunc("/centrifugo/rpc", rpcHandler(sim))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) })
	mux.HandleFunc("/config", configHandler(cfg, mode))

	switch mode {
	case "fanout":
		startFanoutMode(ctx, api, sim, cfg)
	case "shared_poll":
		startSharedPollMode(ctx, api, sim, cfg, hmacSecret, mux)
	}

	server := &http.Server{Addr: ":3001", Handler: mux}
	go func() {
		log.Printf("backend listening on :3001 (mode=%s)", mode)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("shutting down")

	cancel()
	shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutCancel()
	_ = server.Shutdown(shutCtx)
}

// ---------- Fan-out mode: one shared bitmap, one channel ----------

func startFanoutMode(ctx context.Context, api *CentrifugoAPI, sim *Sim, cfg SimConfig) {
	frames := make(chan []byte, 2)
	go func() {
		for bitmap := range frames {
			if err := api.PublishBinary(ctx, channelFanout, bitmap); err != nil && ctx.Err() == nil {
				log.Printf("publish err: %v", err)
			}
		}
	}()

	go sim.Run(ctx, func(worldBuf []byte) {
		if ctx.Err() != nil {
			return
		}
		bitmap := packViewport(worldBuf, cfg)
		select {
		case frames <- bitmap:
		default:
			// Drop oldest, queue freshest.
			select {
			case <-frames:
			default:
			}
			select {
			case frames <- bitmap:
			default:
			}
		}
	})
}

// ---------- Shared-poll mode: 16×16 tile grid, batch publish ----------

func startSharedPollMode(ctx context.Context, api *CentrifugoAPI, sim *Sim, cfg SimConfig, hmacSecret string, mux *http.ServeMux) {
	// Single monotonic version counter shared by all tiles per tick.
	// Seeded from the current Unix time in milliseconds so versions
	// always advance across backend restarts. Without this seed, an
	// in-process counter would reset to 0 on restart while Centrifugo
	// still holds the higher version from before — clients tracking
	// with cached high versions would never see fresh updates and
	// the picture would freeze until they re-track.
	var globalVersion uint64 = uint64(time.Now().UnixMilli())

	mux.HandleFunc("/api/tiles", tilesHandler(sim, cfg))
	mux.HandleFunc("/api/track_refresh", trackRefreshHandler(hmacSecret, channelSharedPoll))
	mux.HandleFunc("/centrifugo/refresh", sharedPollRefreshHandler(sim, cfg))

	frames := make(chan [][]byte, 2)
	go func() {
		for tilePayloads := range frames {
			v := atomic.AddUint64(&globalVersion, 1)
			items := make([]SharedPollItem, 0, len(tilePayloads))
			for i, payload := range tilePayloads {
				items = append(items, SharedPollItem{
					Key:     TileKey(i%TilesPerSide, i/TilesPerSide),
					Data:    payload,
					Version: v,
				})
			}
			if err := api.BatchSharedPollPublish(ctx, channelSharedPoll, items); err != nil && ctx.Err() == nil {
				log.Printf("batch publish err: %v", err)
			}
		}
	}()

	go sim.Run(ctx, func(worldBuf []byte) {
		if ctx.Err() != nil {
			return
		}
		tiles := PackAllTiles(worldBuf, cfg.WorldWidth, cfg.WorldHeight)
		select {
		case frames <- tiles:
		default:
			select {
			case <-frames:
			default:
			}
			select {
			case frames <- tiles:
			default:
			}
		}
	})
}

// /api/tiles?keys=t_0_0,t_0_1,... — returns initial tile data + signature
// for the centrifuge-js getSignature callback.
func tilesHandler(sim *Sim, cfg SimConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// We don't need actual initial state for the demo — tiles republish
		// every tick anyway. Just return signature.
		// (We could fetch worldBuf snapshot here, but it's racy without sim
		// integration; cold-key auto-poll triggers a fresh publish quickly.)
		_ = sim
		_ = cfg
		_ = r
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}
}

// /api/track_refresh — body { keys, channel } → { keys, signature }.
// Called by the SDK's getSignature callback whenever the tracked-key set
// changes.
func trackRefreshHandler(secret, channel string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Keys    []string `json:"keys"`
			Channel string   `json:"channel"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		ch := req.Channel
		if ch == "" {
			ch = channel
		}
		sig := MakeTrackSignature(secret, ch, req.Keys, "", 60)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"keys":      req.Keys,
			"signature": sig,
		})
	}
}

// /centrifugo/refresh — Centrifugo's safety-net poll proxy. Returns
// current tile payloads + versions for keys requested by Centrifugo.
// Called on cold-key tracking and on the configured refresh interval.
func sharedPollRefreshHandler(sim *Sim, cfg SimConfig) http.HandlerFunc {
	type respItem struct {
		Key     string `json:"key"`
		B64Data string `json:"b64data"`
		Version uint64 `json:"version"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Channel string `json:"channel"`
			Items   []struct {
				Key string `json:"key"`
			} `json:"items"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		// We don't have direct sync access to worldBuf here without locking
		// the sim. The fast path is shared_poll_publish from the sim
		// callback; this refresh proxy is just a safety net. Return empty
		// items — Centrifugo will deliver the next published tile shortly.
		_ = sim
		_ = cfg
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"result": map[string]any{"items": []respItem{}},
		})
	}
}

// ---------- Common HTTP handlers ----------

func configHandler(cfg SimConfig, mode string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		bw, bh := cfg.BitmapDims()
		out := map[string]any{
			"mode":          mode,
			"worldWidth":    cfg.WorldWidth,
			"worldHeight":   cfg.WorldHeight,
			"viewportX":     cfg.ViewportX,
			"viewportY":     cfg.ViewportY,
			"viewportW":     cfg.ViewportW,
			"viewportH":     cfg.ViewportH,
			"bitmapW":       bw,
			"bitmapH":       bh,
			"downsample":    cfg.Downsample,
			"particleCount": cfg.ParticleCount,
		}
		if mode == "shared_poll" {
			out["channel"] = channelSharedPoll
			out["tilesPerSide"] = TilesPerSide
			out["tileWorldSide"] = TileWorldSide
			out["tilePackedWidth"] = TilePackedWidth
			out["tilePackedRowBytes"] = TilePackedRowBytes
		} else {
			out["channel"] = channelFanout
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	}
}

func rpcHandler(sim *Sim) http.HandlerFunc {
	type rpcReqEnvelope struct {
		Method  string          `json:"method"`
		Data    json.RawMessage `json:"data"`
		B64Data string          `json:"b64data"`
		Client  string          `json:"client"`
		User    string          `json:"user"`
	}
	type inputBody struct {
		X    float32 `json:"x"`
		Y    float32 `json:"y"`
		Down bool    `json:"down"`
	}

	respOK := func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":{"data":{"ok":true}}}`))
	}
	respErr := func(w http.ResponseWriter, code int, msg string) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"code": code, "message": msg},
		})
	}

	return func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body rpcReqEnvelope
		if err := json.Unmarshal(raw, &body); err != nil {
			respErr(w, 100, "bad request: "+err.Error())
			return
		}
		switch body.Method {
		case "input":
			var dataBytes []byte
			if len(body.Data) > 0 && !strings.EqualFold(string(body.Data), "null") {
				dataBytes = body.Data
			} else if body.B64Data != "" {
				decoded, err := base64.StdEncoding.DecodeString(body.B64Data)
				if err != nil {
					respErr(w, 100, "bad b64data: "+err.Error())
					return
				}
				dataBytes = decoded
			}
			var in inputBody
			if len(dataBytes) > 0 {
				if err := json.Unmarshal(dataBytes, &in); err != nil {
					respErr(w, 100, "bad input data: "+err.Error())
					return
				}
			}
			sim.SetInput(body.Client, in.X, in.Y, in.Down)
			respOK(w)
		default:
			respErr(w, 404, "unknown method: "+body.Method)
		}
	}
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func envInt(k string, d int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return d
}
