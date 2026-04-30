package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"
)

// Tile layout for shared-poll mode. The world is split into a grid of
// fixed-size tiles; each tile becomes one shared-poll key. Viewers track
// only the tiles their viewport intersects; the backend publishes all
// tiles every tick via Centrifugo's batch API.
const (
	TilesPerSide       = 32                                // 32 × 32 = 1024 tiles per world
	TileWorldSide      = 69                                // ceil(2200 / 32); last row/col tiles overhang slightly
	TilePackedWidth    = 72                                // round up to multiple of 8 for clean per-row byte packing
	TilePackedRowBytes = TilePackedWidth / 8               // 9
	TilePackedSize     = TilePackedRowBytes * TileWorldSide // 9 × 69 = 621 bytes per tile
)

// TileKey returns the shared-poll key for tile (tx, ty).
func TileKey(tx, ty int) string {
	return fmt.Sprintf("t_%d_%d", tx, ty)
}

// PackTile extracts a TileWorldSide × TileWorldSide rectangle from the
// world buffer at tile (tx, ty) and packs it as 1 bit per pixel, padded
// to TilePackedWidth columns (multiple of 8).
//
// Uses the 8-byte LUT trick from packFull for interior 8-cell chunks
// (one lookup per 8 cells); the tail near the world boundary falls
// through to a per-cell loop. Cells past the world boundary stay zero.
func PackTile(worldBuf []byte, worldW, worldH, tx, ty int) []byte {
	lut := getPackLUT()
	out := make([]byte, TilePackedSize)
	startX := tx * TileWorldSide
	startY := ty * TileWorldSide

	for row := 0; row < TileWorldSide; row++ {
		wy := startY + row
		if wy >= worldH {
			break // remaining rows stay zero
		}
		rowOff := row * TilePackedRowBytes
		worldRow := wy * worldW

		col := 0
		// Fast path: full 8-cell chunks while we have a complete
		// in-bounds 8-byte window in the world buffer.
		for col+8 <= TileWorldSide && startX+col+8 <= worldW {
			i := worldRow + startX + col
			key := bytesToUint64Unsafe(worldBuf[i : i+8])
			out[rowOff+(col>>3)] = lut[key]
			col += 8
		}
		// Tail: partial chunk at the world boundary or tile edge.
		for ; col < TileWorldSide; col++ {
			wx := startX + col
			if wx >= worldW {
				break
			}
			if worldBuf[worldRow+wx] != 0 {
				out[rowOff+(col>>3)] |= 1 << (col & 7)
			}
		}
	}
	return out
}

// PackAllTiles packs every tile in the grid. The result preserves index
// order (row-major: ty=0 row first), so the caller can correlate with
// per-tile version counters.
func PackAllTiles(worldBuf []byte, worldW, worldH int) [][]byte {
	out := make([][]byte, TilesPerSide*TilesPerSide)
	for ty := 0; ty < TilesPerSide; ty++ {
		for tx := 0; tx < TilesPerSide; tx++ {
			out[ty*TilesPerSide+tx] = PackTile(worldBuf, worldW, worldH, tx, ty)
		}
	}
	return out
}

// MakeTrackSignature signs a (channel, keys, user, expiry) tuple for the
// centrifuge-js shared-poll `getSignature` callback. Keys are hashed in
// the order they appear; the backend must return them in the same order.
// Format matches Centrifugo's expected `<now>:<expiry>:<hex(hmac)>`.
func MakeTrackSignature(secret, channel string, keys []string, user string, ttlSec int) string {
	now := time.Now().Unix()
	expiry := now + int64(ttlSec)
	keysHash := sha256.Sum256([]byte(strings.Join(keys, "\x00")))
	payload := fmt.Sprintf("%d:%d:%s:%s:%x", now, expiry, user, channel, keysHash)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return fmt.Sprintf("%d:%d:%x", now, expiry, mac.Sum(nil))
}
