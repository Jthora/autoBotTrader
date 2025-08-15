package httpapi

import (
    "context"
    "net/http"
    "net/http/httptest"
    "sync"
    "testing"
    "time"
    "errors"
    "github.com/Jthora/autoBotTrader/api/internal/providers"
)

// failing astrology provider
type failAstro struct{}
func (f failAstro) Name() string { return "failAstro" }
func (f failAstro) Fetch(ctx context.Context) (providers.AstrologyData, error) { return providers.AstrologyData{}, errors.New("boom") }

// grav provider that records fetch count
type countGrav struct { cnt *int; mu *sync.Mutex }
func (c countGrav) Name() string { return "count" }
func (c countGrav) Fetch(ctx context.Context) (providers.GravimetricData, error) { c.mu.Lock(); *c.cnt++; c.mu.Unlock(); return providers.GravimetricData{LunarTideForce: 100}, nil }
func (c countGrav) Mode() string { return "mock" }
func (c countGrav) DatasetID() string { return "" }
func (c countGrav) Stale(t time.Time) bool { return false }

func TestPredictAstrologyFailureShortCircuit(t *testing.T) {
    cnt := 0
    h := &Handlers{Astro: failAstro{}, Grav: countGrav{cnt: &cnt, mu: &sync.Mutex{}}}
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/predict", nil)
    NewRouter(h).ServeHTTP(rr, req)
    if rr.Code != http.StatusServiceUnavailable { t.Fatalf("expected 503 got %d", rr.Code) }
    if cnt != 0 { t.Fatalf("grav provider should not be fetched when astrology fails first; got %d", cnt) }
}

func TestAstrologyFetchErrorEndpoint(t *testing.T) {
    h := &Handlers{Astro: failAstro{}, Grav: providers.MockGravimetric{}}
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodGet, "/astrology", nil)
    NewRouter(h).ServeHTTP(rr, req)
    if rr.Code != http.StatusServiceUnavailable { t.Fatalf("expected 503 got %d", rr.Code) }
}

func TestParallelPredictAndPushStress(t *testing.T) {
    h := &Handlers{Astro: providers.MockAstrology{}, Grav: providers.MockGravimetric{}, Chain: nil}
    srv := httptest.NewServer(NewRouter(h))
    defer srv.Close()
    client := http.Client{Timeout: 2 * time.Second}
    var wg sync.WaitGroup
    n := 20
    wg.Add(2*n)
    for i := 0; i < n; i++ {
        go func() {
            defer wg.Done()
            resp, err := client.Get(srv.URL + "/predict")
            if err == nil { resp.Body.Close() }
        }()
        go func() {
            defer wg.Done()
            resp, err := client.Post(srv.URL + "/push", "application/json", nil)
            if err == nil { resp.Body.Close() }
        }()
    }
    done := make(chan struct{})
    go func(){ wg.Wait(); close(done) }()
    select {
    case <-done:
    case <-time.After(5 * time.Second):
        t.Fatal("timeout waiting for parallel predict/push")
    }
}

// custom astro provider to force out-of-range pre-normalization (simulate bad input) -- we directly craft abnormal data
type rawAstro struct { v float64 }
func (r rawAstro) Name() string { return "rawAstro" }
func (r rawAstro) Fetch(ctx context.Context) (providers.AstrologyData, error) { return providers.AstrologyData{VolatilityIndex: r.v}, nil }

func TestPushScoreValidation(t *testing.T) {
    // Use absurd volatility to ensure any future change still yields >100 before clamp if normalization adjusts policy.
    h := &Handlers{Astro: rawAstro{v: 9e9}, Grav: providers.MockGravimetric{}}
    rr := httptest.NewRecorder()
    req := httptest.NewRequest(http.MethodPost, "/push", nil)
    NewRouter(h).ServeHTTP(rr, req)
    // Expect 200 because normalization clamps; validation should pass (defensive check uses clamped values)
    if rr.Code != 200 { t.Fatalf("expected 200 got %d", rr.Code) }
}
