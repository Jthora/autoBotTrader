package chain

import (
    "context"
    "testing"
    "time"
)

func TestClientPushPredictionValidation(t *testing.T) {
    c := NewWithConfig(Config{})
    if _, err := c.PushPrediction(context.Background(), 101, 10); err == nil { t.Fatal("expected error") }
    if _, err := c.PushPrediction(context.Background(), 10, 200); err == nil { t.Fatal("expected error") }
}

func TestClientPushPredictionMockSuccess(t *testing.T) {
    c := NewWithConfig(Config{})
    ctx, cancel := context.WithTimeout(context.Background(), time.Second)
    defer cancel()
    if hash, err := c.PushPrediction(ctx, 10, 20); err != nil || hash == "" { t.Fatalf("mock failed err=%v hash=%s", err, hash) }
}

func TestClientPushPredictionCancel(t *testing.T) {
    c := NewWithConfig(Config{})
    ctx, cancel := context.WithCancel(context.Background())
    cancel()
    if _, err := c.PushPrediction(ctx, 1, 1); err == nil { t.Fatalf("expected cancellation error") }
}

func TestUpdateWeightsValidation(t *testing.T) {
    c := NewWithConfig(Config{})
    if _, err := c.UpdateWeights(context.Background(), 0,0,0); err == nil { t.Fatalf("expected zero weights error") }
    if hash, err := c.UpdateWeights(context.Background(), 1,0,0); err != nil || hash == "" { t.Fatalf("expected hash err=%v", err) }
}
