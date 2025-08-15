package chain

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    "time"
)

func enabledTestConfig(url string, timeout time.Duration) Config {
    return Config{ContractAddress: "0x1", RPCURL: url, PrivateKey: "k", AccountAddress: "0x2", Timeout: timeout}
}

func TestDisabledConfigMockPath(t *testing.T) {
    c := NewWithConfig(Config{RPCURL: "", ContractAddress: ""})
    if c.cfg.IsEnabled() { t.Fatal("expected disabled") }
    if hash, err := c.PushPrediction(context.Background(), 1, 2); err != nil || hash == "" { t.Fatalf("mock hash expected err=%v hash=%s", err, hash) }
}

func TestRPCFailureClassification(t *testing.T) {
    t.Run("non_200_status_error", func(t *testing.T) {
        srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){ w.WriteHeader(503) }))
        defer srv.Close()
        c := NewWithConfig(enabledTestConfig(srv.URL, 200*time.Millisecond))
        if _, err := c.PushPrediction(context.Background(), 1, 2); err == nil { t.Fatal("expected error on non-200") }
    })
    t.Run("timeout_error", func(t *testing.T) {
        srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){ time.Sleep(150 * time.Millisecond); w.WriteHeader(200); w.Write([]byte(`{"jsonrpc":"2.0","result":1}`)) }))
        defer srv.Close()
        c := NewWithConfig(enabledTestConfig(srv.URL, 50*time.Millisecond))
        if _, err := c.PushPrediction(context.Background(), 1, 2); err == nil { t.Fatal("expected timeout") }
    })
    t.Run("malformed_json_body_ignored", func(t *testing.T) {
        srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){ w.WriteHeader(200); w.Write([]byte("{not json}")) }))
        defer srv.Close()
        c := NewWithConfig(enabledTestConfig(srv.URL, 200*time.Millisecond))
        if _, err := c.PushPrediction(context.Background(), 1, 2); err != nil { t.Fatalf("push prediction should succeed err=%v", err) }
        if v, err := c.GetComposite(context.Background()); err != nil || v != 0 { t.Fatalf("get composite expected 0 err=%v v=%d", err, v) }
    })
}

func TestPushPredictionContextCancel(t *testing.T) {
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){ time.Sleep(100 * time.Millisecond) }))
    defer srv.Close()
    c := NewWithConfig(enabledTestConfig(srv.URL, 500*time.Millisecond))
    ctx, cancel := context.WithCancel(context.Background())
    cancel()
    if _, err := c.PushPrediction(ctx, 1, 1); err == nil { t.Fatal("expected context cancel error") }
}

func TestPushPredictionScoreValidation(t *testing.T) {
    c := NewWithConfig(Config{}) // disabled config path still enforces validation
    if _, err := c.PushPrediction(context.Background(), 101, 0); err == nil { t.Fatal("expected error for astro >100") }
    if _, err := c.PushPrediction(context.Background(), 0, 101); err == nil { t.Fatal("expected error for grav >100") }
}
