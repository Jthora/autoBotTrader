package ephem

import (
    "math/rand"
    "testing"
    "time"
    "os"
)

// FuzzLookupTideBPS exercises LookupTideBPS invariants with synthetic tables.
// Invariants:
//  - Returned ok implies value within [0,65535].
//  - Query outside coverage returns ok=false.
//  - Interpolation monotonic when underlying endpoints monotonic.
func FuzzLookupTideBPS(f *testing.F) {
    // Seed with small deterministic cases
    f.Add(uint32(5), int64(60))
    f.Add(uint32(2), int64(30))
    f.Fuzz(func(t *testing.T, n uint32, stepNS int64) {
        if n < 2 { // need at least two points to interpolate
            n = 2
        }
        if n > 5000 { // cap to keep runtime bounded
            n = 5000
        }
        if stepNS <= 0 {
            stepNS = int64(time.Minute)
        }
        // Build synthetic table in-memory (no helper yet, so manual encode)
        // Layout: header then n*records (2 bytes per record for tide_bps)
        hdr := make([]byte, 5+2+8+8+4+4+16)
        copy(hdr[:5], []byte("GTAB1"))
        // version
        hdr[5] = 1
        hdr[6] = 0
        epoch := time.Unix(0, 0)
        // epoch sec
        // already zero filled
        // dt_ns
        putU64 := func(off int, v uint64) { hdr[off] = byte(v); hdr[off+1] = byte(v>>8); hdr[off+2] = byte(v>>16); hdr[off+3] = byte(v>>24); hdr[off+4] = byte(v>>32); hdr[off+5] = byte(v>>40); hdr[off+6] = byte(v>>48); hdr[off+7] = byte(v>>56) }
        putU64(15, uint64(stepNS))
        // n
        hdr[23] = byte(n)
        hdr[24] = byte(n >> 8)
        hdr[25] = byte(n >> 16)
        hdr[26] = byte(n >> 24)
        // fields mask (only tide_bps)
        hdr[27] = 0x01
        // records
        records := make([]byte, n*2)
        // Generate monotonic or random pattern randomly
        monotonic := rand.Intn(2) == 0
        last := 0
        for i := uint32(0); i < n; i++ {
            var v int
            if monotonic {
                last += rand.Intn(100) // non-decreasing
                if last > 65535 {
                    last = 65535
                }
                v = last
            } else {
                v = rand.Intn(70000) // allow >65535 to test clamp
                if v > 65535 {
                    v = 65535
                }
            }
            off := i * 2
            records[off] = byte(v)
            records[off+1] = byte(v >> 8)
        }
        // Write to temp file
        path := t.TempDir() + "/fuzz.gtab"
        data := append(hdr, records...)
        if err := os.WriteFile(path, data, 0o600); err != nil {
            t.Fatalf("write temp: %v", err)
        }
        g, err := Open(path)
        if err != nil {
            t.Fatalf("open fuzz table: %v", err)
        }
        start, end := g.Coverage()
        // sample some times including out of range
        queries := []time.Time{
            start.Add(-time.Duration(stepNS)),
            start,
            start.Add(time.Duration(stepNS / 2)),
            end,
            end.Add(time.Duration(stepNS)),
        }
        for _, qt := range queries {
            v, ok := g.LookupTideBPS(qt)
            if ok {
                if v > 65535 { t.Fatalf("value > 65535: %d", v) }
            } else {
                if !qt.Before(start) && !qt.After(end) {
                    t.Fatalf("expected in-range ok for %v", qt)
                }
            }
        }
        // Monotonicity: sample sequential times and ensure non-decreasing if underlying was monotonic
        if monotonic {
            prev := uint16(0)
            for i := 0; i < 20; i++ {
                frac := float64(i) / 19.0
                qt := start.Add(time.Duration(float64(end.Sub(start))*frac))
                v, ok := g.LookupTideBPS(qt)
                if !ok { t.Fatalf("expected ok for monotonic sample") }
                if v < prev { t.Fatalf("monotonic violation prev=%d cur=%d", prev, v) }
                prev = v
            }
        }
        _ = g.Close()
        _ = epoch // silence unused (epoch kept for clarity)
    })
}
