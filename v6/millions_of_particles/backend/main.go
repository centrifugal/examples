// 2 million particle multiplayer simulation served over Centrifugo.
//
// The backend runs the simulation locally and publishes a single shared
// viewport bitmap (1 bit per pixel) to the `particles:frame` channel
// every 30 fps via Centrifugo's HTTP API. All connected browsers
// subscribe to the same channel and render the same frame — they're
// looking at the same window into the world. Each client can also
// affect the simulation by sending an attractor position via RPC,
// which Centrifugo proxies to this backend.
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
	"syscall"
	"time"
)

const channel = "particles:frame"

func main() {
	apiURL := envOr("CENTRIFUGO_API_URL", "http://localhost:8000/api")
	apiKey := envOr("CENTRIFUGO_API_KEY", "particles-demo-api-key")

	cfg := SimConfig{
		WorldWidth:         envInt("WORLD_WIDTH", 2200),
		WorldHeight:        envInt("WORLD_HEIGHT", 2200),
		ParticleCount:      envInt("PARTICLE_COUNT", 2_000_000),
		ViewportX:          envInt("VIEWPORT_X", 0),
		ViewportY:          envInt("VIEWPORT_Y", 0),
		ViewportW:          envInt("VIEWPORT_W", 2200),
		ViewportH:          envInt("VIEWPORT_H", 2200),
		// 1 = full resolution (~605 KB/frame). 3 ≈ 9× smaller (~67 KB).
		Downsample:         envInt("DOWNSAMPLE", 3),
		FPS:                60,
		PublishEveryNTicks: 1, // publish every sim tick — 60 fps
	}

	api := NewCentrifugoAPI(apiURL, apiKey)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Printf("waiting for Centrifugo at %s ...", apiURL)
	if err := api.WaitReady(ctx); err != nil {
		log.Fatalf("Centrifugo not reachable: %v", err)
	}
	log.Printf("Centrifugo ready")

	sim := NewSim(cfg)

	// Publish off the sim's ticker goroutine — the HTTP round-trip can run
	// 5–10ms which would otherwise push tick work past the 16ms budget at
	// 60 Hz and force ticks to drop. Buffer of 2: the sim hands off and
	// keeps simulating; if the publisher falls behind we drop the older
	// pending frame rather than queue.
	frames := make(chan []byte, 2)
	go func() {
		for bitmap := range frames {
			if err := api.PublishBinary(ctx, channel, bitmap); err != nil && ctx.Err() == nil {
				log.Printf("publish err: %v", err)
			}
		}
	}()

	go sim.Run(ctx, func(bitmap []byte) {
		if ctx.Err() != nil {
			return
		}
		select {
		case frames <- bitmap:
		default:
			// Publisher is behind — drop the oldest pending frame and
			// queue the freshest. Viewer prefers latest over complete.
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

	mux := http.NewServeMux()
	mux.HandleFunc("/centrifugo/rpc", rpcHandler(sim))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		bw, bh := cfg.BitmapDims()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]int{
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
		})
	})

	server := &http.Server{Addr: ":3001", Handler: mux}
	go func() {
		log.Printf("backend listening on :3001 (channel %q, viewport %dx%d at %d,%d)",
			channel, cfg.ViewportW, cfg.ViewportH, cfg.ViewportX, cfg.ViewportY)
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
		log.Printf("rpc body: %s", string(raw))
		var body rpcReqEnvelope
		if err := json.Unmarshal(raw, &body); err != nil {
			respErr(w, 100, "bad request: "+err.Error())
			return
		}
		switch body.Method {
		case "input":
			var dataBytes []byte
			if len(body.Data) > 0 {
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
					log.Printf("input parse err: %v (raw: %s)", err, string(dataBytes))
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
