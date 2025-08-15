package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/Jthora/autoBotTrader/api/internal/chain"
    httpapi "github.com/Jthora/autoBotTrader/api/internal/http"
    "github.com/Jthora/autoBotTrader/api/internal/providers"
)

func main() {
    cfg := chain.LoadConfigFromEnv()
    // Always construct client (it internally decides enabled vs mock path)
    var chainClient httpapi.ChainClient = chain.NewWithConfig(cfg)
    // Select grav provider
    var grav providers.GravimetricProvider = providers.MockGravimetric{}
    gravMode := "mock"
    if os.Getenv("EPHEM_MODE") == "file" {
        table := os.Getenv("EPHEM_TABLE_PATH")
        if table == "" {
            // Fallback search: ./ephem/gtab_1s.bin relative to working dir
            if _, err := os.Stat("./ephem/gtab_1s.bin"); err == nil {
                table = "./ephem/gtab_1s.bin"
            }
        }
        if table != "" {
            if fg, err := providers.NewFileGravimetric(table, os.Getenv("EPHEM_DATASET_ID")); err == nil {
                grav = fg
                gravMode = "file"
            } else {
                log.Printf("[startup] EPHEM_MODE=file but init failed (table=%s): %v — falling back to mock", table, err)
            }
        } else {
            log.Printf("[startup] EPHEM_MODE=file but EPHEM_TABLE_PATH empty and ./ephem/gtab_1s.bin not found — using mock")
        }
    }
    log.Printf("[startup] grav provider mode: %s", gravMode)
    h := &httpapi.Handlers{Astro: providers.MockAstrology{}, Grav: grav, Chain: chainClient}
    mux := httpapi.NewRouter(h)
    port := os.Getenv("API_PORT")
    if port == "" { port = "8080" }
    srv := &http.Server{
        Addr:              ":"+port,
        Handler:           mux,
        ReadHeaderTimeout: 5 * time.Second,
        ReadTimeout:       10 * time.Second,
        WriteTimeout:      10 * time.Second,
        IdleTimeout:       60 * time.Second,
    }

    // Graceful shutdown on SIGINT/SIGTERM
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        log.Printf("api listening on :%s", port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("server error: %v", err)
        }
    }()

    <-stop
    log.Printf("shutdown requested; closing server...")
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := srv.Shutdown(ctx); err != nil {
        log.Printf("server shutdown error: %v", err)
    }
    // Close file-backed provider if present
    if c, ok := grav.(interface{ Close() error }); ok {
        if err := c.Close(); err != nil {
            log.Printf("provider close error: %v", err)
        }
    }
    log.Printf("shutdown complete")
}

