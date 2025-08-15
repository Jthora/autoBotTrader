package ephem

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// helper to build an in-temp GTAB with linear ramp bps
func benchWriteGTAB(b *testing.B, n int, dtNS int64) string {
    dir := b.TempDir()
    epoch := time.Now().UTC().Unix()
    hdr := make([]byte, 5+2+8+8+4+4+16)
    copy(hdr[:5], []byte("GTAB1"))
    binary.LittleEndian.PutUint16(hdr[5:7], 1)
    binary.LittleEndian.PutUint64(hdr[7:15], uint64(epoch))
    binary.LittleEndian.PutUint64(hdr[15:23], uint64(dtNS))
    binary.LittleEndian.PutUint32(hdr[23:27], uint32(n))
    binary.LittleEndian.PutUint32(hdr[27:31], 1) // tide_bps
    rec := make([]byte, n*2)
    for i := 0; i < n; i++ {
        v := uint16((i * 10000) / (n - 1))
        binary.LittleEndian.PutUint16(rec[i*2:(i+1)*2], v)
    }
    path := filepath.Join(dir, "bench.bin")
    if err := os.WriteFile(path, append(hdr, rec...), 0o644); err != nil { b.Fatalf("write: %v", err) }
    return path
}

func BenchmarkGTAB_Lookup(b *testing.B) {
    path := benchWriteGTAB(b, 60*60, 1_000_000_000) // 1 hour @1s
    g, err := Open(path)
    if err != nil { b.Fatalf("open: %v", err) }
    defer g.Close()
    start, _ := g.Coverage()
    ts := make([]time.Time, b.N)
    for i := 0; i < b.N; i++ {
        ts[i] = start.Add(time.Duration(i%3600) * time.Second)
    }
    b.ResetTimer()
    var sum uint32
    for i := 0; i < b.N; i++ {
        v, _ := g.LookupTideBPS(ts[i])
        sum += uint32(v)
    }
    _ = sum
}
