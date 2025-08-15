package ephem

import (
    "encoding/binary"
    "net"
    "os"
    "path/filepath"
    "testing"
    "time"
)

func writeGTABValues(t *testing.T, dir string, name string, epoch int64, dtNS int64, vals []uint16) string {
    t.Helper()
    n := uint32(len(vals))
    hdr := make([]byte, 5+2+8+8+4+4+16)
    copy(hdr[:5], []byte("GTAB1"))
    binary.LittleEndian.PutUint16(hdr[5:7], 1)
    binary.LittleEndian.PutUint64(hdr[7:15], uint64(epoch))
    binary.LittleEndian.PutUint64(hdr[15:23], uint64(dtNS))
    binary.LittleEndian.PutUint32(hdr[23:27], n)
    binary.LittleEndian.PutUint32(hdr[27:31], FieldTideBPS)
    rec := make([]byte, n*2)
    for i,v := range vals { binary.LittleEndian.PutUint16(rec[i*2:(i+1)*2], v) }
    p := filepath.Join(dir, name)
    if err := os.WriteFile(p, append(hdr, rec...), 0o644); err != nil { t.Fatalf("write: %v", err) }
    return p
}

func TestLookupTideBPSBoundaryAndInterpolationRounding(t *testing.T) {
    dir := t.TempDir()
    epoch := time.Unix(1_726_000_000,0).UTC().Unix()
    p := writeGTABValues(t, dir, "vals.bin", epoch, 1_000_000_000, []uint16{100,102})
    g, err := Open(p); if err != nil { t.Fatalf("open: %v", err) }
    t.Cleanup(func(){ _ = g.Close() })
    if v, ok := g.LookupTideBPS(time.Unix(epoch, 500_000_000).UTC()); !ok || v != 101 { t.Fatalf("midpoint round expected 101 got %d ok=%v", v, ok) }
    if _, ok := g.LookupTideBPS(time.Unix(epoch-1,0).UTC()); ok { t.Fatal("expected before start false") }
    if _, ok := g.LookupTideBPS(time.Unix(epoch+2,0).UTC()); ok { t.Fatal("expected after end false") }
}

func TestMissingTideFieldReturnsFalse(t *testing.T) {
    dir := t.TempDir()
    hdr := make([]byte, 5+2+8+8+4+4+16)
    copy(hdr[:5], []byte("GTAB1"))
    binary.LittleEndian.PutUint16(hdr[5:7], 1)
    epoch := time.Unix(1_726_000_100,0).UTC().Unix()
    binary.LittleEndian.PutUint64(hdr[7:15], uint64(epoch))
    binary.LittleEndian.PutUint64(hdr[15:23], uint64(1_000_000_000))
    binary.LittleEndian.PutUint32(hdr[23:27], 2)
    binary.LittleEndian.PutUint32(hdr[27:31], 0)
    p := filepath.Join(dir, "nofield.bin")
    if err := os.WriteFile(p, hdr, 0o644); err != nil { t.Fatalf("write: %v", err) }
    if _, err := Open(p); err == nil { t.Fatalf("expected error for empty fields mask") }
}

func TestGTABAccessAfterCloseSafe(t *testing.T) {
    dir := t.TempDir()
    epoch := time.Unix(1_726_000_500,0).UTC().Unix()
    p := writeGTABValues(t, dir, "vals.bin", epoch, 1_000_000_000, []uint16{10,20,30})
    g, err := Open(p); if err != nil { t.Fatalf("open: %v", err) }
    if err := g.Close(); err != nil { t.Fatalf("close: %v", err) }
    _, _ = g.LookupTideBPS(time.Unix(epoch,0).UTC())
}

func unusedPort() string {
    l, _ := net.Listen("tcp", "127.0.0.1:0")
    if l == nil { return "127.0.0.1:65534" }
    addr := l.Addr().String()
    _ = l.Close()
    return addr
}
