package providers

import (
	"context"
	"math/rand"
	"time"
)

type GravimetricData struct {
    LunarTideForce float64 `json:"lunar_tide_force"`
}

type GravimetricProvider interface {
    Name() string
    Fetch(ctx context.Context) (GravimetricData, error)
    // Optional metadata
    Mode() string              // e.g., "mock" | "file" | "algo"
    DatasetID() string         // identifier of dataset when Mode==file
    Stale(now time.Time) bool  // true if data is beyond table range
}

type MockGravimetric struct{}

func (m MockGravimetric) Name() string { return "mock_grav_v1" }

func (m MockGravimetric) Mode() string { return "mock" }
func (m MockGravimetric) DatasetID() string { return "" }
func (m MockGravimetric) Stale(now time.Time) bool { return false }

func (m MockGravimetric) Fetch(ctx context.Context) (GravimetricData, error) {
    rand.Seed(time.Now().UnixNano())
    return GravimetricData{LunarTideForce: 80 + rand.Float64()*50}, nil
}
