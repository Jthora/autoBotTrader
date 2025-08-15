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

	"github.com/Jthora/autoBotTrader/api/internal/normalize"
	"github.com/Jthora/autoBotTrader/api/internal/providers"
)

// local helper (duplicate kept local for test isolation)
func writeMinimalGTABPredict(t *testing.T, dir string, name string, epochSecond int64, dtNS int64, bps []uint16) string {
    t.Helper()
    n := uint32(len(bps))
    hdr := make([]byte, 5+2+8+8+4+4+16)
    copy(hdr[:5], []byte("GTAB1"))
    binary.LittleEndian.PutUint16(hdr[5:7], 1)
    binary.LittleEndian.PutUint64(hdr[7:15], uint64(epochSecond))
    binary.LittleEndian.PutUint64(hdr[15:23], uint64(dtNS))
    binary.LittleEndian.PutUint32(hdr[23:27], n)
    binary.LittleEndian.PutUint32(hdr[27:31], 1) // tide_bps
    rec := make([]byte, n*2)
    for i, v := range bps { binary.LittleEndian.PutUint16(rec[i*2:(i+1)*2], v) }
    path := filepath.Join(dir, name)
    if err := os.WriteFile(path, append(hdr, rec...), 0o644); err != nil { t.Fatalf("write gtab: %v", err) }
    return path
}

func TestFileModeIntegration_Predict(t *testing.T) {
    dir := t.TempDir()
    // Use a fixed epoch in the past to avoid boundary/stale flakiness.
    start := time.Now().UTC().Add(-30 * time.Second).Unix()
    table := writeMinimalGTABPredict(t, dir, "gtab_1s.bin", start, 1_000_000_000, []uint16{1000, 3000, 5000, 7000, 9000, 10000})
    // meta
    if err := os.WriteFile(filepath.Join(dir, "gtab.meta.json"), []byte(`{"dataset_id":"itest_predict"}`), 0o644); err != nil { t.Fatalf("meta: %v", err) }
    fg, err := providers.NewFileGravimetric(table, "")
    if err != nil { t.Fatalf("file grav: %v", err) }
    defer fg.Close()
    h := &Handlers{Astro: providers.MockAstrology{}, Grav: fg}
    srv := httptest.NewServer(NewRouter(h))
    defer srv.Close()

    resp, err := http.Get(srv.URL + "/predict")
    if err != nil { t.Fatalf("request: %v", err) }
    if resp.StatusCode != 200 { t.Fatalf("status %d", resp.StatusCode) }
    var body struct {
        Astrology struct { NormalizedScore uint32 `json:"normalized_score"` } `json:"astrology"`
        Gravimetrics struct {
            NormalizedScore uint32 `json:"normalized_score"`
            Mode string `json:"mode"`
            DatasetID string `json:"dataset_id"`
            Stale bool `json:"stale"`
        } `json:"gravimetrics"`
        CompositePreview uint32 `json:"composite_preview"`
    }
    if err := json.NewDecoder(resp.Body).Decode(&body); err != nil { t.Fatalf("decode: %v", err) }
    if body.Gravimetrics.Mode != "file" || body.Gravimetrics.DatasetID != "itest_predict" { t.Fatalf("grav meta missing: %+v", body.Gravimetrics) }
    // Stale may be true if current time surpassed final sample; we tolerate either but assert score present.
    if body.Gravimetrics.NormalizedScore == 0 { t.Fatalf("grav normalized score zero (stale=%v)", body.Gravimetrics.Stale) }
    // Composite should be mean of the two scores (weights 50/50)
    expected := (body.Astrology.NormalizedScore + body.Gravimetrics.NormalizedScore) / 2
    if body.CompositePreview != expected { t.Fatalf("composite mismatch got %d want %d", body.CompositePreview, expected) }
    // sanity: scores within 0..100 range (normalization contract)
    if body.Gravimetrics.NormalizedScore > 100 || body.Astrology.NormalizedScore > 100 { t.Fatalf("score out of range") }
    _ = normalize.AstrologyScore // reference to keep import (future: maybe recompute expected)
}
