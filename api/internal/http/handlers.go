package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/Jthora/autoBotTrader/api/internal/normalize"
	"github.com/Jthora/autoBotTrader/api/internal/providers"
)
func writeJSONError(w http.ResponseWriter, code int, msg string) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(code)
    _ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

type Handlers struct {
    Astro providers.AstrologyProvider
    Grav  providers.GravimetricProvider
    Chain ChainClient
}

// ChainClient abstraction for Starknet interactions.
type ChainClient interface {
    PushPrediction(ctx context.Context, astro, grav uint32) (string, error)
    GetComposite(ctx context.Context) (uint32, error)
    UpdateWeights(ctx context.Context, a, g, ml uint32) (string, error)
}

type AstrologyResponse struct {
    Provider        string                  `json:"provider"`
    Raw             providers.AstrologyData `json:"raw"`
    NormalizedScore uint32                  `json:"normalized_score"`
    CalcVersion     string                  `json:"calc_version"`
}

type GravResponse struct {
    Provider        string                    `json:"provider"`
    Raw             providers.GravimetricData `json:"raw"`
    NormalizedScore uint32                    `json:"normalized_score"`
    CalcVersion     string                    `json:"calc_version"`
    Mode            string                    `json:"mode,omitempty"`
    DatasetID       string                    `json:"dataset_id,omitempty"`
    Stale           bool                      `json:"stale,omitempty"`
}

type PredictResponse struct {
    Astrology        AstrologyResponse `json:"astrology"`
    Gravimetrics     GravResponse      `json:"gravimetrics"`
    CompositePreview uint32            `json:"composite_preview"`
    Weights          map[string]uint32 `json:"weights"`
    Version          string            `json:"version"`
}

type PushRequest struct {
    Force bool `json:"force"`
}

type PushResponse struct {
    TxHash string `json:"tx_hash"`
    DryRun bool   `json:"dry_run"`
    Composite uint32 `json:"composite"`
}

func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    mode, dataset, stale := "", "", false
    if h != nil && h.Grav != nil {
        if m, ok := any(h.Grav).(interface{ Mode() string }); ok { mode = m.Mode() }
        if d, ok := any(h.Grav).(interface{ DatasetID() string }); ok { dataset = d.DatasetID() }
        if s, ok := any(h.Grav).(interface{ Stale(time.Time) bool }); ok { stale = s.Stale(time.Now().UTC()) }
    }
    resp := map[string]any{
        "status": "ok",
        "ts": time.Now().UTC().Format(time.RFC3339Nano),
    }
    if mode != "" { resp["grav_mode"] = mode }
    if dataset != "" { resp["grav_dataset_id"] = dataset }
    if mode != "" { resp["grav_stale"] = stale }
    _ = json.NewEncoder(w).Encode(resp)
}

func (h *Handlers) Astrology(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
    defer cancel()
    data, err := h.Astro.Fetch(ctx)
    if err != nil { writeJSONError(w, http.StatusServiceUnavailable, "astrology_fetch_failed"); return }
    resp := AstrologyResponse{
        Provider: h.Astro.Name(),
        Raw: data,
        NormalizedScore: normalize.AstrologyScore(data.VolatilityIndex),
        CalcVersion: "v1",
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

func (h *Handlers) Gravimetrics(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
    defer cancel()
    data, err := h.Grav.Fetch(ctx)
    if err != nil { writeJSONError(w, http.StatusServiceUnavailable, "gravimetrics_fetch_failed"); return }
    // Optional meta if provider supports it
    mode, dataset, stale := "", "", false
    if m, ok := any(h.Grav).(interface{ Mode() string }); ok { mode = m.Mode() }
    if d, ok := any(h.Grav).(interface{ DatasetID() string }); ok { dataset = d.DatasetID() }
    if s, ok := any(h.Grav).(interface{ Stale(time.Time) bool }); ok { stale = s.Stale(time.Now().UTC()) }
    resp := GravResponse{
        Provider: h.Grav.Name(),
        Raw: data,
        NormalizedScore: normalize.GravimetricScore(data.LunarTideForce),
        CalcVersion: "v1",
        Mode: mode,
        DatasetID: dataset,
        Stale: stale,
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

func composite(a, b, aw, bw uint32) uint32 {
    total := aw + bw
    if total == 0 { return 0 }
    return (a*aw + b*bw) / total
}

func (h *Handlers) Predict(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
    defer cancel()
    aData, aErr := h.Astro.Fetch(ctx); if aErr != nil { writeJSONError(w, http.StatusServiceUnavailable, "astrology_fetch_failed"); return }
    gData, gErr := h.Grav.Fetch(ctx); if gErr != nil { writeJSONError(w, http.StatusServiceUnavailable, "gravimetrics_fetch_failed"); return }
    aScore := normalize.AstrologyScore(aData.VolatilityIndex)
    gScore := normalize.GravimetricScore(gData.LunarTideForce)
    aw, gw := uint32(50), uint32(50)
    // Include meta from Grav provider in Predict as well
    mode, dataset, stale := "", "", false
    if m, ok := any(h.Grav).(interface{ Mode() string }); ok { mode = m.Mode() }
    if d, ok := any(h.Grav).(interface{ DatasetID() string }); ok { dataset = d.DatasetID() }
    if s, ok := any(h.Grav).(interface{ Stale(time.Time) bool }); ok { stale = s.Stale(time.Now().UTC()) }
    resp := PredictResponse{
        Astrology: AstrologyResponse{Provider: h.Astro.Name(), Raw: aData, NormalizedScore: aScore, CalcVersion: "v1"},
        Gravimetrics: GravResponse{Provider: h.Grav.Name(), Raw: gData, NormalizedScore: gScore, CalcVersion: "v1", Mode: mode, DatasetID: dataset, Stale: stale},
        CompositePreview: composite(aScore, gScore, aw, gw),
        Weights: map[string]uint32{"astrology": aw, "gravity": gw, "ml": 0},
        Version: "v1",
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

// Push simulates pushing prediction inputs on-chain (stubbed until Starknet client wired in).
func (h *Handlers) Push(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
    defer cancel()
    aData, _ := h.Astro.Fetch(ctx)
    gData, _ := h.Grav.Fetch(ctx)
    aScore := normalize.AstrologyScore(aData.VolatilityIndex)
    gScore := normalize.GravimetricScore(gData.LunarTideForce)
    if aScore > 100 || gScore > 100 { // defensive, normalization should clamp but guard anyway
        writeJSONError(w, http.StatusBadRequest, "score_out_of_range")
        return
    }
    txHash := "0xDRYRUN"
    dry := true
    if h.Chain != nil && (os.Getenv("PUSH_REAL") == "1") {
        if hash, err := h.Chain.PushPrediction(ctx, aScore, gScore); err == nil {
            txHash = hash
            dry = false
        }
    }
    var composite uint32
    if h.Chain != nil {
        if c, err := h.Chain.GetComposite(ctx); err == nil { composite = c }
    }
    resp := PushResponse{TxHash: txHash, DryRun: dry, Composite: composite}
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(resp)
}
