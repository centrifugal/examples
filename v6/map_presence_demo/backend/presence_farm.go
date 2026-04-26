package main

import (
	"context"
	"fmt"
	"log"
	"math/rand/v2"
	"sync"
	"time"
)

// presenceFarmConfig describes a synthetic map_clients presence load.
// Entries are pushed via the Centrifugo HTTP API — no real WebSocket
// clients — so a single browser tab can watch the entire population.
//
// After the initial bulk fill, churn runs as paired "replace" events: each
// tick removes one random live entry and publishes one random non-live
// entry from the fixed pool. Population stays exactly at InitialCount, and
// every key is reused over time so each grid cell flickers with activity.
type presenceFarmConfig struct {
	Channel      string
	PoolSize     int // total stable id pool — keys are c_0..c_<PoolSize-1>
	InitialCount int
	ChurnPerSec  int
}

func runPresenceFarm(ctx context.Context, api *CentrifugoAPI, cfg presenceFarmConfig) {
	if cfg.PoolSize <= 0 {
		cfg.PoolSize = 102400
	}
	if cfg.InitialCount <= 0 {
		cfg.InitialCount = cfg.PoolSize - cfg.PoolSize/40
	}
	if cfg.InitialCount > cfg.PoolSize {
		log.Printf("farm %s: count %d exceeds pool %d, clamping", cfg.Channel, cfg.InitialCount, cfg.PoolSize)
		cfg.InitialCount = cfg.PoolSize
	}
	if cfg.ChurnPerSec <= 0 {
		cfg.ChurnPerSec = 200
	}

	keyOf := func(idx int) string { return fmt.Sprintf("c_%d", idx) }

	publish := func(idx int) {
		if err := api.MapPublish(ctx, cfg.Channel, keyOf(idx), []byte(`{}`)); err != nil && ctx.Err() == nil {
			log.Printf("farm %s publish err: %v", cfg.Channel, err)
		}
	}

	remove := func(idx int) {
		if err := api.MapRemove(ctx, cfg.Channel, keyOf(idx)); err != nil && ctx.Err() == nil {
			log.Printf("farm %s remove err: %v", cfg.Channel, err)
		}
	}

	all := make([]int, cfg.PoolSize)
	for i := range all {
		all[i] = i
	}
	rand.Shuffle(len(all), func(i, j int) { all[i], all[j] = all[j], all[i] })

	var (
		mu      sync.Mutex
		live    = all[:cfg.InitialCount]
		notLive = all[cfg.InitialCount:]
	)

	popRandom := func(s *[]int) (int, bool) {
		if len(*s) == 0 {
			return 0, false
		}
		i := rand.IntN(len(*s))
		v := (*s)[i]
		(*s)[i] = (*s)[len(*s)-1]
		*s = (*s)[:len(*s)-1]
		return v, true
	}

	log.Printf("farm %s: populating %d entries (pool=%d) ...", cfg.Channel, cfg.InitialCount, cfg.PoolSize)
	start := time.Now()
	// Parallelize the initial bulk publish — over HTTP each call is ~1ms,
	// so a serial loop on 1M entries would take far too long.
	const workers = 32
	idxCh := make(chan int, workers*2)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range idxCh {
				publish(idx)
			}
		}()
	}
	for _, idx := range live {
		if ctx.Err() != nil {
			break
		}
		idxCh <- idx
	}
	close(idxCh)
	wg.Wait()
	log.Printf("farm %s: populated %d entries in %s", cfg.Channel, cfg.InitialCount, time.Since(start))

	if len(notLive) == 0 {
		<-ctx.Done()
		return
	}

	churnTick := time.NewTicker(time.Second / time.Duration(cfg.ChurnPerSec))
	defer churnTick.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-churnTick.C:
			mu.Lock()
			leaveIdx, lok := popRandom(&live)
			joinIdx, jok := popRandom(&notLive)
			if lok {
				notLive = append(notLive, leaveIdx)
			}
			if jok {
				live = append(live, joinIdx)
			}
			mu.Unlock()
			if lok {
				remove(leaveIdx)
			}
			if jok {
				publish(joinIdx)
			}
		}
	}
}
