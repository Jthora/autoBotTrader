package httpapi

import (
    "context"
    "encoding/json"
    "errors"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"

    "github.com/Jthora/autoBotTrader/api/internal/providers"
)

type mockChain struct { hash string; err error }
func (m mockChain) PushPrediction(ctx context.Context, a, g uint32) (string, error) { return m.hash, m.err }
func (m mockChain) GetComposite(ctx context.Context) (uint32, error) { return 0, nil }
func (m mockChain) UpdateWeights(ctx context.Context, a, g, ml uint32) (string, error) { return "", nil }

func newHandlers(chain ChainClient) *Handlers {
    return &Handlers{Astro: providers.MockAstrology{}, Grav: providers.MockGravimetric{}, Chain: chain}
}

// failing grav provider for error path testing
type failGrav struct{}
func (f failGrav) Name() string { return "fail" }
func (f failGrav) Fetch(ctx context.Context) (providers.GravimetricData, error) { return providers.GravimetricData{}, errors.New("boom") }
func (f failGrav) Mode() string { return "mock" }
func (f failGrav) DatasetID() string { return "" }
func (f failGrav) Stale(t time.Time) bool { return false }

func TestHealthIncludesMeta(t *testing.T) {
    h := newHandlers(nil)
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/health", nil)
    NewRouter(h).ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("expected 200 got %d", rr.Code) }
    if ct := rr.Header().Get("Content-Type"); ct != "application/json" { t.Fatalf("unexpected content-type %s", ct) }
    var body map[string]any
    if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil { t.Fatalf("decode: %v", err) }
    if body["status"] != "ok" { t.Fatalf("missing status ok: %+v", body) }
    if _, ok := body["grav_mode"]; !ok { t.Fatalf("expected grav_mode meta present") }
    if _, ok := body["ts"]; !ok { t.Fatalf("expected ts present") }
}

func TestGravimetricsSuccessSchema(t *testing.T) {
    h := newHandlers(nil)
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/gravimetrics", nil)
    NewRouter(h).ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("expected 200 got %d", rr.Code) }
    if ct := rr.Header().Get("Content-Type"); ct != "application/json" { t.Fatalf("content-type: %s", ct) }
    var body struct {
        Provider string `json:"provider"`
        Raw struct { LunarTideForce float64 `json:"lunar_tide_force"` } `json:"raw"`
        NormalizedScore uint32 `json:"normalized_score"`
        CalcVersion string `json:"calc_version"`
        Mode string `json:"mode"`
        Stale bool `json:"stale"`
    }
    if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil { t.Fatalf("decode: %v", err) }
    if body.Provider == "" || body.CalcVersion == "" { t.Fatalf("missing fields: %+v", body) }
    if body.NormalizedScore == 0 { t.Fatalf("expected non-zero normalized score") }
    if body.Mode == "" { t.Fatalf("expected mode meta") }
}

func TestGravimetricsErrorPath(t *testing.T) {
    h := &Handlers{Astro: providers.MockAstrology{}, Grav: failGrav{}}
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/gravimetrics", nil)
    NewRouter(h).ServeHTTP(rr, req)
    if rr.Code != http.StatusServiceUnavailable { t.Fatalf("expected 503 got %d", rr.Code) }
    var body map[string]string
    if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil { t.Fatalf("decode: %v", err) }
    if body["error"] != "gravimetrics_fetch_failed" { t.Fatalf("unexpected error body: %+v", body) }
}

func TestPredictEndpoint(t *testing.T) {
    h := newHandlers(nil)
    srv := httptest.NewServer(NewRouter(h))
    defer srv.Close()
    resp, err := http.Get(srv.URL + "/predict")
    if err != nil { t.Fatalf("request error: %v", err) }
    if resp.StatusCode != 200 { t.Fatalf("expected 200 got %d", resp.StatusCode) }
    var body struct { CompositePreview uint32 `json:"composite_preview"`; Weights map[string]uint32 `json:"weights"` }
    if err := json.NewDecoder(resp.Body).Decode(&body); err != nil { t.Fatalf("decode: %v", err) }
    if body.Weights["astrology"]+body.Weights["gravity"] == 0 { t.Fatal("weights not set") }
}

func TestPushDryRun(t *testing.T) {
    h := newHandlers(nil)
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodPost, "/push", nil)
    NewRouter(h).ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("expected 200 got %d", rr.Code) }
    var body struct { TxHash string `json:"tx_hash"`; DryRun bool `json:"dry_run"` }
    if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil { t.Fatalf("decode: %v", err) }
    if !body.DryRun || body.TxHash == "" { t.Fatalf("expected dryrun with hash got %+v", body) }
}

func TestPushRealChain(t *testing.T) {
    t.Setenv("PUSH_REAL", "1")
    h := newHandlers(mockChain{hash: "0xABC", err: nil})
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodPost, "/push", nil)
    NewRouter(h).ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("expected 200 got %d", rr.Code) }
    var body struct { TxHash string `json:"tx_hash"`; DryRun bool `json:"dry_run"` }
    if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil { t.Fatalf("decode: %v", err) }
    if body.DryRun { t.Fatalf("expected real push got dryrun") }
    if body.TxHash != "0xABC" { t.Fatalf("unexpected hash %s", body.TxHash) }
}

func TestPushChainErrorFallsBack(t *testing.T) {
    t.Setenv("PUSH_REAL", "1")
    h := newHandlers(mockChain{hash: "", err: context.DeadlineExceeded})
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodPost, "/push", nil)
    // set a timeout context to simulate cancellation
    ctx, cancel := context.WithTimeout(req.Context(), 1*time.Nanosecond)
    cancel()
    req = req.WithContext(ctx)
    NewRouter(h).ServeHTTP(rr, req)
    if rr.Code != 200 { t.Fatalf("expected 200 got %d", rr.Code) }
    // Should still return dryrun because chain call failed
    var body struct { DryRun bool `json:"dry_run"` }
    if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil { t.Fatalf("decode: %v", err) }
    if !body.DryRun { t.Fatalf("expected dryrun on error") }
}
