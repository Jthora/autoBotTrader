package providers

import (
    "context"
    "encoding/json"
    "os"
    "path/filepath"
    "strconv"
    "sync"
    "time"

    "github.com/Jthora/autoBotTrader/api/internal/ephem"
)

// FileGravimetric implements GravimetricProvider using a GTAB file with tide_bps.
type FileGravimetric struct {
    name      string
    mode      string
    datasetID string
    gtab      *ephem.GTAB
    start     time.Time
    end       time.Time
    // hysteresis in basis points; if 0, no hysteresis
    hysteresis uint16
    // last value cached for hysteresis; 65535 denotes "unset"
    lastBPS uint16
    mu sync.Mutex // protects lastBPS and hysteresis-related decisions
}

// NewFileGravimetric opens the given GTAB file and returns a provider.
// datasetID is a human-readable id from meta JSON; may be empty.
func NewFileGravimetric(path string, datasetID string) (*FileGravimetric, error) {
    g, err := ephem.Open(path)
    if err != nil { return nil, err }
    s, e := g.Coverage()
    hys := uint16(0)
    if v := os.Getenv("HYSTERESIS_BPS"); v != "" {
        if n, err := strconv.Atoi(v); err == nil && n >= 0 && n <= 10000 { hys = uint16(n) }
    }
    if datasetID == "" {
        // Try to read sibling gtab.meta.json
        metaPath := filepath.Join(filepath.Dir(path), "gtab.meta.json")
        if b, err := os.ReadFile(metaPath); err == nil {
            var meta struct{ DatasetID string `json:"dataset_id"` }
            if json.Unmarshal(b, &meta) == nil && meta.DatasetID != "" {
                datasetID = meta.DatasetID
            }
        }
    }
    return &FileGravimetric{
        name: filepath.Base(path),
        mode: "file",
        datasetID: datasetID,
        gtab: g,
        start: s,
        end: e,
    hysteresis: hys,
    lastBPS: 65535,
    }, nil
}

func (f *FileGravimetric) Name() string { return f.name }
func (f *FileGravimetric) Mode() string { return f.mode }
func (f *FileGravimetric) DatasetID() string { return f.datasetID }
func (f *FileGravimetric) Stale(now time.Time) bool { return now.Before(f.start) || now.After(f.end) }

func (f *FileGravimetric) Fetch(ctx context.Context) (GravimetricData, error) {
    if f == nil || f.gtab == nil { return GravimetricData{}, context.Canceled }
    select { case <-ctx.Done(): return GravimetricData{}, ctx.Err(); default: }
    return f.FetchAt(time.Now().UTC())
}

// FetchAt fetches using a specific timestamp (testability and determinism). Not part of the interface.
func (f *FileGravimetric) FetchAt(at time.Time) (GravimetricData, error) {
    if f == nil || f.gtab == nil { return GravimetricData{}, context.Canceled }
    bps, ok := f.gtab.LookupTideBPS(at)
    if !ok {
        if at.Before(f.start) {
            if v, ok := f.gtab.LookupTideBPS(f.start); ok { bps = v } else { bps = 0 }
        } else {
            if v, ok := f.gtab.LookupTideBPS(f.end); ok { bps = v } else { bps = 0 }
        }
    }
    f.mu.Lock()
    if f.hysteresis > 0 && f.lastBPS != 65535 {
        if bps > f.lastBPS {
            if bps - f.lastBPS < f.hysteresis { bps = f.lastBPS }
        } else if f.lastBPS > bps {
            if f.lastBPS - bps < f.hysteresis { bps = f.lastBPS }
        }
    }
    f.lastBPS = bps
    f.mu.Unlock()
    tide := 80.0 + (float64(bps) / 10000.0) * 50.0
    return GravimetricData{LunarTideForce: tide}, nil
}

// Close releases underlying resources (file handles).
func (f *FileGravimetric) Close() error {
    if f != nil && f.gtab != nil { return f.gtab.Close() }
    return nil
}
