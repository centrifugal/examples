// Particle simulation lifted from
// https://github.com/dgerrells/how-fast-is-it/tree/main/go-land
// (MIT-style license, see referenced blog post for context).
//
// Key changes vs. the original:
//   - Single shared viewport, not per-client cameras (every browser tab
//     sees the same crop of the world). This lets one published frame
//     fan out to all subscribers via Centrifugo pub/sub.
//   - Inputs keyed by Centrifugo client id (string) instead of a fixed
//     array slot, with TTL-based pruning to drop disconnected clients.
//   - Frame bytes are returned via callback instead of being shipped on
//     a raw WebSocket; the caller publishes them to Centrifugo.
package main

import (
	"context"
	"log"
	"math"
	"math/rand"
	"runtime"
	"sync"
	"time"
	"unsafe"
)

type Particle struct {
	x, y, dx, dy float32
}

type Input struct {
	X, Y       float32
	IsTouchDown bool
	updated    time.Time
}

type SimConfig struct {
	WorldWidth, WorldHeight int
	ParticleCount           int
	ViewportX, ViewportY    int
	ViewportW, ViewportH    int
	// Downsample factor. 1 = pack one output cell per world pixel (full
	// resolution). K = OR over each K×K world block — output bitmap is
	// 1/K² the size, blurry per-particle but full-world coverage.
	Downsample         int
	FPS                int
	PublishEveryNTicks int
	InputTTL           time.Duration
}

// BitmapDims returns the dimensions of the packed bitmap the viewer
// receives — output cells along each axis after downsampling, with
// width rounded up to a multiple of 8 for clean bit-packing. The
// padded columns at the right edge are always zero (no world data).
func (cfg SimConfig) BitmapDims() (w, h int) {
	K := cfg.Downsample
	if K < 1 {
		K = 1
	}
	wOut := (cfg.ViewportW + K - 1) / K
	hOut := (cfg.ViewportH + K - 1) / K
	wPadded := ((wOut + 7) / 8) * 8
	return wPadded, hOut
}

type Sim struct {
	cfg SimConfig

	particles []Particle

	mu     sync.Mutex
	inputs map[string]*Input

	frameCount uint64
}

type simJob struct {
	startIdx, endIdx int
	dt               float32
	inputs           []Input
	writeWorldBuf    bool
}

func NewSim(cfg SimConfig) *Sim {
	if cfg.FPS <= 0 {
		cfg.FPS = 60
	}
	if cfg.PublishEveryNTicks <= 0 {
		cfg.PublishEveryNTicks = 2
	}
	if cfg.InputTTL <= 0 {
		cfg.InputTTL = 2 * time.Second
	}

	s := &Sim{
		cfg:       cfg,
		particles: make([]Particle, cfg.ParticleCount),
		inputs:    make(map[string]*Input),
	}
	w := float32(cfg.WorldWidth)
	h := float32(cfg.WorldHeight)
	for i := range s.particles {
		s.particles[i].x = rand.Float32() * w
		s.particles[i].y = rand.Float32() * h
	}
	return s
}

// SetInput updates a client's attractor position.
func (s *Sim) SetInput(clientID string, x, y float32, down bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	in, ok := s.inputs[clientID]
	if !ok {
		in = &Input{}
		s.inputs[clientID] = in
	}
	in.X = x
	in.Y = y
	in.IsTouchDown = down
	in.updated = time.Now()
}

// snapshotInputs returns a copy of currently-valid attractors and prunes stale ones.
func (s *Sim) snapshotInputs() []Input {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	active := make([]Input, 0, len(s.inputs))
	for id, in := range s.inputs {
		if now.Sub(in.updated) > s.cfg.InputTTL {
			delete(s.inputs, id)
			continue
		}
		if in.IsTouchDown {
			active = append(active, *in)
		}
	}
	return active
}

// Run drives the simulation loop. onTick is called every PublishEveryNTicks
// ticks with the fresh world buffer (one byte per world pixel, 0 or 1).
// The callback runs synchronously inside the tick — it's the caller's job
// to pack/copy whatever it needs and return quickly. The buffer is reused
// next tick.
func (s *Sim) Run(ctx context.Context, onTick func([]byte)) {
	numThreads := int(math.Min(math.Max(float64(runtime.NumCPU()-1), 1), 8))
	particlesPerThread := s.cfg.ParticleCount / numThreads

	worldArea := s.cfg.WorldWidth * s.cfg.WorldHeight
	worldBuf := make([]byte, worldArea)

	ticker := time.NewTicker(time.Second / time.Duration(s.cfg.FPS))
	defer ticker.Stop()
	lastTime := time.Now()

	jobs := make(chan simJob, numThreads)
	var wg sync.WaitGroup
	for i := 0; i < numThreads; i++ {
		go s.worker(jobs, &wg, worldBuf)
	}

	for {
		select {
		case <-ctx.Done():
			close(jobs)
			return
		case now := <-ticker.C:
			s.frameCount++
			dt := float32(now.Sub(lastTime).Seconds())
			lastTime = now

			if s.frameCount%30 == 0 {
				log.Printf("FPS: %.1f  inputs: %d", 1/dt, len(s.inputs))
			}

			willPublish := s.frameCount%uint64(s.cfg.PublishEveryNTicks) == 0
			if willPublish {
				for i := range worldBuf {
					worldBuf[i] = 0
				}
			}

			active := s.snapshotInputs()

			wg.Add(numThreads)
			for i := 0; i < numThreads; i++ {
				start := i * particlesPerThread
				end := start + particlesPerThread
				if i == numThreads-1 {
					end = s.cfg.ParticleCount
				}
				jobs <- simJob{
					startIdx:      start,
					endIdx:        end,
					dt:            dt,
					inputs:        active,
					writeWorldBuf: willPublish,
				}
			}
			wg.Wait()

			if willPublish {
				onTick(worldBuf)
			}
		}
	}
}

func (s *Sim) worker(jobs <-chan simJob, wg *sync.WaitGroup, worldBuf []byte) {
	for j := range jobs {
		s.runJob(j, worldBuf)
		wg.Done()
	}
}

const friction float32 = 0.988

func (s *Sim) runJob(j simJob, worldBuf []byte) {
	frictionFactor := float32(math.Pow(float64(friction), float64(j.dt*60)))
	gravPower := j.dt * 5
	const pullDist float32 = 32300
	w := float32(s.cfg.WorldWidth)
	h := float32(s.cfg.WorldHeight)

	for i := j.startIdx; i < j.endIdx; i++ {
		p := &s.particles[i]
		for _, in := range j.inputs {
			dx := in.X - p.x
			dy := in.Y - p.y
			dist := dx*dx + dy*dy
			if dist < pullDist && dist > 1 {
				grav := 4 / float32(math.Sqrt(float64(dist)))
				p.dx += dx * gravPower * grav
				p.dy += dy * gravPower * grav
			}
		}
		p.x += p.dx
		p.y += p.dy
		p.dx *= frictionFactor
		p.dy *= frictionFactor

		if p.x < 0 || p.x >= w {
			p.x -= p.dx
			p.dx *= -1
		}
		if p.y < 0 || p.y >= h {
			p.y -= p.dy
			p.dy *= -1
		}

		if j.writeWorldBuf && p.x >= 0 && p.x < w && p.y >= 0 && p.y < h {
			x := uint32(p.x)
			y := uint32(p.y)
			worldBuf[y*uint32(s.cfg.WorldWidth)+x] = 1
		}
	}
}

// uint64ToByte LUT — packs 8 consecutive 0/1 bytes (one byte per pixel
// in worldBuf) into a single bit-per-pixel byte using the original
// trick from the source repo.
var uint64ToByteLUT [256]struct{}

func init() {} // placeholder — actual LUT built lazily in packViewport

// bytesToUint64Unsafe interprets b as a little-endian uint64 without copying.
// Caller must guarantee len(b) >= 8.
func bytesToUint64Unsafe(b []byte) uint64 {
	return *(*uint64)(unsafe.Pointer(&b[0]))
}

// packViewport extracts the configured rectangle from worldBuf (1 byte per
// pixel, 0 or 1) and packs it into a 1-bpp bitmap. With Downsample > 1,
// each output cell is the OR of a K×K world block — which loses
// per-particle precision in dense regions but cuts bytes by K² and lets
// the same channel publish carry the entire world to every viewer at a
// fraction of the full-resolution cost.
func packViewport(worldBuf []byte, cfg SimConfig) []byte {
	x, y := cfg.ViewportX, cfg.ViewportY
	width, height := cfg.ViewportW, cfg.ViewportH
	if x < 0 || y < 0 || width <= 0 || height <= 0 ||
		x+width > cfg.WorldWidth || y+height > cfg.WorldHeight {
		return nil
	}
	K := cfg.Downsample
	if K < 1 {
		K = 1
	}
	if K == 1 {
		return packFull(worldBuf, cfg)
	}
	return packDownsampled(worldBuf, cfg, K)
}

// packFull is the fast path: output cell == world pixel, 8-byte LUT pack
// per row chunk. Requires width to be a multiple of 8.
func packFull(worldBuf []byte, cfg SimConfig) []byte {
	x, y := cfg.ViewportX, cfg.ViewportY
	width, height := cfg.ViewportW, cfg.ViewportH
	lut := getPackLUT()
	out := make([]byte, (width*height+7)/8)
	idx := 0
	for row := 0; row < height; row++ {
		yOff := (y + row) * cfg.WorldWidth
		for col := 0; col < width; col += 8 {
			i := yOff + (x + col)
			key := bytesToUint64Unsafe(worldBuf[i : i+8])
			out[idx] = lut[key]
			idx++
		}
	}
	return out
}

// packDownsampled writes one output bit per K×K world block (set if any
// world pixel in the block is nonzero). Output width is rounded up to
// a multiple of 8 with the trailing columns left as zero (no world data
// there) so the on-wire format is byte-aligned per row and the decoder
// stays a flat per-byte unpack loop.
func packDownsampled(worldBuf []byte, cfg SimConfig, K int) []byte {
	bw, bh := cfg.BitmapDims()
	wOut := (cfg.ViewportW + K - 1) / K
	bytesPerRow := bw / 8

	out := make([]byte, bytesPerRow*bh)

	for outY := 0; outY < bh; outY++ {
		startWY := cfg.ViewportY + outY*K
		rowOff := outY * bytesPerRow
		for outX := 0; outX < wOut; outX++ {
			startWX := cfg.ViewportX + outX*K
			on := false
		check:
			for dy := 0; dy < K; dy++ {
				wy := startWY + dy
				if wy >= cfg.WorldHeight {
					break
				}
				rowW := wy * cfg.WorldWidth
				for dx := 0; dx < K; dx++ {
					wx := startWX + dx
					if wx >= cfg.WorldWidth {
						break
					}
					if worldBuf[rowW+wx] != 0 {
						on = true
						break check
					}
				}
			}
			if on {
				out[rowOff+outX>>3] |= 1 << (outX & 7)
			}
		}
	}
	return out
}

var (
	packLUT     map[uint64]byte
	packLUTOnce sync.Once
)

func getPackLUT() map[uint64]byte {
	packLUTOnce.Do(func() {
		packLUT = make(map[uint64]byte, 256)
		buf := make([]byte, 8)
		for v := 0; v < 256; v++ {
			for bit := 0; bit < 8; bit++ {
				if (v>>bit)&1 == 1 {
					buf[bit] = 1
				} else {
					buf[bit] = 0
				}
			}
			packLUT[bytesToUint64Unsafe(buf)] = byte(v)
		}
	})
	return packLUT
}
