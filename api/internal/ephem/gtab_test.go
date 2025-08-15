package ephem

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeGTAB(t *testing.T, dir string, name string, epoch int64, dtNS int64, n uint32, fields uint32, tideVals []uint16) string {
    t.Helper()
    path := filepath.Join(dir, name)
    hdr := make([]byte, 5+2+8+8+4+4+16)
    copy(hdr[:5], []byte("GTAB1"))
    binary.LittleEndian.PutUint16(hdr[5:7], 1)
    binary.LittleEndian.PutUint64(hdr[7:15], uint64(epoch))
    binary.LittleEndian.PutUint64(hdr[15:23], uint64(dtNS))
    binary.LittleEndian.PutUint32(hdr[23:27], n)
    binary.LittleEndian.PutUint32(hdr[27:31], fields)
    // reserved already zeroed
    rec := []byte{}
    if fields&FieldTideBPS != 0 {
        rec = make([]byte, int(n)*2)
        for i, v := range tideVals {
            binary.LittleEndian.PutUint16(rec[i*2:(i+1)*2], v)
        }
    }
    if err := os.WriteFile(path, append(hdr, rec...), 0o644); err != nil { t.Fatalf("write: %v", err) }
    return path
}

func TestGTAB_Open_ValidAndLookup(t *testing.T) {
    dir := t.TempDir()
    epoch := time.Date(2025,8,1,0,0,0,0,time.UTC).Unix()
    // n=2, dt=1s, tide_bps present
    p := writeGTAB(t, dir, "ok.bin", epoch, 1_000_000_000, 2, FieldTideBPS, []uint16{100, 200})
    g, err := Open(p)
    if err != nil { t.Fatalf("open: %v", err) }
    t.Cleanup(func(){ _ = g.Close() })
    start, end := g.Coverage()
    if start.Unix() != epoch { t.Fatalf("start epoch mismatch: %v", start) }
    if end.Unix() != epoch+1 { t.Fatalf("end epoch mismatch: %v", end) }
    // exact sample 0
    if v, ok := g.LookupTideBPS(time.Unix(epoch,0).UTC()); !ok || v != 100 { t.Fatalf("lookup[0]: %v %v", v, ok) }
    // exact sample 1
    if v, ok := g.LookupTideBPS(time.Unix(epoch+1,0).UTC()); !ok || v != 200 { t.Fatalf("lookup[1]: %v %v", v, ok) }
    // halfway -> ~150
    if v, ok := g.LookupTideBPS(time.Unix(epoch, 500_000_000).UTC()); !ok || v < 149 || v > 151 { t.Fatalf("interp: %v %v", v, ok) }
    // out of range
    if _, ok := g.LookupTideBPS(time.Unix(epoch-1,0).UTC()); ok { t.Fatalf("expected out-of-range before start") }
    if _, ok := g.LookupTideBPS(time.Unix(epoch+2,0).UTC()); ok { t.Fatalf("expected out-of-range after end") }
}

func TestGTAB_Open_HeaderErrors(t *testing.T) {
    dir := t.TempDir()
    epoch := time.Now().UTC().Unix()
    // bad magic
    p := filepath.Join(dir, "badmagic.bin")
    b := make([]byte, 5+2+8+8+4+4+16)
    copy(b[:5], []byte("XXXXX"))
    os.WriteFile(p, b, 0o644)
    if _, err := Open(p); err == nil { t.Fatal("expected error for bad magic") }

    // bad version
    p2 := filepath.Join(dir, "badver.bin")
    copy(b[:5], []byte("GTAB1"))
    binary.LittleEndian.PutUint16(b[5:7], 2)
    os.WriteFile(p2, b, 0o644)
    if _, err := Open(p2); err == nil { t.Fatal("expected error for bad version") }

    // dt <= 0
    p3 := writeGTAB(t, dir, "baddt.bin", epoch, 0, 2, FieldTideBPS, []uint16{1,2})
    if _, err := Open(p3); err == nil { t.Fatal("expected error for dt<=0") }

    // n == 0
    p4 := writeGTAB(t, dir, "badn.bin", epoch, 1_000_000_000, 0, FieldTideBPS, nil)
    if _, err := Open(p4); err == nil { t.Fatal("expected error for n==0") }

    // fields_mask = 0
    p5 := writeGTAB(t, dir, "badmask.bin", epoch, 1_000_000_000, 2, 0, nil)
    if _, err := Open(p5); err == nil { t.Fatal("expected error for empty fields mask") }
}

func TestGTAB_Open_Truncated(t *testing.T) {
    dir := t.TempDir()
    epoch := time.Now().UTC().Unix()
    // Write a header indicating n=10 with tide_bps (2 bytes each) but cut data short
    p := writeGTAB(t, dir, "trunc.bin", epoch, 1_000_000_000, 10, FieldTideBPS, make([]uint16, 5))
    // Overwrite file to be smaller than expected by truncating tail
    fi, _ := os.Stat(p)
    os.Truncate(p, fi.Size()-2)
    if _, err := Open(p); err == nil { t.Fatal("expected truncated file error") }
}
