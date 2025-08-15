package httpapi

import (
	"encoding/binary"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Jthora/autoBotTrader/api/internal/providers"
)

// writeMinimalGTAB creates a GTAB file with n samples (tide_bps) starting at epochSecond.
func writeMinimalGTAB(t *testing.T, dir string, name string, epochSecond int64, dtNS int64, bps []uint16) string {
    t.Helper()
    n := uint32(len(bps))
    hdr := make([]byte, 5+2+8+8+4+4+16)
    copy(hdr[:5], []byte("GTAB1"))
    binary.LittleEndian.PutUint16(hdr[5:7], 1) // version
    binary.LittleEndian.PutUint64(hdr[7:15], uint64(epochSecond))
    binary.LittleEndian.PutUint64(hdr[15:23], uint64(dtNS))
    binary.LittleEndian.PutUint32(hdr[23:27], n)
    // fields mask: bit0 = tide_bps
    binary.LittleEndian.PutUint32(hdr[27:31], 1)
    rec := make([]byte, n*2)
    for i, v := range bps {
        binary.LittleEndian.PutUint16(rec[i*2:(i+1)*2], v)
    }
    path := filepath.Join(dir, name)
    if err := os.WriteFile(path, append(hdr, rec...), 0o644); err != nil { t.Fatalf("write gtab: %v", err) }
    return path
}

func TestFileModeIntegration_Gravimetrics(t *testing.T) {
    dir := t.TempDir()
    // Table coverage: start 10s ago, dt=1s, 120 samples (~2 min window)
    start := time.Now().UTC().Add(-10 * time.Second).Unix()
    table := writeMinimalGTAB(t, dir, "gtab_1s.bin", start, 1_000_000_000, func() []uint16 {
        vals := make([]uint16, 120)
        for i := range vals { vals[i] = 1000 + uint16(i*(8000/119)) } // ramp 1000 -> ~9000
        return vals
    }())
    // meta file for dataset id
    meta := map[string]string{"dataset_id": "itest_ds"}
    mb, _ := json.Marshal(meta)
    if err := os.WriteFile(filepath.Join(dir, "gtab.meta.json"), mb, 0o644); err != nil { t.Fatalf("write meta: %v", err) }

    fg, err := providers.NewFileGravimetric(table, "")
    if err != nil { t.Fatalf("new file provider: %v", err) }
    defer fg.Close()
    h := &Handlers{Astro: providers.MockAstrology{}, Grav: fg}
    srv := httptest.NewServer(NewRouter(h))
    defer srv.Close()
    // Call endpoint
    resp, err := http.Get(srv.URL + "/gravimetrics")
    if err != nil { t.Fatalf("request: %v", err) }
    if resp.StatusCode != 200 { t.Fatalf("status: %d", resp.StatusCode) }
    var body struct {
        Mode string `json:"mode"`
        DatasetID string `json:"dataset_id"`
        Stale bool `json:"stale"`
        Raw struct { LunarTideForce float64 `json:"lunar_tide_force"` } `json:"raw"`
        NormalizedScore uint32 `json:"normalized_score"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&body); err != nil { t.Fatalf("decode: %v", err) }
    if body.Mode != "file" { t.Fatalf("expected file mode, got %s", body.Mode) }
    if body.DatasetID != "itest_ds" { t.Fatalf("dataset id mismatch: %s", body.DatasetID) }
    if body.Stale { t.Fatalf("expected non-stale data") }
    if body.Raw.LunarTideForce < 80 || body.Raw.LunarTideForce > 130 { t.Fatalf("force out of mapped range: %v", body.Raw.LunarTideForce) }
    if body.NormalizedScore == 0 { t.Fatalf("expected non-zero normalized score") }
}
