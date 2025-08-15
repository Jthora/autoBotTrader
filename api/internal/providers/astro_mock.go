package providers

import (
    "context"
    "math/rand"
    "time"
)

type AstrologyData struct {
    VolatilityIndex float64 `json:"volatility_index"`
}

type AstrologyProvider interface {
    Name() string
    Fetch(ctx context.Context) (AstrologyData, error)
}

type MockAstrology struct{}

func (m MockAstrology) Name() string { return "mock_astro_v1" }

func (m MockAstrology) Fetch(ctx context.Context) (AstrologyData, error) {
    rand.Seed(time.Now().UnixNano())
    return AstrologyData{VolatilityIndex: rand.Float64() * 720}, nil
}
