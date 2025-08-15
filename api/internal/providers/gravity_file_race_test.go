package providers

import (
    "context"
    "sync"
    "testing"
    "time"
)

// TestFileGravimetricConcurrentFetch ensures Fetch / FetchAt are race-free.
func TestFileGravimetricConcurrentFetch(t *testing.T) {
    fg := &FileGravimetric{}
    ctx := context.Background()
    var wg sync.WaitGroup
    n := 50
    for i := 0; i < n; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            _, _ = fg.Fetch(ctx)
            _, _ = fg.FetchAt(time.Now())
        }()
    }
    wg.Wait()
}
