package providers

import (
    "encoding/binary"
    "os"
    "path/filepath"
    "testing"
    "time"
    "github.com/Jthora/autoBotTrader/api/internal/ephem"
)

// helper replicating minimal GTAB writer (bps only)
func writeGTAB(t *testing.T, dir string, name string, epoch int64, dtNS int64, vals []uint16) string {
    t.Helper()
    hdr := make([]byte, 5+2+8+8+4+4+16)
    copy(hdr[:5], []byte("GTAB1"))
    binary.LittleEndian.PutUint16(hdr[5:7], 1)
    binary.LittleEndian.PutUint64(hdr[7:15], uint64(epoch))
    binary.LittleEndian.PutUint64(hdr[15:23], uint64(dtNS))
    binary.LittleEndian.PutUint32(hdr[23:27], uint32(len(vals)))
    binary.LittleEndian.PutUint32(hdr[27:31], ephem.FieldTideBPS)
    rec := make([]byte, len(vals)*2)
    for i,v := range vals { binary.LittleEndian.PutUint16(rec[i*2:(i+1)*2], v) }
    p := filepath.Join(dir, name)
    if err := os.WriteFile(p, append(hdr, rec...), 0o644); err != nil { t.Fatalf("write: %v", err) }
    return p
}

func TestHysteresisEdgeEqualBand(t *testing.T) {
    dir := t.TempDir()
    epoch := time.Date(2025,8,1,0,0,0,0,time.UTC).Unix()
    path := writeGTAB(t, dir, "hys.bin", epoch, 1_000_000_000, []uint16{1000,1100,1201})
    t.Setenv("HYSTERESIS_BPS", "100")
    prov, err := NewFileGravimetric(path, "")
    if err != nil { t.Fatalf("new: %v", err) }
    // first -> 1000 maps
    v1,_ := prov.FetchAt(time.Unix(epoch,0).UTC())
    v2,_ := prov.FetchAt(time.Unix(epoch+1,0).UTC()) // +100 exactly should update (edge inclusive)
    if v2.LunarTideForce <= v1.LunarTideForce { t.Fatalf("expected update on edge band: %v vs %v", v2, v1) }
    v3,_ := prov.FetchAt(time.Unix(epoch+2,0).UTC()) // +101 should definitely update
    if v3.LunarTideForce <= v2.LunarTideForce { t.Fatalf("expected update beyond band: %v <= %v", v3, v2) }
}

func TestOutOfRangeClampingAndStale(t *testing.T) {
    dir := t.TempDir()
    epoch := time.Date(2025,8,1,0,0,0,0,time.UTC).Unix()
    path := writeGTAB(t, dir, "clamp.bin", epoch, 1_000_000_000, []uint16{2000,3000})
    prov, err := NewFileGravimetric(path, "")
    if err != nil { t.Fatalf("new: %v", err) }
    before, _ := prov.FetchAt(time.Unix(epoch-10,0).UTC()) // should clamp to first sample
    first,_ := prov.FetchAt(time.Unix(epoch,0).UTC())
    after,_ := prov.FetchAt(time.Unix(epoch+100,0).UTC()) // beyond end -> clamp last sample
    if before.LunarTideForce != first.LunarTideForce { t.Fatalf("before start not clamped: %v vs %v", before, first) }
    if after.LunarTideForce == first.LunarTideForce { t.Fatalf("after end should use last (different) sample") }
    if !prov.Stale(time.Unix(epoch-10,0).UTC()) || !prov.Stale(time.Unix(epoch+100,0).UTC()) { t.Fatalf("stale flag logic incorrect") }
}

func TestDatasetIDResolutionModes(t *testing.T) {
    dir := t.TempDir()
    epoch := time.Now().UTC().Unix()
    path := writeGTAB(t, dir, "ds.bin", epoch, 1_000_000_000, []uint16{0})
    // case 1: explicit id param
    p1, _ := NewFileGravimetric(path, "explicitX")
    if p1.DatasetID() != "explicitX" { t.Fatalf("explicit id not set") }
    // case 2: meta file
    os.WriteFile(filepath.Join(dir, "gtab.meta.json"), []byte("{\n \"dataset_id\": \"metaY\"\n}"), 0o644)
    p2, _ := NewFileGravimetric(path, "")
    if p2.DatasetID() != "metaY" { t.Fatalf("meta id not resolved: %s", p2.DatasetID()) }
    // case 3: malformed meta ignored
    os.WriteFile(filepath.Join(dir, "gtab.meta.json"), []byte("not json"), 0o644)
    p3, _ := NewFileGravimetric(path, "")
    if p3.DatasetID() != "" { t.Fatalf("malformed meta should yield empty id: %s", p3.DatasetID()) }
}
