package providers

import (
	"context"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Jthora/autoBotTrader/api/internal/ephem"
)

func writeGTABWithBPS(t *testing.T, dir, name string, epoch int64, vals []uint16) string {
    t.Helper()
    hdr := make([]byte, 5+2+8+8+4+4+16)
    copy(hdr[:5], []byte("GTAB1"))
    binary.LittleEndian.PutUint16(hdr[5:7], 1)
    binary.LittleEndian.PutUint64(hdr[7:15], uint64(epoch))
    binary.LittleEndian.PutUint64(hdr[15:23], uint64(1_000_000_000))
    binary.LittleEndian.PutUint32(hdr[23:27], uint32(len(vals)))
    binary.LittleEndian.PutUint32(hdr[27:31], ephem.FieldTideBPS)
    rec := make([]byte, len(vals)*2)
    for i, v := range vals { binary.LittleEndian.PutUint16(rec[i*2:(i+1)*2], v) }
    p := filepath.Join(dir, name)
    if err := os.WriteFile(p, append(hdr, rec...), 0o644); err != nil { t.Fatalf("write: %v", err) }
    return p
}

func writeMeta(t *testing.T, dir string, datasetID string) {
    t.Helper()
    _ = os.WriteFile(filepath.Join(dir, "gtab.meta.json"), []byte("{\n  \"dataset_id\": \""+datasetID+"\"\n}\n"), 0o644)
}

func TestFileGravimetric_BasicMappingAndHysteresis(t *testing.T) {
    dir := t.TempDir()
    epoch := time.Date(2025,8,1,0,0,0,0,time.UTC).Unix()
    table := writeGTABWithBPS(t, dir, "gtab_1s.bin", epoch, []uint16{0, 10, 5000, 10000})
    writeMeta(t, dir, "ds123")
    os.Setenv("HYSTERESIS_BPS", "100")
    t.Cleanup(func(){ os.Unsetenv("HYSTERESIS_BPS") })
    prov, err := NewFileGravimetric(table, "")
    if err != nil { t.Fatalf("new provider: %v", err) }
    t.Cleanup(func(){ _ = prov.Close() })
    // First value (sentinel) should not stick; mapping 0 -> 80.0
    v, err := prov.FetchAt(time.Unix(epoch,0).UTC())
    if err != nil { t.Fatalf("fetch: %v", err) }
    if v.LunarTideForce < 79.9 || v.LunarTideForce > 80.1 { t.Fatalf("map0: %v", v.LunarTideForce) }
    // Next value is +10 bps; less than hysteresis 100, should stick to previous
    v2, _ := prov.FetchAt(time.Unix(epoch+1,0).UTC())
    if v2.LunarTideForce < 79.9 || v2.LunarTideForce > 80.1 { t.Fatalf("hysteresis stick expected: %v", v2.LunarTideForce) }
    // Large jump to 5000 bps should pass
    v3, _ := prov.FetchAt(time.Unix(epoch+2,0).UTC())
    if v3.LunarTideForce <= v2.LunarTideForce { t.Fatalf("expected larger after big jump: %v vs %v", v3.LunarTideForce, v2.LunarTideForce) }
    // Max 10000 -> 130.0 approx
    v4, _ := prov.FetchAt(time.Unix(epoch+3,0).UTC())
    if v4.LunarTideForce < 129.9 || v4.LunarTideForce > 130.1 { t.Fatalf("map10000: %v", v4.LunarTideForce) }
    // dataset id read from meta
    if prov.DatasetID() != "ds123" { t.Fatalf("dataset id: %s", prov.DatasetID()) }
}

func TestFileGravimetric_StaleAndCtxCancel(t *testing.T) {
    dir := t.TempDir()
    epoch := time.Date(2025,8,1,0,0,0,0,time.UTC).Unix()
    table := writeGTABWithBPS(t, dir, "gtab_1s.bin", epoch, []uint16{1000, 2000})
    prov, err := NewFileGravimetric(table, "id")
    if err != nil { t.Fatalf("new provider: %v", err) }
    t.Cleanup(func(){ _ = prov.Close() })
    // Before start -> stale true
    if !prov.Stale(time.Unix(epoch-1,0).UTC()) { t.Fatal("expected stale before start") }
    // After end -> stale true
    if !prov.Stale(time.Unix(epoch+10,0).UTC()) { t.Fatal("expected stale after end") }
    // Context cancellation
    ctx, cancel := context.WithCancel(context.Background()); cancel()
    if _, err := prov.Fetch(ctx); err == nil { t.Fatal("expected ctx error") }
}
